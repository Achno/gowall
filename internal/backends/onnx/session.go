package onnx

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Achno/gowall/internal/logger"
	"github.com/Achno/gowall/utils"
	ort "github.com/yalue/onnxruntime_go"
)

type Session struct {
	model       Model
	modelPath   string
	runtimePath string
	inputs      []ort.InputOutputInfo
	outputs     []ort.InputOutputInfo
	inputNames  []string
	outputNames []string
	session     *ort.DynamicAdvancedSession
}

var (
	envMu          sync.Mutex
	envInitialized bool
	envLibraryPath string
)

func NewSession(model Model) (*Session, error) {
	if model == nil {
		return nil, fmt.Errorf("model is required")
	}

	runtimePath, err := ensureRuntimeAvailable()
	if err != nil {
		return nil, err
	}

	if err := ensureEnvironment(runtimePath); err != nil {
		return nil, err
	}

	modelPath, err := ensureModelAvailable(model)
	if err != nil {
		return nil, err
	}

	inputs, outputs, err := ort.GetInputOutputInfo(modelPath)
	if err != nil {
		return nil, fmt.Errorf("inspect ONNX model I/O: %w", err)
	}
	if len(inputs) == 0 {
		return nil, fmt.Errorf("model %q has no inputs", model.Name())
	}
	if len(outputs) == 0 {
		return nil, fmt.Errorf("model %q has no outputs", model.Name())
	}

	inputNames := namesFromIO(inputs)
	outputNames := namesFromIO(outputs)

	dynamicSession, err := ort.NewDynamicAdvancedSession(modelPath, inputNames, outputNames, nil)
	if err != nil {
		return nil, fmt.Errorf("create ONNX session: %w", err)
	}

	return &Session{
		model:       model,
		modelPath:   modelPath,
		runtimePath: runtimePath,
		inputs:      cloneIOInfo(inputs),
		outputs:     cloneIOInfo(outputs),
		inputNames:  append([]string(nil), inputNames...),
		outputNames: append([]string(nil), outputNames...),
		session:     dynamicSession,
	}, nil
}

func (s *Session) Close() error {
	if s == nil || s.session == nil {
		return nil
	}

	err := s.session.Destroy()
	s.session = nil
	return err
}

func (s *Session) Predict(img image.Image) ([]image.Image, error) {
	if s == nil || s.model == nil {
		return nil, fmt.Errorf("session model is not initialized")
	}

	inputs, err := s.model.Normalize(img)
	if err != nil {
		return nil, fmt.Errorf("normalize image: %w", err)
	}
	defer destroyValues(inputs)

	outputs := make([]ort.Value, len(s.outputNames))
	if err := s.Run(inputs, outputs); err != nil {
		destroyValues(outputs)
		return nil, fmt.Errorf("run session: %w", err)
	}
	defer destroyValues(outputs)

	masks, err := s.model.Predict(img, outputs)
	if err != nil {
		return nil, fmt.Errorf("decode outputs: %w", err)
	}

	return masks, nil
}

func (s *Session) Run(inputs, outputs []ort.Value) error {
	if s == nil || s.session == nil {
		return fmt.Errorf("session is not initialized")
	}

	return s.session.Run(inputs, outputs)
}

func (s *Session) InputInfo() []ort.InputOutputInfo {
	return cloneIOInfo(s.inputs)
}

func (s *Session) OutputInfo() []ort.InputOutputInfo {
	return cloneIOInfo(s.outputs)
}

func (s *Session) InputNames() []string {
	return append([]string(nil), s.inputNames...)
}

func (s *Session) OutputNames() []string {
	return append([]string(nil), s.outputNames...)
}

func (s *Session) ModelPath() string {
	return s.modelPath
}

func (s *Session) RuntimePath() string {
	return s.runtimePath
}

func Shutdown() error {
	envMu.Lock()
	defer envMu.Unlock()

	if !envInitialized {
		return nil
	}

	if err := ort.DestroyEnvironment(); err != nil {
		return err
	}

	envInitialized = false
	envLibraryPath = ""
	return nil
}

func ensureRuntimeAvailable() (string, error) {
	runtimePath, err := CheckOnnxRuntimeInstalled()
	if err == nil {
		return runtimePath, nil
	}

	prompt := fmt.Sprintf("%s ◈ ONNX Runtime is not installed. Would you like to set it up?%s", utils.BlueColor, utils.ResetColor)
	if !utils.Confirm(prompt) {
		return "", fmt.Errorf("onnx runtime download declined")
	}

	if err := SetupOnnxRuntime(); err != nil {
		return "", err
	}

	return CheckOnnxRuntimeInstalled()
}

func ensureEnvironment(libraryPath string) error {
	envMu.Lock()
	defer envMu.Unlock()

	if envInitialized {
		if envLibraryPath != libraryPath {
			return fmt.Errorf("onnx runtime already initialized with %q", envLibraryPath)
		}
		return nil
	}

	ort.SetSharedLibraryPath(libraryPath)
	if err := ort.InitializeEnvironment(); err != nil {
		return fmt.Errorf("initialize onnxruntime: %w", err)
	}

	envInitialized = true
	envLibraryPath = libraryPath
	return nil
}

func ensureModelAvailable(model Model) (string, error) {
	baseDir := OnnxRuntimeFolder()
	modelPath := modelCachePath(baseDir, model)

	if _, err := os.Stat(modelPath); err == nil {
		return modelPath, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("check model path %q: %w", modelPath, err)
	}

	prompt := fmt.Sprintf("%s ◈ Model %q is missing. Download it?%s", utils.BlueColor, model.Name(), utils.ResetColor)
	if !utils.Confirm(prompt) {
		return "", fmt.Errorf("model download declined for %q", model.Name())
	}

	// Get file size for progress info
	size := utils.GetRemoteFileSize(model.DownloadURL())
	sizeInfo := ""
	if size != "" {
		sizeInfo = fmt.Sprintf(" size: %s,", size)
	}

	logger.Print(fmt.Sprintf("%s ➜ Downloading %s,%s sit back and relax%s", utils.BlueColor, model.Name(), sizeInfo, utils.ResetColor))

	if err := os.MkdirAll(filepath.Dir(modelPath), 0o755); err != nil {
		return "", fmt.Errorf("create model directory: %w", err)
	}

	if err := model.Download(DownloadOptions{DestPath: modelPath}); err != nil {
		return "", fmt.Errorf("download model %q: %w", model.Name(), err)
	}

	if _, err := os.Stat(modelPath); err != nil {
		return "", fmt.Errorf("model %q was not written to %s: %w", model.Name(), modelPath, err)
	}

	logger.Print(fmt.Sprintf("%s ➜ Process complete. Model %s downloaded%s", utils.BlueColor, model.Name(), utils.ResetColor))

	return modelPath, nil
}

// modelCachePath returns ~/.u2net/<model>.onnx
func modelCachePath(baseDir string, model Model) string {
	name := strings.TrimSpace(model.Name())
	fileName := name
	if filepath.Ext(fileName) == "" {
		fileName += ".onnx"
	}
	return filepath.Join(baseDir, fileName)
}

func namesFromIO(infos []ort.InputOutputInfo) []string {
	names := make([]string, 0, len(infos))
	for _, info := range infos {
		names = append(names, info.Name)
	}
	return names
}

func cloneIOInfo(infos []ort.InputOutputInfo) []ort.InputOutputInfo {
	if len(infos) == 0 {
		return nil
	}

	cloned := make([]ort.InputOutputInfo, len(infos))
	copy(cloned, infos)
	return cloned
}

func destroyValues(values []ort.Value) {
	for _, value := range values {
		if value == nil {
			continue
		}
		value.Destroy()
	}
}
