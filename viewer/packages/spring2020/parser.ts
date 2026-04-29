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

/**
 * Produce an in-between FrameData for smooth playback. Pacs slide from their
 * previous coord to their new coord; pellets stay frozen on the `from` frame
 * so an eaten pellet remains visible under the pac until the turn commits.
 * Pacs present in `from` but absent in `to` (hidden by fog or dead) keep
 * their last coord for the whole transition.
 */
export function lerpFrame(from: FrameData, to: FrameData, t: number): FrameData {
  if (t <= 0) return from
  if (t >= 1) return to

  const fromMap = new Map(from.pacs.map((p) => [pacKey(p), p]))
  const toMap = new Map(to.pacs.map((p) => [pacKey(p), p]))

  const pacs: Pac[] = []
  for (const toPac of to.pacs) {
    const fromPac = fromMap.get(pacKey(toPac))
    if (!fromPac) {
      pacs.push(toPac)
      continue
    }
    pacs.push({
      ...toPac,
      x: fromPac.x + (toPac.x - fromPac.x) * t,
      y: fromPac.y + (toPac.y - fromPac.y) * t,
    })
  }
  for (const fromPac of from.pacs) {
    if (!toMap.has(pacKey(fromPac))) {
      pacs.push(fromPac)
    }
  }

  return { myScore: from.myScore, oppScore: from.oppScore, pacs, pellets: from.pellets }
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
  game_input: { p0?: string[]; p1?: string[] }
  p0_output: string
  p1_output: string
  timing?: TraceTurnTiming
  traces?: TurnTrace[]
}

export interface TraceTurnTiming {
  response: [number, number]
}

export interface TurnTrace {
  label: string
  payload: string
}

function pacKey(pac: Pac): string {
  return `${pac.mine ? 0 : 1}:${pac.id}`
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
 * Merge frames from p0 and p1 views. Each player sees their own pacs plus any
 * opponent pacs within line-of-sight; union gives the most complete picture.
 * Pellets are unioned by coord. Scores: p0 view's "my/opp" = p0/p1.
 */
export function mergeFrames(p0: FrameData, p1: FrameData): FrameData {
  const byId = new Map<string, Pac>()
  for (const pac of p0.pacs) {
    // p0 view: mine=true -> belongs to p0 (side 0)
    const mergedPac = { ...pac, mine: pac.mine }
    byId.set(pacKey(mergedPac), mergedPac)
  }
  for (const pac of p1.pacs) {
    // p1 view: mine=true -> belongs to p1 (side 1); invert to keep "mine==side 0"
    const mergedPac = { ...pac, mine: !pac.mine }
    if (!byId.has(pacKey(mergedPac))) {
      byId.set(pacKey(mergedPac), mergedPac)
    }
  }

  const pelletKey = (x: number, y: number) => `${x},${y}`
  const pelletMap = new Map<string, Pellet>()
  for (const p of p0.pellets) pelletMap.set(pelletKey(p.x, p.y), p)
  for (const p of p1.pellets) {
    const k = pelletKey(p.x, p.y)
    if (!pelletMap.has(k)) pelletMap.set(k, p)
  }

  return {
    myScore: p0.myScore,
    oppScore: p0.oppScore,
    pacs: [...byId.values()].sort((a, b) => Number(b.mine) - Number(a.mine) || a.id - b.id),
    pellets: [...pelletMap.values()],
  }
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
