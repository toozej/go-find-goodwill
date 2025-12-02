#!/bin/sh
set -e
rm -rf manpages
mkdir manpages
go run ./cmd/go-find-goodwill/ man | gzip -c -9 >manpages/go-find-goodwill.1.gz
