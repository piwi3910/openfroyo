package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadStack loads a stack from a YAML file
func LoadStack(path string) (*Stack, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read stack file %s: %w", path, err)
	}

	var stack Stack
	if err := yaml.Unmarshal(data, &stack); err != nil {
		return nil, fmt.Errorf("failed to parse stack file %s: %w", path, err)
	}

	// Set defaults
	for i := range stack.Run {
		if stack.Run[i].Strategy == "" {
			stack.Run[i].Strategy = "serial"
		}
	}

	return &stack, nil
}
