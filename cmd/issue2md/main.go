package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/bigwhite/issue2md/internal/converter"
	"github.com/bigwhite/issue2md/internal/github"
)

var enableReactions = flag.Bool("enable-reactions", false, "Include reactions in the output.")
var enableUserLinks = flag.Bool("enable-user-links", false, "Enable user profile links in the output.")

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: issue2md [flags] url [markdown-file]\n")
	fmt.Fprintf(os.Stderr, "Arguments:\n")
	fmt.Fprintf(os.Stderr, "  url                 The URL of the GitHub issue, discussion, or pull request to convert.\n")
	fmt.Fprintf(os.Stderr, "  markdown-file       (optional) The output markdown file.\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Error: url is required.")
		usage()
		return
	}

	itemURL := args[0]
	var markdownFile string
	if len(args) >= 2 {
		markdownFile = args[1]
	}

	owner, repo, itemNumber, itemType, err := github.ParseURL(itemURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		return
	}

	token := os.Getenv("GITHUB_TOKEN")

	var markdown string
	switch itemType {
	case "issue":
		issue, err := github.FetchIssue(owner, repo, itemNumber, token)
		if err != nil {
			fmt.Printf("Error fetching issue: %v\n", err)
			return
		}

		comments, err := github.FetchComments(owner, repo, itemNumber, token, *enableReactions, *enableUserLinks)
		if err != nil {
			fmt.Printf("Error fetching comments: %v\n", err)
			return
		}
		markdown = converter.IssueToMarkdown(issue, comments, *enableUserLinks)

	case "pull":
		pullRequest, err := github.FetchPullRequest(owner, repo, itemNumber, token)
		if err != nil {
			fmt.Printf("Error fetching pull request: %v\n", err)
			return
		}

		// Pull Request comments are fetched via the issues API endpoint
		comments, err := github.FetchComments(owner, repo, itemNumber, token, *enableReactions, *enableUserLinks)
		if err != nil {
			fmt.Printf("Error fetching comments: %v\n", err)
			return
		}
		markdown = converter.PullRequestToMarkdown(pullRequest, comments, *enableUserLinks)

	case "discussion":
		discussion, err := github.FetchDiscussion(owner, repo, itemNumber, token)
		if err != nil {
			fmt.Printf("Error fetching discussion: %v\n", err)
			return
		}

		discussionComments, err := github.FetchDiscussionComments(owner, repo, itemNumber, token, *enableReactions)
		if err != nil {
			fmt.Printf("Error fetching discussion comments: %v\n", err)
			return
		}
		markdown = converter.DiscussionToMarkdown(discussion, discussionComments, *enableUserLinks)

	default:
		fmt.Printf("Unsupported URL type: %s\n", itemType)
		return
	}

	if markdownFile == "" {
		if _, err := io.WriteString(os.Stdout, markdown); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to stdout: %v\n", err)
			os.Exit(1)
		}
		return
	}

	file, err := os.Create(markdownFile)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	_, err = io.WriteString(file, markdown)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}

	fmt.Printf("Content saved as Markdown in file %s\n", markdownFile)
}
