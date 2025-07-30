package go_dynamic_questionnaire

import (
	"fmt"
)

// Error type constants for consistent error identification.
// These can be used programmatically to handle specific error types.
const (
	// emptyQuestionIDErrType indicates a question was defined without an ID.
	// This violates the questionnaire structure requirements.
	emptyQuestionIDErrType = "empty_question_id"

	// duplicateQuestionIDErrType indicates multiple questions share the same ID.
	// Question IDs must be unique within a questionnaire.
	duplicateQuestionIDErrType = "duplicate_question_id"

	// emptyAnswersErrType indicates a question has no answer options.
	// All questions must have at least one possible answer.
	emptyAnswersErrType = "empty_answers"

	// invalidQuestionIdErrType indicates an answer was provided for a non-existent question.
	// All answer keys must correspond to valid question IDs.
	invalidQuestionIdErrType = "invalid_question_id"

	// invalidAnswerRangeErrType indicates an answer value is outside the valid range.
	// Answer values must be between 1 and the number of available answers for that question.
	invalidAnswerRangeErrType = "invalid_answer_range"

	// invalidDependencyErrType indicates a question depends on a non-existent question.
	// All question IDs in depends_on must correspond to valid questions.
	invalidDependencyErrType = "invalid_dependency"

	// circularDependencyErrType indicates circular dependencies exist in the questionnaire.
	// Questions cannot depend on themselves directly or indirectly.
	circularDependencyErrType = "circular_dependency"

	// conditionDependencyMismatchErrType indicates condition references don't match depends_on.
	// Questions should declare dependencies for all question IDs used in conditions.
	conditionDependencyMismatchErrType = "condition_dependency_mismatch"
)

// validationError represents an error that occurs during questionnaire validation.
// It provides structured information about validation failures, including
// error type, descriptive message, and contextual data for debugging.
//
// This error type implements the standard error interface and can be used
// for both user-facing error messages and programmatic error handling.
//
// Example usage:
//
//	var validationErr validationError
//	if errors.As(err, &validationErr) {
//	    switch validationErr.Type {
//	    case invalidQuestionIdErrType:
//	        // Handle invalid question ID specifically
//	    case invalidAnswerRangeErrType:
//	        // Handle out-of-range answer specifically
//	    }
//	}
type validationError struct {
	Type    string                 // Error type identifier (see constants above)
	Message string                 // Human-readable error description
	Context map[string]interface{} // Additional context data for debugging
}

// Error returns a formatted error message that includes both the error type and message.
// This implements the standard error interface.
//
// Format: "validation error (error_type): error_message"
//
// Example output:
//
//	"validation error (invalid_question_id): question 'xyz' does not exist"
func (e validationError) Error() string {
	return fmt.Sprintf("validation error (%s): %s", e.Type, e.Message)
}

// emptyQuestionIDError creates a validation error for questions missing an ID.
// This error occurs during questionnaire loading when a question is defined
// without a required ID field.
//
// Returns:
//
//	error: A validationError with type emptyQuestionIDErrType.
//
// Example scenario:
//
//	questions:
//	  - text: "What's your favorite color?"  # Missing 'id' field
//	    answers: ["Red", "Blue", "Green"]
func emptyQuestionIDError() error {
	return validationError{
		Type:    emptyQuestionIDErrType,
		Message: "questionnaire contains a question with no ID",
	}
}

// duplicateQuestionIDError creates a validation error for duplicate question IDs.
// This error occurs during questionnaire loading when multiple questions
// share the same ID, which would cause conflicts in answer processing.
//
// Parameters:
//
//	questionID: The ID that appears multiple times in the questionnaire.
//
// Returns:
//
//	error: A validationError with type duplicateQuestionIDErrType and
//	       context containing the conflicting question ID.
//
// Example scenario:
//
//	questions:
//	  - id: "q1"
//	    text: "First question"
//	    answers: ["Yes", "No"]
//	  - id: "q1"  # Duplicate ID
//	    text: "Another question"
//	    answers: ["A", "B", "C"]
func duplicateQuestionIDError(questionID string) error {
	return validationError{
		Type:    duplicateQuestionIDErrType,
		Message: "duplicated question ID",
		Context: map[string]interface{}{"question_id": questionID},
	}
}

// emptyAnswersError creates a validation error for questions with no answer options.
// This error occurs during questionnaire loading when a question is defined
// without any possible answers, making it impossible for users to respond.
//
// Parameters:
//
//	questionID: The ID of the question that lacks answer options.
//
// Returns:
//
//	error: A validationError with type emptyAnswersErrType and
//	       context containing the affected question ID.
//
// Example scenario:
//
//	questions:
//	  - id: "broken_question"
//	    text: "What do you think?"
//	    answers: []  # Empty answers array
func emptyAnswersError(questionID string) error {
	return validationError{
		Type:    emptyAnswersErrType,
		Message: "question has no answer options",
		Context: map[string]interface{}{"question_id": questionID},
	}
}

// invalidQuestionIDError creates a validation error for non-existent question references.
// This error occurs during answer processing when a user provides an answer
// for a question ID that doesn't exist in the questionnaire.
//
// Parameters:
//
//	questionID: The invalid question ID that was referenced.
//	answer: The answer value that was provided (included for context).
//
// Returns:
//
//	error: A validationError with type invalidQuestionIdErrType and
//	       context containing both the invalid ID and the attempted answer.
//
// Example scenario:
//
//	// Questionnaire has questions "q1", "q2", "q3"
//	answers := map[string]int{
//	    "q1": 1,
//	    "q4": 2,  // "q4" doesn't exist
//	}
func invalidQuestionIDError(questionID string, answer int) error {
	return validationError{
		Type:    invalidQuestionIdErrType,
		Message: "question does not exist",
		Context: map[string]interface{}{
			"question_id": questionID,
			"answer":      answer,
		},
	}
}

// invalidAnswerRangeError creates a validation error for out-of-range answer values.
// This error occurs during answer processing when a user provides an answer
// that is outside the valid range for a specific question.
//
// Answer values must be 1-indexed and within the range [1, number_of_answers].
// For example, if a question has 3 answer choices, valid values are 1, 2, or 3.
//
// Parameters:
//
//	q: The question for which an invalid answer was provided.
//	answer: The out-of-range answer value that was provided.
//
// Returns:
//
//	error: A validationError with type invalidAnswerRangeErrType and
//	       comprehensive context including question details and valid range.
//
// Example scenario:
//
//	question:
//	  id: "color"
//	  text: "What's your favorite color?"
//	  answers: ["Red", "Blue", "Green"]  # Valid answers: 1, 2, 3
//
//	// User provides answer 5 (out of range)
//	answers := map[string]int{"color": 5}  # Error: valid range is 1-3
func invalidAnswerRangeError(q *question, answer int) error {
	return validationError{
		Type:    invalidAnswerRangeErrType,
		Message: "answer is out of range",
		Context: map[string]interface{}{
			"question_id":   q.Id,
			"question_text": q.Text,
			"answer":        answer,
			"valid_range":   fmt.Sprintf("1-%d", len(q.Answers)),
		},
	}
}

// invalidDependencyError creates a validation error for invalid question dependencies.
// This error occurs during questionnaire loading when a question declares a dependency
// on a question ID that doesn't exist in the questionnaire.
//
// Parameters:
//
//	questionID: The ID of the question that has the invalid dependency.
//	invalidDependencyID: The non-existent question ID that was referenced.
//
// Returns:
//
//	error: A validationError with type invalidDependencyErrType and
//	       context containing both question IDs.
//
// Example scenario:
//
//	questions:
//	  - id: "q1"
//	    text: "First question"
//	    answers: ["Yes", "No"]
//	  - id: "q2"
//	    text: "Second question"
//	    answers: ["A", "B", "C"]
//	    depends_on: ["nonexistent"]  # "nonexistent" doesn't exist
func invalidDependencyError(questionID, invalidDependencyID string) error {
	return validationError{
		Type:    invalidDependencyErrType,
		Message: fmt.Sprintf("question '%s' depends on non-existent question '%s'", questionID, invalidDependencyID),
		Context: map[string]interface{}{
			"question_id":           questionID,
			"invalid_dependency_id": invalidDependencyID,
		},
	}
}

// circularDependencyError creates a validation error for circular dependencies.
// This error occurs during questionnaire loading when questions have circular
// dependency relationships, which would create infinite loops.
//
// Parameters:
//
//	cycle: A slice of question IDs representing the circular dependency path.
//	       The last element should be the same as an earlier element to show the cycle.
//
// Returns:
//
//	error: A validationError with type circularDependencyErrType and
//	       context containing the full dependency cycle.
//
// Example scenario:
//
//	questions:
//	  - id: "q1"
//	    depends_on: ["q2"]
//	  - id: "q2"
//	    depends_on: ["q3"]
//	  - id: "q3"
//	    depends_on: ["q1"]  # Creates cycle: q1 -> q2 -> q3 -> q1
func circularDependencyError(cycle []string) error {
	cycleStr := fmt.Sprintf("%s", cycle[0])
	for i := 1; i < len(cycle); i++ {
		cycleStr += " -> " + cycle[i]
	}

	return validationError{
		Type:    circularDependencyErrType,
		Message: fmt.Sprintf("circular dependency detected: %s", cycleStr),
		Context: map[string]interface{}{
			"cycle":        cycle,
			"cycle_length": len(cycle) - 1, // -1 because last element repeats the first
		},
	}
}

// conditionDependencyMismatchError creates a validation error when question conditions
// reference question IDs that are not declared in the depends_on field.
// This error helps ensure consistency between explicit dependencies and condition logic.
//
// Parameters:
//
//	questionID: The ID of the question with the mismatched condition.
//	declared: The dependencies declared in depends_on.
//	condition: The dependencies referenced in condition.
//
// Returns:
//
//	error: A validationError with type conditionDependencyMismatchErrType and
//	       context containing both question IDs.
//
// Example scenario:
//
//	questions:
//	  - id: "q1"
//	    text: "First question"
//	    answers: ["Yes", "No"]
//	  - id: "q2"
//	    text: "Second question"
//	    answers: ["A", "B"]
//	    depends_on: ["q1"]
//	    condition: 'answers["q1"] == 1 && answers["q3"] == 2'  # References q3 but doesn't declare it
func conditionDependencyMismatchError(questionID string, declared, condition []string) error {
	return validationError{
		Type:    conditionDependencyMismatchErrType,
		Message: fmt.Sprintf("question '%s' conditions don't match the declared dependencies %v", questionID, declared),
		Context: map[string]interface{}{
			"question_id":  questionID,
			"condition":    condition,
			"dependencies": declared,
		},
	}
}
