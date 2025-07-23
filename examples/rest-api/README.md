# Dynamic Questionnaire REST API

A comprehensive, stateless REST API for dynamic questionnaires built with Gin framework.
Questions appear dynamically based on previous answers, with real-time progress tracking and personalized closing remarks, creating engaging survey experiences.

## Quick Start

### Prerequisites

- Go 1.23+
- curl (for testing)
- jq (optional, for pretty JSON formatting)

### Start the Server

```bash
cd examples/rest-api
go run main.go
```

Server starts on `http://localhost:8081`

```
Starting server on :8081
Available endpoints:
  GET  /questionnaires      - List available questionnaires
  POST /questionnaires/{id} - Get questions with progress and completion
```

## API Reference

### GET /questionnaires

List all available questionnaires.

**Example:**

```bash
curl -s http://localhost:8081/questionnaires | jq
```

**Response:**

```json
{
  "questionnaires": [
    {
      "id": "survey",
      "name": "Survey"
    }
  ]
}
```

### POST /questionnaires/{id}

Get questions for a questionnaire with progress tracking and closing remarks.  
Send current answers to get next questions or completion status.

## Complete Usage Guide

### Understanding Response Format

The API returns a unified response with all necessary information:

```json
{
  "questions": [...],           // Current questions to display
  "closing_remarks": [...],     // Shown only when completed
  "completed": false,           // Completion status
  "progress": {                 // Progress tracking (null when completed)
    "current": 2,
    "total": 5
  },
  "message": "Status message"   // Human-readable status
}
```

### Understanding Answer Format

- **Answers are 1-indexed integers** corresponding to answer options:
  - `1` = First answer option
  - `2` = Second answer option  
  - `3` = Third answer option, etc.

### Step-by-Step questionnaire Flow

#### 1. Start the Questionnaire

```bash
curl -s -X POST http://localhost:8081/questionnaires/survey | jq
```

**Response:**

```json
{
  "questions": [
    {
      "Id": "satisfaction",
      "Text": "How satisfied are you with our service overall?",
      "Answers": [
        "Very Satisfied",
        "Satisfied",
        "Neutral",
        "Dissatisfied",
        "Very Dissatisfied"
      ]
    },
    {
      "Id": "support_quality",
      "Text": "How would you rate the quality of our customer support?",
      "Answers": [
        "Excellent",
        "Good",
        "Average",
        "Poor",
        "Very Poor"
      ]
    },
    {
      "Id": "usage_frequency",
      "Text": "How often do you use our service?",
      "Answers": [
        "Daily",
        "Several times a week",
        "Weekly",
        "Monthly",
        "Rarely"
      ]
    }
  ],
  "completed": false,
  "progress": {
    "current": 0,
    "total": 3
  },
  "message": "Questionnaire started"
}
```

#### 2. Answer the Initial Questions

```bash
curl -s -X POST http://localhost:8081/questionnaires/survey \
  -H "Content-Type: application/json" \
  -d '{
    "answers": {
      "satisfaction": 4,
      "support_quality": 3,
      "usage_frequency": 2
    }
  }' | jq
```

**Response:** (Dynamic, based on your answers)

```json
{
  "questions": [
    {
      "Id": "improvement_areas",
      "Text": "What areas do you think we need to improve the most?",
      "Answers": [
        "Product Features",
        "Customer Support",
        "Pricing",
        "User Experience",
        "Documentation"
      ]
    },
    {
      "Id": "support_channel",
      "Text": "What is your preferred way to get customer support?",
      "Answers": [
        "Live Chat",
        "Email Support",
        "Phone Support",
        "Self-Service Portal",
        "Video Call"
      ]
    }
  ],
  "completed": false,
  "progress": {
    "current": 3,
    "total": 5
  },
  "message": "Next questions retrieved"
}
```

#### 3. Continue until completion

Keep adding answers to your JSON and submitting:

```bash
curl -s -X POST http://localhost:8081/questionnaires/survey \
  -H "Content-Type: application/json" \
  -d '{
    "answers": {
      "satisfaction": 4,
      "support_quality": 3,
      "usage_frequency": 2,
      "improvement_areas": 2,
      "support_channel": 1
    }
  }' | jq
```

**When completed:**

```json
{
  "questions": [],
  "closing_remarks": [
    {
      "id": "thank_you",
      "text": "Thank you for taking the time to provide feedback!"
    },
    {
      "id": "improvement_focus",
      "text": "We appreciate your insights on our customer support. We're working to enhance this area."
    }
  ],
  "completed": true,
  "message": "Questionnaire completed"
}
```

## Error Handling

The API provides clear error responses:

```json
{
  "error": "questionnaire 'invalid-id' not found"
}
```

**Common HTTP Status Codes:**

- `200` - Successful request
- `404` - Questionnaire not found
- `500` - Internal server error

## Customization

### Adding New Questionnaires

1. Create a new YAML file in the directory
2. Add it to the `questionnaires` map in `main.go`:

```go
var questionnaires = map[string]string{
    "survey": "survey.yaml",
    "your-questionnaire": "your-file.yaml",
}
```

### Custom Response Format

Modify the `QuestionsResponse` struct to include additional fields:

```go
type QuestionsResponse struct {
    Questions      []gdq.Question      `json:"questions"`
    ClosingRemarks []gdq.ClosingRemark `json:"closing_remarks,omitempty"`
    Completed      bool                `json:"completed"`
    Progress       *gdq.Progress       `json:"progress,omitempty"`
    Message        string              `json:"message"`
    CustomField    string              `json:"custom_field,omitempty"`
}
```
