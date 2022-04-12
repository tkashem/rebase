package carrycommits

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		errShouldContain string
		expected         *Record
	}{
		{
			name:             "revert",
			input:            "\td032c6e6463\t\t\tUPSTREAM: revert: <carry>: Unskip OCP SDN related tests\thttps://github.com/openshift/kubernetes/commit/d032c6e6463?w=1",
			errShouldContain: "",
			expected: &Record{
				SHA:               "d032c6e6463",
				CommitType:        "revert",
				Message:           "<carry>: Unskip OCP SDN related tests",
				MessageWithPrefix: "UPSTREAM: revert: <carry>: Unskip OCP SDN related tests",
				OpenShiftCommit:   "https://github.com/openshift/kubernetes/commit/d032c6e6463?w=1",
				UpstreamPR:        "",
			},
		},
		{
			name:             "pick",
			input:            "\tdb4c4bbd6d6\t\t\tUPSTREAM: 107900: Add an e2e test for updating a static pod while it restarts\thttps://github.com/openshift/kubernetes/commit/db4c4bbd6d6?w=1\thttps://github.com/kubernetes/kubernetes/pull/107900",
			errShouldContain: "",
			expected: &Record{
				SHA:               "db4c4bbd6d6",
				CommitType:        "107900",
				Message:           "Add an e2e test for updating a static pod while it restarts",
				MessageWithPrefix: "UPSTREAM: 107900: Add an e2e test for updating a static pod while it restarts",
				OpenShiftCommit:   "https://github.com/openshift/kubernetes/commit/db4c4bbd6d6?w=1",
				UpstreamPR:        "https://github.com/kubernetes/kubernetes/pull/107900",
			},
		},
		{
			name:             "carry",
			input:            "\td7b268fffba\t\t\tUPSTREAM: <carry>: use console-public config map for console redirect\thttps://github.com/openshift/kubernetes/commit/d7b268fffba?w=1",
			errShouldContain: "",
			expected: &Record{
				SHA:               "d7b268fffba",
				CommitType:        "carry",
				Message:           "use console-public config map for console redirect",
				MessageWithPrefix: "UPSTREAM: <carry>: use console-public config map for console redirect",
				OpenShiftCommit:   "https://github.com/openshift/kubernetes/commit/d7b268fffba?w=1",
				UpstreamPR:        "",
			},
		},
		{
			name:             "drop",
			input:            "\tc77caa826a0\t\t\tUPSTREAM: <drop>: update vendor files\thttps://github.com/openshift/kubernetes/commit/c77caa826a0?w=1",
			errShouldContain: "",
			expected: &Record{
				SHA:               "c77caa826a0",
				CommitType:        "drop",
				Message:           "update vendor files",
				MessageWithPrefix: "UPSTREAM: <drop>: update vendor files",
				OpenShiftCommit:   "https://github.com/openshift/kubernetes/commit/c77caa826a0?w=1",
				UpstreamPR:        "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record, err := parse(test.input)
			switch {
			case len(test.errShouldContain) > 0:
				if err == nil {
					t.Errorf("Expected error: %s, but got none", test.errShouldContain)
					return
				}

				if !strings.Contains(err.Error(), test.errShouldContain) {
					t.Errorf("Expected error: %s, but got %v", test.errShouldContain, err)
				}
			case test.expected != nil:
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
					return
				}

				if !reflect.DeepEqual(test.expected, record) {
					t.Errorf("Expected commit log to match: %s", cmp.Diff(test.expected, record))
				}
			}
		})
	}
}
