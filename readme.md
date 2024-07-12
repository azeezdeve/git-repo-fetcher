# GitHub Repository Fetcher

This project provides a service to fetch and update GitHub repository information and commits using GitHub's public API. The service also supports continuous monitoring for changes and persists the data in a PostgreSQL database.

## Features

- Fetch repository information (name, description, URL, language, forks count, stars count, open issues count, watchers count, and created/updated dates).
- Fetch commit information (commit message, author, date, and URL).
- Continuously monitor the repository for changes.
- Handle GitHub API rate limits gracefully.
- Persist data into a PostgreSQL database.
- Expose APIs for fetching repository data.

## Requirements

- Go
- PostgreSQL
- GitHub API token (optional for higher rate limits)

## Setup

### Clone the Repository

```sh
git clone https://github.com/azeezdeve/git-repo-fetcher.git
cd git-repo-fetcher
go mod tidy
```

```bash
DATABASE_URL=postgres://user:password@localhost:5432/github_repos
GITHUB_BASE_URL=https://api.github.com
```

#### Run
```sh
cd server
go run main.go
```