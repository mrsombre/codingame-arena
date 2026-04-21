import { Button } from "@shared/components/ui/button.tsx"
import { Card, CardContent, CardHeader, CardTitle } from "@shared/components/ui/card.tsx"
import { Slider } from "@shared/components/ui/slider.tsx"
import { AppleIcon, ArrowDownIcon, BrickWallIcon, ChevronLeftIcon, ChevronRightIcon, ChevronsLeftIcon, ChevronsRightIcon, PauseIcon, PlayIcon, RotateCcwIcon, SkullIcon, SwordsIcon } from "lucide-react"
import { type ReactNode, useCallback, useEffect, useRef, useState } from "react"
import { type FrameData, lerpFrame, type MapData, parseFrameLines, type TraceMatch, type TraceTurn } from "./parser.ts"
import { destroyRenderer, initRenderer, updateFrame } from "./renderer.ts"

type EventKind = "EAT" | "HIT_WALL" | "HIT_ITSELF" | "HIT_ENEMY" | "DEAD" | "FALL"

interface MoveEvent {
  kind: EventKind
  coord?: string
}

interface MoveRow {
  birdId: number
  mine: boolean
  direction: string
  /** Alive segments at the end of this turn; undefined when the bird is no longer present. */
  size?: number
  events: MoveEvent[]
}

const EVENT_ORDER: Record<EventKind, number> = {
  EAT: 0,
  HIT_WALL: 1,
  HIT_ENEMY: 2,
  HIT_ITSELF: 3,
  FALL: 4,
  DEAD: 5,
}

function EventBadge({ event }: { event: MoveEvent }) {
  const iconClass = "size-3.5"
  const icon = (() => {
    switch (event.kind) {
      case "EAT":
        return <AppleIcon className={`${iconClass} text-red-500`} />
      case "HIT_WALL":
        return <BrickWallIcon className={`${iconClass} text-amber-600`} />
      case "HIT_ITSELF":
        return <RotateCcwIcon className={`${iconClass} text-purple-500`} />
      case "HIT_ENEMY":
        return <SwordsIcon className={`${iconClass} text-orange-500`} />
      case "DEAD":
        return <SkullIcon className={`${iconClass} text-neutral-600`} />
      case "FALL":
        return <ArrowDownIcon className={`${iconClass} text-sky-500`} />
    }
  })()
  return (
    <span className="inline-flex items-center gap-0.5" title={event.kind}>
      {icon}
      {event.coord && <span>{event.coord}</span>}
    </span>
  )
}

function parseMoves(turn: TraceTurn, myIds: Set<number>, frame: FrameData | undefined): MoveRow[] {
  const directions = new Map<number, string>()
  for (const output of [turn.p0_output, turn.p1_output]) {
    if (!output) continue
    for (const cmd of output.split(";")) {
      const parts = cmd.trim().split(/\s+/)
      if (parts.length >= 2) {
        const id = Number(parts[0])
        if (!Number.isNaN(id) && parts[1]) {
          directions.set(id, parts[1])
        }
      }
    }
  }

  const eventsByBird = new Map<number, MoveEvent[]>()
  for (const ev of turn.events ?? []) {
    const parts = ev.payload.split(" ")
    const bid = Number(parts[0])
    if (Number.isNaN(bid)) continue
    const coord = parts[1]?.includes(",") ? parts[1] : undefined
    const kind = ev.label as EventKind
    if (!(kind in EVENT_ORDER)) continue
    const list = eventsByBird.get(bid) ?? []
    list.push({ kind, coord })
    eventsByBird.set(bid, list)
  }
  for (const list of eventsByBird.values()) {
    list.sort((a, b) => EVENT_ORDER[a.kind] - EVENT_ORDER[b.kind])
  }

  const sizes = new Map<number, number>()
  if (frame) {
    for (const bird of frame.birds) sizes.set(bird.id, bird.body.length)
  }

  const birdIds = new Set<number>([...directions.keys(), ...eventsByBird.keys()])
  return [...birdIds]
    .sort((a, b) => a - b)
    .map((birdId) => ({
      birdId,
      mine: myIds.has(birdId),
      direction: directions.get(birdId) ?? "",
      size: sizes.get(birdId),
      events: eventsByBird.get(birdId) ?? [],
    }))
}

interface ReplayViewerProps {
  mapData: MapData
  trace: TraceMatch
  status?: string
  leftSlot?: ReactNode
}

export function ReplayViewer({ mapData, trace, status, leftSlot }: ReplayViewerProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const framesRef = useRef<FrameData[]>([])
  const turnsRef = useRef<(TraceTurn | null)[]>([])
  const myBirdIdsRef = useRef<number[]>(mapData.myBirdIds)
  const turnRef = useRef(0)

  const [turnMoves, setTurnMoves] = useState<MoveRow[]>([])
  const [turnDisplay, setTurnDisplay] = useState("")
  const [sliderMax, setSliderMax] = useState(0)
  const [sliderValue, setSliderValue] = useState(0)
  const [ready, setReady] = useState(false)
  const [currentScores, setCurrentScores] = useState<[number, number]>([0, 0])
  const [playing, setPlaying] = useState(false)
  // Pause-at-next-commit flag. Clicking pause while a turn is mid-animation
  // lets the snake finish sliding to the new tile before playback stops.
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
      updateFrame(frame, myBirdIdsRef.current)
      const myIdSet = new Set(myBirdIdsRef.current)
      let s0 = 0
      let s1 = 0
      for (const bird of frame.birds) {
        const segs = bird.body.length
        if (myIdSet.has(bird.id)) s0 += segs
        else s1 += segs
      }
      setCurrentScores([s0, s1])
    }
    setSliderValue(clamped)
    setTurnDisplay(`turn ${clamped} / ${frames.length - 1}`)

    const turn = turnsRef.current[clamped]
    if (turn) {
      setTurnMoves(parseMoves(turn, new Set(myBirdIdsRef.current), frame))
    } else {
      setTurnMoves([])
    }
  }, [])

  const togglePlay = useCallback(() => {
    if (playing) {
      // Defer pause to the next commit so the snake finishes its slide.
      // A second click before that commit cancels the pause request.
      setPauseRequested((prev) => !prev)
      return
    }
    // Starting play: rewind if already at the end and clear any stale request.
    setPauseRequested(false)
    if (framesRef.current.length > 0 && turnRef.current >= framesRef.current.length - 1) {
      goToTurn(0)
    }
    setPlaying(true)
  }, [playing, goToTurn])

  // Playback: while `playing`, animate each turn transition over 1 second via
  // RAF + lerpFrame, then commit via goToTurn so state (score, log, slider)
  // updates at the turn boundary. Auto-stops at the last frame. Draws are
  // throttled to ~20 Hz and the apple layer is skipped during interpolation
  // (apples are frozen to the `from` frame) to keep pixi fed without melting
  // the browser on matches with many apples.
  useEffect(() => {
    if (!playing) return
    const duration = 1000
    const minRenderIntervalMs = 50
    let rafId: number | null = null
    let currentTurn = turnRef.current
    let segmentStart = performance.now()
    let lastRender = 0

    const tick = (now: number) => {
      // Re-sync if the user scrubbed the slider or used arrows mid-playback.
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
        updateFrame(lerpFrame(fromFrame, toFrame, t), myBirdIdsRef.current, { skipApples: true })
        lastRender = now
      }
      rafId = requestAnimationFrame(tick)
    }

    rafId = requestAnimationFrame(tick)
    return () => {
      if (rafId !== null) cancelAnimationFrame(rafId)
    }
  }, [playing, goToTurn])

  // Initialise renderer when mapData/trace change.
  //
  // Slider layout (N = trace.turns.length):
  //   - slider 0: initial state (what players saw at turn 0), no moves/events
  //   - slider i (1..N): state AFTER turn (i-1)'s update, with moves/events
  //     from turn (i-1). State is taken from turns[i].game_input.p0 when it
  //     exists, else falls back to turns[i-1].game_input.p0 for the final turn.
  useEffect(() => {
    const container = containerRef.current
    if (!container) return
    let cancelled = false

    setReady(false)
    setPlaying(false)
    setPauseRequested(false)
    myBirdIdsRef.current = mapData.myBirdIds

    const N = trace.turns.length
    const frames: FrameData[] = []
    const turns: (TraceTurn | null)[] = []

    const initialTurn = trace.turns[0]
    if (initialTurn) {
      frames.push(parseFrameLines(initialTurn.game_input.p0))
      turns.push(null)
      for (let i = 1; i <= N; i++) {
        const source = trace.turns[i] ?? trace.turns[i - 1]!
        frames.push(parseFrameLines(source.game_input.p0))
        turns.push(trace.turns[i - 1] ?? null)
      }
    }
    framesRef.current = frames
    turnsRef.current = turns

    const run = async () => {
      await initRenderer(container, mapData)
      if (cancelled) return
      turnRef.current = 0
      const firstFrame = frames[0]
      if (firstFrame) {
        updateFrame(firstFrame, mapData.myBirdIds)
        const myIdSet = new Set(mapData.myBirdIds)
        let s0 = 0
        let s1 = 0
        for (const bird of firstFrame.birds) {
          const segs = bird.body.length
          if (myIdSet.has(bird.id)) s0 += segs
          else s1 += segs
        }
        setCurrentScores([s0, s1])
      } else {
        setCurrentScores([0, 0])
      }
      setSliderMax(Math.max(0, frames.length - 1))
      setSliderValue(0)
      setTurnDisplay(`turn 0 / ${Math.max(0, frames.length - 1)}`)
      setTurnMoves([])
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
                <span>{turnDisplay}</span>
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
                  {turnMoves.map((row) => (
                    <div key={row.birdId} className="flex items-center gap-1.5">
                      <span className={`w-6 shrink-0 ${row.mine ? "text-sky-400" : "text-red-400"}`}>S{row.birdId}</span>
                      <span className="w-7 shrink-0 tabular-nums text-foreground/80">{row.size !== undefined ? `[${row.size}]` : ""}</span>
                      <span className="w-3 shrink-0">{row.mine ? "\u2192" : "\u2190"}</span>
                      <span className="w-14 shrink-0">{row.direction}</span>
                      {row.events.map((e, i) => (
                        <EventBadge key={`${e.kind}-${i}`} event={e} />
                      ))}
                    </div>
                  ))}
                </div>
              ) : (
                <p className="font-mono text-xs text-muted-foreground">initial state</p>
              )}
            </CardContent>
          </Card>
        )}
      </div>

      <div className="min-w-0 flex-1 overflow-hidden">
        {status && <p className="mb-3 font-mono text-xs text-muted-foreground">{status}</p>}
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
