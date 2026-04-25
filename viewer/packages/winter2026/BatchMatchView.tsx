import { Button } from "@shared/components/ui/button.tsx"
import { Link } from "@tanstack/react-router"
import { ArrowLeftIcon } from "lucide-react"
import { batchMatchCache } from "./MassView.tsx"
import { ReplayViewer } from "./ReplayViewer.tsx"

interface BatchMatchViewProps {
  matchId: string
}

export function BatchMatchView({ matchId }: BatchMatchViewProps) {
  const entry = batchMatchCache.get(matchId)

  if (!entry) {
    return (
      <div className="flex gap-8">
        <div className="w-80 shrink-0">
          <Button asChild variant="outline" size="sm" className="self-start">
            <Link to="/batch">
              <ArrowLeftIcon data-icon="inline-start" /> Back to list
            </Link>
          </Button>
        </div>
        <div className="min-w-0 flex-1">
          <p className="font-mono text-xs text-muted-foreground">Match not found. Run a batch first.</p>
        </div>
      </div>
    )
  }

  const { match: m, mapData, trace } = entry
  const winnerLabel = m.winner === -1 ? "draw" : `p${m.winner}`
  const replayStatus = `match #${m.id}  seed=${m.seed}  ${m.p0_bot} vs ${m.p1_bot}  winner=${winnerLabel}  score=${m.score_p0}:${m.score_p1}  turns=${m.turns}  p0 ttfo=${m.ttfo_p0_ms.toFixed(0)}ms aot=${m.aot_p0_ms.toFixed(0)}ms  p1 ttfo=${m.ttfo_p1_ms.toFixed(0)}ms aot=${m.aot_p1_ms.toFixed(0)}ms`

  const backCard = (
    <Button asChild variant="outline" size="sm" className="self-start">
      <Link to="/batch">
        <ArrowLeftIcon data-icon="inline-start" /> Back to list
      </Link>
    </Button>
  )

  return <ReplayViewer mapData={mapData} trace={trace} status={replayStatus} leftSlot={backCard} />
}
