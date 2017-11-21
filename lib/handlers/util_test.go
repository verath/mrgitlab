package handlers

import "testing"

func TestMarkdownQuote(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{
			text:     "abc",
			expected: "> abc",
		},
		{
			text:     "abc\ndef\n>xyz",
			expected: "> abc\n> def\n> >xyz",
		},
		// Empty / only whitespace strings
		{
			text:     "",
			expected: "",
		},
		{
			text:     "  \n\t ",
			expected: "",
		},
	}

	for _, test := range tests {
		actual := markdownQuote(test.text)
		if actual != test.expected {
			t.Errorf(""+
				"markdownQuote\n"+
				"\t text:\n"+
				"'%s'\n"+
				"\t expected:\n"+
				"'%s'\n"+
				"\t actual:\n"+
				"'%s'\n",
				test.text, test.expected, actual)
		}
	}
}

func TestFilterGitLabReferences(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{"abc\ndef", "abc\ndef"},
		{"", ""},
		// @ user mention
		{"@petere", "`@`petere"},
		{"aa@petere", "aa`@`petere"},
		{"aa@", "aa@"},
		// # issue
		{"#123", "`#`123"},
		{"#abc", "#abc"},
		{"123#", "123#"},
		// ! merge request
		{"!123", "`!`123"},
		{"!abc", "!abc"},
		{"123!", "123!"},
		// $ snippet
		{"$123", "`$`123"},
		{"$abc", "$abc"},
		{"123$", "123$"},
		// ~ label
		{"~123", "`~`123"},
		{"~abc", "`~`abc"},
		{"~\"feature request\"", "`~`\"feature request\""},
		{"123~", "123~"},
		// % milestone
		{"%123", "`%`123"},
		{"%abc", "`%`abc"},
		{"%\"release candidate\"", "`%`\"release candidate\""},
		{"123%", "123%"},
	}
	for _, test := range tests {
		actual := filterGitLabReferences(test.text)
		if actual != test.expected {
			t.Errorf(""+
				"filterGitLabReferences\n"+
				"\t text:\n"+
				"'%s'\n"+
				"\t expected:\n"+
				"'%s'\n"+
				"\t actual:\n"+
				"'%s'\n",
				test.text, test.expected, actual)
		}
	}
}
