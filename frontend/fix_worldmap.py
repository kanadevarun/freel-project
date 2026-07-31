from PIL import Image
import os

img = Image.open('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/hero-documentation.png')
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'

def crop(name, x1, y1, x2, y2):
    region = img.crop((x1, y1, x2, y2))
    region.save(os.path.join(out, name))
    print(f"Saved {name}: {x2-x1}x{y2-y1}")

# The world map dots are in the right portion of the sprite, but the label "ADDITIONAL ASSETS USED IN HERO SECTION"
# text appears at around y=508-530. The actual dots only start lower.
# Let's just skip the world map section from the sprite (it has label overlay)
# Instead crop just the connector/dots pattern from the bottom of the sprite
crop('hero-world-map.png', 1022, 560, 1536, 720)  # dots only, below any label

# Also re-crop the hero images below the label row
crop('hero-cargo-ship.png', 32, 560, 310, 710)
crop('hero-container.png', 318, 560, 508, 710)
crop('hero-port.png', 516, 560, 680, 710)
crop('hero-globe.png', 692, 555, 858, 710)
crop('hero-clipboard.png', 866, 555, 1010, 710)

print("Done!")
