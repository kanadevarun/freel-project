from PIL import Image, ImageDraw, ImageFont
import os

src = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-journey.png'
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'
img = Image.open(src).convert('RGBA')

draw = ImageDraw.Draw(img)
width, height = img.size

print(f"Image size: {img.size}")

# Draw grid lines every 100px
for x in range(0, width, 100):
    draw.line([(x, 0), (x, height)], fill=(255, 0, 0, 128), width=1)
for y in range(0, height, 100):
    draw.line([(0, y), (width, y)], fill=(255, 0, 0, 128), width=1)

try:
    font = ImageFont.truetype("arial.ttf", 20)
except:
    font = ImageFont.load_default()

for x in range(0, width, 200):
    for y in range(0, height, 200):
        draw.text((x+2, y+2), f"{x},{y}", fill=(255, 0, 0, 255), font=font)

path = os.path.join(out, 'dj-debug-grid.png')
img.save(path, 'PNG')
print(f"Grid image saved to {path}")
