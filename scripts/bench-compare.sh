#!/usr/bin/env bash
# Compares benchmarks between two checkouts with benchstat and writes a
# Markdown report. Runs alternate between base and head so runner noise
# affects both sides equally.
#
#   BASE_DIR=../base HEAD_DIR=. REPORT_FILE=report.md scripts/bench-compare.sh
set -euo pipefail

: "${BASE_DIR:?BASE_DIR is required}"
: "${HEAD_DIR:?HEAD_DIR is required}"
: "${REPORT_FILE:?REPORT_FILE is required}"
count="${COUNT:-6}"
benchtime="${BENCH_TIME:-500ms}"
base_sha="${BASE_SHA:-$(git -C "$BASE_DIR" rev-parse HEAD)}"
head_sha="${HEAD_SHA:-$(git -C "$HEAD_DIR" rev-parse HEAD)}"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

(cd "$BASE_DIR" && go test -c -o "$work/base.test" .)
(cd "$HEAD_DIR" && go test -c -o "$work/head.test" .)

run() {
	(cd "$1" && "$2" -test.run '^$' -test.bench . -test.benchmem -test.benchtime "$benchtime" -test.count 1)
}
for _ in $(seq "$count"); do
	run "$BASE_DIR" "$work/base.test" >>"$work/base.txt"
	run "$HEAD_DIR" "$work/head.test" >>"$work/head.txt"
done

{
	echo '<!-- xterm-go-benchstat -->'
	echo '## xterm-go benchmarks'
	echo
	echo "Comparing \`${base_sha:0:12}\` (base) with \`${head_sha:0:12}\` (PR) on the same runner: $count alternating runs at $benchtime per benchmark."
	echo
	echo '```text'
	benchstat -ignore pkg base="$work/base.txt" pr="$work/head.txt"
	echo '```'
} >"$REPORT_FILE"
