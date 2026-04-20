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

  birdId?: number
  playerId?: number
  coord?: { x: number; y: number }
  numberOfCells?: number
  growth?: number
  
}
export interface FrameDataDto {
  events: EventDto[]
  ms: number[]
  messages: MessageDto[]
  marks: CoordDto[][]
}
export interface FrameData extends FrameDataDto {
  previous: FrameData
  frameInfo: FrameInfo

  apples: CoordDto[]
  // indexed by id
  birds: BirdDto[]
}
export interface GlobalBirdDto {
  id: number,
  body: CoordDto[]
}
export interface BirdDto extends GlobalBirdDto {
  owner: number
  bodyAfterMove: CoordDto[]
  // Tracked locally
  alive: boolean
}

export type GlobalDataDto = {
  width: number
  height: number
  map: number[][]
  apples: CoordDto[]
  playerBirds: GlobalBirdDto[][]
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

export interface AppleEffect extends TEffect<PIXI.Container> {
  sparks: PIXI.AnimatedSprite
}
export interface BirdEffect extends TEffect<PIXI.Container> {
  pIdx?: number
  sprite?: PIXI.Sprite
  face?: PIXI.Sprite
  body?: PIXI.Container
  id: PIXI.Text
  faceScaler?: PIXI.Container
  eating?: PIXI.AnimatedSprite
  dying?: PIXI.AnimatedSprite
  transforming?: PIXI.AnimatedSprite
  transformingScaler?: PIXI.Container
  getTooltip: () => string
  hands?: PIXI.Sprite[]
}
export interface BirdAnimatedEffect extends BirdEffect {
}

export interface GlobalData extends GlobalDataDto {
  players: PlayerInfo[]
  playerCount: number
}

export interface MessageDto {
  playerIdx: number
  birdId: number
  text: string
}