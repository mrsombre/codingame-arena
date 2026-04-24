import type { BotEntry } from "@shared/api.ts"
import { Button } from "@shared/components/ui/button.tsx"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@shared/components/ui/card.tsx"
import { Input } from "@shared/components/ui/input.tsx"
import { Label } from "@shared/components/ui/label.tsx"
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "@shared/components/ui/select.tsx"
import { useNavigate } from "@tanstack/react-router"
import { LoaderIcon, PlayIcon } from "lucide-react"
import { useState } from "react"
import { type MapData, parseSerializeResponse, type TraceMatch } from "./parser.ts"

interface MassViewProps {
  bots: BotEntry[]
}

export interface BatchMatchCacheEntry {
  match: BatchMatch
  mapData: MapData
  trace: TraceMatch
}

export const batchMatchCache = new Map<string, BatchMatchCacheEntry>()

let lastBatch: { response: BatchResponse; league: string } | null = null

export interface BatchMatch {
  id: number
  seed: string
  winner: number
  score_p0: number
  score_p1: number
  turns: number
  ttfo_p0_ms: number
  ttfo_p1_ms: number
  aot_p0_ms: number
  aot_p1_ms: number
  p0_bot: string
  p1_bot: string
}

interface BatchResponse {
  simulations: number
  wins_p0: number
  wins_p1: number
  draws: number
  avg_score_p0: number
  avg_score_p1: number
  avg_turns: number
  avg_ttfo_p0_ms: number
  avg_ttfo_p1_ms: number
  avg_aot_p0_ms: number
  avg_aot_p1_ms: number
  seed: string
  p0_bot: string
  p1_bot: string
  matches: BatchMatch[]
}

export function MassView({ bots }: MassViewProps) {
  const navigate = useNavigate()
  const [seed, setSeed] = useState("")
  const [league, setLeague] = useState(lastBatch?.league ?? "4")
  const [p0Bot, setP0Bot] = useState(bots[0]?.path ?? "")
  const [p1Bot, setP1Bot] = useState(bots[1]?.path ?? bots[0]?.path ?? "")
  const [simulations, setSimulations] = useState("50")
  const [maxTurns, setMaxTurns] = useState("200")

  const [status, setStatus] = useState("")
  const [running, setRunning] = useState(false)
  const [batch, setBatch] = useState<BatchResponse | null>(lastBatch?.response ?? null)

  const [loadingMatch, setLoadingMatch] = useState<number | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!p0Bot || !p1Bot) {
      setStatus("select bots for both players")
      return
    }
    const sims = Number(simulations)
    if (!Number.isFinite(sims) || sims < 1) {
      setStatus("simulations must be >= 1")
      return
    }
    const turns = Number(maxTurns)
    if (!Number.isFinite(turns) || turns < 1) {
      setStatus("turns must be >= 1")
      return
    }

    setRunning(true)
    setStatus(`running ${sims} match${sims === 1 ? "" : "es"}\u2026`)
    setBatch(null)
    lastBatch = null
    batchMatchCache.clear()

    try {
      const body: Record<string, unknown> = {
        p0Bin: p0Bot,
        p1Bin: p1Bot,
        simulations: sims,
        maxTurns: turns,
        gameOptions: { league },
      }
      const seedTrimmed = seed.trim()
      if (seedTrimmed) body.seed = seedTrimmed

      const res = await fetch("/api/batch", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      })
      if (!res.ok) {
        setStatus(`batch error ${res.status}: ${await res.text()}`)
        return
      }
      const data: BatchResponse = await res.json()
      lastBatch = { response: data, league }
      setBatch(data)
      setStatus("")
    } catch (err) {
      setStatus(`error: ${String(err)}`)
    } finally {
      setRunning(false)
    }
  }

  const openMatch = async (m: BatchMatch) => {
    setLoadingMatch(m.id)
    setStatus(`loading match ${m.id}\u2026`)
    try {
      const [serRes, traceRes] = await Promise.all([fetch(`/api/serialize?seed=${encodeURIComponent(m.seed)}&league=${encodeURIComponent(league)}`), fetch(`/api/matches/${m.id}`)])
      if (!serRes.ok) {
        setStatus(`serialize error ${serRes.status}: ${await serRes.text()}`)
        return
      }
      if (!traceRes.ok) {
        setStatus(`trace error ${traceRes.status}: ${await traceRes.text()}`)
        return
      }
      const serText = await serRes.text()
      const traceJson: TraceMatch = await traceRes.json()
      const mapData = parseSerializeResponse(serText)
      batchMatchCache.set(String(m.id), { match: m, mapData, trace: traceJson })
      setStatus("")
      navigate({ to: "/batch/$matchId", params: { matchId: String(m.id) } })
    } catch (err) {
      setStatus(`error: ${String(err)}`)
    } finally {
      setLoadingMatch(null)
    }
  }

  const form = (
    <Card size="sm">
      <CardContent>
        <form id="mass-form" className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <div className="flex gap-4">
            <div className="flex flex-1 flex-col gap-1.5">
              <Label htmlFor="mass-seed">Start seed</Label>
              <Input id="mass-seed" inputMode="numeric" autoComplete="off" spellCheck={false} placeholder="random" value={seed} onChange={(e) => setSeed(e.target.value)} />
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
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
              <Label>P0</Label>
              <Select value={p0Bot} onValueChange={setP0Bot}>
                <SelectTrigger className="w-full" size="sm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {bots.map((b) => (
                      <SelectItem key={b.path} value={b.path}>
                        {b.name}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
              <Label>P1</Label>
              <Select value={p1Bot} onValueChange={setP1Bot}>
                <SelectTrigger className="w-full" size="sm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {bots.map((b) => (
                      <SelectItem key={b.path} value={b.path}>
                        {b.name}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="flex gap-4">
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
              <Label htmlFor="mass-sims">Matches</Label>
              <Input id="mass-sims" inputMode="numeric" autoComplete="off" spellCheck={false} value={simulations} onChange={(e) => setSimulations(e.target.value)} />
            </div>
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
              <Label htmlFor="mass-turns">Turns</Label>
              <Input id="mass-turns" inputMode="numeric" autoComplete="off" spellCheck={false} value={maxTurns} onChange={(e) => setMaxTurns(e.target.value)} />
            </div>
          </div>
        </form>
      </CardContent>
      <CardFooter className="border-t">
        <Button type="submit" form="mass-form" className="w-full" disabled={running}>
          {running ? <LoaderIcon data-icon="inline-start" className="animate-spin" /> : <PlayIcon data-icon="inline-start" />}
          {running ? "Simulating\u2026" : "Simulate"}
        </Button>
      </CardFooter>
    </Card>
  )

  const summary =
    batch &&
    (() => {
      const n = batch.simulations || 1
      const winPct = (v: number) => `${((v / n) * 100).toFixed(1)}%`
      return (
        <Card size="sm">
          <CardHeader>
            <CardTitle className="text-xs">Summary</CardTitle>
          </CardHeader>
          <CardContent>
            <dl className="grid grid-cols-2 gap-x-4 gap-y-1 font-mono text-xs text-muted-foreground">
              <dt>Matches</dt>
              <dd>{batch.simulations}</dd>
              <dt className="text-sky-400">{batch.p0_bot} wins</dt>
              <dd>
                {batch.wins_p0} ({winPct(batch.wins_p0)})
              </dd>
              <dt className="text-red-400">{batch.p1_bot} wins</dt>
              <dd>
                {batch.wins_p1} ({winPct(batch.wins_p1)})
              </dd>
              <dt>Draws</dt>
              <dd>
                {batch.draws} ({winPct(batch.draws)})
              </dd>
              <dt className="text-sky-400">{batch.p0_bot} avg</dt>
              <dd>{batch.avg_score_p0.toFixed(2)}</dd>
              <dt className="text-red-400">{batch.p1_bot} avg</dt>
              <dd>{batch.avg_score_p1.toFixed(2)}</dd>
              <dt>Avg turns</dt>
              <dd>{batch.avg_turns.toFixed(1)}</dd>
              <dt className="text-sky-400">p0 ttfo/aot</dt>
              <dd>
                {batch.avg_ttfo_p0_ms.toFixed(0)}ms/{batch.avg_aot_p0_ms.toFixed(0)}ms
              </dd>
              <dt className="text-red-400">p1 ttfo/aot</dt>
              <dd>
                {batch.avg_ttfo_p1_ms.toFixed(0)}ms/{batch.avg_aot_p1_ms.toFixed(0)}ms
              </dd>
              <dt>Seed</dt>
              <dd className="truncate" title={String(batch.seed)}>
                {batch.seed}
              </dd>
            </dl>
          </CardContent>
        </Card>
      )
    })()

  return (
    <div className="flex gap-8">
      <div className="flex w-80 shrink-0 flex-col gap-4 overflow-hidden">
        {form}
        {summary}
      </div>
      <div className="min-w-0 flex-1 overflow-hidden">
        {status && <p className="mb-3 font-mono text-xs text-muted-foreground">{status}</p>}
        {batch && batch.matches.length > 0 && (
          <div className="overflow-auto rounded-sm border">
            <table className="w-full font-mono text-xs">
              <thead className="bg-muted text-left text-muted-foreground">
                <tr>
                  <th className="px-3 py-2">#</th>
                  <th className="px-3 py-2">Seed</th>
                  <th className="px-3 py-2">Sides</th>
                  <th className="px-3 py-2">Winner</th>
                  <th className="px-3 py-2">Score</th>
                  <th className="px-3 py-2">Turns</th>
                  <th className="px-3 py-2">Timing</th>
                  <th className="px-3 py-2" />
                </tr>
              </thead>
              <tbody>
                {batch.matches.map((m) => {
                  // Color by user-selected role: user's P0 bot is always blue,
                  // user's P1 bot is always red, no matter which in-match side
                  // they played. Winner column picks up the winning bot's color.
                  const userColor = (botName: string) => {
                    if (botName === batch.p0_bot) return "text-sky-400"
                    if (botName === batch.p1_bot) return "text-red-400"
                    return ""
                  }
                  const winnerLabel = m.winner === -1 ? "draw" : `p${m.winner}`
                  const winnerClass = m.winner === 0 ? userColor(m.p0_bot) : m.winner === 1 ? userColor(m.p1_bot) : "text-muted-foreground"
                  return (
                    <tr key={m.id} className="border-t hover:bg-accent/40">
                      <td className="px-3 py-1.5">{m.id}</td>
                      <td className="px-3 py-1.5 text-muted-foreground">{m.seed}</td>
                      <td className="px-3 py-1.5">
                        <span className={userColor(m.p0_bot)}>{m.p0_bot}</span>
                        <span className="text-muted-foreground"> vs </span>
                        <span className={userColor(m.p1_bot)}>{m.p1_bot}</span>
                      </td>
                      <td className={`px-3 py-1.5 ${winnerClass}`}>{winnerLabel}</td>
                      <td className="px-3 py-1.5">
                        {m.score_p0}:{m.score_p1}
                      </td>
                      <td className="px-3 py-1.5 text-muted-foreground">{m.turns}</td>
                      <td className="px-3 py-1.5 text-muted-foreground">
                        {m.ttfo_p0_ms.toFixed(0)}/{m.aot_p0_ms.toFixed(0)} vs {m.ttfo_p1_ms.toFixed(0)}/{m.aot_p1_ms.toFixed(0)}ms
                      </td>
                      <td className="px-3 py-1.5 text-right">
                        <Button variant="outline" size="sm" disabled={loadingMatch !== null} onClick={() => openMatch(m)}>
                          {loadingMatch === m.id ? <LoaderIcon className="size-3 animate-spin" /> : "Replay"}
                        </Button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
