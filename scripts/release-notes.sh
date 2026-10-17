#!/usr/bin/env bash
set -euo pipefail

# The workflow checks out the event tag, not the current release branch.
tag=${1:-}
if [[ ! "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
  echo 'expected a vX.Y.Z release tag' >&2
  exit 1
fi
if [[ $(git rev-parse HEAD) != "$(git rev-parse --verify "${tag}^{commit}")" ]]; then
  echo 'checkout does not match release tag' >&2
  exit 1
fi

# Follow fork release history, not tags on imported side branches. Use a Git
# range so old upstream commits merged since the last release are included.
if ! previous=$(git describe --tags --abbrev=0 --first-parent --match 'v[0-9]*' "${tag}^" 2>/dev/null); then
  echo 'cannot find previous fork release; fetch complete fork history and tags' >&2
  exit 1
fi
exec git-cliff --config cliff.toml --strip header "${previous}..${tag}"
