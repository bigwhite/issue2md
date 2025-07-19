package converter

import (
	"fmt"
	"strings"

	"github.com/bigwhite/issue2md/internal/github"
)

const githubBaseURL = "https://github.com"

// writeComment formats a comment into a markdown string with user profile link.
func writeComment(sb *strings.Builder, i int, user github.User, body string) {
	fmt.Fprintf(
		sb,
		"### Comment %d by [%s](%s/%s)\n\n%s\n\n",
		i+1, user.Login, githubBaseURL, user.Login, body,
	)
}

func IssueToMarkdown(issue *github.Issue, comments []github.Comment) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", issue.Title))
	sb.WriteString(fmt.Sprintf("**Issue Number**: #%d\n", issue.Number))
	sb.WriteString(fmt.Sprintf("**URL**: %s\n", issue.URL))
	sb.WriteString(fmt.Sprintf("**Created by**: [%s](%s/%s)\n\n", issue.User.Login, githubBaseURL, issue.User.Login))
	sb.WriteString(fmt.Sprintf("## Description\n\n%s\n\n", issue.Body))

	if len(comments) > 0 {
		sb.WriteString("## Comments\n\n")
		for i, comment := range comments {
			writeComment(&sb, i, comment.User, comment.Body)
		}
	}

	return sb.String()
}

func DiscussionToMarkdown(discussion *github.Discussion, discussionComments []github.DiscussionComment) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", discussion.Title))
	sb.WriteString(fmt.Sprintf("**Discussion Number**: #%d\n", discussion.Number))
	sb.WriteString(fmt.Sprintf("**URL**: %s\n", discussion.URL))
	sb.WriteString(fmt.Sprintf("**Created by**: [%s](%s/%s)\n\n", discussion.User.Login, githubBaseURL, discussion.User.Login))
	sb.WriteString(fmt.Sprintf("## Description\n%s\n\n", discussion.Body))

	if len(discussionComments) > 0 {
		sb.WriteString("## Comments\n\n")
		for i, comment := range discussionComments {
			writeComment(&sb, i, comment.User, comment.Body)
		}
	}

	return sb.String()
}
