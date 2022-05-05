package carry

import (
	"bufio"
	"fmt"
	"os"
)

type CommitSummary struct {
	SHA                         string
	EffectiveType, OriginalType string
	Message                     string
	MessageWithPrefix           string
	OpenShiftCommit             string
	UpstreamPR                  string
}

func (r *CommitSummary) String() string {
	return fmt.Sprintf("%s(%s): %s - %s", r.SHA, r.EffectiveType, r.Message, r.OpenShiftCommit)
}

type csvReader struct {
	fpath string
}

func (r *csvReader) Read() ([]*CommitSummary, error) {
	file, err := os.Open(r.fpath)
	if err != nil {
		return nil, fmt.Errorf("error loading file %q - %w", r.fpath, err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	records := make([]*CommitSummary, 0)
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
