from PIL import Image

img_path = '/Users/varun.kanade/.gemini/antigravity-ide/brain/da92fa42-416b-4b10-a079-93aeb08bd2ca/.tempmediaStorage/media_da92fa42-416b-4b10-a079-93aeb08bd2ca_1782909763288.png'
img = Image.open(img_path)

# Crop the illustration portion
# 2880 x 2320 total size
# Adjust crop coordinates to get just the right side illustration
left = 1550
top = 180
right = 2750
bottom = 1430

cropped = img.crop((left, top, right, bottom))
cropped.save('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/hero-illustration.png')
print("Cropped illustration saved!")
