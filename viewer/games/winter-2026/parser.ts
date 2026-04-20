/** Parsed output from /api/serialize (global info + first frame). */
export interface MapData {
	myId: number;
	width: number;
	height: number;
	/** Row-major grid: true = wall, false = empty. */
	walls: boolean[];
	birdsPerPlayer: number;
	myBirdIds: number[];
	oppBirdIds: number[];
	apples: Coord[];
	birds: Bird[];
}

export interface Coord {
	x: number;
	y: number;
}

export interface Bird {
	id: number;
	body: Coord[];
}

/** Dynamic per-turn state (apples + birds). */
export interface FrameData {
	apples: Coord[];
	birds: Bird[];
}

/** Trace JSON from GET /api/matches/{id}. */
export interface TraceMatch {
	match_id: number;
	seed: number;
	winner: number;
	scores: [number, number];
	turns: TraceTurn[];
}

export interface TraceTurn {
	turn: number;
	game_input: { p0: string[]; p1: string[] };
	p0_output: string;
	p1_output: string;
}

/**
 * Parse frame lines (apple positions + bird bodies) from a string array.
 * Used both by parseSerializeResponse and for per-turn trace game_input.
 *
 * Format:
 *   <appleCount>
 *   <x> <y>  (per apple)
 *   <birdCount>
 *   <id> <x0,y0>:<x1,y1>:...  (per bird, head first)
 */
export function parseFrameLines(lines: string[]): FrameData {
	let i = 0;
	const next = (): string => {
		const line = lines[i];
		if (i >= lines.length || line === undefined) {
			throw new Error(`unexpected end of frame input at line ${i}`);
		}
		i++;
		return line;
	};

	const appleCount = Number.parseInt(next(), 10);
	const apples: Coord[] = [];
	for (let a = 0; a < appleCount; a++) {
		const parts = next().split(" ");
		apples.push({
			x: Number.parseInt(parts[0] ?? "0", 10),
			y: Number.parseInt(parts[1] ?? "0", 10),
		});
	}

	const birdCount = Number.parseInt(next(), 10);
	const birds: Bird[] = [];
	for (let b = 0; b < birdCount; b++) {
		const line = next();
		const spaceIdx = line.indexOf(" ");
		const id = Number.parseInt(line.slice(0, spaceIdx), 10);
		const segments = line.slice(spaceIdx + 1).split(":");
		const body: Coord[] = segments.map((s) => {
			const parts = s.split(",");
			return {
				x: Number.parseInt(parts[0] ?? "0", 10),
				y: Number.parseInt(parts[1] ?? "0", 10),
			};
		});
		birds.push({ id, body });
	}

	return { apples, birds };
}

/**
 * Parse the plain-text response from `/api/serialize` which concatenates
 * global info and first-frame info.
 *
 * Format (see games/winter2026/engine/serializer.go):
 *
 *   <myId>
 *   <width>
 *   <height>
 *   <row0> ... <rowH-1>      (chars: '.' empty, '#' wall)
 *   <birdsPerPlayer>
 *   <myBird0Id> ... <myBirdNId>
 *   <oppBird0Id> ... <oppBirdNId>
 *   --- frame data ---
 *   <appleCount>
 *   <x> <y>  (per apple)
 *   <birdCount>
 *   <id> <x0,y0>:<x1,y1>:...  (per bird, head first)
 */
export function parseSerializeResponse(text: string): MapData {
	const lines = text.split("\n").filter((l) => l !== "");
	let i = 0;
	const next = (): string => {
		const line = lines[i];
		if (i >= lines.length || line === undefined) {
			throw new Error(`unexpected end of input at line ${i}`);
		}
		i++;
		return line;
	};

	const myId = Number.parseInt(next(), 10);
	const width = Number.parseInt(next(), 10);
	const height = Number.parseInt(next(), 10);

	const walls: boolean[] = new Array(width * height);
	for (let y = 0; y < height; y++) {
		const row = next();
		for (let x = 0; x < width; x++) {
			walls[y * width + x] = row[x] === "#";
		}
	}

	const birdsPerPlayer = Number.parseInt(next(), 10);
	const myBirdIds: number[] = [];
	for (let b = 0; b < birdsPerPlayer; b++) {
		myBirdIds.push(Number.parseInt(next(), 10));
	}
	const oppBirdIds: number[] = [];
	for (let b = 0; b < birdsPerPlayer; b++) {
		oppBirdIds.push(Number.parseInt(next(), 10));
	}

	// Delegate frame parsing to parseFrameLines
	const frame = parseFrameLines(lines.slice(i));

	return { myId, width, height, walls, birdsPerPlayer, myBirdIds, oppBirdIds, ...frame };
}
