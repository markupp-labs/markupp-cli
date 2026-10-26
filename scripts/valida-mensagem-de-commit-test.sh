#!/usr/bin/env bash
# Tests for valida-mensagem-de-commit.sh. Each case feeds a commit message to
# the validator, once through a file and once through stdin, and asserts the
# exit status. Run it with `make test-scripts`.
set -uo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
validator="$here/valida-mensagem-de-commit.sh"
workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT
failures=0
cases=0

# check_case NAME WANT MESSAGE runs the validator on MESSAGE and compares the
# exit status with WANT (0 accepted, 1 rejected).
check_case() {
  local name="$1" want="$2" message="$3" got
  cases=$((cases + 1))
  printf '%s' "$message" > "$workdir/msg"
  bash "$validator" "$workdir/msg" > /dev/null 2> "$workdir/err"
  got=$?
  report "$name (file)" "$want" "$got"
  printf '%s' "$message" | bash "$validator" - > /dev/null 2> "$workdir/err"
  got=$?
  report "$name (stdin)" "$want" "$got"
}

report() {
  local name="$1" want="$2" got="$3"
  if [ "$want" = "$got" ]; then
    return 0
  fi
  failures=$((failures + 1))
  printf 'FAIL %s: want exit %s, got %s\n' "$name" "$want" "$got" >&2
  sed 's/^/  stderr: /' "$workdir/err" >&2
}

accepted_cases() {
  local t
  for t in feat fix docs test chore refact release merge; do
    check_case "type $t" 0 "$t: adiciona algo"
  done
  check_case "body after a blank line" 0 $'feat: adiciona busca\n\nExplica o motivo, com acentuação.\n'
  check_case "git comments are ignored" 0 $'fix: corrige saida\n# Please enter the commit message\n# Co-Authored-By: someone\n'
  check_case "scissors cut the verbose diff" 0 $'docs: atualiza leia-me\n# ------------------------ >8 ------------------------\n+ \xe2\x80\x94 Co-Authored-By: x\n'
  check_case "git merge message" 0 "Merge branch 'dev' into feat/busca"
  check_case "git revert message" 0 'Revert "feat: adiciona busca"'
  check_case "leading blank lines" 0 $'\n\nchore: atualiza dependencias\n'
}

rejected_cases() {
  check_case "refactor instead of refact" 1 "refactor: extrai funcao"
  check_case "unknown type" 1 "feature: adiciona busca"
  check_case "uppercase type" 1 "Feat: adiciona busca"
  check_case "scope in parentheses" 1 "feat(cli): adiciona busca"
  check_case "refactor with scope" 1 "refactor(cli): extrai funcao"
  check_case "missing space after colon" 1 "feat:adiciona busca"
  check_case "empty description" 1 "feat: "
  check_case "no type" 1 "adiciona busca"
  check_case "empty message" 1 ""
  check_case "only comments" 1 $'# nothing here\n'
  check_case "emoji in subject" 1 $'feat: adiciona busca \xf0\x9f\x9a\x80'
  check_case "emoji in body" 1 $'feat: adiciona busca\n\nPronto \xe2\x9c\xa8\n'
  check_case "em dash in subject" 1 $'feat: adiciona busca \xe2\x80\x94 rapida'
  check_case "em dash in body" 1 $'fix: corrige\n\nmotivo \xe2\x80\x94 detalhe\n'
  check_case "co-authored-by trailer" 1 $'feat: adiciona busca\n\nCo-Authored-By: Someone <a@b.c>\n'
  check_case "co-authored-by lowercase" 1 $'feat: adiciona busca\n\nco-authored-by: someone\n'
  check_case "generated with footer" 1 $'feat: adiciona busca\n\nGenerated with some tool\n'
  check_case "generated with lowercase" 1 $'feat: adiciona busca\n\ngenerated WITH a tool\n'
}

usage_cases() {
  cases=$((cases + 1))
  bash "$validator" > /dev/null 2> "$workdir/err"
  report "no argument is a usage error" 2 "$?"
  cases=$((cases + 1))
  bash "$validator" "$workdir/does-not-exist" > /dev/null 2> "$workdir/err"
  report "missing file is a usage error" 2 "$?"
}

accepted_cases
rejected_cases
usage_cases
if [ "$failures" -ne 0 ]; then
  printf '%s failure(s) in %s case(s)\n' "$failures" "$cases" >&2
  exit 1
fi
printf 'ok: %s case(s)\n' "$cases"
