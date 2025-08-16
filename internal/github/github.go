package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Issue struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Number   int    `json:"number"`
	URL      string `json:"html_url"` // Use html_url for consistency
	Comments int    `json:"comments"`
	User     User   `json:"user"`
}

type PullRequest struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Number   int    `json:"number"`
	URL      string `json:"html_url"` // Use html_url for the web link
	Comments int    `json:"comments"`
	User     User   `json:"user"`
}

type Discussion struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Number   int    `json:"number"`
	URL      string `json:"html_url"` // Use html_url for discussions
	Comments int    `json:"comments_count"`
	User     User   `json:"user"`
}

type DiscussionComment struct {
	Body      string     `json:"body"`
	User      User       `json:"user"`
	ID        int        `json:"id"`
	Reactions []Reaction `json:"-"`
}

type Comment struct {
	Body      string     `json:"body"`
	User      User       `json:"user"`
	ID        int        `json:"id"`
	Reactions []Reaction `json:"-"`
}

type Reaction struct {
	Content string `json:"content"`
	User    User   `json:"user"`
}

type User struct {
	Login string `json:"login"`
}

var sharedHTTPClient = &http.Client{}

func ParseURL(issueURL string) (owner, repo string, number int, issueType string, err error) {
	parsedURL, err := url.Parse(issueURL)
	if err != nil {
		return
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) < 4 {
		err = fmt.Errorf("invalid GitHub URL format")
		return
	}

	if parts[2] == "issues" && len(parts) == 4 {
		issueType = "issue"
	} else if parts[2] == "discussions" && len(parts) == 4 {
		issueType = "discussion"
	} else if parts[2] == "pull" && len(parts) == 4 {
		issueType = "pull"
	} else {
		err = fmt.Errorf("invalid GitHub issue, discussion, or pull request URL format")
		return
	}
	owner = parts[0]
	repo = parts[1]
	issueNumber, err := strconv.Atoi(parts[3])
	if err != nil {
		return
	}

	return owner, repo, issueNumber, issueType, nil
}

func FetchIssue(owner, repo string, issueNumber int, token string) (*Issue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, issueNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

func FetchPullRequest(owner, repo string, pullNumber int, token string) (*PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, pullNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var pullRequest PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pullRequest); err != nil {
		return nil, err
	}

	return &pullRequest, nil
}

func FetchComments(owner, repo string, issueNumber int, token string, enableReactions bool, enableUserLinks bool) ([]Comment, error) {
	baseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, issueNumber)
	var allComments []Comment

	nextURL := baseURL // Initial URL

	for nextURL != "" {
		req, err := http.NewRequest("GET", nextURL, nil)
		if err != nil {
			return nil, err
		}

		if token != "" {
			req.Header.Set("Authorization", "token "+token)
		}

		resp, err := sharedHTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
		}

		var currentComments []Comment
		if err := json.NewDecoder(resp.Body).Decode(&currentComments); err != nil {
			return nil, err
		}

		// Fetch reactions for each comment
		if enableReactions {
			for i := range currentComments {
				reactions, err := FetchReactionsForComment(owner, repo, currentComments[i].ID, token)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch reactions for comment %d in %s/%s issue %d: %v. Ensure you have set a valid GITHUB_TOKEN", currentComments[i].ID, owner, repo, issueNumber, err)
				}
				currentComments[i].Reactions = reactions
			}
		}

		allComments = append(allComments, currentComments...)

		nextURL = "" // Reset nextURL, will be updated from Link header if exists
		linkHeader := resp.Header.Get("Link")
		if linkHeader != "" {
			// Parse Link header to find "next" page URL
			for _, link := range strings.Split(linkHeader, ",") {
				link = strings.TrimSpace(link)
				parts := strings.Split(link, ";")
				if len(parts) != 2 {
					continue
				}
				urlPart := strings.Trim(parts[0], "<>")
				relPart := strings.TrimSpace(parts[1])
				if relPart == `rel="next"` {
					nextURL = urlPart
					break // Found "next", no need to check other links
				}
			}
		}
		// If nextURL is still "", it means no "next" page, so exit loop
	}

	return allComments, nil
}

func FetchReactionsForComment(owner, repo string, commentID int, token string) ([]Reaction, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/comments/%d/reactions", owner, repo, commentID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.squirrel-girl-preview+json")

	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var reactions []Reaction
	if err := json.NewDecoder(resp.Body).Decode(&reactions); err != nil {
		return nil, err
	}
	return reactions, nil
}

func FetchDiscussion(owner, repo string, discussionNumber int, token string) (*Discussion, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/discussions/%d", owner, repo, discussionNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}
	req.Header.Set("Accept", "application/vnd.github+json") // Important for Discussions API
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")    // Specify API version

	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var discussion Discussion
	if err := json.NewDecoder(resp.Body).Decode(&discussion); err != nil {
		return nil, err
	}

	return &discussion, nil
}

func FetchDiscussionComments(owner, repo string, discussionNumber int, token string, enableReactions bool) ([]DiscussionComment, error) {
	baseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/discussions/%d/comments?state=all", owner, repo, discussionNumber)
	var allComments []DiscussionComment

	nextURL := baseURL // Initial URL

	for nextURL != "" {
		req, err := http.NewRequest("GET", nextURL, nil)
		if err != nil {
			return nil, err
		}

		if token != "" {
			req.Header.Set("Authorization", "token "+token)
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		resp, err := sharedHTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
		}

		var currentComments []DiscussionComment
		if err := json.NewDecoder(resp.Body).Decode(&currentComments); err != nil {
			return nil, err
		}

		// Fetch reactions for each discussion comment
		if enableReactions {
			for i := range currentComments {
				reactions, err := FetchReactionsForComment(owner, repo, currentComments[i].ID, token)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch reactions for comment %d in %s/%s discussion %d: %v. Ensure you have set a valid GITHUB_TOKEN", currentComments[i].ID, owner, repo, discussionNumber, err)
				}
				currentComments[i].Reactions = reactions
			}
		}

		allComments = append(allComments, currentComments...)

		nextURL = "" // Reset nextURL, will be updated from Link header if exists
		linkHeader := resp.Header.Get("Link")
		if linkHeader != "" {
			// Parse Link header to find "next" page URL
			for _, link := range strings.Split(linkHeader, ",") {
				link = strings.TrimSpace(link)
				parts := strings.Split(link, ";")
				if len(parts) != 2 {
					continue
				}
				urlPart := strings.Trim(parts[0], "<>")
				relPart := strings.TrimSpace(parts[1])
				if relPart == `rel="next"` {
					nextURL = urlPart
					break // Found "next", no need to check other links
				}
			}
		}
		// If nextURL is still "", it means no "next" page, so exit loop
	}

	return allComments, nil
}
