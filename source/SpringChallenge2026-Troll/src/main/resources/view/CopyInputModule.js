import { WIDTH, HEIGHT } from './core/constants.js'
import { EntityFactory } from './entity-module/EntityFactory.js'

export class CopyInputModule {
  static get moduleName() {
    return 'inputmodule'
  }
  static get dependencies() {
    return ['entitymodule']
  }

  getInput(turn, swapPlayer) {
    let result = this.globalData
    if (swapPlayer) {
        result = ""
        let inGrid = false
        for (let c of this.globalData) {
            if (c === '\n') inGrid = true
            let rep = c
            if (inGrid && c === '0') rep = '1'
            if (inGrid && c === '1') rep = '0'
            result += rep
        }
    }
    let turnLines = this.turnData[turn].split('\n')
    let inventory = [turnLines.shift(), turnLines.shift()]
    result += '\n' + inventory[swapPlayer ? 1 : 0]
    result += '\n' + inventory[swapPlayer ? 0 : 1]
    let treeCount = +turnLines.shift()
    result += '\n' + treeCount
    for (let i = 0; i < treeCount; i++) {
        result += '\n' + turnLines.shift()
    }
    let unitCount = +turnLines.shift()
    result += '\n' + unitCount
    for (let i = 0; i < unitCount; i++) {
        let line = turnLines.shift()
        let parts = line.split(' ')
        if (swapPlayer) parts[1] = 1 - parts[1]
        result += '\n' + parts.join(" ")
    }
    return result
  }

  updateScene(previousData, currentData, progress) {
    let turn = currentData
    if (progress < 1) turn--
    if (this.currentFrame == turn) return
    this.currentFrame = turn

    let avatar = entityModule.entities.get(13)
    avatar.container.interactive = true
    avatar.interactive = true
    avatar.container.click = (ev) => {
        navigator.clipboard.writeText(this.getInput(turn, false))
    }

     avatar = entityModule.entities.get(14)
     avatar.container.interactive = true
     avatar.interactive = true
     avatar.container.click = (ev) => {
         navigator.clipboard.writeText(this.getInput(turn, true))
     }

     this.p1Area.value = this.getInput(turn, false)
     this.p2Area.value = this.getInput(turn, true)
  }

  handleFrameData(frameInfo, data = []) {
    this.turnData[frameInfo.number] = data
    return frameInfo.number
  }

  handleGlobalData (players, globalData) {
    this.globalData = globalData
    this.turnData = {}
    this.currentFrame = -1

    // blatant theft from https://github.com/reCurs3/codingame-chess/blob/master/src/main/resources/view/ChessViewerModule.js#L92
    var settings = document.getElementsByClassName('settings_panel_form')[0];
    var createOption = (headerText, elementName, id) => {
        var element = document.getElementById(id);
        if (element) {
            document.getElementById(id + '-header').innerText = headerText;
            return element;
        }

        var option = document.createElement('div');
        option.className = 'settings_option';
        var header = document.createElement('h3');
        header.innerText = headerText;
        option.appendChild(header);
        element = document.createElement(elementName);
        element.className = 'settings_button';
        element.id = id;
        header.id = id + '-header';
        element.style.width = '100%';
        element.style.background = 'transparent';
        element.style.color = '#b3b9ad';
        element.style.padding = '3px';
        element.style.resize = 'none';
        element.style.height = '70px';
        element.style.whiteSpace = 'pre';
        element.value = '';
        element.readOnly = true;
        element.onfocus = function() { this.select(); };
        element.onfocusout = function() { this.selectionStart = this.selectionEnd; };
        option.appendChild(element);
        settings.appendChild(option);
        return element;
    };

    this.p1Area = createOption('Input for ' + players[0].name, 'textarea', 'turn_input_p1');
    this.p2Area = createOption('Input for ' + players[1].name, 'textarea', 'turn_input_p2');
  }

  reinitScene(container) {
    container.interactive = true
  }
}