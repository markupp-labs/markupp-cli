#!/usr/bin/env bash
# Validates a commit message against the conventions in AGENTS.md.
#
# Usage: valida-mensagem-de-commit.sh FILE   (FILE "-" reads stdin)
#
# Exit status: 0 accepted, 1 rejected, 2 usage error. It runs as the
# commit-msg hook (pre-commit passes the message file) and in CI, once per
# commit of a pull request. Byte level patterns under LC_ALL=C keep it
# portable to BSD grep, which has no -P.
set -uo pipefail
export LC_ALL=C

readonly ALLOWED_TYPES='feat|fix|docs|test|chore|refact|release|merge'
readonly SCISSORS='^# -+ >8 -+$'
# UTF-8 lead bytes of U+1F000-1FAFF, U+2600-27BF, U+2B00-2BFF and U+FE0F.
readonly EMOJI_BYTES=$'\xf0\x9f[\x80-\xab]|\xe2[\x98-\x9e]|\xe2[\xac-\xaf]|\xef\xb8\x8f'
readonly EM_DASH=$'\xe2\x80\x94'

reject() {
  printf 'commit message rejected: %s\n' "$1" >&2
  exit 1
}

usage_error() {
  printf 'usage: %s FILE (or - for stdin): %s\n' "${0##*/}" "$1" >&2
  exit 2
}

# read_message prints the message without the verbose diff after the
# scissors line and without git comment lines.
read_message() {
  local source="$1"
  [ "$source" = "-" ] && source=/dev/stdin
  sed -E "/$SCISSORS/,\$d" "$source" | grep -v '^#'
}

first_line() {
  printf '%s\n' "$1" | grep -v '^[[:space:]]*$' | head -n 1
}

is_generated_by_git() {
  printf '%s\n' "$1" | grep -qE '^(Merge|Revert) '
}

check_subject() {
  local subject="$1"
  [ -n "$subject" ] || reject "the message is empty, expected '<type>: <description>'"
  if printf '%s\n' "$subject" | grep -qE '^refactor[(:]'; then
    reject "use 'refact:', never 'refactor:'. got subject: $subject"
  fi
  if printf '%s\n' "$subject" | grep -qE '^[A-Za-z]+\([^)]*\)!?:'; then
    reject "no scope in parentheses, expected '<type>: <description>'. got subject: $subject"
  fi
  printf '%s\n' "$subject" | grep -qE "^($ALLOWED_TYPES): [^[:space:]]" && return 0
  reject "expected '<type>: <description>' with type in ($ALLOWED_TYPES). got subject: $subject"
}

check_body() {
  local message="$1"
  if printf '%s\n' "$message" | grep -qE "$EMOJI_BYTES"; then
    reject "the message contains an emoji, which no text in this project uses"
  fi
  if printf '%s\n' "$message" | grep -qF "$EM_DASH"; then
    reject "the message contains an em dash, use a comma, a colon or parentheses"
  fi
  if printf '%s\n' "$message" | grep -qiE 'co-authored-by|generated with'; then
    reject "the message credits a tool as author, the person is the author of the work"
  fi
}

main() {
  [ "$#" -eq 1 ] || usage_error "expected one argument, got $#"
  [ "$1" = "-" ] || [ -r "$1" ] || usage_error "cannot read message file '$1'"
  local message subject
  message="$(read_message "$1")"
  subject="$(first_line "$message")"
  if is_generated_by_git "$subject"; then
    exit 0
  fi
  check_subject "$subject"
  check_body "$message"
}

main "$@"
