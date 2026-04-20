import {
	type FrameData,
	parseFrameLines,
	parseSerializeResponse,
	type TraceMatch,
} from "./parser.ts";
import { destroyRenderer, initRenderer, updateFrame } from "./renderer.ts";

// Replay state
let frames: FrameData[] = [];
let myBirdIds: number[] = [];
let currentTurn = 0;

export function render(root: HTMLElement): void {
	root.innerHTML = `
		<div class="app">
			<nav class="app-nav" role="tablist">
				<button class="tab is-active" type="button" data-pane="map" role="tab" aria-selected="true">Game</button>
			</nav>

			<section class="pane-layout" data-pane="map">
				<div class="col-map">
					<div id="game-status" class="map-status" hidden></div>
					<div id="map-container" class="map-container" hidden></div>
					<div id="turn-nav" class="turn-nav" hidden>
						<button class="btn-nav" id="btn-prev" type="button" aria-label="previous turn">&#8249;</button>
						<input type="range" id="turn-slider" min="0" value="0" />
						<button class="btn-nav" id="btn-next" type="button" aria-label="next turn">&#8250;</button>
						<span id="turn-label" class="turn-label"></span>
					</div>
				</div>
				<div class="col-controls">
					<form class="controls" id="game-form">
						<label class="field">
							<span class="field-label">seed</span>
							<input
								type="text"
								name="seed"
								inputmode="numeric"
								autocomplete="off"
								spellcheck="false"
								placeholder="e.g. 100030005000"
							/>
						</label>
						<label class="field">
							<span class="field-label">league</span>
							<select name="league">
								<option value="1">1 — Bronze</option>
								<option value="2">2 — Silver</option>
								<option value="3">3 — Gold</option>
								<option value="4" selected>4 — Legend</option>
							</select>
						</label>
						<label class="field">
							<span class="field-label">p0 bot</span>
							<select name="p0Bot" id="p0-bot-select">
								<option value="" disabled selected>loading…</option>
							</select>
						</label>
						<label class="field">
							<span class="field-label">p1 bot</span>
							<select name="p1Bot" id="p1-bot-select">
								<option value="" disabled selected>loading…</option>
							</select>
						</label>
						<button class="btn-action" type="submit" id="btn-run" aria-label="run">&#9654; Run</button>
					</form>
				</div>
			</section>
		</div>
	`;

	const form = root.querySelector<HTMLFormElement>("#game-form");
	const status = root.querySelector<HTMLDivElement>("#game-status");
	const container = root.querySelector<HTMLDivElement>("#map-container");
	const turnNav = root.querySelector<HTMLDivElement>("#turn-nav");
	const btnRun = root.querySelector<HTMLButtonElement>("#btn-run");
	const btnPrev = root.querySelector<HTMLButtonElement>("#btn-prev");
	const btnNext = root.querySelector<HTMLButtonElement>("#btn-next");
	const slider = root.querySelector<HTMLInputElement>("#turn-slider");
	const turnLabel = root.querySelector<HTMLSpanElement>("#turn-label");
	const p0Select = root.querySelector<HTMLSelectElement>("#p0-bot-select");
	const p1Select = root.querySelector<HTMLSelectElement>("#p1-bot-select");

	if (
		!form ||
		!status ||
		!container ||
		!turnNav ||
		!btnRun ||
		!btnPrev ||
		!btnNext ||
		!slider ||
		!turnLabel ||
		!p0Select ||
		!p1Select
	) {
		throw new Error("layout: missing elements");
	}

	// Load bots on page init
	loadBots(p0Select, p1Select);

	// Navigation helpers
	function updateTurnLabel(): void {
		if (!turnLabel) return;
		turnLabel.textContent = `turn ${currentTurn} / ${frames.length - 1}`;
	}

	function goToTurn(t: number): void {
		if (!slider) return;
		currentTurn = Math.max(0, Math.min(t, frames.length - 1));
		const frame = frames[currentTurn];
		if (frame) updateFrame(frame, myBirdIds);
		slider.value = String(currentTurn);
		updateTurnLabel();
	}

	btnPrev.addEventListener("click", () => goToTurn(currentTurn - 1));
	btnNext.addEventListener("click", () => goToTurn(currentTurn + 1));
	slider.addEventListener("input", () => goToTurn(Number(slider.value)));

	// Keyboard navigation
	document.addEventListener("keydown", (e) => {
		if (frames.length === 0) return;
		if (e.key === "ArrowLeft") {
			e.preventDefault();
			goToTurn(currentTurn - 1);
		} else if (e.key === "ArrowRight") {
			e.preventDefault();
			goToTurn(currentTurn + 1);
		}
	});

	// Run match
	form.addEventListener("submit", async (e) => {
		e.preventDefault();
		const fd = new FormData(form);
		const seedRaw = String(fd.get("seed") ?? "").trim();
		const league = String(fd.get("league") ?? "4");
		const p0Bin = String(fd.get("p0Bot") ?? "");
		const p1Bin = String(fd.get("p1Bot") ?? "");

		if (!p0Bin || !p1Bin) {
			status.textContent = "select bots for both players";
			status.hidden = false;
			return;
		}

		btnRun.disabled = true;
		status.textContent = "running match…";
		status.hidden = false;
		turnNav.hidden = true;
		container.hidden = true;
		destroyRenderer();

		try {
			// Step 1: Run match
			const runBody: Record<string, unknown> = {
				p0Bin,
				p1Bin,
				gameOptions: { league },
			};
			if (seedRaw) {
				runBody.seed = Number(seedRaw);
			}

			const runRes = await fetch("/api/run", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(runBody),
			});
			if (!runRes.ok) {
				const text = await runRes.text();
				status.textContent = `run error ${runRes.status}: ${text}`;
				return;
			}
			const runData = await runRes.json();
			const actualSeed: number = runData.seed;

			status.textContent = "loading replay…";

			// Step 2: Parallel fetch of serialize (walls) + trace (turns)
			const [serRes, traceRes] = await Promise.all([
				fetch(`/api/serialize?seed=${actualSeed}&league=${encodeURIComponent(league)}`),
				fetch("/api/matches/0"),
			]);

			if (!serRes.ok) {
				status.textContent = `serialize error ${serRes.status}: ${await serRes.text()}`;
				return;
			}
			if (!traceRes.ok) {
				status.textContent = `trace error ${traceRes.status}: ${await traceRes.text()} — is --trace-dir set?`;
				return;
			}

			const serText = await serRes.text();
			const traceJson: TraceMatch = await traceRes.json();

			// Step 3: Parse
			const mapData = parseSerializeResponse(serText);
			myBirdIds = mapData.myBirdIds;

			// Build frames: trace turns contain frame data at start of each turn
			frames = traceJson.turns.map((t) => parseFrameLines(t.game_input.p0));

			if (frames.length === 0) {
				status.textContent = "no turns in trace";
				return;
			}

			// Step 4: Init renderer with static map
			await initRenderer(container, mapData);

			// Step 5: Show turn 0
			currentTurn = 0;
			const firstFrame = frames[0];
			if (firstFrame) updateFrame(firstFrame, myBirdIds);

			slider.max = String(frames.length - 1);
			slider.value = "0";
			updateTurnLabel();

			const winnerStr = runData.winner === -1 ? "draw" : `p${runData.winner}`;
			status.textContent = `seed=${actualSeed}  ${mapData.width}x${mapData.height}  winner=${winnerStr}  score=${runData.score_p0}:${runData.score_p1}  turns=${runData.turns}`;
			container.hidden = false;
			turnNav.hidden = false;

			// Update seed field with actual seed used
			const seedInput = form.querySelector<HTMLInputElement>("input[name=seed]");
			if (seedInput && !seedRaw) {
				seedInput.value = String(actualSeed);
			}
		} catch (err) {
			status.textContent = `error: ${String(err)}`;
		} finally {
			btnRun.disabled = false;
		}
	});
}

async function loadBots(p0: HTMLSelectElement, p1: HTMLSelectElement): Promise<void> {
	try {
		const res = await fetch("/api/bots");
		if (!res.ok) throw new Error(`${res.status}`);
		const bots: { name: string; path: string }[] = await res.json();
		if (bots.length === 0) {
			for (const sel of [p0, p1]) {
				sel.innerHTML = '<option value="" disabled selected>no bots found</option>';
			}
			return;
		}
		for (const sel of [p0, p1]) {
			sel.innerHTML = "";
			for (const bot of bots) {
				const opt = document.createElement("option");
				opt.value = bot.path;
				opt.textContent = bot.name;
				sel.appendChild(opt);
			}
		}
	} catch {
		for (const sel of [p0, p1]) {
			sel.innerHTML = '<option value="" disabled selected>failed to load</option>';
		}
	}
}
