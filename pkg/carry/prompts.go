package carry

import (
	"fmt"
	"os"

	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

type Prompt interface {
	GetAnswer(dropSHA string) string
}

type prompt struct {
	answers map[string]string
}

func (p *prompt) GetAnswer(dropSHA string) string {
	return p.answers[dropSHA]
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
		Prompts map[string]string `json:"prompts,omitempty"`
	}{}

	if err := decoder.Decode(o); err != nil {
		return nil, fmt.Errorf("failed to decode prompts from %q - %w", fpath, err)
	}

	return &prompt{answers: o.Prompts}, nil
}
