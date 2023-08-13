#!/bin/bash
cat <<EOF >>"$HOME"/.netrc
machine gitlab.com
  login master_token
  password $GITLAB_TOKEN
EOF

git config --global url."https://oauth2:${GITLAB_TOKEN}@gitlab.com".insteadOf "https://gitlab.com"
