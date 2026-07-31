from PIL import Image
import os

img = Image.open('/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/hero-documentation.png')
W, H = img.size  # 1536 x 1024

out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'
os.makedirs(out, exist_ok=True)

def crop(name, x1, y1, x2, y2):
    region = img.crop((x1, y1, x2, y2))
    path = os.path.join(out, name)
    region.save(path)
    print(f"Saved {name}: {x2-x1}x{y2-y1}")

# Based on 1536x1024 sprite — top half visible at 1536x512

# ─── DOCUMENT CARDS (floating, left side) ─────────────────────────────────────
# Top row: 4 cards at y ≈ 295–400
crop('card-commercial-invoice.png',  48, 295, 192, 405)
crop('card-packing-list.png',       210, 295, 353, 405)
crop('card-bill-of-lading.png',     370, 295, 513, 405)
crop('card-certificate-origin.png', 530, 295, 685, 405)
# Bottom row: 3 cards at y ≈ 430–510
crop('card-letter-credit.png',       48, 430, 192, 516)
crop('card-insurance.png',          210, 430, 353, 516)
crop('card-customs-declaration.png',370, 430, 513, 516)

# ─── BACKGROUND JOURNEY ICONS (right half top row) ───────────────────────────
# 8 icons at y ≈ 300–395, x from 700 to 1536
crop('icon-factory.png',   700, 302, 812, 400)
crop('icon-warehouse.png', 812, 302, 930, 400)
crop('icon-port.png',      930, 302,1043, 400)
crop('icon-cargo-ship.png',1043, 302,1163, 400)
crop('icon-truck.png',    1163, 302,1275, 400)
crop('icon-airplane.png', 1275, 302,1390, 400)
crop('icon-customs.png',  1390, 302,1463, 400)
crop('icon-importer.png', 1463, 302,1536, 400)

# ─── DOCUMENT FLOW STRIP icons (right half bottom) ───────────────────────────
# 7 icons at y ≈ 455–512
crop('flow-commercial-invoice.png',  710, 460, 806, 512)
crop('flow-packing-list.png',        815, 460, 907, 512)
crop('flow-certificate-origin.png',  918, 460,1010, 512)
crop('flow-bill-lading.png',        1020, 460,1117, 512)
crop('flow-insurance.png',          1128, 460,1222, 512)
crop('flow-customs-declaration.png',1232, 460,1325, 512)
crop('flow-success.png',            1336, 460,1430, 512)

# ─── LARGE HERO ASSETS (bottom half, row 3 at y≈512–680) ─────────────────────
bot = img.crop((0, 512, W, H))
# Save bottom half guide
bot.save(os.path.join(out, 'guide_bot.png'))
# Coordinates within full image (y + 512)
crop('hero-cargo-ship.png',   30, 540, 318, 700)
crop('hero-container.png',   325, 545, 510, 698)
crop('hero-port.png',        517, 545, 686, 698)
crop('hero-globe.png',       694, 540, 860, 698)
crop('hero-clipboard.png',   868, 540,1012, 698)
crop('hero-world-map.png',  1018, 510,1536, 700)

# ─── FLAT DOC IMAGES (row 4, y≈700–820) ──────────────────────────────────────
crop('doc-invoice.png',       30, 704, 198, 820)
crop('doc-packing-list.png', 206, 704, 390, 820)
crop('doc-bill-lading.png',  398, 704, 545, 820)
crop('doc-cert-origin.png',  553, 704, 700, 820)
crop('doc-insurance.png',    708, 704, 858, 820)
crop('doc-customs.png',      866, 704,1008, 820)
crop('shield-blue.png',     1016, 704,1118, 820)
crop('shield-green.png',    1126, 704,1225, 820)

# ─── STATUS PILLS (row 5, y≈868–915) ─────────────────────────────────────────
crop('status-verified.png',   30, 868, 143, 915)
crop('status-generated.png', 152, 868, 278, 915)
crop('status-issued.png',    286, 868, 384, 915)
crop('status-certified.png', 392, 868, 514, 915)
crop('status-approved.png',  522, 868, 644, 915)
crop('status-filed.png',     652, 868, 754, 915)

print("All crops done!")
