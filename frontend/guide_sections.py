from PIL import Image, ImageDraw
import os

img = Image.open('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/hero-documentation.png')
W, H = img.size  # 1536 x 1024

# Save individual sections as guides
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'

# Save top half (rows 1 and 2)
top_half = img.crop((0, 0, W, 512))
top_half.save(os.path.join(out, 'guide_top.png'))

# Save bottom half (rows 3-5)
bot_half = img.crop((0, 512, W, H))
bot_half.save(os.path.join(out, 'guide_bot.png'))

print(f"Guides saved. Full size: {W}x{H}")
