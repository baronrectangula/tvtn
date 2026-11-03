#!/bin/sh
set -eu

head=$(git rev-parse --verify "${HEAD:?HEAD required}^{commit}")
case "${BASE:-}" in
    *[!0]*)
        base=$(git rev-parse --verify "$BASE^{commit}")
        ;;
    *)
        # A branch's first push has an empty or all-zero before SHA. Exclude history
        # already on the default branch instead of checking the entire repo.
        base=$(git rev-parse --verify "refs/remotes/origin/${DEFAULT_BRANCH:?DEFAULT_BRANCH required}^{commit}")
        ;;
esac

git rev-list --reverse "$base..$head"
