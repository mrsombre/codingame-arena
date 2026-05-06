/** Pac type enum — matches engine "ROCK"/"PAPER"/"SCISSORS"/"NEUTRAL"/"DEAD". */
export type PacType = "ROCK" | "PAPER" | "SCISSORS" | "NEUTRAL" | "DEAD"

export interface Coord {
  x: number
  y: number
}

export interface Pac {
  id: number
  mine: boolean
  x: number
  y: number
  type: PacType
  abilityDuration: number
  abilityCooldown: number
}

/** Pellet entry with its point value (1 regular, 10 cherry). */
export interface Pellet {
  x: number
  y: number
  value: number
}

/** Parsed output from /api/serialize (global info + first frame). */
export interface MapData {
  width: number
  height: number
  /** Row-major grid: true = wall, false = floor. */
  walls: boolean[]
  myScore: number
  oppScore: number
  pacs: Pac[]
  pellets: Pellet[]
}

/** Dynamic per-turn state. */
export interface FrameData {
  myScore: number
  oppScore: number
  pacs: Pac[]
  pellets: Pellet[]
}

/** Trace JSON from GET /api/matches/{id}. */
export interface TraceMatch {
  match_id: number
  seed: string
  /** CodinGame-style ranks: 0 = first place, [0,0] = draw. */
  ranks: [number, number]
  scores: [number, number]
  players: [string, string]
  timing?: TraceTiming
  turns: TraceTurn[]
}

/** Returns 0/1 winner index, or -1 for a draw, given a TraceMatch.ranks pair. */
export function winnerFromRanks(ranks: [number, number]): number {
  if (ranks[0] === 0 && ranks[1] === 0) return -1
  return ranks[0] === 0 ? 0 : 1
}

export interface TraceTiming {
  first_response: [number, number]
  response_average: [number, number]
  response_median: [number, number]
}

export interface TraceTurn {
  turn: number
  /** Stdin lines fed to the blue side this turn (the user's bot — see
   * TraceMatch.blue). Absent on speed sub-turns and turns where blue did
   * not execute. */
  game_input?: string[]
  /** Raw stdout per side: [left, right]. Empty entry means the side did
   * not execute this turn. Absent when both sides were silent. */
  output?: [string, string]
  timing?: TraceTurnTiming
  /** Per-player trace events. Index 0 owns player 0's events; index 1 player 1. */
  traces?: [TurnTrace[], TurnTrace[]]
}

export interface TraceTurnTiming {
  response: [number, number]
}

export interface TurnTrace<M = unknown> {
  type: string
  data?: M
}

export interface PacMeta {
  pac: number
}

export interface EatMeta {
  pac: number
  coord: [number, number]
  cost: number
}

export interface KilledMeta {
  pac: number
  coord: [number, number]
  killer: number
}

export interface SwitchMeta {
  pac: number
  type: string
}

/**
 * Parse frame lines from a player's game_input.
 *
 * Format:
 *   <myScore> <oppScore>
 *   <pacCount>
 *   <id> <mine> <x> <y> <type> <abilDur> <abilCool>  (per pac)
 *   <pelletCount>
 *   <x> <y> <value>  (per pellet)
 */
export function parseFrameLines(lines: string[]): FrameData {
  let i = 0
  const next = (): string => {
    if (i >= lines.length) {
      throw new Error(`unexpected end of frame input at line ${i}`)
    }
    const line = lines[i]
    if (line === undefined) {
      throw new Error(`unexpected empty frame input at line ${i}`)
    }
    i++
    return line
  }

  const scoreParts = next().split(/\s+/)
  const myScore = Number.parseInt(scoreParts[0] ?? "0", 10)
  const oppScore = Number.parseInt(scoreParts[1] ?? "0", 10)

  const pacCount = Number.parseInt(next(), 10)
  const pacs: Pac[] = []
  for (let p = 0; p < pacCount; p++) {
    const parts = next().split(/\s+/)
    pacs.push({
      id: Number.parseInt(parts[0] ?? "0", 10),
      mine: parts[1] === "1",
      x: Number.parseInt(parts[2] ?? "0", 10),
      y: Number.parseInt(parts[3] ?? "0", 10),
      type: (parts[4] ?? "NEUTRAL") as PacType,
      abilityDuration: Number.parseInt(parts[5] ?? "0", 10),
      abilityCooldown: Number.parseInt(parts[6] ?? "0", 10),
    })
  }

  const pelletCount = Number.parseInt(next(), 10)
  const pellets: Pellet[] = []
  for (let p = 0; p < pelletCount; p++) {
    const parts = next().split(/\s+/)
    pellets.push({
      x: Number.parseInt(parts[0] ?? "0", 10),
      y: Number.parseInt(parts[1] ?? "0", 10),
      value: Number.parseInt(parts[2] ?? "1", 10),
    })
  }

  return { myScore, oppScore, pacs, pellets }
}

/**
 * Parse the plain-text response from `/api/serialize`: global header (size +
 * map rows) followed by first-frame info.
 *
 * Format (see games/spring2020/engine/serializer.go):
 *
 *   <width> <height>
 *   <row0> ... <rowH-1>      (chars: ' ' floor, '#' wall)
 *   --- frame data ---
 *   <myScore> <oppScore>
 *   <pacCount>
 *   ...
 *   <pelletCount>
 *   ...
 */
export function parseSerializeResponse(text: string): MapData {
  // Keep all lines (floor rows contain spaces but aren't empty — border walls
  // guarantee non-empty rows — however split("\n") may leave a trailing "").
  const raw = text.split("\n")
  // Drop only the final empty line if present (trailing newline).
  if (raw.length > 0 && raw[raw.length - 1] === "") raw.pop()

  const header = raw[0]
  if (!header) throw new Error("empty serialize response")
  const [wStr, hStr] = header.split(/\s+/)
  const width = Number.parseInt(wStr ?? "0", 10)
  const height = Number.parseInt(hStr ?? "0", 10)

  const walls: boolean[] = new Array(width * height)
  for (let y = 0; y < height; y++) {
    const row = raw[1 + y] ?? ""
    for (let x = 0; x < width; x++) {
      walls[y * width + x] = row[x] === "#"
    }
  }

  const frame = parseFrameLines(raw.slice(1 + height))

  return { width, height, walls, ...frame }
}
