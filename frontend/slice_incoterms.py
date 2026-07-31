from PIL import Image, ImageChops
import os

def trim(im):
    bg = Image.new(im.mode, im.size, im.getpixel((0,0)))
    diff = ImageChops.difference(im, bg)
    diff = ImageChops.add(diff, diff, 2.0, -100)
    bbox = diff.getbbox()
    if bbox:
        return im.crop(bbox)
    return im

img_path = 'public/images/import-export/incoterm-section.png'
img = Image.open(img_path).convert("RGBA")
width, height = 1536, 1024
out_dir = 'public/images/import-export/incoterms_new'
os.makedirs(out_dir, exist_ok=True)

# We will just write a function to slice a row and remove text
def extract_row(y1, y2, num_items, prefix, crop_top_pct=0.0, crop_bot_pct=0.0, start_x=0, end_x=1536):
    row_w = end_x - start_x
    item_w = row_w / num_items
    for i in range(num_items):
        box = (start_x + i * item_w, y1, start_x + (i+1) * item_w, y2)
        slice_img = img.crop(box)
        w, h = slice_img.size
        # Crop text
        slice_img = slice_img.crop((0, int(h * crop_top_pct), w, int(h * (1 - crop_bot_pct))))
        # Trim white
        trimmed = trim(slice_img)
        if trimmed:
            trimmed.save(f"{out_dir}/{prefix}_{i}.png")
            print(f"Saved {prefix}_{i}.png")

# Row 1 Left: Hero & Journey (7 items)
# from x=0 to x=950 roughly
extract_row(50, 220, 7, 'hero', crop_top_pct=0.0, crop_bot_pct=0.15, start_x=0, end_x=950)

# Row 1 Right: Documents (4 items)
# from x=950 to 1536
extract_row(50, 220, 4, 'doc', crop_top_pct=0.0, crop_bot_pct=0.15, start_x=1000, end_x=1536)

# Row 2 Left: Any Mode Cards (7 items)
extract_row(260, 460, 7, 'any', crop_top_pct=0.15, crop_bot_pct=0.20, start_x=20, end_x=950)

# Row 2 Right: Sea Mode Cards (4 items)
extract_row(260, 460, 4, 'sea', crop_top_pct=0.15, crop_bot_pct=0.20, start_x=960, end_x=1500)

# Row 3 Left: Timeline Icons (9 items)
extract_row(480, 600, 9, 'timeline', crop_top_pct=0.0, crop_bot_pct=0.25, start_x=0, end_x=1000)

print("Extraction complete!")
