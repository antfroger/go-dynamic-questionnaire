# Dynamic Questionnaire

The Go Dynamic Questionnaire is a Go library that creates dynamic questionnaires.  
Want to create a questionnaire for your Go project that adapts the next questions based on previous answers?
Go Dynamic Questionnaire is what you're looking for!
You can even provide recommendations or closing remarks based on the responses.

Dynamic Questionnaire requires Go >= 1.23

[![PkgGoDev](https://pkg.go.dev/badge/github.com/antfroger/go-dynamic-questionnaire)](https://pkg.go.dev/github.com/antfroger/go-dynamic-questionnaire)
[![CI](https://github.com/antfroger/go-dynamic-questionnaire/actions/workflows/go.yml/badge.svg)](https://github.com/antfroger/go-dynamic-questionnaire/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/antfroger/go-dynamic-questionnaire)](https://goreportcard.com/report/github.com/antfroger/go-dynamic-questionnaire)
[![Release](https://img.shields.io/github/release/antfroger/go-dynamic-questionnaire.svg?style=flat-square)](https://github.com/antfroger/go-dynamic-questionnaire/releases)

## Installation

```shell
go get "github.com/antfroger/go-dynamic-questionnaire"
```

## Basic Usage

Use `questionnaire.New("config.yaml")` to create and initialize the questionnaire based on your config.

```go
import "github.com/antfroger/go-dynamic-questionnaire"

func main() {
    q, err := questionnaire.New("config.yaml")
    if err != nil {
        // ...
    }

    answers := make(map[string]int)
    for {
        response, err := q.Next(answers)
        if err != nil {
            // ...
        }

        // When the questionnaire is completed, display closing remarks
        if response.Completed {
            displayClosingRemarks(response.ClosingRemarks)
            break
        }

        // Display progress if available
        if response.Progress != nil {
            fmt.Printf("Progress: %d/%d questions answered\n", response.Progress.Current, response.Progress.Total)
        }

        // Ask the questions and get the answers from the user
        for _, question := range response.Questions {
            answer := getUserAnswer(question)
            answers[question.Id] = answer
        }
    }
}
```

## Features

### Unified API

Single `Next()` method returns everything you need:

- Questions to display
- Closing remarks (when completed)
- Completion status
- Progress tracking

### Progress Tracking

Track user progress through the questionnaire:

```go
type Progress struct {
    Current int `json:"current"`  // Questions answered
    Total   int `json:"total"`    // Total questions in the current path
}
```

### Closing Remarks

Show personalized messages when questionnaire is completed:

```yaml
closing_remarks:
  - id: "general"
    text: "Thank you for completing the questionnaire!"
  - id: "specific"
    text: "Based on your answers, here's our recommendation..."
    condition: 'answers["interest"] == 1'
```

### Conditional Logic

Dynamic question flow based on previous answers:

```yaml
questions:
  - id: "experience"
    text: "Do you have programming experience?"
    answers:
      - "Yes"
      - "No"
  - id: "language"
    text: "Which language do you prefer?"
    condition: 'answers["experience"] == 1'
    answers:
      - "Go"
      - "Python"
      - "JavaScript"
```

## Response Structure

The `Next()` method returns a unified response:

```go
type Response struct {
    Questions      []Question      `json:"questions"`
    ClosingRemarks []ClosingRemark `json:"closing_remarks,omitempty"`
    Completed      bool            `json:"completed"`
    Progress       *Progress       `json:"progress,omitempty"`
}
```

## Configuration Format

Create questionnaires using YAML:

```yaml
questions:
  - id: "satisfaction"
    text: "How satisfied are you with our service?"
    answers:
      - "Very Satisfied"
      - "Satisfied"
      - "Neutral"
      - "Dissatisfied"
      - "Very Dissatisfied"

  - id: "recommendation"
    text: "Would you recommend us to others?"
    condition: 'answers["satisfaction"] in [1,2]'
    answers:
      - "Definitely"
      - "Probably"
      - "Maybe"

closing_remarks:
  - id: "thank_you"
    text: "Thank you for your feedback!"
  - id: "follow_up"
    text: "We'll reach out to address your concerns."
    condition: 'answers["satisfaction"] >= 4'
```

## Examples

### CLI Application

Interactive command-line questionnaire with progress display:

```bash
cd examples/cli
go run main.go tech.yaml
```

[More details in the dedicated README.](examples/cli/README.md)

### REST API Server

Complete web service with questionnaire endpoints:

```bash
cd examples/rest-api
go run main.go
```

API endpoints:

- `GET /questionnaires` - List available questionnaires
- `POST /questionnaires/{id}` - Get questions with progress and closing remarks

[More details in the dedicated README.](examples/rest-api/README.md)

## Advanced Features

### Thread-Safe Design

The library is stateless and thread-safe by design. Each questionnaire instance can be safely used across multiple goroutines.

### Expression Engine

Powerful condition expressions using the [`expr`](https://github.com/expr-lang/expr) library:

```yaml
condition: 'answers["age"] >= 18 and answers["country"] == "US"'
condition: 'answers["score"] in 1..5'
condition: 'len(answers) >= 3'
```

### Flexible Input

Load questionnaires from files or byte arrays:

```go
// From file
q, err := questionnaire.New("config.yaml")

// From bytes
yamlData := []byte(`questions: ...`)
q, err := questionnaire.New(yamlData)
```

## TODOs

- [ ] Add loaders for different configuration formats
  - read the configuration from a JSON file
  - read the configuration from JSON bytes
  - pass pre-configured questions
