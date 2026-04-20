import { Application, Container, Graphics, Text, type TextStyleOptions } from "pixi.js";
import type { MapData } from "./parser.ts";

const CELL = 24;
const GAP = 1;

const COLOR_BG = 0x1a1a2e;
const COLOR_WALL = 0x4a4a6a;
const COLOR_WALL_TOP = 0x5c5c80;
const COLOR_EMPTY = 0x0f0f23;
const COLOR_APPLE = 0xffcc00;
const COLOR_APPLE_GLOW = 0xffee88;
const COLOR_GRID = 0x2a2a3e;

const PLAYER_COLORS: { head: number; body: number; outline: number }[] = [
	{ head: 0x4fc3f7, body: 0x29719c, outline: 0x1a4f6e },
	{ head: 0xef5350, body: 0xa13533, outline: 0x6e1f1f },
];

const LABEL_STYLE: TextStyleOptions = {
	fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, monospace",
	fontWeight: "bold",
	fill: 0xffffff,
	align: "center",
};

let app: Application | null = null;

export async function renderGame(container: HTMLElement, data: MapData): Promise<void> {
	const { width, height, walls, apples, birds, myBirdIds } = data;

	const step = CELL + GAP;
	const cw = step * width + GAP;
	const ch = step * height + GAP;

	// Reuse or create application
	if (app) {
		app.destroy(true, { children: true });
	}
	app = new Application();
	await app.init({
		width: cw,
		height: ch,
		background: COLOR_BG,
		antialias: true,
		resolution: window.devicePixelRatio || 1,
		autoDensity: true,
	});
	container.innerHTML = "";
	container.appendChild(app.canvas);

	// Grid layer
	const gridLayer = new Container();
	app.stage.addChild(gridLayer);

	const gridLines = new Graphics();
	gridLines.setFillStyle({ color: COLOR_GRID, alpha: 0.4 });
	for (let x = 0; x <= width; x++) {
		gridLines.rect(x * step, 0, GAP, ch).fill();
	}
	for (let y = 0; y <= height; y++) {
		gridLines.rect(0, y * step, cw, GAP).fill();
	}
	gridLayer.addChild(gridLines);

	// Cells layer
	const cellLayer = new Container();
	app.stage.addChild(cellLayer);

	const cellGfx = new Graphics();
	for (let y = 0; y < height; y++) {
		for (let x = 0; x < width; x++) {
			const isWall = walls[y * width + x];
			const px = x * step + GAP;
			const py = y * step + GAP;

			if (isWall) {
				// Wall with top highlight for depth
				cellGfx.rect(px, py, CELL, CELL).fill({ color: COLOR_WALL });
				cellGfx.rect(px, py, CELL, 3).fill({ color: COLOR_WALL_TOP, alpha: 0.5 });
			} else {
				cellGfx.rect(px, py, CELL, CELL).fill({ color: COLOR_EMPTY });
			}
		}
	}
	cellLayer.addChild(cellGfx);

	// Apples layer
	const appleLayer = new Container();
	app.stage.addChild(appleLayer);

	for (const a of apples) {
		const cx = a.x * step + GAP + CELL / 2;
		const cy = a.y * step + GAP + CELL / 2;

		const glow = new Graphics();
		glow.circle(cx, cy, CELL * 0.45).fill({ color: COLOR_APPLE_GLOW, alpha: 0.15 });
		appleLayer.addChild(glow);

		const dot = new Graphics();
		dot.circle(cx, cy, CELL * 0.3).fill({ color: COLOR_APPLE });
		// Small shine highlight
		dot.circle(cx - 2, cy - 2, CELL * 0.12).fill({ color: 0xffffff, alpha: 0.4 });
		appleLayer.addChild(dot);
	}

	// Birds layer
	const birdLayer = new Container();
	app.stage.addChild(birdLayer);

	const myIdSet = new Set(myBirdIds);

	for (const bird of birds) {
		const playerIdx = myIdSet.has(bird.id) ? 0 : 1;
		const colors = PLAYER_COLORS[playerIdx] ?? PLAYER_COLORS[0];
		if (!colors) continue;

		const birdContainer = new Container();
		birdLayer.addChild(birdContainer);

		// Draw body segments (tail to head so head is on top)
		for (let s = bird.body.length - 1; s >= 0; s--) {
			const seg = bird.body[s];
			if (!seg) continue;
			const isHead = s === 0;

			const px = seg.x * step + GAP;
			const py = seg.y * step + GAP;
			const pad = isHead ? 1 : 3;

			const segGfx = new Graphics();

			// Outline
			segGfx
				.roundRect(
					px + pad - 1,
					py + pad - 1,
					CELL - pad * 2 + 2,
					CELL - pad * 2 + 2,
					isHead ? 4 : 2,
				)
				.fill({ color: colors.outline });

			// Fill
			segGfx
				.roundRect(px + pad, py + pad, CELL - pad * 2, CELL - pad * 2, isHead ? 3 : 2)
				.fill({ color: isHead ? colors.head : colors.body });

			if (isHead) {
				// Eyes
				const ecx = seg.x * step + GAP + CELL / 2;
				const ecy = seg.y * step + GAP + CELL / 2 + 1;
				segGfx.circle(ecx - 4, ecy, 2.5).fill({ color: 0xffffff });
				segGfx.circle(ecx + 4, ecy, 2.5).fill({ color: 0xffffff });
				segGfx.circle(ecx - 4, ecy, 1.2).fill({ color: COLOR_BG });
				segGfx.circle(ecx + 4, ecy, 1.2).fill({ color: COLOR_BG });
			}

			birdContainer.addChild(segGfx);
		}

		// Bird ID label above head
		const head = bird.body[0];
		if (head) {
			const label = new Text({
				text: String(bird.id),
				style: { ...LABEL_STYLE, fontSize: CELL * 0.38 },
			});
			label.anchor.set(0.5, 1);
			label.x = head.x * step + GAP + CELL / 2;
			label.y = head.y * step + GAP - 1;
			birdContainer.addChild(label);
		}
	}

	// Hover coordinate label
	const coordLabel = new Text({
		text: "",
		style: {
			fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, monospace",
			fontSize: 11,
			fill: 0xcccccc,
			letterSpacing: 0.5,
		},
	});
	coordLabel.visible = false;
	app.stage.addChild(coordLabel);

	const highlight = new Graphics();
	highlight.visible = false;
	app.stage.addChild(highlight);

	let prevCellX = -1;
	let prevCellY = -1;

	app.canvas.addEventListener("mousemove", (e: MouseEvent) => {
		if (!app) return;
		const rect = app.canvas.getBoundingClientRect();
		const mx = e.clientX - rect.left;
		const my = e.clientY - rect.top;

		const cellX = Math.floor(mx / step);
		const cellY = Math.floor(my / step);

		if (cellX < 0 || cellX >= width || cellY < 0 || cellY >= height) {
			highlight.visible = false;
			coordLabel.visible = false;
			prevCellX = -1;
			prevCellY = -1;
			return;
		}

		if (cellX === prevCellX && cellY === prevCellY) return;
		prevCellX = cellX;
		prevCellY = cellY;

		// Cell highlight
		const px = cellX * step + GAP;
		const py = cellY * step + GAP;
		highlight.clear();
		highlight.rect(px, py, CELL, CELL).fill({ color: 0xffffff, alpha: 0.08 });
		highlight.visible = true;

		// Coord label to the right of the cell
		coordLabel.text = `${cellX},${cellY}`;
		coordLabel.x = px + CELL + 4;
		coordLabel.y = py + (CELL - coordLabel.height) / 2;

		// If label overflows right edge, show to the left instead
		if (coordLabel.x + coordLabel.width > cw) {
			coordLabel.x = px - coordLabel.width - 4;
		}
		coordLabel.visible = true;
	});

	app.canvas.addEventListener("mouseleave", () => {
		highlight.visible = false;
		coordLabel.visible = false;
		prevCellX = -1;
		prevCellY = -1;
	});
}

export function destroyRenderer(): void {
	if (app) {
		app.destroy(true, { children: true });
		app = null;
	}
}
