# Command convert

Re-simulate downloaded replay JSON files through the arena engine and write verified arena trace files.

The output traces are the same format `arena run --trace` produces and feed into `arena analyze` and the web viewer.

## Quick start

```shell
# Convert every replay in ./replays/
bin/arena convert --game=winter2026

# Convert specific replay IDs
bin/arena convert --game=winter2026 875142454 875142455
```

## How it works

For each replay file:

1. Parse the replay JSON and extract the seed from `refereeInput`.
2. Skip if the replay's `puzzleId` doesn't match the selected game.
3. Re-run the engine using the recorded player moves and the league parsed from the replay title (or `--league` override).
4. Verify final scores and turn count match the replay; if they diverge, skip without writing.
5. Write the verified trace to `--trace-dir` as a replay-typed trace file keyed by replay ID.

Existing trace files are skipped unless `--force` is set.

## Options

| Flag            | Default      | Description                                                              |
|-----------------|--------------|--------------------------------------------------------------------------|
| `-l, --league`  | from replay  | League level override (default: parsed from replay title)                |
| `--replay-dir`  | `./replays`  | Directory to scan for replay JSON files                                  |
| `--trace-dir`   | `./traces`   | Directory to write converted trace files                                 |
| `-f, --force`   | `false`      | Overwrite existing trace files                                           |

Positional args (optional): one or more replay IDs to convert. If omitted, every `<gameId>.json` file in `--replay-dir` is processed.

## Output

Per-replay status lines, then a final summary:

```
[1/3] save 875142454 (league=4 turns=187 scores=24.0:18.0)
[2/3] skip 875142455 (trace exists)
[3/3] skip 875142456 (replay mismatch: score mismatch: replay=[...] engine=[...])
done: 1 saved, 1 skipped-existing, 0 skipped-puzzle, 1 skipped-mismatch (replays=3 out=./traces)
```

Skip categories:

- **skipped-existing** — trace already on disk (use `--force` to overwrite)
- **skipped-puzzle** — replay belongs to a different game
- **skipped-mismatch** — engine output (scores or turn count) disagrees with the replay; usually means the engine doesn't yet match the recorded league behavior

A mismatch is logged but does not abort the batch.
