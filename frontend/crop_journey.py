from PIL import Image
import os

src = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-categories.png'
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'
img = Image.open(src).convert('RGBA')

def crop_save(name, box):
    c = img.crop(box)
    path = os.path.join(out, f'dc-deco-{name}.png')
    c.save(path, 'PNG')

# Y: 730 to 850
crop_save('port',      (620, 730, 720, 850))
crop_save('ship',      (20,  730, 130, 850))
crop_save('truck',     (140, 730, 240, 850))
crop_save('warehouse', (490, 730, 600, 850))
crop_save('document',  (990, 730, 1090, 850))

print("Journey icons cropped!")
