# github-pr-prometheus-exporter
Prometheus Exporter for Github Pull Requests

## Preparation

Copy .envrc and load it.

```
$ cp .envrc.sample .envrc
$ # edit .envrc
$ # source .envrc
```

The target repositories are specified by GITHUB_REPOSITORIES environment varibales, that should be written in org/reponame, separated by commas.

>export GITHUB_REPOSITORIES="chaspy/datadog-github-pr,chaspy/favsearch"

## How to run

### Local

```
$ go run main.go
```

### Binary

Get the binary file from [Releases](https://github.com/chaspy/datadog-github-pr/releases) and run it.

### Docker

```
$ docker run -e GITHUB_TOKEN="${GITHUB_TOKEN}" -e GITHUB_REPOSITORIES="${GITHUB_REPOSITORIES}" chaspy/github-pr-ptometheus-exporter:v0.1.0
```

## Metrics

```
$ curl -s localhost:8080/metrics | grep github_pr_prometheus_exporter_pull_request_count
# HELP github_pr_prometheus_exporter_pull_request_count Number of Pull Requests
# TYPE github_pr_prometheus_exporter_pull_request_count gauge
github_pr_prometheus_exporter_pull_request_count{author="chaspy",label="",number="1470",repo="quipper/kubernetes-clusters",reviewer=""} 1
github_pr_prometheus_exporter_pull_request_count{author="dependabot-preview[bot]",label="dependencies,security",number="5563",repo="quipper/server-templates",reviewer=""} 1
github_pr_prometheus_exporter_pull_request_count{author="renovate[bot]",label="renovate:datadog,renovate:datadog/2.6.13",number="1798",repo="quipper/kubernetes-clusters",reviewer="chaspy"} 1
github_pr_prometheus_exporter_pull_request_count{author="renovate[bot]",label="renovate:ingress-nginx,renovate:ingress-nginx/3.20.1",number="1739",repo="quipper/kubernetes-clusters",reviewer="chaspy"} 1
```
