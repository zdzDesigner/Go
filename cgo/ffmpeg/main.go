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
	areas.SetLogLevel(astiav.LogLevelInfo)
	// 设置日志回调函数，用于打印FFmpeg的内部日志
	areas.SetLogCallback(func(c astiav.Classer, l astiav.LogLevel, fmt, msg string) {
		log.Printf("ffmpeg log: %s", strings.TrimSpace(msg))
	})

	// 设置命令行参数
	output := flag.String("o", "output.m4a", "The path to the output M4A file.")
	flag.Parse()
	inputs := flag.Args()

	// 校验输入参数
	if *output == "" || len(inputs) == 0 {
		log.Println("使用方法: go run . -o <输出文件> <输入文件1> <输入文件2> ...")
		return
	}

	if err := concatenate(inputs, *output); err != nil {
		log.Fatalf("拼接过程中发生错误: %v", err)
	}

	log.Printf("成功将 %d 个文件拼接到 %s\n", len(inputs), *output)
}

func concatenate(inputPaths []string, outputPath string) (err error) {
	// 1. 为输出文件分配一个格式上下文，并让FFmpeg根据文件名推断格式
	outputFormatContext, err := astiav.AllocOutputFormatContext(nil, "", outputPath)
	if err != nil {
		return fmt.Errorf("分配输出格式上下文失败: %w", err)
	}
	defer outputFormatContext.Free()

	outputFormat := outputFormatContext.OutputFormat()
	if outputFormat == nil {
		return fmt.Errorf("无法为 %s 找到合适的输出格式", outputPath)
	}

	var outputStream *astiav.Stream
	var ptsOffset int64 = 0

	// 2. 仅用第一个文件来设置输出流的参数
	if len(inputPaths) > 0 {
		firstInputPath := inputPaths[0]
	
ictx, err := openInput(firstInputPath)
		if err != nil {
			return err
		}
		defer ictx.CloseInput()

		istream, err := findAudioStream(ictx)
		if err != nil {
			return fmt.Errorf("在第一个文件 %s 中: %w", firstInputPath, err)
		}

		// 创建输出流
		outputStream = outputFormatContext.NewStream(nil)
		if outputStream == nil {
			return errors.New("创建输出流失败")
		}

		// 对于M4A输出，我们需要一个压缩编码器，如AAC
		enc := astiav.FindEncoder(astiav.CodecIDAac)
		if enc == nil {
			return errors.New("找不到AAC编码器")
		}

		encCtx := astiav.AllocCodecContext(enc)
		if encCtx == nil {
			return errors.New("分配编码器上下文失败")
		}
		defer encCtx.Free()

		// 设置编码器参数
		encCtx.SetSampleRate(44100) // 推荐使用一个标准采样率，如44100
		encCtx.SetSampleFormat(enc.SampleFormats()[0])
		encCtx.SetChannelLayout(astiav.ChannelLayoutStereo) // 推荐使用标准声道布局，如立体声
		encCtx.SetBitRate(128000)
		encCtx.SetTimeBase(astiav.NewRational(1, encCtx.SampleRate()))

		if outputFormat.Flags().Has(astiav.IOFormatFlagGlobalheader) {
			encCtx.SetFlags(encCtx.Flags().Add(astiav.CodecContextFlagGlobalHeader))
		}

		if err = encCtx.Open(enc, nil); err != nil {
			return fmt.Errorf("打开编码器失败: %w", err)
		}

		if err = outputStream.CodecParameters().FromCodecContext(encCtx); err != nil {
			return fmt.Errorf("从编码器上下文复制参数失败: %w", err)
		}
		outputStream.SetTimeBase(encCtx.TimeBase())
	} else {
		return errors.New("没有输入文件")
	}

	// 3. 打开输出文件IO
	if !outputFormat.Flags().Has(astiav.IOFormatFlagNofile) {
		ioCtx, err := astiav.OpenIOContext(outputPath, astiav.NewIOContextFlags(astiav.IOContextFlagWrite), nil, nil)
		if err != nil {
			return fmt.Errorf("为 %s 打开IO上下文失败: %w", outputPath, err)
		}
		outputFormatContext.SetPb(ioCtx)
	}

	// 4. 写入文件头
	if err = outputFormatContext.WriteHeader(nil); err != nil {
		return fmt.Errorf("写入文件头失败: %w", err)
	}

	// 5. 循环处理所有文件
	for _, inputPath := range inputPaths {
		log.Printf("正在处理输入文件: %s\n", inputPath)
	
ictx, err := openInput(inputPath)
		if err != nil {
			log.Printf("警告: 打开 %s 失败: %v, 跳过此文件。", inputPath, err)
			continue
		}

		istream, err := findAudioStream(ictx)
		if err != nil {
			log.Printf("警告: 在 %s 中未找到音频流: %v, 跳过此文件。", inputPath, err)
		
ictx.CloseInput()
			continue
		}

		// 总是进行转码，以确保所有片段都符合输出流的格式
		err = transcodeAndMux(outputFormatContext, outputStream, ictx, istream, ptsOffset)
		if err != nil {
		
ictx.CloseInput()
			return fmt.Errorf("处理 %s 时出错: %w", inputPath, err)
		}

		durationAV := ictx.Duration()
		if durationAV > 0 {
			durationOutTb := astiav.RescaleQ(durationAV, astiav.NewRational(1, 1000000), outputStream.TimeBase())
			ptsOffset += durationOutTb
		} else {
			log.Printf("警告: 无法确定 %s 的时长。下一个文件的时间戳可能不正确。", inputPath)
		}
	
ictx.CloseInput()
	}

	if err = outputFormatContext.WriteTrailer(); err != nil {
		return fmt.Errorf("写入文件尾失败: %w", err)
	}

	return nil
}

func openInput(path string) (*astiav.FormatContext, error) {
	ictx := astiav.AllocFormatContext()
	if err := ictx.OpenInput(path, nil, nil); err != nil {
		ictx.Free()
		return nil, fmt.Errorf("打开输入文件 %s 失败: %w", path, err)
	}
	if err := ictx.FindStreamInfo(nil); err != nil {
		ictx.CloseInput()
		return nil, fmt.Errorf("查找 %s 的流信息失败: %w", path, err)
	}
	return ictx, nil
}

func findAudioStream(ictx *astiav.FormatContext) (*astiav.Stream, error) {
	for _, stream := range ictx.Streams() {
		if stream.CodecParameters().MediaType() == astiav.MediaTypeAudio {
			return stream, nil
		}
	}
	return nil, errors.New("未找到音频流")
}

func transcodeAndMux(octx *astiav.FormatContext, ostream *astiav.Stream, ictx *astiav.FormatContext, istream *astiav.Stream, ptsOffset int64) (err error) {
	// 1. 设置解码器
	dec := astiav.FindDecoder(istream.CodecParameters().CodecID())
	if dec == nil {
		return errors.New("查找解码器失败")
	}
	decCtx := astiav.AllocCodecContext(dec)
	if decCtx == nil {
		return errors.New("分配解码器上下文失败")
	}
	defer decCtx.Free()
	if err = istream.CodecParameters().ToCodecContext(decCtx); err != nil {
		return fmt.Errorf("复制解码器参数失败: %w", err)
	}
	if err = decCtx.Open(dec, nil); err != nil {
		return fmt.Errorf("打开解码器上下文失败: %w", err)
	}

	// 2. 设置编码器
	enc := astiav.FindEncoder(ostream.CodecParameters().CodecID())
	if enc == nil {
		return errors.New("查找编码器失败")
	}
	encCtx := astiav.AllocCodecContext(enc)
	if encCtx == nil {
		return errors.New("分配编码器上下文失败")
	}
	defer encCtx.Free()
	encCtx.SetSampleRate(ostream.CodecParameters().SampleRate())
	encCtx.SetSampleFormat(ostream.CodecParameters().SampleFormat())
	encCtx.SetChannelLayout(ostream.CodecParameters().ChannelLayout())
	encCtx.SetBitRate(128000)
	encCtx.SetTimeBase(astiav.NewRational(1, encCtx.SampleRate()))
	if octx.OutputFormat().Flags().Has(astiav.IOFormatFlagGlobalheader) {
		encCtx.SetFlags(encCtx.Flags().Add(astiav.CodecContextFlagGlobalHeader))
	}
	if err = encCtx.Open(enc, nil); err != nil {
		return fmt.Errorf("打开编码器上下文失败: %w", err)
	}

	// 3. 设置滤镜图
	filterGraph := astiav.AllocFilterGraph()
	if filterGraph == nil {
		return errors.New("分配滤镜图失败")
	}
	defer filterGraph.Free()

	buffersrc := astiav.FindFilterByName("abuffer")
	buffersrcCtx, err := filterGraph.NewBuffersrcFilterContext(buffersrc, "in")
	if err != nil {
		return fmt.Errorf("创建源滤镜上下文失败: %w", err)
	}
	buffersrcCtxParams := astiav.AllocBuffersrcFilterContextParameters()
	defer buffersrcCtxParams.Free()
	buffersrcCtxParams.SetChannelLayout(decCtx.ChannelLayout())
	buffersrcCtxParams.SetSampleFormat(decCtx.SampleFormat())
	buffersrcCtxParams.SetSampleRate(decCtx.SampleRate())
	buffersrcCtxParams.SetTimeBase(decCtx.TimeBase())
	if err = buffersrcCtx.SetParameters(buffersrcCtxParams); err != nil {
		return fmt.Errorf("设置源滤镜参数失败: %w", err)
	}
	if err = buffersrcCtx.Initialize(nil); err != nil {
		return fmt.Errorf("初始化源滤镜上下文失败: %w", err)
	}

	buffersink := astiav.FindFilterByName("abuffersink")
	buffersinkCtx, err := filterGraph.NewBuffersinkFilterContext(buffersink, "out")
	if err != nil {
		return fmt.Errorf("创建池滤镜上下文失败: %w", err)
	}

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

	filterStr := fmt.Sprintf("aformat=sample_fmts=%s:sample_rates=%d:channel_layouts=%s", encCtx.SampleFormat().Name(), encCtx.SampleRate(), encCtx.ChannelLayout().String())
	log.Printf("正在使用滤镜图: %s", filterStr)

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

	processAndWrite := func(f *astiav.Frame) error {
		if err := buffersrcCtx.AddFrame(f, astiav.NewBuffersrcFlags()); err != nil {
			return fmt.Errorf("向滤镜图添加帧失败: %w", err)
		}
		for {
			err := buffersinkCtx.GetFrame(frame, astiav.NewBuffersinkFlags())
			if err != nil {
				if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
					break
				}
				return fmt.Errorf("从滤镜图获取帧失败: %w", err)
			}
			if err := encCtx.SendFrame(frame); err != nil {
				return fmt.Errorf("向编码器发送帧失败: %w", err)
			}
			frame.Unref()
			for {
				outPkt := astiav.AllocPacket()
				err := encCtx.ReceivePacket(outPkt)
				if err != nil {
					outPkt.Free()
					if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
						break
					}
					return fmt.Errorf("从编码器接收数据包失败: %w", err)
				}
				if err := writePacket(outPkt, encCtx, octx, ostream, ptsOffset); err != nil {
					outPkt.Free()
					return err
				}
				outPkt.Free()
			}
		}
		return nil
	}

	for {
		err = ictx.ReadFrame(inPkt)
		if err != nil {
			if errors.Is(err, astiav.ErrEof) {
				break
			}
			return fmt.Errorf("读取数据包失败: %w", err)
		}
		if inPkt.StreamIndex() == istream.Index() {
			if err = decCtx.SendPacket(inPkt); err != nil {
				return fmt.Errorf("向解码器发送数据包失败: %w", err)
			}
			inPkt.Unref()
			for {
				err = decCtx.ReceiveFrame(frame)
				if err != nil {
					if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
						break
					}
					return fmt.Errorf("从解码器接收帧失败: %w", err)
				}
				frame.SetPts(frame.Pts())
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
	if err = decCtx.SendPacket(nil); err != nil {
		return fmt.Errorf("清空解码器失败: %w", err)
	}
	for {
		err = decCtx.ReceiveFrame(frame)
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
	if err = processAndWrite(nil); err != nil {
		return err
	}
	if err = encCtx.SendFrame(nil); err != nil {
		return fmt.Errorf("清空编码器失败: %w", err)
	}
	for {
		outPkt := astiav.AllocPacket()
		err = encCtx.ReceivePacket(outPkt)
		if err != nil {
			outPkt.Free()
			if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
				break
			}
			return fmt.Errorf("从编码器接收数据包失败: %w", err)
		}
		if err := writePacket(outPkt, encCtx, octx, ostream, ptsOffset); err != nil {
			outPkt.Free()
			return err
		}
		outPkt.Free()
	}

	return nil
}

func writePacket(pkt *astiav.Packet, encCtx *astiav.CodecContext, octx *astiav.FormatContext, ostream *astiav.Stream, ptsOffset int64) error {
	pkt.RescaleTs(encCtx.TimeBase(), ostream.TimeBase())
	pkt.SetPts(pkt.Pts() + ptsOffset)
	pkt.SetDts(pkt.Dts() + ptsOffset)
	pkt.SetStreamIndex(ostream.Index())
	pkt.SetPos(-1)

	if err := octx.WriteInterleavedFrame(pkt); err != nil {
		log.Printf("警告: 写入交错帧失败: %v\n", err)
	}
	return nil
}