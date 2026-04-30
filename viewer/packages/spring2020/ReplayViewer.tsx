import { Button } from "@shared/components/ui/button.tsx"
import { Card, CardContent, CardHeader, CardTitle } from "@shared/components/ui/card.tsx"
import { Slider } from "@shared/components/ui/slider.tsx"
import { ChevronLeftIcon, ChevronRightIcon, ChevronsLeftIcon, ChevronsRightIcon, PauseIcon, PlayIcon, ZapIcon } from "lucide-react"
import { type ReactNode, useCallback, useEffect, useRef, useState } from "react"
import { type FrameData, lerpFrame, type MapData, parseFrameLines, type TraceMatch, type TraceTurn } from "./parser.ts"
import { destroyRenderer, initRenderer, updateFrame } from "./renderer.ts"

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
    for (const cmd of output.split("|")) {
      const trimmed = cmd.trim()
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
  const lines = turn.game_input
  if (lines) {
    return parseFrameLines(lines)
  }
  if (fallback) {
    return fallback
  }
  throw new Error(`turn ${turn.turn} has no frame input`)
}

interface ReplayViewerProps {
  mapData: MapData
  trace: TraceMatch
  status?: ReactNode
  leftSlot?: ReactNode
}

export function ReplayViewer({ mapData, trace, status, leftSlot }: ReplayViewerProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const framesRef = useRef<FrameData[]>([])
  const turnsRef = useRef<(TraceTurn | null)[]>([])
  const mainTurnIndexRef = useRef<number[]>([])
  const totalMainTurnsRef = useRef(0)
  const turnRef = useRef(0)

  const [turnMoves, setTurnMoves] = useState<MoveRow[]>([])
  const [isSpeedTurn, setIsSpeedTurn] = useState(false)
  const [turnDisplay, setTurnDisplay] = useState("")
  const [canvasWidth, setCanvasWidth] = useState<number | undefined>(undefined)
  const [sliderMax, setSliderMax] = useState(0)
  const [sliderValue, setSliderValue] = useState(0)
  const [ready, setReady] = useState(false)
  const [currentScores, setCurrentScores] = useState<[number, number]>([0, 0])
  const [playing, setPlaying] = useState(false)
  const [pauseRequested, setPauseRequested] = useState(false)
  const pauseRequestedRef = useRef(false)
  useEffect(() => {
    pauseRequestedRef.current = pauseRequested
  }, [pauseRequested])

  const goToTurn = useCallback((t: number) => {
    const frames = framesRef.current
    if (frames.length === 0) return
    const clamped = Math.max(0, Math.min(t, frames.length - 1))
    turnRef.current = clamped
    const frame = frames[clamped]
    if (frame) {
      updateFrame(frame)
      setCurrentScores([frame.myScore, frame.oppScore])
    }
    setSliderValue(clamped)

    const turn = turnsRef.current[clamped]
    const speed = !!turn && !turn.game_input
    const mainIdx = mainTurnIndexRef.current[clamped] ?? 0
    setTurnDisplay(`turn ${mainIdx} / ${totalMainTurnsRef.current}`)
    if (turn) {
      setTurnMoves(parseMoves(turn))
      setIsSpeedTurn(speed)
    } else {
      setTurnMoves([])
      setIsSpeedTurn(false)
    }
  }, [])

  const togglePlay = useCallback(() => {
    if (playing) {
      setPauseRequested((prev) => !prev)
      return
    }
    setPauseRequested(false)
    if (framesRef.current.length > 0 && turnRef.current >= framesRef.current.length - 1) {
      goToTurn(0)
    }
    setPlaying(true)
  }, [playing, goToTurn])

  useEffect(() => {
    if (!playing) return
    const duration = 600
    const minRenderIntervalMs = 50
    let rafId: number | null = null
    let currentTurn = turnRef.current
    let segmentStart = performance.now()
    let lastRender = 0

    const tick = (now: number) => {
      if (turnRef.current !== currentTurn) {
        currentTurn = turnRef.current
        segmentStart = now
        lastRender = 0
      }

      const frames = framesRef.current
      if (currentTurn >= frames.length - 1) {
        setPlaying(false)
        return
      }
      const fromFrame = frames[currentTurn]
      const toFrame = frames[currentTurn + 1]
      if (!fromFrame || !toFrame) {
        setPlaying(false)
        return
      }

      const elapsed = now - segmentStart
      if (elapsed >= duration) {
        goToTurn(currentTurn + 1)
        currentTurn += 1
        segmentStart = now
        lastRender = now
        if (pauseRequestedRef.current || currentTurn >= frames.length - 1) {
          setPlaying(false)
          setPauseRequested(false)
          return
        }
        rafId = requestAnimationFrame(tick)
        return
      }

      if (now - lastRender >= minRenderIntervalMs) {
        const t = elapsed / duration
        updateFrame(lerpFrame(fromFrame, toFrame, t), { skipPellets: true })
        lastRender = now
      }
      rafId = requestAnimationFrame(tick)
    }

    rafId = requestAnimationFrame(tick)
    return () => {
      if (rafId !== null) cancelAnimationFrame(rafId)
    }
  }, [playing, goToTurn])

  useEffect(() => {
    const container = containerRef.current
    if (!container) return
    let cancelled = false

    setReady(false)
    setPlaying(false)
    setPauseRequested(false)

    const N = trace.turns.length
    const frames: FrameData[] = []
    const turns: (TraceTurn | null)[] = []
    const mainTurnIdx: number[] = []

    // slider 0: initial state (from first turn's game_input, no moves yet)
    // slider i (1..N): state AFTER turn (i-1); taken from turns[i] game_input,
    // or fall back to turns[i-1] for the final turn.
    const initialTurn = trace.turns[0]
    let mainCount = 0
    if (initialTurn) {
      frames.push(frameFromTurn(initialTurn, undefined))
      turns.push(null)
      mainTurnIdx.push(0)
      for (let i = 1; i <= N; i++) {
        const previous = trace.turns[i - 1]
        if (!previous) continue
        const source = trace.turns[i] ?? previous
        frames.push(frameFromTurn(source, frames[frames.length - 1]))
        turns.push(previous)
        if (previous.game_input) mainCount++
        mainTurnIdx.push(mainCount)
      }
    }
    framesRef.current = frames
    turnsRef.current = turns
    mainTurnIndexRef.current = mainTurnIdx
    totalMainTurnsRef.current = mainCount

    const run = async () => {
      const dims = await initRenderer(container, mapData)
      if (cancelled) return
      setCanvasWidth(dims.width)
      turnRef.current = 0
      const firstFrame = frames[0]
      if (firstFrame) {
        updateFrame(firstFrame)
        setCurrentScores([firstFrame.myScore, firstFrame.oppScore])
      } else {
        setCurrentScores([0, 0])
      }
      setSliderMax(Math.max(0, frames.length - 1))
      setSliderValue(0)
      setTurnDisplay(`turn 0 / ${totalMainTurnsRef.current}`)
      setTurnMoves([])
      setIsSpeedTurn(false)
      setReady(true)
    }
    run()

    return () => {
      cancelled = true
      destroyRenderer()
    }
  }, [mapData, trace])

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (framesRef.current.length === 0) return
      if (e.key === "ArrowLeft") {
        e.preventDefault()
        goToTurn(turnRef.current - 1)
      } else if (e.key === "ArrowRight") {
        e.preventDefault()
        goToTurn(turnRef.current + 1)
      }
    }
    document.addEventListener("keydown", handler)
    return () => document.removeEventListener("keydown", handler)
  }, [goToTurn])

  return (
    <div className="flex gap-8">
      <div className="flex w-80 shrink-0 flex-col gap-4 overflow-hidden">
        {leftSlot}
        {ready && (
          <Card size="sm">
            <CardHeader>
              <CardTitle className="flex items-center justify-between text-xs">
                <span className="flex items-center gap-1">
                  {turnDisplay}
                  {isSpeedTurn && <ZapIcon className="size-3 text-yellow-400" />}
                </span>
                <span className="font-mono">
                  <span className="text-sky-400">{currentScores[0]}</span>
                  <span className="mx-0.5 text-muted-foreground">:</span>
                  <span className="text-red-400">{currentScores[1]}</span>
                </span>
              </CardTitle>
            </CardHeader>
            <CardContent>
              {turnMoves.length > 0 ? (
                <div className="flex flex-col gap-1 font-mono text-xs text-muted-foreground">
                  {turnMoves.map((row, i) => (
                    <div key={`${row.pacId}-${row.mine}-${i}`} className="flex items-center gap-1.5">
                      <span className={`w-6 shrink-0 ${row.mine ? "text-sky-400" : "text-red-400"}`}>P{row.pacId}</span>
                      <span className="w-3 shrink-0">{row.mine ? "\u2192" : "\u2190"}</span>
                      <span className="truncate">{row.command}</span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="font-mono text-xs text-muted-foreground">{isSpeedTurn ? "speed sub-turn" : "initial state"}</p>
              )}
            </CardContent>
          </Card>
        )}
      </div>

      <div className="min-w-0 flex-1 overflow-hidden">
        <div className="mb-3 flex min-h-9 items-center justify-between gap-3" style={canvasWidth ? { maxWidth: canvasWidth } : undefined}>
          {status && <p className="min-w-0 truncate font-mono text-xs text-muted-foreground">{status}</p>}
        </div>
        <div ref={containerRef} className="[&_canvas]:block [&_canvas]:rounded-sm" />
        {ready && (
          <div className="mt-3 flex items-center gap-3">
            <Button variant="default" size="icon-lg" onClick={togglePlay} aria-label={playing && !pauseRequested ? "pause" : "play"}>
              {playing && !pauseRequested ? <PauseIcon className="size-5" /> : <PlayIcon className="size-5" />}
            </Button>
            <Button variant="outline" size="icon-sm" onClick={() => goToTurn(0)} aria-label="first turn">
              <ChevronsLeftIcon />
            </Button>
            <Button variant="outline" size="icon-sm" onClick={() => goToTurn(turnRef.current - 1)} aria-label="previous turn">
              <ChevronLeftIcon />
            </Button>
            <Slider
              className="flex-1 max-w-sm"
              min={0}
              max={sliderMax}
              value={[sliderValue]}
              onValueChange={([v]) => {
                if (v !== undefined) goToTurn(v)
              }}
            />
            <Button variant="outline" size="icon-sm" onClick={() => goToTurn(turnRef.current + 1)} aria-label="next turn">
              <ChevronRightIcon />
            </Button>
            <Button variant="outline" size="icon-sm" onClick={() => goToTurn(framesRef.current.length - 1)} aria-label="last turn">
              <ChevronsRightIcon />
            </Button>
            <span className="shrink-0 font-mono text-xs text-muted-foreground">{turnDisplay}</span>
          </div>
        )}
      </div>
    </div>
  )
}
