export const GAME_ZONE_RECT = {
    x: 53, y: 159, w: 1820, h: 909
};
export const GRID_LINE_WIDTH = 0;
export const ZONE_LINE_WIDTH = 4;
export const RAIL_DATA = {
    byTrackCode: (n, withEnds = false) => {
        if (withEnds) {
            switch (n) {
                case 0b0001:
                    return { texture: 'end', rotation: 0 };
                case 0b0100:
                    return { texture: 'end', rotation: Math.PI };
                case 0b0010:
                    return { texture: 'end', rotation: -Math.PI / 2 };
                case 0b1000:
                    return { texture: 'end', rotation: Math.PI / 2 };
                default:
                    break;
            }
        }
        switch (n) {
            case 0b0000:
            case 0b1111:
                return { texture: 'all', rotation: 0 };
            case 0b0001:
            case 0b0100:
            case 0b0101:
                return { texture: 'straight', rotation: 0 };
            case 0b0010:
            case 0b1000:
            case 0b1010:
                return { texture: 'straight', rotation: Math.PI / 2 };
            case 0b0011:
                return { texture: 'turn', rotation: -Math.PI / 2 };
            case 0b0110:
                return { texture: 'turn', rotation: Math.PI };
            case 0b1100:
                return { texture: 'turn', rotation: Math.PI / 2 };
            case 0b1001:
                return { texture: 'turn', rotation: 0 };
            case 0b1011:
                return { texture: 't', rotation: 0 };
            case 0b0111:
                return { texture: 't', rotation: -Math.PI / 2 };
            case 0b1110:
                return { texture: 't', rotation: Math.PI };
            case 0b1101:
                return { texture: 't', rotation: Math.PI / 2 };
        }
    }
};
export const HUD_COLORS = [
    '#ff1d5c',
    '#22a1e4',
];
export const AVATAR_RECT = {
    x: 16, y: 15, w: 93, h: 92
};
export const NAME_RECT = {
    x: 124, y: 35, w: 286, h: 33
};
export const SCORE_RECT = {
    x: 454, y: 20, w: 147, h: 61
};
export const MESSAGE_RECT = {
    x: 37, y: 183, w: 215, h: 827
};
export const CONVEYOR_SCALE = 0.6;
export const CONVEYOR_HEIGHT = 32;
export const CONVEYOR_WIDTH = 27;
export const POI_FRAMES = ['poi', 'poi2', 'poi3'];
export const TOWN_FRAME = 'town';
export const MOUNTAIN_FRAMES_1x1 = [
    'Montagne1',
    'Montagne2',
    'Montagne3'
];
export const MOUNTAIN_FRAMES_2x2 = [
    'Montagne4',
    'Montagne7',
];
export const MOUNTAIN_FRAMES_2x1 = [
    'Montagne5',
    'Montagne6',
];
export const GRASS_FRAMES = ['Herbe1', 'Herbe2', 'Herbe3'];
export const TRAIN_FRAME = 'train.png';
export const BRUSH_FRAMES = [[
        'Anime_R0001',
        'Anime_R0003',
        'Anime_R0005',
        'Anime_R0007',
        'Anime_R0009',
        'Anime_R0011',
        'Anime_R0013',
        'Anime_R0015',
        'Anime_R0017',
        'Anime_R0019',
        'Anime_R0021',
        'Anime_R0023',
        'Anime_R0025',
        'Anime_R0027',
        'Anime_R0029',
        'Anime_R0031',
        'Anime_R0033',
        'Anime_R0035',
        'Anime_R0037',
        'Anime_R0039',
        'Anime_R0041',
        'Anime_R0043',
        'Anime_R0045',
        'Anime_R0047',
        'Anime_R0049',
        'Anime_R0051',
    ], [
        'Anime_B0001',
        'Anime_B0003',
        'Anime_B0005',
        'Anime_B0007',
        'Anime_B0009',
        'Anime_B0011',
        'Anime_B0013',
        'Anime_B0015',
        'Anime_B0017',
        'Anime_B0019',
        'Anime_B0021',
        'Anime_B0023',
        'Anime_B0025',
        'Anime_B0027',
        'Anime_B0029',
        'Anime_B0031',
        'Anime_B0033',
        'Anime_B0035',
        'Anime_B0037',
        'Anime_B0039',
        'Anime_B0041',
        'Anime_B0043',
        'Anime_B0045',
        'Anime_B0047',
        'Anime_B0049',
        'Anime_B0051',
    ]];
export const INK_FRAMES = [[
        'Anime_Encrier_R0001',
        'Anime_Encrier_R0003',
        'Anime_Encrier_R0005',
        'Anime_Encrier_R0007',
        'Anime_Encrier_R0009',
        'Anime_Encrier_R0011',
        'Anime_Encrier_R0013',
        'Anime_Encrier_R0015',
        'Anime_Encrier_R0017',
        'Anime_Encrier_R0019',
        'Anime_Encrier_R0021',
        'Anime_Encrier_R0023',
        'Anime_Encrier_R0025',
        'Anime_Encrier_R0027',
        'Anime_Encrier_R0029',
        'Anime_Encrier_R0031',
        'Anime_Encrier_R0033',
        'Anime_Encrier_R0035',
        'Anime_Encrier_R0037',
        'Anime_Encrier_R0039',
        'Anime_Encrier_R0041',
        'Anime_Encrier_R0043',
        'Anime_Encrier_R0045',
        'Anime_Encrier_R0047',
        'Anime_Encrier_R0049',
        'Anime_Encrier_R0051',
        'Anime_Encrier_R0053',
        'Anime_Encrier_R0055',
        'Anime_Encrier_R0057',
    ], [
        'Anime_Encrier_B0001',
        'Anime_Encrier_B0003',
        'Anime_Encrier_B0005',
        'Anime_Encrier_B0007',
        'Anime_Encrier_B0009',
        'Anime_Encrier_B0011',
        'Anime_Encrier_B0013',
        'Anime_Encrier_B0015',
        'Anime_Encrier_B0017',
        'Anime_Encrier_B0019',
        'Anime_Encrier_B0021',
        'Anime_Encrier_B0023',
        'Anime_Encrier_B0025',
        'Anime_Encrier_B0027',
        'Anime_Encrier_B0029',
        'Anime_Encrier_B0031',
        'Anime_Encrier_B0033',
        'Anime_Encrier_B0035',
        'Anime_Encrier_B0037',
        'Anime_Encrier_B0039',
        'Anime_Encrier_B0041',
        'Anime_Encrier_B0043',
        'Anime_Encrier_B0045',
        'Anime_Encrier_B0047',
        'Anime_Encrier_B0049',
        'Anime_Encrier_B0051',
        'Anime_Encrier_B0053',
        'Anime_Encrier_B0055',
        'Anime_Encrier_B0057',
    ]];
