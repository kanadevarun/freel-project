from PIL import Image
import os

src = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-categories.png'
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'
img = Image.open(src).convert('RGBA')

def crop_save(name, box):
    c = img.crop(box)
    path = os.path.join(out, f'dc-debug-{name}.png')
    c.save(path, 'PNG')

# Block 1 (y: 68-458): contains Eco, Stats, Cards
# Let's crop the full width to see all X coords
crop_save('full-block1', (0, 68, 1536, 458))

# Let's also get the labels above to align X coords
crop_save('labels-row1', (0, 30, 1536, 68))

print("Debug crops saved")
