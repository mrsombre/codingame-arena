import { HEIGHT, WIDTH } from '../core/constants.js'
import { flagForDestructionOnReinit, getRenderer } from '../core/rendering.js'
import { bell, easeIn } from '../core/transitions.js'
import {
  fitAspectRatio,
  lerp,
  lerpColor,
  lerpAngle,
  lerpPosition,
  unlerp,
  unlerpUnclamped,
} from '../core/utils.js'
import {
  AnimatedEffect,
  AnimData,
  CanvasInfo,
  Connection,
  ConnectionAnim,
  ConnectionData,
  ContainerConsumer,
  CoordDto,
  Effect,
  FrameData,
  FrameInfo,
  GlobalData,
  PlayerDto,
  PlayerInfo,
  TEffect,
  Tile,
  TrackEffect,
  ZoneDto,
  ZoneState,
} from '../types.js'
import { parseData, parseGlobalData } from './Deserializer.js'
import { TooltipManager } from './TooltipManager.js'
import ev from './events.js'
import { computeRotationAngle, normalizeAngle, rotateAround } from './trigo.js'
import {
  angleDiff,
  choice,
  fit,
  keyOf,
  last,
  setAnimationProgress,
} from './utils.js'
import {
  AVATAR_RECT,
  CONVEYOR_HEIGHT,
  CONVEYOR_SCALE,
  CONVEYOR_WIDTH,
  GAME_ZONE_RECT,
  GRASS_FRAMES,
  BRUSH_FRAMES,
  HUD_COLORS,
  MESSAGE_RECT,
  MOUNTAIN_FRAMES_1x1,
  MOUNTAIN_FRAMES_2x2,
  MOUNTAIN_FRAMES_2x1,
  NAME_RECT,
  POI_FRAMES,
  RAIL_DATA,
  SCORE_RECT,
  TOWN_FRAME,
  TRAIN_FRAME,
  ZONE_LINE_WIDTH,
  INK_FRAMES,
} from './assetConstants.js'
import { TYPE_GRASS, TYPE_MOUNTAIN, TYPE_POI, TYPE_WATER } from './gameConstants.js'

const BASE_OVERLAY_ALPHA = 0.3
const ZONE_GRADIENT_LENGTH = 12
const COLOR_NEUTRAL = 0xffffff
const TRAIN_COLOR = 0xff7f00

interface EffectPool {
  [key: string]: Effect[];
}

const api = {
  setDebugMode: (value: boolean) => {
    api.options.debugMode = value
  },
  renderUpdate: () => {},
  options: {
    debugMode: false,
    territories: 1,
    showOthersMessages: true,
    showMyMessages: true,
    meInGame: false,
    connectionDisplay: 0,
    trackConstrast: 0
  },
}
export { api }

export class ViewModule {
  states: FrameData[];
  globalData: GlobalData;
  pool: EffectPool;
  api: any;
  playerSpeed: number;
  previousData: FrameData;
  currentData: FrameData;
  progress: number;
  oversampling: number;
  container: PIXI.Container;
  time: number;
  canvasData: CanvasInfo;
  gameZone: PIXI.Container;
  tileSize: number;

  tooltipManager: TooltipManager;
  tiles: Tile[][];
  trainLayer: PIXI.Container;
  trackLayer: PIXI.Container;
  borderLayer: PIXI.Container;
  brushLayer: PIXI.Container;
  messageLayer: PIXI.Container;
  hoverTownId?: number;
  huds: any[];
  liveConnectionsToAnimate: ConnectionAnim[];
  messages: {text: PIXI.Text, frame: PIXI.Sprite}[]
  conveyors: PIXI.TilingSprite[]
  conveyorLayer: PIXI.Container
  waterMap: Record<string, PIXI.Sprite>
  inkMask: PIXI.Graphics
  inkAnim: PIXI.Graphics
  regionMask: PIXI.Graphics
  inkAnimTexture: PIXI.RenderTexture
  inkAnimLayer: PIXI.Container
  trainTime: number
  bumpingTiles: Tile[]

  constructor() {
    this.states = []
    this.pool = {}
    this.time = 0
    this.trainTime = 0
    window.debug = this
    this.tooltipManager = new TooltipManager()
    this.api = api
    this.api.setDebugMode = (value: boolean) => {
      //hack for hiding ranking
      this.api.options.debugMode = value
      this.container.parent.children[1].visible = !value
    }
  }

  static get moduleName() {
    return 'graphics'
  }

  registerTooltip(container: PIXI.Container, getString: () => string) {
    container.interactive = true
    this.tooltipManager.register(container, getString)
  }

  // Effects
  getFromPool<T extends Effect = Effect>(type: string): T {
    if (!this.pool[type]) {
      this.pool[type] = []
    }

    for (const e of this.pool[type]) {
      if (!e.busy) {
        e.busy = true
        e.display.visible = true
        return e as T
      }
    }

    const e = this.createEffect(type)
    this.pool[type].push(e)
    e.busy = true
    return e as T
  }

  createEffect(type: string): Effect {
    let display = null
    if (type === 'train') {
      display = new PIXI.Container()
      const sprite = PIXI.Sprite.from(TRAIN_FRAME)
      sprite.anchor.set(0.5)
      fit(sprite, this.tileSize * 1, this.tileSize * 1)
      display.addChild(sprite)
      this.trainLayer.addChild(display)
    } else if (type === 'track') {
      display = new PIXI.Container()
      const sprite = PIXI.Sprite.from('all')
      const scaler = new PIXI.Container()
      sprite.anchor.set(0.5)
      scaler.addChild(sprite)
      display.addChild(scaler)
      this.trackLayer.addChild(display)
      const effect: TrackEffect =  { busy: false, display, sprite, overlay: null, scaler }
      return effect
    } else if (type === 'gradient') {
      display = PIXI.Sprite.from('gradient_10px.png')
      this.borderLayer.addChild(display)
    } else if (type === 'gradient_corner') {
      display = PIXI.Sprite.from('gradient_10px_corner.png')
      this.borderLayer.addChild(display)
    } else if (type === 'conveyor') {
      display = PIXI.TilingSprite.from('arrow', { width: CONVEYOR_WIDTH, height: CONVEYOR_HEIGHT })
      display.anchor.x = 0
      display.anchor.y = 0.5
      this.conveyors.push(display)
      this.conveyorLayer.addChild(display)
    } else if (type.startsWith('brush')) {
      const pIdx = parseInt(type.split('_')[1])
      display = PIXI.AnimatedSprite.fromFrames(BRUSH_FRAMES[pIdx])
      display.anchor.set(49/205,175/201)
      this.brushLayer.addChild(display)
    } else if (type.startsWith('ink')) {
      const pIdx = parseInt(type.split('_')[1])
      display = PIXI.AnimatedSprite.fromFrames(INK_FRAMES[pIdx])
      display.anchor.set(32/100,109/132)

      this.brushLayer.addChild(display)
    }

    return { busy: false, display }
  }

  updateHud() {
    let currentScores = this.previousData.players.map((p) => p.score)
    this.currentData.events
      .filter((e) => e.type === ev.EARN_POINTS)
      .forEach((e) => {
        let p = this.getAnimProgress(e.animData, this.progress)
        if (p > 0) {
          currentScores[e.playerIdx] += e.score
        }
      })

    for (let player of this.globalData.players) {
      const { score} = this.huds[player.index]
      score.text = currentScores[player.index].toString()
      score.scale.set(1)
      fit(score, SCORE_RECT.w, SCORE_RECT.h)
    }
  }

  zoneStateAt(zoneId: number, progress: number): ZoneState {
    let zone: ZoneState = {
      ...this.previousData.zones[zoneId],
      prevInstability: this.previousData.zones[zoneId].instability,
      disruptP: 1,
      inkP: 1,
      prevInked: this.previousData.zones[zoneId].inked,
    }
    for (const event of this.currentData.events) {
      if (event.zoneId === zone.id) {
        if (event.type === ev.INK) {
          const p = this.getAnimProgress(event.animData, progress)
          if (p >= 0) {
            zone = { ...zone, inked: true, inkP: Math.min(1, p)}
          }
        } else if (event.type === ev.DISRUPT) {
          const p = this.getAnimProgress(event.animData, progress)
          if (p >= 0) {
            zone = { ...zone, instability: zone.instability + 1, disruptP: Math.min(1, p)}
          }
        }
      }
    }

    return zone
  }

  drawData = [
    { direction: { x: -1, y: 0 }, angle: 0, offset: { x: 0, y: 0 } },
    { direction: { x: 0, y: -1 }, angle: Math.PI / 2, offset: { x: 1, y: 0 } },
    { direction: { x: 1, y: 0 }, angle: Math.PI, offset: { x: 1, y: 1 } },
    {
      direction: { x: 0, y: 1 },
      angle: (3 * Math.PI) / 2,
      offset: { x: 0, y: 1 },
    },
  ];
  cornerData = [
    {
      directions: [
        { x: -1, y: 0 },
        { x: 0, y: -1 },
      ],
      angle: 0,
      offset: { x: 0, y: 0 },
    },
    {
      directions: [
        { x: 0, y: -1 },
        { x: 1, y: 0 },
      ],
      angle: Math.PI / 2,
      offset: { x: 1, y: 0 },
    },
    {
      directions: [
        { x: 1, y: 0 },
        { x: 0, y: 1 },
      ],
      angle: Math.PI,
      offset: { x: 1, y: 1 },
    },
    {
      directions: [
        { x: 0, y: 1 },
        { x: -1, y: 0 },
      ],
      angle: (3 * Math.PI) / 2,
      offset: { x: 0, y: 1 },
    },
  ];
  updateBorders() {
    const data = this.currentData
    const allZones = data.zones
      .map((v) => this.zoneStateAt(v.id, this.progress))
      .filter((state) => {
        return state.instability > 0 && !state.prevInked
      })

    const color = 0
    for (const zone of allZones) {
      const allCells = zone.coords
      const zoneSet = new Set(allCells.map(({ x, y }) => `${x},${y}`))

      const fromAlpha = this.alphaFromInstability(zone.prevInstability)
      const toAlpha = this.alphaFromInstability(zone.instability)
      let alphaP = lerp(fromAlpha, toAlpha, zone.disruptP)
      if (zone.inked && zone.inkP > 0) {
        alphaP = lerp(alphaP, 0, zone.inkP)
      }

      for (const cell of allCells) {
        const x = cell.x
        const y = cell.y
        // For each 4 cardinal directions, check if we draw a line at the perimeter of this tile
        for (const d of this.drawData) {
          const key = `${x + d.direction.x},${y + d.direction.y}`
          if (!zoneSet.has(key)) {
            const gradient = this.getFromPool<TEffect<PIXI.Sprite>>('gradient')
            gradient.display.height = this.tileSize

            gradient.display.width = ZONE_GRADIENT_LENGTH

            let gx = x * this.tileSize + this.tileSize * d.offset.x
            let gy = y * this.tileSize + this.tileSize * d.offset.y

            gradient.display.rotation = d.angle
            gradient.display.position.set(gx, gy)
            gradient.display.tint = color
            gradient.display.alpha = alphaP
          }
        }

        // For each 4 corner directions, check if we draw a square at the perimeter of this tile
        for (const d of this.cornerData) {
          const key1 = `${x + d.directions[0].x},${y + d.directions[0].y}`
          const key2 = `${x + d.directions[1].x},${y + d.directions[1].y}`
          const key3 = `${x + d.directions[0].x + d.directions[1].x},${
            y + d.directions[0].y + d.directions[1].y
          }`

          if (zoneSet.has(key1) && zoneSet.has(key2) && !zoneSet.has(key3)) {
            const gradient =this.getFromPool<TEffect<PIXI.Sprite>>('gradient_corner')

            gradient.display.width = ZONE_GRADIENT_LENGTH
            gradient.display.height = ZONE_GRADIENT_LENGTH

            let gx = x * this.tileSize + this.tileSize * d.offset.x
            let gy = y * this.tileSize + this.tileSize * d.offset.y

            gradient.display.rotation = d.angle
            gradient.display.position.set(gx, gy)
            gradient.display.tint = color
            gradient.display.alpha = alphaP
          }
        }
      }
    }
  }

  alphaFromInstability(instability: number):number {
    return unlerp(0, 3, instability)
  }

  updateScene(
    previousData: FrameData,
    currentData: FrameData,
    progress: number,
    playerSpeed?: number
  ) {
    const frameChange = this.currentData !== currentData
    const fullProgressChange = (this.progress === 1) !== (progress === 1)

    this.previousData = previousData
    this.currentData = currentData
    this.progress = progress
    this.playerSpeed = playerSpeed || 0

    this.resetEffects()
    this.updateTracks()
    this.updateGrid()
    this.updateDisrupts()
    this.updateBorders()
    this.updateHud()
    this.updateTrains()
    this.updateMessages()
    this.updateConveyors()

    // Time-saving hack for hiding ranking
    this.container.parent.children[1].visible = !this.api.options.debugMode
  }

  updateDisrupts() {
    const disruptEvents =  this.currentData.events
      .filter((e) => e.type === ev.DISRUPT)

    for (const e of disruptEvents) {
      const p = this.getAnimProgress(e.animData, this.progress)
      if (p > 0 && p < 1) {
        const center = this.getZoneCenter(e.zoneId)
        const pIdx = e.playerIdx
        const {display:inkFx} = this.getFromPool<AnimatedEffect>(`ink_${pIdx}`)
        this.placeInGameZone(inkFx, center)
        setAnimationProgress(inkFx, p)
        inkFx.scale.x = pIdx ? -1 : 1
      }

    }

  }

  updateConveyors() {
    if (api.options.connectionDisplay !== 1) {
      return
    }
    for (const lcta of this.liveConnectionsToAnimate) {
      const fromTown = this.globalData.towns[lcta.fromTownId]
      const toTown = this.globalData.towns[lcta.toTownId]
      const { display } = this.getFromPool<TEffect<PIXI.TilingSprite>>('conveyor')

      const fromPos = this.toBoardPos(fromTown.coord)
      const toPos = this.toBoardPos(toTown.coord)

      const scaleMult = (this.tileSize * CONVEYOR_SCALE / CONVEYOR_HEIGHT)

      display.width = Math.sqrt((toPos.x - fromPos.x) ** 2 + (toPos.y - fromPos.y) ** 2)
      display.height = this.tileSize * CONVEYOR_SCALE
      display.tileScale.set(scaleMult)
      const rotation = Math.atan2(toPos.y - fromPos.y, toPos.x - fromPos.x)
      display.rotation = rotation
      display.position.copyFrom({
        x: fromPos.x + this.tileSize / 2,
        y: fromPos.y + this.tileSize / 2,
      })
      display.tint = TRAIN_COLOR
    }
  }

  updateMessages() {
    // Update message
    for (const {text} of this.messages.flat()) {
      text.text = ''
    }
    for (const player of this.globalData.players) {
      const messageText = this.currentData.messages[player.index]
      if (messageText !== '') {

        const {text} = this.messages[player.index]

        const x = player.index == 1 ? WIDTH - 1 - MESSAGE_RECT.x : MESSAGE_RECT.x
        const y = MESSAGE_RECT.y

        text.position.set(x, y)
        text.anchor.x = player.index == 1 ? 1 : 0
        text.text = messageText

        while (text.height > MESSAGE_RECT.h) {
          text.text = text.text.slice(0, -6) + '...'
        }
      }
    }
  }

  updateGrid() {
    this.regionMask.clear()
    this.regionMask.beginFill(0xffffff)

    this.inkAnim.clear()
    this.inkAnim.beginFill(0xffffff)

    this.inkMask.clear()
    this.inkMask.beginFill(0xffffff)

    for (let y = 0; y < this.globalData.height; y++) {
      for (let x = 0; x < this.globalData.width; x++) {
        const tile = this.tiles[y][x]
        const tileData = this.globalData.tiles[this.flat({ x, y })]

        tile.overlay.tint = tile.baseTint
        tile.overlay.alpha = BASE_OVERLAY_ALPHA
        let showWater = true

        const zoneState = this.zoneStateAt(tileData.zoneId, this.progress)
        if (zoneState.inked) {
          if (zoneState.inkP < 1) {
            this.regionMask.drawRect(this.tileSize * x, this.tileSize * y, this.tileSize, this.tileSize)
          }

          showWater = zoneState.inkP < 1
          if (zoneState.inkP >= 1) {
            this.inkMask.drawRect(this.tileSize * x, this.tileSize * y, this.tileSize, this.tileSize)
          }
        }
      }
    }

    const inkEvents =  this.currentData.events
      .filter((e) => e.type === ev.INK)

    for (const e of inkEvents) {
      const p = this.getAnimProgress(e.animData, this.progress)
      if (p > 0 && p < 1) {
        // Centre of the region
        const center = this.getZoneCenter(e.zoneId)

        this.inkAnim.drawCircle(
          center.x * this.tileSize + this.tileSize / 2,
          center.y * this.tileSize + this.tileSize / 2,
          p * 4 * this.tileSize,
        )
      }
    }

    this.inkMask.endFill()
    this.inkAnim.endFill()
    this.regionMask.endFill()
    getRenderer().render(this.inkAnimLayer, this.inkAnimTexture, true)

    for (const {sprite, overlay} of this.bumpingTiles) {
      sprite.width = this.tileSize
      sprite.height = this.tileSize
      overlay.width = this.tileSize
      overlay.height = this.tileSize
    }
    this.bumpingTiles = []
    const bumps = this.currentData.events
      .filter((e) => e.type === ev.POI_BUMP)

    bumps.forEach((e) => {
      const p = this.getAnimProgress(e.animData, this.progress)
      if (p <= 0 || p > 1) {
        return
      }

      const tile = this.tiles[e.coord.y][e.coord.x]

      const spriteScale = tile.sprite.scale.x
      const overlayScale = tile.overlay.scale.x
      tile.sprite.scale.set( lerp(spriteScale, spriteScale * 2, bell(p)))
      tile.overlay.scale.set( lerp(overlayScale, overlayScale * 2, bell(p)))
      this.bumpingTiles.push(tile)

    })


  }

  getZoneCenter(zoneId: number): CoordDto {
    const zone = this.globalData.zoneMap[zoneId]
    let sumX = 0
    let sumY = 0
    for (const coord of zone.coords) {
      sumX += coord.x
      sumY += coord.y
    }
    return {
      x: sumX / zone.coords.length,
      y: sumY / zone.coords.length,
    }
  }

  flat(coord: CoordDto): number {
    return coord.y * this.globalData.width + coord.x
  }

  getAnimProgress({ start, end }: AnimData, progress: number) {
    return unlerpUnclamped(start, end, progress)
  }

  upThenDown(t: number) {
    return Math.min(1, bell(t) * 2)
  }

  toBoardPos(coord: PIXI.IPointData) {
    return {
      x: coord.x * this.tileSize,
      y: coord.y * this.tileSize,
    }
  }

  placeInGameZone(display: PIXI.DisplayObject, coord: PIXI.IPointData) {
    const pos = this.toBoardPos(coord)
    display.position.set(
      pos.x + this.tileSize / 2,
      pos.y + this.tileSize / 2
    )
  }

  updateTrains() {
    const activeConnectionMap: Record<string, ConnectionData> = {}

    this.currentData.connections.forEach((c) => {
      activeConnectionMap[keyOf(c.fromTownId, c.toTownId)] = c
    })

    this.currentData.events
      .filter((e) => e.type === ev.CONNECTION_GAINED)
      .forEach((e) => {
        const p = this.getAnimProgress(e.animData, this.progress)
        if (p > 0) {
          const fadeInP = Math.min(1, p)

          activeConnectionMap[keyOf(e.fromTownId, e.toTownId)] = {
            fromTownId: e.fromTownId,
            toTownId: e.toTownId,
            coords: e.coords,
            fadeInP: fadeInP
          }
        } else {
          // too early to animate
          delete activeConnectionMap[keyOf(e.fromTownId, e.toTownId)]
        }
      })

    this.currentData.events
      .filter((e) => e.type === ev.CONNECTION_LOST)
      .forEach((e) => {
        const p = this.getAnimProgress(e.animData, this.progress)
        if (p >= 1) {
          // no longer connected
          delete activeConnectionMap[keyOf(e.fromTownId, e.toTownId)]
        } else {
          const c = this.previousData.connections.find((c => c.fromTownId === e.fromTownId && c.toTownId === e.toTownId))
          activeConnectionMap[keyOf(c.fromTownId, c.toTownId)] = {
            ...c,
            fadeOutP: Math.max(0, p)
          }
        }
      })

    this.liveConnectionsToAnimate = []
    for (const key in activeConnectionMap) {
      const connection = activeConnectionMap[key]
      const trainEffects = []
      if (api.options.connectionDisplay === 0) {
        const trainCount = 1
        for (let i = 0; i < trainCount; i++) {
          let effect = this.getFromPool<TEffect<PIXI.Container>>('train')
          trainEffects.push(effect.display)
          let alpha = 1
          if (connection.fadeInP !== undefined) {
            alpha = connection.fadeInP
          } else if (connection.fadeOutP !== undefined) {
            alpha = 1 - connection.fadeOutP
          }
          effect.display.alpha = alpha
        }
      }
      this.liveConnectionsToAnimate.push({ ...connection, trainEffects })
    }
  }

  tintTrack(sprite: PIXI.Sprite, pId: number) {
    if (pId === 2) {
      sprite.tint = COLOR_NEUTRAL
    } else {
      sprite.tint = this.globalData.players[pId].color
      if (api.options.trackConstrast === 1) {
        if (pId === 0) {
          // #ff8f16 to #000000
          sprite.tint = lerpColor(0xff8f16, 0x0, 0.3)
        } else if (pId === 1) {
          // #3ac5ca
          sprite.tint = 0x3ac5ca
        }
      }
    }
  }

  updateTracks() {
    const railsToDraw: Set<string> = new Set()

    const buildEvents =  this.currentData.events
      .filter((e) => e.type === ev.BUILD)

    for (let y = 0; y < this.globalData.height; y++) {
      for (let x = 0; x < this.globalData.width; x++) {
        const tileTrack = this.getTileTrackAt({ x, y }, this.progress)
        if (tileTrack > -1) {
          railsToDraw.add(`${x},${y}`)
        }
      }
    }

    buildEvents.forEach((e) => {
      railsToDraw.add(`${e.coord.x},${e.coord.y}`)
    })

    const railMap: Record<string, TrackEffect> = {}

    for (const key of railsToDraw) {
      const [x, y] = key.split(',').map((v) => parseInt(v))
      const tileTrack = this.getTileTrackAt({ x, y }, this.progress)

      const effect = this.getFromPool<TrackEffect>('track')
      const sprite = effect.sprite

      effect.scaler.scale.set(1)
      effect.display.visible = true
      effect.display.zIndex = 0
      if (tileTrack !== -1) {
        this.tintTrack(sprite, tileTrack)
      }

      let trackCode = 0

      this.drawData
        .map((d) => d.direction)
        .forEach((d, index) => {
          const nCoord = { x: x + d.x, y: y + d.y }
          if (
            nCoord.x >= 0 &&
                nCoord.x < this.globalData.width &&
                nCoord.y >= 0 &&
                nCoord.y < this.globalData.height
          ) {
            const nTileTrack = this.getTileTrackAt(nCoord, this.progress)
            if (
              nTileTrack > -1 ||
                  this.globalData.towns.some(
                    (t) => t.coord.x === nCoord.x && t.coord.y === nCoord.y
                  )
            ) {
              trackCode = trackCode | (1 << (3 - index))
            }
          }
        })

      const railData = RAIL_DATA.byTrackCode(trackCode)
      sprite.texture = PIXI.Texture.from(railData.texture)
      sprite.rotation = railData.rotation

      this.placeInGameZone(effect.display, { x, y })
      effect.display.width = this.tileSize
      effect.display.height = this.tileSize

      railMap[key] = effect

    }

    buildEvents
      .forEach((e) => {
        const effect = railMap[`${e.coord.x},${e.coord.y}`]
        if (!effect) {
          console.log('effect not found, i guess this zone got inked')
          return
        }
        let p = this.getAnimProgress(e.animData, this.progress)
        if (p < 0) {
          effect.display.visible = false
          return
        }
        p = Math.min(1, p)
        effect.display.visible = true
        this.tintTrack(effect.sprite, e.playerIdx)
        const scaler = effect.scaler
        scaler.scale.set(p < 0.5 ? 0 : 1)

        if (p > 0 && p < 1) {
          const brushPlayers = []
          if (e.playerIdx === 2) {
            brushPlayers.push(0, 1)
          } else {
            brushPlayers.push(e.playerIdx)
          }
          for (const pIdx of brushPlayers) {
            const {display:brushFx} = this.getFromPool<AnimatedEffect>(`brush_${pIdx}`)
            this.placeInGameZone(brushFx, e.coord)
            setAnimationProgress(brushFx, p)
            brushFx.scale.x = pIdx ? -1 : 1
          }

        }
      })

    const bumps = this.currentData.events
      .filter((e) => e.type === ev.BUMP)

    bumps.forEach((e) => {
      const p = this.getAnimProgress(e.animData, this.progress)
      if (p <= 0 || p > 1) {
        return
      }

      const coord = e.coord
      const effect = railMap[`${coord.x},${coord.y}`]
      if (!effect) {
        return
      }
      const scaler = effect.scaler
      const scale = Math.max(scaler.scale.x, lerp(1, 2, bell(p)))
      scaler.scale.set(scale)
      effect.display.zIndex = scale
    })


  }
  getTileTrackAt(coord: CoordDto, progress: number): number {
    const tileDto = this.currentData.tiles[this.flat(coord)]
    // Find first item in history with p <= progress
    let track = this.previousData.tiles[this.flat(coord)].track
    for (let i = tileDto.history.length - 1; i >= 0; i--) {
      const h = tileDto.history[i]
      if (h.p <= progress) {
        track = h.track
        break
      }
    }

    return track
  }

  resetEffects() {
    for (const type in this.pool) {
      for (const effect of this.pool[type]) {
        effect.display.visible = false
        effect.busy = false
      }
    }
  }

  renderLiveConnections() {
    const globalTime = this.trainTime
    if (api.options.connectionDisplay !== 0) {
      return
    }
    for (const liveConnection of this.liveConnectionsToAnimate) {
      const trainCount = liveConnection.trainEffects.length
      let timePerCoord = 400

      const totalJourneyTime = timePerCoord * (liveConnection.coords.length - 1)
      for (let i = 0; i < trainCount; i++) {
        const localTime = globalTime + i * (totalJourneyTime / (trainCount + 2))
        const tripP = unlerp(
          0,
          totalJourneyTime,
          localTime % totalJourneyTime
        )

        const p = tripP

        const coords = liveConnection.coords
        const effect = liveConnection.trainEffects[i]
        const t = lerp(0, coords.length - 1, p)
        const from = coords[Math.floor(t)]
        const to = coords[Math.ceil(t)]
        const fromToP = t - Math.floor(t)

        const current = {
          x: lerp(from.x, to.x, fromToP),
          y: lerp(from.y, to.y, fromToP),
        }
        this.placeInGameZone(effect, current)
        if (from.x > to.x) {
          effect.scale.set(-1, 1)
        } else if (from.x < to.x) {
          effect.scale.set(1, 1)
        }
      }
    }
  }

  renderMessages(delta: number) {
    for (let player of this.globalData.players) {
      const message = this.messages[player.index]

      const options = this.api.options
      const stepFactor = Math.pow(0.99, delta)
      const targetMessageAlpha = (options.showMyMessages && player.isMe) || (options.showOthersMessages && !player.isMe) ? 1 : 0
      message.text.alpha = message.text.alpha * stepFactor + targetMessageAlpha * (1 - stepFactor)
      const targetFrameAlpha = message.text.text === '' ? 0 : targetMessageAlpha
      message.frame.alpha = message.frame.alpha * stepFactor + targetFrameAlpha * (1 - stepFactor)
    }
  }

  renderConveyors(delta: number) {
    for (const c of this.conveyors) {
      if (c.visible) {
        c.tilePosition.x = this.time * c.tileScale.x / 16
      }
    }
  }

  animateScene(delta: number) {
    this.time += delta

    let timeCoeff = 1
    if (this.playerSpeed > 0) {
      timeCoeff = this.playerSpeed
    }
    this.trainTime += delta * timeCoeff
    this.renderLiveConnections()
    this.renderMessages(delta)
    this.renderConveyors(delta)
  }

  asLayer(func: ContainerConsumer): PIXI.Container {
    const layer = new PIXI.Container()
    func.bind(this)(layer)
    return layer
  }

  reinitScene(container: PIXI.Container, canvasData: CanvasInfo) {
    (window as any).g = new PIXI.Graphics()
    this.time = 0
    this.trainTime = 0
    this.oversampling = canvasData.oversampling
    this.container = container
    this.pool = {}
    this.canvasData = canvasData
    this.conveyors = []
    this.bumpingTiles = []

    const gameZone = new PIXI.Container()

    this.tileSize = Math.min(
      GAME_ZONE_RECT.w / this.globalData.width,
      GAME_ZONE_RECT.h / this.globalData.height
    )

    const tooltipLayer = this.tooltipManager.reinit()
    this.messageLayer = this.asLayer(this.initMessages)
    this.inkMask = new PIXI.Graphics()
    this.inkAnim = new PIXI.Graphics()
    this.regionMask = new PIXI.Graphics()

    const background = PIXI.Sprite.from('Background.jpg')
    background.tint = 0xffffff
    background.width = WIDTH
    background.height = HEIGHT
    this.trainLayer = new PIXI.Container()
    this.trackLayer = new PIXI.Container()
    this.borderLayer = new PIXI.Container()
    this.brushLayer = new PIXI.Container()
    this.conveyorLayer = new PIXI.Container()

    const riverLayer = this.asLayer(this.initRiverLayer)
    const townLayer = this.asLayer(this.initTownLayer)
    const grid = this.asLayer(this.initGrid)
    const zoneLines = this.asLayer(this.initZoneLines)
    const hud = this.asLayer(this.initHud)


    gameZone.x = GAME_ZONE_RECT.x
    gameZone.y = GAME_ZONE_RECT.y
    const gameWidth = this.globalData.width * this.tileSize
    const gameHeight = this.globalData.height * this.tileSize
    gameZone.x += (GAME_ZONE_RECT.w - gameWidth) / 2
    gameZone.y += (GAME_ZONE_RECT.h - gameHeight) / 2
    this.gameZone = gameZone

    const ink = PIXI.Sprite.from ('cover.png')
    ink.alpha = 1
    ink.mask = this.inkMask
    ink.width = gameWidth
    ink.height = gameHeight

    this.inkAnim.mask = this.regionMask
    this.inkAnimTexture = PIXI.RenderTexture.create({ width: gameWidth / 2, height: gameHeight / 2})
    flagForDestructionOnReinit(this.inkAnimTexture)
    const inkAnimSprite = new PIXI.Sprite(this.inkAnimTexture)
    inkAnimSprite.width = gameWidth
    inkAnimSprite.height = gameHeight

    const ink2 = PIXI.Sprite.from ('cover.png')
    ink2.alpha = 1
    ink2.mask = inkAnimSprite
    ink2.width = gameWidth
    ink2.height = gameHeight

    this.inkAnimLayer = new PIXI.Container()
    this.inkAnimLayer.addChild(this.regionMask)
    this.inkAnimLayer.addChild(this.inkAnim)
    this.inkAnimLayer.scale.set(0.5)

    this.trackLayer.sortableChildren = true
    gameZone.addChild(grid)
    gameZone.addChild(riverLayer)


    gameZone.addChild(zoneLines)
    gameZone.addChild(this.borderLayer)
    gameZone.addChild(this.trackLayer)
    gameZone.addChild(this.inkMask)
    gameZone.addChild(ink)
    gameZone.addChild(ink2)
    gameZone.addChild(inkAnimSprite)

    gameZone.addChild(this.conveyorLayer)
    gameZone.addChild(this.trainLayer)
    gameZone.addChild(townLayer)
    gameZone.addChild(this.brushLayer)


    container.addChild(background)
    container.addChild(hud)
    container.addChild(gameZone)
    container.addChild(this.messageLayer)
    container.addChild(tooltipLayer)

    container.interactiveChildren = false


    container.interactive = true
    container.on('mousemove', (event: PIXI.InteractionEvent) => {
      this.tooltipManager.moveTooltip(event)
      const pos = event.data.getLocalPosition(gameZone)
      const x = Math.floor(pos.x / this.tileSize)
      const y = Math.floor(pos.y / this.tileSize)
      this.hoverTownId = this.globalData.townMap[`${x},${y}`]?.id ?? null
    })

    this.tooltipManager.registerGlobal((data) => {
      const pos = data.getLocalPosition(gameZone)
      const x = Math.floor(pos.x / this.tileSize)
      const y = Math.floor(pos.y / this.tileSize)
      const index = this.flat({ x, y })
      const tile = this.globalData.tiles[index]
      if (
        x < 0 ||
        x >= this.globalData.width ||
        y < 0 ||
        y >= this.globalData.height
      ) {
        return null
      }
      const blocks = []
      blocks.push(`(${x}, ${y}) ` + getTileTypeName(tile.type))

      let regionBlock = `Region ${tile.zoneId}`
      const inst = this.currentData.zones[tile.zoneId].instability
      if (inst > 0) {
        regionBlock += `\ninstability: ${inst}`
      }
      blocks.push(regionBlock)
      const town = this.globalData.towns.find(
        (t) => t.coord.x === x && t.coord.y === y
      )
      if (town) {
        let townBlock = `Town ${town.id}\nDesired: ${
          town.desiredConnections.length
            ? town.desiredConnections.join(', ')
            : 'None'
        }`

        const connectedTo = this.currentData.connections.filter((c) => c.fromTownId === town.id).map((c) => c.toTownId)
        if (connectedTo.length) {
          townBlock += `\nConnected: ${connectedTo.join(', ')}`
        }


        blocks.push(townBlock)
      }

      return blocks.join('\n--------\n')
    })

    api.renderUpdate = () => {
      this.updateScene(this.previousData, this.currentData, this.progress, this.playerSpeed)
    }
  }

  initRiverLayer(layer: PIXI.Container) {
    this.waterMap = {}
    for (let y = 0; y < this.globalData.height; y++) {
      for (let x = 0; x < this.globalData.width; x++) {
        const tile = this.globalData.tiles[this.flat({ x, y })]
        if (tile.type === TYPE_WATER) {
          let trackCode = 0

          this.drawData
            .map((d) => d.direction)
            .forEach((d, index) => {
              const nCoord = { x: x + d.x, y: y + d.y }
              if (
                nCoord.x >= 0 &&
                nCoord.x < this.globalData.width &&
                nCoord.y >= 0 &&
                nCoord.y < this.globalData.height
              ) {
                const nTile = this.globalData.tiles[this.flat(nCoord)]
                if (nTile.type === TYPE_WATER) {
                  trackCode = trackCode | (1 << (3 - index))
                }
              } else {
                trackCode = trackCode | (1 << (3 - index))
              }
            })
          const sprite = new PIXI.Sprite(PIXI.Texture.WHITE)
          const railData = RAIL_DATA.byTrackCode(trackCode, true)
          sprite.texture = PIXI.Texture.from('water_' + railData.texture)
          sprite.rotation = railData.rotation
          sprite.width = this.tileSize
          sprite.height = this.tileSize
          sprite.anchor.set(0.5, 0.5)
          this.placeInGameZone(sprite, { x, y })
          layer.addChild(sprite)

          this.waterMap[`${x},${y}`] = sprite
        }
      }
    }
  }

  initTownLayer(layer: PIXI.Container) {
    for (const town of this.globalData.towns) {
      const sprite = new PIXI.Sprite(PIXI.Texture.from(TOWN_FRAME))
      sprite.width = this.tileSize * 1.5
      sprite.height = this.tileSize * 1.5
      sprite.anchor.set(0.5, 0.5)
      this.placeInGameZone(sprite, town.coord)
      layer.addChild(sprite)

      const text = new PIXI.Text(town.id.toString(), {
        fontSize: '32px',
        fill: 'white',
        fontWeight: 'bold',
        strokeThickness: 2,
        stroke: 'black'

      })
      text.anchor.set(0.5)
      this.placeInGameZone(text, town.coord)
      text.y -= this.tileSize /2
      text.x-= this.tileSize /3
      layer.addChild(text)
    }
  }

  initMessages(layer: PIXI.Container) {
    this.messages = []
    for (const player of this.globalData.players) {
      const messageFrame = PIXI.Sprite.from(player.index === 0 ? 'Tchat_Rouge' : 'Tchat_Bleu')
      messageFrame.position.x = player.index == 1 ? WIDTH - 1 - (MESSAGE_RECT.x - 10) : (MESSAGE_RECT.x - 10)
      messageFrame.position.y = MESSAGE_RECT.y - 155
      messageFrame.anchor.x = player.index == 1 ? 1 : 0
      layer.addChild(messageFrame)

      const messageText = new PIXI.Text('', {
        fontFamily: 'Arial',
        fontWeight: '700',
        fontSize: 34,
        fill: player.color,
        // stroke: '#000000',
        // strokeThickness: 4,
        align: 'left',
        wordWrap: true,
        wordWrapWidth: MESSAGE_RECT.w
      })

      this.messages.push({text: messageText, frame: messageFrame})
      layer.addChild(messageText)

    }
  }

  initHud(layer: PIXI.Container) {
    const background = PIXI.Sprite.from('HUD.png')
    layer.addChild(background)

    this.huds = []
    for (const player of this.globalData.players) {
      const avatar = PIXI.Sprite.from(player.avatar)
      const name = new PIXI.Text(player.name, {
        fontSize: '48px',
        fill: HUD_COLORS[player.index],
        fontWeight: 'bold',
      })
      const score = new PIXI.Text('999', {
        fontSize: '48px',
        fill: HUD_COLORS[player.index],
        fontWeight: 'bold',
      })

      this.placeInHUD(avatar, AVATAR_RECT, player.index)
      this.placeInHUD(name, NAME_RECT, player.index)
      this.placeInHUD(score, SCORE_RECT, player.index)

      layer.addChild(avatar, name, score)

      this.huds.push({ avatar, name, score })
    }

    const companyLogoFrame = PIXI.Sprite.from('Cadre_Logo.png')
    companyLogoFrame.position.x = WIDTH / 2
    companyLogoFrame.position.y = 55
    companyLogoFrame.anchor.set(0.5, 0.5)
    fit(companyLogoFrame, 316+148, 78+42)
    layer.addChild(companyLogoFrame)

    const companyLogo = PIXI.Sprite.from('codingame.png')
    companyLogo.position.x = WIDTH / 2
    companyLogo.position.y = 50
    companyLogo.anchor.set(0.5, 0.5)
    fit(companyLogo, 316, 78)
    layer.addChild(companyLogo)

  }

  placeInHUD(
    element: PIXI.Text | PIXI.Sprite,
    { x, y, w, h }: { x: number; y: number; w: number; h: number },
    pIdx: number
  ) {
    fit(element, w, h)
    element.position.set(pIdx ? WIDTH - 1 - x : x, y)
    // element.anchor.set(pIdx ? 1 : 0, 0)

    element.position.set(pIdx ? WIDTH - 1 - x - w / 2 : x + w / 2, y)

    element.anchor.x = 0.5
  }

  initZoneLines(layer: PIXI.Container) {
    const gridLines = new PIXI.Container()
    gridLines.x = ZONE_LINE_WIDTH
    gridLines.y = ZONE_LINE_WIDTH
    const alreadyDrawnBorders = new Set<string>()
    for (const zone of this.globalData.zones) {
      const zoneSet = new Set(zone.coords.map(({ x, y }) => `${x},${y}`))

      for (const cell of zone.coords) {
        const x = cell.x
        const y = cell.y
        // For each 4 cardinal directions, check if we draw a line at the perimeter of this tile
        for (const d of this.drawData) {
          const key = `${x + d.direction.x},${y + d.direction.y}`
          if (!zoneSet.has(key)) {
            const borderKey = `${Math.min(x, x + d.direction.x)},${Math.min(y, y + d.direction.y)}>${Math.max(x, x + d.direction.x)},${Math.max(y, y + d.direction.y)}`
            if (alreadyDrawnBorders.has(borderKey)) {
              continue
            }
            alreadyDrawnBorders.add(borderKey)
            const line = PIXI.Sprite.from('trait.png')
            line.height = this.tileSize

            let gx = x * this.tileSize + this.tileSize * d.offset.x
            let gy = y * this.tileSize + this.tileSize * d.offset.y

            line.anchor.x = 0.5

            line.rotation = d.angle
            line.position.set(gx, gy)
            gridLines.addChild(line)

          }
        }
      }
    }
    const texture = PIXI.RenderTexture.create({ width: WIDTH, height: HEIGHT })
    flagForDestructionOnReinit(texture)
    getRenderer().render(gridLines, texture)

    const gridLineSprite = new PIXI.Sprite(texture)
    gridLineSprite.alpha = 1 //0.2
    // An offset was needed to fit the lines into the render texture, another to center it around the tiles
    gridLineSprite.position.set(-ZONE_LINE_WIDTH - 1, -ZONE_LINE_WIDTH - 1)
    layer.addChild(gridLineSprite)
    Object.defineProperty(gridLineSprite, 'visible', {
      get: () => api.options.territories === 1,
    })
  }

  mix(a:number, b:number):number {
    let x = Math.sin(a * 12.9898 + b * 78.233) * 43758.5453
    return x - Math.floor(x) // value in [0, 1)
  }

  initGrid(layer: PIXI.Container) {
    this.tiles = []
    let poiIdx = 0

    const ignoredMountains = new Set<string>()

    for (let y = 0; y < this.globalData.height; y++) {
      const row: Tile[] = []
      for (let x = 0; x < this.globalData.width; x++) {
        const tile = this.globalData.tiles[this.flat({ x, y })]

        const tileContainer = new PIXI.Container()
        tileContainer.x = this.tileSize * x
        tileContainer.y = this.tileSize * y
        tileContainer.zIndex = 0

        let sprite: PIXI.Sprite = null
        let overlay: PIXI.Sprite = new PIXI.Sprite(PIXI.Texture.WHITE)
        overlay.width = this.tileSize
        overlay.height = this.tileSize

        if (ignoredMountains.has(`${x},${y}`)) {
          overlay.tint = lerpColor(0x8B4513, 0xFFFFFF, 0.5)
          sprite = new PIXI.Sprite(PIXI.Texture.EMPTY)
        } else if (tile.type === TYPE_GRASS || tile.type === TYPE_WATER) {
          sprite = new PIXI.Sprite(PIXI.Texture.from(choice(GRASS_FRAMES)))
          overlay.tint =lerpColor(0x007700, 0xFFFFFF, 0.7)

          sprite.alpha = choice([0,1])
          sprite.anchor.set(0.5)
          sprite.x = this.tileSize / 2
          sprite.y = this.tileSize / 2
          sprite.width = this.tileSize
          sprite.height = this.tileSize

        } else if (tile.type === TYPE_POI) {
          sprite = new PIXI.Sprite(PIXI.Texture.from(POI_FRAMES[poiIdx % POI_FRAMES.length]))
          poiIdx++
          overlay.tint = lerpColor(0xFFC0CB, 0xFFFFFF, 0.1)
          overlay.anchor.set(0.5)
          overlay.x = this.tileSize / 2
          overlay.y = this.tileSize / 2

          tileContainer.zIndex = 9
          sprite.anchor.set(0.5)
          sprite.x = this.tileSize / 2
          sprite.y = this.tileSize / 2
          sprite.width = this.tileSize
          sprite.height = this.tileSize
        } else if (tile.type === TYPE_MOUNTAIN) {
          overlay.tint = lerpColor(0x8B4513, 0xFFFFFF, 0.5)

          // check right and bottom and bottom right
          const rightCell = x < this.globalData.width - 1 ? this.globalData.tiles[this.flat({ x: x + 1, y })] : null
          const bottomCell = y < this.globalData.height - 1 ? this.globalData.tiles[this.flat({ x, y: y + 1 })] : null
          const bottomRightCell = (x < this.globalData.width - 1 && y < this.globalData.height - 1) ? this.globalData.tiles[this.flat({ x: x + 1, y: y + 1 })] : null

          if (rightCell?.type === TYPE_MOUNTAIN && bottomCell?.type === TYPE_MOUNTAIN && bottomRightCell?.type === TYPE_MOUNTAIN) {
            // 2x2
            ignoredMountains.add(`${x+1},${y}`)
            ignoredMountains.add(`${x},${y+1}`)
            ignoredMountains.add(`${x+1},${y+1}`)

            sprite = new PIXI.Sprite(PIXI.Texture.from(choice(MOUNTAIN_FRAMES_2x2)))
            sprite.width = this.tileSize * 2
            sprite.height = this.tileSize * 2
          } else if (rightCell?.type === TYPE_MOUNTAIN) {
            // 2x1
            ignoredMountains.add(`${x+1},${y}`)

            sprite = new PIXI.Sprite(PIXI.Texture.from(choice(MOUNTAIN_FRAMES_2x1)))
            sprite.width = this.tileSize * 2
            sprite.height = this.tileSize
          } else {
            sprite = new PIXI.Sprite(PIXI.Texture.from(choice(MOUNTAIN_FRAMES_1x1)))
            sprite.width = this.tileSize
            sprite.height = this.tileSize
          }
        }
        tileContainer.addChild(sprite)
        tileContainer.addChild(overlay)

        overlay.alpha = BASE_OVERLAY_ALPHA

        layer.addChild(tileContainer)
        row.push({
          container: tileContainer,
          sprite,
          overlay,
          baseTint: overlay.tint
        })


      }
      this.tiles.push(row)
      layer.sortableChildren = true
    }
  }

  easeOutElastic(x: number): number {
    const c4 = (2 * Math.PI) / 3

    return x === 0
      ? 0
      : x === 1
        ? 1
        : Math.pow(2, -10 * x) * Math.sin((x * 10 - 0.75) * c4) + 1
  }

  handleGlobalData(players: PlayerInfo[], raw: string): void {
    const globalData = parseGlobalData(raw)
    api.options.meInGame = !!players.find((p) => p.isMe)

    this.globalData = {
      ...globalData,
      players: players,
      playerCount: players.length,
      townMap: globalData.towns.reduce((acc, town) => {
        acc[`${town.coord.x},${town.coord.y}`] = town.id
        return acc
      }, {})
    }
  }

  trackStateAt(coord: CoordDto, progress: number): any {
  }

  handleFrameData(frameInfo: FrameInfo, raw: string): FrameData {
    const dto = parseData(raw, this.globalData)
    const previousFrame = last(this.states)
    const frameData: FrameData = {
      ...dto,
      previous: null,
      frameInfo,
      connections: previousFrame?.connections ?? [],
      zones:
        previousFrame?.zones ??
        this.globalData.zones.map((z) => ({
          ...z,
          inked: false,
          instability: 0,
        })),
      tiles:
        previousFrame?.tiles ??
        new Array(this.globalData.width * this.globalData.height).fill({
          track: -1,
        }),
      players:
        previousFrame != null
          ? previousFrame.players.map((v) => ({ ...v }))
          : this.globalData.players.map((p) => ({
            score: 0,
          })),
    }

    frameData.previous = previousFrame ?? frameData
    frameData.tiles = frameData.tiles.map(t => ({ ...t, history: [] }))

    for (const event of dto.events) {
      event.animData.start /= frameInfo.frameDuration
      event.animData.end /= frameInfo.frameDuration
    }

    for (const event of dto.events) {
      if (event.type === ev.BUILD) {
        const t = frameData.tiles[this.flat(event.coord)]
        frameData.tiles[this.flat(event.coord)] = {
          track: event.playerIdx,
          history: [
            ...t.history,
            {p: event.animData.start, track: event.playerIdx}
          ]
        }
      } else if (event.type === ev.DISRUPT) {
        frameData.zones = [...frameData.zones]
        frameData.zones[event.zoneId] = {
          ...frameData.zones[event.zoneId],
          instability: frameData.zones[event.zoneId].instability + 1,
        }
      } else if (event.type === ev.EARN_POINTS) {
        frameData.players = frameData.players.map((p, idx) => {
          if (idx === event.playerIdx) {
            return {
              ...p,
              score: p.score + event.score,
            }
          }
          return p
        })
      } else if (event.type === ev.INK) {
        frameData.zones = [...frameData.zones]
        frameData.zones[event.zoneId] = {
          ...frameData.zones[event.zoneId],
          inked: true,
        }
        event.coords.forEach((c) => {
          const t = frameData.tiles[this.flat(c)]
          frameData.tiles[this.flat(c)] = {
            track: -1,
            history: [
              ...t.history,
              {p: event.animData.end, track: -1}
            ]
          }
        })
      } else if (event.type === ev.CONNECTION_GAINED) {
        //TODO: connections probl need some sort of history if they appear then dissapear during the same frame?
        frameData.connections = [...frameData.connections]
        frameData.connections.push({
          fromTownId: event.fromTownId,
          toTownId: event.toTownId,
          coords: event.coords,
        })
      } else if (event.type === ev.CONNECTION_LOST) {
        frameData.connections = [...frameData.connections]
        const idx = frameData.connections.findIndex(
          (c) =>
            c.fromTownId === event.fromTownId && c.toTownId === event.toTownId
        )
        frameData.connections.splice(idx, 1)
      }
    }

    this.states.push(frameData)
    return frameData
  }
}
function getTileTypeName(type: number) {
  switch (type) {
  case TYPE_GRASS:
    return 'plains'
  case TYPE_WATER:
    return 'river'
  case TYPE_MOUNTAIN:
    return 'mountain'
  case TYPE_POI:
    return 'P.O.I.'
  }
  return ''
}

