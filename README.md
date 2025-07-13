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

## TODOs

- [ ] Configure GitHub actions to run tests (cf. https://docs.github.com/en/actions/how-tos/writing-workflows/building-and-testing/building-and-testing-go)
- [ ] Add recommendations and closing remarks
- [ ] Add a function to check if the questionnaire is finished
- [ ] Add more examples
- [ ] Add loaders for different configuration formats
  - read the configuration from a JSON file
  - read the configuration from JSON bytes
  - pass pre-configured questions
