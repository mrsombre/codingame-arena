import { Application, Container, Graphics, Text, type TextStyleOptions } from "pixi.js"
import type { FrameData, MapData, Pac } from "./parser.ts"

const CELL_PREFERRED = 28
const CELL_MIN = 10
const GAP = 1

let CELL = CELL_PREFERRED

const COLOR_BG = 0x0a0a1e
const COLOR_WALL = 0x2044aa
const COLOR_WALL_TOP = 0x3a66dd
const COLOR_FLOOR = 0x07071a
const COLOR_GRID = 0x1a1a2e
const COLOR_PELLET = 0xffd166
const COLOR_CHERRY = 0xef476f
const COLOR_CHERRY_GLOW = 0xffa3b8

const PLAYER_COLORS: { body: number; outline: number }[] = [
  { body: 0x4fc3f7, outline: 0x1a4f6e },
  { body: 0xef5350, outline: 0x6e1f1f },
]

const LABEL_STYLE: TextStyleOptions = {
  fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, monospace",
  fontWeight: "bold",
  fill: 0xffffff,
  align: "center",
}

let app: Application | null = null
let pelletLayer: Container | null = null
let pacLayer: Container | null = null
let labelLayer: Container | null = null
let appToken = 0

let gridWidth = 0
let gridHeight = 0

export async function initRenderer(container: HTMLElement, data: MapData): Promise<{ width: number; height: number }> {
  const myToken = ++appToken

  const { width, height, walls } = data
  gridWidth = width
  gridHeight = height

  const availableWidth = container.clientWidth || 800
  const maxCell = Math.floor((availableWidth - GAP) / width - GAP)
  CELL = Math.max(CELL_MIN, Math.min(CELL_PREFERRED, maxCell))

  const step = CELL + GAP
  const cw = step * width + GAP
  const ch = step * height + GAP

  if (app) {
    const oldApp = app
    app = null
    pelletLayer = null
    pacLayer = null
    labelLayer = null
    oldApp.destroy(true, { children: true })
  }

  const next = new Application()
  await next.init({
    width: cw,
    height: ch,
    background: COLOR_BG,
    antialias: true,
    resolution: window.devicePixelRatio || 1,
    autoDensity: true,
  })

  if (appToken !== myToken) {
    next.destroy(true, { children: true })
    return { width: cw, height: ch }
  }

  app = next
  container.innerHTML = ""
  app.canvas.style.maxWidth = "100%"
  app.canvas.style.height = "auto"
  container.appendChild(app.canvas)

  // Grid lines (static)
  const gridLayer = new Container()
  app.stage.addChild(gridLayer)
  const gridLines = new Graphics()
  gridLines.setFillStyle({ color: COLOR_GRID, alpha: 0.35 })
  for (let x = 0; x <= width; x++) {
    gridLines.rect(x * step, 0, GAP, ch).fill()
  }
  for (let y = 0; y <= height; y++) {
    gridLines.rect(0, y * step, cw, GAP).fill()
  }
  gridLayer.addChild(gridLines)

  // Cells / walls (static)
  const cellLayer = new Container()
  app.stage.addChild(cellLayer)
  const cellGfx = new Graphics()
  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      const isWall = walls[y * width + x]
      const px = x * step + GAP
      const py = y * step + GAP
      if (isWall) {
        cellGfx.rect(px, py, CELL, CELL).fill({ color: COLOR_WALL })
        cellGfx.rect(px, py, CELL, 3).fill({ color: COLOR_WALL_TOP, alpha: 0.6 })
      } else {
        cellGfx.rect(px, py, CELL, CELL).fill({ color: COLOR_FLOOR })
      }
    }
  }
  cellLayer.addChild(cellGfx)

  // Dynamic layers
  pelletLayer = new Container()
  app.stage.addChild(pelletLayer)

  pacLayer = new Container()
  app.stage.addChild(pacLayer)

  labelLayer = new Container()
  app.stage.addChild(labelLayer)

  // Hover overlay
  const coordLabel = new Text({
    text: "",
    style: {
      fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, monospace",
      fontSize: 11,
      fill: 0xcccccc,
      letterSpacing: 0.5,
    },
  })
  coordLabel.visible = false
  app.stage.addChild(coordLabel)

  const highlight = new Graphics()
  highlight.visible = false
  app.stage.addChild(highlight)

  let prevCellX = -1
  let prevCellY = -1

  app.canvas.addEventListener("mousemove", (e: MouseEvent) => {
    if (!app) return
    const rect = app.canvas.getBoundingClientRect()
    const mx = e.clientX - rect.left
    const my = e.clientY - rect.top

    const cellX = Math.floor(mx / step)
    const cellY = Math.floor(my / step)

    if (cellX < 0 || cellX >= width || cellY < 0 || cellY >= height) {
      highlight.visible = false
      coordLabel.visible = false
      prevCellX = -1
      prevCellY = -1
      return
    }
    if (cellX === prevCellX && cellY === prevCellY) return
    prevCellX = cellX
    prevCellY = cellY

    const px = cellX * step + GAP
    const py = cellY * step + GAP
    highlight.clear()
    highlight.rect(px, py, CELL, CELL).fill({ color: 0xffffff, alpha: 0.08 })
    highlight.visible = true

    coordLabel.text = `${cellX},${cellY}`
    coordLabel.x = px + CELL + 4
    coordLabel.y = py + (CELL - coordLabel.height) / 2
    if (coordLabel.x + coordLabel.width > cw) {
      coordLabel.x = px - coordLabel.width - 4
    }
    coordLabel.visible = true
  })

  app.canvas.addEventListener("mouseleave", () => {
    highlight.visible = false
    coordLabel.visible = false
    prevCellX = -1
    prevCellY = -1
  })

  return { width: cw, height: ch }
}

export interface UpdateFrameOptions {
  /** Skip rebuilding the pellet layer. Use during interpolation. */
  skipPellets?: boolean
}

function typeGlyph(type: Pac["type"]): string {
  switch (type) {
    case "ROCK":
      return "R"
    case "PAPER":
      return "P"
    case "SCISSORS":
      return "S"
    case "DEAD":
      return "×"
    default:
      return "?"
  }
}

export function updateFrame(frame: FrameData, options?: UpdateFrameOptions): void {
  if (!pelletLayer || !pacLayer || !labelLayer) return

  const step = CELL + GAP

  pacLayer.removeChildren()
  labelLayer.removeChildren()

  if (!options?.skipPellets) {
    pelletLayer.removeChildren()
    for (const p of frame.pellets) {
      if (p.x < 0 || p.x >= gridWidth || p.y < 0 || p.y >= gridHeight) continue
      const cx = p.x * step + GAP + CELL / 2
      const cy = p.y * step + GAP + CELL / 2
      if (p.value >= 10) {
        const glow = new Graphics()
        glow.circle(cx, cy, CELL * 0.5).fill({ color: COLOR_CHERRY_GLOW, alpha: 0.2 })
        pelletLayer.addChild(glow)
        const dot = new Graphics()
        dot.circle(cx, cy, CELL * 0.32).fill({ color: COLOR_CHERRY })
        dot.circle(cx - CELL * 0.09, cy - CELL * 0.09, CELL * 0.12).fill({ color: 0xffffff, alpha: 0.5 })
        pelletLayer.addChild(dot)
      } else {
        const dot = new Graphics()
        dot.circle(cx, cy, CELL * 0.12).fill({ color: COLOR_PELLET })
        pelletLayer.addChild(dot)
      }
    }
  }

  for (const pac of frame.pacs) {
    const owner = pac.mine ? 0 : 1
    const colors = PLAYER_COLORS[owner] ?? PLAYER_COLORS[0]
    if (!colors) continue
    const dead = pac.type === "DEAD"
    const alpha = dead ? 0.3 : 1

    const scx = pac.x * step + GAP + CELL / 2
    const scy = pac.y * step + GAP + CELL / 2
    const r = CELL * 0.4

    const gfx = new Graphics()
    gfx.circle(scx, scy, r + 1).fill({ color: colors.outline, alpha })
    gfx.circle(scx, scy, r).fill({ color: colors.body, alpha })
    if (pac.abilityDuration > 0 && !dead) {
      // Speed boost ring
      gfx.setStrokeStyle({ width: 2, color: 0xffff99, alpha: 0.9 })
      gfx.circle(scx, scy, r + 3).stroke()
    }
    pacLayer.addChild(gfx)

    // Type glyph inside pac
    const glyph = new Text({
      text: typeGlyph(pac.type),
      style: { ...LABEL_STYLE, fontSize: CELL * 0.42, fill: dead ? 0x888888 : 0xffffff },
    })
    glyph.anchor.set(0.5, 0.5)
    glyph.x = scx
    glyph.y = scy
    pacLayer.addChild(glyph)

    // Id label above
    const label = new Text({
      text: String(pac.id),
      style: { ...LABEL_STYLE, fontSize: CELL * 0.34 },
    })
    label.anchor.set(0.5, 1)
    label.x = scx
    label.y = pac.y * step + GAP - 1
    labelLayer.addChild(label)

    // Cooldown hint below (small)
    if (pac.abilityCooldown > 0 && !dead) {
      const cd = new Text({
        text: `cd${pac.abilityCooldown}`,
        style: {
          fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, monospace",
          fontSize: CELL * 0.26,
          fill: 0xaaaaaa,
        },
      })
      cd.anchor.set(0.5, 0)
      cd.x = scx
      cd.y = pac.y * step + GAP + CELL + 1
      labelLayer.addChild(cd)
    }
  }
}

export async function renderGame(container: HTMLElement, data: MapData): Promise<void> {
  await initRenderer(container, data)
  updateFrame({ myScore: data.myScore, oppScore: data.oppScore, pacs: data.pacs, pellets: data.pellets })
}

export function destroyRenderer(): void {
  appToken++
  if (app) {
    const oldApp = app
    app = null
    oldApp.destroy(true, { children: true })
  }
  pelletLayer = null
  pacLayer = null
  labelLayer = null
}
