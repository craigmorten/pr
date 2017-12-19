package main

import (
	"context"
	"log"
	"os"
	"sort"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type byRepoName []*github.Repository

func (r byRepoName) Len() int           { return len(r) }
func (r byRepoName) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r byRepoName) Less(i, j int) bool { return r[i].GetName() < r[j].GetName() }

func setupClient(apiKey string) (context.Context, *github.Client) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: apiKey},
	)
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return ctx, client
}

func getGithubRepositories(ctx context.Context, client *github.Client, org string) []*github.Repository {
	var repositories []*github.Repository

	repoListOpts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	// Loop over as per pagination
	for {
		out, resp, err := client.Repositories.ListByOrg(ctx, org, repoListOpts)
		if err != nil {
			log.Fatalf("\tError: %s\n", err.Error())
		}

		repositories = append(repositories, out...)

		if resp.NextPage == 0 {
			break
		}

		repoListOpts.Page = resp.NextPage
	}

	// Sort the repositories by repository name
	sort.Sort(byRepoName(repositories))
	return repositories
}

func printPullRequests(ctx context.Context, client *github.Client, org string, repositories []*github.Repository) {
	for i, repository := range repositories {
		// Get PRs
		pullRequests, _, err := client.PullRequests.List(ctx, org, repository.GetName(), &github.PullRequestListOptions{})
		if err != nil {
			log.Fatalf("\tError: %s\n", err.Error())
		}

		repoPipe := "├"
		basePipe := "│"
		if i == len(repositories)-1 {
			repoPipe = "└"
			basePipe = " "
		}

		if len(pullRequests) > 0 {
			log.Printf("%s %s:\n", repoPipe, repository.GetName())

			for j, pullRequest := range pullRequests {
				// Get reviewers
				reviewers, _, err := client.PullRequests.ListReviewers(ctx, org, repository.GetName(), pullRequest.GetNumber(), &github.ListOptions{})
				if err != nil {
					log.Fatalf("\tError: %s\n", err.Error())
				}

				prPipe := "├"
				userPipe := "│"
				if j == len(pullRequests)-1 {
					prPipe = "└"
					userPipe = " "
				}

				log.Printf("%s %s PR: %s\n", basePipe, prPipe, pullRequest.GetTitle())
				ownerPipe := "├"

				owner := pullRequest.GetUser().GetLogin()
				assignee := pullRequest.GetAssignee().GetLogin()

				if assignee == "" && len(reviewers.Users) == 0 {
					ownerPipe = "└"
				}

				if owner != "" {
					log.Printf("%s %s %s Owner: %s\n", basePipe, userPipe, ownerPipe, owner)
				}

				for k, reviewer := range reviewers.Users {
					if assignee == "" && k == len(reviewers.Users)-1 {
						ownerPipe = "└"
					}
					log.Printf("%s %s %s Reviewer %d: %s\n", basePipe, userPipe, ownerPipe, k+1, reviewer.GetLogin())
				}

				if assignee != "" {
					log.Printf("%s %s └ Assignee: %s\n", basePipe, userPipe, assignee)
				}
			}
		}
	}
}

func main() {
	// Get the org as the first argument passed on command line
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	org := os.Args[1]
	if len(org) == 0 {
		os.Exit(1)
	}

	// Get the api key as an environment variable
	apiKey := os.Getenv("GITHUB_API_KEY")
	if len(apiKey) == 0 {
		os.Exit(1)
	}

	// Setup the github context and client
	ctx, client := setupClient(apiKey)
	if ctx == nil {
		os.Exit(1)
	}
	if client == nil {
		os.Exit(1)
	}

	// Get all the repositories
	repositories := getGithubRepositories(ctx, client, org)

	// Output the results
	printPullRequests(ctx, client, org, repositories)
}
