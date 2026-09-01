#!/usr/bin/env node
import { readdirSync, readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";

const directory = process.argv[2];
if (!directory) {
  throw new Error("usage: node scripts/summarize-kingtc-benchmark-matrix.mjs REPORT_DIRECTORY");
}

const reports = findJson(resolve(directory)).map((path) => JSON.parse(readFileSync(path, "utf8")))
  .filter((report) => report.schemaVersion === 1 && Array.isArray(report.results));
if (!reports.length) throw new Error(`No benchmark reports found in ${directory}`);

const rows = reports.map((report) => {
  const wrapperCpuMs = report.results.map((result) => result.wrapper.medianCpuNs / 1_000_000);
  const wallMs = report.results.map((result) => result.wallTimeMs);
  const instructions = report.results.map((result) => result.counters.instructions.count);
  return {
    host: report.environment.hostname,
    lane: report.runName.replace(/^[^-]+-/, ""),
    cpus: report.environment.cpuLimit || "full",
    workers: report.environment.workers,
    wrapperCpuMs: median(wrapperCpuMs),
    wallMs: median(wallMs),
    instructions: median(instructions),
    instructionSpreadPpm: relativeRange(instructions) * 1_000_000,
  };
}).sort((left, right) => `${left.host}:${left.lane}`.localeCompare(`${right.host}:${right.lane}`));

const lines = [
  "# KingTC Benchmark Matrix",
  "",
  `Collection: \`${directory}\``,
  "",
  "Each row is the median of three wrapper launches. Each wrapper launch contains two warmups and five measured prime-count passes.",
  "",
  "| Host | Lane | CPU quota | Wrappers | Thread CPU / pass | Wall / sequence | Retired instructions | Instruction spread |",
  "|---|---|---:|---:|---:|---:|---:|---:|",
  ...rows.map((row) => `| ${row.host} | ${row.lane} | ${row.cpus} | ${row.workers} | ${row.wrapperCpuMs.toFixed(1)} ms | ${row.wallMs.toFixed(1)} ms | ${Math.round(row.instructions)} | ${row.instructionSpreadPpm.toFixed(1)} ppm |`),
  "",
  "`Thread CPU / pass` is the wrapper-reported median of its five measured passes. `Wall / sequence` covers the two warmups and five measured passes. The contention lane profiles one probe while seven identical wrapper benchmarks run concurrently.",
  "",
];
const output = resolve(directory, "matrix-summary.md");
writeFileSync(output, `${lines.join("\n")}\n`);
console.log(lines.join("\n"));
console.log(`Wrote ${output}`);

function findJson(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = resolve(directory, entry.name);
    if (entry.isDirectory()) return findJson(path);
    return entry.name.endsWith(".json") ? [path] : [];
  });
}

function median(values) {
  const sorted = [...values].sort((left, right) => left - right);
  const middle = Math.floor(sorted.length / 2);
  return sorted.length % 2 ? sorted[middle] : (sorted[middle - 1] + sorted[middle]) / 2;
}

function relativeRange(values) {
  const center = median(values);
  return center ? (Math.max(...values) - Math.min(...values)) / center : 0;
}
