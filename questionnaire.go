package go_dynamic_questionnaire

import (
	"fmt"
	"os"

	"github.com/expr-lang/expr"
	"github.com/goccy/go-yaml"
)

type (
	Questionnaire interface {
		Next(answers map[string]int) (*Response, error)
	}

	// questionnaire represents a dynamic questionnaire that users can answer.
	// It contains a list of questions loaded from the YAML data and can determine which questions to show based on user answers.
	questionnaire struct {
		Questions []question      `yaml:"questions"`
		Remarks   []closingRemark `yaml:"closing_remarks"`
	}

	// question represents a single question in the questionnaire.
	question struct {
		Id        string   `yaml:"id"`
		Text      string   `yaml:"text"`
		Answers   []string `yaml:"answers"`
		Condition string   `yaml:"condition,omitempty"`
	}

	// closingRemark represents a closing remark that can be shown when the questionnaire is completed.
	closingRemark struct {
		Id        string `yaml:"id"`
		Text      string `yaml:"text"`
		Condition string `yaml:"condition,omitempty"`
	}

	// Question represents a question that can be presented to the user.
	Question struct {
		Id      string
		Text    string
		Answers []string
	}

	// ClosingRemark represents a closing remark that can be presented to the user when the questionnaire is completed.
	ClosingRemark struct {
		Id   string
		Text string
	}

	// Response represents the response to a questionnaire, including the next set of questions and any closing remarks.
	Response struct {
		Questions      []Question      `json:"questions"`
		ClosingRemarks []ClosingRemark `json:"closing_remarks,omitempty"`
		Completed      bool            `json:"completed"`
		Progress       *Progress       `json:"progress,omitempty"`
	}

	// Progress represents the progress of the questionnaire, indicating how many questions have been answered and how many are left.
	Progress struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	}

	// config is a generic interface that can be used to pass either a file path (string) or YAML content ([]byte).
	config interface {
		string | []byte
	}
)

// New creates a new Questionnaire instance by either
// - reading the configuration from the specified YAML or JSON (TODO) file
// - reading the given YAML or JSON configuration (TODO)
// - using the given configuration (TODO)
func New[T config](config T) (Questionnaire, error) {
	q := &questionnaire{}
	// TODO: introduce a loader interface to handle different config types
	// these loaders would be responsible for reading from files, parsing YAML/JSON, etc.
	if err := loadConfig(config, q); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return q, nil
}

// loadConfig loads a questionnaire configuration from a file path or YAML content.
func loadConfig[T config](config T, q *questionnaire) error {
	switch v := any(config).(type) {
	case string:
		return loadYamlFileConfig(v, q)
	case []byte:
		return loadYamlConfig(v, q)
	}

	return fmt.Errorf("unsupported config type: expected string (file path) or []byte (YAML content), got %T", config)
}

// loadYamlFileConfig loads a questionnaire configuration from a YAML file.
func loadYamlFileConfig(configPath string, q *questionnaire) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %q: %w", configPath, err)
	}

	return loadYamlConfig(data, q)
}

// loadYamlConfig loads a questionnaire configuration from YAML content.
func loadYamlConfig(data []byte, q *questionnaire) error {
	if err := yaml.Unmarshal(data, q); err != nil {
		return fmt.Errorf("failed to parse questionnaire config: %w", err)
	}
	return nil
}

// Next retrieves the next set of questions based on the provided answers.
// It also calculates the progress, determines if the questionnaire is completed, and if so, retrieves the closing remarks.
func (q *questionnaire) Next(answers map[string]int) (*Response, error) {
	questions, err := q.getNextQuestions(answers)
	if err != nil {
		return nil, fmt.Errorf("failed to get next questions: %w", err)
	}

	completed := len(questions) == 0
	var remarks []ClosingRemark

	if completed {
		remarks, err = q.getClosingRemarks(answers)
		if err != nil {
			return nil, fmt.Errorf("failed to get next closing remarks: %w", err)
		}
	}

	progress := q.calculateProgress(answers, len(questions))

	return &Response{
		Questions:      questions,
		ClosingRemarks: remarks,
		Completed:      completed,
		Progress:       progress,
	}, nil
}

// getNextQuestions retrieves the next set of questions based on the provided answers.
func (q *questionnaire) getNextQuestions(answers map[string]int) ([]Question, error) {
	var nextQuestions []Question

	for _, qu := range q.Questions {
		show, err := shouldShowQuestion(qu, answers)
		if err != nil {
			return nil, fmt.Errorf("failed to show question: %w", err)
		}
		if show {
			nextQuestions = append(nextQuestions, Question{Id: qu.Id, Text: qu.Text, Answers: qu.Answers})
		}
	}

	return nextQuestions, nil
}

// getClosingRemarks retrieves the closing remarks based on the provided answers.
func (q *questionnaire) getClosingRemarks(answers map[string]int) ([]ClosingRemark, error) {
	var remarks []ClosingRemark

	for _, remark := range q.Remarks {
		show, err := shouldShowClosingRemark(remark, answers)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate closing remark condition: %w", err)
		}
		if show {
			remarks = append(remarks, ClosingRemark{Id: remark.Id, Text: remark.Text})
		}
	}

	return remarks, nil
}

// calculateProgress calculates the progress of the questionnaire based on the provided answers and the number of available questions.
func (q *questionnaire) calculateProgress(answers map[string]int, availableQuestions int) *Progress {
	if availableQuestions == 0 {
		return nil
	}

	current := len(answers)
	total := current + availableQuestions

	return &Progress{
		Current: current,
		Total:   total,
	}
}

// shouldShowQuestion determines if a question should be shown based on its condition and the provided answers.
func shouldShowQuestion(q question, answers map[string]int) (bool, error) {
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
func isQuestionAnswered(question question, answers map[string]int) bool {
	_, exists := answers[question.Id]
	return exists
}

// shouldShowClosingRemark determines if a closing remark should be shown based on its condition and the provided answers.
func shouldShowClosingRemark(remark closingRemark, answers map[string]int) (bool, error) {
	if remark.Condition == "" {
		return true, nil
	}

	env := map[string]interface{}{
		"answers": answers,
	}

	program, err := expr.Compile(remark.Condition, expr.Env(env))
	if err != nil {
		return false, fmt.Errorf("failed to compile condition expression: %w", err)
	}
	result, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}
	show, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("condition '%s' does not return a boolean", remark.Condition)
	}
	return show, nil
}
