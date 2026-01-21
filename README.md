# go-find-goodwill

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/toozej/go-find-goodwill)
[![Go Report Card](https://goreportcard.com/badge/github.com/toozej/go-find-goodwill)](https://goreportcard.com/report/github.com/toozej/go-find-goodwill)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/toozej/go-find-goodwill/cicd.yaml)
![Docker Pulls](https://img.shields.io/docker/pulls/toozej/go-find-goodwill)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/toozej/go-find-goodwill/total)


### Setup Instructions

- set up new repository in quay.io web console
  - (DockerHub and GitHub Container Registry do this automatically on first push/publish)
  - name must match Git repo name
  - grant robot user with username stored in QUAY_USERNAME "write" permissions (your quay.io account should already have admin permissions)
- set built packages visibility in GitHub packages to public
  - navigate to https://github.com/users/$USERNAME/packages/container/$REPO/settings
  - scroll down to "Danger Zone"
  - change visibility to public

## changes required to update golang version

- `make update-golang-version`
