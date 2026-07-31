import cv2
import numpy as np
import os

img_path = '/Users/varun.kanade/.gemini/antigravity-ide/brain/0438df59-754c-494a-bf4e-8bd5ba8af0e1/customs-clearance-docs.png'
out_dir = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/import-export/customs/'
os.makedirs(out_dir, exist_ok=True)

img = cv2.imread(img_path, cv2.IMREAD_UNCHANGED)
if img is None:
    print("Could not load image")
    exit(1)

# Extract alpha channel
alpha = img[:, :, 3]

# Threshold to get mask
_, thresh = cv2.threshold(alpha, 10, 255, cv2.THRESH_BINARY)

# Find contours
contours, _ = cv2.findContours(thresh, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)

count = 0
for i, cnt in enumerate(contours):
    x, y, w, h = cv2.boundingRect(cnt)
    # Filter out small noise or lines
    if w > 30 and h > 30:
        crop = img[y:y+h, x:x+w]
        # Ignore mostly horizontal/vertical lines (like the dividers in the sheet)
        if w/h > 15 or h/w > 15: continue
        cv2.imwrite(os.path.join(out_dir, f'sprite_{count}.png'), crop)
        count += 1

print(f"Extracted {count} sprites to {out_dir}")
