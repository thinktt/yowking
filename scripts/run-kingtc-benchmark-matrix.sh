#!/usr/bin/env bash
set -Eeuo pipefail

host_label="${HOST_LABEL:?set HOST_LABEL to identify this host}"
out_dir="${OUT_DIR:?set OUT_DIR to the collection directory}"
repeats="${REPEATS:-3}"
isolated_cpus="${ISOLATED_CPUS:-14,15}"
isolated_slice="${ISOLATED_SLICE:-king-benchmark-isolated.slice}"

mkdir -p "$out_dir"

run_profile() {
  local name=$1
  shift
  node scripts/profile-kingtc-benchmark.mjs \
    --run "${host_label}-${name}" \
    --repeats "$repeats" \
    --out-dir "$out_dir" \
    "$@"
}

run_isolated() {
  local name=$1
  shift
  ISOLATED_CPUS="$isolated_cpus" \
  ISOLATED_SLICE="$isolated_slice" \
  scripts/run-isolated-benchmark.sh \
    node scripts/profile-kingtc-benchmark.mjs \
      --run "${host_label}-${name}" \
      --repeats "$repeats" \
      --out-dir "$out_dir" \
      --cpuset "$isolated_cpus" \
      --cgroup-parent "$isolated_slice" \
      "$@"
}

run_isolated "isolated-full" --workers 1
run_isolated "isolated-quarter" --workers 1 --cpus 0.25
run_profile "ordinary-full" --workers 1
run_profile "ordinary-quarter" --workers 1 --cpus 0.25
run_profile "contention-eight" --workers 8
