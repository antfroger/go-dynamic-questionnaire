package go_dynamic_questionnaire

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Loader", func() {
	Describe("getLoaderForConfig", func() {
		Context("with string file paths", func() {
			It("should return yamlLoader for .yaml files", func() {
				loader, err := getLoaderForConfig("test.yaml")
				Expect(err).ToNot(HaveOccurred())
				Expect(loader).To(BeAssignableToTypeOf(&yamlLoader{}))
			})

			It("should return yamlLoader for .yml files", func() {
				loader, err := getLoaderForConfig("test.yml")
				Expect(err).ToNot(HaveOccurred())
				Expect(loader).To(BeAssignableToTypeOf(&yamlLoader{}))
			})

			It("should return jsonLoader for .json files", func() {
				loader, err := getLoaderForConfig("test.json")
				Expect(err).ToNot(HaveOccurred())
				Expect(loader).To(BeAssignableToTypeOf(&jsonLoader{}))
			})

			It("should return error for unsupported file extensions", func() {
				loader, err := getLoaderForConfig("test.txt")
				Expect(err).To(MatchError("unsupported file extension .txt: expected .yaml, .yml, or .json"))
				Expect(loader).To(BeNil())
			})
		})

		Context("with byte array content", func() {
			It("should return jsonLoader for JSON content", func() {
				jsonContent := []byte(`{"questions": []}`)
				loader, err := getLoaderForConfig(jsonContent)
				Expect(err).ToNot(HaveOccurred())
				Expect(loader).To(BeAssignableToTypeOf(&jsonLoader{}))
			})

			It("should return yamlLoader for YAML content", func() {
				yamlContent := []byte("questions: []")
				loader, err := getLoaderForConfig(yamlContent)
				Expect(err).ToNot(HaveOccurred())
				Expect(loader).To(BeAssignableToTypeOf(&yamlLoader{}))
			})

			It("should return yamlLoader for empty content (default)", func() {
				emptyContent := []byte("")
				loader, err := getLoaderForConfig(emptyContent)
				Expect(err).ToNot(HaveOccurred())
				Expect(loader).To(BeAssignableToTypeOf(&yamlLoader{}))
			})

			It("should return jsonLoader for array content", func() {
				arrayContent := []byte(`[{"id": "test"}]`)
				loader, err := getLoaderForConfig(arrayContent)
				Expect(err).ToNot(HaveOccurred())
				Expect(loader).To(BeAssignableToTypeOf(&jsonLoader{}))
			})
		})

		Context("with unsupported types", func() {
			It("should return error for unsupported types", func() {
				loader, err := getLoaderForConfig(123)
				Expect(err).To(MatchError("unsupported config type: expected string (file path) or []byte (content), got int"))
				Expect(loader).To(BeNil())
			})
		})
	})

	Describe("yamlLoader", func() {
		var loader *yamlLoader

		BeforeEach(func() {
			loader = &yamlLoader{}
		})

		Context("with string file path", func() {
			It("should load YAML from file", func() {
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

				q := &questionnaire{}
				err = loader.Load(tmpFile.Name(), q)
				Expect(err).To(BeNil())
				Expect(q).NotTo(BeNil())
			})

			It("should return error for non-existent file", func() {
				q := &questionnaire{}
				err := loader.Load("non-existent.yaml", q)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to read file"))
			})
		})

		Context("with byte array content", func() {
			It("should load YAML from bytes", func() {
				yamlContent := []byte(`
questions:
  - id: test1
    text: Test question 1
    answers: ["Yes", "No"]
  - id: test2
    text: Test question 2
    answers: ["A", "B", "C"]
`)
				q := &questionnaire{}
				err := loader.Load(yamlContent, q)
				Expect(err).ToNot(HaveOccurred())
				Expect(q.Questions).ToNot(BeNil())
				Expect(len(q.Questions)).To(Equal(2))
				Expect(q.Questions[0].Id).To(Equal("test1"))
				Expect(q.Questions[1].Id).To(Equal("test2"))
			})

			It("should return error for invalid YAML", func() {
				invalidYaml := []byte("invalid: yaml: content: [")
				q := &questionnaire{}
				err := loader.Load(invalidYaml, q)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse content"))
			})
		})

		Context("with unsupported data types", func() {
			It("should return error for unsupported types", func() {
				q := &questionnaire{}
				err := loader.Load(123, q)
				Expect(err).To(MatchError("unsupported data type for loader: int"))
			})
		})
	})

	Describe("jsonLoader", func() {
		var loader *jsonLoader

		BeforeEach(func() {
			loader = &jsonLoader{}
		})

		Context("with string file path", func() {
			It("should load JSON from file", func() {
				content := []byte(`
{
  "questions": [
    {
      "id": "q1",
      "text": "Question 1?",
      "answers": [
        "Answer 1",
        "Answer 2"
      ]
    }
  ]
}
`)
				tmpFile, err := os.CreateTemp("", "questionnaire-*.json")
				Expect(err).To(BeNil())
				defer func(name string) {
					_ = os.Remove(name)
				}(tmpFile.Name())

				_, err = tmpFile.Write(content)
				Expect(err).To(BeNil())
				err = tmpFile.Close()
				Expect(err).To(BeNil())

				q := &questionnaire{}
				err = loader.Load(tmpFile.Name(), q)
				Expect(err).To(BeNil())
				Expect(q).NotTo(BeNil())
			})

			It("should return error for non-existent file", func() {
				q := &questionnaire{}
				err := loader.Load("non-existent.json", q)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to read file"))
			})
		})

		Context("with byte array content", func() {
			It("should load JSON from bytes", func() {
				jsonContent := []byte(`{
  "questions": [
    {
      "id": "test1",
      "text": "Test question 1",
      "answers": ["Yes", "No"]
    },
    {
      "id": "test2",
      "text": "Test question 2",
      "answers": ["A", "B", "C"]
    }
  ]
}`)
				q := &questionnaire{}
				err := loader.Load(jsonContent, q)
				Expect(err).ToNot(HaveOccurred())
				Expect(q.Questions).ToNot(BeNil())
				Expect(len(q.Questions)).To(Equal(2))
				Expect(q.Questions[0].Id).To(Equal("test1"))
				Expect(q.Questions[1].Id).To(Equal("test2"))
			})

			It("should return error for invalid JSON", func() {
				invalidJson := []byte(`{"questions": [invalid json}`)
				q := &questionnaire{}
				err := loader.Load(invalidJson, q)
				Expect(err).To(MatchError(ContainSubstring("failed to parse content")))
			})
		})

		Context("with unsupported data types", func() {
			It("should return error for unsupported types", func() {
				q := &questionnaire{}
				err := loader.Load(123, q)
				Expect(err).To(MatchError("unsupported data type for loader: int"))
			})
		})
	})

	Describe("validateLoadedQuestionnaire", func() {
		It("should initialize nil slices", func() {
			q := &questionnaire{}
			err := validateLoadedQuestionnaire(q)
			Expect(err).ToNot(HaveOccurred())
			Expect(q.Questions).ToNot(BeNil())
			Expect(q.Remarks).ToNot(BeNil())
			Expect(q.Questions).To(HaveLen(0))
			Expect(q.Remarks).To(HaveLen(0))
		})

		It("should not modify existing slices", func() {
			q := &questionnaire{
				Questions: []question{{Id: "test", Text: "Test", Answers: []string{"Yes"}}},
				Remarks:   []closingRemark{{Id: "remark", Text: "Test remark"}},
			}
			err := validateLoadedQuestionnaire(q)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(q.Questions)).To(Equal(1))
			Expect(len(q.Remarks)).To(Equal(1))
		})
	})
})
