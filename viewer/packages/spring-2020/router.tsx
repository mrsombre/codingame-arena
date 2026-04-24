import { createHashHistory, createRootRoute, createRoute, createRouter, redirect, useParams } from "@tanstack/react-router"
import App, { useBots } from "./App.tsx"
import { MassView } from "./MassView.tsx"
import { PlayView } from "./PlayView.tsx"
import { ReplaysView } from "./ReplaysView.tsx"
import { ReplayView } from "./ReplayView.tsx"

const rootRoute = createRootRoute({ component: App })

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  beforeLoad: () => {
    throw redirect({ to: "/play" })
  },
})

function PlayRoute() {
  return <PlayView bots={useBots()} />
}

function MassRoute() {
  return <MassView bots={useBots()} />
}

function ReplayRoute() {
  const { id } = useParams({ from: "/replay/$id" })
  return <ReplayView id={id} />
}

const playRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/play",
  component: PlayRoute,
})

const massRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/mass",
  component: MassRoute,
})

const replaysRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/replays",
  component: ReplaysView,
})

const replayRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/replay/$id",
  component: ReplayRoute,
})

const routeTree = rootRoute.addChildren([indexRoute, playRoute, massRoute, replaysRoute, replayRoute])

export const router = createRouter({
  routeTree,
  history: createHashHistory(),
})

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router
  }
}
