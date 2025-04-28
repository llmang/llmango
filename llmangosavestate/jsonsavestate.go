package llmangosavestate

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/llmang/llmango/llmango"
)

// goalForJSON is an intermediate struct for JSON marshalling/unmarshalling.
// It excludes fields that should not be persisted or are functions.
type goalForJSON struct {
	UID         string `json:"UID"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   int    `json:"createdAt"`
	UpdatedAt   int    `json:"updatedAt"`
	// InputExample and OutputExample removed as they are hardcoded
}

// mangoConfigFile defines the structure of the JSON configuration file.
type mangoConfigFile struct {
	Goals   map[string]*goalForJSON    `json:"goals"`
	Prompts map[string]*llmango.Prompt `json:"prompts"`
}

// jsonSaveStateFunc saves the current state of the LLMangoManager to a JSON file.
func jsonSaveStateFunc(mango *llmango.LLMangoManager, fileName string) error {
	configToSave := mangoConfigFile{
		Goals:   make(map[string]*goalForJSON),
		Prompts: make(map[string]*llmango.Prompt),
		// PinnedModels could be fetched from somewhere if needed, or left empty
	}

	// Populate Goals for saving
	// Corrected: Iterate over items retrieved using GetAll() AND removed Input/Output Examples
	for uid, goal := range mango.Goals.Snapshot() {
		if goal == nil {
			continue // skip nil entries if any
		}
		configToSave.Goals[uid] = &goalForJSON{
			UID:         goal.UID,
			Title:       goal.Title,
			Description: goal.Description,
			CreatedAt:   goal.CreatedAt,
			UpdatedAt:   goal.UpdatedAt,
			// InputExample and OutputExample removed
		}
	}

	// Populate Prompts for saving
	for uid, prompt := range mango.Prompts.Snapshot() {
		if prompt == nil {
			continue // skip nil entries if any
		}
		// Ensure prompt UID matches key before saving
		if prompt.UID != uid {
			log.Printf("WARN: MANGO SAVESTATE: Prompt UID %s does not match map key %s during save. Using map key.", prompt.UID, uid)
			prompt.UID = uid // Correct it before saving? Or just log? Let's log for now.
		}
		configToSave.Prompts[uid] = prompt
	}

	// Write updated config to file
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print JSON
	if err := encoder.Encode(configToSave); err != nil {
		return err
	}
	log.Printf("INFO: MANGO SAVESTATE: State saved to %s", fileName)
	return nil
}

// WithJSONSaveState configures the LLMangoManager to use a JSON file for saving and loading state.
func WithJSONSaveState(fileName string, llmangoManager *llmango.LLMangoManager) (*llmango.LLMangoManager, error) {
	if fileName == "" {
		fileName = "mango.json"
	}

	// Setup saveStateFunc
	saveStateFunc := func() error {
		return jsonSaveStateFunc(llmangoManager, fileName)
	}
	llmangoManager.SaveState = saveStateFunc

	// Attempt to open the file
	file, err := os.Open(fileName)
	if os.IsNotExist(err) {
		// File doesn't exist, create an empty one and return
		log.Printf("INFO: MANGO SAVESTATE: Config file '%s' not found, creating an empty one.", fileName)
		initialConfig := &mangoConfigFile{
			Goals:   make(map[string]*goalForJSON),
			Prompts: make(map[string]*llmango.Prompt),
		}
		emptyFile, createErr := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if createErr != nil {
			return nil, createErr
		}
		defer emptyFile.Close()
		encoder := json.NewEncoder(emptyFile)
		encoder.SetIndent("", "  ")
		if encodeErr := encoder.Encode(initialConfig); encodeErr != nil {
			return nil, encodeErr
		}
		// No data to load, just return the manager
		return llmangoManager, nil
	} else if err != nil {
		// Other error opening file
		return nil, err
	}
	defer file.Close()

	// Read file content
	fileBytes, readErr := io.ReadAll(file)
	if readErr != nil {
		return nil, readErr
	}

	// Decode file contents into config
	var loadedConfig mangoConfigFile
	if err := json.Unmarshal(fileBytes, &loadedConfig); err != nil {
		log.Printf("ERROR: MANGO SAVESTATE: Failed to decode config file '%s': %v. Check JSON validity.", fileName, err)
		// Decide whether to return error or proceed with default empty config
		// For now, let's return the error.
		return nil, err // fmt.Errorf("failed to decode config file '%s': %w", fileName, err)
	}

	// --- Load Goals ---
	if loadedConfig.Goals != nil {
		goalsToLoad := make([]*llmango.Goal, 0, len(loadedConfig.Goals))
		for uid, gj := range loadedConfig.Goals {
			if gj == nil {
				log.Printf("WARN: MANGO SAVESTATE: Skipping nil goal entry with key %s in config file.", uid)
				continue
			}
			if gj.UID == "" {
				gj.UID = uid // Assign map key as UID if missing in JSON object
				log.Printf("WARN: MANGO SAVESTATE: Assigning map key '%s' as UID for goal with empty UID field.", uid)
			} else if gj.UID != uid {
				log.Printf("WARN: MANGO SAVESTATE: Goal UID '%s' in JSON object does not match map key '%s'. Using UID from JSON object.", gj.UID, uid)
				// Potentially problematic, but we'll trust the JSON object's UID primarily.
			}

			// Create the llmango.Goal struct. Input/Output validators are not persisted.
			goal := &llmango.Goal{
				UID:         gj.UID,
				Title:       gj.Title,
				Description: gj.Description,
				CreatedAt:   gj.CreatedAt,
				UpdatedAt:   gj.UpdatedAt,
				// PromptUIDs will be populated by AddPrompts later
				// InputOutput field is omitted here; AddOrUpdateGoals preserves the existing one
			}
			goalsToLoad = append(goalsToLoad, goal)
		}
		// Use AddOrUpdateGoals which preserves existing Input/Output handlers if goal already exists
		llmangoManager.AddOrUpdateGoals(goalsToLoad...)
		log.Printf("INFO: MANGO SAVESTATE: Loaded/Updated %d goals from %s", len(goalsToLoad), fileName)
	} else {
		log.Printf("INFO: MANGO SAVESTATE: No 'goals' section found or it's empty in %s", fileName)
	}

	// --- Load Prompts ---
	if loadedConfig.Prompts != nil {
		promptsToLoad := make([]*llmango.Prompt, 0, len(loadedConfig.Prompts))
		for uid, p := range loadedConfig.Prompts {
			if p == nil {
				log.Printf("WARN: MANGO SAVESTATE: Skipping nil prompt entry with key %s in config file.", uid)
				continue
			}
			if p.UID == "" {
				p.UID = uid // Assign map key as UID if missing
				log.Printf("WARN: MANGO SAVESTATE: Assigning map key '%s' as UID for prompt with empty UID field.", uid)
			} else if p.UID != uid {
				log.Printf("WARN: MANGO SAVESTATE: Prompt UID '%s' in JSON object does not match map key '%s'. Using UID from JSON object.", p.UID, uid)
			}
			// GoalUID association will be handled by AddPrompts
			promptsToLoad = append(promptsToLoad, p)
		}
		// AddPrompts handles associating prompts with goals
		llmangoManager.AddPrompts(promptsToLoad...)
		log.Printf("INFO: MANGO SAVESTATE: Loaded/Updated %d prompts from %s", len(promptsToLoad), fileName)
	} else {
		log.Printf("INFO: MANGO SAVESTATE: No 'prompts' section found or it's empty in %s", fileName)
	}

	// Save state immediately after loading to potentially fix UIDs or add timestamps if logic requires it
	// (Though AddOrUpdateGoals/AddPrompts should handle timestamps now)
	if err := llmangoManager.SaveState(); err != nil {
		log.Printf("ERROR: MANGO SAVESTATE: Failed to save state immediately after loading: %v", err)
		// Decide if this should be a fatal error
	}

	return llmangoManager, nil
}
