# yowdeps

Build pipeline for Ye Old Wizard dependency artifacts.

This repo prepares Chessmaster-derived assets used by the Ye Old Wizard runtime, primarily for:
- https://github.com/thinktt/yowking

Primary purpose: prepare dependency artifacts used by the yowworker in that repo.

## What It Builds

`task build:dist` produces `../dist` (repo root `dist/`):
- `dist/books/*.bin`: opening books converted from Chessmaster `.obk`
- `dist/personalities.json`: personality metadata derived from `.CMP` + `assets/personalities.cfg`
- `dist/TheKing350noOpk.exe`: patched King executable with OPK disabled
- `dist/runbook`: Alpine Linux binary built from the JavaScript runbook for opening-book move lookup

## Prerequisites

- Docker (daemon running)
- Taskfile (`task`) CLI
- Local Chessmaster install folder with `TheKing350.exe`, `Data/Opening Books/`, and `Data/Personalities/`

Install Taskfile (`task`) if needed:

```bash
npm install --global @go-task/cli
```

For other installation methods, use:
- https://taskfile.dev/installation/

## Setup

Create or update `.env` in repo root with your Chessmaster path:

```bash
CM_DIR=/path/to/your/Chessmaster/install
```

Then import Chessmaster assets:

```bash
task cm:import
```

This populates `assets/cm` with books, personalities, and `TheKing350.exe`.

## Build

Run:

```bash
task build:dist
```

Build flow:
1. Builds Docker image `yowdeps`
2. Runs container pipeline (`scripts/build-all.sh`)
3. Generates all artifacts in `/build/dist` inside container
4. Copies artifacts back to local repo root `../dist`

## Runbook Book-Path Behavior

`dist/runbook` resolves books relative to the runbook binary location, not the current working directory.

It always looks for:
- `<runbook directory>/books/<bookName>`

Example:

```bash
../dist/runbook '{"moves":["d2d4"],"book":"EarlyQueen.bin"}'
```

## Useful Tasks

```bash
task --list
```

- `cm:import`: import Chessmaster files from `CM_DIR` into `assets/cm`
- `build:dist`: build all distributable artifacts
- `clean:all`: cleanup assets, dist, and related Docker artifacts
