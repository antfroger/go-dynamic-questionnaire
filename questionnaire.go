/*
Package go_dynamic_questionnaire provides a flexible, condition-based questionnaire system
for Go applications.

This package allows you to create dynamic questionnaires where the questions shown to users
depend on their previous answers. Questions can have conditional logic using expressions,
and the system supports progress tracking and closing remarks.

# Basic Usage

See README.md for a quick start guide.
For more examples and advanced usage, see the examples/ directory.

# Thread Safety

All operations are thread-safe. The same questionnaire instance can be used
concurrently by multiple goroutines without any synchronization.

# Error Handling

The package provides rich error information through custom error types.
Validation errors include context about what went wrong and suggestions
for fixing the issue.
*/
package go_dynamic_questionnaire

import (
	"fmt"
	"os"

	"github.com/expr-lang/expr"
	"github.com/goccy/go-yaml"
)

type (
	// Questionnaire represents a dynamic questionnaire that can process user answers
	// and determine the next questions to show based on conditional logic.
	//
	// All methods are thread-safe and can be called concurrently from multiple goroutines.
	// The questionnaire instance is stateless - all state is passed through the answers parameter.
	//
	// Example usage:
	//   q, err := gdq.New("questionnaire.yaml")
	//   if err != nil {
	//       return err
	//   }
	//
	//   answers := map[string]int{"q1": 2, "q2": 1}
	//   response, err := q.Next(answers)
	//   if err != nil {
	//       return err
	//   }
	//
	//   if response.Completed {
	//       // Handle completion and closing remarks
	//   } else {
	//       // Present next questions to user
	//   }
	Questionnaire interface {
		// Next processes the provided answers and returns the next set of questions,
		// progress information, and completion status.
		//
		// Parameters:
		//   answers: A map where keys are question IDs and values are 1-indexed answer choices.
		//            For example, if a question has answers ["Yes", "No", "Maybe"],
		//            a value of 1 means "Yes", 2 means "No", and 3 means "Maybe".
		//
		// Returns:
		//   *Response: Contains the next questions to show, completion status,
		//             progress information, and closing remarks (if completed).
		//   error: Returns validation errors for invalid question IDs, out-of-range answers,
		//          or condition evaluation errors.
		//
		// The method validates all provided answers before processing. If any answer
		// is invalid, the entire operation fails and returns a validation error with
		// details about what went wrong.
		Next(answers map[string]int) (*Response, error)
	}

	// config is a constraint interface for configuration inputs to the New function.
	// It accepts either a file path (string) or raw YAML content ([]byte).
	//
	// Examples:
	//   New("path/to/questionnaire.yaml")  // Load from file
	//   New([]byte("questions: ..."))      // Load from YAML content
	config interface {
		string | []byte
	}

	// questionnaire is the internal implementation of the Questionnaire interface.
	// It contains the parsed questions and closing remarks from YAML configuration.
	//
	// This struct is not exported as users should interact with the Questionnaire interface.
	// Instances are created through the New function and are immutable after creation.
	questionnaire struct {
		Questions []question      `yaml:"questions"`       // List of all questions in the questionnaire
		Remarks   []closingRemark `yaml:"closing_remarks"` // List of all closing remarks
	}

	// question represents a single question in the questionnaire configuration.
	// Questions can have conditional logic that determines when they should be shown.
	question struct {
		Id        string   `yaml:"id"`                  // Unique identifier for the question
		Text      string   `yaml:"text"`                // The question text shown to users
		Answers   []string `yaml:"answers"`             // List of possible answer choices
		Condition string   `yaml:"condition,omitempty"` // Optional expression to determine if question should be shown
	}

	// closingRemark represents a message shown when the questionnaire is completed.
	// Like questions, closing remarks can have conditional logic.
	closingRemark struct {
		Id        string `yaml:"id"`                  // Unique identifier for the remark
		Text      string `yaml:"text"`                // The remark text shown to users
		Condition string `yaml:"condition,omitempty"` // Optional expression to determine if remark should be shown
	}

	// Response represents the complete response from processing a questionnaire step.
	// It contains all information needed to either continue the questionnaire or handle completion.
	//
	// The response structure allows clients to:
	//   - Display the next questions to users (Questions field)
	//   - Show progress information (Progress field)
	//   - Handle questionnaire completion (Completed field)
	//   - Display closing messages (ClosingRemarks field)
	//
	// Example JSON output:
	//   {
	//     "questions": [{"id": "q1", "text": "...", "answers": ["Yes", "No"]}],
	//     "completed": false,
	//     "progress": {"current": 2, "total": 5}
	//   }
	Response struct {
		Questions      []Question      `json:"questions"`                 // Next questions to show (empty if completed)
		ClosingRemarks []ClosingRemark `json:"closing_remarks,omitempty"` // Closing remarks (only when completed)
		Completed      bool            `json:"completed"`                 // Whether the questionnaire is finished
		Progress       *Progress       `json:"progress,omitempty"`        // Progress information (nil when completed)
	}

	// Question represents a question that should be presented to the user.
	// This is the external representation used in API responses.
	Question struct {
		Id      string   `json:"id"`      // Unique identifier for the question
		Text    string   `json:"text"`    // The question text to display
		Answers []string `json:"answers"` // List of answer choices (1-indexed when referenced)
	}

	// ClosingRemark represents a message shown to users when the questionnaire is completed.
	// Multiple remarks can be shown based on the user's answers and conditional logic.
	//
	// Example usage in JSON:
	//   {
	//     "id": "thank_you",
	//     "text": "Thank you for your feedback!"
	//   }
	ClosingRemark struct {
		Id   string `json:"id"`   // Unique identifier for the remark
		Text string `json:"text"` // The message text to display
	}

	// Progress represents the user's progress through the questionnaire.
	// It provides information about how many questions have been processed
	// and how many remain, useful for displaying progress bars or indicators.
	//
	// The progress calculation is based on available questions at each step,
	// taking into account conditional logic. Questions that cannot be reached
	// due to user answers are not counted in the total.
	//
	// Note: Progress is nil when the questionnaire is completed.
	Progress struct {
		Current int `json:"current"` // Number of questions answered so far
		Total   int `json:"total"`   // Total number of questions that could be answered
	}
)

// New creates a new Questionnaire instance from either a file path or YAML content.
//
// The function accepts two types of input:
//   - string: Path to a YAML configuration file
//   - []byte: Raw YAML content as bytes
//
// Parameters:
//
//	config: Either a file path (string) or YAML content ([]byte).
//	        The YAML must contain 'questions' and optionally 'closing_remarks' sections.
//
// Returns:
//
//	Questionnaire: A fully configured questionnaire instance ready for use.
//	               The instance is immutable and thread-safe.
//	error: Returns configuration errors, file reading errors, YAML parsing errors,
//	       or validation errors if the questionnaire structure is invalid.
//
// Example usage with file path:
//
//	q, err := gdq.New("surveys/customer-feedback.yaml")
//	if err != nil {
//	    log.Fatalf("Failed to load questionnaire: %v", err)
//	}
//
// Example usage with YAML content:
//
//	yamlData := []byte(`
//	questions:
//	  - id: "satisfaction"
//	    text: "How satisfied are you with our service?"
//	    answers: ["Very satisfied", "Satisfied", "Neutral", "Dissatisfied"]
//	  - id: "recommend"
//	    text: "Would you recommend us?"
//	    answers: ["Yes", "No"]
//	    condition: 'answers["satisfaction"] <= 2'
//	closing_remarks:
//	  - id: "thanks"
//	    text: "Thank you for your feedback!"
//	`)
//
//	q, err := gdq.New(yamlData)
//	if err != nil {
//	    return fmt.Errorf("questionnaire creation failed: %w", err)
//	}
//
// The function validates the questionnaire structure during creation, checking for:
//   - Duplicate question IDs
//   - Empty question IDs
//   - Questions without answer options
//   - Invalid YAML syntax
func New[T config](config T) (Questionnaire, error) {
	q := &questionnaire{}
	if err := loadConfig(config, q); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	if err := q.validateQuestionnaireIntegrity(); err != nil {
		return nil, fmt.Errorf("questionnaire validation failed: %w", err)
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

// validateQuestionnaireIntegrity validates the questionnaire configuration at load time
func (q *questionnaire) validateQuestionnaireIntegrity() error {
	questionIDs := make(map[string]bool)

	for _, question := range q.Questions {
		if question.Id == "" {
			return emptyQuestionIDError()
		}
		if questionIDs[question.Id] {
			return duplicateQuestionIDError(question.Id)
		}
		if len(question.Answers) == 0 {
			return emptyAnswersError(question.Id)
		}
		questionIDs[question.Id] = true
	}

	return nil
}

// Next processes user answers and returns the next step in the questionnaire flow.
//
// This is the main method for progressing through a questionnaire. It evaluates
// all provided answers, determines which questions should be shown next based on
// conditional logic, calculates progress, and handles questionnaire completion.
//
// Parameters:
//
//	answers: Map of question ID to answer choice (1-indexed).
//	         Keys must be valid question IDs from the questionnaire.
//	         Values must be in the range [1, number_of_answers] for each question.
//
//	         Example: map[string]int{
//	             "satisfaction": 2,    // Second answer choice
//	             "recommend": 1,       // First answer choice
//	             "category": 3,        // Third answer choice
//	         }
//
// Returns:
//
//	*Response: Complete response containing:
//	           - Questions: Next questions to show (empty if completed)
//	           - Completed: Whether questionnaire is finished
//	           - Progress: Current progress (nil when completed)
//	           - ClosingRemarks: Final messages (only when completed)
//	error: Validation errors for invalid inputs, or condition evaluation errors.
//
// Behavior:
//   - Validates all answers before processing
//   - Evaluates question conditions to determine visibility
//   - Filters out already-answered questions
//   - Calculates progress based on reachable questions
//   - Returns closing remarks only when questionnaire is complete
//   - Thread-safe: can be called concurrently
//
// Example usage:
//
//	// Start questionnaire (no answers yet)
//	response, err := q.Next(map[string]int{})
//	if err != nil {
//	    return err
//	}
//
//	// Show initial questions to user...
//
//	// Process user answers
//	answers := map[string]int{"q1": 2, "q2": 1}
//	response, err = q.Next(answers)
//	if err != nil {
//	    return err
//	}
//
//	if response.Completed {
//	    // Display closing remarks
//	    for _, remark := range response.ClosingRemarks {
//	        fmt.Println(remark.Text)
//	    }
//	} else {
//	    // Show next questions
//	    fmt.Printf("Progress: %d/%d\n", response.Progress.Current, response.Progress.Total)
//	    for _, question := range response.Questions {
//	        fmt.Printf("%s\n", question.Text)
//	    }
//	}
//
// Common errors:
//   - Invalid question ID: "question 'xyz' does not exist"
//   - Out-of-range answer: "answer 5 is out of range for question 'q1' (valid: 1-3)"
//   - Condition evaluation error: "failed to evaluate condition for question 'q2'"
func (q *questionnaire) Next(answers map[string]int) (*Response, error) {
	if err := q.validateAnswers(answers); err != nil {
		return nil, fmt.Errorf("invalid answers provided: %w", err)
	}

	questions, err := q.getNextQuestions(answers)
	if err != nil {
		return nil, fmt.Errorf("failed to get next questions: %w", err)
	}

	completed := len(questions) == 0
	var remarks []ClosingRemark

	if completed {
		remarks, err = q.getClosingRemarks(answers)
		if err != nil {
			return nil, fmt.Errorf("failed to get closing remarks: %w", err)
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

// validateAnswers performs comprehensive validation on the provided answers
func (q *questionnaire) validateAnswers(answers map[string]int) error {
	for questionID, answer := range answers {
		if err := q.validateSingleAnswer(questionID, answer); err != nil {
			return err
		}
	}
	return nil
}

// validateSingleAnswer validates a single answer for a specific question
func (q *questionnaire) validateSingleAnswer(questionID string, answer int) error {
	question := q.findQuestionByID(questionID)
	if question == nil {
		return invalidQuestionIDError(questionID, answer)
	}

	if answer < 1 || answer > len(question.Answers) {
		return invalidAnswerRangeError(question, answer)
	}

	return nil
}

// findQuestionByID finds a question by its ID
func (q *questionnaire) findQuestionByID(id string) *question {
	for i := range q.Questions {
		if q.Questions[i].Id == id {
			return &q.Questions[i]
		}
	}
	return nil
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
