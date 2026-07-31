from PIL import Image
import os

src = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-categories.png'
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'

img = Image.open(src).convert('RGBA')
W, H = img.size  # 1536 x 1024
print(f"Image size: {W} x {H}")

def crop_save(name, box):
    c = img.crop(box)
    path = os.path.join(out, f'dc-{name}.png')
    c.save(path, 'PNG')
    print(f"Saved dc-{name}.png  {c.size}")

# ─────────────────────────────────────────────────────────────────
# SECTION LABELS in the sprite sheet (reading the image carefully):
#
# The image has a title row "DOCUMENT CATEGORIES - ALL IMAGES & ASSETS"
# Then labeled section headers like "SECTION BADGE", "LEFT ILLUSTRATION (ECOSYSTEM)",
# "RIGHT STAT CARDS", "CATEGORY CARD ILLUSTRATIONS"
# etc.
#
# Based on the 1536 x 1024 image and visual inspection of the reference:
#
# TITLE BAR: y ~ 0-50
# SECTION LABELS row 1: y ~ 50-75
#
# MAIN CONTENT BLOCK 1 (y 75-450):
#   - Section Badge: x 0-240, y 75-140
#   - Ecosystem Diagram: x 240-640, y 75-450
#   - Right Stat Cards: x 640-1010, y 75-450
#   - Category Card Illustrations (2 rows x 3): x 1010-1536, y 75-450
#
# LABELS ROW 2: y 450-480 (Category Badges, Difficulty, Contains)
#
# CONTENT BLOCK 2 (y 480-600):
#   - Category Badges: x 0-660, y 480-600
#   - Difficulty indicators: x 660-940, y 480-600
#   - Contains bullet styles: x 940-1536, y 480-600
#
# LABELS ROW 3: y 600-630 (Bottom Legend, Connector Elements)
#
# CONTENT BLOCK 3 (y 630-710):
#   - Bottom legend chips: x 0-680, y 630-710
#   - Connector elements: x 680-1536, y 630-710
#
# LABELS ROW 4: y 710-740 (Additional Illustrations)
#
# CONTENT BLOCK 4 (y 740-930):
#   - Decorative icons row: 9 items across full width
#     Cargo Ship, Truck, Airplane, Factory, Warehouse, Port/Crane,
#     Customs Office, World Map (dotted), Document Icon
#
# LABELS ROW 5: y 930-960 (UI Elements)
#
# CONTENT BLOCK 5 (y 960-1024):
#   - UI elements row
# ─────────────────────────────────────────────────────────────────

# Save grid slices to verify coordinates
# Save overall sections to figure out coordinates
crop_save('debug-top-block',    (0, 50, 1536, 460))
crop_save('debug-ecosystem',    (230, 65, 645, 455))
crop_save('debug-stat-cards',   (635, 65, 1010, 455))
crop_save('debug-card-illust',  (1005, 65, 1536, 455))
crop_save('debug-badge-row',    (0, 455, 680, 615))
crop_save('debug-deco-row',     (0, 700, 1536, 940))

print("\nDebug crops done")
