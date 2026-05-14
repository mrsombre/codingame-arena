import { Drawer } from './core/Drawer.js'

let replacedDrawer = false

export class CompressionModule {
    constructor(assets) {
        if (replacedDrawer) return
        replacedDrawer = true
        const originalParseFrame = Drawer.prototype.parseFrame;
        Drawer.prototype.parseFrame = function (...args) {
            let frame = args[0]

            // restore full state from delta
            const resources = ['PLUM', 'LEMON', 'APPLE', 'BANANA', 'IRON', 'WOOD']
            const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
            let prevState = {}
            if (args[1].length > 0) prevState = args[1][args[1].length - 1].data.diffState || {}
            let currentState = {}
            for (const [key, value] of Object.entries(prevState)) {
                let obj = {...value}
                if (obj.objectType == 'plant') {
                    if (obj.health <= 0) continue
                    if (obj.currentCooldown > 0) obj.currentCooldown--
                    if (obj.currentCooldown == 0 && obj.fruits < 3) {
                        obj.currentCooldown = obj.totalCooldown
                        if (obj.size < 4) obj.size++
                        else obj.fruits++
                    }
                }
                currentState[key] = obj
            }

            if (frame.diff !== undefined && frame.diff !== "") {
                for (let diff of frame.diff.split(';')) {
                    let parts = diff.split(' ')
                    let id = +parts[0]
                    let obj = {...(currentState || {})[id]} || {}
                    if (diff === '867 x7 y5')
                        console.log('test')
                    if (parts[1] === 'W') {
                        obj = {
                            objectType: "unit",
                            id: alphabet.indexOf(parts[2][0]),
                            x: alphabet.indexOf(parts[2][1]),
                            y: alphabet.indexOf(parts[2][2]),
                            playerId: alphabet.indexOf(parts[2][3]),
                            moveSpeed: alphabet.indexOf(parts[2][4]),
                            carryCapacity: alphabet.indexOf(parts[2][5]),
                            harvestPower: alphabet.indexOf(parts[2][6]),
                            chopPower: alphabet.indexOf(parts[2][7]),
                            inventory: [0,0,0,0,0,0],
                        }
                    } else if (parts[1] === 'P') {
                        obj = {
                            objectType: "plant",
                            x: alphabet.indexOf(parts[2][0]),
                            y: alphabet.indexOf(parts[2][1]),
                            type: resources[alphabet.indexOf(parts[2][2])],
                            size: alphabet.indexOf(parts[2][3]),
                            fruits: 0,
                            currentCooldown: alphabet.indexOf(parts[2][4]),
                            health: alphabet.indexOf(parts[2][5]),
                            totalCooldown: alphabet.indexOf(parts[2][6]),
                        }
                        if (obj.objectType === 'plant' && obj.size >= 4) {
                            obj.fruits = obj.size - 4
                            obj.size = 4
                        }
                    } else {
                        parts.shift()
                        for (let update of parts) {
                            if (update[0] == 's') {
                                obj.size = alphabet.indexOf(update[1])
                                if (obj.size >= 4) {
                                    obj.fruits = obj.size - 4
                                    obj.size = 4
                                }
                            } else if (update[0] == 'h') {
                                obj.health = alphabet.indexOf(update[1])
                            } else if (update[0] == 'c') {
                                obj.currentCooldown = alphabet.indexOf(update[1])
                            } else if (update[0] == 'x') {
                                obj.x = alphabet.indexOf(update[1])
                            } else if (update[0] == 'y') {
                                obj.y = alphabet.indexOf(update[1])
                            } else {
                                obj.inventory[+update[0]] = alphabet.indexOf(update[1])
                            }
                        }
                    }
                    currentState[id] = obj
                }
            }

            let plantIds = Object.keys(currentState).filter(k => currentState[k].objectType === 'plant').sort((a, b) => a - b)
            let unitIds = Object.keys(currentState).filter(k => currentState[k].objectType === 'unit').sort((a, b) => a - b)

            // set tooltips
            if (!frame.tooltips) frame.tooltips = [{}]
            for (let id of plantIds) {
                let plant = currentState[id]
                let tip = `${plant.type}\nsize: ${plant.size}\nhealth: ${plant.health}\nfruits: ${plant.fruits}\ncooldown: ${plant.currentCooldown}`
                if (plant.health <= 0) tip = ''
                frame.tooltips[0][id] = tip
            }
            plantIds = Object.keys(currentState).filter(k => currentState[k].objectType === 'plant' && currentState[k].health > 0).sort((a, b) => a - b)
            for (let id of unitIds) {
                let unit = currentState[id]
                let tip = `TROLL\nid: ${unit.id}\nmovementSpeed: ${unit.moveSpeed}\ncarryCapacity: ${unit.carryCapacity}\nharvestPower: ${unit.harvestPower}\nchopPower: ${unit.chopPower}`
                for (let i = 0; i < unit.inventory.length; i++) {
                    if (unit.inventory[i] !== 0) tip += "\n" + resources[i] + ": " + unit.inventory[i]
                }
                frame.tooltips[0][id] = tip
            }

            // restore player input
            if (!!frame.inputmodule && unitIds.length > 0) {
                frame.inputmodule += "\n" + plantIds.length
                for (let id of plantIds) {
                    let plant = currentState[id]
                    frame.inputmodule += `\n${plant.type} ${plant.x} ${plant.y} ${plant.size} ${plant.health} ${plant.fruits} ${plant.currentCooldown}`
                }
                frame.inputmodule += "\n" + unitIds.length
                for (let id of unitIds) {
                    let unit = currentState[id]
                    frame.inputmodule += `\n${unit.id} ${unit.playerId} ${unit.x} ${unit.y} ${unit.moveSpeed} ${unit.carryCapacity} ${unit.harvestPower} ${unit.chopPower} ${unit.inventory.join(" ")}`
                }
            }

            frame = originalParseFrame.apply(this, args);
            frame.data.diffState = currentState
            return frame
        };
    }

    static get moduleName() {
        return 'compressionmodule'
    }
    static get dependencies() {
        return []
    }

    updateScene(previousData, currentData, progress) { }

    handleFrameData(frameInfo, data = []) { }

    reinitScene(container) { }
}