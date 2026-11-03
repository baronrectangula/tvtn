#!/bin/sh
# Require an explicit local-development confirmation when config.toml or
# .env.sample changes, to prompt review before committing deployment credentials.
# This checks the author's acknowledgement; it does not detect secrets.
# Usage: validate-config-change-confirmation.sh MESSAGE_FILE | --commit SHA
# The hook checks staged changes; CI checks the given commit, skipping merges.
# Exit 0 if confirmed/not applicable, 3 for a missing footer, otherwise an error.
set -eu

# Git passes a message file; CI passes a commit to inspect.
if [ "${1:-}" = "--commit" ]; then
    commit=$(git rev-parse --verify "${2:?commit required}^{commit}")
    # Generated merge messages do not carry the original commits' footers.
    # Those commits remain in the CI range and are validated individually.
    parents=$(git show -s --format=%P "$commit")
    case "$parents" in
        *' '*) exit 0 ;;
    esac
    changes=$(git diff-tree --root --no-commit-id --name-only -r --no-renames \
        "$commit" -- \
        ':(top,literal)config.toml' ':(top,literal).env.sample')
    [ -n "$changes" ] || exit 0
    message=$(git show -s --format=%B "$commit")
else
    message=$(cat "${1:?commit message file required}")
    # Inspect Git's active index, including partial commits and deletions.
    if git diff --cached --quiet --no-ext-diff --no-renames -- \
        ':(top,literal)config.toml' ':(top,literal).env.sample'; then
        exit 0
    else
        status=$?
        [ "$status" -eq 1 ] || exit "$status"
    fi
fi

required_en='Config-Change: I confirm these settings and credentials are for local development only.'
required_vi='Config-Change: Tôi xác nhận cấu hình và thông tin đăng nhập này chỉ dùng để phát triển trên máy local.'
# Accept a Conventional Commits breaking note alongside the confirmation.
trailers=$(printf '%s\n' "$message" | sed 's/^BREAKING CHANGE:/BREAKING-CHANGE:/' |
    git -c trailer.separators=: interpret-trailers --parse)

# Require both a recognized footer and the exact line, without normalization.
for required in "$required_en" "$required_vi"; do
    if printf '%s\n' "$trailers" | grep -Fqx -- "$required" &&
        printf '%s\n' "$message" | grep -Fqx -- "$required"; then
        exit 0
    fi
done

printf '%s\n' \
    'Configuration change confirmation required: config.toml or .env.sample changed.' \
    'Review them for deployment credentials, then add either exact footer after a blank line:' \
    '' "$required_en" "or" "$required_vi" >&2
# Distinguish a missing confirmation from Git or execution errors in CI.
exit 3
