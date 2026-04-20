import { parseSerializeResponse } from "./parser.ts";
import { renderMap } from "./renderer.ts";

export function render(root: HTMLElement): void {
	root.innerHTML = `
		<div class="app">
			<nav class="app-nav" role="tablist">
				<button class="tab is-active" type="button" data-pane="map" role="tab" aria-selected="true">Map</button>
			</nav>

			<section class="pane" data-pane="map">
				<form class="controls" id="map-form">
					<label class="field">
						<span class="field-label">seed</span>
						<input
							type="text"
							name="seed"
							inputmode="numeric"
							autocomplete="off"
							spellcheck="false"
							placeholder="e.g. 100030005000"
							required
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
					<button class="btn-action" type="submit" aria-label="render">&#9654;</button>
				</form>

				<div id="map-status" class="map-status" hidden></div>
				<div id="map-container" class="map-container" hidden></div>
			</section>
		</div>
	`;

	const form = root.querySelector<HTMLFormElement>("#map-form");
	const status = root.querySelector<HTMLDivElement>("#map-status");
	const container = root.querySelector<HTMLDivElement>("#map-container");

	if (!form || !status || !container) {
		throw new Error("layout: missing elements");
	}

	form.addEventListener("submit", async (e) => {
		e.preventDefault();
		const fd = new FormData(form);
		const seed = String(fd.get("seed") ?? "").trim();
		if (!seed) return;
		const league = String(fd.get("league") ?? "4");

		status.textContent = "loading...";
		status.hidden = false;
		container.hidden = true;

		try {
			const res = await fetch(
				`/api/serialize?seed=${encodeURIComponent(seed)}&league=${encodeURIComponent(league)}`,
			);
			const text = await res.text();

			if (!res.ok) {
				status.textContent = `error ${res.status}: ${text}`;
				return;
			}

			const data = parseSerializeResponse(text);
			await renderMap(container, data);

			status.textContent = `seed=${seed}  ${data.width}x${data.height}  apples=${data.apples.length}  birds=${data.birds.length}`;
			container.hidden = false;
		} catch (err) {
			status.textContent = `error: ${String(err)}`;
		}
	});
}
