package converter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bigwhite/issue2md/internal/github"
)

const githubBaseURL = "https://github.com"

func formatUser(login string, enableUserLinks bool) string {
	if enableUserLinks {
		return fmt.Sprintf("[%s](%s/%s)", login, githubBaseURL, login)
	}
	return login
}

// writeComment formats a comment into a markdown string with user profile link.
func writeComment(sb *strings.Builder, i int, user github.User, body string, enableUserLinks bool) {
	fmt.Fprintf(
		sb,
		"### Comment %d by %s\n\n%s\n\n",
		i+1, formatUser(user.Login, enableUserLinks), body,
	)
}

// Helper to write reactions for a comment or discussion comment.
func writeReactions(sb *strings.Builder, reactions []github.Reaction, enableUserLinks bool) {
	if len(reactions) == 0 {
		return
	}
	sb.WriteString("**Reactions:**\n")
	reactionMap := make(map[string][]github.User)
	for _, reaction := range reactions {
		reactionMap[reaction.Content] = append(reactionMap[reaction.Content], reaction.User)
	}
	// Sort reaction types for deterministic output
	reactionTypes := make([]string, 0, len(reactionMap))
	for reaction := range reactionMap {
		reactionTypes = append(reactionTypes, reaction)
	}
	sort.Strings(reactionTypes)
	for _, reaction := range reactionTypes {
		users := reactionMap[reaction]
		sb.WriteString(fmt.Sprintf("- :%s: by %d user(s):\n", reaction, len(users)))
		for _, user := range users {
			sb.WriteString(fmt.Sprintf("  - %s\n", formatUser(user.Login, enableUserLinks)))
		}
	}
	sb.WriteString("\n")
}

func IssueToMarkdown(issue *github.Issue, comments []github.Comment, enableUserLinks bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", issue.Title))
	sb.WriteString(fmt.Sprintf("- **Issue Number**: #%d\n", issue.Number))
	sb.WriteString(fmt.Sprintf("- **URL**: %s\n", issue.URL))
	sb.WriteString(fmt.Sprintf("- **Created by**: %s\n\n", formatUser(issue.User.Login, enableUserLinks)))
	sb.WriteString(fmt.Sprintf("## Description\n\n%s\n\n", issue.Body))
	writeReactions(&sb, issue.Reactions, enableUserLinks)

	if len(comments) > 0 {
		sb.WriteString("## Comments\n\n")
		for i, comment := range comments {
			writeComment(&sb, i, comment.User, comment.Body, enableUserLinks)
			writeReactions(&sb, comment.Reactions, enableUserLinks)
		}
	}

	return sb.String()
}

func PullRequestToMarkdown(pr *github.PullRequest, comments []github.Comment, enableUserLinks bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", pr.Title))
	sb.WriteString(fmt.Sprintf("- **Pull Request Number**: #%d\n", pr.Number))
	sb.WriteString(fmt.Sprintf("- **URL**: %s\n", pr.URL))
	sb.WriteString(fmt.Sprintf("- **Created by**: %s\n\n", formatUser(pr.User.Login, enableUserLinks)))
	sb.WriteString(fmt.Sprintf("## Description\n\n%s\n\n", pr.Body))
	writeReactions(&sb, pr.Reactions, enableUserLinks)

	if len(comments) > 0 {
		sb.WriteString("## Comments\n\n")
		for i, comment := range comments {
			writeComment(&sb, i, comment.User, comment.Body, enableUserLinks)
			writeReactions(&sb, comment.Reactions, enableUserLinks)
		}
	}

	return sb.String()
}

func DiscussionToMarkdown(discussion *github.Discussion, discussionComments []github.DiscussionComment, enableUserLinks bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", discussion.Title))
	sb.WriteString(fmt.Sprintf("**Discussion Number**: #%d\n", discussion.Number))
	sb.WriteString(fmt.Sprintf("**URL**: %s\n", discussion.URL))
	sb.WriteString(fmt.Sprintf("**Created by**: %s\n\n", formatUser(discussion.User.Login, enableUserLinks)))
	sb.WriteString(fmt.Sprintf("## Description\n%s\n\n", discussion.Body))

	if len(discussionComments) > 0 {
		sb.WriteString("## Comments\n\n")
		for i, comment := range discussionComments {
			writeComment(&sb, i, comment.User, comment.Body, enableUserLinks)
			writeReactions(&sb, comment.Reactions, enableUserLinks)
		}
	}

	return sb.String()
}
