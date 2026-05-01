import type { BotEntry } from "@shared/api.ts"
import { useNavigate } from "@tanstack/react-router"
import { type ComponentType, type FormEvent, type ReactNode, useEffect, useState } from "react"
import { ArenaWorkspaceLayout, BackButton, BatchMatchesList, BatchRunPanel, BatchSummaryPanel, LoadingList, ReplaysList, RunMatchPanel, StatusLine } from "./components.tsx"
import { ReplayViewer } from "./ReplayViewer.tsx"
import type { BatchMatch, BatchResponse, GameViewerAdapter, ReplayEntry, ReplayResponse, RunResponse, TraceMatchBase, TraceTurnBase } from "./types.ts"

export interface BatchMatchCacheEntry<TMapData, TTurn extends TraceTurnBase> {
  match: BatchMatch
  mapData: TMapData
  trace: TraceMatchBase<TTurn>
}

export interface SharedGameViews<TMapData, TTurn extends TraceTurnBase> {
  SingleView: ComponentType<{ bots: BotEntry[] }>
  BatchView: ComponentType<{ bots: BotEntry[] }>
  BatchMatchView: ComponentType<{ matchId: string }>
  ReplaysView: ComponentType
  ReplayView: ComponentType<{ replayId: string }>
  ReplayViewer: ComponentType<{ mapData: TMapData; trace: TraceMatchBase<TTurn>; status?: ReactNode; leftSlot?: ReactNode; topRightSlot?: ReactNode }>
  batchMatchCache: Map<string, BatchMatchCacheEntry<TMapData, TTurn>>
}

function serializeUrl(game: string, seed: string, league: string | number | undefined) {
  const leagueQuery = league !== undefined && String(league) !== "" && Number(league) > 0 ? `&league=${encodeURIComponent(String(league))}` : ""
  return `/api/serialize?game=${encodeURIComponent(game)}&seed=${encodeURIComponent(seed)}${leagueQuery}`
}

export function createGameViewerViews<TMapData, TFrame, TTurn extends TraceTurnBase, TMeta = unknown>(adapter: GameViewerAdapter<TMapData, TFrame, TTurn, TMeta>): SharedGameViews<TMapData, TTurn> {
  const defaultLeague = adapter.defaultLeague ?? adapter.leagueOptions.at(-1)?.value ?? "4"
  let lastSingleMatch: { mapData: TMapData; trace: TraceMatchBase<TTurn>; status: ReactNode } | null = null
  let lastBatch: { response: BatchResponse; league: string } | null = null
  const batchMatchCache = new Map<string, BatchMatchCacheEntry<TMapData, TTurn>>()

  function AdapterReplayViewer(props: { mapData: TMapData; trace: TraceMatchBase<TTurn>; status?: ReactNode; leftSlot?: ReactNode; topRightSlot?: ReactNode }) {
    return <ReplayViewer adapter={adapter} {...props} />
  }

  function SingleView({ bots }: { bots: BotEntry[] }) {
    const [seed, setSeed] = useState("")
    const [league, setLeague] = useState(defaultLeague)
    const [blueBot, setBlueBot] = useState(bots[0]?.path ?? "")
    const [redBot, setRedBot] = useState(bots[1]?.path ?? bots[0]?.path ?? "")
    const [status, setStatus] = useState<ReactNode>(lastSingleMatch?.status ?? "")
    const [running, setRunning] = useState(false)
    const [mapData, setMapData] = useState<TMapData | null>(lastSingleMatch?.mapData ?? null)
    const [trace, setTrace] = useState<TraceMatchBase<TTurn> | null>(lastSingleMatch?.trace ?? null)

    const handleSubmit = async (event: FormEvent) => {
      event.preventDefault()
      if (!blueBot || !redBot) {
        setStatus("select bots for both players")
        return
      }

      setRunning(true)
      setStatus("running match...")
      setMapData(null)
      setTrace(null)
      lastSingleMatch = null

      try {
        const runBody: Record<string, unknown> = {
          blueBin: blueBot,
          redBin: redBot,
          game: adapter.game,
          gameOptions: { league },
        }
        const seedTrimmed = seed.trim()
        if (seedTrimmed) runBody.seed = seedTrimmed

        const runRes = await fetch("/api/run", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(runBody),
        })
        if (!runRes.ok) {
          setStatus(`run error ${runRes.status}: ${await runRes.text()}`)
          return
        }

        const runData = (await runRes.json()) as RunResponse
        setStatus("loading replay...")

        const [serRes, traceRes] = await Promise.all([fetch(serializeUrl(adapter.game, runData.seed, league)), fetch("/api/matches/cg-match")])
        if (!serRes.ok) {
          setStatus(`serialize error ${serRes.status}: ${await serRes.text()}`)
          return
        }
        if (!traceRes.ok) {
          setStatus(`trace error ${traceRes.status}: ${await traceRes.text()} - is --trace-dir set?`)
          return
        }

        const map = adapter.parseSerializeResponse(await serRes.text())
        const traceJson = (await traceRes.json()) as TraceMatchBase<TTurn>
        if (traceJson.turns.length === 0) {
          setStatus("no turns in trace")
          return
        }

        const statusLine = adapter.formatRunStatus({ actualSeed: runData.seed, mapData: map, trace: traceJson, run: runData })
        lastSingleMatch = { mapData: map, trace: traceJson, status: statusLine }
        setStatus(statusLine)
        setMapData(map)
        setTrace(traceJson)
      } catch (error) {
        setStatus(`error: ${String(error)}`)
      } finally {
        setRunning(false)
      }
    }

    const form = (
      <RunMatchPanel
        bots={bots}
        leagueOptions={adapter.leagueOptions}
        seed={seed}
        league={league}
        blueBot={blueBot}
        redBot={redBot}
        running={running}
        onSeedChange={setSeed}
        onLeagueChange={setLeague}
        onBlueBotChange={setBlueBot}
        onRedBotChange={setRedBot}
        onSubmit={handleSubmit}
      />
    )

    if (mapData && trace) {
      return <AdapterReplayViewer mapData={mapData} trace={trace} status={status} leftSlot={form} />
    }
    return (
      <ArenaWorkspaceLayout left={form}>
        <StatusLine>{status}</StatusLine>
      </ArenaWorkspaceLayout>
    )
  }

  function BatchView({ bots }: { bots: BotEntry[] }) {
    const navigate = useNavigate()
    const [seed, setSeed] = useState("")
    const [league, setLeague] = useState(lastBatch?.league ?? defaultLeague)
    const [blueBot, setBlueBot] = useState(bots[0]?.path ?? "")
    const [redBot, setRedBot] = useState(bots[1]?.path ?? bots[0]?.path ?? "")
    const [simulations, setSimulations] = useState("50")
    const [maxTurns, setMaxTurns] = useState("200")
    const [status, setStatus] = useState("")
    const [running, setRunning] = useState(false)
    const [batch, setBatch] = useState<BatchResponse | null>(lastBatch?.response ?? null)
    const [loadingMatch, setLoadingMatch] = useState<number | null>(null)

    const handleSubmit = async (event: FormEvent) => {
      event.preventDefault()
      if (!blueBot || !redBot) {
        setStatus("select bots for both players")
        return
      }
      const sims = Number(simulations)
      if (!Number.isFinite(sims) || sims < 1) {
        setStatus("simulations must be >= 1")
        return
      }
      const turns = Number(maxTurns)
      if (!Number.isFinite(turns) || turns < 1) {
        setStatus("turns must be >= 1")
        return
      }

      setRunning(true)
      setStatus(`running batch: ${sims} match${sims === 1 ? "" : "es"}...`)
      setBatch(null)
      lastBatch = null
      batchMatchCache.clear()

      try {
        const body: Record<string, unknown> = {
          blueBin: blueBot,
          redBin: redBot,
          game: adapter.game,
          simulations: sims,
          maxTurns: turns,
          gameOptions: { league },
        }
        const seedTrimmed = seed.trim()
        if (seedTrimmed) body.seed = seedTrimmed

        const response = await fetch("/api/batch", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(body),
        })
        if (!response.ok) {
          setStatus(`batch error ${response.status}: ${await response.text()}`)
          return
        }
        const data = (await response.json()) as BatchResponse
        lastBatch = { response: data, league }
        setBatch(data)
        setStatus("")
      } catch (error) {
        setStatus(`error: ${String(error)}`)
      } finally {
        setRunning(false)
      }
    }

    const openMatch = async (match: BatchMatch) => {
      setLoadingMatch(match.id)
      setStatus(`loading match ${match.id}...`)
      try {
        const [serRes, traceRes] = await Promise.all([fetch(serializeUrl(adapter.game, match.seed, league)), fetch(`/api/matches/${match.id}`)])
        if (!serRes.ok) {
          setStatus(`serialize error ${serRes.status}: ${await serRes.text()}`)
          return
        }
        if (!traceRes.ok) {
          setStatus(`trace error ${traceRes.status}: ${await traceRes.text()}`)
          return
        }
        const map = adapter.parseSerializeResponse(await serRes.text())
        const traceJson = (await traceRes.json()) as TraceMatchBase<TTurn>
        batchMatchCache.set(String(match.id), { match, mapData: map, trace: traceJson })
        setStatus("")
        navigate({ to: "/batch/$matchId", params: { matchId: String(match.id) } })
      } catch (error) {
        setStatus(`error: ${String(error)}`)
      } finally {
        setLoadingMatch(null)
      }
    }

    const form = (
      <BatchRunPanel
        bots={bots}
        leagueOptions={adapter.leagueOptions}
        seed={seed}
        league={league}
        blueBot={blueBot}
        redBot={redBot}
        running={running}
        simulations={simulations}
        maxTurns={maxTurns}
        onSeedChange={setSeed}
        onLeagueChange={setLeague}
        onBlueBotChange={setBlueBot}
        onRedBotChange={setRedBot}
        onSimulationsChange={setSimulations}
        onMaxTurnsChange={setMaxTurns}
        onSubmit={handleSubmit}
      />
    )

    return (
      <ArenaWorkspaceLayout
        left={
          <>
            {form}
            {batch && <BatchSummaryPanel batch={batch} />}
          </>
        }
      >
        <StatusLine>{status}</StatusLine>
        {batch && batch.matches.length > 0 && <BatchMatchesList batch={batch} loadingMatch={loadingMatch} onOpenMatch={openMatch} />}
      </ArenaWorkspaceLayout>
    )
  }

  function ReplaysView() {
    const [list, setList] = useState<ReplayEntry[] | null>(null)
    const [listError, setListError] = useState("")

    useEffect(() => {
      fetch("/api/replays")
        .then((response) => {
          if (!response.ok) throw new Error(`${response.status}`)
          return response.json()
        })
        .then((data: ReplayEntry[]) => setList(data))
        .catch((error) => setListError(String(error)))
    }, [])

    return (
      <ArenaWorkspaceLayout>
        {listError && <StatusLine destructive>{listError}</StatusLine>}
        {list === null && !listError && <LoadingList />}
        {list !== null && list.length === 0 && (
          <p className="font-mono text-xs text-muted-foreground">
            No replays. Use <code>arena replay &lt;url|id&gt;</code> to download.
          </p>
        )}
        {list !== null && list.length > 0 && <ReplaysList list={list} />}
      </ArenaWorkspaceLayout>
    )
  }

  function ReplayView({ replayId }: { replayId: string }) {
    const [status, setStatus] = useState<ReactNode>(`loading replay ${replayId}...`)
    const [mapData, setMapData] = useState<TMapData | null>(null)
    const [trace, setTrace] = useState<TraceMatchBase<TTurn> | null>(null)
    const [replay, setReplay] = useState<ReplayResponse<TTurn> | null>(null)
    const backButton = <BackButton to="/replays" label="Back to list" />

    useEffect(() => {
      let cancelled = false
      setMapData(null)
      setTrace(null)
      setReplay(null)
      setStatus(`loading replay ${replayId}...`)
      ;(async () => {
        try {
          const traceRes = await fetch(`/api/replays/${encodeURIComponent(replayId)}?game=${encodeURIComponent(adapter.game)}`)
          if (!traceRes.ok) {
            if (!cancelled) setStatus(`replay error ${traceRes.status}: ${await traceRes.text()}`)
            return
          }
          const replayJson = (await traceRes.json()) as ReplayResponse<TTurn>
          const serRes = await fetch(serializeUrl(adapter.game, replayJson.seed, replayJson.league))
          if (!serRes.ok) {
            if (!cancelled) setStatus(`serialize error ${serRes.status}: ${await serRes.text()}`)
            return
          }
          const map = adapter.parseSerializeResponse(await serRes.text())
          if (cancelled) return
          setMapData(map)
          setTrace(replayJson)
          setReplay(replayJson)
          setStatus("")
        } catch (error) {
          if (!cancelled) setStatus(`error: ${String(error)}`)
        }
      })()
      return () => {
        cancelled = true
      }
    }, [replayId])

    if (mapData && trace && replay) {
      return <AdapterReplayViewer mapData={mapData} trace={trace} status={adapter.formatReplayStatus({ replayId, mapData, trace, replay })} leftSlot={backButton} />
    }
    return (
      <ArenaWorkspaceLayout left={backButton}>
        <StatusLine>{status}</StatusLine>
      </ArenaWorkspaceLayout>
    )
  }

  function BatchMatchView({ matchId }: { matchId: string }) {
    const entry = batchMatchCache.get(matchId)
    const backButton = <BackButton to="/batch" label="Back to list" />

    if (!entry) {
      return (
        <ArenaWorkspaceLayout left={backButton}>
          <p className="font-mono text-xs text-muted-foreground">Match not found. Run a batch first.</p>
        </ArenaWorkspaceLayout>
      )
    }

    const { match, mapData, trace } = entry
    return <AdapterReplayViewer mapData={mapData} trace={trace} status={adapter.formatBatchMatchStatus({ match, mapData, trace })} leftSlot={backButton} />
  }

  return {
    SingleView,
    BatchView,
    BatchMatchView,
    ReplaysView,
    ReplayView,
    ReplayViewer: AdapterReplayViewer,
    batchMatchCache,
  }
}
