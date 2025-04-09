package llmangosavestate

import (
	"encoding/json"
	"log"
	"maps"
	"os"

	"github.com/llmang/llmango/llmango"
)

type mangoConfigFile struct {
	PinnedModels []string                     `json:"pinnedModels"`
	Goals        map[string]*llmango.GoalInfo `json:"goals"`
	Prompts      map[string]*llmango.Prompt   `json:"prompts"`
}

func JSONSaveStateFunc(mango *llmango.LLMangoManager, fileName string) error {
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
		return JSONSaveStateFunc(mango, fileName)
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

	LoadConfig(mango, config.Goals)

	return mango, nil
}

// this will parse the config if there currently is one and load the prompts and solutions into the object
func LoadConfig(m *llmango.LLMangoManager, fileGoalsInfo map[string]*llmango.GoalInfo) error {
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
			maps.Copy(goal.Solutions, info.Solutions)
		}
	}
	return nil
}
