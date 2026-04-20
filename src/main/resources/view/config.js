import { EndScreenModule } from './endscreen-module/EndScreenModule.js'
import { ViewModule, api } from './graphics/ViewModule.js'

export const modules = [
  ViewModule,
  EndScreenModule
]

export const playerColors = [
  '#ef26c0', // shocking pink
  '#01ff09', // full green

  '#6ac371', // mantis green
  '#ff1d5c', // radical red
  '#22a1e4', // curious blue
  '#de6ddf', // lavender pink
  '#9975e2', // medium purple
  '#3ac5ca', // scooter blue
  '#ff0000' // solid red
]

export const gameName = 'Winter2025'

export const stepByStepAnimateSpeed = 3

export const options = [
  {
    title: 'HIDE RANKING',
    get: function () {
      return api.options.debugMode
    },
    set: function (value) {
      api.setDebugMode(value)
    },
    values: {
      'ON': true,
      'OFF': false
    },
  },
  {
    title: 'MY MESSAGES',
    get: function () {
      return api.options.showMyMessages
    },
    set: function (value) {
      api.options.showMyMessages = value
    },
    enabled: function () {
      return api.options.meInGame
    },
    values: {
      'ON': true,
      'OFF': false
    }
  }, {
    title: 'OTHERS\' MESSAGES',
    get: function () {
      return api.options.showOthersMessages
    },
    set: function (value) {
      api.options.showOthersMessages = value
    },

    values: {
      'ON': true,
      'OFF': false
    }
  },
  {
    title: 'MARKS',
    get: function () {
      return api.options.showMarks
    },
    set: function (value) {
      api.options.showMarks = value
    },
    values: { // Can't be zero for some reason
      'ON': 3,
      '1': 1,
      '2': 2,
      'OFF': -1
    }
  }
]
