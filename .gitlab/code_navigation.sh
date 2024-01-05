#!/bin/bash
.gitlab/common/git.sh

TAG=v1.9.3
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed -e 's/x86_64/amd64/')
echo "${TAG}" "${OS}" "${ARCH}"
curl https://github.com/sourcegraph/lsif-go/releases/download/"${TAG}"/src_"${OS}"_"${ARCH}" -o /usr/local/bin/lsif-go
chmod +x /usr/local/bin/lsif-go
lsif-go
