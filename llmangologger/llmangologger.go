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
