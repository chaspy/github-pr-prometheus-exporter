package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/oauth2"
)

type PR struct {
	Number             int
	Labels             []*github.Label
	User               string
	RequestedReviewers []*github.User
	Repo               string
}

var (
	//nolint:gochecknoglobals
	PullRequestCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "github_pr",
		Subsystem: "prometheus_exporter",
		Name:      "pull_request_count",
		Help:      "Number of Pull Requests",
	},
		[]string{"number", "label", "author", "reviewer", "repo"},
	)
)

func main() {
	interval, err := getInterval()
	if err != nil {
		log.Fatal(err)
	}

	prometheus.MustRegister(PullRequestCount)

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)

		// register at first time
		err := snapshot()
		if err != nil {
			log.Fatal(err)
		}

		// register metrics as background
		for range ticker.C {
			err := snapshot()
			if err != nil {
				log.Fatal(err)
			}
		}
	}()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func snapshot() error {
	PullRequestCount.Reset()

	githubToken, err := readGithubConfig()
	if err != nil {
		return fmt.Errorf("failed to read Datadog Config: %w", err)
	}

	repositories, err := getRepositories()
	if err != nil {
		return fmt.Errorf("failed to get GitHub repository name: %w", err)
	}

	repositoryList := parseRepositories(repositories)

	prs, err := getPullRequests(githubToken, repositoryList)
	if err != nil {
		return fmt.Errorf("failed to get PullRequests: %w", err)
	}

	prInfos := getPRInfos(prs)

	for _, prInfo := range prInfos {
		labelsTag := make([]string, len(prInfo.Labels))
		for i, label := range prInfo.Labels {
			labelsTag[i] = *label.Name
		}

		reviewersTag := make([]string, len(prInfo.RequestedReviewers))
		for i, reviewer := range prInfo.RequestedReviewers {
			reviewersTag[i] = *reviewer.Login
		}

		labels := prometheus.Labels{
			"number":   strconv.Itoa(prInfo.Number),
			"label":    strings.Join(labelsTag, ","),
			"author":   prInfo.User,
			"reviewer": strings.Join(reviewersTag, ","),
			"repo":     prInfo.Repo,
		}

		PullRequestCount.With(labels).Set(1)
	}

	return nil
}

func getInterval() (int, error) {
	const defaultGithubAPIIntervalSecond = 300
	githubAPIInterval := os.Getenv("GITHUB_API_INTERVAL")
	if len(githubAPIInterval) == 0 {
		return defaultGithubAPIIntervalSecond, nil
	}

	integerGithubAPIInterval, err := strconv.Atoi(githubAPIInterval)
	if err != nil {
		return 0, fmt.Errorf("failed to read Datadog Config: %w", err)
	}

	return integerGithubAPIInterval, nil
}

func readGithubConfig() (string, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if len(githubToken) == 0 {
		return "", fmt.Errorf("missing environment variable: GITHUB_TOKEN")
	}

	return githubToken, nil
}

func getPullRequests(githubToken string, githubRepositories []string) ([]*github.PullRequest, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	const perPage = 100
	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: perPage},
	}

	prs := []*github.PullRequest{}
	allPrsInRepo := []*github.PullRequest{}

	for _, githubRepository := range githubRepositories {
		repo := strings.Split(githubRepository, "/")
		org := repo[0]
		name := repo[1]

		for {
			prsInRepo, resp, err := client.PullRequests.List(ctx, org, name, opt)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub Pull Requests: %w", err)
			}
			allPrsInRepo = append(allPrsInRepo, prsInRepo...)
			if resp.NextPage == 0 {
				opt.Page = 0
				break
			}
			opt.Page = resp.NextPage
		}

		prs = append(prs, allPrsInRepo...)
	}

	return prs, nil
}

func getRepositories() (string, error) {
	githubRepositories := os.Getenv("GITHUB_REPOSITORIES")
	if len(githubRepositories) == 0 {
		return "", fmt.Errorf("missing environment variable: GITHUB_REPOSITORIES")
	}

	return githubRepositories, nil
}

func parseRepositories(repositories string) []string {
	return strings.Split(repositories, ",")
}

func getPRInfos(prs []*github.PullRequest) []PR {
	prInfos := make([]PR, len(prs))

	for i, pr := range prs {
		repos := strings.Split(pr.GetURL(), "/")

		prInfos[i] = PR{
			Number:             pr.GetNumber(),
			Labels:             pr.Labels,
			User:               pr.User.GetLogin(),
			RequestedReviewers: pr.RequestedReviewers,
			Repo:               repos[4] + "/" + repos[5],
		}
	}

	return prInfos
}
