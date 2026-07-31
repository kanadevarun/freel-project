import sys

target = 'frontend/src/pages/public/trade-intelligence/guides/import-export-basics/Page.jsx'
with open(target, 'r') as f:
    content = f.read()

with open('/Users/varun.kanade/.gemini/antigravity-ide/brain/0438df59-754c-494a-bf4e-8bd5ba8af0e1/scratch/inco_jsx.js', 'r') as f:
    jsx = f.read()

marker = 'export default function ImportExportBasicsPage'
if marker in content:
    new_content = content.replace(marker, jsx + '\n\n' + marker)
    with open(target, 'w') as f:
        f.write(new_content)
    print("Injected successfully!")
else:
    print("Marker not found!")
