package apply

import (
	"encoding/json"
	"fmt"
	"github.com/tkashem/rebase/pkg/carry"
	"k8s.io/klog/v2"
	"os"
)

type Overrider interface {
	Override([]*carry.Commit)
}

func newOverrider(fpath string) (Overrider, error) {
	if len(fpath) == 0 {
		return noOverride{}, nil
	}

	return newOverriderFromFile(fpath)
}

type noOverride struct{}

func (noOverride) Override(commits []*carry.Commit) {
	klog.InfoS("override: none specified")
}

type Overrides struct {
	Overrides []Override `json:"overrides,omitempty"`
}

type Override struct {
	SHA string `json:"sha,omitempty"`
	Do  string `json:"do,omitempty"`
}

func (o *Override) String() string { return fmt.Sprintf("sha: %s, action: %s", o.SHA, o.Do) }

type overrider struct {
	overrides []Override
}

func newOverriderFromFile(fpath string) (*overrider, error) {
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
	decoder := json.NewDecoder(file)

	o := &Overrides{}
	if err := decoder.Decode(o); err != nil {
		return nil, fmt.Errorf("failed to decode overrides from %q - %w", fpath, err)
	}

	return &overrider{overrides: o.Overrides}, nil
}

func (o *overrider) Override(commits []*carry.Commit) {
	o.info()
	overrides := toMap(o.overrides)

	for i := range commits {
		commit := commits[i]
		if override, ok := overrides[commit.SHA]; ok {
			klog.Infof("override(%s): %s->%s", commit.SHA, commit.CommitType, override.Do)
			commit.CommitType = override.Do
		}
	}
}

func (o *overrider) info() {
	klog.Infof("override: %d specified", len(o.overrides))
	for i := range o.overrides {
		klog.Infof("override: %s", o.overrides[i].String())
	}
}

func toMap(overrides []Override) map[string]Override {
	m := map[string]Override{}
	for i := range overrides {
		override := overrides[i]
		m[override.SHA] = override
	}
	return m
}
