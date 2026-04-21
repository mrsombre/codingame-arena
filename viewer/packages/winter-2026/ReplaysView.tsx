import { Button } from "@shared/components/ui/button.tsx"
import { ArrowLeftIcon, LoaderIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { type MapData, parseSerializeResponse, type TraceMatch } from "./parser.ts"
import { ReplayViewer } from "./ReplayViewer.tsx"

interface ReplayEntry {
  id: string
  size: number
  mtime: string
  p0_name?: string
  p1_name?: string
  league?: number
  score_p0: number
  score_p1: number
  winner: number
}

interface ReplayResponse extends TraceMatch {
  league: number
}

export function ReplaysView() {
  const [list, setList] = useState<ReplayEntry[] | null>(null)
  const [listError, setListError] = useState("")

  const [loadingId, setLoadingId] = useState<string | null>(null)
  const [status, setStatus] = useState("")
  const [mapData, setMapData] = useState<MapData | null>(null)
  const [trace, setTrace] = useState<TraceMatch | null>(null)
  const [selected, setSelected] = useState<ReplayEntry | null>(null)

  useEffect(() => {
    fetch("/api/replays")
      .then((res) => {
        if (!res.ok) throw new Error(`${res.status}`)
        return res.json()
      })
      .then((data: ReplayEntry[]) => setList(data))
      .catch((err) => setListError(String(err)))
  }, [])

  const openReplay = async (entry: ReplayEntry) => {
    setLoadingId(entry.id)
    setStatus(`loading replay ${entry.id}\u2026`)
    try {
      const traceRes = await fetch(`/api/replays/${encodeURIComponent(entry.id)}`)
      if (!traceRes.ok) {
        setStatus(`replay error ${traceRes.status}: ${await traceRes.text()}`)
        return
      }
      const replay: ReplayResponse = await traceRes.json()
      const leagueQuery = replay.league > 0 ? `&league=${replay.league}` : ""
      const serRes = await fetch(`/api/serialize?seed=${encodeURIComponent(replay.seed)}${leagueQuery}`)
      if (!serRes.ok) {
        setStatus(`serialize error ${serRes.status}: ${await serRes.text()}`)
        return
      }
      const map = parseSerializeResponse(await serRes.text())
      setMapData(map)
      setTrace(replay)
      setSelected(entry)
      setStatus("")
    } catch (err) {
      setStatus(`error: ${String(err)}`)
    } finally {
      setLoadingId(null)
    }
  }

  const back = () => {
    setMapData(null)
    setTrace(null)
    setSelected(null)
  }

  if (mapData && trace && selected) {
    const p0 = selected.p0_name ?? trace.bots[0] ?? "p0"
    const p1 = selected.p1_name ?? trace.bots[1] ?? "p1"
    const winnerLabel = trace.winner === -1 ? "draw" : `p${trace.winner}`
    const replayStatus = `replay ${selected.id}  seed=${trace.seed}  ${p0} vs ${p1}  winner=${winnerLabel}  score=${trace.scores[0]}:${trace.scores[1]}  turns=${trace.turns.length}`
    const backCard = (
      <Button variant="outline" size="sm" className="self-start" onClick={back}>
        <ArrowLeftIcon data-icon="inline-start" /> Back to list
      </Button>
    )
    return <ReplayViewer mapData={mapData} trace={trace} status={replayStatus} leftSlot={backCard} />
  }

  return (
    <div className="flex gap-8">
      <div className="w-80 shrink-0" />
      <div className="min-w-0 flex-1 overflow-hidden">
        {status && <p className="mb-3 font-mono text-xs text-muted-foreground">{status}</p>}
        {listError && <p className="mb-3 font-mono text-xs text-destructive">{listError}</p>}
        {list === null && !listError && <p className="font-mono text-xs text-muted-foreground">loading…</p>}
        {list !== null && list.length === 0 && (
          <p className="font-mono text-xs text-muted-foreground">
            No replays. Use <code>arena replay &lt;url|id&gt;</code> to download.
          </p>
        )}
        {list !== null && list.length > 0 && (
          <div className="overflow-auto rounded-sm border">
            <table className="w-full font-mono text-xs">
              <thead className="bg-muted text-left text-muted-foreground">
                <tr>
                  <th className="px-3 py-2">Replay ID</th>
                  <th className="px-3 py-2">Players</th>
                  <th className="px-3 py-2">Winner</th>
                  <th className="px-3 py-2">Score</th>
                  <th className="px-3 py-2">Modified</th>
                  <th className="px-3 py-2" />
                </tr>
              </thead>
              <tbody>
                {list.map((r) => {
                  const winnerLabel = r.winner === 0 ? (r.p0_name ?? "p0") : r.winner === 1 ? (r.p1_name ?? "p1") : "draw"
                  const winnerClass = r.winner === 0 ? "text-sky-400" : r.winner === 1 ? "text-red-400" : "text-muted-foreground"
                  return (
                    <tr key={r.id} className="border-t hover:bg-accent/40">
                      <td className="px-3 py-1.5">{r.id}</td>
                      <td className="px-3 py-1.5">
                        <span className="text-sky-400">{r.p0_name ?? "p0"}</span>
                        <span className="text-muted-foreground"> vs </span>
                        <span className="text-red-400">{r.p1_name ?? "p1"}</span>
                      </td>
                      <td className={`px-3 py-1.5 ${winnerClass}`}>{winnerLabel}</td>
                      <td className="px-3 py-1.5">
                        <span className="text-sky-400">{r.score_p0}</span>
                        <span className="text-muted-foreground">:</span>
                        <span className="text-red-400">{r.score_p1}</span>
                      </td>
                      <td className="px-3 py-1.5 text-muted-foreground">{new Date(r.mtime).toLocaleString()}</td>
                      <td className="px-3 py-1.5 text-right">
                        <Button variant="outline" size="sm" disabled={loadingId !== null} onClick={() => openReplay(r)}>
                          {loadingId === r.id ? <LoaderIcon className="size-3 animate-spin" /> : "Replay"}
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
