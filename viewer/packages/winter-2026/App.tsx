import { useCallback, useEffect, useRef, useState } from "react"
import { type BotEntry, fetchBots } from "@shared/api.ts"
import { Button } from "@shared/components/ui/button.tsx"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@shared/components/ui/card.tsx"
import { Input } from "@shared/components/ui/input.tsx"
import { Label } from "@shared/components/ui/label.tsx"
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "@shared/components/ui/select.tsx"
import { Separator } from "@shared/components/ui/separator.tsx"
import { Slider } from "@shared/components/ui/slider.tsx"
import { ChevronsLeftIcon, ChevronLeftIcon, ChevronRightIcon, ChevronsRightIcon, PlayIcon, LoaderIcon } from "lucide-react"
import { type FrameData, parseFrameLines, parseSerializeResponse, type TraceMatch, type TraceTurn } from "./parser.ts"
import { destroyRenderer, initRenderer, updateFrame } from "./renderer.ts"

export default function App() {
  const [bots, setBots] = useState<BotEntry[] | null>(null)
  const [botsError, setBotsError] = useState(false)
  const [status, setStatus] = useState("")
  const [running, setRunning] = useState(false)
  const [showMap, setShowMap] = useState(false)
  const [showNav, setShowNav] = useState(false)

  const [seed, setSeed] = useState("")
  const [league, setLeague] = useState("4")
  const [p0Bot, setP0Bot] = useState("")
  const [p1Bot, setP1Bot] = useState("")

  const framesRef = useRef<FrameData[]>([])
  const turnsRef = useRef<TraceTurn[]>([])
  const myBirdIdsRef = useRef<number[]>([])
  const [turnMoves, setTurnMoves] = useState<string[]>([])
  const turnRef = useRef(0)
  const [turnDisplay, setTurnDisplay] = useState("")
  const [sliderMax, setSliderMax] = useState(0)
  const [sliderValue, setSliderValue] = useState(0)

  const containerRef = useRef<HTMLDivElement>(null)

  // Load bots
  useEffect(() => {
    fetchBots()
      .then((b) => {
        setBots(b)
        if (b.length >= 2) {
          setP0Bot(b[0]!.path)
          setP1Bot(b[1]!.path)
        } else if (b[0]) {
          setP0Bot(b[0].path)
          setP1Bot(b[0].path)
        }
      })
      .catch(() => setBotsError(true))
  }, [])

  // Parse move lines from trace output: "0 RIGHT;1 LEFT" → ["S0 → RIGHT", "S1 → LEFT"]
  const parseMoves = useCallback((turn: TraceTurn, myIds: Set<number>) => {
    const lines: string[] = []
    for (const [output, arrow] of [[turn.p0_output, "\u2192"], [turn.p1_output, "\u2190"]] as const) {
      if (!output) continue
      for (const cmd of output.split(";")) {
        const parts = cmd.trim().split(" ")
        if (parts.length >= 2) {
          const id = parts[0]
          const dir = parts[1]
          const prefix = myIds.has(Number(id)) ? "\u2192" : "\u2190"
          lines.push(`S${id} ${prefix} ${dir}`)
        }
      }
    }
    return lines
  }, [])

  // Turn navigation
  const goToTurn = useCallback((t: number) => {
    const frames = framesRef.current
    const clamped = Math.max(0, Math.min(t, frames.length - 1))
    turnRef.current = clamped
    const frame = frames[clamped]
    if (frame) updateFrame(frame, myBirdIdsRef.current)
    setSliderValue(clamped)
    setTurnDisplay(`turn ${clamped} / ${frames.length - 1}`)

    const turn = turnsRef.current[clamped]
    if (turn) {
      setTurnMoves(parseMoves(turn, new Set(myBirdIdsRef.current)))
    } else {
      setTurnMoves([])
    }
  }, [parseMoves])

  // Keyboard nav
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

  // Run match
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!p0Bot || !p1Bot) {
      setStatus("select bots for both players")
      return
    }

    setRunning(true)
    setStatus("running match\u2026")
    setShowNav(false)
    setShowMap(false)
    destroyRenderer()

    try {
      const runBody: Record<string, unknown> = {
        p0Bin: p0Bot,
        p1Bin: p1Bot,
        gameOptions: { league },
      }
      const seedTrimmed = seed.trim()
      if (seedTrimmed) {
        runBody.seed = Number(seedTrimmed)
      }

      const runRes = await fetch("/api/run", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(runBody),
      })
      if (!runRes.ok) {
        setStatus(`run error ${runRes.status}: ${await runRes.text()}`)
        return
      }
      const runData = await runRes.json()
      const actualSeed: number = runData.seed

      setStatus("loading replay\u2026")

      const [serRes, traceRes] = await Promise.all([
        fetch(`/api/serialize?seed=${actualSeed}&league=${encodeURIComponent(league)}`),
        fetch("/api/matches/0"),
      ])

      if (!serRes.ok) {
        setStatus(`serialize error ${serRes.status}: ${await serRes.text()}`)
        return
      }
      if (!traceRes.ok) {
        setStatus(`trace error ${traceRes.status}: ${await traceRes.text()} \u2014 is --trace-dir set?`)
        return
      }

      const serText = await serRes.text()
      const traceJson: TraceMatch = await traceRes.json()

      const mapData = parseSerializeResponse(serText)
      myBirdIdsRef.current = mapData.myBirdIds

      const frames = traceJson.turns.map((t) => parseFrameLines(t.game_input.p0))
      framesRef.current = frames
      turnsRef.current = traceJson.turns

      if (frames.length === 0) {
        setStatus("no turns in trace")
        return
      }

      const container = containerRef.current
      if (!container) return
      container.hidden = false
      await initRenderer(container, mapData)

      turnRef.current = 0
      const firstFrame = frames[0]
      if (firstFrame) updateFrame(firstFrame, myBirdIdsRef.current)

      setSliderMax(frames.length - 1)
      setSliderValue(0)
      setTurnDisplay(`turn 0 / ${frames.length - 1}`)

      const winnerStr = runData.winner === -1 ? "draw" : `p${runData.winner}`
      setStatus(
        `seed=${actualSeed}  ${mapData.width}\u00d7${mapData.height}  winner=${winnerStr}  score=${runData.score_p0}:${runData.score_p1}  turns=${runData.turns}`,
      )
      setShowMap(true)
      setShowNav(true)

    } catch (err) {
      setStatus(`error: ${String(err)}`)
    } finally {
      setRunning(false)
    }
  }

  const botSelectPlaceholder = botsError ? "failed to load" : "loading\u2026"
  const botName = (path: string) => bots?.find((b) => b.path === path)?.name ?? path

  return (
    <div className="flex gap-8 p-10">
      {/* Left column: controls + turn log */}
      <div className="flex w-80 shrink-0 flex-col gap-4 overflow-hidden">
        <Card size="sm">
          <CardHeader className="border-b">
            <CardTitle>Winter 2026</CardTitle>
          </CardHeader>
          {bots === null ? (
            <CardContent className="flex items-center justify-center py-10">
              <LoaderIcon className="size-5 animate-spin text-muted-foreground" />
            </CardContent>
          ) : botsError ? (
            <CardContent className="py-10 text-center text-sm text-muted-foreground">
              Failed to load bots
            </CardContent>
          ) : (
          <>
          <CardContent>
            <form id="game-form" className="flex flex-col gap-4" onSubmit={handleSubmit}>
              <div className="flex gap-4">
                <div className="flex flex-1 flex-col gap-1.5">
                  <Label htmlFor="seed">Seed</Label>
                  <Input
                    id="seed"
                    inputMode="numeric"
                    autoComplete="off"
                    spellCheck={false}
                    placeholder="random"
                    value={seed}
                    onChange={(e) => setSeed(e.target.value)}
                  />
                </div>
                <div className="flex w-28 flex-col gap-1.5">
                  <Label>League</Label>
                  <Select value={league} onValueChange={setLeague}>
                    <SelectTrigger className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="1">Bronze</SelectItem>
                        <SelectItem value="2">Silver</SelectItem>
                        <SelectItem value="3">Gold</SelectItem>
                        <SelectItem value="4">Legend</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="flex gap-4">
                <div className="flex flex-1 flex-col gap-1.5">
                  <Label>P0</Label>
                  <Select value={p0Bot} onValueChange={setP0Bot}>
                    <SelectTrigger className="w-full" size="sm">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {bots.map((b) => (
                          <SelectItem key={b.path} value={b.path}>{b.name}</SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex flex-1 flex-col gap-1.5">
                  <Label>P1</Label>
                  <Select value={p1Bot} onValueChange={setP1Bot}>
                    <SelectTrigger className="w-full" size="sm">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {bots.map((b) => (
                          <SelectItem key={b.path} value={b.path}>{b.name}</SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </form>
          </CardContent>
          <CardFooter className="border-t">
            <Button type="submit" form="game-form" className="w-full" disabled={running}>
              {running ? (
                <LoaderIcon data-icon="inline-start" className="animate-spin" />
              ) : (
                <PlayIcon data-icon="inline-start" />
              )}
              {running ? "Running\u2026" : "Run Match"}
            </Button>
          </CardFooter>
          </>
          )}
        </Card>

        {/* Turn log */}
        {turnMoves.length > 0 && (
          <Card size="sm">
            <CardHeader>
              <CardTitle className="text-xs">{turnDisplay}</CardTitle>
            </CardHeader>
            <CardContent>
              <pre className="font-mono text-xs leading-relaxed text-muted-foreground">
                {turnMoves.join("\n")}
              </pre>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Map area */}
      <div className="min-w-0 flex-1 overflow-hidden">
        {status && (
          <p className="mb-3 font-mono text-xs text-muted-foreground">{status}</p>
        )}
        <div ref={containerRef} hidden={!showMap} className="[&_canvas]:block [&_canvas]:rounded-sm" />
        {showNav && (
          <div className="mt-3 flex items-center gap-3">
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
              onValueChange={([v]) => { if (v !== undefined) goToTurn(v) }}
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
