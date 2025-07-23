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

	Describe("Start", func() {
		When("the questionnaire has just been loaded", func() {
			It("should return the first batch of questions", func() {
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
    condition: "answers['q1'] == 1"
`))
				Expect(err).ToNot(HaveOccurred())

				questions := q.Start()
				Expect(questions).To(Equal([]gdq.Question{
					{Id: "q1", Text: "Question 1?", Answers: []string{"Yes", "No"}},
				}))
			})
		})

		When("the questionnaire contains several questions without conditions", func() {
			It("should return all of these questions", func() {
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
  - id: "q3"
    text: "Question 3?"
    answers:
      - "Yes"
      - "No"
    condition: "answers['q1'] == 1 and answers['q2'] == 1"
`))
				Expect(err).ToNot(HaveOccurred())

				questions := q.Start()
				Expect(questions).To(Equal([]gdq.Question{
					{Id: "q1", Text: "Question 1?", Answers: []string{"Yes", "No"}},
					{Id: "q2", Text: "Question 2?", Answers: []string{"Yes", "No"}},
				}))
			})
		})

		When("the questionnaire is empty", func() {
			It("should return an empty slice", func() {
				q, err := gdq.New([]byte(``))
				Expect(err).ToNot(HaveOccurred())

				questions := q.Start()
				Expect(questions).To(BeEmpty())
			})
		})

		When("the questionnaire has no question without conditions", func() {
			It("should return an empty slice", func() {
				q, err := gdq.New([]byte(`
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"
    condition: "false"
`))
				Expect(err).ToNot(HaveOccurred())

				questions := q.Start()
				Expect(questions).To(BeEmpty())
			})
		})
	})

	Describe("Next", func() {
		var q gdq.Questionnaire

		Context("the questionnaire is valid", func() {
			BeforeEach(func() {
				var err error
				q, err = gdq.New([]byte(`
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
    condition: 'answers["q1"] == 1'

  - id: "q4"
    text: "Question 4?"
    answers:
      - "Answer 1"
      - "Answer 2"
    condition: 'answers["q1"] == 2'

  - id: "q5"
    text: "Question 5?"
    answers:
      - "Answer 1"
      - "Answer 2"
    condition: 'answers["q2"] == 2 and answers["q3"] == 2'

  - id: "q6"
    text: "Question 6?"
    answers:
      - "Answer 1"
      - "Answer 2"
      - "Answer 3"
    condition: 'answers["q2"] in 1..3'
`))
				Expect(err).ToNot(HaveOccurred())
			})

			When("no answers are given", func() {
				It("should return the first batch of questions", func() {
					questions, err := q.Next(map[string]int{})
					Expect(err).ToNot(HaveOccurred())
					Expect(questions).To(Equal([]gdq.Question{
						{Id: "q1", Text: "Question 1?", Answers: []string{"Answer 1", "Answer 2"}},
					}))
				})
			})

			When("one question has been answered", func() {
				It("should return the next questions depending on the answer", func() {
					questions, err := q.Next(map[string]int{"q1": 1})
					Expect(err).ToNot(HaveOccurred())
					Expect(questions).To(Equal([]gdq.Question{
						{Id: "q2", Text: "Question 2?", Answers: []string{"Answer 1", "Answer 2", "Answer 3"}},
						{Id: "q3", Text: "Question 3?", Answers: []string{"Answer 1", "Answer 2", "Answer 3"}},
					}))
				})
			})

			When("several questions have been answered", func() {
				When("more questions are available", func() {
					It("should return the next questions depending on the answers", func() {
						questions, err := q.Next(map[string]int{"q1": 1})
						Expect(err).ToNot(HaveOccurred())
						Expect(questions).To(Equal([]gdq.Question{
							{Id: "q2", Text: "Question 2?", Answers: []string{"Answer 1", "Answer 2", "Answer 3"}},
							{Id: "q3", Text: "Question 3?", Answers: []string{"Answer 1", "Answer 2", "Answer 3"}},
						}))

						questions, err = q.Next(map[string]int{"q1": 1, "q2": 2, "q3": 2})
						Expect(questions).To(Equal([]gdq.Question{
							{Id: "q5", Text: "Question 5?", Answers: []string{"Answer 1", "Answer 2"}},
							{Id: "q6", Text: "Question 6?", Answers: []string{"Answer 1", "Answer 2", "Answer 3"}},
						}))
						Expect(err).ToNot(HaveOccurred())
					})
				})

				When("the questionnaire has reached its end", func() {
					It("should return an empty response", func() {
						questions, err := q.Next(map[string]int{"q4": 1})
						Expect(questions).To(BeEmpty())
						Expect(err).ToNot(HaveOccurred())
					})
				})

				When("the given answers are not valid", func() {
					It("should return an empty response", func() {
						questions, err := q.Next(map[string]int{"q1": 10})
						Expect(questions).To(BeEmpty())
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})

		Context("the questionnaire has invalid conditions", func() {
			When("the condition is not a valid expression", func() {
				It("should return an error", func() {
					q, err := gdq.New([]byte(`
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"
    condition: '1 : 2'
`))
					Expect(err).ToNot(HaveOccurred())

					questions, err := q.Next(map[string]int{})
					Expect(questions).To(BeNil())
					Expect(err).To(MatchError(ContainSubstring("failed to show question: failed to compile condition expression: ")))
				})
			})

			When("the condition does not return a boolean", func() {
				It("should return an error", func() {
					q, err := gdq.New([]byte(`
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"
    condition: '123'
`))
					Expect(err).ToNot(HaveOccurred())

					questions, err := q.Next(map[string]int{})
					Expect(questions).To(BeNil())
					Expect(err).To(MatchError("failed to show question: condition '123' does not return a boolean"))
				})

			})
		})
	})

	Describe("Completed", func() {
		var q gdq.Questionnaire

		BeforeEach(func() {
			var err error
			q, err = gdq.New([]byte(`
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
    condition: 'answers["q1"] == 1'
`))
			Expect(err).ToNot(HaveOccurred())
		})

		When("the questionnaire has just been created", func() {
			It("should return false", func() {
				Expect(q.Completed()).To(BeFalse())
			})
		})

		When("the questionnaire has been started but not finished", func() {
			It("should return false", func() {
				questions := q.Start()
				Expect(questions).ToNot(BeEmpty())
				Expect(q.Completed()).To(BeFalse())

				questions, err := q.Next(map[string]int{"q1": 1})
				Expect(err).ToNot(HaveOccurred())
				Expect(questions).ToNot(BeEmpty())
				Expect(q.Completed()).To(BeFalse())
			})
		})

		When("the questionnaire has been completed", func() {
			It("should return true when no more questions are available", func() {
				questions, err := q.Next(map[string]int{"q1": 2, "q2": 1})
				Expect(err).ToNot(HaveOccurred())
				Expect(questions).To(BeEmpty())
				Expect(q.Completed()).To(BeTrue())
			})
		})
	})

	Describe("ClosingRemarks", func() {
		var q gdq.Questionnaire

		Context("the questionnaire has closing remarks", func() {
			BeforeEach(func() {
				var err error
				q, err = gdq.New([]byte(`
questions:
  - id: "q1"
    text: "Do you like programming?"
    answers:
      - "Yes"
      - "No"
  - id: "q2"
    text: "Which language do you prefer?"
    answers:
      - "Go"
      - "Python"
      - "JavaScript"
    condition: 'answers["q1"] == 1'

closing_remarks:
  - id: "general"
    text: "Thank you for completing the questionnaire!"
  - id: "programming_lover"
    text: "Great to hear you love programming!"
    condition: 'answers["q1"] == 1'
  - id: "go_developer"
    text: "Go is an excellent choice for backend development!"
    condition: 'answers["q1"] == 1 and answers["q2"] == 1'
  - id: "not_interested"
    text: "That's okay, programming isn't for everyone."
    condition: 'answers["q1"] == 2'
`))
				Expect(err).ToNot(HaveOccurred())
			})

			When("the questionnaire is not completed", func() {
				It("should return an empty slice", func() {
					remarks, err := q.ClosingRemarks(map[string]int{"q1": 1})
					Expect(err).ToNot(HaveOccurred())
					Expect(remarks).To(BeEmpty())
				})
			})

			When("the questionnaire is completed", func() {
				It("should return remarks without conditions", func() {
					answers := map[string]int{"q1": 2}
					_, err := q.Next(answers)
					Expect(err).ToNot(HaveOccurred())
					Expect(q.Completed()).To(BeTrue())

					remarks, err := q.ClosingRemarks(answers)
					Expect(err).ToNot(HaveOccurred())
					Expect(remarks).To(Equal([]gdq.ClosingRemark{
						{Id: "general", Text: "Thank you for completing the questionnaire!"},
						{Id: "not_interested", Text: "That's okay, programming isn't for everyone."},
					}))
				})

				It("should return remarks that match the conditions", func() {
					answers := map[string]int{"q1": 1, "q2": 1}
					_, err := q.Next(answers)
					Expect(err).ToNot(HaveOccurred())
					Expect(q.Completed()).To(BeTrue())

					remarks, err := q.ClosingRemarks(answers)
					Expect(err).ToNot(HaveOccurred())
					Expect(remarks).To(Equal([]gdq.ClosingRemark{
						{Id: "general", Text: "Thank you for completing the questionnaire!"},
						{Id: "programming_lover", Text: "Great to hear you love programming!"},
						{Id: "go_developer", Text: "Go is an excellent choice for backend development!"},
					}))
				})
			})
		})

		Context("the questionnaire has no closing remarks", func() {
			BeforeEach(func() {
				var err error
				q, err = gdq.New([]byte(`
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Yes"
      - "No"
`))
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return an empty slice", func() {
				answers := map[string]int{"q1": 1}
				_, err := q.Next(answers)
				Expect(err).ToNot(HaveOccurred())
				Expect(q.Completed()).To(BeTrue())

				remarks, err := q.ClosingRemarks(answers)
				Expect(err).ToNot(HaveOccurred())
				Expect(remarks).To(BeEmpty())
			})
		})

		Context("the questionnaire has invalid closing remark conditions", func() {
			When("the condition is not a valid expression", func() {
				It("should return an error when questionnaire is completed", func() {
					q, err := gdq.New([]byte(`
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"

closing_remarks:
  - id: "invalid"
    text: "Invalid remark"
    condition: '1 : 2'
`))
					Expect(err).ToNot(HaveOccurred())

					answers := map[string]int{"q1": 1}
					_, err = q.Next(answers)
					Expect(err).ToNot(HaveOccurred())
					Expect(q.Completed()).To(BeTrue())

					remarks, err := q.ClosingRemarks(answers)
					Expect(remarks).To(BeNil())
					Expect(err).To(MatchError(ContainSubstring("failed to evaluate closing remark condition: failed to compile condition expression: ")))
				})
			})

			When("the condition does not return a boolean", func() {
				It("should return an error when questionnaire is completed", func() {
					q, err := gdq.New([]byte(`
questions:
  - id: "q1"
    text: "Question 1?"
    answers:
      - "Answer 1"
      - "Answer 2"

closing_remarks:
  - id: "non_boolean"
    text: "Non-boolean remark"
    condition: '123'
`))
					Expect(err).ToNot(HaveOccurred())

					answers := map[string]int{"q1": 1}
					_, err = q.Next(answers)
					Expect(err).ToNot(HaveOccurred())
					Expect(q.Completed()).To(BeTrue())

					remarks, err := q.ClosingRemarks(answers)
					Expect(remarks).To(BeNil())
					Expect(err).To(MatchError("failed to evaluate closing remark condition: condition '123' does not return a boolean"))
				})
			})
		})
	})

})
