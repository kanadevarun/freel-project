from PIL import Image

img_path = '/Users/varun.kanade/.gemini/antigravity-ide/brain/da92fa42-416b-4b10-a079-93aeb08bd2ca/media__1782912865438.png'
img = Image.open(img_path)

# Subtle Background (Top Right)
bg = img.crop((520, 0, 1024, 300))
bg.save('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/subtle-background.png')

# Air Waybill (Row 3, Col 2)
awb = img.crop((390, 580, 490, 680))
awb.save('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/card-air-waybill.png')

# Export License (Row 3, Col 3)
el = img.crop((695, 580, 795, 680))
el.save('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/card-export-license.png')

print("Cropped from original reference image!")
