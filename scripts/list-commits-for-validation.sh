#!/bin/sh
# List commit SHAs, oldest first, for CI message and configuration validation.
# PRs compare against their target SHA; normal pushes compare against before.
# New or rewritten branches compare against origin's default branch so existing
# history is not checked again. Requires checkout with fetch-depth: 0.
# Inputs: HEAD, BASE, EVENT_NAME (push/pull_request), DEFAULT_BRANCH.
# Output: one SHA per line. Missing PR bases or fallback refs are errors.
set -eu

default_branch_commit() {
    git rev-parse --verify "refs/remotes/origin/${DEFAULT_BRANCH:?DEFAULT_BRANCH required}^{commit}"
}

head=$(git rev-parse --verify "${HEAD:?HEAD required}^{commit}")
case "${BASE:-}" in
    *[!0]*)
        if ! base=$(git rev-parse --verify "$BASE^{commit}" 2>/dev/null); then
            if [ "${EVENT_NAME:-}" != push ]; then
                echo 'Cannot resolve the commit validation base.' >&2
                exit 1
            fi
            base=$(default_branch_commit)
        fi
        # After a history rewrite, inspect the branch's changes relative to
        # the default branch rather than replaying newly inherited history.
        if [ "${EVENT_NAME:-}" = push ] && ! git merge-base --is-ancestor "$base" "$head"; then
            base=$(default_branch_commit)
        fi
        ;;
    *)
        # A branch's first push has an empty or all-zero before SHA. Exclude history
        # already on the default branch instead of checking the entire repo.
        base=$(default_branch_commit)
        ;;
esac

git rev-list --reverse "$base..$head"
