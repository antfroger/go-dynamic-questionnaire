package go_dynamic_questionnaire_test

import (
	"errors"
	"os"

	gdq "github.com/antfroger/go-dynamic-questionnaire"
	"github.com/goccy/go-yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Questionnaire", func() {
	Describe("New", func() {
		When("config is a file", func() {
			When("the given file does not exist", func() {
				It("returns an error", func() {
					_, err := gdq.New("testdata/missing.yaml")
					Expect(err).To(MatchError(ContainSubstring(`failed to read config file "testdata/missing.yaml"`)))
					Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
				})
			})

			When("the given file exists", func() {
				It("should load a questionnaire from the file", func() {
					content := []byte(`
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"
`)
					tmpFile, err := os.CreateTemp("", "questionnaire-*.yaml")
					Expect(err).To(BeNil())
					defer func(name string) {
						_ = os.Remove(name)
					}(tmpFile.Name())

					_, err = tmpFile.Write(content)
					Expect(err).To(BeNil())
					err = tmpFile.Close()
					Expect(err).To(BeNil())

					q, err := gdq.New(tmpFile.Name())
					Expect(err).To(BeNil())
					Expect(q).NotTo(BeNil())
				})
			})
		})

		When("config is yaml content", func() {
			It("should load the questionnaire from bytes", func() {
				q, err := gdq.New([]byte(`
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"
`))
				Expect(err).To(BeNil())
				Expect(q).NotTo(BeNil())
			})

			It("should handle invalid YAML content", func() {
				_, err := gdq.New([]byte(`invalid yaml`))
				Expect(err).To(MatchError(ContainSubstring(`failed to parse questionnaire config`)))
				var yamlErr *yaml.UnexpectedNodeTypeError
				Expect(errors.As(err, &yamlErr)).To(BeTrue())
			})
		})
	})

	Describe("Next", func() {
		var (
			config string
			q      gdq.Questionnaire
			err    error
		)
		JustBeforeEach(func() {
			q, err = gdq.New([]byte(config))
			Expect(err).ToNot(HaveOccurred())
		})

		When("no answers are provided", func() {
			BeforeEach(func() {
				config = `
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Yes"
      - "No"
  - id: "q2"
    text: "Question 2?"
    answers:
      - "Yes"
      - "No"
  - id: "q3"
    text: "Question 3?"
    answers:
      - "Yes"
      - "No"
    condition: "answers['q1'] == 1"`
			})

			It("should return the first batch of questions", func() {
				r, err := q.Next(map[string]int{})

				Expect(err).ToNot(HaveOccurred())
				Expect(r.Questions).To(Equal([]gdq.Question{
					{Id: "q1", Text: "Question 1?", Answers: []string{"Yes", "No"}},
					{Id: "q2", Text: "Question 2?", Answers: []string{"Yes", "No"}},
				}))
				Expect(r.Completed).To(BeFalse())
				Expect(r.ClosingRemarks).To(BeEmpty())
				Expect(r.Progress).To(Equal(&gdq.Progress{Current: 0, Total: 2}))
			})
		})

		When("the questionnaire is empty", func() {
			BeforeEach(func() {
				config = ``
			})

			It("should return completed with no questions", func() {
				r, err := q.Next(map[string]int{})

				Expect(err).ToNot(HaveOccurred())
				Expect(r.Questions).To(BeEmpty())
				Expect(r.Completed).To(BeTrue())
				Expect(r.Progress).To(BeNil())
			})
		})

		When("the questionnaire has no question without conditions", func() {
			BeforeEach(func() {
				config = `
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"
    condition: "false"`
			})

			It("should return completed with no questions", func() {
				r, err := q.Next(map[string]int{})

				Expect(err).ToNot(HaveOccurred())
				Expect(r.Questions).To(BeEmpty())
				Expect(r.Completed).To(BeTrue())
			})
		})

		When("one or more questions have been answered", func() {
			BeforeEach(func() {
				config = `
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"
  - id: "q2"
    text: "Question 2?"
    answers:
      - "Answer 1"
      - "Answer 2"
      - "Answer 3"
    condition: 'answers["q1"] == 1'
  - id: "q3"
    text: "Question 3?"
    answers:
      - "Answer 1"
      - "Answer 2"
      - "Answer 3"
    condition: 'answers["q1"] == 1'`
			})

			It("should return the next questions depending on the answer", func() {
				r, err := q.Next(map[string]int{"q1": 1})
				Expect(err).ToNot(HaveOccurred())
				Expect(r.Questions).To(Equal([]gdq.Question{
					{Id: "q2", Text: "Question 2?", Answers: []string{"Answer 1", "Answer 2", "Answer 3"}},
					{Id: "q3", Text: "Question 3?", Answers: []string{"Answer 1", "Answer 2", "Answer 3"}},
				}))
				Expect(r.Completed).To(BeFalse())
				Expect(r.Progress).To(Equal(&gdq.Progress{Current: 1, Total: 3}))
			})
		})

		When("the questionnaire reaches completion", func() {
			BeforeEach(func() {
				config = `
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"`
			})

			It("should return completed with no questions", func() {
				r, err := q.Next(map[string]int{"q1": 1})

				Expect(err).ToNot(HaveOccurred())
				Expect(r.Questions).To(BeEmpty())
				Expect(r.Completed).To(BeTrue())
				Expect(r.Progress).To(BeNil())
			})
		})

		When("the questionnaire has invalid conditions", func() {
			When("condition is not a valid expression", func() {
				BeforeEach(func() {
					config = `
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"
    condition: '1 : 2'`
				})

				It("should return an error", func() {
					_, err = q.Next(map[string]int{})
					Expect(err).To(MatchError(ContainSubstring("failed to show question: failed to compile condition expression: ")))
				})
			})

			When("condition does not return a boolean", func() {
				BeforeEach(func() {
					config = `
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"
    condition: '123'`
				})

				It("should return an error", func() {
					_, err = q.Next(map[string]int{})
					Expect(err).To(MatchError("failed to get next questions: failed to show question: condition '123' does not return a boolean"))
				})
			})
		})
	})

	Describe("Closing Remarks", func() {
		var (
			config string
			q      gdq.Questionnaire
			err    error
		)
		JustBeforeEach(func() {
			q, err = gdq.New([]byte(config))
			Expect(err).ToNot(HaveOccurred())
		})

		When("questionnaire has closing remarks", func() {
			BeforeEach(func() {
				config = `
questions:
  - id: "q1"
    text: "Do you like programming?"
    answers:
      - "Yes"
      - "No"

closing_remarks:
  - id: "general"
    text: "Thank you for completing the questionnaire!"
  - id: "programming_lover"
    text: "Great to hear you love programming!"
    condition: 'answers["q1"] == 1'
  - id: "not_interested"
    text: "That's okay, programming isn't for everyone."
    condition: 'answers["q1"] == 2'`
			})

			It("should return remarks when questionnaire is completed", func() {
				r, err := q.Next(map[string]int{"q1": 1})

				Expect(err).ToNot(HaveOccurred())
				Expect(r.Questions).To(BeEmpty())
				Expect(r.Completed).To(BeTrue())
				Expect(r.ClosingRemarks).To(Equal([]gdq.ClosingRemark{
					{Id: "general", Text: "Thank you for completing the questionnaire!"},
					{Id: "programming_lover", Text: "Great to hear you love programming!"},
				}))
			})

			It("should not return remarks when questionnaire is not completed", func() {
				q, err := gdq.New([]byte(`
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Yes"
      - "No"
  - id: "q2"
    text: "Question 2?"
    answers:
      - "Yes"
      - "No"
    condition: 'answers["q1"] == 1'

closing_remarks:
  - id: "general"
    text: "Thank you!"
`))
				Expect(err).ToNot(HaveOccurred())

				response, err := q.Next(map[string]int{})
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Completed).To(BeFalse())
				Expect(response.ClosingRemarks).To(BeEmpty())
			})
		})

		When("questionnaire has no closing remarks", func() {
			BeforeEach(func() {
				config = `
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Yes"
      - "No"`
			})

			It("should return empty remarks when completed", func() {
				r, err := q.Next(map[string]int{"q1": 1})

				Expect(err).ToNot(HaveOccurred())
				Expect(r.Completed).To(BeTrue())
				Expect(r.ClosingRemarks).To(BeEmpty())
			})
		})

		When("questionnaire has invalid closing remark conditions", func() {
			When("condition is not a valid expression", func() {
				BeforeEach(func() {
					config = `
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"

closing_remarks:
  - id: "invalid"
    text: "Invalid remark"
    condition: '1 : 2'`
				})

				It("should return an error", func() {
					_, err = q.Next(map[string]int{"q1": 1})
					Expect(err).To(MatchError(ContainSubstring("failed to evaluate closing remark condition: failed to compile condition expression: ")))
				})
			})

			When("condition does not return a boolean", func() {
				BeforeEach(func() {
					config = `
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"

closing_remarks:
  - id: "non_boolean"
    text: "Non-boolean remark"
    condition: '123'`
				})

				It("should return an error", func() {
					_, err = q.Next(map[string]int{"q1": 1})
					Expect(err).To(MatchError("failed to get next closing remarks: failed to evaluate closing remark condition: condition '123' does not return a boolean"))
				})
			})
		})
	})

	Describe("Progress Tracking", func() {
		var (
			config string
			q      gdq.Questionnaire
			err    error
		)
		JustBeforeEach(func() {
			q, err = gdq.New([]byte(config))
			Expect(err).ToNot(HaveOccurred())
		})

		When("questionnaire has multiple conditional paths", func() {
			BeforeEach(func() {
				config = `
questions:
  - id: "q1"
    text: "Path selector?"
    answers:
      - "Path A"
      - "Path B"
  - id: "q2a"
    text: "Question 2A?"
    answers:
      - "Yes"
      - "No"
    condition: 'answers["q1"] == 1'
  - id: "q3a"
    text: "Question 3A?"
    answers:
      - "Yes"
      - "No"
    condition: 'answers["q1"] == 1'
  - id: "q2b"
    text: "Question 2B?"
    answers:
      - "Option 1"
      - "Option 2"
    condition: 'answers["q1"] == 2'`
			})

			It("should calculate progress correctly for different paths", func() {
				response, err := q.Next(map[string]int{})
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Progress).To(Equal(&gdq.Progress{Current: 0, Total: 1}))

				response, err = q.Next(map[string]int{"q1": 1})
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Progress).To(Equal(&gdq.Progress{Current: 1, Total: 3}))

				response, err = q.Next(map[string]int{"q1": 1, "q2a": 1, "q3a": 2})
				Expect(err).ToNot(HaveOccurred())
				Expect(response.Completed).To(BeTrue())
				Expect(response.Progress).To(BeNil())
			})
		})
	})
})
