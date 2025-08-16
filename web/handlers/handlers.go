package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/bigwhite/issue2md/internal/converter"
	"github.com/bigwhite/issue2md/internal/github"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ConvertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	itemURL := r.FormValue("issue_url")
	if itemURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	enableReactions := r.FormValue("enable_reactions") == "true"
	enableUserLinks := r.FormValue("enable_user_links") == "true"

	owner, repo, itemNumber, itemType, err := github.ParseURL(itemURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing URL: %v", err), http.StatusBadRequest)
		return
	}

	token := os.Getenv("GITHUB_TOKEN")

	var markdown string
	var filename string

	switch itemType {
	case "issue":
		issue, err := github.FetchIssue(owner, repo, itemNumber, token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching issue: %v", err), http.StatusInternalServerError)
			return
		}

		comments, err := github.FetchComments(owner, repo, itemNumber, token, enableReactions, enableUserLinks)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching comments: %v", err), http.StatusInternalServerError)
			return
		}
		markdown = converter.IssueToMarkdown(issue, comments, enableUserLinks)
		filename = fmt.Sprintf("%s_%s_issue_%d.md", owner, repo, issue.Number)
	case "pull":
		pullRequest, err := github.FetchPullRequest(owner, repo, itemNumber, token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching pull request: %v", err), http.StatusInternalServerError)
			return
		}
		// Pull Request comments are fetched via the issues API endpoint
		comments, err := github.FetchComments(owner, repo, itemNumber, token, enableReactions, enableUserLinks)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching comments: %v", err), http.StatusInternalServerError)
			return
		}
		markdown = converter.PullRequestToMarkdown(pullRequest, comments, enableUserLinks)
		filename = fmt.Sprintf("%s_%s_pull_%d.md", owner, repo, pullRequest.Number)
	case "discussion":
		discussion, err := github.FetchDiscussion(owner, repo, itemNumber, token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching discussion: %v", err), http.StatusInternalServerError)
			return
		}
		discussionComments, err := github.FetchDiscussionComments(owner, repo, itemNumber, token, enableReactions)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching discussion comments: %v", err), http.StatusInternalServerError)
			return
		}
		markdown = converter.DiscussionToMarkdown(discussion, discussionComments, enableUserLinks)
		filename = fmt.Sprintf("%s_%s_discussion_%d.md", owner, repo, discussion.Number)
	default:
		http.Error(w, fmt.Sprintf("Unsupported URL type: %s", itemType), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/markdown")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	_, err = w.Write([]byte(markdown))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error writing response: %v", err), http.StatusInternalServerError)
	}
}
