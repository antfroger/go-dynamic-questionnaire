# Dynamic Questionnaire

Dynamic Questionnaire is a Go library that generates dynamic questionnaires for you.
The Go Dynamic Questionnaire is a Go library that creates dynamic questionnaires. Want to create a questionnaire for your
Go project that adapts the next questions based on previous answers? Go Dynamic Questionnaire is what you're looking for!
You can even provide recommendations or closing remarks based on the responses.

Dynamic Questionnaire requires Go >= 1.20

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
