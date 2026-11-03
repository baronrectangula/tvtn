#!/bin/sh
set -eu

head=$(git rev-parse --verify "${HEAD:?HEAD required}^{commit}")
case "${BASE:-}" in
    *[!0]*)
        if base=$(git rev-parse --verify "$BASE^{commit}" 2>/dev/null); then
            :
        elif [ "${EVENT_NAME:-}" = push ]; then
            # Rebased/force-pushed commits may no longer be fetched by checkout.
            base=$(git rev-parse --verify "refs/remotes/origin/${DEFAULT_BRANCH:?DEFAULT_BRANCH required}^{commit}")
        else
            echo 'Cannot resolve the commit validation base.' >&2
            exit 1
        fi
        # After a history rewrite, inspect the branch's changes relative to
        # the default branch rather than replaying newly inherited history.
        if [ "${EVENT_NAME:-}" = push ] && ! git merge-base --is-ancestor "$base" "$head"; then
            base=$(git rev-parse --verify "refs/remotes/origin/${DEFAULT_BRANCH:?DEFAULT_BRANCH required}^{commit}")
        fi
        ;;
    *)
        # A branch's first push has an empty or all-zero before SHA. Exclude history
        # already on the default branch instead of checking the entire repo.
        base=$(git rev-parse --verify "refs/remotes/origin/${DEFAULT_BRANCH:?DEFAULT_BRANCH required}^{commit}")
        ;;
esac

git rev-list --reverse "$base..$head"
