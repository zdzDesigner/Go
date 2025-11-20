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
	// Handle ffmpeg logs
	astiav.SetLogLevel(astiav.LogLevelInfo)
	astiav.SetLogCallback(func(c astiav.Classer, l astiav.LogLevel, fmt, msg string) {
		log.Printf("ffmpeg log: %s", strings.TrimSpace(msg))
	})

	// Setup flags
	output := flag.String("o", "output.wav", "The path to the output WAV file.")
	flag.Parse()
	inputs := flag.Args()

	// Validate inputs
	if *output == "" || len(inputs) == 0 {
		log.Println("Usage: go run . -o <output file> <input file 1> <input file 2> ...")
		return
	}

	if err := concatenate(inputs, *output); err != nil {
		log.Fatalf("Error during concatenation: %v", err)
	}

	log.Printf("Successfully concatenated %d files into %s\n", len(inputs), *output)
}

func concatenate(inputPaths []string, outputPath string) (err error) {
	var outputFormatContext *astiav.FormatContext
	if outputFormatContext, err = astiav.AllocOutputFormatContext(nil, "", outputPath); err != nil {
		return fmt.Errorf("allocating output context failed: %w", err)
	}
	defer outputFormatContext.Free()

	var outputStream *astiav.Stream
	var ptsOffset int64 = 0

	for i, inputPath := range inputPaths {
		log.Printf("Processing input: %s\n", inputPath)

		inputFormatContext := astiav.AllocFormatContext()
		if err = inputFormatContext.OpenInput(inputPath, nil, nil); err != nil {
			inputFormatContext.Free()
			return fmt.Errorf("opening input %s failed: %w", inputPath, err)
		}

		if err = inputFormatContext.FindStreamInfo(nil); err != nil {
			inputFormatContext.CloseInput()
			inputFormatContext.Free()
			return fmt.Errorf("finding stream info for %s failed: %w", inputPath, err)
		}

		var currentInputStream *astiav.Stream
		for _, is := range inputFormatContext.Streams() {
			if is.CodecParameters().MediaType() == astiav.MediaTypeAudio {
				currentInputStream = is
				break
			}
		}

		if currentInputStream == nil {
			inputFormatContext.CloseInput()
			inputFormatContext.Free()
			log.Printf("Warning: no audio stream found in %s, skipping file.", inputPath)
			continue
		}

		if i == 0 {
			outputStream = outputFormatContext.NewStream(nil)
			if outputStream == nil {
				inputFormatContext.CloseInput()
				inputFormatContext.Free()
				return errors.New("failed to create new stream in output context")
			}

			if err = currentInputStream.CodecParameters().Copy(outputStream.CodecParameters()); err != nil {
				inputFormatContext.CloseInput()
				inputFormatContext.Free()
				return fmt.Errorf("copying codec parameters failed: %w", err)
			}
			outputStream.CodecParameters().SetCodecTag(0)

			if !outputFormatContext.OutputFormat().Flags().Has(astiav.IOFormatFlagNofile) {
				var ioCtx *astiav.IOContext
				ioCtx, err = astiav.OpenIOContext(outputPath, astiav.NewIOContextFlags(astiav.IOContextFlagWrite), nil, nil)
				if err != nil {
					inputFormatContext.CloseInput()
					inputFormatContext.Free()
					return fmt.Errorf("opening IO context for %s failed: %w", outputPath, err)
				}
				outputFormatContext.SetPb(ioCtx)
			}

			if err = outputFormatContext.WriteHeader(nil); err != nil {
				inputFormatContext.CloseInput()
				inputFormatContext.Free()
				return fmt.Errorf("writing header failed: %w", err)
			}

			log.Println("First file. Remuxing...")
			err = remux(outputFormatContext, inputFormatContext, outputStream, currentInputStream, ptsOffset)
		} else {
			outParams := outputStream.CodecParameters()
			inParams := currentInputStream.CodecParameters()

			if outParams.SampleRate() == inParams.SampleRate() &&
				outParams.SampleFormat() == inParams.SampleFormat() &&
				outParams.ChannelLayout().Equal(inParams.ChannelLayout()) {
				log.Println("Audio parameters match. Remuxing...")
				err = remux(outputFormatContext, inputFormatContext, outputStream, currentInputStream, ptsOffset)
			} else {
				log.Printf("Audio parameters do not match. Transcoding from %dHz/%s to %dHz/%s...", 
					inParams.SampleRate(), inParams.SampleFormat(), outParams.SampleRate(), outParams.SampleFormat())
				err = transcodeAndMux(outputFormatContext, outputStream, inputFormatContext, currentInputStream, ptsOffset)
			}
		}

		if err != nil {
			inputFormatContext.CloseInput()
			inputFormatContext.Free()
			return err
		}

		durationAV := inputFormatContext.Duration()
		if durationAV > 0 {
			durationOutTb := astiav.RescaleQ(durationAV, astiav.NewRational(1, 1000000), outputStream.TimeBase())
			ptsOffset += durationOutTb
		} else {
			log.Printf("Warning: could not determine duration for %s. Timestamping for next file may be incorrect.", inputPath)
		}

		inputFormatContext.CloseInput()
		inputFormatContext.Free()
	}

	if err = outputFormatContext.WriteTrailer(); err != nil {
		return fmt.Errorf("writing trailer failed: %w", err)
	}

	return nil
}

func remux(octx *astiav.FormatContext, ictx *astiav.FormatContext, ostream *astiav.Stream, istream *astiav.Stream, ptsOffset int64) error {
	pkt := astiav.AllocPacket()
	defer pkt.Free()

	for {
		err := ictx.ReadFrame(pkt)
		if errors.Is(err, astiav.ErrEof) {
			break
		} else if err != nil {
			return fmt.Errorf("reading frame failed: %w", err)
		}

		if pkt.StreamIndex() == istream.Index() {
			pkt.RescaleTs(istream.TimeBase(), ostream.TimeBase())
			pkt.SetPts(pkt.Pts() + ptsOffset)
			pkt.SetDts(pkt.Dts() + ptsOffset)
			pkt.SetStreamIndex(ostream.Index())
			pkt.SetPos(-1)

			if err = octx.WriteInterleavedFrame(pkt); err != nil {
				log.Printf("Warning: writing frame failed: %v\n", err)
			}
		}
		pkt.Unref()
	}
	return nil
}

func transcodeAndMux(octx *astiav.FormatContext, ostream *astiav.Stream, ictx *astiav.FormatContext, istream *astiav.Stream, ptsOffset int64) (err error) {
	// 1. Setup Decoder
	dec := astiav.FindDecoder(istream.CodecParameters().CodecID())
	if dec == nil {
		return errors.New("finding decoder failed")
	}
	decCtx := astiav.AllocCodecContext(dec)
	if decCtx == nil {
		return errors.New("allocating decoder context failed")
	}
	defer decCtx.Free()
	if err = istream.CodecParameters().ToCodecContext(decCtx); err != nil {
		return fmt.Errorf("copying decoder parameters failed: %w", err)
	}
	if err = decCtx.Open(dec, nil); err != nil {
		return fmt.Errorf("opening decoder context failed: %w", err)
	}

	// 2. Setup Encoder
	enc := astiav.FindEncoder(ostream.CodecParameters().CodecID())
	if enc == nil {
		return errors.New("finding encoder failed")
	}
	encCtx := astiav.AllocCodecContext(enc)
	if encCtx == nil {
		return errors.New("allocating encoder context failed")
	}
	defer encCtx.Free()
	encCtx.SetSampleRate(ostream.CodecParameters().SampleRate())
	encCtx.SetSampleFormat(ostream.CodecParameters().SampleFormat())
	encCtx.SetChannelLayout(ostream.CodecParameters().ChannelLayout())
	encCtx.SetTimeBase(astiav.NewRational(1, ostream.CodecParameters().SampleRate()))
	if err = encCtx.Open(enc, nil); err != nil {
		return fmt.Errorf("opening encoder context failed: %w", err)
	}

	// 3. Setup Filter Graph
	filterGraph := astiav.AllocFilterGraph()
	if filterGraph == nil {
		return errors.New("allocating filter graph failed")
	}
	defer filterGraph.Free()

	buffersrc := astiav.FindFilterByName("abuffer")
	buffersrcCtx, err := filterGraph.NewBuffersrcFilterContext(buffersrc, "in")
	if err != nil {
		return fmt.Errorf("creating buffersrc context failed: %w", err)
	}
	buffersrcCtxParams := astiav.AllocBuffersrcFilterContextParameters()
	defer buffersrcCtxParams.Free()
	buffersrcCtxParams.SetChannelLayout(decCtx.ChannelLayout())
	buffersrcCtxParams.SetSampleFormat(decCtx.SampleFormat())
	buffersrcCtxParams.SetSampleRate(decCtx.SampleRate())
	buffersrcCtxParams.SetTimeBase(decCtx.TimeBase())
	if err = buffersrcCtx.SetParameters(buffersrcCtxParams); err != nil {
		return fmt.Errorf("setting buffersrc parameters failed: %w", err)
	}
	if err = buffersrcCtx.Initialize(nil); err != nil {
		return fmt.Errorf("initializing buffersrc context failed: %w", err)
	}

	buffersink := astiav.FindFilterByName("abuffersink")
	buffersinkCtx, err := filterGraph.NewBuffersinkFilterContext(buffersink, "out")
	if err != nil {
		return fmt.Errorf("creating buffersink context failed: %w", err)
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
	log.Printf("Using filter graph: %s", filterStr)

	if err = filterGraph.Parse(filterStr, inputs, outputs); err != nil {
		return fmt.Errorf("parsing filter graph failed: %w", err)
	}
	if err = filterGraph.Configure(); err != nil {
		return fmt.Errorf("configuring filter graph failed: %w", err)
	}

	// 4. Start Processing
	inPkt := astiav.AllocPacket()
	defer inPkt.Free()
	frame := astiav.AllocFrame()
	defer frame.Free()

	processAndWrite := func(f *astiav.Frame) error {
		if err := buffersrcCtx.AddFrame(f, astiav.NewBuffersrcFlags()); err != nil {
			return fmt.Errorf("adding frame to filter graph failed: %w", err)
		}
		for {
			err := buffersinkCtx.GetFrame(frame, astiav.NewBuffersinkFlags())
			if err != nil {
				if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
					break
				}
				return fmt.Errorf("getting frame from filter graph failed: %w", err)
			}
			if err := encCtx.SendFrame(frame); err != nil {
				return fmt.Errorf("sending frame to encoder failed: %w", err)
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
					return fmt.Errorf("receiving packet from encoder failed: %w", err)
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
			return fmt.Errorf("reading frame failed: %w", err)
		}
		if inPkt.StreamIndex() == istream.Index() {
			if err = decCtx.SendPacket(inPkt); err != nil {
				return fmt.Errorf("sending packet to decoder failed: %w", err)
			}
			inPkt.Unref()
			for {
				err = decCtx.ReceiveFrame(frame)
				if err != nil {
					if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
						break
					}
					return fmt.Errorf("receiving frame from decoder failed: %w", err)
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

	// 5. Flush pipeline
	if err = decCtx.SendPacket(nil); err != nil {
		return fmt.Errorf("flushing decoder failed: %w", err)
	}
	for {
		err = decCtx.ReceiveFrame(frame)
		if err != nil {
			if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
				break
			}
			return fmt.Errorf("receiving frame from decoder failed: %w", err)
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
		return fmt.Errorf("flushing encoder failed: %w", err)
	}
	for {
		outPkt := astiav.AllocPacket()
		err = encCtx.ReceivePacket(outPkt)
		if err != nil {
			outPkt.Free()
			if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
				break
			}
			return fmt.Errorf("receiving packet from encoder failed: %w", err)
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
		log.Printf("Warning: writing frame failed: %v\n", err)
	}
	return nil
}
