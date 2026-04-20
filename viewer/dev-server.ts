import index from "./games/winter-2026/index.html";

const apiTarget = process.env.ARENA_API ?? "http://localhost:5757";
const port = Number(process.env.PORT ?? 3000);

const server = Bun.serve({
	port,
	development: { hmr: true, console: true },
	routes: {
		"/": Response.redirect("/winter-2026/", 302),
		"/winter-2026/": index,
	},
	async fetch(req) {
		const url = new URL(req.url);
		if (url.pathname.startsWith("/api/") || url.pathname === "/healthz") {
			const upstream = apiTarget + url.pathname + url.search;
			try {
				return await fetch(upstream, {
					method: req.method,
					headers: req.headers,
					body: req.method === "GET" || req.method === "HEAD" ? undefined : req.body,
				});
			} catch (err) {
				return new Response(
					JSON.stringify({
						error: `proxy to ${apiTarget} failed — is arena front running? (${String(err)})`,
					}),
					{ status: 502, headers: { "content-type": "application/json" } },
				);
			}
		}
		return new Response("not found", { status: 404 });
	},
});

console.log(`viewer dev → ${server.url}`);
console.log(`  proxy /api/*, /healthz → ${apiTarget}`);
