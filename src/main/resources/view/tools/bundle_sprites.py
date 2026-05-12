import os
from PIL import Image

def combine_sprite_sheets(input_folder, output_path, sprite_w=34, sprite_h=34):
    # Filter for common image extensions
    valid_extensions = ('.png', '.jpg', '.jpeg', '.bmp', '.webp')
    files = [f for f in sorted(os.listdir(input_folder)) if f.lower().endswith(valid_extensions)]
    
    if not files:
        print("No images found in the directory.")
        return

    processed_rows = []
    max_sprites_in_a_row = 0

    for file in files:
        img = Image.open(os.path.join(input_folder, file)).convert("RGBA")
        img_w, img_h = img.size
        
        # Calculate how many sprites are in this specific file
        cols = img_w // sprite_w
        rows = img_h // sprite_h
        total_sprites = cols * rows
        
        # Keep track of the widest row to determine final canvas width
        max_sprites_in_a_row = max(max_sprites_in_a_row, total_sprites)
        
        # Create a new blank strip for this file
        row_strip = Image.new("RGBA", (total_sprites * sprite_w, sprite_h))
        
        idx = 0
        for r in range(rows):
            for c in range(cols):
                # Crop the individual sprite
                left = c * sprite_w
                top = r * sprite_h
                right = left + sprite_w
                bottom = top + sprite_h
                sprite = img.crop((left, top, right, bottom))
                
                # Paste it into the horizontal strip
                row_strip.paste(sprite, (idx * sprite_w, 0))
                idx += 1
        
        processed_rows.append(row_strip)

    # Create the final master canvas
    final_width = max_sprites_in_a_row * sprite_w
    final_height = len(processed_rows) * sprite_h
    master_sheet = Image.new("RGBA", (final_width, final_height), (0, 0, 0, 0))

    # Stack all strips vertically
    for i, row in enumerate(processed_rows):
        master_sheet.paste(row, (0, i * sprite_h))

    # Save the result
    master_sheet.save(output_path)
    print(f"Success! Master sheet saved at: {output_path}")
    print(f"Dimensions: {final_width}x{final_height}")

# Usage
combine_sprite_sheets(
    input_folder='cat', 
    output_path='cat.png'
)
