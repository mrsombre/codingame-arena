from PIL import Image
import os, math

# CONFIG
INPUT_DIR = "Front_Drop"          # folder with png files
CELL_WIDTH = 840
CELL_HEIGHT = 720
COLUMNS = 4                     # sprites per row
BACKGROUND_COLOR = (0, 0, 0, 0) # transparent

# Load images
images = [
    Image.open(os.path.join(INPUT_DIR, f)).convert("RGBA")
    for f in sorted(os.listdir(INPUT_DIR))
    if f.lower().endswith(".png")
]

if not images:
    raise ValueError("No PNG files found")

# Calculate grid size
rows = math.ceil(len(images) / COLUMNS)
sheet_width = COLUMNS * CELL_WIDTH
sheet_height = rows * CELL_HEIGHT

# Create sprite sheet
sheet = Image.new("RGBA", (sheet_width, sheet_height), BACKGROUND_COLOR)

for index, img in enumerate(images):
    col = index % COLUMNS
    row = index // COLUMNS

    cell_x = col * CELL_WIDTH
    cell_y = row * CELL_HEIGHT

    # Center image in cell
    offset_x = cell_x + (CELL_WIDTH - img.width) // 2
    offset_y = cell_y + (CELL_HEIGHT - img.height) // 2

    sheet.paste(img, (offset_x, offset_y), img)

# Save result
sheet = sheet.resize(
    (sheet.width // 4, sheet.height // 4),
    Image.NEAREST
)

sheet.save(INPUT_DIR + '.png')

