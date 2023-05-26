#!/bin/bash
.gitlab/common/git.sh
go install gotest.tools/gotestsum@latest
gotestsum --junitfile report.xml --format testname
