package carry

import (
	"bufio"
	"fmt"
	"os"
)

type CommitReader interface {
	Read() ([]*Commit, error)
}

func NewCommitReaderFromFile(fpath string) (CommitReader, error) {
	return &reader{fpath: fpath}, nil
}

type Commit struct {
	SHA               string
	CommitType        string
	Message           string
	MessageWithPrefix string
	OpenShiftCommit   string
	UpstreamPR        string
}

func (r *Commit) String() string {
	return fmt.Sprintf("%s, %s, %s, %s, %q", r.SHA, r.CommitType, r.OpenShiftCommit, r.UpstreamPR, r.Message)
}

func (r *Commit) ShortString() string {
	return fmt.Sprintf("%s(%s): %s - %s", r.SHA, r.CommitType, r.Message, r.OpenShiftCommit)
}

type reader struct {
	fpath string
}

func (r *reader) Read() ([]*Commit, error) {
	file, err := os.Open(r.fpath)
	if err != nil {
		return nil, fmt.Errorf("error loading file %q - %w", r.fpath, err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	records := make([]*Commit, 0)
	// we assume first line is not the header
	for scanner.Scan() {
		record, err := parse(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("parsing failed: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}
