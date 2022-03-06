#!/usr/bin/env bash

docker buildx build --push --platform linux/arm64 -t williamhaley/meilisearch:armv7 -f Dockerfile.meilisearch .

