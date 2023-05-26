#!/bin/bash
.gitlab/common/git.sh
CGO_ENABLED=0 go test ./... -coverprofile=coverage-report.out
go tool cover -html=coverage-report.out -o coverage-report.html
go tool cover -func=coverage-report.out
