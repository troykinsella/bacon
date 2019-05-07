#!/usr/bin/env bash

set -e

PREFIX=${PREFIX:-${1:-/usr/local}}

export GH_USER=troykinsella
export GH_REPO=bacon
export ASSET=bacon_$(uname | tr '[:upper:]' '[:lower:]')_amd64
bash <(wget -q -O - https://gist.githubusercontent.com/troykinsella/b0cf38006d6ff4afb5056129d14f0738/raw/fetch_gh_release.sh)

DEST=$PREFIX/bin/bacon
mkdir -p $(dirname $DEST)
mv $ASSET $DEST
chmod +x $DEST

echo "Installed: $DEST"
