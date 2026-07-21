# KingTC Benchmark Matrix

Collection: `reports/kingtc-benchmark-matrix-2026-07-21-three-host`

Each row is the median of three wrapper launches. Each wrapper launch contains two warmups and five measured prime-count passes.

| Host | Lane | CPU quota | Wrappers | Thread CPU / pass | Wall / sequence | Retired instructions | Instruction spread |
|---|---|---:|---:|---:|---:|---:|---:|
| ace | contention-eight | full | 8 | 680.0 ms | 10806.7 ms | 10150509120 | 16.2 ppm |
| ace | isolated-full | full | 1 | 410.0 ms | 3156.2 ms | 10140990907 | 2.7 ppm |
| ace | isolated-quarter | 0.25 | 1 | 420.0 ms | 13154.4 ms | 10141691043 | 17.9 ppm |
| ace | ordinary-full | full | 1 | 410.0 ms | 3114.6 ms | 10142257620 | 6.6 ppm |
| ace | ordinary-quarter | 0.25 | 1 | 490.0 ms | 15857.8 ms | 10145120305 | 47.9 ppm |
| bee | contention-eight | full | 8 | 200.0 ms | 2461.4 ms | 10140948890 | 23.0 ppm |
| bee | isolated-full | full | 1 | 370.0 ms | 2985.0 ms | 10140635661 | 16.3 ppm |
| bee | isolated-quarter | 0.25 | 1 | 380.0 ms | 11778.1 ms | 10141078950 | 27.4 ppm |
| bee | ordinary-full | full | 1 | 170.0 ms | 1366.1 ms | 10140107157 | 18.4 ppm |
| bee | ordinary-quarter | 0.25 | 1 | 170.0 ms | 5258.5 ms | 10141424443 | 71.4 ppm |
| ryz | contention-eight | full | 8 | 190.0 ms | 2392.7 ms | 10140839579 | 26.4 ppm |
| ryz | isolated-full | full | 1 | 380.0 ms | 3060.0 ms | 10140618955 | 4.0 ppm |
| ryz | isolated-quarter | 0.25 | 1 | 380.0 ms | 11858.1 ms | 10140973160 | 1.8 ppm |
| ryz | ordinary-full | full | 1 | 160.0 ms | 1338.0 ms | 10140121032 | 18.6 ppm |
| ryz | ordinary-quarter | 0.25 | 1 | 170.0 ms | 5187.4 ms | 10141221240 | 24.3 ppm |

`Thread CPU / pass` is the wrapper-reported median of its five measured passes. `Wall / sequence` covers the two warmups and five measured passes. The contention lane profiles one probe while seven identical wrapper benchmarks run concurrently.

