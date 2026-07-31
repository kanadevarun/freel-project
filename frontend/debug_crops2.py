from PIL import Image
import os

src = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-categories.png'
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'
img = Image.open(src).convert('RGBA')

def crop_save(name, box):
    c = img.crop(box)
    path = os.path.join(out, f'dc-{name}.png')
    c.save(path, 'PNG')

# Use the debug-top-block crop (y:50-460) as reference
# That showed that the content starts below the section header labels
# The "LEFT ILLUSTRATION (ECOSYSTEM)" label takes up roughly y 50-80
# The actual ecosystem diagram starts y ~80

# Let's crop the ecosystem more precisely
# Looking at debug-ecosystem image (x:230-645, y:65-455):
# The title "LEFT ILLUSTRATION" is at top
# Commercial node is at roughly relative y=60-170 (absolute y=125-235)
# Transport node: relative x=210-415 y=60-170 (absolute x=440-645, y=125-235) 
# Financial node: relative x=0-150 y=200-330 (absolute x=230-380, y=265-395)
# Customs node: relative x=280-415 y=200-330 (absolute x=510-645, y=265-395)
# Insurance node: relative x=30-170 y=345-455 (absolute x=260-400, y=410-520)
# Wait - y goes to 520? Let me recalc.

# The top block spans y:50-460. The section labels "LEFT ILLUSTRATION" starts at y=50.
# The actual diagram: starts below label. Let's say labels take ~30px each.
# So diagram starts at y ~80.

# From debug-ecosystem (415x390, mapping x:230-645, y:65-455):
# Commercial node (document + calculator): at relative (45,60)-(215,180) 
#   => absolute (275, 125)-(445, 245)
# Transport node (cargo ship): relative (215,60)-(415,180)
#   => absolute (445, 125)-(645, 245)
# Center (blue doc icon): relative (155,155)-(305,345)
#   => absolute (385, 220)-(535, 410)
# Financial (bank building): relative (0,190)-(145,330)
#   => absolute (230, 255)-(375, 395)
# Customs (bank+shield): relative (275,175)-(415,320)
#   => absolute (505, 240)-(645, 385)
# Insurance (blue shield): relative (25,330)-(155,435)
#   => absolute (255, 395)-(385, 500) - but max was 455, so y2=455
# Certificates (paper+medal): relative (195,325)-(375,435)
#   => absolute (425, 390)-(605, 520) - but crop was only to 455

# The key issue: the ecosystem diagram extends BELOW y=455 for bottom nodes!
# Let me re-examine what the actual y range for the ecosystem is.

# Let me crop a wider vertical range for the whole ecosystem
crop_save('debug-eco-wider', (228, 60, 650, 500))

# Also let me check the header text position to know where content really starts
crop_save('debug-top-50', (0, 0, 1536, 100))  # Title area
crop_save('debug-stat-wider', (635, 60, 1015, 465))  # Stats with more room
crop_save('debug-card-wider', (1005, 60, 1536, 465))  # Cards

print("Done")
