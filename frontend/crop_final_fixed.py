from PIL import Image
import os

src = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-categories.png'
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'
img = Image.open(src).convert('RGBA')

def crop_save(name, box):
    c = img.crop(box)
    path = os.path.join(out, f'dc-{name}.png')
    c.save(path, 'PNG')

# ECOSYSTEM (X: 250 to 680, Y: 95 to 470)
# Started at 250 to avoid the "ES" from "CATEGORIES"
crop_save('ecosystem-full', (250, 95, 680, 470))

# STATS ICONS (X: 720 to 790)
crop_save('stat-categories-icon', (720, 105, 790, 175))
crop_save('stat-documents-icon',  (720, 205, 790, 275))
crop_save('stat-countries-icon',  (720, 305, 790, 375))
crop_save('stat-essential-icon',  (720, 400, 790, 470))

# CATEGORY CARDS (X: 1010-1165, 1205-1360, 1400-1536)
crop_save('card-commercial',   (1010, 105, 1165, 235))
crop_save('card-transport',    (1205, 105, 1360, 235))
crop_save('card-customs',      (1400, 105, 1536, 235))
crop_save('card-financial',    (1010, 300, 1165, 430))
crop_save('card-insurance',    (1205, 300, 1360, 430))
crop_save('card-certificates', (1400, 300, 1536, 430))

# BADGE ICONS (Y: 535 to 570)
# We only want the icon, not the text next to it.
crop_save('badge-commercial',   (30,  535, 65,  570))
crop_save('badge-transport',    (185, 535, 220, 570))
crop_save('badge-customs',      (305, 535, 340, 570))
crop_save('badge-finance',      (465, 535, 500, 570))
crop_save('badge-insurance',    (615, 535, 650, 570))
crop_save('badge-certificates', (765, 535, 800, 570))

print("Final perfect crops generated!")
