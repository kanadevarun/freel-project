from PIL import Image
import os

src = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-categories.png'
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'
img = Image.open(src).convert('RGBA')

def crop_save(name, box):
    c = img.crop(box)
    path = os.path.join(out, f'dc-{name}.png')
    c.save(path, 'PNG')

# CATEGORY CARDS
# Col 1: 1000-1120 (w:120)
# Col 2: 1195-1315 (w:120)
# Col 3: 1390-1510 (w:120)
# Row 1: 105-235 (h:130)
# Row 2: 300-430 (h:130)
crop_save('card-commercial',   (1000, 105, 1120, 235))
crop_save('card-transport',    (1195, 105, 1315, 235))
crop_save('card-customs',      (1390, 105, 1510, 235))
crop_save('card-financial',    (1000, 300, 1120, 430))
crop_save('card-insurance',    (1195, 300, 1315, 430))
crop_save('card-certificates', (1390, 300, 1510, 430))

print("Crops generated!")
