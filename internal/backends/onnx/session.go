package onnx

import (
	"fmt"
	"image"

	ort "github.com/yalue/onnxruntime_go"
)

type Session struct {
	model       Model
	modelPath   string
	runtimePath string
	inputs      []ort.InputOutputInfo
	outputs     []ort.InputOutputInfo
	session     *ort.DynamicAdvancedSession
}

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

	outputs := make([]ort.Value, len(s.outputs))
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
