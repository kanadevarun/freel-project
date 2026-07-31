from PIL import Image
import os

img = Image.open('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/hero-documentation.png')
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'

def crop(name, x1, y1, x2, y2):
    region = img.crop((x1, y1, x2, y2))
    region.save(os.path.join(out, name))
    print(f"Saved {name}: {x2-x1}x{y2-y1}")

# Just fix the hero cargo ship - the image starts below "ADDITIONAL ASSETS" label
# Label is ~y=508-535, actual cargo ship starts y~535
crop('hero-cargo-ship.png',   30, 545, 318, 708)
crop('hero-container.png',   325, 548, 508, 700)
crop('hero-port.png',        517, 548, 684, 700)
crop('hero-globe.png',       694, 543, 860, 700)
crop('hero-clipboard.png',   868, 543,1012, 700)
# World map bg - just the dots portion
crop('hero-world-map.png',  1022, 520,1536, 710)

print("Fixed assets done!")
