package llmangosavestate

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"time"

	"github.com/llmang/llmango/llmango"
)

type mangoConfigFile struct {
	PinnedModels []string                     `json:"pinnedModels"`
	Goals        map[string]*llmango.GoalInfo `json:"goals"`
	Prompts      map[string]*llmango.Prompt   `json:"prompts"`
}

func jsonSaveStateFunc(mango *llmango.LLMangoManager, fileName string) error {
	writeObject := struct {
		Goals   map[string]any             `json:"goals"`
		Prompts map[string]*llmango.Prompt `json:"prompts"`
	}{
		Goals:   mango.Goals,
		Prompts: mango.Prompts,
	}

	// Write updated config to file
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(writeObject); err != nil {
		return err
	}
	return nil
}

// You create an internal LLMang package where you can setup your calls etc this is done manually so that you can use proper goalng structs. You can part it out into different files, you can either place it in your main or you can have it separatly
func WithJSONSaveState(fileName string, llmangoManager *llmango.LLMangoManager) (*llmango.LLMangoManager, error) {
	if fileName == "" {
		fileName = "mango.json"
	}

	//setup savestate Func
	var saveStateFunc func() error = func() error {
		log.Printf("JSON SAVE FUNCTION TRIGGERING for LLMangoManager: %p", llmangoManager)
		return jsonSaveStateFunc(llmangoManager, fileName)
	}
	llmangoManager.SaveState = saveStateFunc

	// Initialize empty config structure
	config := &mangoConfigFile{
		Goals:   make(map[string]*llmango.GoalInfo),
		Prompts: make(map[string]*llmango.Prompt),
	}

	file, err := os.Open(fileName)

	if os.IsNotExist(err) {
		file, err = os.Create(fileName)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		if err := json.NewEncoder(file).Encode(config); err != nil {
			return nil, err
		}
		log.Printf("INFO: MANGO: new empty config file created at %s", fileName)
		return llmangoManager, nil
	}

	if err != nil {
		return nil, err
	}

	defer file.Close()

	// Decode file contents into config
	if err := json.NewDecoder(file).Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	if config.Prompts == nil {
		config.Prompts = make(map[string]*llmango.Prompt)
	}

	if config.Goals == nil {
		config.Goals = make(map[string]*llmango.GoalInfo)
	}

	//initialize prompts
	if llmangoManager.Prompts == nil {
		llmangoManager.Prompts = make(map[string]*llmango.Prompt)
	}

	// Ensure prompt UIDs match their map keys
	for key, prompt := range config.Prompts {
		if prompt.UID == "" {
			prompt.UID = key
		} else if prompt.UID != key {
			log.Printf("INFO: MANGO: Updating prompt UID from %s to %s to match key", prompt.UID, key)
			prompt.UID = key
		}
	}

	maps.Copy(llmangoManager.Prompts, config.Prompts)

	loadConfig(llmangoManager, config.Goals)

	// Save the state to persist any updates
	err = llmangoManager.SaveState()
	if err != nil {
		log.Printf("ERROR: MANGO: Failed to save state after loading: %v", err)
	}

	return llmangoManager, nil
}

// this will parse the config if there currently is one and load the prompts and solutions into the object
func loadConfig(m *llmango.LLMangoManager, fileGoalsInfo map[string]*llmango.GoalInfo) error {
	if fileGoalsInfo == nil {
		return nil
	}

	// Iterate through config and update goals
	for uid, info := range fileGoalsInfo {
		if goalAny, exists := m.Goals[uid]; exists {
			if goal, ok := goalAny.(*llmango.Goal[any, any]); ok {
				// Update UID if needed
				if goal.UID != uid {
					goal.UID = uid
				}

				// Update title and description
				if info.Title != "" {
					goal.Title = info.Title
				}
				if info.Description != "" {
					goal.Description = info.Description
				}
			} else {
				log.Printf("ERROR: MANGO: Could not type assert goal %s as *llmango.Goal[any, any], actual type: %T", uid, goalAny)
			}
		} else {
			log.Printf("WARN: MANGO: Goal %s from file not found in manager", uid)
		}
	}

	return nil
}

// updateTimestamps sets CreatedAt and UpdatedAt fields to current Unix time if they are empty (0)
func updateTimestamps(m *llmango.LLMangoManager) {
	currentTime := int(time.Now().Unix())
	log.Printf("INFO: MANGO: Setting timestamps to %d for empty values", currentTime)

	// Update timestamps for prompts using map keys
	for key := range m.Prompts {
		prompt := m.Prompts[key]
		log.Printf("INFO: MANGO: Prompt %s has CreatedAt=%d, UpdatedAt=%d", key, prompt.CreatedAt, prompt.UpdatedAt)
		if m.Prompts[key].CreatedAt == 0 {
			m.Prompts[key].CreatedAt = currentTime
			log.Printf("INFO: MANGO: Updated CreatedAt for prompt %s", key)
		}
		if m.Prompts[key].UpdatedAt == 0 {
			m.Prompts[key].UpdatedAt = currentTime
			log.Printf("INFO: MANGO: Updated UpdatedAt for prompt %s", key)
		}
	}

	// Update timestamps for goals using type assertion
	for key := range m.Goals {
		log.Printf("INFO: MANGO: Processing goal %s", key)
		if goal, ok := m.Goals[key].(*llmango.Goal[any, any]); ok {
			log.Printf("INFO: MANGO: Goal %s has CreatedAt=%d, UpdatedAt=%d", key, goal.CreatedAt, goal.UpdatedAt)
			if goal.CreatedAt == 0 {
				goal.CreatedAt = currentTime
				log.Printf("INFO: MANGO: Updated CreatedAt for goal %s", key)
			}
			if goal.UpdatedAt == 0 {
				goal.UpdatedAt = currentTime
				log.Printf("INFO: MANGO: Updated UpdatedAt for goal %s", key)
			}
		} else {
			log.Printf("ERROR: MANGO: Could not type assert goal %s as *llmango.Goal[any, any]", key)
			// Try to find the actual type for debugging
			goalType := fmt.Sprintf("%T", m.Goals[key])
			log.Printf("INFO: MANGO: Goal %s has actual type %s", key, goalType)
		}
	}
}
