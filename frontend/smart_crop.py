from PIL import Image, ImageChops

def trim(im):
    bg = Image.new(im.mode, im.size, im.getpixel((0,0)))
    diff = ImageChops.difference(im, bg)
    diff = ImageChops.add(diff, diff, 2.0, -100)
    bbox = diff.getbbox()
    if bbox:
        return im.crop(bbox)
    return im

img_path = 'public/images/import-export/make-essential-documents.png'
img = Image.open(img_path).convert("RGBA")

# Make background white instead of transparent for trimming if necessary, but RGBA usually has transparent bg.
# Actually, the background is white.
width, height = 1536, 1024
out_dir = 'public/images/import-export/documents_new'

# Row 1: Participants (5 items)
p_names = ['manufacturer', 'exporter', 'forwarder', 'customs', 'importer']
for i, name in enumerate(p_names):
    box = (i * (width//5), 0, (i+1) * (width//5), height//3)
    slice_img = img.crop(box)
    # Crop off bottom 25% to remove text
    w, h = slice_img.size
    slice_img = slice_img.crop((0, 0, w, int(h * 0.75)))
    trim(slice_img).save(f"{out_dir}/participant_{name}.png")

# Row 2: Documents (6 items)
d_names = ['invoice', 'packing-list', 'bill-of-lading', 'certificate-of-origin', 'insurance', 'customs-declaration']
for i, name in enumerate(d_names):
    box = (i * (width//6), height//3, (i+1) * (width//6), (height//3)*2)
    slice_img = img.crop(box)
    # Crop off bottom 30% to remove text
    w, h = slice_img.size
    slice_img = slice_img.crop((0, 0, w, int(h * 0.7)))
    # Crop off top 10% to remove any bleed from row 1
    slice_img = slice_img.crop((0, int(h * 0.1), w, slice_img.size[1]))
    trim(slice_img).save(f"{out_dir}/doc_{name}.png")

# Row 3: Transport (6 items)
t_names = ['ship', 'airplane', 'container', 'port', 'warehouse', 'truck']
for i, name in enumerate(t_names):
    box = (i * (width//6), (height//3)*2, (i+1) * (width//6), height)
    slice_img = img.crop(box)
    # Crop off bottom 25% to remove text
    w, h = slice_img.size
    slice_img = slice_img.crop((0, 0, w, int(h * 0.75)))
    trim(slice_img).save(f"{out_dir}/transport_{name}.png")

print("Smart cropped successfully!")
