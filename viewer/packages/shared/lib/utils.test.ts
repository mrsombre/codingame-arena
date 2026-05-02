import { describe, expect, it } from "vitest"
import { cn } from "./utils.ts"

describe("cn", () => {
  it("joins class names", () => {
    expect(cn("a", "b")).toBe("a b")
  })

  it("drops falsy values", () => {
    expect(cn("a", false, undefined, null, "b")).toBe("a b")
  })

  it("merges conflicting tailwind utilities, last wins", () => {
    expect(cn("p-2", "p-4")).toBe("p-4")
    expect(cn("text-sm text-base")).toBe("text-base")
  })

  it("supports conditional object syntax", () => {
    expect(cn("base", { active: true, hidden: false })).toBe("base active")
  })
})
