#!/bin/sh
set -e
rm -rf completions
mkdir completions
task build
for sh in bash zsh fish; do
	./cloudflare-utils completion "$sh" >"completions/cloudflare-utils.$sh"
done