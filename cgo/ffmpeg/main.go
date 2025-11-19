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
	// Allocate output format context
	var outputFormatContext *astiav.FormatContext
	if outputFormatContext, err = astiav.AllocOutputFormatContext(nil, "", outputPath); err != nil {
		return fmt.Errorf("allocating output context failed: %w", err)
	}
	defer outputFormatContext.Free()

	var firstInputStream *astiav.Stream
	var outputStream *astiav.Stream

	// Process each input file
	for _, inputPath := range inputPaths {
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

		// On the first file, setup the output stream
		if firstInputStream == nil {
			for _, is := range inputFormatContext.Streams() {
				if is.CodecParameters().MediaType() == astiav.MediaTypeAudio {
					firstInputStream = is
					break
				}
			}

			if firstInputStream == nil {
				inputFormatContext.CloseInput()
				inputFormatContext.Free()
				return errors.New("no audio stream found in the first input file")
			}

			outputStream = outputFormatContext.NewStream(nil)
			if outputStream == nil {
				inputFormatContext.CloseInput()
				inputFormatContext.Free()
				return errors.New("failed to create new stream in output context")
			}

			if err = firstInputStream.CodecParameters().Copy(outputStream.CodecParameters()); err != nil {
				inputFormatContext.CloseInput()
				inputFormatContext.Free()
				return fmt.Errorf("copying codec parameters failed: %w", err)
			}
			outputStream.CodecParameters().SetCodecTag(0)

			if outputFormatContext.OutputFormat().Flags().Has(astiav.IOFormatFlagGlobalheader) {
				outputStream.Codec().SetFlags(outputStream.Codec().Flags().Add(astiav.CodecFlagGlobalHeader))
			}

			if err = outputFormatContext.OpenIO(nil); err != nil {
				inputFormatContext.CloseInput()
				inputFormatContext.Free()
				return fmt.Errorf("opening output IO failed: %w", err)
			}

			if err = outputFormatContext.WriteHeader(nil); err != nil {
				inputFormatContext.CloseInput()
				inputFormatContext.Free()
				return fmt.Errorf("writing header failed: %w", err)
			}
		}

		// Read packets from input and write to output
		pkt := astiav.AllocPacket()
		for {
			err = inputFormatContext.ReadFrame(pkt)
			if errors.Is(err, astiav.ErrEof) {
				break // End of this file
			} else if err != nil {
				pkt.Free()
				inputFormatContext.CloseInput()
				inputFormatContext.Free()
				return fmt.Errorf("reading frame from %s failed: %w", inputPath, err)
			}

			// Write packet to output
			if pkt.StreamIndex() == firstInputStream.Index() {
				pkt.SetStreamIndex(outputStream.Index())
				pkt.SetPos(-1)
				// No need to rescale timestamps for simple wav concatenation
				if err = outputFormatContext.WriteInterleavedFrame(pkt); err != nil {
					log.Printf("Warning: writing frame failed: %v\n", err)
				}
			}
			pkt.Unref()
		}
		pkt.Free()
		inputFormatContext.CloseInput()
		inputFormatContext.Free()
	}

	// Write trailer
	if err = outputFormatContext.WriteTrailer(); err != nil {
		return fmt.Errorf("writing trailer failed: %w", err)
	}

	return nil
}
