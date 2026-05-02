import type { BotEntry } from "@shared/api.ts"
import { Button } from "@shared/components/ui/button.tsx"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@shared/components/ui/card.tsx"
import { Input } from "@shared/components/ui/input.tsx"
import { Label } from "@shared/components/ui/label.tsx"
import { ScrollArea } from "@shared/components/ui/scroll-area.tsx"
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "@shared/components/ui/select.tsx"
import { Skeleton } from "@shared/components/ui/skeleton.tsx"
import { Slider } from "@shared/components/ui/slider.tsx"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@shared/components/ui/table.tsx"
import { Tooltip, TooltipContent, TooltipTrigger } from "@shared/components/ui/tooltip.tsx"
import { cn } from "@shared/lib/utils.ts"
import { Link } from "@tanstack/react-router"
import { ArrowLeftIcon, ChevronLeftIcon, ChevronRightIcon, ChevronsLeftIcon, ChevronsRightIcon, LoaderIcon, PlayIcon } from "lucide-react"
import type { FormEvent, ReactNode, RefObject } from "react"
import type { BatchMatch, BatchResponse, LeagueOption, ReplayEntry } from "./types.ts"

export function ArenaWorkspaceLayout({ left, children, className }: { left?: ReactNode; children: ReactNode; className?: string }) {
  return (
    <div className={cn("grid gap-6 xl:grid-cols-[20rem_minmax(0,1fr)]", className)}>
      <aside className="flex min-w-0 flex-col gap-4 xl:w-80">{left}</aside>
      <main className="min-w-0 overflow-hidden">{children}</main>
    </div>
  )
}

function LeagueSelect({ value, onChange, options }: { value: string; onChange: (value: string) => void; options: LeagueOption[] }) {
  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger className="w-full">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {options.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  )
}

function BotSelect({ value, onChange, bots }: { value: string; onChange: (value: string) => void; bots: BotEntry[] }) {
  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger className="w-full" size="sm">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {bots.map((bot) => (
            <SelectItem key={bot.path} value={bot.path}>
              {bot.name}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  )
}

interface MatchPanelProps {
  bots: BotEntry[]
  leagueOptions: LeagueOption[]
  seed: string
  league: string
  blueBot: string
  redBot: string
  running: boolean
  onSeedChange: (value: string) => void
  onLeagueChange: (value: string) => void
  onBlueBotChange: (value: string) => void
  onRedBotChange: (value: string) => void
  onSubmit: (event: FormEvent) => void
}

export function RunMatchPanel(props: MatchPanelProps) {
  return (
    <Card size="sm">
      <CardContent>
        <form id="play-form" className="flex flex-col gap-4" onSubmit={props.onSubmit}>
          <div className="grid grid-cols-[minmax(0,1fr)_7rem] gap-4">
            <div className="flex min-w-0 flex-col gap-1.5">
              <Label htmlFor="play-seed">Seed</Label>
              <Input id="play-seed" inputMode="numeric" autoComplete="off" spellCheck={false} placeholder="random" value={props.seed} onChange={(event) => props.onSeedChange(event.target.value)} />
            </div>
            <div className="flex min-w-0 flex-col gap-1.5">
              <Label>League</Label>
              <LeagueSelect value={props.league} onChange={props.onLeagueChange} options={props.leagueOptions} />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="flex min-w-0 flex-col gap-1.5">
              <Label>Blue</Label>
              <BotSelect value={props.blueBot} onChange={props.onBlueBotChange} bots={props.bots} />
            </div>
            <div className="flex min-w-0 flex-col gap-1.5">
              <Label>Red</Label>
              <BotSelect value={props.redBot} onChange={props.onRedBotChange} bots={props.bots} />
            </div>
          </div>
        </form>
      </CardContent>
      <CardFooter className="border-t">
        <Button type="submit" form="play-form" className="w-full" disabled={props.running}>
          {props.running ? <LoaderIcon data-icon="inline-start" className="animate-spin" /> : <PlayIcon data-icon="inline-start" />}
          {props.running ? "Running..." : "Run Match"}
        </Button>
      </CardFooter>
    </Card>
  )
}

interface BatchRunPanelProps extends MatchPanelProps {
  simulations: string
  maxTurns: string
  onSimulationsChange: (value: string) => void
  onMaxTurnsChange: (value: string) => void
}

export function BatchRunPanel(props: BatchRunPanelProps) {
  return (
    <Card size="sm">
      <CardContent>
        <form id="batch-form" className="flex flex-col gap-4" onSubmit={props.onSubmit}>
          <div className="grid grid-cols-[minmax(0,1fr)_7rem] gap-4">
            <div className="flex min-w-0 flex-col gap-1.5">
              <Label htmlFor="batch-seed">Start seed</Label>
              <Input id="batch-seed" inputMode="numeric" autoComplete="off" spellCheck={false} placeholder="random" value={props.seed} onChange={(event) => props.onSeedChange(event.target.value)} />
            </div>
            <div className="flex min-w-0 flex-col gap-1.5">
              <Label>League</Label>
              <LeagueSelect value={props.league} onChange={props.onLeagueChange} options={props.leagueOptions} />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="flex min-w-0 flex-col gap-1.5">
              <Label>Blue</Label>
              <BotSelect value={props.blueBot} onChange={props.onBlueBotChange} bots={props.bots} />
            </div>
            <div className="flex min-w-0 flex-col gap-1.5">
              <Label>Red</Label>
              <BotSelect value={props.redBot} onChange={props.onRedBotChange} bots={props.bots} />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="flex min-w-0 flex-col gap-1.5">
              <Label htmlFor="batch-sims">Matches</Label>
              <Input id="batch-sims" inputMode="numeric" autoComplete="off" spellCheck={false} value={props.simulations} onChange={(event) => props.onSimulationsChange(event.target.value)} />
            </div>
            <div className="flex min-w-0 flex-col gap-1.5">
              <Label htmlFor="batch-turns">Turns</Label>
              <Input id="batch-turns" inputMode="numeric" autoComplete="off" spellCheck={false} value={props.maxTurns} onChange={(event) => props.onMaxTurnsChange(event.target.value)} />
            </div>
          </div>
        </form>
      </CardContent>
      <CardFooter className="border-t">
        <Button type="submit" form="batch-form" className="w-full" disabled={props.running}>
          {props.running ? <LoaderIcon data-icon="inline-start" className="animate-spin" /> : <PlayIcon data-icon="inline-start" />}
          {props.running ? "Running..." : "Run Batch"}
        </Button>
      </CardFooter>
    </Card>
  )
}

export function BatchSummaryPanel({ batch }: { batch: BatchResponse }) {
  const n = batch.simulations || 1
  const winPct = (value: number) => `${((value / n) * 100).toFixed(1)}%`

  return (
    <Card size="sm">
      <CardHeader>
        <CardTitle className="text-xs">Summary</CardTitle>
      </CardHeader>
      <CardContent>
        <dl className="grid grid-cols-2 gap-x-4 gap-y-1 font-mono text-xs text-muted-foreground">
          <dt>Matches</dt>
          <dd>{batch.simulations}</dd>
          <dt className="text-sky-400">{batch.blue_bot} wins</dt>
          <dd>
            {batch.wins_blue} ({winPct(batch.wins_blue)})
          </dd>
          <dt className="text-red-400">{batch.red_bot} wins</dt>
          <dd>
            {batch.wins_red} ({winPct(batch.wins_red)})
          </dd>
          <dt>Draws</dt>
          <dd>
            {batch.draws} ({winPct(batch.draws)})
          </dd>
          <dt className="text-sky-400">{batch.blue_bot} avg</dt>
          <dd>{batch.avg_score_blue.toFixed(2)}</dd>
          <dt className="text-red-400">{batch.red_bot} avg</dt>
          <dd>{batch.avg_score_red.toFixed(2)}</dd>
          <dt>Avg turns</dt>
          <dd>{batch.avg_turns.toFixed(1)}</dd>
          <dt className="text-sky-400">blue ttfo/aot</dt>
          <dd>
            {batch.avg_ttfo_blue_ms.toFixed(0)}ms/{batch.avg_aot_blue_ms.toFixed(0)}ms
          </dd>
          <dt className="text-red-400">red ttfo/aot</dt>
          <dd>
            {batch.avg_ttfo_red_ms.toFixed(0)}ms/{batch.avg_aot_red_ms.toFixed(0)}ms
          </dd>
          <dt>Seed</dt>
          <dd className="truncate" title={String(batch.seed)}>
            {batch.seed}
          </dd>
        </dl>
      </CardContent>
    </Card>
  )
}

export function StatusLine({ children, destructive = false }: { children: ReactNode; destructive?: boolean }) {
  if (!children) return null
  return <p className={cn("mb-3 min-w-0 truncate font-mono text-xs", destructive ? "text-destructive" : "text-muted-foreground")}>{children}</p>
}

export function ReplayStage({ status, canvasRef, ready, controls, maxWidth, topRight }: { status?: ReactNode; canvasRef: RefObject<HTMLDivElement | null>; ready: boolean; controls: ReactNode; maxWidth?: number; topRight?: ReactNode }) {
  return (
    <div className="min-w-0">
      <div className="mb-3 flex min-h-9 items-center justify-between gap-3" style={maxWidth ? { maxWidth } : undefined}>
        <StatusLine>{status}</StatusLine>
        {topRight}
      </div>
      <div ref={canvasRef} className="overflow-auto [&_canvas]:block [&_canvas]:rounded-sm" />
      {ready && controls}
    </div>
  )
}

export function TurnLogPanel({ turnLabel, score, marker, children, emptyLabel }: { turnLabel: ReactNode; score: [number, number]; marker?: ReactNode; children: ReactNode; emptyLabel: ReactNode }) {
  return (
    <Card size="sm">
      <CardHeader>
        <CardTitle className="flex items-center justify-between gap-3 text-xs">
          <span className="flex min-w-0 items-center gap-1 truncate">
            {turnLabel}
            {marker}
          </span>
          <span className="shrink-0 font-mono">
            <span className="text-sky-400">{score[0]}</span>
            <span className="mx-0.5 text-muted-foreground">:</span>
            <span className="text-red-400">{score[1]}</span>
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ScrollArea className="max-h-72">
          <div className="font-mono text-xs text-muted-foreground">{children || <p>{emptyLabel}</p>}</div>
        </ScrollArea>
      </CardContent>
    </Card>
  )
}

export function PlaybackControls({
  sliderMax,
  sliderValue,
  turnLabel,
  onFirst,
  onPrevious,
  onNext,
  onLast,
  onSliderChange,
}: {
  sliderMax: number
  sliderValue: number
  turnLabel: ReactNode
  onFirst: () => void
  onPrevious: () => void
  onNext: () => void
  onLast: () => void
  onSliderChange: (value: number) => void
}) {
  const iconButton = (label: string, onClick: () => void, icon: ReactNode) => (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button variant="outline" size="icon-sm" onClick={onClick} aria-label={label}>
          {icon}
        </Button>
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  )

  return (
    <div className="mt-3 flex max-w-3xl items-center gap-3">
      {iconButton("first turn", onFirst, <ChevronsLeftIcon />)}
      {iconButton("previous turn", onPrevious, <ChevronLeftIcon />)}
      <Slider
        className="min-w-28 flex-1"
        min={0}
        max={sliderMax}
        value={[sliderValue]}
        onValueChange={([value]) => {
          if (value !== undefined) onSliderChange(value)
        }}
      />
      {iconButton("next turn", onNext, <ChevronRightIcon />)}
      {iconButton("last turn", onLast, <ChevronsRightIcon />)}
      <span className="hidden shrink-0 font-mono text-xs text-muted-foreground sm:inline">{turnLabel}</span>
    </div>
  )
}

export function BackButton({ to, label }: { to: "/batch" | "/replays"; label: string }) {
  return (
    <Button asChild variant="outline" size="sm" className="self-start">
      <Link to={to}>
        <ArrowLeftIcon data-icon="inline-start" /> {label}
      </Link>
    </Button>
  )
}

export function LoadingList() {
  return (
    <div className="rounded-sm border p-3">
      <div className="flex flex-col gap-2">
        <Skeleton className="h-8 w-full" />
        <Skeleton className="h-8 w-11/12" />
        <Skeleton className="h-8 w-10/12" />
      </div>
    </div>
  )
}

export function ReplaysList({ list }: { list: ReplayEntry[] }) {
  return (
    <div className="overflow-hidden rounded-sm border">
      <Table className="font-mono text-xs">
        <TableHeader className="bg-muted">
          <TableRow>
            <TableHead>Replay ID</TableHead>
            <TableHead>Players</TableHead>
            <TableHead>Winner</TableHead>
            <TableHead>Score</TableHead>
            <TableHead>Modified</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {list.map((replay) => {
            const winnerLabel = replay.winner === 0 ? (replay.left_name ?? "left") : replay.winner === 1 ? (replay.right_name ?? "right") : "draw"
            const winnerClass = replay.winner === 0 ? "text-sky-400" : replay.winner === 1 ? "text-red-400" : "text-muted-foreground"
            return (
              <TableRow key={replay.id}>
                <TableCell>{replay.id}</TableCell>
                <TableCell>
                  <span className="text-sky-400">{replay.left_name ?? "left"}</span>
                  <span className="text-muted-foreground"> vs </span>
                  <span className="text-red-400">{replay.right_name ?? "right"}</span>
                </TableCell>
                <TableCell className={winnerClass}>{winnerLabel}</TableCell>
                <TableCell>
                  <span className="text-sky-400">{replay.score_left}</span>
                  <span className="text-muted-foreground">:</span>
                  <span className="text-red-400">{replay.score_right}</span>
                </TableCell>
                <TableCell className="text-muted-foreground">{new Date(replay.mtime).toLocaleString()}</TableCell>
                <TableCell className="text-right">
                  <Button asChild variant="outline" size="sm">
                    <Link to="/replays/$replayId" params={{ replayId: replay.id }}>
                      Replay
                    </Link>
                  </Button>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}

export function BatchMatchesList({ batch, loadingMatch, onOpenMatch }: { batch: BatchResponse; loadingMatch: number | null; onOpenMatch: (match: BatchMatch) => void }) {
  const userColor = (botName: string) => {
    if (botName === batch.blue_bot) return "text-sky-400"
    if (botName === batch.red_bot) return "text-red-400"
    return ""
  }

  return (
    <div className="overflow-hidden rounded-sm border">
      <Table className="font-mono text-xs">
        <TableHeader className="bg-muted">
          <TableRow>
            <TableHead>#</TableHead>
            <TableHead>Seed</TableHead>
            <TableHead>Sides</TableHead>
            <TableHead>Winner</TableHead>
            <TableHead>Score</TableHead>
            <TableHead>Turns</TableHead>
            <TableHead>Timing</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {batch.matches.map((match) => {
            const winnerLabel = match.winner === -1 ? "draw" : match.winner === 0 ? match.left_bot : match.right_bot
            const winnerClass = match.winner === 0 ? userColor(match.left_bot) : match.winner === 1 ? userColor(match.right_bot) : "text-muted-foreground"
            return (
              <TableRow key={match.id}>
                <TableCell>{match.id}</TableCell>
                <TableCell className="text-muted-foreground">{match.seed}</TableCell>
                <TableCell>
                  <span className={userColor(match.left_bot)}>{match.left_bot}</span>
                  <span className="text-muted-foreground"> vs </span>
                  <span className={userColor(match.right_bot)}>{match.right_bot}</span>
                </TableCell>
                <TableCell className={winnerClass}>{winnerLabel}</TableCell>
                <TableCell>
                  {match.score_left}:{match.score_right}
                </TableCell>
                <TableCell className="text-muted-foreground">{match.turns}</TableCell>
                <TableCell className="text-muted-foreground">
                  {match.ttfo_left_ms.toFixed(0)}/{match.aot_left_ms.toFixed(0)} vs {match.ttfo_right_ms.toFixed(0)}/{match.aot_right_ms.toFixed(0)}ms
                </TableCell>
                <TableCell className="text-right">
                  <Button variant="outline" size="sm" disabled={loadingMatch !== null} onClick={() => onOpenMatch(match)}>
                    {loadingMatch === match.id ? <LoaderIcon data-icon="inline-start" className="animate-spin" /> : null}
                    Replay
                  </Button>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
