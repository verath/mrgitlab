package handlers

import (
	"context"
	"regexp"
	"strings"

	"github.com/verath/mrgitlab/lib/gitlab"
)

// MergeRequestHandlerFunc is a wrapper allowing a func to implement the
// MergeRequestHandler interface
type MergeRequestHandlerFunc func(context.Context, *gitlab.MergeRequestWebhook) (string, error)

// HandleMergeRequest implements the MergeRequestHandler by calling itself.
func (f MergeRequestHandlerFunc) HandleMergeRequest(ctx context.Context, webhook *gitlab.MergeRequestWebhook) (string, error) {
	return f(ctx, webhook)
}

// markdownQuote takes a text as input and adds `> ` in front of each line,
// making the text render as a quote in markdown. Returns an empty string
// if the provided text is empty or the provided text only contains whitespace.
func markdownQuote(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = "> " + line
	}
	return strings.Join(lines, "\n")
}

var gitLabReferenceReplaceTables = []struct {
	pattern *regexp.Regexp
	replace string
}{
	// @ user mention
	{regexp.MustCompile(`@(\w+)`), "`@`$1"},
	// # issue
	{regexp.MustCompile(`#(\d+)`), "`#`$1"},
	// ! merge request
	{regexp.MustCompile(`!(\d+)`), "`!`$1"},
	// $ snippet
	{regexp.MustCompile(`\$(\d+)`), "`$$`$1"},
	// ~ label
	{regexp.MustCompile(`~((\w|"|')+)`), "`~`$1"},
	// % milestone
	{regexp.MustCompile(`%((\w|"|')+)`), "`%`$1"},
}

// filterGitLabReferences takes a text and "escapes" any GitLab
// special references in it by putting them in an inline code
// block. E.g. "@user" -> "`@`user", "!123" -> "`!`123"
// See https://docs.gitlab.com/ee/user/markdown.html#special-gitlab-references
// Note: this should only be considered a best effort, bypassing this
// filter is certainly possible.
func filterGitLabReferences(text string) string {
	for _, rep := range gitLabReferenceReplaceTables {
		text = rep.pattern.ReplaceAllString(text, rep.replace)
	}
	return text
}
