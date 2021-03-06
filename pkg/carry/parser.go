package carry

import (
	"fmt"
	"strconv"
	"strings"
)

func parse(s string) (*CommitSummary, error) {
	// an example commit log
	// {commit-sha}\t\t\tUPSTREAM: {type}: {message}\t{openshift-commit}\t{upstream-pr}
	// type == 'carry|revert|drop|{upstream-pr-number}'
	s = strings.TrimSpace(s)
	split := strings.SplitN(s, "\t\t\t", 2)
	if len(split) < 2 {
		return nil, fmt.Errorf("malformed commit log, separator: %q not found", "\t\t\t")
	}
	summary := &CommitSummary{
		SHA: split[0],
	}

	// remaining: 'UPSTREAM: {type}: {message}\t{openshift-commit}\t{upstream-pr}'
	s = split[1]
	s = strings.TrimSpace(s)
	split = strings.SplitN(s, "\t", 2)
	if len(split) < 2 {
		return nil, fmt.Errorf("malformed commit log, separator: %q not found", "\t")
	}
	summary.MessageWithPrefix = split[0]

	// {openshift-commit}\t{upstream-pr}'
	split = strings.Split(split[1], "\t")
	if len(split) < 1 {
		return nil, fmt.Errorf("malformed commit log, missing openshift commit url")
	}
	summary.OpenShiftCommit = split[0]
	if len(split) == 2 {
		summary.UpstreamPR = split[1]
	}

	// extract type
	split = strings.SplitN(summary.MessageWithPrefix, ":", 3)
	if len(split) < 3 {
		return nil, fmt.Errorf("malformed commit log, did not find the commit SHA")
	}
	if split[0] != "UPSTREAM" {
		return nil, fmt.Errorf("malformed commit log, missing 'UPSTREAM' prefix")
	}

	commitType, err := sanitize(split[1])
	if err != nil {
		return nil, fmt.Errorf("malformed commit log, unknown commit type: %w", err)
	}

	// both effective and original type are equal when
	// a summary object is initialized.
	summary.EffectiveType = commitType
	summary.OriginalType = commitType

	summary.Message = strings.TrimSpace(split[2])
	return summary, nil
}

func sanitize(t string) (string, error) {
	t = strings.TrimSpace(t)
	t = strings.TrimPrefix(t, "<")
	t = strings.TrimSuffix(t, ">")
	t = strings.ToLower(t)

	switch {
	case t == "revert" || t == "drop" || t == "carry":
	default:
		if _, err := strconv.Atoi(t); err != nil {
			return "", fmt.Errorf("unknown type: %s", t)
		}
	}

	return t, nil
}
