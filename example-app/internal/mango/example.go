package mango

import (
	"fmt"

	"github.com/llmang/llmango/llmango"
)

// Sentiment Analysis Types (Go-defined goal)
type SentimentInput struct {
	Text string `json:"text"`
}

type SentimentOutput struct {
	Sentiment  string  `json:"sentiment"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// Text Summary Types (Go-defined goal)
type SummaryInput struct {
	Text string `json:"text"`
}

type SummaryOutput struct {
	Summary   string   `json:"summary"`
	KeyPoints []string `json:"key_points"`
	WordCount int      `json:"word_count"`
}

// Code Review Types (Go-defined goal)
type CodeReviewInput struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

type CodeReviewOutput struct {
	Issues      []string `json:"issues"`
	Suggestions []string `json:"suggestions"`
	Rating      int      `json:"rating"`
	Summary     string   `json:"summary"`
}

// Translation Types (Go-defined goal)
type TranslationInput struct {
	Text       string `json:"text"`
	TargetLang string `json:"target_lang"`
}

type TranslationOutput struct {
	Translation string  `json:"translation"`
	Confidence  float64 `json:"confidence"`
	SourceLang  string  `json:"source_lang"`
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

// Inline Go Goals (using NewGoal - these return *Goal pointers)
var sentimentGoal = llmango.NewGoal(
	"sentiment-analysis",
	"Sentiment Analysis",
	"Analyzes the sentiment of text input (defined in Go code)",
	SentimentInput{Text: "I absolutely love this new AI system! It's incredibly helpful."},
	SentimentOutput{
		Sentiment:  "positive",
		Confidence: 0.95,
		Reasoning:  "Contains positive language like 'love' and 'incredibly helpful'",
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

var codeReviewGoal = llmango.NewGoal(
	"code-review",
	"Code Review",
	"Reviews code and provides suggestions (defined in Go code)",
	CodeReviewInput{
		Code:     "func main() {\n\tfmt.Println(\"Hello World\")\n}",
		Language: "go",
	},
	CodeReviewOutput{
		Issues:      []string{"Missing package declaration", "No error handling"},
		Suggestions: []string{"Add package main", "Consider adding comments"},
		Rating:      7,
		Summary:     "Basic Go program with minor issues",
	},
	llmango.TypedValidator[CodeReviewInput, CodeReviewOutput]{
		ValidateInput: func(input CodeReviewInput) error {
			if input.Code == "" {
				return fmt.Errorf("code is required")
			}
			if input.Language == "" {
				return fmt.Errorf("language is required")
			}
			return nil
		},
		ValidateOutput: func(output CodeReviewOutput) error {
			if output.Rating < 1 || output.Rating > 10 {
				return fmt.Errorf("rating must be between 1 and 10")
			}
			if output.Summary == "" {
				return fmt.Errorf("summary is required")
			}
			return nil
		},
	},
)

// Translation Goal (using NewGoal - these return *Goal pointers)
var translationGoal = llmango.NewGoal(
	"translation",
	"Translation",
	"Translates text between languages (defined in Go code)",
	TranslationInput{
		Text:       "Hello, how are you?",
		TargetLang: "spanish",
	},
	TranslationOutput{
		Translation: "Hola, ¿cómo estás?",
		Confidence:  0.95,
		SourceLang:  "english",
	},
	llmango.TypedValidator[TranslationInput, TranslationOutput]{
		ValidateInput: func(input TranslationInput) error {
			if input.Text == "" {
				return fmt.Errorf("text is required")
			}
			if input.TargetLang == "" {
				return fmt.Errorf("target language is required")
			}
			return nil
		},
		ValidateOutput: func(output TranslationOutput) error {
			if output.Translation == "" {
				return fmt.Errorf("translation is required")
			}
			if output.Confidence < 0 || output.Confidence > 1 {
				return fmt.Errorf("confidence must be between 0 and 1")
			}
			return nil
		},
	},
)

// Text Summary Goal (using NewGoal - these return *Goal pointers)
var summaryGoal = llmango.NewGoal(
	"text-summary",
	"Text Summary",
	"Summarizes long text into key points (defined in Go code)",
	SummaryInput{
		Text: "Artificial intelligence (AI) is a rapidly evolving field that encompasses machine learning, natural language processing, and computer vision. Recent advances in large language models have revolutionized how we interact with computers, enabling more natural conversations and automated content generation. However, challenges remain in areas such as bias mitigation, energy efficiency, and ensuring AI systems remain aligned with human values.",
	},
	SummaryOutput{
		Summary:   "AI is rapidly advancing with breakthroughs in language models, but faces challenges in bias, efficiency, and alignment.",
		KeyPoints: []string{"AI encompasses ML, NLP, and computer vision", "Large language models enable natural interaction", "Challenges include bias, energy use, and human alignment"},
		WordCount: 3,
	},
	llmango.TypedValidator[SummaryInput, SummaryOutput]{
		ValidateInput: func(input SummaryInput) error {
			if input.Text == "" {
				return fmt.Errorf("text is required")
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
			if output.WordCount < 0 {
				return fmt.Errorf("word count must be non-negative")
			}
			return nil
		},
	},
)

