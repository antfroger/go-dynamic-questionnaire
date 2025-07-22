package go_dynamic_questionnaire

import (
	"fmt"
	"os"

	"github.com/expr-lang/expr"
	"github.com/goccy/go-yaml"
)

type (
	Questionnaire struct {
		Questions   []Question `yaml:"questions"`
		isCompleted bool
	}

	Question struct {
		Id        string
		Text      string
		Answers   []string
		Condition string
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
	q := &Questionnaire{}
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
		if question.Condition == "" {
			nextQuestions = append(nextQuestions, question)
		}
	}

	return nextQuestions
}

// Next returns the next batch of questions in the questionnaire.
func (q *Questionnaire) Next(answers map[string]int) ([]Question, error) {
	var nextQuestions []Question

	for _, question := range q.Questions {
		show, err := shouldShowQuestion(question, answers)
		if err != nil {
			return nil, fmt.Errorf("failed to show question: %w", err)
		}
		if show {
			nextQuestions = append(nextQuestions, question)
		}
	}
	if len(nextQuestions) == 0 {
		q.isCompleted = true
	}

	return nextQuestions, nil
}

// shouldShowQuestion determines if a question should be shown based on its condition and the provided answers.
func shouldShowQuestion(q Question, answers map[string]int) (bool, error) {
	if isQuestionAnswered(q, answers) {
		return false, nil
	}

	if q.Condition == "" {
		if len(answers) == 0 {
			return true, nil
		}
		return false, nil
	}

	env := map[string]interface{}{
		"answers": answers,
	}

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

// isQuestionAnswered checks if a question has been answered based on the provided answers map.
func isQuestionAnswered(question Question, answers map[string]int) bool {
	_, exists := answers[question.Id]
	return exists
}

// Completed returns true if the questionnaire has been completed, false otherwise.
func (q *Questionnaire) Completed() bool {
	return q.isCompleted
}
