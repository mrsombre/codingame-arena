import type { ReactNode } from "react"

export interface LeagueOption {
  value: string
  label: string
}

export interface TraceTiming {
  first_response: [number, number]
  response_average: [number, number]
  response_median: [number, number]
}

export interface TraceTurnTiming {
  response: [number, number]
}

export interface TurnTrace<M = unknown> {
  type: string
  data?: M
}

export interface TraceTurnBase {
  turn: number
  game_input?: string[]
  output?: [string, string]
  timing?: TraceTurnTiming
  /**
   * Per-player trace events for this turn. Index 0 is everything player 0
   * owned; index 1 player 1. Cross-owner events are mirrored into both slots.
   */
  traces?: [TurnTrace[], TurnTrace[]]
}

export interface TraceMatchBase<TTurn extends TraceTurnBase = TraceTurnBase> {
  match_id: number
  seed: string
  ranks: [number, number]
  scores: [number, number]
  players: [string, string]
  timing?: TraceTiming
  turns: TTurn[]
}

export interface ReplayEntry {
  id: string
  size: number
  mtime: string
  left_name?: string
  right_name?: string
  league?: number
  score_left: number
  score_right: number
  winner: number
}

export interface ReplayResponse<TTurn extends TraceTurnBase = TraceTurnBase> extends TraceMatchBase<TTurn> {
  league: number
}

export interface BatchMatch {
  id: number
  seed: string
  winner: number
  score_left: number
  score_right: number
  turns: number
  ttfo_left_ms: number
  ttfo_right_ms: number
  aot_left_ms: number
  aot_right_ms: number
  left_bot: string
  right_bot: string
}

export interface BatchResponse {
  simulations: number
  wins_blue: number
  wins_red: number
  draws: number
  avg_score_blue: number
  avg_score_red: number
  avg_turns: number
  avg_ttfo_blue_ms: number
  avg_ttfo_red_ms: number
  avg_aot_blue_ms: number
  avg_aot_red_ms: number
  seed: string
  blue_bot: string
  red_bot: string
  matches: BatchMatch[]
}

export interface RunResponse {
  id: number
  seed: string
  winner: number
  score_blue: number
  score_red: number
  turns: number
  loss_reason_blue?: string
  loss_reason_red?: string
  ttfo_ms?: [number, number]
  aot_ms?: [number, number]
  swapped?: boolean
  trace: TraceMatchBase
}

export interface FrameTimeline<TFrame, TTurn extends TraceTurnBase, TMeta = unknown> {
  frames: TFrame[]
  turns: (TTurn | null)[]
  meta?: TMeta[]
}

export interface FrameContext<TMapData, TFrame, TTurn extends TraceTurnBase, TMeta = unknown> {
  mapData: TMapData
  trace: TraceMatchBase<TTurn>
  frame: TFrame
  turn: TTurn | null
  frameIndex: number
  frameCount: number
  meta: TMeta | undefined
}

export interface GameViewerAdapter<TMapData, TFrame, TTurn extends TraceTurnBase = TraceTurnBase, TMeta = unknown> {
  game: string
  title: string
  leagueOptions: LeagueOption[]
  defaultLeague?: string
  parseSerializeResponse: (source: string) => TMapData
  buildTimeline: (mapData: TMapData, trace: TraceMatchBase<TTurn>) => FrameTimeline<TFrame, TTurn, TMeta>
  initRenderer: (container: HTMLElement, mapData: TMapData) => Promise<{ width?: number; height?: number } | undefined>
  updateFrame: (frame: TFrame, context: { mapData: TMapData }) => void
  destroyRenderer: () => void
  getScore: (frame: TFrame, mapData: TMapData) => [number, number]
  formatTurnLabel: (context: FrameContext<TMapData, TFrame, TTurn, TMeta>) => ReactNode
  renderTurnLog: (context: FrameContext<TMapData, TFrame, TTurn, TMeta>) => ReactNode
  turnLogEmptyLabel: (context: FrameContext<TMapData, TFrame, TTurn, TMeta>) => ReactNode
  turnMarker?: (context: FrameContext<TMapData, TFrame, TTurn, TMeta>) => ReactNode
  formatRunStatus: (context: { actualSeed: string; mapData: TMapData; trace: TraceMatchBase<TTurn>; run: RunResponse }) => ReactNode
  formatReplayStatus: (context: { replayId: string; mapData: TMapData; trace: TraceMatchBase<TTurn>; replay: ReplayResponse<TTurn> }) => ReactNode
  formatBatchMatchStatus: (context: { match: BatchMatch; mapData: TMapData; trace: TraceMatchBase<TTurn> }) => ReactNode
}

export function winnerFromRanks(ranks: [number, number]): number {
  if (ranks[0] === 0 && ranks[1] === 0) return -1
  return ranks[0] === 0 ? 0 : 1
}
