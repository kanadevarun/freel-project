from PIL import Image
import os

img_path = 'public/images/import-export/make-essential-documents.png'
img = Image.open(img_path).convert("RGBA")
width, height = img.size

row_height = height // 3
row3_y = row_height * 2

item_width = width // 6
out_dir = 'public/images/import-export/transport'
os.makedirs(out_dir, exist_ok=True)

names = ['ship', 'airplane', 'container', 'port', 'warehouse', 'truck']

for i, name in enumerate(names):
    left = i * item_width
    right = (i + 1) * item_width
    top = row3_y
    bottom = height
    
    box_height = bottom - top
    crop_bottom = bottom - int(box_height * 0.15) # remove text label
    
    box = (left, top, right, crop_bottom)
    cropped = img.crop(box)
    
    # Try to trim whitespace by getting bounding box of non-transparent pixels
    # Since background is white, let's first make white transparent
    
    # Get bounding box of all pixels that are not pure white or transparent
    # Actually, simplest is to just save it. The frontend can use object-fit: contain
    cropped.save(f"{out_dir}/{name}.png")
    print(f"Saved {name}.png")

