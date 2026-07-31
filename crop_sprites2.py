import os
from PIL import Image

def dilate(mask, w, h, radius=3):
    new_mask = bytearray(w * h)
    for y in range(h):
        for x in range(w):
            if mask[y * w + x]:
                for dy in range(-radius, radius+1):
                    for dx in range(-radius, radius+1):
                        ny, nx = y + dy, x + dx
                        if 0 <= ny < h and 0 <= nx < w:
                            new_mask[ny * w + nx] = 1
    return new_mask

def crop_sprites(img_path, out_dir):
    os.makedirs(out_dir, exist_ok=True)
    img = Image.open(img_path).convert("RGB")
    w, h = img.size
    
    # Create binary mask (1 if NOT background)
    pixels = img.load()
    mask = bytearray(w * h)
    for y in range(h):
        for x in range(w):
            r, g, b = pixels[x, y]
            # Ignore mostly white pixels and the blue separator lines (approx 0, 0, 200)
            if not (r > 240 and g > 240 and b > 240) and not (r < 100 and g < 100 and b > 150):
                mask[y * w + x] = 1

    print("Dilating mask (this may take a few seconds)...")
    dilated_mask = dilate(mask, w, h, radius=8)
    
    visited = bytearray(w * h)
    blobs = []
    
    for y in range(h):
        for x in range(w):
            idx = y * w + x
            if dilated_mask[idx] and not visited[idx]:
                q = [(x, y)]
                visited[idx] = 1
                min_x, max_x = x, x
                min_y, max_y = y, y
                
                while q:
                    cx, cy = q.pop(0)
                    for dy, dx in [(-1,0), (1,0), (0,-1), (0,1), (-1,-1), (-1,1), (1,-1), (1,1)]:
                        nx, ny = cx + dx, cy + dy
                        if 0 <= nx < w and 0 <= ny < h:
                            nidx = ny * w + nx
                            if dilated_mask[nidx] and not visited[nidx]:
                                visited[nidx] = 1
                                q.append((nx, ny))
                                if nx < min_x: min_x = nx
                                if nx > max_x: max_x = nx
                                if ny < min_y: min_y = ny
                                if ny > max_y: max_y = ny
                
                bw, bh = max_x - min_x, max_y - min_y
                if bw > 30 and bh > 30:
                    if bw > w * 0.8 or bh > h * 0.8:
                        continue
                    if bw < 50 and bh < 50:
                        continue
                    blobs.append((min_x, min_y, max_x, max_y))
    
    print(f"Found {len(blobs)} blobs. Saving...")
    blobs.sort(key=lambda b: (b[1]//150, b[0]))
    
    for i, (min_x, min_y, max_x, max_y) in enumerate(blobs):
        real_min_x, real_max_x = max_x, min_x
        real_min_y, real_max_y = max_y, min_y
        has_pixels = False
        for cy in range(min_y, max_y+1):
            for cx in range(min_x, max_x+1):
                if mask[cy * w + cx]:
                    if cx < real_min_x: real_min_x = cx
                    if cx > real_max_x: real_max_x = cx
                    if cy < real_min_y: real_min_y = cy
                    if cy > real_max_y: real_max_y = cy
                    has_pixels = True
        
        if has_pixels:
            real_min_x = max(0, real_min_x - 5)
            real_min_y = max(0, real_min_y - 5)
            real_max_x = min(w-1, real_max_x + 5)
            real_max_y = min(h-1, real_max_y + 5)
            
            crop = img.crop((real_min_x, real_min_y, real_max_x, real_max_y))
            # Convert background to transparent if needed, or just leave it white.
            # The reference images have white background! The design specifies white background.
            crop.save(os.path.join(out_dir, f'cc_sprite_{i:02d}.png'))
            print(f"Saved cc_sprite_{i:02d}.png (Size: {real_max_x-real_min_x}x{real_max_y-real_min_y})")

img_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/import-export/customs-clearance-docs.png'
out_dir = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/import-export/customs/'
crop_sprites(img_path, out_dir)
