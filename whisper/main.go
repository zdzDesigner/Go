package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultAddr       = "127.0.0.1:8318"
	defaultMaxBytes   = 25 << 20
	defaultTimeout    = 120 * time.Second
	transcriptionsURL = "/v1/audio/transcriptions"
)

type Config struct {
	Addr     string
	CliPath  string
	Model    string
	APIKey   string
	Language string
	Timeout  time.Duration
	MaxBytes int64
}

type Server struct {
	config Config
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type TranscriptionResponse struct {
	Text string `json:"text"`
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	server := &Server{config: config}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.health)
	mux.HandleFunc(transcriptionsURL, server.transcriptions)

	log.Printf("whiper listening on http://%s", config.Addr)
	log.Printf("transcription endpoint: http://%s%s", config.Addr, transcriptionsURL)
	if err := http.ListenAndServe(config.Addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func loadConfig() (Config, error) {
	config := Config{
		Addr: envOr("WHIPER_ADDR", defaultAddr),
		// CliPath:  strings.TrimSpace(os.Getenv("WHISPER_CPP_CLI")),
		// Model:    strings.TrimSpace(os.Getenv("WHISPER_CPP_MODEL")),
		// Language: strings.TrimSpace(os.Getenv("WHIPER_LANGUAGE")),
		APIKey:   strings.TrimSpace(os.Getenv("WHIPER_API_KEY")),
		CliPath:  "/home/zdz/Documents/Try/LLM/fork/whisper.cpp/build/bin/whisper-cli",
		Model:    "/home/zdz/Documents/Try/LLM/fork/whisper.cpp/models/ggml-base.bin",
		Language: "zh",
		Timeout:  defaultTimeout,
		MaxBytes: defaultMaxBytes,
	}

	if config.CliPath == "" {
		return Config{}, errors.New("WHISPER_CPP_CLI is required")
	}
	if config.Model == "" {
		return Config{}, errors.New("WHISPER_CPP_MODEL is required")
	}
	if value := strings.TrimSpace(os.Getenv("WHIPER_TIMEOUT")); value != "" {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf("WHIPER_TIMEOUT=%q is invalid: %w", value, err)
		}
		config.Timeout = duration
	}

	return config, nil
}

func envOr(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func (server *Server) health(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed", "only GET is allowed")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func (server *Server) transcriptions(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed", "only POST is allowed")
		return
	}
	if !server.isAuthorized(request) {
		writeError(writer, http.StatusUnauthorized, "unauthorized", "invalid bearer token")
		return
	}

	request.Body = http.MaxBytesReader(writer, request.Body, server.config.MaxBytes)
	if err := request.ParseMultipartForm(server.config.MaxBytes); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request_error", fmt.Sprintf("failed to parse multipart form: %v", err))
		return
	}

	file, header, err := request.FormFile("file")
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request_error", fmt.Sprintf("missing form file field %q: %v", "file", err))
		return
	}
	defer file.Close()

	workDir, err := os.MkdirTemp("", "whiper-*")
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "server_error", fmt.Sprintf("failed to create temp dir: %v", err))
		return
	}
	defer os.RemoveAll(workDir)

	inputPath := filepath.Join(workDir, cleanFilename(header.Filename))
	input, err := os.Create(inputPath)
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "server_error", fmt.Sprintf("failed to create temp audio file: %v", err))
		return
	}
	if _, err := io.Copy(input, file); err != nil {
		input.Close()
		writeError(writer, http.StatusInternalServerError, "server_error", fmt.Sprintf("failed to save uploaded audio file: %v", err))
		return
	}
	if err := input.Close(); err != nil {
		writeError(writer, http.StatusInternalServerError, "server_error", fmt.Sprintf("failed to close uploaded audio file: %v", err))
		return
	}

	language := strings.TrimSpace(request.FormValue("language"))
	text, err := server.transcribe(request.Context(), inputPath, workDir, language)
	if err != nil {
		log.Printf("transcription failed: file=%s temp=%s error=%v", header.Filename, inputPath, err)
		writeError(writer, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	log.Printf("transcription success: file=%s temp=%s text=%q", header.Filename, inputPath, text)

	writeJSON(writer, http.StatusOK, TranscriptionResponse{Text: text})
}

func (server *Server) isAuthorized(request *http.Request) bool {
	if server.config.APIKey == "" {
		return true
	}
	const prefix = "Bearer "
	header := request.Header.Get("Authorization")
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return subtle.ConstantTimeCompare([]byte(token), []byte(server.config.APIKey)) == 1
}

func (server *Server) transcribe(ctx context.Context, inputPath string, workDir string, language string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, server.config.Timeout)
	defer cancel()

	outputPrefix := filepath.Join(workDir, "transcript")
	args := []string{"-otxt", "-np", "-m", server.config.Model, "-f", inputPath, "-of", outputPrefix}
	if language == "" {
		language = server.config.Language
	}
	if language != "" {
		args = append(args, "-l", language)
	}

	command := exec.CommandContext(ctx, server.config.CliPath, args...)
	log.Printf("running whisper.cpp: cli=%s args=%q", server.config.CliPath, args)
	output, err := command.CombinedOutput()
	log.Printf("whisper.cpp output: audio=%s output=%q", inputPath, strings.TrimSpace(string(output)))
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("whisper.cpp timed out after %s, cli=%s, model=%s, audio=%s", server.config.Timeout, server.config.CliPath, server.config.Model, inputPath)
	}
	if err != nil {
		return "", fmt.Errorf("whisper.cpp failed: cli=%s, model=%s, audio=%s, error=%v, output=%s", server.config.CliPath, server.config.Model, inputPath, err, strings.TrimSpace(string(output)))
	}

	textPath := outputPrefix + ".txt"
	textBytes, err := os.ReadFile(textPath)
	if err != nil {
		return "", fmt.Errorf("failed to read transcription output %s: %w", textPath, err)
	}
	text := strings.TrimSpace(string(textBytes))
	if text == "" {
		return "", fmt.Errorf("empty transcription from whisper.cpp, cli=%s, model=%s, audio=%s", server.config.CliPath, server.config.Model, inputPath)
	}
	if language == "zh" {
		text, err = convertToSimplified(ctx, text)
		if err != nil {
			return "", err
		}
	}
	return text, nil
}

func convertToSimplified(ctx context.Context, text string) (string, error) {
	command := exec.CommandContext(ctx, "opencc", "-c", "t2s")
	command.Stdin = strings.NewReader(text)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("opencc t2s conversion failed: error=%v, output=%s", err, strings.TrimSpace(string(output)))
	}
	converted := strings.TrimSpace(string(output))
	if converted == "" {
		return "", errors.New("opencc t2s conversion returned empty text")
	}
	return converted, nil
}

func cleanFilename(name string) string {
	base := filepath.Base(name)
	if base == "." || base == string(filepath.Separator) || strings.TrimSpace(base) == "" {
		return "audio.wav"
	}
	return base
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		log.Printf("failed to write json response: %v", err)
	}
}

func writeError(writer http.ResponseWriter, status int, errorType string, message string) {
	writeJSON(writer, status, ErrorResponse{Error: ErrorDetail{Type: errorType, Message: message}})
}
