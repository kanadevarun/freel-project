from PIL import Image
import os

src = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-journey.png'
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'
img = Image.open(src).convert('RGBA')

def crop_save(name, box):
    c = img.crop(box)
    path = os.path.join(out, f'dj-{name}.png')
    c.save(path, 'PNG')

# 1. Timeline Stages
width_per_stage = 1536 // 12
for i in range(12):
    left = i * width_per_stage
    right = (i + 1) * width_per_stage
    crop_save(f'stage-{i+1:02d}', (left, 150, right, 320))

# 2. Decorative Header Illustration
crop_save('decor-factory',    (0,   600, 180, 770))
crop_save('decor-truck',      (180, 600, 320, 770))
crop_save('decor-crane',      (320, 600, 450, 770))
crop_save('decor-ship',       (450, 600, 590, 770))
crop_save('decor-warehouse',  (590, 600, 710, 770))
crop_save('decor-worldmap',   (710, 600, 900, 770))
crop_save('decor-cloud',      (900, 600, 1020, 770))

# 3. Learning Card Icons
crop_save('icon-why',     (630, 420, 760, 570))
crop_save('icon-mistake', (770, 420, 900, 570))
crop_save('icon-tip',     (910, 420, 1040, 570))

# 4. Connecting line and circles (Optional, but good to have)
crop_save('ui-arrow',       (10,  850, 90,  980))
crop_save('ui-dot-circle',  (100, 850, 170, 980))
crop_save('ui-step-circle', (180, 850, 260, 980))
crop_save('ui-dot-line',    (270, 850, 350, 980))

# Also crop the book icon if it happens to be hiding around 1100, 850
crop_save('ui-book',        (1080, 850, 1200, 980))

print("All Document Journey crops generated successfully!")
