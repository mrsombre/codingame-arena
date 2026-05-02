import { TooltipProvider } from "@shared/components/ui/tooltip.tsx"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import type { ReactElement } from "react"
import { describe, expect, it, vi } from "vitest"
import { BatchSummaryPanel, PlaybackControls, StatusLine } from "./components.tsx"
import type { BatchResponse } from "./types.ts"

function renderWithTooltip(node: ReactElement) {
  return render(<TooltipProvider>{node}</TooltipProvider>)
}

describe("StatusLine", () => {
  it("renders children when provided", () => {
    render(<StatusLine>turn 4 / 10</StatusLine>)
    expect(screen.getByText("turn 4 / 10")).toBeInTheDocument()
  })

  it("renders nothing when children is falsy", () => {
    const { container } = render(<StatusLine>{null}</StatusLine>)
    expect(container).toBeEmptyDOMElement()
  })

  it("applies destructive styling on the destructive flag", () => {
    render(<StatusLine destructive>boom</StatusLine>)
    expect(screen.getByText("boom")).toHaveClass("text-destructive")
  })
})

describe("PlaybackControls", () => {
  function setup() {
    const handlers = {
      onFirst: vi.fn(),
      onPrevious: vi.fn(),
      onNext: vi.fn(),
      onLast: vi.fn(),
      onSliderChange: vi.fn(),
    }
    renderWithTooltip(<PlaybackControls sliderMax={10} sliderValue={3} turnLabel="turn 3 / 10" {...handlers} />)
    return handlers
  }

  it("renders the four step buttons and the turn label", () => {
    setup()
    expect(screen.getByLabelText("first turn")).toBeInTheDocument()
    expect(screen.getByLabelText("previous turn")).toBeInTheDocument()
    expect(screen.getByLabelText("next turn")).toBeInTheDocument()
    expect(screen.getByLabelText("last turn")).toBeInTheDocument()
    expect(screen.getByText("turn 3 / 10")).toBeInTheDocument()
  })

  it("does not render any play/pause button", () => {
    setup()
    expect(screen.queryByLabelText("play")).toBeNull()
    expect(screen.queryByLabelText("pause")).toBeNull()
  })

  it("invokes the matching handler for each button", async () => {
    const handlers = setup()
    const user = userEvent.setup()
    await user.click(screen.getByLabelText("first turn"))
    await user.click(screen.getByLabelText("previous turn"))
    await user.click(screen.getByLabelText("next turn"))
    await user.click(screen.getByLabelText("last turn"))
    expect(handlers.onFirst).toHaveBeenCalledTimes(1)
    expect(handlers.onPrevious).toHaveBeenCalledTimes(1)
    expect(handlers.onNext).toHaveBeenCalledTimes(1)
    expect(handlers.onLast).toHaveBeenCalledTimes(1)
  })
})

describe("BatchSummaryPanel", () => {
  const batch: BatchResponse = {
    simulations: 10,
    blue_bot: "blue.cpp",
    red_bot: "red.py",
    wins_blue: 6,
    wins_red: 3,
    draws: 1,
    avg_score_blue: 12.34,
    avg_score_red: 8.5,
    avg_turns: 42.7,
    avg_ttfo_blue_ms: 4,
    avg_ttfo_red_ms: 7,
    avg_aot_blue_ms: 11,
    avg_aot_red_ms: 14,
    seed: "12345",
    matches: [],
  }

  it("renders win counts as N (XX.X%)", () => {
    render(<BatchSummaryPanel batch={batch} />)
    expect(screen.getByText("6 (60.0%)")).toBeInTheDocument()
    expect(screen.getByText("3 (30.0%)")).toBeInTheDocument()
    expect(screen.getByText("1 (10.0%)")).toBeInTheDocument()
  })

  it("formats numeric averages", () => {
    render(<BatchSummaryPanel batch={batch} />)
    expect(screen.getByText("12.34")).toBeInTheDocument()
    expect(screen.getByText("42.7")).toBeInTheDocument()
    expect(screen.getByText("4ms/11ms")).toBeInTheDocument()
  })
})
