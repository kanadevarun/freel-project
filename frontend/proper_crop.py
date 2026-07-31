from PIL import Image
import os

img_path = 'public/images/import-export/make-essential-documents.png'
img = Image.open(img_path).convert("RGBA")
width, height = 1536, 1024

out_dir = 'public/images/import-export/documents_new'
os.makedirs(out_dir, exist_ok=True)

# ROW 1 (5 items)
row1_y1 = 0
row1_y2 = height // 3
item_w_r1 = width // 5

p_names = ['manufacturer', 'exporter', 'forwarder', 'customs', 'importer']
for i, name in enumerate(p_names):
    box = (i * item_w_r1, row1_y1, (i+1) * item_w_r1, row1_y2)
    img.crop(box).save(f"{out_dir}/participant_{name}.png")

# ROW 2 (6 items)
row2_y1 = height // 3
row2_y2 = (height // 3) * 2
item_w_r2 = width // 6

d_names = ['invoice', 'packing-list', 'bill-of-lading', 'certificate-of-origin', 'insurance', 'customs-declaration']
for i, name in enumerate(d_names):
    box = (i * item_w_r2, row2_y1, (i+1) * item_w_r2, row2_y2)
    img.crop(box).save(f"{out_dir}/doc_{name}.png")

# ROW 3 (6 items)
row3_y1 = (height // 3) * 2
row3_y2 = height
item_w_r3 = width // 6

t_names = ['ship', 'airplane', 'container', 'port', 'warehouse', 'truck']
for i, name in enumerate(t_names):
    box = (i * item_w_r3, row3_y1, (i+1) * item_w_r3, row3_y2)
    img.crop(box).save(f"{out_dir}/transport_{name}.png")

print("All extracted!")
