package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	gdq "github.com/antfroger/go-dynamic-questionnaire"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Available questionnaires
var questionnaires = map[string]string{
	"survey": "survey.yaml",
}

// Response structures
type (
	QuestionnairesResponse struct {
		Questionnaires []Questionnaire `json:"questionnaires"`
	}
	Questionnaire struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	QuestionsRequest struct {
		Answers map[string]int `json:"answers,omitempty"`
	}
	QuestionsResponse struct {
		Questions      []gdq.Question      `json:"questions"`
		ClosingRemarks []gdq.ClosingRemark `json:"closing_remarks,omitempty"`
		Completed      bool                `json:"completed"`
		Progress       *gdq.Progress       `json:"progress,omitempty"`
		Message        string              `json:"message"`
	}
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	r.GET("/questionnaires", handleQuestionnaires)
	r.POST("/questionnaires/:id", handleQuestions)

	log.Println("Starting server on :8081")
	log.Println("Available endpoints:")
	log.Println("  GET  /questionnaires      - List available questionnaires")
	log.Println("  POST /questionnaires/{id} - Get questions (with optional answers)")

	log.Fatal(r.Run(":8081"))
}

// GET /questionnaires - List available questionnaires
func handleQuestionnaires(c *gin.Context) {
	caser := cases.Title(language.English)

	var list []Questionnaire
	for id := range questionnaires {
		list = append(list, Questionnaire{
			ID:   id,
			Name: caser.String(strings.ReplaceAll(id, "-", " ")),
		})
	}

	response := QuestionnairesResponse{Questionnaires: list}
	c.JSON(http.StatusOK, response)
}

// POST /questionnaires/{id}/questions - Get questions based on current answers
func handleQuestions(c *gin.Context) {
	q, err := loadQuestionnaire(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var r QuestionsRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		// If body is empty or invalid, treat as starting questionnaire
		r.Answers = make(map[string]int)
	}

	response, err := q.Next(r.Answers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get next questions: %v", err)})
		return
	}

	message := "Next questions retrieved"
	if len(r.Answers) == 0 {
		message = "Questionnaire started"
	} else if response.Completed {
		message = "Questionnaire completed"
	}

	apiResponse := QuestionsResponse{
		Questions:      response.Questions,
		ClosingRemarks: response.ClosingRemarks,
		Completed:      response.Completed,
		Progress:       response.Progress,
		Message:        message,
	}

	c.JSON(http.StatusOK, apiResponse)
}

func loadQuestionnaire(id string) (gdq.Questionnaire, error) {
	path, err := getConfig(id)
	if err != nil {
		return nil, err
	}

	questionnaire, err := gdq.New(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load questionnaire %s", id)
	}
	return questionnaire, nil
}

func getConfig(questionnaireID string) (string, error) {
	configPath, exists := questionnaires[questionnaireID]
	if !exists {
		return "", fmt.Errorf("questionnaire '%s' not found", questionnaireID)
	}

	return getQuestionnairePath(configPath), nil
}

// getQuestionnairePath returns the full path to a questionnaire file
func getQuestionnairePath(filename string) string {
	if _, err := os.Stat(filename); err == nil {
		return filename
	}

	cwd, _ := os.Getwd()
	fullPath := filepath.Join(cwd, filename)
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath
	}

	// fallback
	return filename
}
