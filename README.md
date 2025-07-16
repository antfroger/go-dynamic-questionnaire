# Dynamic Questionnaire

The Go Dynamic Questionnaire is a Go library that creates dynamic questionnaires.  
Want to create a questionnaire for your Go project that adapts the next questions based on previous answers?
Go Dynamic Questionnaire is what you're looking for!
You can even provide recommendations or closing remarks based on the responses.

Dynamic Questionnaire requires Go >= 1.20

[![PkgGoDev](https://pkg.go.dev/badge/github.com/antfroger/go-dynamic-questionnaire)](https://pkg.go.dev/github.com/antfroger/go-dynamic-questionnaire)
[![CI](https://github.com/antfroger/go-dynamic-questionnaire/actions/workflows/check.yml/badge.svg)](https://github.com/antfroger/go-dynamic-questionnaire/actions/workflows/check.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/antfroger/go-dynamic-questionnaire)](https://goreportcard.com/report/github.com/antfroger/go-dynamic-questionnaire)

## Installation

Add this to your Go file

```go
import "github.com/antfroger/go-dynamic-questionnaire"
```

And run `go get` to get the package.

## Basic Usage

Use `questionnaire.New("config.yaml")` to create and initialize the questionnaire based on your config.

```go
import "github.com/antfroger/go-dynamic-questionnaire"

func main() {
    q := questionnaire.New("config.yaml")

    // Start the questionnaire and returns the first set of questions; the one(s) without conditions
    questions := q.Start()

    // Collect answer(s) from the user
    var answers map[string]int

    // After collecting answers, call Next to get the next set of questions based on the answers
    questions, err := q.Next(answers)
}
```

## TODOs

- [ ] Add recommendations and closing remarks
- [ ] Add a function to check if the questionnaire is finished
- [ ] Add more examples
- [ ] Add loaders for different configuration formats
  - read the configuration from a JSON file
  - read the configuration from JSON bytes
  - pass pre-configured questions
