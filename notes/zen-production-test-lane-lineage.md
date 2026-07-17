# Zen Production Test-Lane Lineage

Date: 2026-07-17

## Scope

This is a local, read-only investigation of the Zen Yowking production image.
It does not modify Zen or publish a replacement image. Its purpose is to define
the smallest credible worker replacement for a deterministic `rnd=0` strength
test against Zen's existing engine lane.

## Production Evidence

The running Zen worker `yowking45` reports:

```text
image reference: zen:5000/yowking
image ID:        sha256:0714e5c95fe40daab8e948a822185ec3340fd5d24985223effcbd1ddd75d79d2
image created:   2024-11-12T15:13:05.85153081Z
repository digest: zen:5000/yowking@sha256:e178bb1a7d717176e33ae3079376b91393f2fc39dec60382be5166e494a5e374
```

The registry retained this exact manifest under `zen:5000/yowking:previous`.
It was verified locally to have the same image ID, creation timestamp, manifest
digest, and runtime artifact hashes as `yowking45`. It is now also preserved as
the stable tag `zen:5000/yowking:prod`:

```text
zen:5000/yowking:prod@sha256:e178bb1a7d717176e33ae3079376b91393f2fc39dec60382be5166e494a5e374
```

Its important runtime artifact hashes are:

```text
bfe695d9eb0449311c6a08120dbb2de14301fb8b016a489a46db542c66cbc8d8  kingworker
0b83bd4519d649f5b4ab575971d278c53a2b39b65afa80addd2da5ee07e4160b  enginewrap.exe
e115e4dbc42b9ec8f880f214f089abc301e8118dd697153ed8f0b365098d2f38  TheKing350noOpk.exe
```

Go build metadata embedded in the production `kingworker` records source
revision `0368cd291e962ee2990b0e76bd3f320c3a51e7b7` on 2023-09-20, with
`vcs.modified=true`. That commit is not in the current checkout. The 2024 image
therefore contains older, locally modified worker artifacts; the image build
date is not the source revision date.

The cached Bee copy of `zen:5000/yowking:latest` is not the same image ID as
the running Zen image, but its `kingworker`, wrapper, and King executable hashes
match the running Zen artifacts above. It was useful before the exact `previous`
tag was identified, but future probes should use `zen:5000/yowking:prod`.

## Runtime Contract

Both the cached production payload and `zen:5000/yowking-nt:test3` use:

```text
Alpine Linux 3.21.3
wine-9.17-r1
musl-1.2.5-r9
xvfb-21.1.16-r0
ENG_CMD=/usr/bin/wine enginewrap.exe
```

All 74 opening-book files have identical hashes between those images. The
personality JSON files are the same size but have different hashes, so a Zen
test lane must preserve the production personality file rather than inherit the
KingNT test image copy.

Zen's live worker uses a mounted calibration file:

```json
{"Easy":4100,"Hard":6350,"GM":9650}
```

The cached image itself contains `Hard: 5750`; it is not the active Zen clock
configuration. A derived test lane must retain or mount the live Zen clocks.

## Wrapper Compatibility Probe

An ephemeral local container based on the cached production payload ran the
same direct, out-of-book Orin request twice at `clockTime: 6350` and `rnd=0`:

```json
{
  "moves": ["e2e4", "c7c5", "g1f3", "b8c6", "d2d4", "c5d4", "f3d4", "e7e6"]
}
```

Results:

| Wrapper | Engine bytes | Result |
| --- | --- | --- |
| Production `enginewrap.exe` | `TheKing350noOpk.exe` | `d4b5`, id `341518` |
| Current `enginewrap.exe` | Alias `TheKing350.exe -> TheKing350noOpk.exe` | `d4b5`, id `341518` |

The current wrapper is functionally compatible with the production Wine/runtime
when an alias supplies the filename it expects. The actual King executable bytes
remain the production `TheKing350noOpk.exe` bytes.

## Replacement Choices

### Functional minimum: `kingworker` only

Replacing only `kingworker` is compatible with the production wrapper and
engine naming. It enables worker tags and `FORCE_RANDOM_OFF=true`, while
preserving Zen's Wine, wrapper, engine, books, personalities, and mounted
clocks. This is the smallest change for a narrow `rnd=0` comparison.

The replacement must be built with the existing container build path or with
`CGO_ENABLED=0`. A default host build is dynamically linked against glibc and
is not a valid Alpine artifact. A `CGO_ENABLED=0` build of the current branch
is a static Linux amd64 executable and is compatible with the production
Alpine runtime.

### Recommended test-lane minimum: worker + wrapper + engine-name alias

For a durable lane, replace:

1. `kingworker` with the current static Linux worker.
2. `enginewrap.exe` with the current wrapper.
3. Add `TheKing350.exe -> TheKing350noOpk.exe` as a symlink or byte-identical
   copy.

Leave `TheKing350noOpk.exe`, Wine, books, production personalities, and Zen
clock volume unchanged. This includes the current worker/wrapper cleanup path
that was introduced after the Zen image was built, without changing the actual
King engine under test.

## Lifecycle Caveat

The direct local probe did show exited Wine/King children under the ephemeral
container's `sleep` PID 1 after each one-shot `docker exec` run. That layout is
not representative of a long-lived `kingworker` as container PID 1, so it does
not prove or disprove the current worker's `reapExitedWineChildren` behavior.

Before using the recommended lane for a long strength run, validate it locally
with a real NATS worker process handling repeated requests, then check that its
process count and zombie count remain flat. The production `kingworker` alone
does not include the current wrapper's explicit stdin close and `Wait` behavior,
so it is not the preferred long-running test lane.

## Built Test Image

The production-derived test image has been built and published as:

```text
zen:5000/yowking:prod-rnd0@sha256:9ffa56b9339306e42ff366c54b04dea1924e2cce94a47655fe6ed63fd1882ac6
```

It is based directly on `zen:5000/yowking:prod` and changes only:

```text
kingworker       current static worker, including FORCE_RANDOM_OFF support
enginewrap.exe   current wrapper
TheKing350.exe   symlink to unchanged TheKing350noOpk.exe
```

Five direct out-of-book Orin probes at `clockTime: 6350` and `rnd=0` were run
through both `yowking:prod` and `yowking:prod-rnd0`. Every run matched on
coordinate move (`d4b5`), algebraic move (`Nb5`), depth (`7003`), evaluation
(`-100`), id (`341518`), and draw decision. The only variation was the
unpatched engine's real-clock `time` field: production reported 28 or 29, and
the derived image reported 29 centiseconds.

## Next Step

Build a local production-derived image that changes only the recommended files,
mounts the verified Zen clocks, and runs a sustained tagged NATS smoke test.
Only after that passes should an equivalent isolated Zen test service be
considered.
