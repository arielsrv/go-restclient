#!/bin/bash
.gitlab/common/git.sh
curl -L  https://github.com/sourcegraph/lsif-go/releases/download/v1.2.0/src_linux_amd64 -o /usr/local/bin/lsif-go
chmod +x /usr/local/bin/lsif-go
lsif-go
