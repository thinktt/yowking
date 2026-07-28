#!/usr/bin/env node
import { spawn, spawnSync } from "node:child_process";
import { createReadStream, createWriteStream, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const args = parseArgs(process.argv.slice(2));
const repeats = positiveInteger(args.repeats || 5, "repeats");
const workers = positiveInteger(args.workers || 1, "workers");
const image = args.image || "zen:5000/yowking:prod-rnd0";
const cpuLimit = args.cpus || null;
if (cpuLimit !== null && !(Number(cpuLimit) > 0)) {
  throw new Error(`cpus must be a positive Docker CPU limit, got ${cpuLimit}`);
}
const cpusetCpus = args.cpuset || null;
const cgroupParent = args["cgroup-parent"] || null;
const runName = slug(args.run || `kingtc-benchmark-${stamp()}`);
const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const wrapperPath = resolve(repoRoot, "dist/enginewrap.exe");
const outDir = resolve(args["out-dir"] || resolve(repoRoot, "reports"));

if (!existsSync(wrapperPath)) {
  throw new Error(`Missing ${wrapperPath}; run task gobuild first`);
}
assertPerfAvailable();
mkdirSync(outDir, { recursive: true });

const results = [];
for (let repeat = 1; repeat <= repeats; repeat++) {
  process.stdout.write(`[${repeat}/${repeats}] `);
  const result = await profileBenchmark({ repeat, workers, image, cpuLimit, cpusetCpus, cgroupParent, wrapperPath });
  results.push(result);
  console.log(`wrapper=${result.wrapper.medianCpuNs}ns instructions=${result.counters.instructions.count}`);
}

const report = {
  schemaVersion: 1,
  createdAt: new Date().toISOString(),
  runName,
  benchmark: "enginewrap.exe --kingtc-benchmark; two warmups then five fixed prime-count passes to 1,500,000",
  environment: {
    hostname: readFileOrNull("/etc/hostname")?.trim() || null,
    cpuModel: cpuModel(),
    cpuLimit,
    cpusetCpus,
    cgroupParent,
    workers,
    image,
  },
  results,
  summary: summarize(results),
};
const jsonPath = resolve(outDir, `${runName}.json`);
const markdownPath = resolve(outDir, `${runName}.md`);
writeFileSync(jsonPath, `${JSON.stringify(report, null, 2)}\n`);
writeFileSync(markdownPath, renderMarkdown(report));
console.log(renderTerminalSummary(report.summary));
console.log(`Wrote ${jsonPath}`);

async function profileBenchmark({ repeat, workers, image, cpuLimit, cpusetCpus, cgroupParent, wrapperPath }) {
  const handles = Array.from({ length: workers }, (_, workerIndex) => startBenchmark({
    name: `kingtc-benchmark-${process.pid}-${repeat}-${workerIndex + 1}`,
    image,
    cpuLimit,
    cpusetCpus,
    cgroupParent,
    wrapperPath,
  }));
  const probe = handles[0];
  let perf;
  try {
    await Promise.all(handles.map((handle) => withTimeout(handle.ready, 15_000, "waiting for benchmark ready")));
    const wrapperPid = await waitForWrapperPid(probe.name, 5_000);
    perf = await controlledPerf(wrapperPid, ["instructions:u", "cycles:u", "branches:u", "branch-misses:u"]);
    await perf.enable();
    const startedAt = process.hrtime.bigint();
    for (const handle of handles) handle.start();
    await Promise.all(handles.map((handle) => handle.complete));
    const wallTimeMs = Number(process.hrtime.bigint() - startedAt) / 1_000_000;
    const perfResult = await perf.stop();
    return {
      repeat,
      workers,
      wrapperPid,
      wallTimeMs,
      wrapper: probe.benchmark(),
      contenderMedianCpuNs: handles.map((handle) => handle.benchmark().medianCpuNs),
      counters: perfResult.counters,
      perfDiagnostics: perfResult.diagnostics,
    };
  } catch (error) {
    if (perf) await perf.abort();
    for (const handle of handles) handle.abort();
    throw error;
  }
}

function startBenchmark({ name, image, cpuLimit, cpusetCpus, cgroupParent, wrapperPath }) {
  const dockerArgs = ["run", "--rm", "-i", "--name", name];
  if (cpuLimit) dockerArgs.push("--cpus", cpuLimit);
  if (cpusetCpus) dockerArgs.push("--cpuset-cpus", cpusetCpus);
  if (cgroupParent) dockerArgs.push("--cgroup-parent", cgroupParent);
  dockerArgs.push(
    "-v", `${wrapperPath}:/benchmark/enginewrap.exe:ro`,
    "--entrypoint", "/bin/sh",
    image,
    "-lc", "cd /benchmark && wine enginewrap.exe --kingtc-benchmark",
  );

  const child = spawn("docker", dockerArgs, { stdio: ["pipe", "pipe", "pipe"] });
  let stdout = "";
  let stderr = "";
  let benchmark = null;
  let readyResolve;
  let readyReject;
  const ready = new Promise((resolvePromise, rejectPromise) => {
    readyResolve = resolvePromise;
    readyReject = rejectPromise;
  });
  const complete = new Promise((resolvePromise, rejectPromise) => {
    child.stdout.on("data", (chunk) => {
      stdout += chunk.toString("utf8");
      for (const line of stdout.split(/\r?\n/)) {
        if (line === "kingtc-benchmark-ready") readyResolve();
        if (line.startsWith("kingtc-benchmark primes=")) benchmark = parseBenchmarkLine(line);
      }
    });
    child.stderr.on("data", (chunk) => { stderr += chunk.toString("utf8"); });
    child.on("error", rejectPromise);
    child.on("close", (code) => {
      if (code !== 0) {
        rejectPromise(new Error(`benchmark container exited ${code}: ${stderr.trim()}`));
        return;
      }
      if (!benchmark) {
        rejectPromise(new Error(`benchmark completed without result: ${stdout.trim()}`));
        return;
      }
      resolvePromise();
    });
  });

  return {
    name,
    ready,
    complete,
    benchmark: () => {
      if (!benchmark) throw new Error(`benchmark result is unavailable for ${name}`);
      return benchmark;
    },
    start: () => child.stdin.write("run\n"),
    abort: () => {
      child.kill("SIGTERM");
      spawnSync("docker", ["rm", "-f", name], { stdio: "ignore" });
      readyReject(new Error(`benchmark aborted: ${name}`));
    },
  };
}

async function waitForWrapperPid(containerName, timeoutMs) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const result = spawnSync("docker", ["top", containerName, "-eo", "pid,comm"], { encoding: "utf8" });
    if (result.status === 0) {
      for (const line of result.stdout.split("\n").slice(1)) {
        const [pid, command] = line.trim().split(/\s+/, 2);
        if (command === "enginewrap.exe" && /^\d+$/.test(pid)) return Number(pid);
      }
    }
    await delay(25);
  }
  throw new Error(`Timed out locating enginewrap.exe in ${containerName}`);
}

async function controlledPerf(pid, events) {
  const controlDir = mkdtempSync(resolve(tmpdir(), "kingtc-benchmark-perf-"));
  const controlPath = resolve(controlDir, "control");
  const ackPath = resolve(controlDir, "ack");
  spawnSync("mkfifo", [controlPath, ackPath], { stdio: "inherit" });
  const child = spawn("sudo", [
    "-n", "perf", "stat", "-x,", "--no-big-num", "--delay=-1",
    "--control", `fifo:${controlPath},${ackPath}`,
    "-e", events.join(","), "-p", String(pid),
  ], { stdio: ["ignore", "ignore", "pipe"] });
  let stderr = "";
  child.stderr.on("data", (chunk) => { stderr += chunk.toString("utf8"); });
  const closed = new Promise((resolvePromise) => child.on("close", (code, signal) => resolvePromise({ code, signal })));
  const ack = createReadStream(ackPath);
  const control = createWriteStream(controlPath);
  await new Promise((resolvePromise, rejectPromise) => {
    control.once("open", resolvePromise);
    control.once("error", rejectPromise);
  });
  let ackBuffer = "";
  const waiters = [];
  ack.on("data", (chunk) => {
    ackBuffer += chunk.toString("utf8").replace(/\0/g, "");
    while (ackBuffer.includes("ack\n") && waiters.length) {
      ackBuffer = ackBuffer.replace("ack\n", "");
      waiters.shift()();
    }
  });
  const command = (value) => new Promise((resolvePromise, rejectPromise) => {
    const timer = setTimeout(() => rejectPromise(new Error(`Timed out waiting for perf ${value} acknowledgement`)), 5_000);
    waiters.push(() => {
      clearTimeout(timer);
      resolvePromise();
    });
    control.write(`${value}\n`);
  });
  const cleanup = () => {
    control.destroy();
    ack.destroy();
    rmSync(controlDir, { recursive: true, force: true });
  };
  return {
    enable: () => command("enable"),
    stop: async () => {
      child.kill("SIGINT");
      const status = await closed;
      cleanup();
      return { status, counters: parsePerf(stderr), diagnostics: stderr.split("\n").filter((line) => line && !line.includes(",")).join("\n") };
    },
    abort: async () => {
      child.kill("SIGINT");
      await closed;
      cleanup();
    },
  };
}

function parsePerf(stderr) {
  const counters = {};
  for (const line of stderr.split("\n")) {
    const fields = line.split(",");
    if (fields.length < 5 || !fields[2]?.includes(":")) continue;
    const name = fields[2].replace(/:u$/, "");
    counters[name] = { count: /^\d+$/.test(fields[0]) ? Number(fields[0]) : null };
  }
  for (const name of ["instructions", "cycles", "branches", "branch-misses"]) {
    if (!(counters[name]?.count >= 0)) throw new Error(`perf did not count ${name}: ${stderr.trim()}`);
  }
  return counters;
}

function parseBenchmarkLine(line) {
  const match = line.match(/^kingtc-benchmark primes=(\d+) limit=(\d+) warmups=(\d+) runs=(\d+) cpu-ns=\[([\d ]*)\] median-cpu-ns=(\d+)$/);
  if (!match) throw new Error(`Could not parse benchmark output: ${line}`);
  return {
    primes: Number(match[1]),
    limit: Number(match[2]),
    warmups: Number(match[3]),
    runs: Number(match[4]),
    cpuNs: match[5].trim().split(/\s+/).filter(Boolean).map(Number),
    medianCpuNs: Number(match[6]),
  };
}

function summarize(results) {
  const values = (selector) => results.map(selector);
  return {
    repeats: results.length,
    wrapperMedianCpuNs: values((result) => result.wrapper.medianCpuNs),
    wallTimesMs: values((result) => result.wallTimeMs),
    instructions: values((result) => result.counters.instructions.count),
    cycles: values((result) => result.counters.cycles.count),
    instructionRange: range(values((result) => result.counters.instructions.count)),
    wrapperCpuRangeNs: range(values((result) => result.wrapper.medianCpuNs)),
  };
}

function renderMarkdown(report) {
  const lines = [
    "# KingTC Wrapper Benchmark",
    "",
    `Generated: ${report.createdAt}`,
    "",
    `Image: \`${report.environment.image}\``,
    `CPU set: \`${report.environment.cpusetCpus || "unrestricted"}\``,
    "",
    "| Repeat | Wrapper median CPU ns | Wall ms | Retired instructions | Cycles |",
    "|---:|---:|---:|---:|---:|",
    ...report.results.map((result) => `| ${result.repeat} | ${result.wrapper.medianCpuNs} | ${result.wallTimeMs.toFixed(1)} | ${result.counters.instructions.count} | ${result.counters.cycles.count} |`),
    "",
    `Wrapper CPU range: ${report.summary.wrapperCpuRangeNs} ns`,
    `Instruction range: ${report.summary.instructionRange}`,
    "",
  ];
  return `${lines.join("\n")}\n`;
}

function renderTerminalSummary(summary) {
  return [
    "\nKingTC benchmark baseline",
    `wrapper median CPU ns: ${summary.wrapperMedianCpuNs.join(", ")}`,
    `wall time ms:           ${summary.wallTimesMs.map((value) => value.toFixed(1)).join(", ")}`,
    `retired instructions:   ${summary.instructions.join(", ")}`,
    `wrapper CPU range:      ${summary.wrapperCpuRangeNs} ns`,
    `instruction range:      ${summary.instructionRange}`,
  ].join("\n");
}

function parseArgs(values) {
  const parsed = {};
  for (let index = 0; index < values.length; index += 2) {
    const key = values[index];
    if (!key?.startsWith("--") || !values[index + 1]) throw new Error(`Expected --key value, got ${values.slice(index).join(" ")}`);
    parsed[key.slice(2)] = values[index + 1];
  }
  return parsed;
}

function assertPerfAvailable() {
  const result = spawnSync("sudo", ["-n", "perf", "--version"], { encoding: "utf8" });
  if (result.status === 0) return;
  const detail = [result.stdout, result.stderr].filter(Boolean).join(" ").trim();
  throw new Error(`host perf is unavailable: ${detail || `exit ${result.status}`}`);
}

function positiveInteger(value, name) {
  const number = Number(value);
  if (!Number.isInteger(number) || number < 1) throw new Error(`${name} must be a positive integer`);
  return number;
}

function range(values) {
  return Math.max(...values) - Math.min(...values);
}

function cpuModel() {
  return readFileOrNull("/proc/cpuinfo")?.match(/^model name\s*: (.+)$/m)?.[1] || null;
}

function readFileOrNull(path) {
  try {
    return readFileSync(path, "utf8");
  } catch {
    return null;
  }
}

function slug(value) {
  return value.replace(/[^a-zA-Z0-9._-]+/g, "-");
}

function stamp() {
  return new Date().toISOString().replace(/[:.]/g, "-");
}

function delay(ms) {
  return new Promise((resolvePromise) => setTimeout(resolvePromise, ms));
}

async function withTimeout(promise, timeoutMs, description) {
  let timer;
  try {
    return await Promise.race([
      promise,
      new Promise((_, rejectPromise) => { timer = setTimeout(() => rejectPromise(new Error(`Timed out ${description}`)), timeoutMs); }),
    ]);
  } finally {
    clearTimeout(timer);
  }
}
