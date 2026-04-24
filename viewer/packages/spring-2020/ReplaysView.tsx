import { Button } from "@shared/components/ui/button.tsx"
import { Link } from "@tanstack/react-router"
import { useEffect, useState } from "react"

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

export function ReplaysView() {
  const [list, setList] = useState<ReplayEntry[] | null>(null)
  const [listError, setListError] = useState("")

  useEffect(() => {
    fetch("/api/replays")
      .then((res) => {
        if (!res.ok) throw new Error(`${res.status}`)
        return res.json()
      })
      .then((data: ReplayEntry[]) => setList(data))
      .catch((err) => setListError(String(err)))
  }, [])

  return (
    <div className="flex gap-8">
      <div className="w-80 shrink-0" />
      <div className="min-w-0 flex-1 overflow-hidden">
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
                        <Button asChild variant="outline" size="sm">
                          <Link to="/replay/$id" params={{ id: r.id }}>
                            Replay
                          </Link>
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
