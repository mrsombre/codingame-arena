import { Application, Container, Graphics, Text, type TextStyleOptions } from "pixi.js"
import type { FrameData, MapData } from "./parser.ts"

const CELL_PREFERRED = 24
const CELL_MIN = 8
const GAP = 1

// Computed per initRenderer based on available space
let CELL = CELL_PREFERRED

const COLOR_BG = 0x1a1a2e
const COLOR_WALL = 0x4a4a6a
const COLOR_WALL_TOP = 0x5c5c80
const COLOR_EMPTY = 0x0f0f23
const COLOR_APPLE = 0xffcc00
const COLOR_APPLE_GLOW = 0xffee88
const COLOR_GRID = 0x2a2a3e

const PLAYER_COLORS: { head: number; body: number; outline: number }[] = [
  { head: 0x4fc3f7, body: 0x29719c, outline: 0x1a4f6e },
  { head: 0xef5350, body: 0xa13533, outline: 0x6e1f1f },
]

const LABEL_STYLE: TextStyleOptions = {
  fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, monospace",
  fontWeight: "bold",
  fill: 0xffffff,
  align: "center",
}

// Module-level state
let app: Application | null = null
let appleLayer: Container | null = null
let birdLayer: Container | null = null
let labelLayer: Container | null = null

/**
 * Initialize the PixiJS application, draw static geometry (grid + walls),
 * and create empty dynamic layers (apples, birds). Must be called once
 * before updateFrame.
 */
export async function initRenderer(container: HTMLElement, data: MapData): Promise<void> {
  const { width, height, walls } = data

  // Fit canvas to container width
  const availableWidth = container.clientWidth || 800
  const maxCell = Math.floor((availableWidth - GAP) / width - GAP)
  CELL = Math.max(CELL_MIN, Math.min(CELL_PREFERRED, maxCell))

  const step = CELL + GAP
  const cw = step * width + GAP
  const ch = step * height + GAP

  if (app) {
    app.destroy(true, { children: true })
  }
  app = new Application()
  await app.init({
    width: cw,
    height: ch,
    background: COLOR_BG,
    antialias: true,
    resolution: window.devicePixelRatio || 1,
    autoDensity: true,
  })
  container.innerHTML = ""
  app.canvas.style.maxWidth = "100%"
  app.canvas.style.height = "auto"
  container.appendChild(app.canvas)

  // Grid lines (static)
  const gridLayer = new Container()
  app.stage.addChild(gridLayer)

  const gridLines = new Graphics()
  gridLines.setFillStyle({ color: COLOR_GRID, alpha: 0.4 })
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
        cellGfx.rect(px, py, CELL, 3).fill({ color: COLOR_WALL_TOP, alpha: 0.5 })
      } else {
        cellGfx.rect(px, py, CELL, CELL).fill({ color: COLOR_EMPTY })
      }
    }
  }
  cellLayer.addChild(cellGfx)

  // Dynamic layers (populated by updateFrame)
  appleLayer = new Container()
  app.stage.addChild(appleLayer)

  birdLayer = new Container()
  app.stage.addChild(birdLayer)

  labelLayer = new Container()
  app.stage.addChild(labelLayer)

  // Hover overlay (always on top)
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
}

/**
 * Redraw dynamic content (apples + birds) for a specific turn.
 * Requires initRenderer to have been called first.
 */
export function updateFrame(frame: FrameData, myBirdIds: number[]): void {
  if (!appleLayer || !birdLayer || !labelLayer) return

  const step = CELL + GAP

  // Clear dynamic layers
  appleLayer.removeChildren()
  birdLayer.removeChildren()
  labelLayer.removeChildren()

  // Apples
  for (const a of frame.apples) {
    const cx = a.x * step + GAP + CELL / 2
    const cy = a.y * step + GAP + CELL / 2

    const glow = new Graphics()
    glow.circle(cx, cy, CELL * 0.45).fill({ color: COLOR_APPLE_GLOW, alpha: 0.15 })
    appleLayer.addChild(glow)

    const dot = new Graphics()
    dot.circle(cx, cy, CELL * 0.3).fill({ color: COLOR_APPLE })
    dot.circle(cx - CELL * 0.08, cy - CELL * 0.08, CELL * 0.12).fill({ color: 0xffffff, alpha: 0.4 })
    appleLayer.addChild(dot)
  }

  // Birds
  const myIdSet = new Set(myBirdIds)

  for (const bird of frame.birds) {
    const playerIdx = myIdSet.has(bird.id) ? 0 : 1
    const colors = PLAYER_COLORS[playerIdx] ?? PLAYER_COLORS[0]
    if (!colors) continue

    const birdContainer = new Container()
    birdLayer.addChild(birdContainer)

    // Helper: center of a cell
    const cx = (seg: { x: number; y: number }) => seg.x * step + GAP + CELL / 2
    const cy = (seg: { x: number; y: number }) => seg.y * step + GAP + CELL / 2

    // Draw connecting chains (lines between consecutive segments)
    const chainGfx = new Graphics()
    chainGfx.setStrokeStyle({ width: CELL * 0.2, color: colors.outline, cap: "round" })
    for (let s = 0; s < bird.body.length - 1; s++) {
      const a = bird.body[s]
      const b = bird.body[s + 1]
      if (!a || !b) continue
      chainGfx.moveTo(cx(a), cy(a)).lineTo(cx(b), cy(b)).stroke()
    }
    birdContainer.addChild(chainGfx)

    // Draw segments as circles (tail to head so head is on top)
    for (let s = bird.body.length - 1; s >= 0; s--) {
      const seg = bird.body[s]
      if (!seg) continue
      const isHead = s === 0

      const scx = cx(seg)
      const scy = cy(seg)
      const r = isHead ? CELL * 0.42 : CELL * 0.3

      const segGfx = new Graphics()
      segGfx.circle(scx, scy, r + 1).fill({ color: colors.outline })
      segGfx.circle(scx, scy, r).fill({ color: isHead ? colors.head : colors.body })

      if (isHead) {
        const eyeOff = CELL * 0.17
        const eyeR = CELL * 0.1
        const pupilR = CELL * 0.05
        segGfx.circle(scx - eyeOff, scy, eyeR).fill({ color: 0xffffff })
        segGfx.circle(scx + eyeOff, scy, eyeR).fill({ color: 0xffffff })
        segGfx.circle(scx - eyeOff, scy, pupilR).fill({ color: COLOR_BG })
        segGfx.circle(scx + eyeOff, scy, pupilR).fill({ color: COLOR_BG })
      }

      birdContainer.addChild(segGfx)
    }

    const head = bird.body[0]
    if (head) {
      const label = new Text({
        text: String(bird.id),
        style: { ...LABEL_STYLE, fontSize: CELL * 0.38 },
      })
      label.anchor.set(0.5, 1)
      label.x = head.x * step + GAP + CELL / 2
      label.y = head.y * step + GAP - 1
      labelLayer.addChild(label)
    }
  }
}

/** Convenience wrapper: init + render first frame. */
export async function renderGame(container: HTMLElement, data: MapData): Promise<void> {
  await initRenderer(container, data)
  updateFrame({ apples: data.apples, birds: data.birds }, data.myBirdIds)
}

export function destroyRenderer(): void {
  if (app) {
    app.destroy(true, { children: true })
    app = null
  }
  appleLayer = null
  birdLayer = null
  labelLayer = null
}
