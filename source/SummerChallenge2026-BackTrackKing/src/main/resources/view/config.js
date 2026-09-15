import { EndScreenModule } from './endscreen-module/EndScreenModule.js'
import { ViewModule, api } from './graphics/ViewModule.js'

export const modules = [
  ViewModule,
  EndScreenModule
]

export const playerColors = [
      '#ff1d5c', // radical red
      '#22a1e4', // curious blue
      '#ff8f16', // west side orange
      '#6ac371', // mantis green
      '#9975e2', // medium purple
      '#3ac5ca', // scooter blue
      '#de6ddf', // lavender pink
      '#ff0000'  // solid red
    ];

export const gameName = 'Trains2026'

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
  }, {
  title: 'REGIONS',
  get: function () {
    return api.options.territories
  },
  set: function (value) {
    api.options.territories = value
  },
  values: {
    'HIDE': 0,
    'SHOW': 1
  }
}, {
    title: 'ACTIVE CONNECTIONS',
    get: function () {
      return api.options.connectionDisplay
    },
    set: function (value) {
      api.options.connectionDisplay = value
      api.renderUpdate()
    },
    values: {
      'TRAINS': 0,
      'ARROWS': 1
    }
},{
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
  }, {
  title: 'TRACK CONTRAST',
  get: function () {
    return api.options.trackConstrast
  },
  set: function (value) {
    api.options.trackConstrast = value
  },
  values: {
    'DEFAULT': 0,
    'A': 1,
    'B': 2,
  }
}
]
