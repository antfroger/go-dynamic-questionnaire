package go_dynamic_questionnaire

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// Loader defines the interface for loading questionnaire configurations.
// Each loader implementation is responsible for parsing a specific format
// (YAML, JSON, etc.) and populating a given questionnaire struct.
//
// The Loader interface allows the system to be easily extended to support
// additional configuration formats without modifying the core questionnaire logic.
type Loader interface {
	// Load parses the configuration data and populates the provided questionnaire struct.
	// The data parameter can be either a file path (string) or raw content ([]byte).
	// The q parameter is a pointer to the questionnaire struct to be populated.
	//
	// Parameters:
	//   data: Either a file path or raw configuration content
	//   q: Pointer to the questionnaire struct to populate
	//
	// Returns:
	//   error: Parsing errors, file reading errors, or validation errors
	Load(data interface{}, q *questionnaire) error
}

// loadConfig loads a questionnaire configuration from either a file path or content.
// This function handles all the internal logic of selecting the appropriate loader
// and parsing the configuration into the provided questionnaire struct.
//
// Parameters:
//
//	config: Either a file path (string) or configuration content ([]byte)
//	q: Pointer to questionnaire struct to populate
//
// Returns:
//
//	error: Configuration errors, file reading errors, parsing errors, or validation errors
func loadConfig[T config](cfg T, q *questionnaire) error {
	loaderInstance, err := getLoaderForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to get loader: %w", err)
	}

	if err := loaderInstance.Load(cfg, q); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return nil
}

// getLoaderForConfig determines the appropriate loader based on the configuration data.
// For file paths, it uses the file extension. For byte arrays, it attempts to detect
// the format by examining the content.
func getLoaderForConfig(cfg interface{}) (Loader, error) {
	switch v := cfg.(type) {
	case string:
		// Determine loader based on file extension
		switch ext := strings.ToLower(filepath.Ext(v)); ext {
		case ".yaml", ".yml":
			return &yamlLoader{}, nil
		case ".json":
			return &jsonLoader{}, nil
		default:
			return nil, fmt.Errorf("unsupported file extension %s: expected .yaml, .yml, or .json", ext)
		}
	case []byte:
		// Try to detect format by examining content
		content := strings.TrimSpace(string(v))
		if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
			return &jsonLoader{}, nil
		}
		// Default to YAML for backward compatibility
		return &yamlLoader{}, nil
	default:
		return nil, fmt.Errorf("unsupported config type: expected string (file path) or []byte (content), got %T", cfg)
	}
}

// yamlLoader implements the Loader interface for YAML configuration files.
type yamlLoader struct{}

// Load parses YAML configuration data and populates the provided questionnaire struct.
func (l *yamlLoader) Load(data interface{}, q *questionnaire) error {
	return loadWithUnmarshaler(data, q, yaml.Unmarshal)
}

// jsonLoader implements the Loader interface for JSON configuration files.
type jsonLoader struct{}

// Load parses JSON configuration data and populates the provided questionnaire struct.
func (l *jsonLoader) Load(data interface{}, q *questionnaire) error {
	return loadWithUnmarshaler(data, q, json.Unmarshal)
}

// unmarshalFunc defines the signature for unmarshal functions.
// This allows different format parsers (JSON, YAML, etc.) to be used interchangeably.
type unmarshalFunc func([]byte, interface{}) error

// loadWithUnmarshaler provides common loading logic for different configuration formats.
// It handles reading data from files or byte arrays, unmarshaling with the provided
// unmarshal function, and validating the resulting questionnaire structure.
//
// Parameters:
//
//	data: Either a file path (string) or raw configuration content ([]byte)
//	q: Pointer to the questionnaire struct to populate
//	unmarshal: The unmarshal function specific to the format (json.Unmarshal, yaml.Unmarshal, etc.)
//
// Returns:
//
//	error: File reading errors, parsing errors, or validation errors
func loadWithUnmarshaler(data interface{}, q *questionnaire, unmarshal unmarshalFunc) error {
	var content []byte
	var err error

	switch v := data.(type) {
	case string:
		// Load from file
		content, err = os.ReadFile(v)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", v, err)
		}
	case []byte:
		// Load from byte array
		content = v
	default:
		return fmt.Errorf("unsupported data type for loader: %T", data)
	}

	// Unmarshal directly into the questionnaire struct
	if err := unmarshal(content, q); err != nil {
		return fmt.Errorf("failed to parse content: %w", err)
	}

	// Basic validation to ensure data structure is valid
	if err := validateLoadedQuestionnaire(q); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// validateLoadedQuestionnaire performs basic structural validation on the loaded questionnaire data.
// This is called by each loader after parsing to ensure the data structure is valid.
// Business logic validation (duplicate IDs, dependencies, etc.) is handled by the main validation.
func validateLoadedQuestionnaire(q *questionnaire) error {
	// Ensure slices are initialized (not nil)
	if q.Questions == nil {
		q.Questions = []question{}
	}
	if q.Remarks == nil {
		q.Remarks = []closingRemark{}
	}

	return nil
}
