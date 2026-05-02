import { describe, expect, it } from "vitest"
import { winnerFromRanks } from "./types.ts"

describe("winnerFromRanks", () => {
  it("returns -1 for a draw", () => {
    expect(winnerFromRanks([0, 0])).toBe(-1)
  })

  it("returns 0 when left ranks first", () => {
    expect(winnerFromRanks([0, 1])).toBe(0)
  })

  it("returns 1 when right ranks first", () => {
    expect(winnerFromRanks([1, 0])).toBe(1)
  })
})
