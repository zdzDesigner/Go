package audio

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/asticode/go-astiav"
)


// ProgressCallback 定义进度回调函数类型
type ProgressCallback func(progress int, currentFile string, fileIndex int, totalFiles int) error

// Joint 合并多个音频文件到指定输出文件，并支持进度回调
func Joint(inputs []string, output string, progressCallback ProgressCallback) {
	// 设置FFmpeg的日志级别为Info
	astiav.SetLogLevel(astiav.LogLevelInfo)
	// 设置日志回调函数，用于打印FFmpeg的内部日志
	astiav.SetLogCallback(func(c astiav.Classer, l astiav.LogLevel, fmt, msg string) {
		log.Printf("ffmpeg log: %s", strings.TrimSpace(msg))
	})

	if err := concatenate(inputs, output, progressCallback); err != nil {
		log.Fatalf("拼接过程中发生错误: %v", err)
	}

	log.Printf("成功将 %d 个文件拼接到 %s\n", len(inputs), output)
}

// 无进度回调版本的Joint函数，保持向后兼容
func JointWithoutProgress(inputs []string, output string) {
	Joint(inputs, output, nil)
}

// 检查可用编码器
func checkAvailableEncoders() {
	log.Println("检查可用编码器...")

	// 尝试获取 MP3 编码器
	mp3Encoder := astiav.FindEncoder(astiav.CodecIDMp3)
	if mp3Encoder != nil {
		log.Println("✓ MP3 编码器可用")
		formats := mp3Encoder.SampleFormats()
		log.Printf("  - 支持的采样格式数量: %d", len(formats))
		for i, format := range formats {
			log.Printf("    %d. %s", i+1, format.Name())
		}
	} else {
		log.Println("✗ MP3 编码器不可用")
	}

	// 尝试获取 AAC 编码器
	aacEncoder := astiav.FindEncoder(astiav.CodecIDAac)
	if aacEncoder != nil {
		log.Println("✓ AAC 编码器可用")
		formats := aacEncoder.SampleFormats()
		log.Printf("  - 支持的采样格式数量: %d", len(formats))
		for i, format := range formats {
			log.Printf("    %d. %s", i+1, format.Name())
		}
	} else {
		log.Println("✗ AAC 编码器不可用")
	}

	// 检查一些常见的音频编码器
	commonEncoders := []struct {
		id   astiav.CodecID
		name string
	}{
		{astiav.CodecIDMp3, "MP3"},
		{astiav.CodecIDAac, "AAC"},
		{astiav.CodecIDOpus, "Opus"},
		{astiav.CodecIDVorbis, "Vorbis"},
		{astiav.CodecIDMp2, "MP2"},
		{astiav.CodecIDFlac, "FLAC"},
	}

	log.Println("其他常见音频编码器状态:")
	for _, enc := range commonEncoders {
		encoder := astiav.FindEncoder(enc.id)
		if encoder != nil {
			log.Printf("  ✓ %s 可用", enc.name)
		} else {
			log.Printf("  ✗ %s 不可用", enc.name)
		}
	}
}

func concatenate(inputPaths []string, outputPath string, progressCallback ProgressCallback) (err error) {
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

		// 检查是否请求MP3输出及编码器可用性
		isMp3Request := strings.HasSuffix(strings.ToLower(outputPath), ".mp3")
		mp3EncoderAvailable := astiav.FindEncoder(astiav.CodecIDMp3) != nil

		// 根据输出文件扩展名选择编码器
		encoderID := getEncoderIDForOutputFile(outputPath)

		// 如果用户请求MP3但编码器不可用，提供清晰的错误信息
		if isMp3Request && !mp3EncoderAvailable {
			return fmt.Errorf("错误: 请求输出MP3格式，但MP3编码器不可用。请使用M4A格式或使用编译了MP3编码支持的FFmpeg库")
		}

		enc := astiav.FindEncoder(encoderID)
		if enc == nil {
			return fmt.Errorf("找不到编码器 ID %v", encoderID)
		}

		// 从输入流获取编码参数
		inputCodecParams := istream.CodecParameters()

		encCtx := astiav.AllocCodecContext(enc)
		if encCtx == nil {
			return errors.New("分配编码器上下文失败")
		}
		defer encCtx.Free()

		// For MP3/AAC, use standard sample rate that is well-supported (44100 or 48000)
		// Use 44100 for compatibility
		encCtx.SetSampleRate(44100)

		// For the encoder, we need to use supported formats
		sampleFormats := enc.SampleFormats()
		if len(sampleFormats) > 0 {
			encCtx.SetSampleFormat(sampleFormats[0]) // Use the first supported sample format of the encoder
		} else {
			// Default to FLTP if no specific format is provided
			encCtx.SetSampleFormat(astiav.SampleFormatFltp)
		}

		// For audio, use a compatible channel layout based on the input
		inputChannelLayout := inputCodecParams.ChannelLayout()
		if inputChannelLayout.Valid() && inputChannelLayout.Channels() > 0 {
			// If input is mono, use mono; if stereo or more, use stereo
			if inputChannelLayout.Channels() == 1 {
				encCtx.SetChannelLayout(astiav.ChannelLayoutMono)
			} else {
				encCtx.SetChannelLayout(astiav.ChannelLayoutStereo)
			}
		} else {
			// Default to stereo if input channel layout is invalid
			encCtx.SetChannelLayout(astiav.ChannelLayoutStereo)
		}

		encCtx.SetBitRate(inputCodecParams.BitRate()) // Use the same bit rate as input, or set a default if zero
		if encCtx.BitRate() == 0 {
			// 设置默认比特率：MP3通常使用128kbps或320kbps，AAC使用128kbps
			if encoderID == astiav.CodecIDMp3 {
				encCtx.SetBitRate(192000) // MP3通常使用更高的比特率获得更好质量
			} else {
				encCtx.SetBitRate(128000) // 默认AAC比特率
			}
		}
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
	for i, inputPath := range inputPaths {
		log.Printf("正在处理输入文件: %s\n", inputPath)

		// 更新进度
		if progressCallback != nil {
			progress := int(float64(i) / float64(len(inputPaths)) * 90) // 预留10%给最后的处理
			if err := progressCallback(progress, inputPath, i+1, len(inputPaths)); err != nil {
				return fmt.Errorf("进度回调错误: %w", err)
			}
		}

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

	// For MP3/AAC encoder compatibility, use standard sample rate (44100)
	encCtx.SetSampleRate(44100)

	// Use encoder's supported sample format instead of the output stream's format
	sampleFormats := enc.SampleFormats()
	if len(sampleFormats) > 0 {
		encCtx.SetSampleFormat(sampleFormats[0]) // Use the first supported sample format of the encoder
	} else {
		// Default to the output stream's format if no specific format is provided by encoder
		encCtx.SetSampleFormat(ostream.CodecParameters().SampleFormat())
	}

	// Use output codec parameters' channel layout if valid, otherwise default to stereo
	outputChannelLayout := ostream.CodecParameters().ChannelLayout()
	if outputChannelLayout.Valid() && outputChannelLayout.Channels() > 0 {
		// If output is mono, use mono; if stereo or more, use stereo
		if outputChannelLayout.Channels() == 1 {
			encCtx.SetChannelLayout(astiav.ChannelLayoutMono)
		} else {
			encCtx.SetChannelLayout(astiav.ChannelLayoutStereo)
		}
	} else {
		// Default to stereo if output channel layout is invalid
		encCtx.SetChannelLayout(astiav.ChannelLayoutStereo)
	}

	// Use the bit rate from output stream, or a default if it's 0
	encCtx.SetBitRate(ostream.CodecParameters().BitRate())
	if encCtx.BitRate() == 0 {
		// 根据输出编码器类型设置默认比特率
		outputCodecID := ostream.CodecParameters().CodecID()
		if outputCodecID == astiav.CodecIDMp3 {
			encCtx.SetBitRate(192000) // MP3通常使用更高的比特率
		} else {
			encCtx.SetBitRate(128000) // 默认AAC比特率
		}
	}
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

	// Ensure the filter uses formats compatible with the encoder
	encSampleFormat := encCtx.SampleFormat()
	encChannelLayout := encCtx.ChannelLayout()
	encSampleRate := encCtx.SampleRate()

	// Use the channels count instead of the layout string representation for channel_layouts
	channelLayoutStr := fmt.Sprintf("aformat=sample_fmts=%s:sample_rates=%d", encSampleFormat.Name(), encSampleRate)
	// Only add channel layout filter if valid
	if encChannelLayout.Valid() && encChannelLayout.Channels() > 0 {
		channelCount := encChannelLayout.Channels()
		if channelCount == 1 {
			channelLayoutStr += ":channel_layouts=mono"
		} else if channelCount == 2 {
			channelLayoutStr += ":channel_layouts=stereo"
		} else {
			// For multi-channel audio, we might need to map to standard layouts or just use channel count
			channelLayoutStr += ":channel_layouts=stereo" // Default to stereo for compatibility
		}
	}

	// Calculate the target frame size based on the encoder's frame size
	frameSize := encCtx.FrameSize()
	if frameSize <= 0 {
		// Default AAC frame size if not specified
		frameSize = 1024
	}

	// Add filters to ensure proper frame sizing for encoder
	filterStr := fmt.Sprintf("%s,aresample=async=1:first_pts=0,asetnsamples=n=%d:p=0", channelLayoutStr, frameSize)
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

// 根据输出文件路径返回相应的编码器ID
func getEncoderIDForOutputFile(outputPath string) astiav.CodecID {
	if strings.HasSuffix(strings.ToLower(outputPath), ".mp3") {
		// 检查MP3编码器是否可用，如果不可用则使用AAC作为替代
		mp3Encoder := astiav.FindEncoder(astiav.CodecIDMp3)
		if mp3Encoder != nil {
			return astiav.CodecIDMp3
		} else {
			// MP3编码器不可用，使用AAC作为替代
			return astiav.CodecIDAac
		}
	} else if strings.HasSuffix(strings.ToLower(outputPath), ".m4a") {
		return astiav.CodecIDAac
	} else {
		// 默认使用AAC，因为M4A是更常见的默认格式
		return astiav.CodecIDAac
	}
}

// 批量处理assets目录下所有1_xxxx.wav文件并转换为mp3格式
func BatchConvert1WavToMp3() {
	// 设置FFmpeg的日志级别为Info
	astiav.SetLogLevel(astiav.LogLevelInfo)
	// 设置日志回调函数，用于打印FFmpeg的内部日志
	astiav.SetLogCallback(func(c astiav.Classer, l astiav.LogLevel, fmt, msg string) {
		log.Printf("ffmpeg log: %s", strings.TrimSpace(msg))
	})

	// 检查可用编码器
	checkAvailableEncoders()

	// 定义assets目录路径
	assetsDir := "/home/zdz/Documents/Try/Go/cgo/ffmpeg/assets"
	outputFile := "batch_output.mp3"

	// 收集所有以"1_"开头的.wav文件
	inputFiles := []string{}

	// 遍历assets目录下的文件
	files, err := ioutil.ReadDir(assetsDir)
	if err != nil {
		log.Fatalf("读取assets目录失败: %v", err)
	}

	for _, file := range files {
		// 检查文件名是否以"1_"开头且以".wav"结尾
		if !file.IsDir() && strings.HasPrefix(file.Name(), "1_") && strings.HasSuffix(strings.ToLower(file.Name()), ".wav") {
			inputFiles = append(inputFiles, filepath.Join(assetsDir, file.Name()))
		}
	}

	if len(inputFiles) == 0 {
		log.Println("未找到符合条件的1_xxxx.wav文件")
		return
	}

	log.Printf("找到 %d 个1_xxxx.wav文件，准备转换为mp3格式\n", len(inputFiles))

	// 定义进度回调函数
	progressCallback := func(progress int, currentFile string, fileIndex int, totalFiles int) error {
		// 打印进度信息
		log.Printf("进度: %d%%, 当前文件: %s (%d/%d)", progress, filepath.Base(currentFile), fileIndex, totalFiles)
		return nil
	}

	// 调用带进度回调的Joint函数进行转换
	Joint(inputFiles, outputFile, progressCallback)

	log.Printf("成功将 %d 个1_xxxx.wav文件转换为 %s\n", len(inputFiles), outputFile)
}
