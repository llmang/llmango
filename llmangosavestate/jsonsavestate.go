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
func WithJSONSaveState(fileName string, mango *llmango.LLMangoManager) (*llmango.LLMangoManager, error) {
	if fileName == "" {
		fileName = "mango.json"
	}

	//setup savestate Func
	var saveStateFunc func() error = func() error {
		return jsonSaveStateFunc(mango, fileName)
	}
	mango.SaveState = saveStateFunc

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
		return mango, nil
	}

	if err != nil {
		return nil, err
	}

	defer file.Close()

	if err := json.NewDecoder(file).Decode(config); err != nil {
		return nil, err
	}

	if config.Prompts == nil {
		config.Prompts = make(map[string]*llmango.Prompt)
	}

	if config.Goals == nil {
		config.Goals = make(map[string]*llmango.GoalInfo)
	}

	//initialize prompts
	if mango.Prompts == nil {
		mango.Prompts = make(map[string]*llmango.Prompt)
	}

	// Ensure prompt UIDs match their map keys
	for key, prompt := range config.Prompts {
		if prompt.UID == "" {
			prompt.UID = key
			log.Printf("INFO: MANGO: Setting empty prompt UID to match key: %s", key)
		} else if prompt.UID != key {
			log.Printf("INFO: MANGO: Updating prompt UID from %s to %s to match key", prompt.UID, key)
			prompt.UID = key
		}
	}

	maps.Copy(mango.Prompts, config.Prompts)

	loadConfig(mango, config.Goals)

	
	// updateTimestamps(mango)

	// Save the state to persist the timestamp updates
	err = mango.SaveState()
	if err != nil {
		log.Printf("ERROR: MANGO: Failed to save state after updating timestamps: %v", err)
	} else {
		log.Printf("INFO: MANGO: Successfully saved state after updating timestamps")
	}

	return mango, nil
}

// this will parse the config if there currently is one and load the prompts and solutions into the object
func loadConfig(m *llmango.LLMangoManager, fileGoalsInfo map[string]*llmango.GoalInfo) error {
	if fileGoalsInfo == nil {
		return nil
	}

	// Iterate through config and update goals
	for uid, info := range fileGoalsInfo {
		if goal, ok := m.Goals[uid].(*llmango.Goal[any, any]); ok {
			// Update UID if needed
			if goal.UID != uid {
				log.Printf("INFO: MANGO: Updating goal UID from %s to %s to match key", goal.UID, uid)
				goal.UID = uid
			}

			// Update title and description
			if info.Title != "" {
				goal.Title = info.Title
			}
			if info.Description != "" {
				goal.Description = info.Description
			}

			// Initialize and copy solutions
			if goal.Solutions == nil {
				goal.Solutions = make(map[string]*llmango.Solution)
			}
			if info.Solutions != nil {
				maps.Copy(goal.Solutions, info.Solutions)
			}
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

			// Update timestamps for solutions using map keys
			for solKey := range goal.Solutions {
				solution := goal.Solutions[solKey]
				log.Printf("INFO: MANGO: Solution %s in goal %s has CreatedAt=%d, UpdatedAt=%d",
					solKey, key, solution.CreatedAt, solution.UpdatedAt)
				if goal.Solutions[solKey].CreatedAt == 0 {
					goal.Solutions[solKey].CreatedAt = currentTime
					log.Printf("INFO: MANGO: Updated CreatedAt for solution %s in goal %s", solKey, key)
				}
				if goal.Solutions[solKey].UpdatedAt == 0 {
					goal.Solutions[solKey].UpdatedAt = currentTime
					log.Printf("INFO: MANGO: Updated UpdatedAt for solution %s in goal %s", solKey, key)
				}
			}
		} else {
			log.Printf("ERROR: MANGO: Could not type assert goal %s as *llmango.Goal[any, any]", key)
			// Try to find the actual type for debugging
			goalType := fmt.Sprintf("%T", m.Goals[key])
			log.Printf("INFO: MANGO: Goal %s has actual type %s", key, goalType)
		}
	}
}
