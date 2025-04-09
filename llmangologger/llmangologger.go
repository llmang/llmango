package llmangologger

import (
	"errors"
	"fmt"
	"log"

	"github.com/llmang/llmango/llmango"
)

type MangoLoggingSetup struct {
	Logger func(*llmango.LLMangoLog) error
	MangoLoggingOptions
}

type MangoLoggingOptions struct {
	LogFullInputOutputMessages bool
	LogPercentage              int
}

func UseLogger(m *llmango.LLMangoManager, logger *log.Logger, opts *MangoLoggingOptions) error {
	if logger == nil {
		return errors.New("logger cannot be nil")
	}

	m.LogResponse = func(mangolog *llmango.LLMangoLog) error {
		logger.Println(mangolog)
		return nil
	}
	m.GetLogs = nil
	m.LogPercentage = opts.LogPercentage
	m.LogFullInputOutputMessages = opts.LogFullInputOutputMessages

	return nil
}

func UseConsoleLogging(m *llmango.LLMangoManager, opts *MangoLoggingOptions) error {
	m.LogResponse = func(mangolog *llmango.LLMangoLog) error {
		fmt.Printf("MANGO: %v\n", mangolog)
		return nil
	}
	m.GetLogs = nil
	m.LogPercentage = opts.LogPercentage
	m.LogFullInputOutputMessages = opts.LogFullInputOutputMessages

	return nil
}
