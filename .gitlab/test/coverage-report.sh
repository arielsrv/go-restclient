#!/bin/bash
.gitlab/common/git.sh
go install
go test ./... -coverprofile=coverage.txt -covermode count
go get github.com/boumenot/gocover-cobertura
go run github.com/boumenot/gocover-cobertura <coverage.txt >coverage.xml
