import { type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from "react"
import { ArenaWorkspaceLayout, PlaybackControls, ReplayStage, TurnLogPanel } from "./components.tsx"
import type { FrameContext, GameViewerAdapter, TraceMatchBase, TraceTurnBase } from "./types.ts"

interface ReplayViewerProps<TMapData, TFrame, TTurn extends TraceTurnBase, TMeta> {
  adapter: GameViewerAdapter<TMapData, TFrame, TTurn, TMeta>
  mapData: TMapData
  trace: TraceMatchBase<TTurn>
  status?: ReactNode
  leftSlot?: ReactNode
  topRightSlot?: ReactNode
}

export function ReplayViewer<TMapData, TFrame, TTurn extends TraceTurnBase, TMeta = unknown>({ adapter, mapData, trace, status, leftSlot, topRightSlot }: ReplayViewerProps<TMapData, TFrame, TTurn, TMeta>) {
  const containerRef = useRef<HTMLDivElement>(null)
  const framesRef = useRef<TFrame[]>([])
  const turnsRef = useRef<(TTurn | null)[]>([])
  const metaRef = useRef<(TMeta | undefined)[]>([])
  const turnRef = useRef(0)

  const [turnLabel, setTurnLabel] = useState<ReactNode>("")
  const [turnLog, setTurnLog] = useState<ReactNode>(null)
  const [turnMarker, setTurnMarker] = useState<ReactNode>(null)
  const [emptyLabel, setEmptyLabel] = useState<ReactNode>("initial state")
  const [canvasWidth, setCanvasWidth] = useState<number | undefined>(undefined)
  const [sliderMax, setSliderMax] = useState(0)
  const [sliderValue, setSliderValue] = useState(0)
  const [ready, setReady] = useState(false)
  const [currentScores, setCurrentScores] = useState<[number, number]>([0, 0])

  const contextFor = useCallback(
    (frameIndex: number): FrameContext<TMapData, TFrame, TTurn, TMeta> | null => {
      const frame = framesRef.current[frameIndex]
      if (!frame) return null
      return {
        mapData,
        trace,
        frame,
        turn: turnsRef.current[frameIndex] ?? null,
        frameIndex,
        frameCount: framesRef.current.length,
        meta: metaRef.current[frameIndex],
      }
    },
    [mapData, trace],
  )

  const commitFrameState = useCallback(
    (frameIndex: number) => {
      const context = contextFor(frameIndex)
      if (!context) return
      adapter.updateFrame(context.frame, { mapData })
      setCurrentScores(adapter.getScore(context.frame, mapData))
      setTurnLabel(adapter.formatTurnLabel(context))
      setTurnLog(adapter.renderTurnLog(context))
      setTurnMarker(adapter.turnMarker?.(context) ?? null)
      setEmptyLabel(adapter.turnLogEmptyLabel(context))
    },
    [adapter, contextFor, mapData],
  )

  const goToTurn = useCallback(
    (target: number) => {
      const frames = framesRef.current
      if (frames.length === 0) return
      const clamped = Math.max(0, Math.min(target, frames.length - 1))
      turnRef.current = clamped
      setSliderValue(clamped)
      commitFrameState(clamped)
    },
    [commitFrameState],
  )

  useEffect(() => {
    const container = containerRef.current
    if (!container) return
    let cancelled = false

    setReady(false)

    const timeline = adapter.buildTimeline(mapData, trace)
    framesRef.current = timeline.frames
    turnsRef.current = timeline.turns
    metaRef.current = timeline.meta ?? []

    const run = async () => {
      const dimensions = await adapter.initRenderer(container, mapData)
      if (cancelled) return
      setCanvasWidth(dimensions?.width)
      turnRef.current = 0
      setSliderMax(Math.max(0, timeline.frames.length - 1))
      setSliderValue(0)
      commitFrameState(0)
      setReady(true)
    }
    run()

    return () => {
      cancelled = true
      adapter.destroyRenderer()
    }
  }, [adapter, commitFrameState, mapData, trace])

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (framesRef.current.length === 0) return
      if (event.key === "ArrowLeft") {
        event.preventDefault()
        goToTurn(turnRef.current - 1)
      } else if (event.key === "ArrowRight") {
        event.preventDefault()
        goToTurn(turnRef.current + 1)
      }
    }
    document.addEventListener("keydown", handler)
    return () => document.removeEventListener("keydown", handler)
  }, [goToTurn])

  const controls = useMemo(
    () => (
      <PlaybackControls
        sliderMax={sliderMax}
        sliderValue={sliderValue}
        turnLabel={turnLabel}
        onFirst={() => goToTurn(0)}
        onPrevious={() => goToTurn(turnRef.current - 1)}
        onNext={() => goToTurn(turnRef.current + 1)}
        onLast={() => goToTurn(framesRef.current.length - 1)}
        onSliderChange={goToTurn}
      />
    ),
    [goToTurn, sliderMax, sliderValue, turnLabel],
  )

  return (
    <ArenaWorkspaceLayout
      left={
        <>
          {leftSlot}
          {ready && (
            <TurnLogPanel turnLabel={turnLabel} score={currentScores} marker={turnMarker} emptyLabel={emptyLabel}>
              {turnLog}
            </TurnLogPanel>
          )}
        </>
      }
    >
      <ReplayStage status={status} canvasRef={containerRef} ready={ready} controls={controls} maxWidth={canvasWidth} topRight={topRightSlot} />
    </ArenaWorkspaceLayout>
  )
}
