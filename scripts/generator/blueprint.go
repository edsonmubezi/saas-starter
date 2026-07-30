package generator

import (
	"encoding/json"
	"fmt"
	"os"
)

type Blueprint struct {
	Entity      string            `json:"entity"`
	Table       string            `json:"table"`
	Fields      map[string]string `json:"fields"`
	CreateInput map[string]any    `json:"create_input"`
}

func LoadBlueprint(path string) (*Blueprint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read blueprint: %w", err)
	}

	var bp Blueprint
	if err := json.Unmarshal(data, &bp); err != nil {
		return nil, fmt.Errorf("failed to parse blueprint JSON: %w", err)
	}

	return &bp, nil
}
