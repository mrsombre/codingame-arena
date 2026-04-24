import { Button } from "@shared/components/ui/button.tsx"
import { Link } from "@tanstack/react-router"
import { ArrowLeftIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { type MapData, parseSerializeResponse, type TraceMatch } from "./parser.ts"
import { ReplayViewer } from "./ReplayViewer.tsx"

interface ReplayResponse extends TraceMatch {
  league: number
}

interface ReplayViewProps {
  replayId: string
}

export function ReplayView({ replayId }: ReplayViewProps) {
  const [status, setStatus] = useState(`loading replay ${replayId}\u2026`)
  const [mapData, setMapData] = useState<MapData | null>(null)
  const [trace, setTrace] = useState<TraceMatch | null>(null)

  useEffect(() => {
    let cancelled = false
    setMapData(null)
    setTrace(null)
    setStatus(`loading replay ${replayId}\u2026`)
    ;(async () => {
      try {
        const traceRes = await fetch(`/api/replays/${encodeURIComponent(replayId)}`)
        if (!traceRes.ok) {
          if (!cancelled) setStatus(`replay error ${traceRes.status}: ${await traceRes.text()}`)
          return
        }
        const replay: ReplayResponse = await traceRes.json()
        const leagueQuery = replay.league > 0 ? `&league=${replay.league}` : ""
        const serRes = await fetch(`/api/serialize?seed=${encodeURIComponent(replay.seed)}${leagueQuery}`)
        if (!serRes.ok) {
          if (!cancelled) setStatus(`serialize error ${serRes.status}: ${await serRes.text()}`)
          return
        }
        const map = parseSerializeResponse(await serRes.text())
        if (cancelled) return
        setMapData(map)
        setTrace(replay)
        setStatus("")
      } catch (err) {
        if (!cancelled) setStatus(`error: ${String(err)}`)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [replayId])

  const backCard = (
    <Button asChild variant="outline" size="sm" className="self-start">
      <Link to="/replays">
        <ArrowLeftIcon data-icon="inline-start" /> Back to list
      </Link>
    </Button>
  )

  if (mapData && trace) {
    const p0 = trace.bots[0] ?? "p0"
    const p1 = trace.bots[1] ?? "p1"
    const winnerLabel = trace.winner === -1 ? "draw" : `p${trace.winner}`
    const replayStatus = `replay ${replayId}  seed=${trace.seed}  ${p0} vs ${p1}  winner=${winnerLabel}  score=${trace.scores[0]}:${trace.scores[1]}  turns=${trace.turns.length}`
    return <ReplayViewer mapData={mapData} trace={trace} status={replayStatus} leftSlot={backCard} />
  }

  return (
    <div className="flex gap-8">
      <div className="w-80 shrink-0">{backCard}</div>
      <div className="min-w-0 flex-1 overflow-hidden">{status && <p className="font-mono text-xs text-muted-foreground">{status}</p>}</div>
    </div>
  )
}
