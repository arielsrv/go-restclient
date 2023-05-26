#!/bin/bash
cat <<EOF >>"$HOME"/.netrc
machine gitlab.com
  login master_token
  password $CICD_TOKEN
EOF
