import { type BotEntry, fetchBots } from "@shared/api.ts"
import { Button } from "@shared/components/ui/button.tsx"
import { Card, CardContent, CardHeader, CardTitle } from "@shared/components/ui/card.tsx"
import { LoaderIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { MassView } from "./MassView.tsx"
import { PlayView } from "./PlayView.tsx"
import { ReplaysView } from "./ReplaysView.tsx"

type Tab = "play" | "mass" | "replays"

export default function App() {
  const [bots, setBots] = useState<BotEntry[] | null>(null)
  const [botsError, setBotsError] = useState(false)
  const [tab, setTab] = useState<Tab>("play")

  useEffect(() => {
    fetchBots()
      .then(setBots)
      .catch(() => setBotsError(true))
  }, [])

  if (bots === null) {
    return (
      <div className="flex min-h-[40vh] items-center justify-center">
        <LoaderIcon className="size-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (botsError || bots.length === 0) {
    return (
      <div className="p-10">
        <Card size="sm">
          <CardHeader>
            <CardTitle>Winter 2026</CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">{botsError ? "Failed to load bots" : "No bots available. Add binaries to --bin-dir."}</CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6 p-10">
      <div className="flex items-center gap-4">
        <h1 className="font-mono text-sm uppercase tracking-wider text-muted-foreground">Winter 2026</h1>
        <div className="flex gap-1 rounded-md border p-1">
          <Button variant={tab === "play" ? "default" : "ghost"} size="sm" onClick={() => setTab("play")}>
            Play
          </Button>
          <Button variant={tab === "mass" ? "default" : "ghost"} size="sm" onClick={() => setTab("mass")}>
            Mass
          </Button>
          <Button variant={tab === "replays" ? "default" : "ghost"} size="sm" onClick={() => setTab("replays")}>
            Replays
          </Button>
        </div>
      </div>

      {tab === "play" && <PlayView bots={bots} />}
      {tab === "mass" && <MassView bots={bots} />}
      {tab === "replays" && <ReplaysView />}
    </div>
  )
}
