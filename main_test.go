package main

import "testing"

func TestYoutrackBranchNameFilter(t *testing.T) {
	tests := []struct {
		branchName string
		expectedID string
	}{
		{"feature/xyz982", "XYZ-982"},
		{"feature/XYZ982", "XYZ-982"},
		{"feature/XYZ-982", ""},
		{"XYZ-982", ""},
		{"feature/XYZ982_some_feature", "XYZ-982"},
		{"feature/some_feature", ""},
		{"feature/some_xyz982_feature", ""},
		{"release-fix/XYZ982_some_feature", "XYZ-982"},
	}

	for _, test := range tests {
		actual := youtrackBranchNameFilter(test.branchName)
		if actual != test.expectedID {
			t.Errorf("expected '%s' for branch name '%s', got: '%s'",
				test.expectedID, test.branchName, actual)
		}
	}
}
