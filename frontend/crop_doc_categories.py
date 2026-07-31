from PIL import Image
import os

src = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-categories.png'
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'
img = Image.open(src).convert('RGBA')

def crop_save(name, box):
    c = img.crop(box)
    path = os.path.join(out, f'dc-{name}.png')
    c.save(path, 'PNG')
    print(f"  dc-{name}.png  {c.size}")

# The sprite image title "DOCUMENT CATEGORIES - ALL IMAGES & ASSETS"
# occupies y: 0-35
# Section header labels: y: 35-68
# 
# Content block 1: y: 68-458
# The "CATEGORY CARD ILLUSTRATIONS" label starts at x:1005, y:35
# which means the actual illustrations start at y:68

# Looking at dc-debug-card-wider which was (x:1005-1536, y:60-465):
# Relative y = 60-68 = label text at top (8px)
# Then 3D art fills the space below
# Top row: 3 illustration cells, each approximately 195px tall 
#   Each cell: illustration from top with label text at bottom (~25px)
#   Art area: y 68 to ~237 (169px)
# Bottom row: y ~243 to ~432

# From the debug image we saw label "Commercial Documents" at bottom of each cell
# So crop tight, trimming ~20px from bottom of each cell:
# Top row cells: y 75-230 (avoiding "CATEGORY CARD ILLUSTRATIONS" header at y:35-68)
# Bottom row cells: y 247-425

print("=== CARD ILLUSTRATIONS (trimmed) ===")
# Top row: y 75-230 (no header text, no bottom label)
crop_save('card-commercial',   (1010, 75,  1183, 230))
crop_save('card-transport',    (1183, 75,  1356, 230))
crop_save('card-customs',      (1356, 75,  1536, 230))
# Bottom row: y 248-428
crop_save('card-financial',    (1010, 248, 1183, 425))
crop_save('card-insurance',    (1183, 248, 1356, 425))
crop_save('card-certificates', (1356, 248, 1536, 425))

print("\n=== DECORATIVE ICONS (trimmed) ===")
# Deco row: label "ADDITIONAL ILLUSTRATIONS" at ~y 722-755
# Content: y 755-930
# Label text at bottom (Cargo Ship, Truck, etc.) starts at ~y 880
# So art is y 755-878 (no bottom label)
items = [
    ('deco-ship',          0,    170),
    ('deco-truck',         172,  340),
    ('deco-airplane',      342,  510),
    ('deco-factory',       512,  682),
    ('deco-warehouse',     684,  854),
    ('deco-port',          856,  1024),
    ('deco-customs-office',1026, 1194),
    ('deco-worldmap',      1196, 1364),
    ('deco-document',      1366, 1536),
]
for name, x1, x2 in items:
    crop_save(name, (x1, 758, x2, 880))

print("\n=== STAT ICONS ===")
# From debug-stat-wider (x:635-1015, y:60-465):
# The icons are colored rounded squares of ~77x77px
# Looking at the reference image more carefully:
# "RIGHT STAT CARDS" label is at y:35-68
# Then 4 stats stacked in y:68-458, each ~97px
# Each stat row: icon at far left, then number, then label
# Icon box roughly: 635-717, heights:
# Row1 (layers): y 95-168 (with some padding)
# Row2 (doc): y 195-265
# Row3 (globe): y 290-360
# Row4 (shield): y 385-455

# Let's just crop the icon squares more precisely
# The actual square icons are about 64x64px with some padding
crop_save('stat-categories-icon', (640, 97,  716, 170))
crop_save('stat-documents-icon',  (640, 198, 716, 268))
crop_save('stat-countries-icon',  (640, 293, 716, 362))
crop_save('stat-essential-icon',  (640, 387, 716, 455))

print("\n=== ECO ICONS (tighter crops) ===")
# These node icons from the ecosystem - crop only the 3D art, not labels
crop_save('eco-commercial-icon',   (272, 70,  430, 155))
crop_save('eco-transport-icon',    (442, 70,  638, 158))
crop_save('eco-center-icon',       (368, 173, 505, 283))
crop_save('eco-financial-icon',    (228, 200, 350, 314))
crop_save('eco-customs-icon',      (506, 198, 642, 312))
crop_save('eco-insurance-icon',    (251, 322, 378, 432))
crop_save('eco-certificates-icon', (436, 323, 588, 432))

print("\nDone!")
