from PIL import Image
import os

img_path = 'public/images/import-export/make-essential-documents.png'
img = Image.open(img_path).convert("RGBA")
width, height = img.size

out_dir = 'public/images/import-export/documents'
os.makedirs(out_dir, exist_ok=True)

# Let's assume it's a 3x3 grid
cols = 3
rows = 2
item_width = width // cols
item_height = height // rows

count = 1
for r in range(rows):
    for c in range(cols):
        left = c * item_width
        right = (c + 1) * item_width
        top = r * item_height
        bottom = (r + 1) * item_height
        
        box = (left, top, right, bottom)
        cropped = img.crop(box)
        cropped.save(f"{out_dir}/doc_{count}.png")
        count += 1

print("Sliced!")
