# Brutaltester Referee Decoupling Plan

## Summary

Implement brutaltester-compatible referee mode as the v1 contract.

- Add `arena referee <game>` as a standalone referee executable mode.
- Make normal `arena run <game>` use a referee subprocess by default, with the default command pointing back to the current binary: `bin/arena referee <game>`.
- Keep embedded Go execution available for parity, debug, and trace workflows.
- Support external Java referees in arena via `--referee-cmd "java -jar referee.jar"` for registered 2-player games.
- Keep v1 result-only for external referees: no viewer trace reconstruction yet.

## Key Changes

- Add brutaltester CLI parsing for `arena referee <game>`:
  - `-p1`, `-p2` required for current games.
  - Accept up to `-p8`, but return a clear unsupported-player-count error unless the selected game supports it.
  - Support `-seed <int>`, `--seed <int>`, and brutaltester seed form `-d seed=<int>`.
  - Support `-league <int>` / `--league <int>`.
  - Support `-l <path>` to write a CodinGame-runner-like JSON log.
  - Support `-h`.
- Refactor match execution behind a small internal interface:
  - `EmbeddedExecutor`: current in-process Go runner.
  - `BrutaltesterExecutor`: shells out to a referee command, parses stdout scores, and optionally reads the `-l` JSON log.
  - `arena run` normal batches default to `BrutaltesterExecutor` using self-command.
  - `arena run --engine=embedded` preserves the old path.
  - `--debug` and `--trace` use embedded execution until external trace conversion exists.
- Define referee stdout contract:
  - Print one integer score per player, in `p1`, `p2`, ... order.
  - Do not print bot stderr to referee stderr during successful matches, because brutaltester treats referee stderr as a match error.
  - Any fatal referee failure exits non-zero and writes a concise error to stderr.
- Define `-l` JSON contract:
  - Write at least `scores`: object keyed by `"0"`, `"1"`, ... with integer final scores.
  - Write `errors`: object keyed by `"0"`, `"1"`, ... with arrays of captured bot stderr strings.
  - Write `failCause` only on fatal referee failure.
  - This is enough for cgarena-style wrappers that inspect `scores` and bot stderr attributes.
- Add arena-side external referee support:
  - `arena run <game> --referee-cmd "java -jar referee.jar"` runs Java/brutaltester-compatible referees.
  - `<game>` still supplies arena metadata, defaults, and registered game identity.
  - Fully registry-less external games are out of v1.

## Test Plan

- Unit tests:
  - Parse `-p1/-p2`, `-seed`, `--seed`, `-d seed=...`, and league flags.
  - Reject missing players, unknown games, bad seeds, and unsupported player counts.
  - Verify stdout score order matches `p1`, `p2`.
  - Verify `-l` JSON includes `scores` and `errors`.
- Integration tests:
  - Run `arena referee winter2026 -p1 <bot> -p2 <bot> -seed <seed>`.
  - Compare one fixed-seed embedded match vs self-referee subprocess match.
  - Run `arena run winter2026 --engine=embedded` and default subprocess mode on the same seed and compare summary result.
  - Run a fake brutaltester-compatible referee command through `--referee-cmd`.
- Validation:
  - `make test-arena`
  - `make lint-arena`

## Assumptions

- v1 targets 2-player registered games first.
- External Java referees are result-only in v1.
- Existing trace/debug behavior remains embedded until a later Java-log-to-arena-trace adapter exists.
- Sources inspected:
  - https://github.com/aangairbender/cgarena/blob/master/docs/configuration.md
  - https://github.com/aangairbender/cgarena/blob/master/assets/play_game.py
  - https://github.com/aangairbender/cgarena/blob/master/docs/making_bt_compatible_referee.md
  - https://github.com/dreignier/cg-brutaltester
