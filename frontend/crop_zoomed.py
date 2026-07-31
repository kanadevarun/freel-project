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
crop_save(img_new, 'commercial', (60,  140, 350, 490))
crop_save(img_new, 'transport',  (360, 140, 650, 490))
crop_save(img_new, 'customs',    (680, 140, 950, 490))
crop_save(img_new, 'insurance',  (190, 560, 470, 910))
crop_save(img_new, 'financial',  (540, 560, 810, 910))

# For Certificates, we must use the old image, but let's crop it tightly to match the proportions
# Old Certificates was X: 1400-1536, Y: 300-430
# Let's crop it very tightly around the document itself.
# Looking at dc-debug-grid from earlier, it's roughly (1410, 310, 1510, 420)
crop_save(img_old, 'certificates', (1410, 310, 1515, 425))

print("Zoomed crops generated!")
