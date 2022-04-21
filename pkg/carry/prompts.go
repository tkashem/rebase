package carry

import (
	"fmt"
	"os"

	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

type Prompt interface {
	ShouldDrop(dropSHA string) bool
}

type prompt struct {
	answers []Override
}

func (p *prompt) ShouldDrop(dropSHA string) bool {
	for _, override := range p.answers {
		if override.SHA == dropSHA && override.Do == "drop" {
			return true
		}
	}
	return false
}

func NewPromptsFromFile(fpath string) (Prompt, error) {
	stat, err := os.Stat(fpath)
	if err != nil {
		return nil, fmt.Errorf("error loading file %q - %w", fpath, err)
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("is a directory %q", err)
	}

	file, err := os.Open(fpath)
	if err != nil {
		return nil, fmt.Errorf("error loading file %q - %w", fpath, err)
	}

	defer file.Close()

	decoder := utilyaml.NewYAMLToJSONDecoder(file)
	o := &struct {
		Overrides []Override `json:"overrides,omitempty"`
	}{}

	if err := decoder.Decode(o); err != nil {
		return nil, fmt.Errorf("failed to decode prompts from %q - %w", fpath, err)
	}

	return &prompt{answers: o.Overrides}, nil
}
