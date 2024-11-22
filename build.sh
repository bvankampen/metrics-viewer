#!/bin/bash

export GITHUB_TOKEN=$GORELEASE_GITHUB_TOKEN

goreleaser release clean
