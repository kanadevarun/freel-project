from PIL import Image
import os

img_path = 'public/images/import-export/make-essential-documents.png'
img = Image.open(img_path).convert("RGBA")
width, height = img.size
print(f"Dimensions: {width} x {height}")

# Slicing the hero
# Let's say top portion is the hero... we will need to coordinate closely.
