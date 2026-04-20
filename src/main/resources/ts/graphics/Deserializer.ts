import { GlobalBirdDto, EventDto, FrameDataDto, GlobalData, GlobalDataDto, MessageDto, CoordDto } from '../types.js'
import ev from './events.js'

const MAIN_SEPERATOR = '|'

function splitLine (str: string) {
  return str.length === 0 ? [] : str.split(' ')
}


export function parseData (unsplit: string, globalData: GlobalData): FrameDataDto {
  const raw = unsplit.split(MAIN_SEPERATOR)
  let idx = 0

  const events: EventDto[] = []
  const eventCount = +raw[idx++]
  for (let i = 0; i < eventCount; ++i) {
    const type = +raw[idx++]
    const start = +raw[idx++]
    const end = +raw[idx++]

    const params = splitLine(raw[idx++]).map(v=>+v)
    const animData = { start, end }
    const event: EventDto = {
      type,
      animData,
      params
    }
    if (event.type === ev.MOVE) {
      event.birdId = +params[0]
      event.playerId = +params[1]
      event.coord = {x: +params[2], y: +params[3]}
      event.growth = +params[4]
    } else if (event.type === ev.FALL) {
      event.birdId = +params[0]
      event.playerId = +params[1]
      event.numberOfCells = +params[2]
    } else if (event.type === ev.EAT) {
      event.birdId = +params[0]
      event.playerId = +params[1]
      event.growth = +params[2]
      event.coord = {x: +params[3], y: +params[4]}
    } else if (event.type === ev.DEATH) {
      event.birdId = +params[0]
      event.playerId = +params[1]
    } else if (event.type === ev.BEHEAD) {
      event.birdId = +params[0]
      event.playerId = +params[1]
    }

    events.push(event)
  }

  const ms = []
  const marks: CoordDto[][] = []
  for (let pIdx = 0; pIdx < 2; ++pIdx) {
    ms.push(+raw[idx++])

    const playerMarks: CoordDto[] = []
    const markCount = +raw[idx++]
    for (let mIdx = 0; mIdx < markCount; ++mIdx) {
      playerMarks.push(parseCoord(raw[idx++]))
    }
    marks.push(playerMarks)
  }
  const messages: MessageDto[] = []
  const messageCount = +raw[idx++]

  for (let i = 0; i < messageCount; ++i) {
    const birdId = +raw[idx++]
    const playerIdx = +raw[idx++]
    const message = raw[idx++]
    messages.push({ playerIdx, birdId, text: message })
  }

  return {
    events,
    ms,
    messages,
    marks
  }
}


export function parseGlobalData (unsplit: string): GlobalDataDto {
  const raw = unsplit.split(MAIN_SEPERATOR)
  let idx = 0
  const width = +raw[idx++]
  const height = +raw[idx++]
  const map = []
  for (let y = 0; y < height; y++) {
    const row = []
    for (let x = 0; x < width; x++) {
      const type = parseInt(raw[idx++])
      row.push(type)
    }
    map.push(row)
  }
  const appleCount = +raw[idx++]
  const apples = []
  for (let i = 0; i < appleCount; ++i) {
    const x = +raw[idx++]
    const y = +raw[idx++]
    apples.push({ x, y })
  }

  const playerBirds = []
  for (let pIdx = 0; pIdx < 2; ++pIdx) {
    const birds: GlobalBirdDto[] = []
    const birdCount = +raw[idx++]
    for (let bIdx = 0; bIdx < birdCount; ++bIdx) {
      const id = +raw[idx++]
      const body = []
      const bodySize = +raw[idx++]
      for (let i = 0; i < bodySize; ++i) {
        const x = +raw[idx++]
        const y = +raw[idx++]
        body.push({ x, y })
      }
      birds.push({id, body})
    }
    playerBirds.push(birds)
  }


  return {
    width,
    height,
    map,
    apples,
    playerBirds
  }
}

function parseCoord (coord: string) {
  const [x, y] = coord.split(' ').map(x => +x)
  return { x, y }
}

