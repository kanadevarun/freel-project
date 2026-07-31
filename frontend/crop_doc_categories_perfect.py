from PIL import Image
import os

src = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-categories.png'
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'
img = Image.open(src).convert('RGBA')

def crop_save(name, box):
    c = img.crop(box)
    path = os.path.join(out, f'dc-{name}.png')
    c.save(path, 'PNG')

# ECOSYSTEM
crop_save('ecosystem-full', (170, 95, 630, 460))

# STAT ICONS
crop_save('stat-categories-icon', (645, 95, 715, 165))
crop_save('stat-documents-icon',  (645, 195, 715, 265))
crop_save('stat-countries-icon',  (645, 292, 715, 362))
crop_save('stat-essential-icon',  (645, 390, 715, 460))

# CATEGORY CARDS
# Col 1: 1015-1155, Col 2: 1210-1350, Col 3: 1405-1536
# Row 1: 105-235, Row 2: 300-430
crop_save('card-commercial',   (1015, 105, 1155, 235))
crop_save('card-transport',    (1210, 105, 1350, 235))
crop_save('card-customs',      (1405, 105, 1536, 235))
crop_save('card-financial',    (1015, 300, 1155, 430))
crop_save('card-insurance',    (1210, 300, 1350, 430))
crop_save('card-certificates', (1405, 300, 1536, 430))

print("Crops generated!")
