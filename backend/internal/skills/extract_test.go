package skills

import (
	"testing"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single skill Go",
			input:    "Looking for a Go developer",
			expected: []string{"Go"},
		},
		{
			name:     "Go substring avoidance",
			input:    "We go to the park",
			expected: []string{},
		},
		{
			name:     "Multiple skills",
			input:    "Experience with React, Node.js and PostgreSQL required.",
			expected: []string{"React", "Node.js", "PostgreSQL"},
		},
		{
			name:     "Case insensitivity",
			input:    "java and python expertise",
			expected: []string{"Java", "Python"},
		},
		{
			name:     "Alternate names regex",
			input:    "Knowledge of Golang and JS",
			expected: []string{"Go", "JavaScript"},
		},
		{
			name:     "Special characters",
			input:    "C++ and C# developer",
			expected: []string{"C++", "C#"},
		},
		{
			name:     "No skills",
			input:    "Project manager with agile experience",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Extract(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("Extract() returned %d skills, expected %d. Got: %v", len(got), len(tt.expected), got)
			}

			// Check content regardless of order
			for _, exp := range tt.expected {
				found := false
				for _, g := range got {
					if g == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Extract() missing expected skill: %s", exp)
				}
			}
		})
	}
}
