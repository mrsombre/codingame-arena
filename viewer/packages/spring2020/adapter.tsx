import { type FrameContext, type GameViewerAdapter, winnerFromRanks } from "@shared/viewer/types.ts"
import { createGameViewerViews } from "@shared/viewer/views.tsx"
import { ZapIcon } from "lucide-react"
import { type FrameData, lerpFrame, type MapData, parseFrameLines, parseSerializeResponse, type TraceTurn } from "./parser.ts"
import { destroyRenderer, initRenderer, updateFrame } from "./renderer.ts"

interface SpringTurnMeta {
  mainTurn: number
  totalMainTurns: number
  isSpeedTurn: boolean
}

interface MoveRow {
  pacId: number
  mine: boolean
  command: string
}

function parseMoves(turn: TraceTurn): MoveRow[] {
  const rows: MoveRow[] = []
  const outputs = turn.output ?? ["", ""]
  for (const [output, mine] of [
    [outputs[0], true],
    [outputs[1], false],
  ] as const) {
    if (!output) continue
    for (const command of output.split("|")) {
      const trimmed = command.trim()
      if (!trimmed) continue
      const parts = trimmed.split(/\s+/)
      const verb = (parts[0] ?? "").toUpperCase()
      const idNum = Number(parts[1])
      if (Number.isNaN(idNum)) continue
      rows.push({ pacId: idNum, mine, command: verb === "MOVE" ? `MOVE ${parts[2] ?? "?"},${parts[3] ?? "?"}` : parts.slice(0, 3).join(" ") })
    }
  }
  return rows.sort((a, b) => a.pacId - b.pacId)
}

function frameFromTurn(turn: TraceTurn, fallback: FrameData | undefined): FrameData {
  if (turn.game_input) return parseFrameLines(turn.game_input)
  if (fallback) return fallback
  throw new Error(`turn ${turn.turn} has no frame input`)
}

function turnContext(context: FrameContext<MapData, FrameData, TraceTurn, SpringTurnMeta>) {
  return context.meta ?? { mainTurn: context.frameIndex, totalMainTurns: context.frameCount - 1, isSpeedTurn: false }
}

function formatWinner(winner: number) {
  if (winner === -1) return "draw"
  return winner === 0 ? "blue" : "red"
}

export const spring2020Adapter: GameViewerAdapter<MapData, FrameData, TraceTurn, SpringTurnMeta> = {
  game: "spring2020",
  title: "Spring 2020",
  defaultLeague: "4",
  leagueOptions: [
    { value: "1", label: "Wood 2" },
    { value: "2", label: "Wood 1" },
    { value: "3", label: "Bronze" },
    { value: "4", label: "Silver+" },
  ],
  parseSerializeResponse,
  buildTimeline: (_mapData, trace) => {
    const totalMainTurns = trace.turns.filter((turn) => turn.game_input).length
    const frames: FrameData[] = []
    const turns: (TraceTurn | null)[] = []
    const meta: SpringTurnMeta[] = []

    const initialTurn = trace.turns[0]
    if (initialTurn) {
      frames.push(frameFromTurn(initialTurn, undefined))
      turns.push(null)
      meta.push({ mainTurn: 0, totalMainTurns, isSpeedTurn: false })
      let mainCount = 0
      for (let i = 1; i <= trace.turns.length; i++) {
        const previous = trace.turns[i - 1]
        if (!previous) continue
        const source = trace.turns[i] ?? previous
        frames.push(frameFromTurn(source, frames[frames.length - 1]))
        turns.push(previous)
        if (previous.game_input) mainCount++
        meta.push({ mainTurn: mainCount, totalMainTurns, isSpeedTurn: !previous.game_input })
      }
    }

    return { frames, turns, meta }
  },
  initRenderer,
  updateFrame: (frame, context) => updateFrame(frame, context.phase === "interpolate" ? { skipPellets: true } : undefined),
  destroyRenderer,
  lerpFrame,
  getScore: (frame) => [frame.myScore, frame.oppScore],
  formatTurnLabel: (context) => {
    const meta = turnContext(context)
    return `turn ${meta.mainTurn} / ${meta.totalMainTurns}`
  },
  renderTurnLog: (context) => {
    if (!context.turn) return null
    const rows = parseMoves(context.turn)
    if (rows.length === 0) return null
    return (
      <div className="flex flex-col gap-1">
        {rows.map((row, index) => (
          <div key={`${row.pacId}-${row.mine}-${index}`} className="flex items-center gap-1.5">
            <span className={row.mine ? "w-6 shrink-0 text-sky-400" : "w-6 shrink-0 text-red-400"}>P{row.pacId}</span>
            <span className="w-4 shrink-0">{row.mine ? "->" : "<-"}</span>
            <span className="truncate">{row.command}</span>
          </div>
        ))}
      </div>
    )
  },
  turnLogEmptyLabel: (context) => (turnContext(context).isSpeedTurn ? "speed sub-turn" : "initial state"),
  turnMarker: (context) => (turnContext(context).isSpeedTurn ? <ZapIcon className="size-3 text-yellow-400" /> : null),
  formatRunStatus: ({ actualSeed, mapData, trace, run }) => {
    const winner = formatWinner(run.winner)
    const mainTurns = trace.turns.filter((turn) => turn.game_input).length
    const ttfo = run.ttfo_ms ?? [0, 0]
    const aot = run.aot_ms ?? [0, 0]
    return (
      <>
        seed={actualSeed} {mapData.width}x{mapData.height} winner={winner} score=<span className="text-sky-400">{run.score_blue}</span>:<span className="text-red-400">{run.score_red}</span> turns={run.turns} [{mainTurns}] blue ttfo={ttfo[0].toFixed(0)}ms
        aot={aot[0].toFixed(0)}ms red ttfo={ttfo[1].toFixed(0)}ms aot={aot[1].toFixed(0)}ms
      </>
    )
  },
  formatReplayStatus: ({ replayId, trace }) => {
    const left = trace.players[0] ?? "left"
    const right = trace.players[1] ?? "right"
    const shortId = replayId.startsWith("replay-") ? replayId.slice("replay-".length) : replayId
    const winner = winnerFromRanks(trace.ranks)
    const winnerName = winner === 0 ? left : winner === 1 ? right : "draw"
    const winnerClass = winner === 0 ? "text-sky-400" : winner === 1 ? "text-red-400" : "text-muted-foreground"
    return (
      <>
        replay: {shortId} seed={trace.seed} <span className="text-sky-400">{left}</span> vs <span className="text-red-400">{right}</span> winner=<span className={winnerClass}>{winnerName}</span> score=<span className="text-sky-400">{trace.scores[0]}</span>
        :<span className="text-red-400">{trace.scores[1]}</span> turns={trace.turns.length}
      </>
    )
  },
  formatBatchMatchStatus: ({ match }) => {
    const winnerLabel = match.winner === -1 ? "draw" : match.winner === 0 ? match.left_bot : match.right_bot
    return `match #${match.id} seed=${match.seed} ${match.left_bot} vs ${match.right_bot} winner=${winnerLabel} score=${match.score_left}:${match.score_right} turns=${match.turns} left ttfo=${match.ttfo_left_ms.toFixed(0)}ms aot=${match.aot_left_ms.toFixed(0)}ms right ttfo=${match.ttfo_right_ms.toFixed(0)}ms aot=${match.aot_right_ms.toFixed(0)}ms`
  },
  playbackDurationMs: 600,
  minRenderIntervalMs: 50,
}

export const spring2020Views = createGameViewerViews(spring2020Adapter)
export const PlayView = spring2020Views.SingleView
export const MassView = spring2020Views.BatchView
export const BatchMatchView = spring2020Views.BatchMatchView
export const ReplaysView = spring2020Views.ReplaysView
export const ReplayView = spring2020Views.ReplayView
export const ReplayViewer = spring2020Views.ReplayViewer
export const batchMatchCache = spring2020Views.batchMatchCache
