from PIL import Image
import os

# ---- configuration ----
INPUT_IMAGE = "Cat_Grey.png"   # path to your sprite sheet

SPRITE_W, SPRITE_H = 32, 32
PAD_W, PAD_H = 34, 34
PADDING_X = (PAD_W - SPRITE_W) // 2  # 1 px
PADDING_Y = (PAD_H - SPRITE_H) // 2  # 1 px
# -----------------------

sheet = Image.open(INPUT_IMAGE).convert("RGBA")
sheet_w, sheet_h = sheet.size

sprites_per_row = sheet_w // SPRITE_W
rows = sheet_h // SPRITE_H

for row in range(rows):
    # normal row image
    row_normal = Image.new(
        "RGBA",
        (sprites_per_row * PAD_W, PAD_H),
        (0, 0, 0, 0)
    )

    # mirrored row image (per-sprite mirror)
    row_mirrored = Image.new(
        "RGBA",
        (sprites_per_row * PAD_W, PAD_H),
        (0, 0, 0, 0)
    )

    for col in range(sprites_per_row):
        x = col * SPRITE_W
        y = row * SPRITE_H

        sprite = sheet.crop((x, y, x + SPRITE_W, y + SPRITE_H))

        # padded sprite
        padded = Image.new("RGBA", (PAD_W, PAD_H), (0, 0, 0, 0))
        padded.paste(sprite, (PADDING_X, PADDING_Y))

        # padded + mirrored sprite
        padded_mirrored = padded.transpose(Image.FLIP_LEFT_RIGHT)

        dest_x = col * PAD_W

        row_normal.paste(padded, (dest_x, 0))
        row_mirrored.paste(padded_mirrored, (dest_x, 0))

    # save outputs
    actions = ['death', 'idle', 'idle2', 'idle3', 'idle4', 'takeoff', 'fly', 'land']
    #row_normal.save(f"{actions[row]}_right.png")
    #row_mirrored.save(f"{actions[row]}_left.png")
    row_normal.save(f"{row}_right.png")
    row_mirrored.save(f"{row}_left.png")

print("Done!")

