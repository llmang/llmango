package llmango

import (
	"errors"
	"fmt"
)

// DeletePrompt deletes a prompt from the manager
func (l *LLMangoManager) DeletePrompt(promptUID string) error {
	if l.Prompts == nil {
		return errors.New("prompts map is not initialized")
	}

	// Check if prompt exists
	if _, exists := l.Prompts[promptUID]; !exists {
		return errors.New("prompt not found")
	}

	// Delete the prompt
	delete(l.Prompts, promptUID)

	// Save state if SaveState function is set
	if l.SaveState != nil {
		if err := l.SaveState(); err != nil {
			return fmt.Errorf("failed to save state after deleting prompt: %w", err)
		}
	}

	return nil
}
