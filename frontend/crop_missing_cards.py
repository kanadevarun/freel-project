from PIL import Image

img_path = '/Users/varun.kanade/.gemini/antigravity-ide/brain/da92fa42-416b-4b10-a079-93aeb08bd2ca/media__1782912865438.png'
img = Image.open(img_path)

# Card dimensions in the reference image (approximate)
# The cards are distributed in 3 rows and 3 columns.
# We just need the illustration part (left side of the card).
# Let's crop a square around the illustration.

# Air Waybill (Row 3, Col 2)
# X: around 530 to 650
# Y: around 720 to 860
awb = img.crop((520, 710, 680, 870))
awb.save('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/card-air-waybill.png')

# Export License (Row 3, Col 3)
# X: around 1030 to 1150
# Y: around 720 to 860
el = img.crop((1020, 710, 1180, 870))
el.save('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/card-export-license.png')

print("Cropped missing icons!")
