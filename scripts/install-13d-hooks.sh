#!/bin/sh
set -eu

repository_root=$(git rev-parse --show-toplevel)
git -C "$repository_root" config core.hooksPath .githooks
printf '%s\n' "Installed repository-local 13D hooks via core.hooksPath=.githooks"
printf '%s\n' "authorityEffect: NONE"
