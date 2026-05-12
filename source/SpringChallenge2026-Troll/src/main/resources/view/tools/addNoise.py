#!/usr/bin/env python3

import argparse
from PIL import Image
import colorsys
import random

RND_RANGE = 10

def process_image(input_path):
    img = Image.open(input_path).convert("RGBA")
    pixels = img.load()
    width, height = img.size

    for y in range(height):
        for x in range(width):
            r, g, b, a = pixels[x, y]
            r += random.randint(-RND_RANGE, RND_RANGE)
            g += random.randint(-RND_RANGE, RND_RANGE)
            b += random.randint(-RND_RANGE, RND_RANGE)
            pixels[x, y] = (r,g,b,a)

    output_path = 'xx' + input_path
    img.save(output_path)
    print(f"Saved: {output_path}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("image", help="Path to input PNG")
    args = parser.parse_args()

    process_image(args.image)

