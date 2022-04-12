package carrycommits

import (
	"bufio"
	"fmt"
	"os"
)

type Reader interface {
	Read() ([]*Record, error)
}

func NewReaderFromFile(fpath string) (Reader, error) {
	stat, err := os.Stat(fpath)
	if err != nil {
		return nil, fmt.Errorf("error loading file %q - %w", fpath, err)
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("is a directory %q", err)
	}

	return &reader{fpath: fpath}, nil
}

type Record struct {
	SHA               string
	CommitType        string
	Message           string
	MessageWithPrefix string
	OpenShiftCommit   string
	UpstreamPR        string
}

func (r *Record) String() string {
	return fmt.Sprintf("%s, %s, %s, %s, %q", r.SHA, r.CommitType, r.OpenShiftCommit, r.UpstreamPR, r.Message)
}

func (r *Record) ShortString() string {
	return fmt.Sprintf("%s(%s): %s - %s", r.SHA, r.CommitType, r.Message, r.OpenShiftCommit)
}

type reader struct {
	fpath string
}

func (r *reader) Read() ([]*Record, error) {
	file, err := os.Open(r.fpath)
	if err != nil {
		return nil, fmt.Errorf("error loading file %q - %w", r.fpath, err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	records := make([]*Record, 0)
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
