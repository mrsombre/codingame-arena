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
  const winnerLabel = m.winner === -1 ? "draw" : m.winner === 0 ? m.left_bot : m.right_bot
  const replayStatus = `match #${m.id}  seed=${m.seed}  ${m.left_bot} vs ${m.right_bot}  winner=${winnerLabel}  score=${m.score_left}:${m.score_right}  turns=${m.turns}  left ttfo=${m.ttfo_left_ms.toFixed(0)}ms aot=${m.aot_left_ms.toFixed(0)}ms  right ttfo=${m.ttfo_right_ms.toFixed(0)}ms aot=${m.aot_right_ms.toFixed(0)}ms`

  const backCard = (
    <Button asChild variant="outline" size="sm" className="self-start">
      <Link to="/batch">
        <ArrowLeftIcon data-icon="inline-start" /> Back to list
      </Link>
    </Button>
  )

  return <ReplayViewer mapData={mapData} trace={trace} status={replayStatus} leftSlot={backCard} />
}
