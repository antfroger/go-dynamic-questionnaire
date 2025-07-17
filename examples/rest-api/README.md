# Dynamic Questionnaire REST API

A simple, stateless REST API for dynamic questionnaires built with Gin framework.
Questions appear dynamically based on previous answers, creating personalized survey experiences.

## 🚀 Quick Start

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
  POST /questionnaires/{id} - Get questions (with optional answers)
```

## 📚 API Reference

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

Get questions for a questionnaire. Send current answers to get next questions.

**Start questionnaire (empty body):**

```bash
curl -s -X POST http://localhost:8081/questionnaires/survey | jq
```

**Submit answers:**

```bash
curl -s -X POST http://localhost:8081/questionnaires/survey \
  -H "Content-Type: application/json" \
  -d '{"answers": {"satisfaction": 1, "support_quality": 2}}' | jq
```

## 🎯 Complete Usage Guide

### Understanding Answer Format

- **Answers are 1-indexed integers** corresponding to answer options:
  - `1` = First answer option
  - `2` = Second answer option  
  - `3` = Third answer option, etc.

### Step-by-Step Questionnaire Flow

#### 1. Start the Questionnaire

```bash
curl -s -X POST http://localhost:8081/questionnaires/survey | jq
```

**Response:**

```json
{
  "questions": [
    {
      "id": "satisfaction",
      "text": "How satisfied are you with our service overall?",
      "answers": [
        "Very Satisfied",
        "Satisfied",
        "Neutral",
        "Dissatisfied",
        "Very Dissatisfied"
      ]
    },
    {
      "id": "support_quality",
      "text": "How would you rate the quality of our customer support?",
      "answers": [
        "Excellent",
        "Good",
        "Average",
        "Poor",
        "Very Poor"
      ]
    },
    {
      "id": "usage_frequency",
      "text": "How often do you use our service?",
      "answers": [
        "Daily",
        "Several times a week",
        "Weekly",
        "Monthly",
        "Rarely"
      ]
    }
  ],
  "completed": false,
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

**Response:** (Dynamic questions based on your answers)

```json
{
  "questions": [
    {
      "id": "improvement_areas",
      "text": "What areas do you think we need to improve the most?",
      "answers": [
        "Product Features",
        "Customer Support",
        "Pricing",
        "User Experience",
        "Documentation"
      ]
    },
    {
      "id": "support_channel",
      "text": "What is your preferred way to get customer support?",
      "answers": [
        "Live Chat",
        "Email Support",
        "Phone Support",
        "Self-Service Portal",
        "Video Call"
      ]
    }
  ],
  "completed": false,
  "message": "Next questions retrieved"
}
```

#### 3. Continue Until Completion

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
  "completed": true,
  "message": "Questionnaire completed"
}
```
