#!/bin/bash

export GITHUB_TOKEN=$GORELEASER_GITHUB_TOKEN

goreleaser release --clean
