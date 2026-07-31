from PIL import Image

img_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/hero-section-documentation-picture.png'
img = Image.open(img_path)

# Let's see the size of this image
print("Original size:", img.size)

# The top navigation + breadcrumb is typically around 120-140 pixels.
# Let's crop from Y=120 downwards to keep only the hero section.
# We will save this back to the same path or a temporary path to verify.
w, h = img.size
cropped = img.crop((0, 130, w, h))
cropped.save('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/hero-section-documentation-picture-cropped.png')
print("Cropped successfully to:", cropped.size)
