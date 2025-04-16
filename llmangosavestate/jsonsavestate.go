package llmangosavestate

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"reflect"
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
	// --- Associate loaded prompts with goals using reflection ---
	promptUIDsForGoal := make(map[string]map[string]struct{}) // GoalUID -> set of PromptUIDs

	// Build the map from GoalUID to its associated PromptUIDs based on loaded prompts
	for _, prompt := range m.Prompts {
		if prompt.GoalUID != "" {
			if _, exists := promptUIDsForGoal[prompt.GoalUID]; !exists {
				promptUIDsForGoal[prompt.GoalUID] = make(map[string]struct{})
			}
			promptUIDsForGoal[prompt.GoalUID][prompt.UID] = struct{}{}
		}
	}

	// Iterate through the goals in the manager and update their PromptUIDs lists
	for goalUID, goalAny := range m.Goals {
		goalValue := reflect.ValueOf(goalAny)

		// Ensure it's a pointer to a struct
		if goalValue.Kind() != reflect.Ptr || goalValue.IsNil() {
			log.Printf("WARN: MANGO: Goal %s is not a non-nil pointer, skipping prompt association. Type: %T", goalUID, goalAny)
			continue
		}
		goalElem := goalValue.Elem()
		if goalElem.Kind() != reflect.Struct {
			log.Printf("WARN: MANGO: Goal %s is not a pointer to a struct, skipping prompt association. Type: %T", goalUID, goalAny)
			continue
		}

		if loadedPromptSet, goalHasPrompts := promptUIDsForGoal[goalUID]; goalHasPrompts {
			promptUIDsField := goalElem.FieldByName("PromptUIDs")
			if !promptUIDsField.IsValid() {
				log.Printf("WARN: MANGO: Goal %s (Type: %T) has no PromptUIDs field, skipping prompt association.", goalUID, goalAny)
				continue
			}
			if !promptUIDsField.CanSet() {
				log.Printf("WARN: MANGO: Cannot set PromptUIDs field for goal %s (Type: %T), skipping prompt association.", goalUID, goalAny)
				continue
			}
			if promptUIDsField.Kind() != reflect.Slice {
				log.Printf("WARN: MANGO: PromptUIDs field for goal %s (Type: %T) is not a slice, skipping prompt association.", goalUID, goalAny)
				continue
			}

			// Initialize PromptUIDs slice if nil
			if promptUIDsField.IsNil() {
				// Create a new slice of the appropriate type (string)
				newSlice := reflect.MakeSlice(promptUIDsField.Type(), 0, len(loadedPromptSet))
				promptUIDsField.Set(newSlice)
			}

			// Get current slice value
			currentPromptUIDsVal := promptUIDsField.Interface()
			currentPromptUIDs, ok := currentPromptUIDsVal.([]string)
			if !ok {
				log.Printf("ERROR: MANGO: Could not assert PromptUIDs field for goal %s (Type: %T) as []string.", goalUID, goalAny)
				continue
			}

			// Create a set of existing prompt UIDs for efficient lookup
			existingPromptSet := make(map[string]struct{}, len(currentPromptUIDs))
			for _, existingUID := range currentPromptUIDs {
				existingPromptSet[existingUID] = struct{}{}
			}

			// Append prompts from the loaded set if they are not already present
			updatedSlice := promptUIDsField // Start with the existing slice
			added := false
			for loadedPromptUID := range loadedPromptSet {
				if _, alreadyExists := existingPromptSet[loadedPromptUID]; !alreadyExists {
					// Append using reflection
					updatedSlice = reflect.Append(updatedSlice, reflect.ValueOf(loadedPromptUID))
					// log.Printf("INFO: MANGO: Associated prompt %s with goal %s", loadedPromptUID, goalUID)
					added = true
				}
			}

			// Only set if changed
			if added {
				promptUIDsField.Set(updatedSlice)
			}
		}
	}
	// --- End of Prompt Association Logic ---

	// --- Update Goal Metadata (Title, Description) using reflection ---
	if fileGoalsInfo == nil {
		return nil
	}

	// Iterate through config goal info and update corresponding manager goals
	for uid, info := range fileGoalsInfo {
		if goalAny, exists := m.Goals[uid]; exists {
			goalValue := reflect.ValueOf(goalAny)

			// Ensure it's a pointer to a struct
			if goalValue.Kind() != reflect.Ptr || goalValue.IsNil() {
				log.Printf("WARN: MANGO: Goal %s is not a non-nil pointer, skipping metadata update. Type: %T", uid, goalAny)
				continue
			}
			goalElem := goalValue.Elem()
			if goalElem.Kind() != reflect.Struct {
				log.Printf("WARN: MANGO: Goal %s is not a pointer to a struct, skipping metadata update. Type: %T", uid, goalAny)
				continue
			}

			// --- Update UID Field ---
			uidField := goalElem.FieldByName("UID")
			if uidField.IsValid() && uidField.CanSet() && uidField.Kind() == reflect.String {
				if uidField.String() != uid {
					log.Printf("WARN: MANGO: Goal UID mismatch for %s, updating manager goal UID.", uid)
					uidField.SetString(uid)
				}
			} else {
				log.Printf("WARN: MANGO: Cannot access or set UID field for goal %s (Type: %T).", uid, goalAny)
			}

			// --- Update Title Field ---
			if info.Title != "" {
				titleField := goalElem.FieldByName("Title")
				if titleField.IsValid() && titleField.CanSet() && titleField.Kind() == reflect.String {
					if titleField.String() != info.Title {
						titleField.SetString(info.Title)
						log.Printf("INFO: MANGO: Updated title for goal %s", uid)
					}
				} else {
					log.Printf("WARN: MANGO: Cannot access or set Title field for goal %s (Type: %T).", uid, goalAny)
				}
			}

			// --- Update Description Field ---
			if info.Description != "" {
				descField := goalElem.FieldByName("Description")
				if descField.IsValid() && descField.CanSet() && descField.Kind() == reflect.String {
					if descField.String() != info.Description {
						descField.SetString(info.Description)
						log.Printf("INFO: MANGO: Updated description for goal %s", uid)
					}
				} else {
					log.Printf("WARN: MANGO: Cannot access or set Description field for goal %s (Type: %T).", uid, goalAny)
				}
			}

		} else {
			log.Printf("WARN: MANGO: Goal %s from file not found in manager", uid)
		}
	}
	// --- End of Goal Metadata Update ---

	return nil
}

// updateTimestamps sets CreatedAt and UpdatedAt fields to current Unix time if they are empty (0)
// Uses reflection to handle generic Goal types
func updateTimestamps(m *llmango.LLMangoManager) {
	currentTime := int64(time.Now().Unix()) // Use int64 for time
	log.Printf("INFO: MANGO: Setting timestamps to %d for empty values", currentTime)

	// Update timestamps for prompts using map keys (Prompts are concrete types)
	for key := range m.Prompts {
		prompt := m.Prompts[key]
		if prompt.CreatedAt == 0 {
			prompt.CreatedAt = int(currentTime) // Assuming Prompt.CreatedAt is int
			log.Printf("INFO: MANGO: Updated CreatedAt for prompt %s", key)
		}
		if prompt.UpdatedAt == 0 {
			prompt.UpdatedAt = int(currentTime) // Assuming Prompt.UpdatedAt is int
			log.Printf("INFO: MANGO: Updated UpdatedAt for prompt %s", key)
		}
	}

	// Update timestamps for goals using reflection
	for key, goalAny := range m.Goals {
		goalValue := reflect.ValueOf(goalAny)

		// Ensure it's a pointer to a struct
		if goalValue.Kind() != reflect.Ptr || goalValue.IsNil() {
			log.Printf("WARN: MANGO: Goal %s is not a non-nil pointer, skipping timestamp update. Type: %T", key, goalAny)
			continue
		}
		goalElem := goalValue.Elem()
		if goalElem.Kind() != reflect.Struct {
			log.Printf("WARN: MANGO: Goal %s is not a pointer to a struct, skipping timestamp update. Type: %T", key, goalAny)
			continue
		}

		// Update CreatedAt
		createdAtField := goalElem.FieldByName("CreatedAt")
		if createdAtField.IsValid() && createdAtField.CanSet() {
			if createdAtField.Kind() == reflect.Int || createdAtField.Kind() == reflect.Int64 {
				if createdAtField.Int() == 0 {
					createdAtField.SetInt(currentTime)
					log.Printf("INFO: MANGO: Updated CreatedAt for goal %s", key)
				}
			} else {
				log.Printf("WARN: MANGO: CreatedAt field for goal %s (Type: %T) is not an int/int64.", key, goalAny)
			}
		} else {
			log.Printf("WARN: MANGO: Cannot access or set CreatedAt field for goal %s (Type: %T).", key, goalAny)
		}

		// Update UpdatedAt
		updatedAtField := goalElem.FieldByName("UpdatedAt")
		if updatedAtField.IsValid() && updatedAtField.CanSet() {
			if updatedAtField.Kind() == reflect.Int || updatedAtField.Kind() == reflect.Int64 {
				if updatedAtField.Int() == 0 {
					updatedAtField.SetInt(currentTime)
					log.Printf("INFO: MANGO: Updated UpdatedAt for goal %s", key)
				}
			} else {
				log.Printf("WARN: MANGO: UpdatedAt field for goal %s (Type: %T) is not an int/int64.", key, goalAny)
			}
		} else {
			log.Printf("WARN: MANGO: Cannot access or set UpdatedAt field for goal %s (Type: %T).", key, goalAny)
		}
	}
}
