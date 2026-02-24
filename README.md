# The King Worker (yowking)

This repo packages Johan de Koning's chess engine [The King](https://www.chessprogramming.org/The_King) (from Chessmaster) as a worker service.

The worker:
- consumes move requests from NATS
- selects book moves / personality settings
- runs The King engine
- replies with a move result

This repo is now worker-focused (engine wrapper + dependency build pipeline + container build/deploy helpers).

## Related Repos

Deploy `yowapi` first (API/orchestrator), then deploy this worker service.

- `yowapi` (API / game server / orchestrator): `https://github.com/thinktt/yowapi`

`yowapi` owns users/games/domain logic and sends move requests to this worker via NATS.

## What This Repo Builds

The runtime `dist/` folder contains:

- `dist/kingworker` (Linux worker binary)
- `dist/enginewrap.exe` (Wine-side wrapper for engine IO)
- `dist/TheKing350.exe` (King engine)
- `dist/runbook` (opening-book helper binary)
- `dist/books/*.bin` (Polyglot-style opening book artifacts)
- `dist/personalities.json` (personality metadata/settings)
- `dist/calibrations/clockTimes.json` (copied calibration timings)

## Prerequisites

- Docker
- Task (`task`) CLI
- Go (for local `gobuild` / `build:dist` steps)
- A local Chessmaster install (for asset import)

Install Task if needed:

```bash
npm install --global @go-task/cli
```

## Environment

This repo uses a single root `.env`.

Minimum values you are likely to need:

```bash
CM_DIR=/path/to/Chessmaster/install
NATS_TOKEN=your-dev-nats-token
```

`CM_DIR` is used by `import:cm`.
`NATS_TOKEN` is used by the dev worker compose task.

## Top-Level Workflow

Current top-level steps:

1. Import Chessmaster assets
2. Build `dist/`
3. Build the main worker image
4. (Optional / current dev flow) deploy one dev worker

Tasks:

```bash
task import:cm
task build:dist
task build:image
task deploy        # deploys one dev worker for now
```

Or run the full dev pipeline:

```bash
task dev:up
```

## Task Summary

List tasks:

```bash
task --list
```

Most important tasks:

- `import:cm` - import Chessmaster engine/books/personalities into `assets/cm`
- `build:dist` - builds dependency artifacts + worker binaries into `dist/`
- `build:image` - builds `zen:5000/yowking:latest`
- `deploy:dev` - starts one worker from `deploy/compose.dev.yaml`
- `deploy:dev:down` - stops the dev worker
- `clean:all` - reset generated artifacts + imported CM assets + helper image

## Notes

- `deploy/compose.yowking.yaml` is generated and ignored.
- `deploy/compose.dev.yaml` is the simple single-worker dev compose file.
