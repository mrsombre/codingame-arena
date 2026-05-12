import random
from PIL import Image

def process_with_lock_average(source_path, lock_path, output_path):
    # Load both images
    img = Image.open(source_path).convert("RGBA")
    lock = Image.open(lock_path).convert("RGBA")
    
    width, height = img.size
    pixels = img.load()
    lock_pixels = lock.load()
    
    new_img = Image.new("RGBA", (width, height))
    new_pixels = new_img.load()

    def is_locked(x, y):
        # Locked if transparent in original OR non-transparent in lock file
        return (pixels[x, y][3] == 0) or (lock_pixels[x, y][3] > 0)

    for y in range(height):
        for x in range(width):
            current_pixel = pixels[x, y]

            # If the pixel is locked, keep it exactly as it is
            if is_locked(x, y):
                new_pixels[x, y] = current_pixel
                continue

            # Identify valid neighbors
            neighbors = []
            for dy in [-1, 0, 1]:
                for dx in [-1, 0, 1]:
                    nx, ny = x + dx, y + dy
                    if 0 <= nx < width and 0 <= ny < height:
                        if not is_locked(nx, ny):
                            neighbors.append(pixels[nx, ny])

            if neighbors:
                # Select a random non-empty subset
                # k is the size of the subset (at least 1, up to all neighbors)
                k = random.randint(1, min(3, len(neighbors)))
                subset = random.sample(neighbors, k)
                
                # Calculate the average color of the subset
                sum_r = sum(p[0] for p in subset)
                sum_g = sum(p[1] for p in subset)
                sum_b = sum(p[2] for p in subset)
                sum_a = sum(p[3] for p in subset)
                
                avg_pixel = (
                    int(sum_r / k),
                    int(sum_g / k),
                    int(sum_b / k),
                    int(sum_a / k)
                )
                new_pixels[x, y] = avg_pixel
            else:
                new_pixels[x, y] = current_pixel

    new_img.save(output_path)
    print(f"Averaged jitter image saved to {output_path}")

# Run the process
process_with_lock_average("tileset_.png", "tileset_lock.png", "tileset.png")
