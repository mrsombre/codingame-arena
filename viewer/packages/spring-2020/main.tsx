import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import App from "./App.tsx"
import "./style.css"

const root = document.getElementById("app")
if (!root) throw new Error("#app element missing")

createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
