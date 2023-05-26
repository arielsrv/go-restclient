#!/bin/bash
.gitlab/common/git.sh
go mod tidy
go build -v ./...
