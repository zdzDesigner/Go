package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/asticode/go-astiav"
)

func main() {
	// 设置FFmpeg的日志级别为Info
	astiav.SetLogLevel(astiav.LogLevelInfo)
	// 设置日志回调函数，用于打印FFmpeg的内部日志
	astiav.SetLogCallback(func(c astiav.Classer, l astiav.LogLevel, fmt, msg string) {
		log.Printf("ffmpeg log: %s", strings.TrimSpace(msg))
	})

	// 设置命令行参数
	// -o: 指定输出文件名，默认为 "output.wav"
	output := flag.String("o", "output.m4a", "The path to the output M4A file.")
	flag.Parse()
	// 获取所有非标志参数作为输入文件列表
	inputs := flag.Args()

	// 校验输入参数
	if *output == "" || len(inputs) == 0 {
		log.Println("使用方法: go run . -o <输出文件> <输入文件1> <输入文件2> ...")
		return
	}

	// 调用核心拼接函数
	if err := concatenate(inputs, *output); err != nil {
		log.Fatalf("拼接过程中发生错误: %v", err)
	}

	log.Printf("成功将 %d 个文件拼接到 %s\n", len(inputs), *output)
}

// concatenate 函数负责处理所有输入文件，并将它们拼接成一个输出文件。
// 它会智能地判断输入文件的音频参数是否一致。
// 如果一致，则执行高效的流拷贝（remux）；
// 如果不一致，则自动进行转码（transcode）以统一格式。
func concatenate(inputPaths []string, outputPath string) (err error) {
	// 为输出文件分配一个格式上下文（FormatContext）
	outputFormatContext, err := astiav.AllocOutputFormatContext(nil, "", outputPath)
	if err != nil {
		return fmt.Errorf("分配输出格式上下文失败: %w", err)
	}
	defer outputFormatContext.Free()

	var outputStream *astiav.Stream // 指向输出文件的音频流
	var ptsOffset int64 = 0         // 用于累加每个文件时长，作为下一个文件的时间戳偏移量

	// 遍历处理每一个输入文件
	for i, inputPath := range inputPaths {
		log.Printf("正在处理输入文件: %s\n", inputPath)

		// 为当前输入文件分配格式上下文
		inputFormatContext := astiav.AllocFormatContext()
		defer inputFormatContext.Free() // 确保在出错时释放

		// 使用 defer 确保在函数退出前关闭输入上下文
		if err = inputFormatContext.OpenInput(inputPath, nil, nil); err != nil {
			return fmt.Errorf("打开输入文件 %s 失败: %w", inputPath, err)
		}
		defer inputFormatContext.CloseInput()

		// 查找流信息
		if err = inputFormatContext.FindStreamInfo(nil); err != nil {
			return fmt.Errorf("查找 %s 的流信息失败: %w", inputPath, err)
		}

		// 找到音频流
		var currentInputStream *astiav.Stream
		for _, stream := range inputFormatContext.Streams() {
			if stream.CodecParameters().MediaType() == astiav.MediaTypeAudio {
				currentInputStream = stream
				break
			}
		}

		// 如果找不到音频流，则跳过此文件
		if currentInputStream == nil {
			log.Printf("警告: 在 %s 中未找到音频流，跳过此文件。", inputPath)
			continue
		}

		// --- 核心逻辑 ---
		if i == 0 {
			// 对于第一个文件，设置输出流的参数并写入文件头
			outputStream = outputFormatContext.NewStream(nil)
			if outputStream == nil {
				return errors.New("在输出上下文中创建新流失败")
			}

			// 将第一个输入流的编解码器参数复制到输出流
			if err = currentInputStream.CodecParameters().Copy(outputStream.CodecParameters()); err != nil {
				return fmt.Errorf("复制编解码器参数失败: %w", err)
			}
			outputStream.CodecParameters().SetCodecTag(0) // WAV格式通常不需要特定的编解码器标签

			// 如果输出格式需要文件I/O（而不是像网络流那样），则打开文件
			if !outputFormatContext.OutputFormat().Flags().Has(astiav.IOFormatFlagNofile) {
				var ioCtx *astiav.IOContext
				ioCtx, err = astiav.OpenIOContext(outputPath, astiav.NewIOContextFlags(astiav.IOContextFlagWrite), nil, nil)
				if err != nil {
					return fmt.Errorf("为 %s 打开IO上下文失败: %w", outputPath, err)
				}
				outputFormatContext.SetPb(ioCtx) // 将IO上下文与格式上下文关联
			}

			// 写入输出文件的文件头
			if err = outputFormatContext.WriteHeader(nil); err != nil {
				return fmt.Errorf("写入文件头失败: %w", err)
			}

			log.Println("第一个文件，执行流拷贝(Remuxing)...")
			err = remux(outputFormatContext, inputFormatContext, outputStream, currentInputStream, ptsOffset)
		} else {
			// 对于后续文件，检查其音频参数是否与第一个文件（即输出流）兼容
			outParams := outputStream.CodecParameters()
			inParams := currentInputStream.CodecParameters()

			if outParams.SampleRate() == inParams.SampleRate() &&
				outParams.SampleFormat() == inParams.SampleFormat() &&
				outParams.ChannelLayout().Equal(inParams.ChannelLayout()) {
				// 参数兼容，直接进行流拷贝
				log.Println("音频参数匹配，执行流拷贝(Remuxing)...")
				err = remux(outputFormatContext, inputFormatContext, outputStream, currentInputStream, ptsOffset)
			} else {
				// 参数不兼容，需要进行转码
				log.Printf("音频参数不匹配。正在转码: 从 %dHz/%s 转为 %dHz/%s...",
					inParams.SampleRate(), inParams.SampleFormat(), outParams.SampleRate(), outParams.SampleFormat())
				err = transcodeAndMux(outputFormatContext, outputStream, inputFormatContext, currentInputStream, ptsOffset)
			}
		}

		// 如果处理过程中出错，立即返回
		if err != nil {
			return err
		}

		// 更新时间戳偏移量，为下一个文件做准备
		durationAV := inputFormatContext.Duration()
		if durationAV > 0 {
			// 将当前文件的时长（以AV_TIME_BASE为单位）转换成输出流的时间基单位
			durationOutTb := astiav.RescaleQ(durationAV, astiav.NewRational(1, 1000000), outputStream.TimeBase())
			ptsOffset += durationOutTb // 累加时长
		} else {
			log.Printf("警告: 无法确定 %s 的时长。下一个文件的时间戳可能不正确。", inputPath)
		}
	}

	// 所有文件处理完毕后，写入输出文件的文件尾
	if err = outputFormatContext.WriteTrailer(); err != nil {
		return fmt.Errorf("写入文件尾失败: %w", err)
	}

	return nil
}

// remux 函数执行简单的流拷贝。它从输入上下文读取数据包，调整时间戳后直接写入输出上下文。
// 适用于音视频参数完全兼容的场景。
func remux(octx *astiav.FormatContext, ictx *astiav.FormatContext, ostream *astiav.Stream, istream *astiav.Stream, ptsOffset int64) error {
	pkt := astiav.AllocPacket()
	defer pkt.Free()

	for {
		// 从输入文件读取一个数据包
		err := ictx.ReadFrame(pkt)
		if errors.Is(err, astiav.ErrEof) {
			return nil // 文件读取完毕
		}
		if err != nil {
			return fmt.Errorf("读取数据包失败: %w", err)
		}

		// 只处理属于我们关心的音频流的数据包
		if pkt.StreamIndex() == istream.Index() {
			// 重新计算时间戳，从输入流的时间基转换到输出流的时间基
			pkt.RescaleTs(istream.TimeBase(), ostream.TimeBase())
			// 加上之前所有文件的时长偏移量
			pkt.SetPts(pkt.Pts() + ptsOffset)
			pkt.SetDts(pkt.Dts() + ptsOffset)
			// 设置包的流索引为输出流的索引
			pkt.SetStreamIndex(ostream.Index())
			pkt.SetPos(-1) // 重置位置信息

			// 将处理后的数据包写入输出文件
			if err = octx.WriteInterleavedFrame(pkt); err != nil {
				log.Printf("警告: 写入交错帧失败: %v\n", err)
			}
		}
		// 释放数据包的引用，准备下一次读取
		pkt.Unref()
	}
}

// transcodeAndMux 函数处理不兼容的音频流。它建立一个完整的“解码 -> 滤镜 -> 编码”流水线。
func transcodeAndMux(octx *astiav.FormatContext, ostream *astiav.Stream, ictx *astiav.FormatContext, istream *astiav.Stream, ptsOffset int64) (err error) {
	// 1. 设置解码器
	dec := astiav.FindDecoder(istream.CodecParameters().CodecID())
	if dec == nil {
		return errors.New("查找解码器失败")
	}
	dec_ctx := astiav.AllocCodecContext(dec)
	if dec_ctx == nil {
		return errors.New("分配解码器上下文失败")
	}
	defer dec_ctx.Free()
	if err = istream.CodecParameters().ToCodecContext(dec_ctx); err != nil {
		return fmt.Errorf("复制解码器参数失败: %w", err)
	}
	if err = dec_ctx.Open(dec, nil); err != nil {
		return fmt.Errorf("打开解码器上下文失败: %w", err)
	}

	// 2. 设置编码器
	enc := astiav.FindEncoder(ostream.CodecParameters().CodecID())
	if enc == nil {
		return errors.New("查找编码器失败")
	}
	enc_ctx := astiav.AllocCodecContext(enc)
	if enc_ctx == nil {
		return errors.New("分配编码器上下文失败")
	}
	defer enc_ctx.Free()
	// 编码器的参数必须与输出流保持一致
	enc_ctx.SetBitRate(128000) // 设置比特率为 128kbps (128000 bits/second)。这个值会影响音质和文件大小。
	enc_ctx.SetSampleRate(ostream.CodecParameters().SampleRate())
	enc_ctx.SetSampleFormat(ostream.CodecParameters().SampleFormat())
	enc_ctx.SetChannelLayout(ostream.CodecParameters().ChannelLayout())
	enc_ctx.SetTimeBase(astiav.NewRational(1, ostream.CodecParameters().SampleRate()))
	if err = enc_ctx.Open(enc, nil); err != nil {
		return fmt.Errorf("打开编码器上下文失败: %w", err)
	}

	// 3. 设置滤镜图 (Filter Graph)
	filterGraph := astiav.AllocFilterGraph()
	if filterGraph == nil {
		return errors.New("分配滤镜图失败")
	}
	defer filterGraph.Free()

	// 创建源滤镜 (abuffer)，用于接收解码后的音频帧
	buffersrc := astiav.FindFilterByName("abuffer")
	buffersrcCtx, err := filterGraph.NewBuffersrcFilterContext(buffersrc, "in")
	if err != nil {
		return fmt.Errorf("创建源滤镜上下文失败: %w", err)
	}
	// 配置源滤镜的参数，使其与解码器的输出匹配
	buffersrcCtxParams := astiav.AllocBuffersrcFilterContextParameters()
	defer buffersrcCtxParams.Free()
	buffersrcCtxParams.SetChannelLayout(dec_ctx.ChannelLayout())
	buffersrcCtxParams.SetSampleFormat(dec_ctx.SampleFormat())
	buffersrcCtxParams.SetSampleRate(dec_ctx.SampleRate())
	buffersrcCtxParams.SetTimeBase(dec_ctx.TimeBase())
	if err = buffersrcCtx.SetParameters(buffersrcCtxParams); err != nil {
		return fmt.Errorf("设置源滤镜参数失败: %w", err)
	}
	if err = buffersrcCtx.Initialize(nil); err != nil {
		return fmt.Errorf("初始化源滤镜上下文失败: %w", err)
	}

	// 创建池滤镜 (abuffersink)，用于输出处理后的音频帧
	buffersink := astiav.FindFilterByName("abuffersink")
	buffersinkCtx, err := filterGraph.NewBuffersinkFilterContext(buffersink, "out")
	if err != nil {
		return fmt.Errorf("创建池滤镜上下文失败: %w", err)
	}

	// 链接滤镜图的输入和输出
	outputs := astiav.AllocFilterInOut()
	defer outputs.Free()
	outputs.SetName("in")
	outputs.SetFilterContext(buffersrcCtx.FilterContext())
	outputs.SetPadIdx(0)

	inputs := astiav.AllocFilterInOut()
	defer inputs.Free()
	inputs.SetName("out")
	inputs.SetFilterContext(buffersinkCtx.FilterContext())
	inputs.SetPadIdx(0)

	// 定义滤镜链。`aformat`滤镜会自动处理采样率、样本格式和声道布局的转换。
	filterStr := fmt.Sprintf("aformat=sample_fmts=%s:sample_rates=%d:channel_layouts=%s", enc_ctx.SampleFormat().Name(), enc_ctx.SampleRate(), enc_ctx.ChannelLayout().String())
	log.Printf("正在使用滤镜图: %s", filterStr)

	// 解析并配置滤镜图
	if err = filterGraph.Parse(filterStr, inputs, outputs); err != nil {
		return fmt.Errorf("解析滤镜图失败: %w", err)
	}
	if err = filterGraph.Configure(); err != nil {
		return fmt.Errorf("配置滤镜图失败: %w", err)
	}

	// 4. 开始处理
	inPkt := astiav.AllocPacket()
	defer inPkt.Free()
	frame := astiav.AllocFrame()
	defer frame.Free()

	// processAndWrite 是一个辅助闭包，处理从解码->滤镜->编码->写入的完整流程
	processAndWrite := func(f *astiav.Frame) error {
		// 将帧送入滤镜图
		if err := buffersrcCtx.AddFrame(f, astiav.NewBuffersrcFlags()); err != nil {
			return fmt.Errorf("向滤镜图添加帧失败: %w", err)
		}
		for {
			// 从滤镜图获取处理后的帧
			err := buffersinkCtx.GetFrame(frame, astiav.NewBuffersinkFlags())
			if err != nil {
				if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
					break
				}
				return fmt.Errorf("从滤镜图获取帧失败: %w", err)
			}
			// 将处理后的帧送入编码器
			if err := enc_ctx.SendFrame(frame); err != nil {
				return fmt.Errorf("向编码器发送帧失败: %w", err)
			}
			frame.Unref()
			for {
				// 从编码器获取编码后的数据包
				outPkt := astiav.AllocPacket()
				err := enc_ctx.ReceivePacket(outPkt)
				if err != nil {
					outPkt.Free()
					if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
						break
					}
					return fmt.Errorf("从编码器接收数据包失败: %w", err)
				}
				// 写入数据包
				if err := writePacket(outPkt, enc_ctx, octx, ostream, ptsOffset); err != nil {
					outPkt.Free()
					return err
				}
				outPkt.Free()
			}
		}
		return nil
	}

	// 主循环：读取 -> 解码 -> 处理
	for {
		err = ictx.ReadFrame(inPkt)
		if err != nil {
			if errors.Is(err, astiav.ErrEof) {
				break // 文件结束
			}
			return fmt.Errorf("读取数据包失败: %w", err)
		}
		if inPkt.StreamIndex() == istream.Index() {
			// 将数据包送入解码器
			if err = dec_ctx.SendPacket(inPkt); err != nil {
				return fmt.Errorf("向解码器发送数据包失败: %w", err)
			}
			inPkt.Unref()
			for {
				// 从解码器获取解码后的帧
				err = dec_ctx.ReceiveFrame(frame)
				if err != nil {
					if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
						break
					}
					return fmt.Errorf("从解码器接收帧失败: %w", err)
				}
				frame.SetPts(frame.Pts()) // 确保PTS有效
				// 处理并写入帧
				if err = processAndWrite(frame); err != nil {
					return err
				}
				frame.Unref()
			}
		} else {
			inPkt.Unref()
		}
	}

	// 5. 清空流水线中剩余的数据
	// 清空解码器
	if err = dec_ctx.SendPacket(nil); err != nil {
		return fmt.Errorf("清空解码器失败: %w", err)
	}
	for {
		err = dec_ctx.ReceiveFrame(frame)
		if err != nil {
			if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
				break
			}
			return fmt.Errorf("从解码器接收帧失败: %w", err)
		}
		if err = processAndWrite(frame); err != nil {
			return err
		}
		frame.Unref()
	}
	// 清空滤镜图
	if err = processAndWrite(nil); err != nil {
		return err
	}
	// 清空编码器
	if err = enc_ctx.SendFrame(nil); err != nil {
		return fmt.Errorf("清空编码器失败: %w", err)
	}
	for {
		outPkt := astiav.AllocPacket()
		err = enc_ctx.ReceivePacket(outPkt)
		if err != nil {
			outPkt.Free()
			if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
				break
			}
			return fmt.Errorf("从编码器接收数据包失败: %w", err)
		}
		if err := writePacket(outPkt, enc_ctx, octx, ostream, ptsOffset); err != nil {
			outPkt.Free()
			return err
		}
		outPkt.Free()
	}

	return nil
}

// writePacket 函数负责调整数据包的时间戳并将其写入输出文件。
func writePacket(pkt *astiav.Packet, enc_ctx *astiav.CodecContext, octx *astiav.FormatContext, ostream *astiav.Stream, ptsOffset int64) error {
	// 从编码器的时间基转换到输出流的时间基
	pkt.RescaleTs(enc_ctx.TimeBase(), ostream.TimeBase())
	// 加上偏移量
	pkt.SetPts(pkt.Pts() + ptsOffset)
	pkt.SetDts(pkt.Dts() + ptsOffset)
	// 设置流索引
	pkt.SetStreamIndex(ostream.Index())
	pkt.SetPos(-1)

	// 写入
	if err := octx.WriteInterleavedFrame(pkt); err != nil {
		log.Printf("警告: 写入交错帧失败: %v\n", err)
	}
	return nil
}
