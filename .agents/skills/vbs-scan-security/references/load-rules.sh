
set -euo pipefail

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PART_BUDGET=20000
PART=1
LANGS=()
while [ $# -gt 0 ]; do
  case "$1" in
    --part) PART="$2"; shift 2 ;;
    --part=*) PART="${1#--part=}"; shift ;;
    *) LANGS+=("$1"); shift ;;
  esac
done

rule_id() { sed -n 's/^id:[[:space:]]*//p' "$1" | head -1; }
is_opt_in() { sed -n '/^---$/,/^---$/p' "$1" | grep -q '^opt_in:'; }

overlay_for() {
  local dir="$SKILL_DIR/rules/languages/$1" f
  [ -d "$dir" ] || return 0
  for f in "$dir"/[0-9]*.md; do
    [ -f "$f" ] && [ "$(rule_id "$f")" = "$2" ] && { echo "$f"; return 0; }
  done
  return 0
}

detection() {
  local rel="${1#"$SKILL_DIR"/}"
  echo "=== RULE $(rule_id "$1") (source: $rel) ==="
  awk '/^## (Examples|Fix recommendation|Cross-references)/ { exit } { print }' "$1"
  echo
}

FILES_SELECTED=()

for generic in "$SKILL_DIR"/rules/generic/[0-9]*.md; do
  is_opt_in "$generic" && continue
  id="$(rule_id "$generic")"
  overlays=()
  all_covered=1
  for lang in ${LANGS[@]+"${LANGS[@]}"}; do
    o="$(overlay_for "$lang" "$id")"
    if [ -n "$o" ]; then overlays+=("$o"); else all_covered=0; fi
  done
  [ ${#LANGS[@]} -eq 0 ] && all_covered=0
  [ "$all_covered" -eq 1 ] || FILES_SELECTED+=("$generic")
  for o in ${overlays[@]+"${overlays[@]}"}; do FILES_SELECTED+=("$o"); done
done

part_of=()
current=1
used=0
for f in "${FILES_SELECTED[@]}"; do
  size=$(detection "$f" | wc -m | tr -d ' ')
  if [ "$used" -gt 0 ] && [ $((used + size)) -gt "$PART_BUDGET" ]; then
    current=$((current + 1))
    used=0
  fi
  part_of+=("$current")
  used=$((used + size))
done
total_parts=$current

if ! [ "$PART" -ge 1 ] 2>/dev/null || [ "$PART" -gt "$total_parts" ]; then
  echo "--part phải từ 1 đến $total_parts" >&2
  exit 1
fi

echo "=== vbsec rules — PART $PART/$total_parts. Đọc toàn bộ output này; không cắt bằng head/tail ==="
for i in "${!FILES_SELECTED[@]}"; do
  [ "${part_of[$i]}" -eq "$PART" ] && detection "${FILES_SELECTED[$i]}"
done
if [ "$PART" -lt "$total_parts" ]; then
  echo "=== PART $PART/$total_parts — CHƯA ĐỦ: chạy tiếp --part $((PART + 1)) đến --part $total_parts ==="
else
  echo "=== PART $PART/$total_parts — đã in đủ ${#FILES_SELECTED[@]} rule ==="
fi
