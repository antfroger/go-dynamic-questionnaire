package go_dynamic_questionnaire

import (
	"fmt"
	"os"

	"github.com/expr-lang/expr"
	"github.com/goccy/go-yaml"
)

type (
	Questionnaire struct {
		Questions      []Question `yaml:"questions"`
		askedQuestions map[string]bool
	}

	Question struct {
		Id        string `json:"id"`
		Text      string `json:"text"`
		Condition string `json:"condition,omitempty"`
	}

	config interface {
		string | []byte
	}
)

// New creates a new Questionnaire instance by either
// - reading the configuration from the specified YAML or JSON (TODO) file
// - reading the given YAML or JSON configuration (TODO)
// - using the given configuration (TODO)
func New[T config](config T) (*Questionnaire, error) {
	q := &Questionnaire{
		askedQuestions: make(map[string]bool),
	}
	// TODO: introduce a loader interface to handle different config types
	// these loaders would be responsible for reading from files, parsing YAML/JSON, etc.
	if err := loadConfig(config, q); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return q, nil
}

// loadConfig loads a questionnaire configuration from a file path or YAML content.
func loadConfig[T config](config T, q *Questionnaire) error {
	switch v := any(config).(type) {
	case string:
		return loadYamlFileConfig(v, q)
	case []byte:
		return loadYamlConfig(v, q)
	}

	return fmt.Errorf("unsupported config type: expected string (file path) or []byte (YAML content), got %T", config)
}

// loadYamlFileConfig loads a questionnaire configuration from a YAML file.
func loadYamlFileConfig(configPath string, q *Questionnaire) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %q: %w", configPath, err)
	}

	return loadYamlConfig(data, q)
}

// loadYamlConfig loads a questionnaire configuration from YAML content.
func loadYamlConfig(data []byte, q *Questionnaire) error {
	if err := yaml.Unmarshal(data, q); err != nil {
		return fmt.Errorf("failed to parse questionnaire config: %w", err)
	}
	return nil
}

// Start starts the questionnaire by returning the first batch of questions.
func (q *Questionnaire) Start() []Question {
	var nextQuestions []Question
	for _, question := range q.Questions {
		if question.Condition == "" && !q.askedQuestions[question.Id] {
			q.askedQuestions[question.Id] = true
			nextQuestions = append(nextQuestions, question)
		}
	}

	return nextQuestions
}

// Next returns the next batch of questions in the questionnaire.
func (q *Questionnaire) Next(answers map[string]int) ([]Question, error) {
	var nextQuestions []Question
	for _, question := range q.Questions {
		if q.askedQuestions[question.Id] {
			continue
		}

		show, err := shouldShowQuestion(question, answers)
		if err != nil {
			return nil, fmt.Errorf("failed to show question: %w", err)
		}
		if show {
			q.askedQuestions[question.Id] = true
			nextQuestions = append(nextQuestions, question)
		}
	}
	return nextQuestions, nil
}

func shouldShowQuestion(q Question, answers map[string]int) (bool, error) {
	if q.Condition == "" {
		if len(answers) == 0 {
			return true, nil
		}
		return false, nil
	}

	env := map[string]interface{}{"answers": answers}
	program, err := expr.Compile(q.Condition, expr.Env(env))
	if err != nil {
		return false, fmt.Errorf("failed to compile condition expression: %w", err)
	}
	result, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}
	show, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("condition '%s' does not return a boolean", q.Condition)
	}
	return show, nil
}
