from PIL import Image

img_path = '/Users/varun.kanade/.gemini/antigravity-ide/brain/da92fa42-416b-4b10-a079-93aeb08bd2ca/media__1782912865438.png'
img = Image.open(img_path)

# Extract the subtle background illustration (top right of the reference image)
# X: around 800 to 1500
# Y: around 0 to 400
bg = img.crop((850, 0, 1500, 350))
bg.save('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/subtle-background.png')

print("Cropped background!")
