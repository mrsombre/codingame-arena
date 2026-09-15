

export type ContainerConsumer = (layer: PIXI.Container) => void

/**
 * Given by the SDK
*/
export interface FrameInfo {
  number: number
  frameDuration: number
  date: number
}
/**
 * Given by the SDK
 */
export interface CanvasInfo {
  width: number
  height: number
  oversampling: number
}
/**
 * Given by the SDK
 */
export interface PlayerInfo {
  name: string
  avatar: PIXI.Texture
  color: number
  index: number
  isMe: boolean
  number: number
  type?: string
}

export interface TrainDto {
  currentTownId: number
  id: number
}

export interface FrameDataDto {
  events: EventDto[]
  messages: string[]
}

export interface CoordDto {
  x: number
  y: number
}

export interface AnimData {
  start: number
  end: number
}

export interface EventDto {
  type: number
  animData: AnimData
  params: number[]
  
  coord?: CoordDto
  cost?: number
  amount?: number
  zoneId?: number
  playerIdx?: number
  score?: number
  fromTownId?: number
  toTownId?: number
  coords?: CoordDto[]
}

export interface PlayerDto {
  score: number
}

export interface Connection {
  fromTownId: number
  toTownId: number
  coords: CoordDto[]
}
export interface ConnectionData extends Connection {
  fadeInP?: number
  fadeOutP?: number
}

export interface ConnectionAnim extends Connection {
  trainEffects: PIXI.Container[]
}

export interface ConnectionState extends Connection {
  startFadeP: number
  endFadeP: number
}
export interface Tile {
  container: PIXI.Container
  sprite: PIXI.Sprite
  baseTint: number
  overlay: PIXI.Sprite
}

export interface FrameData extends FrameDataDto {
  previous: FrameData
  frameInfo: FrameInfo

  tiles: TileDto[]
  zones: ZoneDto[]
  players: PlayerDto[]
  connections: Connection[]
}

export interface GlobalZoneDto {
  coords: CoordDto[]
  id: number
  owner: number
}

export interface ZoneDto extends GlobalZoneDto {
  instability: number
  inked: boolean
}
export interface ZoneState extends ZoneDto {
  disruptP: number
  inkP: number
  prevInstability: number
  prevInked: boolean
}

export interface TownDto {
  id: number
  coord: CoordDto
  desiredConnections: number[]
}

export interface GlobalDataDto {
  width: number
  height: number
  passiveIncome: number
  tiles: GlobalTileDto  []
  zoneMap: Record<number, GlobalZoneDto>
  zones: GlobalZoneDto[]
  towns: TownDto[]
}

export type GlobalTileDto = {
  coord: CoordDto  
  type: number
  zoneId: number
}
export type TileDto = {
  track: number
  history: {p: number, track: number}[]
}

export interface Effect {
  busy: boolean
  display: PIXI.DisplayObject
}
export interface AnimatedEffect extends Effect{
  busy: boolean
  display: PIXI.AnimatedSprite
}
export interface TEffect<T extends PIXI.DisplayObject> extends Effect {
  busy: boolean
  display: T
}
export interface TrackEffect extends TEffect<PIXI.Container> {
  sprite: PIXI.Sprite
  overlay: PIXI.Sprite
  scaler: PIXI.Container
}

export interface GlobalData extends GlobalDataDto {
  players: PlayerInfo[]
  playerCount: number
  townMap: Record<string, TownDto>
}


export interface RailData {
  rotation: number
  texture: string
}