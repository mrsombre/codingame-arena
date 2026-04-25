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
    const p0 = trace.players[0] ?? "p0"
    const p1 = trace.players[1] ?? "p1"
    const shortId = replayId.startsWith("replay-") ? replayId.slice("replay-".length) : replayId
    const winnerName = trace.winner === 0 ? p0 : trace.winner === 1 ? p1 : null
    const winnerClass = trace.winner === 0 ? "text-sky-400" : trace.winner === 1 ? "text-red-400" : "text-muted-foreground"
    const replayStatus = (
      <>
        replay: {shortId}&nbsp;&nbsp;seed={trace.seed}&nbsp;&nbsp;<span className="text-sky-400">{p0}</span> vs <span className="text-red-400">{p1}</span>&nbsp;&nbsp;winner=
        <span className={winnerClass}>{winnerName ?? "draw"}</span>&nbsp;&nbsp;score=<span className="text-sky-400">{trace.scores[0]}</span>:<span className="text-red-400">{trace.scores[1]}</span>
        &nbsp;&nbsp;turns={trace.turns.length}
      </>
    )
    return <ReplayViewer mapData={mapData} trace={trace} status={replayStatus} leftSlot={backCard} />
  }

  return (
    <div className="flex gap-8">
      <div className="w-80 shrink-0">{backCard}</div>
      <div className="min-w-0 flex-1 overflow-hidden">{status && <p className="font-mono text-xs text-muted-foreground">{status}</p>}</div>
    </div>
  )
}
