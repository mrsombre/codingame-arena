/** Parsed output from /api/serialize (global info + first frame). */
export interface MapData {
  myId: number
  width: number
  height: number
  /** Row-major grid: true = wall, false = empty. */
  walls: boolean[]
  birdsPerPlayer: number
  myBirdIds: number[]
  oppBirdIds: number[]
  apples: Coord[]
  birds: Bird[]
}

export interface Coord {
  x: number
  y: number
}

export interface Bird {
  id: number
  body: Coord[]
}

/** Dynamic per-turn state (apples + birds). */
export interface FrameData {
  apples: Coord[]
  birds: Bird[]
}

/**
 * Produce an in-between FrameData for smooth playback.
 *
 * Body segments are interpolated index by index — each `to` segment slides
 * from its counterpart in `from`; new segments that extended past the old
 * body anchor at the old tail so they stay put until the snake catches up.
 * Birds present in `from` but missing from `to` (died this turn) keep their
 * last body for the whole transition — the commit step will remove them.
 * Apples are held at the `from` state so an eaten tile stays visible under
 * the snake's head until the turn commits.
 */
export function lerpFrame(from: FrameData, to: FrameData, t: number): FrameData {
  if (t <= 0) return from
  if (t >= 1) return to

  const fromMap = new Map(from.birds.map((b) => [b.id, b]))
  const toMap = new Map(to.birds.map((b) => [b.id, b]))

  const birds: Bird[] = []
  for (const toBird of to.birds) {
    const fromBird = fromMap.get(toBird.id)
    if (!fromBird) {
      birds.push(toBird)
      continue
    }
    const fromBody = fromBird.body
    const tailFallback = fromBody[fromBody.length - 1]
    const body: Coord[] = new Array(toBird.body.length)
    for (let i = 0; i < toBird.body.length; i++) {
      const toSeg = toBird.body[i] as Coord
      const fromSeg = fromBody[i] ?? tailFallback ?? toSeg
      body[i] = {
        x: fromSeg.x + (toSeg.x - fromSeg.x) * t,
        y: fromSeg.y + (toSeg.y - fromSeg.y) * t,
      }
    }
    birds.push({ id: toBird.id, body })
  }
  for (const fromBird of from.birds) {
    if (!toMap.has(fromBird.id)) {
      birds.push(fromBird)
    }
  }

  return { apples: from.apples, birds }
}

/** Trace JSON from GET /api/matches/{id}. */
export interface TraceMatch {
  match_id: number
  seed: string
  /** CodinGame-style ranks: 0 = first place, [0,0] = draw. */
  ranks: [number, number]
  scores: [number, number]
  /** Player (bot) basenames per in-match side: [left, right]. */
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
   * TraceMatch.blue). Absent on turns where blue did not execute. */
  game_input?: string[]
  /** Raw stdout per side: [left, right]. Empty entry means the side did
   * not execute this turn. Absent when both sides were silent. */
  output?: [string, string]
  timing?: TraceTurnTiming
  traces?: TurnTrace[]
}

export interface TraceTurnTiming {
  response: [number, number]
}

export interface TurnTrace<M = unknown> {
  type: string
  meta?: M
}

export interface BirdMeta {
  bird: number
}

export interface BirdCoordMeta {
  bird: number
  coord: [number, number]
}

/**
 * Parse frame lines (apple positions + bird bodies) from a string array.
 * Used both by parseSerializeResponse and for per-turn trace game_input.
 *
 * Format:
 *   <appleCount>
 *   <x> <y>  (per apple)
 *   <birdCount>
 *   <id> <x0,y0>:<x1,y1>:...  (per bird, head first)
 */
export function parseFrameLines(lines: string[]): FrameData {
  let i = 0
  const next = (): string => {
    const line = lines[i]
    if (i >= lines.length || line === undefined) {
      throw new Error(`unexpected end of frame input at line ${i}`)
    }
    i++
    return line
  }

  const appleCount = Number.parseInt(next(), 10)
  const apples: Coord[] = []
  for (let a = 0; a < appleCount; a++) {
    const parts = next().split(" ")
    apples.push({
      x: Number.parseInt(parts[0] ?? "0", 10),
      y: Number.parseInt(parts[1] ?? "0", 10),
    })
  }

  const birdCount = Number.parseInt(next(), 10)
  const birds: Bird[] = []
  for (let b = 0; b < birdCount; b++) {
    const line = next()
    const spaceIdx = line.indexOf(" ")
    const id = Number.parseInt(line.slice(0, spaceIdx), 10)
    const segments = line.slice(spaceIdx + 1).split(":")
    const body: Coord[] = segments.map((s) => {
      const parts = s.split(",")
      return {
        x: Number.parseInt(parts[0] ?? "0", 10),
        y: Number.parseInt(parts[1] ?? "0", 10),
      }
    })
    birds.push({ id, body })
  }

  return { apples, birds }
}

/**
 * Parse the plain-text response from `/api/serialize` which concatenates
 * global info and first-frame info.
 *
 * Format (see games/winter2026/engine/serializer.go):
 *
 *   <myId>
 *   <width>
 *   <height>
 *   <row0> ... <rowH-1>      (chars: '.' empty, '#' wall)
 *   <birdsPerPlayer>
 *   <myBird0Id> ... <myBirdNId>
 *   <oppBird0Id> ... <oppBirdNId>
 *   --- frame data ---
 *   <appleCount>
 *   <x> <y>  (per apple)
 *   <birdCount>
 *   <id> <x0,y0>:<x1,y1>:...  (per bird, head first)
 */
export function parseSerializeResponse(text: string): MapData {
  const lines = text.split("\n").filter((l) => l !== "")
  let i = 0
  const next = (): string => {
    const line = lines[i]
    if (i >= lines.length || line === undefined) {
      throw new Error(`unexpected end of input at line ${i}`)
    }
    i++
    return line
  }

  const myId = Number.parseInt(next(), 10)
  const width = Number.parseInt(next(), 10)
  const height = Number.parseInt(next(), 10)

  const walls: boolean[] = new Array(width * height)
  for (let y = 0; y < height; y++) {
    const row = next()
    for (let x = 0; x < width; x++) {
      walls[y * width + x] = row[x] === "#"
    }
  }

  const birdsPerPlayer = Number.parseInt(next(), 10)
  const myBirdIds: number[] = []
  for (let b = 0; b < birdsPerPlayer; b++) {
    myBirdIds.push(Number.parseInt(next(), 10))
  }
  const oppBirdIds: number[] = []
  for (let b = 0; b < birdsPerPlayer; b++) {
    oppBirdIds.push(Number.parseInt(next(), 10))
  }

  // Delegate frame parsing to parseFrameLines
  const frame = parseFrameLines(lines.slice(i))

  return { myId, width, height, walls, birdsPerPlayer, myBirdIds, oppBirdIds, ...frame }
}
