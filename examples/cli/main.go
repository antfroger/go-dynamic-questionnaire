package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	gdq "github.com/antfroger/go-dynamic-questionnaire"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run main.go <questionnaire file: tech.yaml|yes-no.yaml>")
	}
	config := os.Args[1]

	// Load questionnaire from YAML file
	questionnaire, err := gdq.New(config)
	if err != nil {
		log.Fatalf("Failed to load questionnaire: %v", err)
	}

	answers := askQuestions(questionnaire)
	displayResults(answers)
}

func askQuestions(questionnaire gdq.Questionnaire) map[string]int {
	answers := make(map[string]int)

	for {
		response, err := questionnaire.Next(answers)
		if err != nil {
			log.Fatalf("Failed to get next questions: %v", err)
		}

		if response.Completed {
			displayClosingRemarks(response.ClosingRemarks)
			break
		}

		displayProgress(response.Progress)

		for _, question := range response.Questions {
			answer := askQuestion(question)
			answers[question.Id] = answer
		}
	}

	return answers
}

func displayProgress(progress *gdq.Progress) {
	if progress == nil {
		return
	}

	fmt.Printf("\n📊 Progress: %d/%d questions answered\n", progress.Current, progress.Total)
	percentage := float64(progress.Current) / float64(progress.Total) * 100
	fmt.Printf("🔄 %.0f%% complete\n", percentage)
}

func askQuestion(question gdq.Question) int {
	fmt.Printf("\n%s (ID: %s)\n", question.Text, question.Id)
	for i, answer := range question.Answers {
		fmt.Printf("  - %s (%d)\n", answer, i+1)
	}

	var input string
	for {
		fmt.Print("Select answer: ")
		fmt.Scanln(&input)

		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || choice < 1 || choice > len(question.Answers) {
			fmt.Printf("Invalid choice. Please enter 1 - %d.\n", len(question.Answers))
			continue
		}

		return choice
	}
}

func displayClosingRemarks(remarks []gdq.ClosingRemark) {
	if len(remarks) == 0 {
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 40))
	fmt.Println("QUESTIONNAIRE COMPLETE!")
	fmt.Println(strings.Repeat("=", 40))

	for _, remark := range remarks {
		fmt.Printf("💬 %s\n", remark.Text)
	}
}

func displayResults(answers map[string]int) {
	fmt.Println("\n" + strings.Repeat("=", 40))
	fmt.Println("YOUR ANSWERS")
	fmt.Println(strings.Repeat("=", 40))

	for id, answer := range answers {
		fmt.Printf("  %s: %d\n", id, answer)
	}
}
