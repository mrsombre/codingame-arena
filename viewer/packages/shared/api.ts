export interface BotEntry {
  name: string
  path: string
}

/** Fetch available bots from the arena server. */
export async function fetchBots(): Promise<BotEntry[]> {
  const res = await fetch("/api/bots")
  if (!res.ok) throw new Error(`${res.status}`)
  return res.json()
}
