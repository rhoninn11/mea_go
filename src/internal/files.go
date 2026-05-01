package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func PngFilename(basename string) string {
	return fmt.Sprintf("%s.png", basename)
}

func YamlFilename(basename string) string {
	return fmt.Sprintf("%s.yaml", basename)
}

func SaveAsYAML(path string, v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func LoadFromYAML[T any](path string) (*T, error) {
	bArr, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("file open failed | %w", err)
	}

	var val T
	err = yaml.Unmarshal(bArr, &val)
	return &val, err
}
