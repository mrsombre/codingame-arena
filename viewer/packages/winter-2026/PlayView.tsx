import { useState } from "react"
import type { BotEntry } from "@shared/api.ts"
import { Button } from "@shared/components/ui/button.tsx"
import { Card, CardContent, CardFooter } from "@shared/components/ui/card.tsx"
import { Input } from "@shared/components/ui/input.tsx"
import { Label } from "@shared/components/ui/label.tsx"
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "@shared/components/ui/select.tsx"
import { LoaderIcon, PlayIcon } from "lucide-react"
import { type MapData, parseSerializeResponse, type TraceMatch } from "./parser.ts"
import { ReplayViewer } from "./ReplayViewer.tsx"

interface PlayViewProps {
  bots: BotEntry[]
}

export function PlayView({ bots }: PlayViewProps) {
  const [seed, setSeed] = useState("")
  const [league, setLeague] = useState("4")
  const [p0Bot, setP0Bot] = useState(bots[0]?.path ?? "")
  const [p1Bot, setP1Bot] = useState(bots[1]?.path ?? bots[0]?.path ?? "")

  const [status, setStatus] = useState("")
  const [running, setRunning] = useState(false)
  const [mapData, setMapData] = useState<MapData | null>(null)
  const [trace, setTrace] = useState<TraceMatch | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!p0Bot || !p1Bot) {
      setStatus("select bots for both players")
      return
    }

    setRunning(true)
    setStatus("running match\u2026")
    setMapData(null)
    setTrace(null)

    try {
      const runBody: Record<string, unknown> = {
        p0Bin: p0Bot,
        p1Bin: p1Bot,
        gameOptions: { league },
      }
      const seedTrimmed = seed.trim()
      if (seedTrimmed) runBody.seed = seedTrimmed

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
      const actualSeed: string = runData.seed

      setStatus("loading replay\u2026")

      const [serRes, traceRes] = await Promise.all([
        fetch(`/api/serialize?seed=${encodeURIComponent(actualSeed)}&league=${encodeURIComponent(league)}`),
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
      const map = parseSerializeResponse(serText)

      if (traceJson.turns.length === 0) {
        setStatus("no turns in trace")
        return
      }

      const winnerStr = runData.winner === -1 ? "draw" : `p${runData.winner}`
      setStatus(
        `seed=${actualSeed}  ${map.width}\u00d7${map.height}  winner=${winnerStr}  score=${runData.score_p0}:${runData.score_p1}  turns=${runData.turns}`,
      )
      setMapData(map)
      setTrace(traceJson)
    } catch (err) {
      setStatus(`error: ${String(err)}`)
    } finally {
      setRunning(false)
    }
  }

  const form = (
    <Card size="sm">
      <CardContent>
        <form id="play-form" className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <div className="flex gap-4">
            <div className="flex flex-1 flex-col gap-1.5">
              <Label htmlFor="play-seed">Seed</Label>
              <Input
                id="play-seed"
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
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
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
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
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
        <Button type="submit" form="play-form" className="w-full" disabled={running}>
          {running ? (
            <LoaderIcon data-icon="inline-start" className="animate-spin" />
          ) : (
            <PlayIcon data-icon="inline-start" />
          )}
          {running ? "Running\u2026" : "Run Match"}
        </Button>
      </CardFooter>
    </Card>
  )

  if (mapData && trace) {
    return <ReplayViewer mapData={mapData} trace={trace} status={status} leftSlot={form} />
  }
  return (
    <div className="flex gap-8">
      <div className="flex w-80 shrink-0 flex-col gap-4 overflow-hidden">{form}</div>
      <div className="min-w-0 flex-1 overflow-hidden">
        {status && <p className="font-mono text-xs text-muted-foreground">{status}</p>}
      </div>
    </div>
  )
}
