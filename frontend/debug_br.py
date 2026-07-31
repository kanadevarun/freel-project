from PIL import Image, ImageDraw, ImageFont
import os

src = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-journey.png'
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/dj-debug-bottom-right.png'
img = Image.open(src).convert('RGBA').crop((1000, 800, 1536, 1024))

draw = ImageDraw.Draw(img)
width, height = img.size

# Draw grid lines every 50px
for x in range(0, width, 50):
    draw.line([(x, 0), (x, height)], fill=(255, 0, 0, 128), width=1)
for y in range(0, height, 50):
    draw.line([(0, y), (width, y)], fill=(255, 0, 0, 128), width=1)

img.save(out, 'PNG')
print("Saved debug crop")
