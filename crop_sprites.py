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
    img = Image.open(img_path).convert("RGBA")
    w, h = img.size
    
    # Create binary mask from alpha
    pixels = img.load()
    mask = bytearray(w * h)
    for y in range(h):
        for x in range(w):
            r, g, b, a = pixels[x, y]
            if a > 30:
                mask[y * w + x] = 1

    # Dilate to connect nearby parts of the same illustration
    print("Dilating mask (this may take a few seconds)...")
    dilated_mask = dilate(mask, w, h, radius=6)
    
    # BFS to find blobs
    visited = bytearray(w * h)
    blobs = []
    
    for y in range(h):
        for x in range(w):
            idx = y * w + x
            if dilated_mask[idx] and not visited[idx]:
                # Start BFS
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
                if bw > 40 and bh > 40:
                    # Ignore the long separator lines
                    if bw > w * 0.8 or bh > h * 0.8:
                        continue
                    # Ignore small circles (step numbers)
                    if bw < 50 and bh < 50:
                        continue
                    blobs.append((min_x, min_y, max_x, max_y))
    
    # Crop and save
    print(f"Found {len(blobs)} blobs. Saving...")
    blobs.sort(key=lambda b: (b[1]//100, b[0])) # Sort top-to-bottom, left-to-right
    
    for i, (min_x, min_y, max_x, max_y) in enumerate(blobs):
        # The bounding box was computed on the dilated mask. We can crop tightly by checking the real mask inside this box.
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
            # Add small 2px padding
            real_min_x = max(0, real_min_x - 2)
            real_min_y = max(0, real_min_y - 2)
            real_max_x = min(w-1, real_max_x + 2)
            real_max_y = min(h-1, real_max_y + 2)
            
            crop = img.crop((real_min_x, real_min_y, real_max_x, real_max_y))
            crop.save(os.path.join(out_dir, f'cc_sprite_{i:02d}.png'))
            print(f"Saved cc_sprite_{i:02d}.png (Size: {real_max_x-real_min_x}x{real_max_y-real_min_y})")

img_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/import-export/customs-clearance-docs.png'
out_dir = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/import-export/customs/'
crop_sprites(img_path, out_dir)
