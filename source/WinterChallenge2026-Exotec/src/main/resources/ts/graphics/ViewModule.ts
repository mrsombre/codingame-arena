import { IPointData, TextureUvs } from 'pixi.js'
import { HEIGHT, WIDTH } from '../core/constants.js'
import { flagForDestructionOnReinit, getRenderer } from '../core/rendering.js'
import { bell, easeIn, easeOut } from '../core/transitions.js'
import { fitAspectRatio, lerp, lerpColor, lerpAngle, lerpPosition, unlerp, unlerpUnclamped } from '../core/utils.js'
import { AnimatedEffect, AnimData, CanvasInfo, ContainerConsumer, Effect, FrameData, FrameInfo, GlobalData, PlayerInfo, BirdEffect, TEffect, BirdDto, CoordDto, EventDto, AppleEffect } from '../types.js'
import { parseData, parseGlobalData } from './Deserializer.js'
import { TooltipManager } from './TooltipManager.js'
import ev from './events.js'
import { angleBetween, computeRotationAngle, normalizeAngle, rotateAround, subtract } from './trigo.js'
import { angleDiff, choice, drawDebugFrameAroundObject, fit, last, setAnimationProgress } from './utils.js'
import { AIGUILLE_FRAME, APPLE_FRAME, AVATAR_RECT, BODY_FRAME, DEATH_FRAMES, EAT_FRAMES, FALL_HEAD_FRAME, GAME_ZONE_RECT, HEAD_FRAME, MARK_FRAME, NAME_RECT, SCORE_RECT, SPARK_FRAMES, TILES, TIME_RECT, TRANSFORM_FRAMES } from './assetConstants.js'
import { MessageContainer, initMessages, renderMessageContainer } from './MessageBoxes.js'
import { initFilters } from './pixi-filters.js'

// meta
//TODO: arena (leagues and threshholds)
//TODO: github repo

const fragment = `
precision mediump float;

varying vec2 vTextureCoord;
uniform sampler2D uSampler;
uniform float hueShift; // radians

const vec3 LUMA = vec3(0.299, 0.587, 0.114);

vec3 hueRotate(vec3 color, float angle) {
  float cosA = cos(angle);
  float sinA = sin(angle);

  mat3 rot = mat3(
    LUMA.x + (1.0 - LUMA.x) * cosA - LUMA.x * sinA,
    LUMA.x - LUMA.x * cosA + 0.143 * sinA,
    LUMA.x - LUMA.x * cosA - (1.0 - LUMA.x) * sinA,

    LUMA.y - LUMA.y * cosA - LUMA.y * sinA,
    LUMA.y + (1.0 - LUMA.y) * cosA + 0.140 * sinA,
    LUMA.y - LUMA.y * cosA + LUMA.y * sinA,

    LUMA.z - LUMA.z * cosA + (1.0 - LUMA.z) * sinA,
    LUMA.z - LUMA.z * cosA - LUMA.z * sinA,
    LUMA.z + (1.0 - LUMA.z) * cosA + LUMA.z * sinA
  );

  return clamp(rot * color, 0.0, 1.0);
}

void main() {
  vec4 color = texture2D(uSampler, vTextureCoord);

  // only rotate saturated colors (skip white/blue eyes)
  float maxC = max(max(color.r, color.g), color.b);
  float minC = min(min(color.r, color.g), color.b);
  float saturation = maxC - minC;

  vec3 result = color.rgb;

  if (saturation > 0.25) {
    result = hueRotate(color.rgb, hueShift);
  }

  gl_FragColor = vec4(result, color.a);
}
`


const BODY_PART_OVERSCALE = 1.05

interface EffectPool {
  [key: string]: Effect[]
}


const api = {
  setDebugMode: (value: boolean) => {
    api.options.debugMode = value
  },
  options: {
    debugMode: false,
    showMarks: 3,
    showOthersMessages: true,
    showMyMessages: true,
    meInGame: false,
  }
}
export { api }



export class ViewModule {
  states: FrameData[]
  globalData: GlobalData
  pool: EffectPool
  api: any
  playerSpeed: number
  previousData: FrameData
  currentData: FrameData
  progress: number
  oversampling: number
  container: PIXI.Container
  time: number
  canvasData: CanvasInfo


  tooltipManager: TooltipManager

  messages: MessageContainer[]
  tileSize: number
  bodyPartScale: number
  gameZone: PIXI.Container
  birdLayer: PIXI.Container
  appleLayer: PIXI.Container
  markLayer: PIXI.Container
  currentMarks: {sprite: PIXI.Sprite, playerIdx: number}[]
  huds: {
    avatar: PIXI.Sprite,
    name: PIXI.Text,
    score: PIXI.Text,
    msText: PIXI.Text
  }[]
  apples: {countdown: number, effect: AppleEffect}[]
  birdContainers: PIXI.Container[]

  constructor () {
    this.states = []
    this.pool = {}
    this.time = 0
    window.debug = this
    this.tooltipManager = new TooltipManager()
    this.api = api
    this.api.setDebugMode = (value: boolean) => {
      //hack for hiding ranking
      this.api.options.debugMode = value
      this.container.parent.children[1].visible = !value
    }

    initFilters()
  }

  static get moduleName () {
    return 'graphics'
  }

  registerTooltip (container: PIXI.Container, getString: () => string) {
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

  createEffect (type: string): Effect {
    let display = null
    if (type.startsWith('head_')) {
      const pIdx = parseInt(type.split('_')[1])
      display = new PIXI.Container()

      const scaler = new PIXI.Container()
      const face = PIXI.Sprite.from(HEAD_FRAME[pIdx])
      face.anchor.set(72/134, 77/141)
      face.scale.set(this.bodyPartScale)


      const eating = PIXI.AnimatedSprite.fromFrames(EAT_FRAMES[pIdx])
      eating.anchor.set(131/354,140/258)
      eating.scale.set(this.bodyPartScale)

      const dying = PIXI.AnimatedSprite.fromFrames(DEATH_FRAMES)
      dying.anchor.set((108+165/2)/350, (63+163/2)/285)
      dying.scale.set(this.bodyPartScale)

      const id = new PIXI.Text('0', { fill: 0x0, fontSize: 22, fontWeight: 'bold' })
      id.anchor.set(0.5)
      id.visible = false

      scaler.addChild(face)
      scaler.addChild(eating)
      scaler.addChild(dying)

      eating.visible = false
      dying.visible = false

      display.addChild(scaler)
      display.addChild(id)

      const effect: BirdEffect = { pIdx, busy: false, display, faceScaler: scaler, face, eating, dying, id, getTooltip: () => '' }


      this.registerTooltip(display, () => {
        return effect.getTooltip()
      })

      return effect
    } else if (type.startsWith('body_')) {
      const pIdx = parseInt(type.split('_')[1])
      display = new PIXI.Container()
      const sprite = PIXI.Sprite.from(BODY_FRAME[pIdx])
      sprite.scale.set(this.bodyPartScale)
      sprite.anchor.set(0.5)

      const hands = []
      const handContainer = new PIXI.Container()
      for (const hIdx of [0, 1]) {
        const hand = PIXI.Sprite.from(AIGUILLE_FRAME[pIdx])
        hand.scale.set(this.bodyPartScale)
        hand.anchor.set(26/90, 25/50)
        handContainer.addChild(hand)
        hands.push(hand)
      }

      const transforming = PIXI.AnimatedSprite.fromFrames(TRANSFORM_FRAMES[pIdx])
      transforming.scale.set(this.bodyPartScale)
      transforming.anchor.set((32+147/2)/200,(46+146/2)/223)
      transforming.visible = false

      const dying = PIXI.AnimatedSprite.fromFrames(DEATH_FRAMES)
      dying.anchor.set((108+165/2)/350, (63+163/2)/285)
      dying.scale.set(this.bodyPartScale)
      dying.visible = false

      const transformingScaler = new PIXI.Container()

      const id = new PIXI.Text('0', { fill: 0x0, fontSize: 22, fontWeight: 'bold' })
      id.anchor.set(0.5)
      id.visible = false

      transformingScaler.addChild(transforming)

      const body = new PIXI.Container()
      display.addChild(body)
      body.addChild(sprite)
      body.addChild(transformingScaler)
      body.addChild(handContainer)
      display.addChild(dying)
      display.addChild(id)


      const effect: BirdEffect = {
        busy: false,
        display,
        sprite,
        dying,
        body,
        id,
        transformingScaler,
        transforming,
        getTooltip: () => '' ,
        hands
      }


      this.registerTooltip(display, () => {
        return effect.getTooltip()
      })

      return effect
    } else if (type === 'apple') {
      display = new PIXI.Container()
      const sprite = PIXI.Sprite.from(APPLE_FRAME)
      let coeff = fitAspectRatio(sprite.width, sprite.height, this.tileSize, this.tileSize)
      sprite.anchor.set(0.5)
      sprite.scale.set(coeff)
      display.addChild(sprite)

      const sparks = PIXI.AnimatedSprite.fromFrames(SPARK_FRAMES)

      coeff = fitAspectRatio(sparks.width, sparks.height, this.tileSize*2, this.tileSize*2)
      sparks.anchor.set(0.5)
      sparks.scale.set(coeff)
      sparks.visible = true
      sparks.animationSpeed = 0.35
      sparks.loop = false
      sparks.gotoAndStop(0)
      display.addChild(sparks)

      this.appleLayer.addChild(display)
      const effect: AppleEffect = {
        busy: false,
        display,
        sparks
      }
      return effect

    } else if (type === 'mark') {
      display = PIXI.Sprite.from('mark')
      display.scale.set(this.tileSize / 95)
      display.anchor.set(0.5)
      this.markLayer.addChild(display)
    }

    return { busy: false, display }
  }

  initHud(layer: PIXI.Container) {
    const background = PIXI.Sprite.from('HUD')
    layer.addChild(background)

    const hudColors = [
      0xff28ff,
      0x00ff0a,
    ]

    this.huds = []
    for (const player of this.globalData.players) {
      const msText = new PIXI.Text('exec time', {
        fill: hudColors[player.index],
        fontSize: 24,
        fontWeight: 'bold',
      })

      const color = hudColors[player.index]

      const avatar = PIXI.Sprite.from(player.avatar)
      const name = new PIXI.Text(player.name, {
        fontSize: '48px',
        fill: color,
        fontWeight: 'bold',
      })
      const score = new PIXI.Text('999', {
        fontSize: '48px',
        fill: color,
        fontWeight: 'bold',
      })

      this.placeInHUD(avatar, AVATAR_RECT, player.index)
      this.placeInHUD(name, NAME_RECT, player.index)
      this.placeInHUD(score, SCORE_RECT, player.index)
      this.placeInHUD(msText, TIME_RECT, player.index)

      layer.addChild(avatar, name, score, msText)

      this.huds.push({ avatar, name, score, msText })
    }

    const companyLogoFrame = PIXI.Sprite.from('Cadre_Logo.png')
    companyLogoFrame.position.x = WIDTH / 2
    companyLogoFrame.position.y = 0
    companyLogoFrame.anchor.set(0.5, 0)
    fit(companyLogoFrame, 316+148, 78+42)

    const companyLogo = PIXI.Sprite.from('company_logo.png')
    const rect = {
      x: 817, y: 4, w: 302, h: 89
    }

    companyLogo.position.x = rect.x + rect.w/2
    companyLogo.position.y = rect.y + rect.h/2
    companyLogo.anchor.set(0.5)

    const pad = 40
    fit(companyLogo, rect.w - pad, rect.h - pad)

    layer.addChild(companyLogoFrame)
    layer.addChild(companyLogo)

  }

  updateHud() {
    for (let pIdx = 0; pIdx < this.globalData.playerCount; ++pIdx) {
      const hud = this.huds[pIdx]
      const ms = this.currentData.ms[pIdx]
      if (ms === -1) {
        hud.msText.text = 'exec time'
      } else {
        hud.msText.text = ms +'ms'

      }

      const data = this.progress === 1 ? this.currentData : this.previousData

      const score = data.birds.filter(b => b.owner === pIdx && b.alive).flatMap(b=>b.body).length
      hud.score.text = score.toString()
    }

  }

  updateMarks() {
    this.currentMarks = []
    const marks = this.currentData.marks
    for (let pIdx = 0; pIdx < this.globalData.playerCount; ++pIdx) {
      const playerMarks: CoordDto[] = marks[pIdx]
      for (const mark of playerMarks) {
        const markEffect = this.getFromPool<TEffect<PIXI.Sprite>>('mark')
        markEffect.display.tint = this.globalData.players[pIdx].color
        this.placeInGameZone(markEffect.display, mark)
        this.currentMarks.push({sprite: markEffect.display, playerIdx: pIdx})
      }
    }

  }

  updateScene (previousData: FrameData, currentData: FrameData, progress: number, playerSpeed?: number) {
    const frameChange = (this.currentData !== currentData)
    const fullProgressChange = ((this.progress === 1) !== (progress === 1))

    this.previousData = previousData
    this.currentData = currentData
    this.progress = progress
    this.playerSpeed = playerSpeed || 0

    this.resetEffects()

    this.updateBirds()
    this.updateApples()
    this.updateHud()
    this.updateMarks()


    // Time-saving hack for hiding ranking
    this.container.parent.children[1].visible = !this.api.options.debugMode
  }



  placeInHUD(element: PIXI.Text | PIXI.Sprite, {x,y,w,h}: {x:number,y:number,w:number,h:number}, pIdx: number) {
    fit(element, w, h)
    element.position.set(pIdx ? WIDTH - 1 - (x + w/2) :  (x + w/2), y + h / 2)
    element.anchor.set(0.5, 0.51)
  }


  updateApples() {
    const appleMap = {}
    this.apples = []

    const prevApples = this.previousData.apples
    for (let i = 0; i < prevApples.length; ++i) {
      const apple = prevApples[i]

      const effect = this.getFromPool<AppleEffect>('apple')
      const display = effect.display

      this.placeInGameZone(display, apple)
      appleMap[`${apple.x},${apple.y}`] = display
      display.scale.set(1)

      this.apples.push({effect, countdown: Math.random() * 20000})
    }

    const eatEvents = this.currentData.events.filter(e => e.type === ev.EAT)
    for (const event of eatEvents) {
      let p = this.getAnimProgress(event.animData, this.progress)
      // This keep the apple gone even after the animation is over
      p = Math.min(1, Math.max(0, p))
      const shrinkP = unlerp(0.4, 1, p)
      const coord = event.coord
      const apple = appleMap[`${coord.x},${coord.y}`]
      apple.scale.set(1 - shrinkP)
    }
  }

  epsEqual(a: number, b: number, eps = 0.0001) {
    return Math.abs(a - b) < eps
  }

  turnHeadSprite(cuicui: BirdEffect, body: CoordDto[]) {
    const  {faceScaler, transformingScaler}  = cuicui

    const dx = body[0].x - body[1].x
    const dy = body[0].y - body[1].y
    const angle = Math.atan2(dy, dx)
    const scaler = faceScaler ?? transformingScaler
    scaler.rotation = angle
    scaler.scale.x = 1
    if (this.epsEqual(angle, Math.PI)){
      scaler.scale.x = -1
      scaler.rotation += Math.PI
    }
  }

  resetBird(cuicui: BirdEffect) {
    const {sprite, display, face, eating,dying, pIdx, transforming, body, hands} = cuicui
    if (sprite != null) {

      sprite.visible = true
      transforming.visible = false
      body.visible = true
      dying.visible = false
      display.scale.set(1)
      hands.forEach(h => h.scale.set(this.bodyPartScale))

    }

    if (face != null) {
      face.visible = true
      face.texture = PIXI.Texture.from(HEAD_FRAME[pIdx])
      eating.visible= false
      dying.visible= false
    }
  }

  updateBirdClocks(birdMap: Record<number, BirdEffect[]>) {
    for (const [bIdxStr, fullBird] of Object.entries(birdMap)) {
      for (let partIdx = 1; partIdx < fullBird.length; ++partIdx) {
        const part = fullBird[partIdx]
        const { hands } = part

        const prevPart = fullBird[partIdx - 1]
        const nextPart = fullBird[partIdx + 1]

        const angle1 = angleBetween(subtract(prevPart.display.position,part.display.position))
        const angle2 = nextPart == null
          ? angle1
          : angleBetween(subtract(nextPart.display.position, part.display.position))

        hands[0].rotation = angle1
        hands[1].rotation = angle2
      }
    }
  }

  updateBirds () {

    // Message reset
    for (const messageContainer of Object.values(this.messages)) {
      messageContainer.updateText('', 0, 0)
    }

    const birdMap: Record<number, BirdEffect[]> = {}
    const headPosMap: Record<number, CoordDto> = {}

    // Birds with initial state
    const birds = this.previousData.birds
    for (let bIdx = 0; bIdx < birds.length; ++bIdx) {
      const bird = birds[bIdx]
      if (!bird.alive) {
        continue
      }
      const pIdx = bird.owner
      const fullBird = []
      for (let partIdx = 0; partIdx < bird.body.length; ++partIdx) {
        const partCoord = bird.body[partIdx]
        const cuicui = partIdx === 0 ? this.getFromPool<BirdEffect>(`head_${pIdx}`) : this.getFromPool<BirdEffect>(`body_${pIdx}`)

        const {display, id} = cuicui

        this.birdContainers[bIdx].addChild(display)
        this.resetBird(cuicui)

        id.text = (bird.id).toString()
        cuicui.getTooltip = () => {
          return `Snakebot ${bird.id} (part ${partIdx})`
        }

        this.placeInGameZone(display, partCoord)

        if (partIdx === 0) {
          headPosMap[bIdx] = partCoord
          this.turnHeadSprite(cuicui, bird.body)
        } else if (partIdx === 1) {
          this.turnHeadSprite(cuicui, bird.body)
        }


        fullBird.push(cuicui)
      }
      birdMap[bIdx] = fullBird

    }

    // Move birds according to events
    const moveEvents = this.currentData.events.filter(e => e.type === ev.MOVE)
    for (const event of moveEvents) {
      let p = this.getAnimProgress(event.animData, this.progress)
      p = Math.min(1, Math.max(0, p))

      const fullBird = birdMap[event.birdId]
      if (event.growth > 0) {
        const newPart = this.getFromPool<BirdEffect>(`body_${event.playerId}`)
        const {sprite, display, face, id} = newPart
        this.birdContainers[event.birdId].addChild(display)

        if (sprite != null) {
          // sprite.tint = this.globalData.players[event.playerId].color
        }
        if (face != null) {
          face.visible = false
        }
        id.text = (event.birdId).toString()
        fullBird.push(newPart)
      }

      //Move each individual part one down
      for (let partIdx = fullBird.length - 1; partIdx >= 0; --partIdx) {
        let nextPos = partIdx === 0 ? event.coord : this.previousData.birds[event.birdId].body[partIdx - 1]
        const from = this.previousData.birds[event.birdId].body[partIdx]
        if (!from) {
          //spawning!
          this.placeInGameZone(fullBird[partIdx].display, nextPos)
          fullBird[partIdx].display.zIndex = partIdx
          continue
        }
        const to = nextPos
        const pos = lerpPosition(from, to, p)
        this.placeInGameZone(fullBird[partIdx].display, pos)
        const newBody = [event.coord, ...this.previousData.birds[event.birdId].body.slice(0, -1)]
        if (partIdx === 0) {
          headPosMap[event.birdId] = pos
          this.turnHeadSprite(fullBird[0], newBody)
        } else if (partIdx === 1) {
          this.turnHeadSprite(fullBird[0], newBody)
        }

        fullBird[partIdx].display.zIndex = -partIdx
      }
    }

    // Animate eating
    const eatEvents = this.currentData.events.filter(e => e.type === ev.EAT)
    for (const event of eatEvents) {
      let p = this.getAnimProgress(event.animData, this.progress)
      if (p <= 0 || p > 1) {
        continue
      }
      const head = birdMap[event.birdId][0]
      head.eating.visible = true
      head.face.visible = false
      setAnimationProgress(head.eating, p)
    }

    // Animate dying
    const deathEvents = this.currentData.events.filter(e => e.type === ev.BEHEAD || e.type === ev.DEATH)
    for (const event of deathEvents) {
      let p = this.getAnimProgress(event.animData, this.progress)
      if (p <= 0) {
        continue
      }
      p = Math.min(1, p)
      const head = birdMap[event.birdId][0]
      head.dying.visible = true
      head.face.visible = p < (13/24)
      setAnimationProgress(head.dying, p)

      if (event.type === ev.DEATH) {
        for (const part of birdMap[event.birdId].slice(1)) {
        //   part.display.scale.set(1 - p)
          part.dying.visible = true
          part.body.visible = p < (13/24)
          setAnimationProgress(part.dying, p)
        }
      } else if (event.type === ev.BEHEAD) {
        const neck = birdMap[event.birdId][1]
        neck.transforming.visible = true
        neck.sprite.visible = false
        neck.hands.forEach(h => h.scale.set((1 - unlerp(0, 0.5, p)) * this.bodyPartScale))

        setAnimationProgress(neck.transforming, p)
      }
    }

    // Move birds according to fall events
    const fallEvents = this.currentData.events.filter(e => e.type === ev.FALL)
    for (const event of fallEvents) {
      let p = this.getAnimProgress(event.animData, this.progress)
      if (p <= 0) {
        continue
      }
      p = Math.min(1, p)

      const fallP = unlerp(0, 0.5, p)
      const fullBird = birdMap[event.birdId]

      for (let partIdx = 0; partIdx < fullBird.length; ++partIdx) {
        let from = this.currentData.birds[event.birdId].bodyAfterMove[partIdx]
        if (!from) {
          //The bird has lost a piece before falling, remove effect
          fullBird[partIdx].display.visible = false
          fullBird[partIdx].busy = false
          birdMap[event.birdId].pop()
          continue
        }
        const to = { x: from.x, y: from.y + event.numberOfCells }
        const pos = lerpPosition(from, to, easeIn(fallP))
        this.placeInGameZone(fullBird[partIdx].display, pos)

        this.resetBird(fullBird[partIdx])

        if (partIdx === 0) {
          headPosMap[event.birdId] = pos
          const isGoingToDieAsAResult = this.currentData.events.find(e => e.type === ev.DEATH && e.birdId === event.birdId) != null
          if (isGoingToDieAsAResult) {
            fullBird[partIdx].face.texture = PIXI.Texture.from(FALL_HEAD_FRAME[fullBird[partIdx].pIdx])
          }
        }
      }
      const shakeP = easeOut(unlerpUnclamped(0.5, 1, p))
      if (shakeP > 0) {
        const offset = (Math.sin(shakeP * 10) * (1 - shakeP) * 0.5)
        for (let partIdx = 0; partIdx < fullBird.length; ++partIdx) {
          const cuicui = fullBird[partIdx]
          cuicui.display.x += 1 * offset * this.tileSize / 4
        }
      }
    }

    // Messages
    for (const message of this.currentData.messages) {
      if (message.text != '') {
        const birdId = message.birdId
        const messageContainer = this.messages[birdId]
        const coord =  headPosMap[birdId]
        if (coord == null) {
          console.log('Unexpected missing snakebot')
          continue
        }
        const boardPos = this.toBoardPos(coord)
        let globalPoint = this.gameZone.toGlobal(boardPos)
        let containerPoint = this.container.toLocal(globalPoint)
        messageContainer.updateText(message.text, containerPoint.x, containerPoint.y)
      }
    }

    this.updateBirdClocks(birdMap)
  }

  getAnimProgress ({ start, end }: AnimData, progress: number) {
    return unlerpUnclamped(start, end, progress)
  }

  resetEffects () {
    for (const type in this.pool) {
      for (const effect of this.pool[type]) {
        effect.display.visible = false
        effect.busy = false
      }
    }
  }

  animateScene (delta: number) {
    this.time += delta
    for (const player of this.globalData.players) {
      for (const birdData of this.globalData.playerBirds[player.index]) {
        const message = this.messages[birdData.id]
        renderMessageContainer.bind(this)(message, player.index, delta)
      }
    }

    this.animateMarkers(delta)
    this.animateApples(delta)
  }


  animateApples(delta: number) {
    const maxAnimating = 3
    let currentlyAnimating = this.apples.filter(a => a.effect.sparks.playing).length

    for (const apple of this.apples) {
      const scaleFactor = 1 + 0.05 * Math.sin(this.time / 200)
      apple.effect.display.rotation = scaleFactor - 1
      if (apple.effect.display.visible && !apple.effect.sparks.playing) {
        if (apple.countdown <= 0) {
          if (currentlyAnimating < maxAnimating) {
            apple.effect.sparks.gotoAndPlay(0)
            currentlyAnimating++
          }
          apple.countdown = 10000 + Math.floor(Math.random() * 15000)
        } else {
          apple.countdown -= delta
        }
      }

    }
  }

  animateMarkers(delta: number) {
    for (const markEffect of this.currentMarks) {
      const mark = markEffect.sprite
      const localTime = this.time + ((mark.tint === this.globalData.players[0].color) ? 0 : Math.PI * 100)
      const scaleFactor = 1 + 0.1 * Math.sin(localTime / 100)
      mark.scale.set(this.tileSize / 95 * scaleFactor)
      if (api.options.showMarks) {
        const stepFactor = Math.pow(0.99, delta)
        const targetMessageAlpha = (api.options.showMarks === 3 || api.options.showMarks === markEffect.playerIdx + 1)
          ? 0.9
          : 0
        mark.alpha = mark.alpha * stepFactor + targetMessageAlpha * (1 - stepFactor)
      }
    }
  }

  asLayer (func: ContainerConsumer): PIXI.Container {
    const layer = new PIXI.Container()
    func.bind(this)(layer)
    return layer
  }

  toBoardPos (coord: PIXI.IPointData) {
    return {
      x: coord.x * this.tileSize,
      y: coord.y * this.tileSize
    }
  }

  placeInGameZone(display: PIXI.DisplayObject, coord: IPointData, center = true) {
    const pos = this.toBoardPos(coord)
    display.position.set(pos.x + (center ? this.tileSize / 2 : 0), pos.y + (center ? this.tileSize / 2 : 0))
  }

  initBirdLayer(layer: PIXI.Container) {
    const hueStep = [0.1, 0.2]
    this.birdContainers = []
    for (const player of this.globalData.players) {
      let idx = 0
      const birds = this.globalData.playerBirds[player.index]
      for (const birdData of birds) {
        const birdContainer = new PIXI.Container()
        birdContainer.sortableChildren = true


        const hueFilter = new PIXI.Filter(undefined, fragment, {
          hueShift: -hueStep[player.index]*(birds.length/2) + (hueStep[player.index] * (idx++))
        })
        birdContainer.filters = [hueFilter]
        layer.addChild(birdContainer)
        this.birdContainers.push(birdContainer)
      }
    }
  }



  initMap (layer: PIXI.Container) {
    for (let y = 0; y < this.globalData.height; y++) {
      for (let x = 0; x < this.globalData.width; x++) {
        const type = this.globalData.map[y][x]
        if (type === 0) {
          continue
        }
        const neighbours = {
          left: this.globalData.map[y][x - 1] || 0,
          top: this.globalData.map[y - 1]?.[x] || 0,
          right: this.globalData.map[y][x + 1] || 0,
          bottom: this.globalData.map[y + 1]?.[x] || 0,
          bottomLeft: this.globalData.map[y + 1]?.[x - 1] || 0,
          bottomRight: this.globalData.map[y + 1]?.[x + 1] || 0,
        }

        let texture = null
        if (neighbours.top && neighbours.bottom) {
          texture = choice(TILES.MIDDLE)
        } else if (neighbours.bottom && !neighbours.top) {
          texture = choice(TILES.TOP)
        } else if (!neighbours.bottom && neighbours.top) {
          texture = choice(TILES.BOTTOM)
        } else if (!neighbours.bottomLeft && !neighbours.bottomRight) {
          texture = choice(TILES.PLATFORM)
        } else {
          texture = choice(TILES.TOP)
        }

        const sprite = PIXI.Sprite.from(texture)
        this.placeInGameZone(sprite, { x, y }, false)
        sprite.scale.set(fitAspectRatio(sprite.width, sprite.height, this.tileSize, Infinity))
        sprite.anchor.set(0, 31/122)
        if (Math.random() < 0.5) {
          sprite.scale.x *= -1
          sprite.anchor.x = 1
        }
        layer.addChild(sprite)
      }
    }
  }

  reinitScene (container: PIXI.Container, canvasData: CanvasInfo) {
    (window as any).g = new PIXI.Graphics()
    this.time = 0
    this.oversampling = canvasData.oversampling
    this.container = container
    this.pool = {}
    this.canvasData = canvasData

    this.currentMarks = []

    this.tileSize = Math.min(GAME_ZONE_RECT.w / this.globalData.width, GAME_ZONE_RECT.h / this.globalData.height)

    const sprite = PIXI.Sprite.from(BODY_FRAME[0])
    this.bodyPartScale = fitAspectRatio(sprite.width, sprite.height, this.tileSize, this.tileSize) * BODY_PART_OVERSCALE

    const tooltipLayer = this.tooltipManager.reinit()

    const background = PIXI.Sprite.from('Background_3.jpg')
    background.width = WIDTH
    background.height = HEIGHT
    const gameZone = new PIXI.Container()

    const mapLayer = this.asLayer(this.initMap)
    this.birdLayer = this.asLayer(this.initBirdLayer)
    this.appleLayer = new PIXI.Container()
    this.markLayer = new PIXI.Container()
    const hudLayer = this.asLayer(this.initHud)
    const messageLayer = this.asLayer(initMessages)

    this.birdLayer.sortableChildren = true

    gameZone.addChild(this.appleLayer)
    gameZone.addChild(this.birdLayer)
    gameZone.addChild(mapLayer)
    gameZone.addChild(this.markLayer)

    gameZone.x = GAME_ZONE_RECT.x
    gameZone.y = GAME_ZONE_RECT.y
    const gameWidth = this.globalData.width * this.tileSize
    const gameHeight = this.globalData.height * this.tileSize
    gameZone.x += (GAME_ZONE_RECT.w - gameWidth) / 2
    gameZone.y += (GAME_ZONE_RECT.h - gameHeight) / 2
    this.gameZone = gameZone

    container.addChild(background)
    container.addChild(gameZone)
    container.addChild(hudLayer)
    container.addChild(messageLayer)
    container.addChild(tooltipLayer)

    // // @ts-ignore
    // const f = new PIXI.filters.OutlineFilter(4, 0x0, 1)
    // this.birdLayer.filters = [f]

    background.interactiveChildren = false
    tooltipLayer.interactiveChildren = false
    hudLayer.interactiveChildren = false
    messageLayer.interactiveChildren = false

    container.interactive = true
    container.on('mousemove', (event) => {
      this.tooltipManager.moveTooltip(event)
    })

    this.tooltipManager.registerGlobal((data) => {
      const pos = data.getLocalPosition(gameZone)
      const x = Math.floor(pos.x / this.tileSize)
      const y = Math.floor(pos.y / this.tileSize)

      if (x < 0 || x >= this.globalData.width || y < 0 || y >= this.globalData.height) {
        return null
      }
      const blocks = []
      blocks.push(`(${x}, ${y})`)
      return blocks.join('\n--------\n')
    })

  }

  easeOutElastic (x: number): number {
    const c4 = (2 * Math.PI) / 3

    return x === 0
      ? 0
      : x === 1
        ? 1
        : Math.pow(2, -10 * x) * Math.sin((x * 10 - 0.75) * c4) + 1
  }

  handleGlobalData (players: PlayerInfo[], raw: string): void {
    const globalData = parseGlobalData(raw)
    api.options.meInGame = !!players.find(p => p.isMe)


    this.globalData = {
      ...globalData,
      players: players,
      playerCount: players.length
    }
  }


  handleFrameData (frameInfo: FrameInfo, raw: string): FrameData {
    const dto = parseData(raw, this.globalData)
    const previousFrame = last(this.states)
    const frameData: FrameData = {
      ...dto,
      previous: null,
      frameInfo,
      apples: (previousFrame ? previousFrame.apples : this.globalData.apples),
      birds: (previousFrame ? previousFrame.birds : this.globalData.playerBirds.map((birds,idx)=> (birds.map(b=>({...b, owner: idx, bodyAfterMove: b.body, alive: true})))).flat())
    }

    // Deep copy
    frameData.birds = frameData.birds.map(bird => ({...bird, body: [...bird.body]}))
    frameData.apples = frameData.apples.map(a => ({...a}))

    frameData.previous = previousFrame ?? frameData


    // [*** NOTE: This next bit of code was AI generated as a test, I have reviewed it, it's good ***]

    // --- Normalize bird events to fit in 1s, preserving simultaneous events ---
    const eventsByBird: Record<number, EventDto[]> = {}
    for (const event of dto.events) {
      if (event.birdId == null) continue
      if (!eventsByBird[event.birdId]) eventsByBird[event.birdId] = []
      eventsByBird[event.birdId].push(event)
    }

    for (const [birdIdStr, events] of Object.entries(eventsByBird)) {
      // Sort by original start time
      events.sort((a, b) => a.animData.start - b.animData.start)

      // Group events by original start time
      const groups: EventDto[][] = []
      let currentGroup: EventDto[] = []
      let lastStart = null
      for (const event of events) {
        if (event.animData.start !== lastStart) {
          if (currentGroup.length > 0) groups.push(currentGroup)
          currentGroup = [event]
          lastStart = event.animData.start
        } else {
          currentGroup.push(event)
        }
      }
      if (currentGroup.length > 0) groups.push(currentGroup)

      // Compute total active time (sum of group durations)
      const groupDurations = groups.map(g => {
        // duration = max of events in the group
        return Math.max(...g.map(e => e.animData.end - e.animData.start))
      })
      const totalActive = groupDurations.reduce((a, b) => a + b, 0)
      if (totalActive === 0) continue

      // Reassign start/end so groups are contiguous, events in the same group keep same relative duration
      let current = 0
      for (let i = 0; i < groups.length; i++) {
        const group = groups[i]
        const proportion = groupDurations[i] / totalActive
        const groupStart = current
        const groupEnd = current + proportion
        for (const event of group) {
          const relDuration = event.animData.end - event.animData.start
          const relProportion = relDuration / groupDurations[i]
          event.animData.start = groupStart
          event.animData.end = groupStart + relProportion * (groupEnd - groupStart)
        }
        current = groupEnd
      }
    }
    // [*** End of block ***]


    for (const event of dto.events) {
      if (event.type === ev.MOVE) {
        const bird = frameData.birds[event.birdId]
        bird.body.unshift(event.coord)
        if (event.growth === 0) {
          bird.body.pop()
        }
        bird.bodyAfterMove = bird.body.map(pos => ({...pos}))
      } else if (event.type === ev.FALL) {
        const bird = frameData.birds[event.birdId]
        bird.body = bird.body.map(pos => ({...pos, y: pos.y + event.numberOfCells}))
      } else if (event.type === ev.EAT) {
        const bird = frameData.birds[event.birdId]
        frameData.apples = frameData.apples.filter(apple => !(apple.x === event.coord.x && apple.y === event.coord.y))
      } else if (event.type === ev.DEATH) {
        const bird = frameData.birds[event.birdId]
        if (bird.alive === false) {
          console.log('ALRADY DEAD: ', event.birdId, frameInfo.number)
        }
        bird.alive = false
      } else if (event.type === ev.BEHEAD) {
        const bird = frameData.birds[event.birdId]
        bird.body.shift()
        bird.bodyAfterMove = [...bird.body]
      }
    }

    this.states.push(frameData)
    return frameData
  }
}

