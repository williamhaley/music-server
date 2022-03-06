#!/usr/bin/env bash

set -e

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

export MUSIC_DIRECTORY=~/Audio/Music
export PERSISTENCE_DIRECTORY="${SCRIPT_DIR}/../db"
export MEILISEARCH_ADDRESS=http://127.0.0.1:7700
export ACCESS_TOKEN=abcd1234
export PORT=4000
export UI_SERVER_ADDRESS=http://localhost:3000

pushd "${SCRIPT_DIR}/.."

mkdir -p "${PERSISTENCE_DIRECTORY}"

# if [ ! -f ./meilisearch ];
# then
#   curl -L https://install.meilisearch.com | sh
# fi

# if ! ps ax | grep -i meilisearch | grep -v grep;
# then
#   ./meilisearch &
# fi

go run .
