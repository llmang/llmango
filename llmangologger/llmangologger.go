package llmangologger

import (
	"errors"
	"log"

	"github.com/llmang/llmango/llmango"
)

type MangoLoggingSetup struct {
	Logger func(*llmango.LLMangoLog) error
	MangoLoggingOptions
}

type MangoLoggingOptions struct {
	LogRawRequestResponse bool
}

func UseLogger(m *llmango.LLMangoManager, logger *log.Logger, opts *MangoLoggingOptions) error {
	if logger == nil {
		return errors.New("logger cannot be nil")
	}

	m.Logging.LogResponse = func(mangolog *llmango.LLMangoLog) error {
		if opts == nil || !opts.LogRawRequestResponse {
			mangolog.RawRequest = ""
			mangolog.RawResponse = ""
		}
		logger.Println(mangolog)
		return nil
	}
	m.Logging.GetLogs = nil

	return nil
}

func UseConsoleLogging(m *llmango.LLMangoManager, opts *MangoLoggingOptions) error {
	m.Logging.LogResponse = func(mangolog *llmango.LLMangoLog) error {
		if opts != nil || !opts.LogRawRequestResponse {
			mangolog.RawRequest = ""
			mangolog.RawResponse = ""
		}
		log.Printf("MANGO: %v\n", mangolog)
		return nil
	}
	m.Logging.GetLogs = nil

	return nil
}

// CreatePrintLogger creates a simple console logger that can be used with WithLogging()
func CreatePrintLogger(logFullRequests bool) *llmango.Logging {
	return &llmango.Logging{
		LogResponse: func(mangolog *llmango.LLMangoLog) error {
			log.Printf("=== MANGO LOG ENTRY ===\n")
			log.Printf("Timestamp: %d\n", mangolog.Timestamp)
			log.Printf("Goal UID: %s\n", mangolog.GoalUID)
			log.Printf("Prompt UID: %s\n", mangolog.PromptUID)
			log.Printf("Input Object: %s\n", mangolog.InputObject)
			log.Printf("Output Object: %s\n", mangolog.OutputObject)
			log.Printf("Input Tokens: %d\n", mangolog.InputTokens)
			log.Printf("Output Tokens: %d\n", mangolog.OutputTokens)
			log.Printf("Cost: %.6f\n", mangolog.Cost)
			log.Printf("Request Time: %.3f\n", mangolog.RequestTime)
			log.Printf("Generation Time: %.3f\n", mangolog.GenerationTime)
			log.Printf("User ID: %s\n", mangolog.UserID)

			if logFullRequests {
				log.Printf("=== RAW REQUEST ===\n")
				log.Printf("%s\n", mangolog.RawRequest)
				log.Printf("=== RAW RESPONSE ===\n")
				log.Printf("%s\n", mangolog.RawResponse)
				log.Printf("====================\n")
			}

			if mangolog.Error != "" {
				log.Printf("ERROR: %s\n", mangolog.Error)
			}

			log.Printf("==============================\n")
			return nil
		},
		GetLogs: nil, // Print logger doesn't support log retrieval
	}
}

// CreateNoOpLogger creates a logger that doesn't log anything
func CreateNoOpLogger() *llmango.Logging {
	return &llmango.Logging{
		LogResponse: func(mangolog *llmango.LLMangoLog) error {
			return nil // Do nothing
		},
		GetLogs: nil,
	}
}
