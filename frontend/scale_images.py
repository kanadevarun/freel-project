import os

css_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/src/pages/public/trade-intelligence/guides/import-export-basics/Page.css'
jsx_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/src/pages/public/trade-intelligence/guides/import-export-basics/Page.jsx'

# 1. UPDATE CSS
with open(css_path, 'r') as f:
    css = f.read()

css = css.replace('.pay-hd-item {\n  position: absolute;\n  height: 80px;', '.pay-hd-item {\n  position: absolute;\n  height: 140px;')
css = css.replace('.pay-method-item img {\n  height: 48px;', '.pay-method-item img {\n  height: 90px;')
css = css.replace('.pay-lc-step img {\n  height: 40px;', '.pay-lc-step img {\n  height: 70px;')
css = css.replace('.pay-eco-item img {\n  height: 48px;', '.pay-eco-item img {\n  height: 90px;')
css = css.replace('.pay-rw-step img {\n  height: 40px;', '.pay-rw-step img {\n  height: 70px;')
css = css.replace('.pay-pt-icon {\n  width: 16px;\n  height: 16px;', '.pay-pt-icon {\n  width: 32px;\n  height: 32px;')

with open(css_path, 'w') as f:
    f.write(css)

# 2. UPDATE JSX (Inline styles for images)
with open(jsx_path, 'r') as f:
    jsx = f.read()

# Fix hero inline heights for smaller items
jsx = jsx.replace("style={{top:'20%', left:'20%', height: 40}}", "style={{top:'20%', left:'20%', height: 70}}")
jsx = jsx.replace("style={{top:'80%', left:'20%', height: 80}}", "style={{top:'80%', left:'20%', height: 120}}")
jsx = jsx.replace("style={{top:'75%', left:'50%', height: 60}}", "style={{top:'75%', left:'50%', height: 100}}")
jsx = jsx.replace("style={{top:'85%', left:'80%', height: 50}}", "style={{top:'85%', left:'80%', height: 90}}")

# Fix mistake items inline height
jsx = jsx.replace("style={{height:40, marginBottom:8, objectFit: 'contain'}}", "style={{height:80, marginBottom:16, objectFit: 'contain'}}")

with open(jsx_path, 'w') as f:
    f.write(jsx)

print("Scaled up images in PaymentFinanceSection")
