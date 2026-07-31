from PIL import Image
import os

src_new = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/cropped-categories-documents.png'
src_old = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-categories.png'
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'

img_new = Image.open(src_new).convert('RGBA')
img_old = Image.open(src_old).convert('RGBA')

def crop_save(img, name, box):
    c = img.crop(box)
    path = os.path.join(out, f'dc-card-{name}.png')
    c.save(path, 'PNG')

# New Zoomed Images
crop_save(img_new, 'commercial', (70,  140, 440, 510))
crop_save(img_new, 'transport',  (570, 140, 940, 510))
crop_save(img_new, 'customs',    (1070, 140, 1440, 510))
crop_save(img_new, 'insurance',  (320, 570, 690, 940))
crop_save(img_new, 'financial',  (820, 570, 1190, 940))

# Certificates from old image (very tight crop to match zoom)
crop_save(img_old, 'certificates', (1410, 310, 1515, 425))

print("Final Zoomed crops generated!")
