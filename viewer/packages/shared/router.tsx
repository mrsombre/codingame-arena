import { type BotEntry, fetchBots } from "@shared/api.ts"
import { Button } from "@shared/components/ui/button.tsx"
import { Card, CardContent, CardHeader, CardTitle } from "@shared/components/ui/card.tsx"
import { createHashHistory, createRootRoute, createRoute, createRouter, Link, Outlet, redirect, useParams, useRouterState } from "@tanstack/react-router"
import { LoaderIcon } from "lucide-react"
import { type ComponentType, createContext, useContext, useEffect, useState } from "react"

// --- Bots context ---

const BotsContext = createContext<BotEntry[] | null>(null)

export function useBots(): BotEntry[] {
  const bots = useContext(BotsContext)
  if (!bots) throw new Error("BotsContext missing")
  return bots
}

// --- App shell (root route component) ---

function AppShell({ title }: { title: string }) {
  const [bots, setBots] = useState<BotEntry[] | null>(null)
  const [botsError, setBotsError] = useState(false)
  const pathname = useRouterState({ select: (s) => s.location.pathname })

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
            <CardTitle>{title}</CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">{botsError ? "Failed to load bots" : "No bots available. Add binaries to --bin-dir."}</CardContent>
        </Card>
      </div>
    )
  }

  const tabLink = (path: "/single" | "/batch" | "/replays", label: string) => {
    const active = path === "/single" ? pathname === "/single" : pathname.startsWith(path)
    return (
      <Button asChild variant={active ? "default" : "ghost"} size="sm">
        <Link to={path}>{label}</Link>
      </Button>
    )
  }

  return (
    <BotsContext.Provider value={bots}>
      <div className="flex flex-col gap-6 p-10">
        <div className="flex items-center gap-4">
          <h1 className="font-mono text-sm uppercase tracking-wider text-muted-foreground">{title}</h1>
          <div className="flex gap-1 rounded-md border p-1">
            {tabLink("/single", "Single")}
            {tabLink("/batch", "Batch")}
            {tabLink("/replays", "Replays")}
          </div>
        </div>
        <Outlet />
      </div>
    </BotsContext.Provider>
  )
}

// --- Router factory ---

export interface GameViews {
  SingleView: ComponentType<{ bots: BotEntry[] }>
  BatchView: ComponentType<{ bots: BotEntry[] }>
  BatchMatchView: ComponentType<{ matchId: string }>
  ReplaysView: ComponentType
  ReplayView: ComponentType<{ replayId: string }>
}

export function createGameRouter(views: GameViews, title: string) {
  const rootRoute = createRootRoute({
    component: () => <AppShell title={title} />,
  })

  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/",
    beforeLoad: () => {
      throw redirect({ to: "/single" })
    },
  })

  const singleRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/single",
    component: function SingleRoute() {
      return <views.SingleView bots={useBots()} />
    },
  })

  const batchRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/batch",
    component: function BatchRoute() {
      return <views.BatchView bots={useBots()} />
    },
  })

  const batchMatchRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/batch/$matchId",
    component: function BatchMatchRoute() {
      const { matchId } = useParams({ from: "/batch/$matchId" })
      return <views.BatchMatchView matchId={matchId} />
    },
  })

  const replaysRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/replays",
    component: function ReplaysRoute() {
      return <views.ReplaysView />
    },
  })

  const replayRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/replays/$replayId",
    component: function ReplayRoute() {
      const { replayId } = useParams({ from: "/replays/$replayId" })
      return <views.ReplayView replayId={replayId} />
    },
  })

  const routeTree = rootRoute.addChildren([indexRoute, singleRoute, batchRoute, batchMatchRoute, replaysRoute, replayRoute])

  return createRouter({
    routeTree,
    history: createHashHistory(),
  })
}
