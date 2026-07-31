from PIL import Image
import os

src = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation/document-categories.png'
img = Image.open(src).convert('RGBA')

# Let's save x: 600-800 to see what's between ecosystem and stats
out = '/Users/varun.kanade/go/src/freel/freel-project/frontend/public/images/documentation'
c = img.crop((600, 68, 800, 458))
c.save(os.path.join(out, 'dc-debug-600-800.png'), 'PNG')

# Let's save x: 800-1000 to see what's between stats and cards
c2 = img.crop((800, 68, 1000, 458))
c2.save(os.path.join(out, 'dc-debug-800-1000.png'), 'PNG')

# Let's save x: 950-1150 to see the start of the cards
c3 = img.crop((950, 68, 1150, 458))
c3.save(os.path.join(out, 'dc-debug-950-1150.png'), 'PNG')
