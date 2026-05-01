import { type FrameContext, type GameViewerAdapter, winnerFromRanks } from "@shared/viewer/types.ts"
import { createGameViewerViews } from "@shared/viewer/views.tsx"
import { AppleIcon, ArrowDownIcon, BrickWallIcon, RotateCcwIcon, SkullIcon, SwordsIcon } from "lucide-react"
import { type BirdCoordMeta, type BirdMeta, type FrameData, lerpFrame, type MapData, parseFrameLines, parseSerializeResponse, type TraceTurn } from "./parser.ts"
import { destroyRenderer, initRenderer, updateFrame } from "./renderer.ts"

type TraceKind = "EAT" | "HIT_SELF" | "HIT_WALL" | "HIT_ENEMY" | "DEAD" | "DEAD_FALL"

interface MoveTrace {
  kind: TraceKind
  coord?: string
}

interface MoveRow {
  birdId: number
  mine: boolean
  direction: string
  size?: number
  traces: MoveTrace[]
}

const TRACE_ORDER: Record<TraceKind, number> = {
  EAT: 0,
  HIT_SELF: 1,
  HIT_WALL: 2,
  HIT_ENEMY: 3,
  DEAD_FALL: 4,
  DEAD: 5,
}

function TraceBadge({ trace }: { trace: MoveTrace }) {
  const icon = (() => {
    switch (trace.kind) {
      case "EAT":
        return <AppleIcon className="size-3.5 text-red-500" />
      case "HIT_SELF":
        return <RotateCcwIcon className="size-3.5 text-purple-500" />
      case "HIT_WALL":
        return <BrickWallIcon className="size-3.5 text-amber-600" />
      case "HIT_ENEMY":
        return <SwordsIcon className="size-3.5 text-orange-500" />
      case "DEAD":
        return <SkullIcon className="size-3.5 text-neutral-600" />
      case "DEAD_FALL":
        return <ArrowDownIcon className="size-3.5 text-sky-500" />
    }
  })()
  return (
    <span className="inline-flex items-center gap-0.5" title={trace.kind}>
      {icon}
      {trace.coord && <span>{trace.coord}</span>}
    </span>
  )
}

function parseMoves(turn: TraceTurn, myIds: Set<number>, frame: FrameData | undefined): MoveRow[] {
  const directions = new Map<number, string>()
  for (const output of turn.output ?? []) {
    if (!output) continue
    for (const command of output.split(";")) {
      const parts = command.trim().split(/\s+/)
      if (parts.length >= 2) {
        const id = Number(parts[0])
        if (!Number.isNaN(id) && parts[1]) {
          directions.set(id, parts[1])
        }
      }
    }
  }

  const tracesByBird = new Map<number, MoveTrace[]>()
  for (const turnTrace of turn.traces ?? []) {
    const kind = turnTrace.type as TraceKind
    if (!(kind in TRACE_ORDER)) continue
    const meta = turnTrace.meta as BirdCoordMeta | BirdMeta | undefined
    if (!meta) continue
    const bid = meta.bird
    if (typeof bid !== "number") continue
    const coord = "coord" in meta && Array.isArray(meta.coord) ? `${meta.coord[0]},${meta.coord[1]}` : undefined
    const list = tracesByBird.get(bid) ?? []
    list.push({ kind, coord })
    tracesByBird.set(bid, list)
  }
  for (const list of tracesByBird.values()) {
    list.sort((a, b) => TRACE_ORDER[a.kind] - TRACE_ORDER[b.kind])
  }

  const sizes = new Map<number, number>()
  if (frame) {
    for (const bird of frame.birds) sizes.set(bird.id, bird.body.length)
  }

  const birdIds = new Set<number>([...directions.keys(), ...tracesByBird.keys()])
  return [...birdIds]
    .sort((a, b) => a - b)
    .map((birdId) => ({
      birdId,
      mine: myIds.has(birdId),
      direction: directions.get(birdId) ?? "",
      size: sizes.get(birdId),
      traces: tracesByBird.get(birdId) ?? [],
    }))
}

function scoreFromFrame(frame: FrameData, mapData: MapData): [number, number] {
  const myIdSet = new Set(mapData.myBirdIds)
  let mine = 0
  let opponent = 0
  for (const bird of frame.birds) {
    const segments = bird.body.length
    if (myIdSet.has(bird.id)) mine += segments
    else opponent += segments
  }
  return [mine, opponent]
}

function renderRows(context: FrameContext<MapData, FrameData, TraceTurn>) {
  if (!context.turn) return null
  const rows = parseMoves(context.turn, new Set(context.mapData.myBirdIds), context.frame)
  if (rows.length === 0) return null
  return (
    <div className="flex flex-col gap-1">
      {rows.map((row) => (
        <div key={row.birdId} className="flex items-center gap-1.5">
          <span className={row.mine ? "w-6 shrink-0 text-sky-400" : "w-6 shrink-0 text-red-400"}>S{row.birdId}</span>
          <span className="w-7 shrink-0 tabular-nums text-foreground/80">{row.size !== undefined ? `[${row.size}]` : ""}</span>
          <span className="w-4 shrink-0">{row.mine ? "->" : "<-"}</span>
          <span className="w-14 shrink-0">{row.direction}</span>
          {row.traces.map((trace, index) => (
            <TraceBadge key={`${trace.kind}-${index}`} trace={trace} />
          ))}
        </div>
      ))}
    </div>
  )
}

export const winter2026Adapter: GameViewerAdapter<MapData, FrameData, TraceTurn> = {
  game: "winter2026",
  title: "Winter 2026",
  defaultLeague: "4",
  leagueOptions: [
    { value: "1", label: "Bronze" },
    { value: "2", label: "Silver" },
    { value: "3", label: "Gold" },
    { value: "4", label: "Legend" },
  ],
  parseSerializeResponse,
  buildTimeline: (_mapData, trace) => {
    const frames: FrameData[] = []
    const turns: (TraceTurn | null)[] = []
    const initialTurn = trace.turns[0]
    const initialInput = initialTurn?.game_input
    if (initialTurn && initialInput) {
      frames.push(parseFrameLines(initialInput))
      turns.push(null)
      for (let i = 1; i <= trace.turns.length; i++) {
        const source = trace.turns[i]?.game_input ?? trace.turns[i - 1]?.game_input ?? initialInput
        frames.push(parseFrameLines(source))
        turns.push(trace.turns[i - 1] ?? null)
      }
    }
    return { frames, turns }
  },
  initRenderer: async (container, mapData) => {
    await initRenderer(container, mapData)
    return undefined
  },
  updateFrame: (frame, context) => updateFrame(frame, context.mapData.myBirdIds, context.phase === "interpolate" ? { skipApples: true } : undefined),
  destroyRenderer,
  lerpFrame,
  getScore: scoreFromFrame,
  formatTurnLabel: (context) => `turn ${context.frameIndex} / ${Math.max(0, context.frameCount - 1)}`,
  renderTurnLog: renderRows,
  turnLogEmptyLabel: () => "initial state",
  formatRunStatus: ({ actualSeed, mapData, run }) => {
    const winner = run.winner === -1 ? "draw" : run.winner === 0 ? "blue" : "red"
    const ttfo = run.ttfo_ms ?? [0, 0]
    const aot = run.aot_ms ?? [0, 0]
    return (
      <>
        seed={actualSeed} {mapData.width}x{mapData.height} winner={winner} score=<span className="text-sky-400">{run.score_blue}</span>:<span className="text-red-400">{run.score_red}</span> turns={run.turns} blue ttfo={ttfo[0].toFixed(0)}ms aot=
        {aot[0].toFixed(0)}ms red ttfo={ttfo[1].toFixed(0)}ms aot={aot[1].toFixed(0)}ms
      </>
    )
  },
  formatReplayStatus: ({ replayId, trace }) => {
    const left = trace.players[0] ?? "left"
    const right = trace.players[1] ?? "right"
    const winner = winnerFromRanks(trace.ranks)
    const winnerName = winner === 0 ? left : winner === 1 ? right : "draw"
    const winnerClass = winner === 0 ? "text-sky-400" : winner === 1 ? "text-red-400" : "text-muted-foreground"
    return (
      <>
        replay: {replayId} seed={trace.seed} <span className="text-sky-400">{left}</span> vs <span className="text-red-400">{right}</span> winner=<span className={winnerClass}>{winnerName}</span> score=
        <span className="text-sky-400">{trace.scores[0]}</span>:<span className="text-red-400">{trace.scores[1]}</span> turns={trace.turns.length}
      </>
    )
  },
  formatBatchMatchStatus: ({ match }) => {
    const winnerLabel = match.winner === -1 ? "draw" : match.winner === 0 ? match.left_bot : match.right_bot
    return `match #${match.id} seed=${match.seed} ${match.left_bot} vs ${match.right_bot} winner=${winnerLabel} score=${match.score_left}:${match.score_right} turns=${match.turns} left ttfo=${match.ttfo_left_ms.toFixed(0)}ms aot=${match.aot_left_ms.toFixed(0)}ms right ttfo=${match.ttfo_right_ms.toFixed(0)}ms aot=${match.aot_right_ms.toFixed(0)}ms`
  },
  playbackDurationMs: 1000,
  minRenderIntervalMs: 50,
}

export const winter2026Views = createGameViewerViews(winter2026Adapter)
export const PlayView = winter2026Views.SingleView
export const MassView = winter2026Views.BatchView
export const BatchMatchView = winter2026Views.BatchMatchView
export const ReplaysView = winter2026Views.ReplaysView
export const ReplayView = winter2026Views.ReplayView
export const ReplayViewer = winter2026Views.ReplayViewer
export const batchMatchCache = winter2026Views.batchMatchCache
