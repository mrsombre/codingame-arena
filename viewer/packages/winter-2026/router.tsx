import { createGameRouter } from "@shared/router.tsx"
import { BatchMatchView } from "./BatchMatchView.tsx"
import { MassView } from "./MassView.tsx"
import { PlayView } from "./PlayView.tsx"
import { ReplaysView } from "./ReplaysView.tsx"
import { ReplayView } from "./ReplayView.tsx"

export const router = createGameRouter(
  {
    SingleView: PlayView,
    BatchView: MassView,
    BatchMatchView,
    ReplaysView,
    ReplayView,
  },
  "Winter 2026",
)

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router
  }
}
