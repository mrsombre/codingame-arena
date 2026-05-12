#!/usr/bin/env python3

import argparse
from PIL import Image
import colorsys

HUE_RANGE_DEG = 15      # ±15°
HUE_SHIFT_DEG = 120         # subtract 120°
VALUE_THRESHOLD = 20 / 255 # normalize to 0–1

def process_image(input_path):
    img = Image.open(input_path).convert("RGBA")
    pixels = img.load()

    width, height = img.size

    for y in range(height):
        for x in range(width):
            r, g, b, a = pixels[x, y]

            # RGB → HSV (colorsys uses 0–1 range)
            h, s, v = colorsys.rgb_to_hsv(
                r / 255, g / 255, b / 255
            )

            hue_deg = h * 360

            is_red = (
                (hue_deg <= HUE_RANGE_DEG or hue_deg >= 360 - HUE_RANGE_DEG)
                and v > VALUE_THRESHOLD
            )
            is_green = (
                (90 - HUE_RANGE_DEG <= hue_deg <= 90 + HUE_RANGE_DEG)
                and v > VALUE_THRESHOLD
            )
            is_brown = (
                (25 <= hue_deg <= 42)
                and v > VALUE_THRESHOLD
            )

            if is_red and 'blue' in input_path:
                hue_deg = (hue_deg - HUE_SHIFT_DEG) % 360
                h = hue_deg / 360

                r, g, b = colorsys.hsv_to_rgb(h, s, v)
                pixels[x, y] = (
                    int(r * 255),
                    int(g * 255),
                    int(b * 255),
                    a
                )

            if is_green:
                hue_deg = (hue_deg - 50) % 360
                h = hue_deg / 360

                r, g, b = colorsys.hsv_to_rgb(h, s, v)
                pixels[x, y] = (
                    int(r * 255),
                    int(g * 255),
                    int(b * 255),
                    a
                )       

            if is_brown:
                if 'red' in input_path: hue_deg = (hue_deg - 30) % 360
                else: hue_deg = (hue_deg - 30 - HUE_SHIFT_DEG) % 360
                h = hue_deg / 360

                r, g, b = colorsys.hsv_to_rgb(h, s, v)
                pixels[x, y] = (
                    int(r * 255),
                    int(g * 255),
                    int(b * 255),
                    a
                )                

    output_path = "shift_" + input_path
    img.save(output_path)
    print(f"Saved: {output_path}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("image", help="Path to input PNG")
    args = parser.parse_args()

    process_image(args.image)

