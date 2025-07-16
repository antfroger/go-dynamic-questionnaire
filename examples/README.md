# Dynamic Questionnaire Examples

This directory contains various examples of how to use the **Go Dynamic Questionnaire** library in different applications and scenarios.
The library provides a flexible way to create interactive questionnaires with conditional logic based on previous answers.

## Examples

### 1. CLI Application (`cli/`)

**Use Case**: Interactive command-line questionnaire tool

**Files**:

- `main.go` - Complete CLI implementation
- `tech.yaml` - Career and technology questionnaire
- `yes-no.yaml` - Simple yes/no questionnaire example

**Key Features**:

- Command-line argument parsing
- Interactive question prompting
- Input validation
- Results display

**Run Example**:

```bash
cd cli/
go run main.go tech.yaml
# or
go run main.go yes-no.yaml
```
