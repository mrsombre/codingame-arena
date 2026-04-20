import { render } from "./layout.ts";

const root = document.getElementById("app");
if (!root) {
	throw new Error("#app element missing");
}
render(root);
