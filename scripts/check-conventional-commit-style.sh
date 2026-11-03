#!/bin/sh
# Keep commit messages consistent in the commit-msg hook and CI.
# Usage: check-conventional-commit-style.sh MESSAGE_FILE | --commit SHA
# Enforce the repository's allowed Conventional Commit types and require a
# descriptive footer for breaking changes. CI skips actual merge commits because
# Git/GitHub generate their messages. Exit 0 on success, nonzero on invalid input.
set -eu

if [ "${1:-}" = "--commit" ]; then
    commit=$(git rev-parse --verify "${2:?commit required}^{commit}")
    # Git/GitHub generate merge subjects. Identify merges by their parents so
    # ordinary commits cannot bypass validation with a merge-like subject.
    parents=$(git show -s --format=%P "$commit")
    case "$parents" in
        *' '*) exit 0 ;;
    esac
    message=$(git show -s --format=%B "$commit")
else
    # Ignore Git's editor comments, just as the default commit cleanup does.
    message=$(git stripspace --strip-comments < "${1:?commit message file required}")
fi

subject=$(printf '%s\n' "$message" | sed -n '1p')
if ! printf '%s\n' "$subject" | grep -Eq '^(feat|chore|fix|test|doc|refactor|perf)(\([^()[:space:]]+\))?!?: [^[:space:]].*$'; then
    printf '%s\n' \
        'Invalid Conventional Commit style. Use: type(scope): description' \
        'Allowed types: feat, chore, fix, test, doc, refactor, perf. Scope is optional.' >&2
    exit 1
fi

if [ -n "$(printf '%s\n' "$message" | sed -n '2p')" ]; then
    printf '%s\n' 'Separate the subject from the body or footers with a blank line.' >&2
    exit 1
fi

# Git trailer tokens cannot contain spaces; normalize only the conventional
# BREAKING CHANGE token to its equivalent BREAKING-CHANGE spelling for parsing.
trailers=$(printf '%s\n' "$message" | sed 's/^BREAKING CHANGE:/BREAKING-CHANGE:/' |
    git -c trailer.separators=: interpret-trailers --parse)
if printf '%s\n' "$subject" | grep -Eq '^(feat|chore|fix|test|doc|refactor|perf)(\([^()[:space:]]+\))?!:' ||
    printf '%s\n' "$message" | grep -Eq '^BREAKING[ -]CHANGE:'; then
    if ! printf '%s\n' "$trailers" | grep -Eq '^BREAKING-CHANGE: [^[:space:]]'; then
        printf '%s\n' \
            'Breaking changes require a footer describing what breaks:' \
            '' 'BREAKING CHANGE: description of the incompatible change' >&2
        exit 1
    fi
fi
