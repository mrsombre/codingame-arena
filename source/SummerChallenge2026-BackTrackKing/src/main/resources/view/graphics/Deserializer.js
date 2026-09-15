import ev from './events.js';
function splitLine(str) {
    return str.length === 0 ? [] : str.split(' ');
}
export function parseData(unsplit, globalData) {
    let idx = 0;
    const raw = unsplit.split('|');
    const events = [];
    const eventCount = +raw[idx++];
    for (let i = 0; i < eventCount; ++i) {
        const type = +raw[idx++];
        const start = +raw[idx++];
        const end = +raw[idx++];
        const params = splitLine(raw[idx++]).map(v => +v);
        const animData = { start, end };
        const event = {
            type,
            animData,
            params
        };
        if (event.type === ev.BUILD) {
            event.playerIdx = +params[0];
            event.coord = { x: +params[1], y: +params[2] };
            event.cost = +params[3];
        }
        else if (event.type === ev.DISRUPT) {
            event.playerIdx = +params[0];
            event.zoneId = +params[1];
        }
        else if (event.type === ev.EARN_POINTS) {
            event.playerIdx = +params[0];
            event.score = +params[1];
        }
        else if (event.type === ev.INK) {
            event.zoneId = +params[0];
            event.coords = [];
            for (let j = 1; j < params.length; j += 2) {
                const coord = { x: params[j], y: params[j + 1] };
                event.coords.push(coord);
            }
        }
        else if (event.type === ev.CONNECTION_GAINED) {
            event.coords = [];
            event.fromTownId = +params[0];
            event.toTownId = +params[1];
            for (let j = 2; j < params.length; j += 2) {
                const coord = { x: params[j], y: params[j + 1] };
                event.coords.push(coord);
            }
        }
        else if (event.type === ev.CONNECTION_LOST) {
            event.coords = [];
            event.fromTownId = +params[0];
            event.toTownId = +params[1];
        }
        else if (event.type === ev.BUMP) {
            event.coord = { x: +params[0], y: +params[1] };
        }
        else if (event.type === ev.POI_BUMP) {
            event.coord = { x: +params[0], y: +params[1] };
            event.playerIdx = +params[2];
        }
        events.push(event);
    }
    let messages = [];
    for (let i = 0; i < globalData.playerCount; ++i) {
        const message = raw[idx++];
        messages.push(message);
    }
    return {
        events,
        messages
    };
}
export function parseGlobalData(unsplit) {
    var _a;
    const raw = unsplit.split('|');
    let idx = 0;
    const width = +raw[idx++];
    const height = +raw[idx++];
    const passiveIncome = +raw[idx++];
    const zoneMap = {};
    const tiles = [];
    for (let y = 0; y < height; y++) {
        for (let x = 0; x < width; x++) {
            const tokens = raw[idx++].split(' ');
            const tile = {
                coord: { x, y },
                zoneId: +tokens[0],
                type: +tokens[1],
            };
            tiles.push(tile);
            zoneMap[tile.zoneId] = (_a = zoneMap[tile.zoneId]) !== null && _a !== void 0 ? _a : { coords: [], id: tile.zoneId, owner: -1 };
            const zone = zoneMap[tile.zoneId];
            zone.coords.push(tile.coord);
        }
    }
    const towns = [];
    const townCount = +raw[idx++];
    for (let i = 0; i < townCount; i++) {
        const tokens = splitLine(raw[idx++]);
        const id = +tokens[0];
        const coord = { x: +tokens[1], y: +tokens[2] };
        const desiredConnections = tokens[3] === 'x' ? [] : tokens[3].split(',').map(x => +x);
        towns.push({ id, coord, desiredConnections });
    }
    towns.sort((a, b) => a.id - b.id);
    return {
        width,
        height,
        tiles,
        zoneMap,
        passiveIncome,
        zones: Object.values(zoneMap).sort((a, b) => a.id - b.id),
        towns
    };
}
function parseCoord(coord) {
    const [x, y] = coord.split(' ').map(x => +x);
    return { x, y };
}
