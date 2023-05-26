#!/bin/bash
cat <<EOF >>"$HOME"/.netrc
machine gitlab.com
  login master_token
  password $CICD_TOKEN
EOF

git config --global url."https://oauth2:${CICD_TOKEN}@gitlab.com".insteadOf "https://gitlab.com"
