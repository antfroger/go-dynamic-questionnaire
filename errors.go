package go_dynamic_questionnaire

import (
	"fmt"
)

const (
	emptyQuestionIDErrType     = "empty_question_id"
	duplicateQuestionIDErrType = "duplicate_question_id"
	emptyAnswersErrType        = "empty_answers"
	invalidQuestionIdErrType   = "invalid_question_id"
	invalidAnswerRangeErrType  = "invalid_answer_range"
)

// validationError defines an error happening when the questionnaire structure or the answers given by the users are
// not valid. It implements the error interface.
type validationError struct {
	Type    string
	Message string
	Context map[string]interface{}
}

func (e validationError) Error() string {
	return fmt.Sprintf("validation error (%s): %s", e.Type, e.Message)
}

func emptyQuestionIDError() error {
	return validationError{
		Type:    emptyQuestionIDErrType,
		Message: "questionnaire contains a question with no ID",
	}
}

func duplicateQuestionIDError(questionID string) error {
	return validationError{
		Type:    duplicateQuestionIDErrType,
		Message: "duplicated question ID",
		Context: map[string]interface{}{"question_id": questionID},
	}
}

func emptyAnswersError(questionID string) error {
	return validationError{
		Type:    emptyAnswersErrType,
		Message: "question has no answer options",
		Context: map[string]interface{}{"question_id": questionID},
	}
}

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
