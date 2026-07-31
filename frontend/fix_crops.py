from PIL import Image

# Use the full sprite sheet which is 1536x1024
img_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/hero-documentation.png'
img = Image.open(img_path)

# 1. Subtle Background (Top Right)
# Let's crop the top right corner of the sprite sheet
bg = img.crop((850, 0, 1500, 350))
bg.save('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/subtle-background.png')

# 2. Air Waybill (Bottom Middle card)
# Air Waybill was not in the original sprite sheet? No wait, the prompt for this chat says:
# "Crop the required illustrations from the existing sprite sheets or PNG assets inside this folder. Examples include: Commercial Invoice ... Air Waybill ... Export License"
# Let's see if Air Waybill is in the sprite sheet. The grid was 3 rows.
awb = img.crop((600, 710, 750, 870)) # rough estimate for middle column
awb.save('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/card-air-waybill.png')

el = img.crop((1100, 710, 1250, 870)) # rough estimate for right column
el.save('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/card-export-license.png')

print("Cropped background and icons from sprite sheet!")
