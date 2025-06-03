package mango

import (
	"fmt"
	
	"github.com/llmang/llmango/llmango"
)

// Sentiment Analysis Types
type SentimentInput struct {
	Text string `json:"text"`
}

type SentimentOutput struct {
	Sentiment  string  `json:"sentiment"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// Text Summary Types
type SummaryInput struct {
	Text string `json:"text"`
}

type SummaryOutput struct {
	Summary    string   `json:"summary"`
	KeyPoints  []string `json:"key_points"`
	WordCount  int      `json:"word_count"`
}

// Config-based goal types (these types support the config-defined goals)
type EmailInput struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Sender  string `json:"sender"`
}

type EmailOutput struct {
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

type LanguageInput struct {
	Text string `json:"text"`
}

type LanguageOutput struct {
	Language     string  `json:"language"`
	LanguageCode string  `json:"language_code"`
	Confidence   float64 `json:"confidence"`
}

// Inline Go Goals (using new dual-mode system)
var sentimentGoal = llmango.NewGoal(
	"sentiment-analysis",
	"Sentiment Analysis",
	"Analyzes the sentiment of text input",
	SentimentInput{Text: "I love this new product! It works perfectly."},
	SentimentOutput{
		Sentiment:  "positive",
		Confidence: 0.95,
		Reasoning:  "Contains positive language like 'love' and 'perfectly'",
	},
	llmango.TypedValidator[SentimentInput, SentimentOutput]{
		ValidateInput: func(input SentimentInput) error {
			if input.Text == "" {
				return fmt.Errorf("text is required")
			}
			return nil
		},
		ValidateOutput: func(output SentimentOutput) error {
			if output.Sentiment == "" {
				return fmt.Errorf("sentiment is required")
			}
			if output.Confidence < 0 || output.Confidence > 1 {
				return fmt.Errorf("confidence must be between 0 and 1")
			}
			return nil
		},
	},
)

// Text Summary Goal (using new dual-mode system)
var summaryGoal = llmango.NewGoal(
	"text-summary",
	"Text Summary",
	"Summarizes long text into key points",
	SummaryInput{Text: "This is a long article about artificial intelligence and its impact on society..."},
	SummaryOutput{
		Summary:   "Article discusses AI's societal impact",
		KeyPoints: []string{"AI transformation", "Social implications", "Future outlook"},
		WordCount: 150,
	},
	llmango.TypedValidator[SummaryInput, SummaryOutput]{
		ValidateInput: func(input SummaryInput) error {
			if input.Text == "" {
				return fmt.Errorf("text is required")
			}
			if len(input.Text) < 10 {
				return fmt.Errorf("text must be at least 10 characters long")
			}
			return nil
		},
		ValidateOutput: func(output SummaryOutput) error {
			if output.Summary == "" {
				return fmt.Errorf("summary is required")
			}
			if len(output.KeyPoints) == 0 {
				return fmt.Errorf("at least one key point is required")
			}
			return nil
		},
	},
)

// Note: Prompts are now defined in llmango.yaml and will be CLI-generated
// This demonstrates the separation between inline Go goals and config-based prompts
