import { readdir, rm } from "node:fs/promises";
import { join } from "node:path";

const gamesDir = "games";
const entries = await readdir(gamesDir, { withFileTypes: true });
const games = entries.filter((d) => d.isDirectory()).map((d) => d.name);

if (games.length === 0) {
	console.error(`no games found in ${gamesDir}/`);
	process.exit(1);
}

for (const game of games) {
	const entry = join(gamesDir, game, "index.html");
	const outdir = join("dist", game);
	await rm(outdir, { recursive: true, force: true });
	const result = await Bun.build({
		entrypoints: [entry],
		outdir,
		minify: true,
	});
	if (!result.success) {
		for (const log of result.logs) {
			console.error(log);
		}
		process.exit(1);
	}
	console.log(`built ${game} → ${outdir}`);
}
