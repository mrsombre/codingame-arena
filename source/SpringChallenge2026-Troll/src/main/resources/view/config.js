import { CompressionModule } from './CompressionModule.js';
import { GraphicEntityModule } from './entity-module/GraphicEntityModule.js';
import { TooltipModule } from './tooltip-module/TooltipModule.js';
import { ToggleModule } from './toggle-module/ToggleModule.js'
import { EndScreenModule } from './endscreen-module/EndScreenModule.js';
import { ExplosionModule } from './ExplosionModule.js';
import { CopyInputModule } from './CopyInputModule.js';

export const modules = [
	CompressionModule,
	GraphicEntityModule,
	TooltipModule,
	ToggleModule,
	EndScreenModule,
	ExplosionModule,
	CopyInputModule
];

export const options = [
  ToggleModule.defineToggle({
    toggle: 'debug',
    title: 'DEBUG',
    values: {
      'ON': true,
      'OFF': false
    },
    default: false
  }),
  ToggleModule.defineToggle({
    toggle: 'darkMode',
    title: 'THEME',
    values: {
      'LIGHT': false,
      'DARK': true,
    },
    default: false
  }),]