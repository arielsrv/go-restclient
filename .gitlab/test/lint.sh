#!/bin/bash
.gitlab/common/git.sh
golangci-lint run --issues-exit-code 0 --print-issued-lines=false --out-format code-climate:gl-code-quality-report.json,line-number
