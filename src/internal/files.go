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
