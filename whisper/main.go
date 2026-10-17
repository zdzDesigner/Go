package main

import (
	"bytes"
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
	defaultLLMTimeout = 5 * time.Second
	transcriptionsURL = "/v1/audio/transcriptions"
)

type Config struct {
	Addr        string
	CliPath     string
	Model       string
	APIKey      string
	Language    string
	Corrections string
	LLMEnabled  bool
	LLMCliPath  string
	LLMModel    string
	LLMHotWords string
	LLMTimeout  time.Duration
	Timeout     time.Duration
	MaxBytes    int64
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
		APIKey:      strings.TrimSpace(os.Getenv("WHIPER_API_KEY")),
		CliPath:     "/home/zdz/Documents/Try/LLM/fork/whisper.cpp/build/bin/whisper-cli",
		Model:       "/home/zdz/Documents/Try/LLM/fork/whisper.cpp/models/ggml-base.bin",
		Language:    "zh",
		Corrections: strings.TrimSpace(envOr("WHIPER_CORRECTIONS", defaultCorrections())),
		LLMCliPath:  strings.TrimSpace(os.Getenv("WHIPER_CORRECTOR_CLI")),
		LLMModel:    strings.TrimSpace(os.Getenv("WHIPER_CORRECTOR_MODEL")),
		LLMHotWords: strings.TrimSpace(envOr("WHIPER_CORRECTOR_HOTWORDS", "欧喷扣=OpenCode，欧盆扣=OpenCode，扣德=code")),
		LLMTimeout:  defaultLLMTimeout,
		Timeout:     defaultTimeout,
		MaxBytes:    defaultMaxBytes,
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
	if value := strings.TrimSpace(os.Getenv("WHIPER_CORRECTOR_ENABLED")); value != "" {
		enabled, err := parseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf("WHIPER_CORRECTOR_ENABLED=%q is invalid: %w", value, err)
		}
		config.LLMEnabled = enabled
	}
	if value := strings.TrimSpace(os.Getenv("WHIPER_CORRECTOR_TIMEOUT")); value != "" {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf("WHIPER_CORRECTOR_TIMEOUT=%q is invalid: %w", value, err)
		}
		config.LLMTimeout = duration
	}
	if config.LLMEnabled {
		if config.LLMCliPath == "" {
			return Config{}, errors.New("WHIPER_CORRECTOR_CLI is required when WHIPER_CORRECTOR_ENABLED is true")
		}
		if config.LLMModel == "" {
			return Config{}, errors.New("WHIPER_CORRECTOR_MODEL is required when WHIPER_CORRECTOR_ENABLED is true")
		}
	}

	return config, nil
}

func defaultCorrections() string {
	return strings.Join([]string{
		"欧喷扣=OpenCode",
		"欧盆扣=OpenCode",
		"扣德=code",
		"向木=项目",
		"衣赖=依赖",
		"从新=重新",
		"侧试=测试",
		"单元侧试=单元测试",
		"全线=权限",
		"权现=权限",
		"效验=校验",
		"权现效验=权限校验",
		"带办=待办",
		"日成=日程",
		"反回=返回",
	}, "，")
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("expected one of true,false,1,0,yes,no,on,off")
	}
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
		// text, err = applyTextCorrections(text, server.config.Corrections)
		// if err != nil {
		// 	return "", err
		// }
		// text, err = server.correctChineseText(ctx, text)
		// if err != nil {
		// 	return "", err
		// }
	}
	return text, nil
}

func applyTextCorrections(text string, corrections string) (string, error) {
	if strings.TrimSpace(corrections) == "" {
		return text, nil
	}
	corrected := text
	for _, item := range strings.FieldsFunc(corrections, func(char rune) bool { return char == ',' || char == '，' || char == '\n' }) {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid WHIPER_CORRECTIONS item %q: expected wrong=right", item)
		}
		wrong := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		if wrong == "" || right == "" {
			return "", fmt.Errorf("invalid WHIPER_CORRECTIONS item %q: wrong and right must be non-empty", item)
		}
		corrected = strings.ReplaceAll(corrected, wrong, right)
	}
	return corrected, nil
}

func (server *Server) correctChineseText(ctx context.Context, text string) (string, error) {
	if !server.config.LLMEnabled {
		return text, nil
	}

	ctx, cancel := context.WithTimeout(ctx, server.config.LLMTimeout)
	defer cancel()

	systemPrompt := "你是中文语音识别纠错器。只修正明显的语音识别错字、同音字、近音字、热词和标点。保留原意，不补充信息，不解释。不确定就原样返回。只输出修正后的文本。"
	if server.config.LLMHotWords != "" {
		systemPrompt = systemPrompt + "热词：" + server.config.LLMHotWords + "。"
	}
	prompt := fmt.Sprintf("原文：%s", text)
	args := []string{"-m", server.config.LLMModel, "-sys", systemPrompt, "-p", prompt, "-n", "128", "--temp", "0", "--no-display-prompt", "--single-turn", "--log-disable", "--no-show-timings", "--simple-io"}
	command := exec.CommandContext(ctx, server.config.LLMCliPath, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	log.Printf("running corrector: cli=%s model=%s text=%q", server.config.LLMCliPath, server.config.LLMModel, text)
	output, err := command.Output()
	trimmedOutput := strings.TrimSpace(string(output))
	trimmedError := strings.TrimSpace(stderr.String())
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("corrector timed out after %s, cli=%s, model=%s, text=%q", server.config.LLMTimeout, server.config.LLMCliPath, server.config.LLMModel, text)
	}
	if err != nil {
		return "", fmt.Errorf("corrector failed: cli=%s, model=%s, text=%q, error=%v, output=%s, stderr=%s", server.config.LLMCliPath, server.config.LLMModel, text, err, trimmedOutput, trimmedError)
	}

	corrected := extractCorrection(trimmedOutput)
	if !isValidCorrection(text, corrected) {
		log.Printf("corrector rejected output: raw=%q corrected=%q", text, corrected)
		return text, nil
	}
	if corrected != text {
		log.Printf("corrector changed text: raw=%q corrected=%q", text, corrected)
	}
	return corrected, nil
}

func extractCorrection(output string) string {
	lines := strings.Split(output, "\n")
	for index := len(lines) - 1; index >= 0; index-- {
		line := strings.Trim(strings.TrimSpace(lines[index]), "\"'")
		if isLLMNoiseLine(line) {
			continue
		}
		return line
	}
	return ""
}

func isLLMNoiseLine(line string) bool {
	if line == "" {
		return true
	}
	noisePrefixes := []string{"Loading model", "build", "model", "modalities", "available commands", "/exit", "/regen", "/clear", "/read", "/glob", ">", "[ Prompt:"}
	for _, prefix := range noisePrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return line == "Exiting..." || line == "using custom system prompt" || strings.Contains(line, "▄▄") || strings.Contains(line, "██") || strings.Contains(line, "▀▀")
}

func isValidCorrection(raw string, corrected string) bool {
	if corrected == "" {
		return false
	}
	if strings.Contains(corrected, "原文：") || strings.Contains(corrected, "修正：") || strings.Contains(corrected, "解释") {
		return false
	}
	rawLen := len([]rune(raw))
	correctedLen := len([]rune(corrected))
	return correctedLen <= rawLen*3+20
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
