import re
import os

jsx_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/src/pages/public/trade-intelligence/guides/import-export-basics/Page.jsx'

with open(jsx_path, 'r') as f:
    content = f.read()

# Update PAY_MISTAKES
old_mistakes = """const PAY_MISTAKES = [
  'Sending goods before payment',
  'Ignoring currency fluctuation',
  'Wrong beneficiary details',
  'Incorrect LC documents',
  'Missing SWIFT codes'
];"""

new_mistakes = """const PAY_MISTAKES = [
  { text: 'Sending goods before payment', img: '/images/import-export/transport/ship.png' },
  { text: 'Ignoring currency fluctuation', img: '/images/import-export/documents_new/doc_insurance.png' },
  { text: 'Wrong beneficiary details', img: '/images/import-export/documents_new/doc_commercial-invoice.png' },
  { text: 'Incorrect LC documents', img: '/images/import-export/incoterms_new/doc_3.png' },
  { text: 'Missing SWIFT codes', img: '/images/import-export/incoterms_new/doc_0.png' }
];"""

content = content.replace(old_mistakes, new_mistakes)

# Update mistake render loop
old_mistake_loop = """            <div className="pay-mistakes-grid">
               {PAY_MISTAKES.map((m, i) => (
                 <div key={i} className="pay-mistake-item">
                   <div className="pay-mistake-warn">!</div>
                   <div className="pay-mistake-img"></div>
                   <p>{m}</p>
                 </div>
               ))}
            </div>"""

new_mistake_loop = """            <div className="pay-mistakes-grid">
               {PAY_MISTAKES.map((m, i) => (
                 <div key={i} className="pay-mistake-item">
                   <div className="pay-mistake-warn">!</div>
                   <img src={m.img} style={{height:40, marginBottom:8, objectFit: 'contain'}} alt=""/>
                   <p>{m.text}</p>
                 </div>
               ))}
            </div>"""

content = content.replace(old_mistake_loop, new_mistake_loop)

# Add SVG to hero diagram
old_hero = """             <div className="pay-hero-diagram">
                {/* Simulated Workflow using absolute positioning */}
                <img src="/images/import-export/modules/factory.png" className="pay-hd-item" style={{top:'40%', left:'0%'}} alt="Exporter"/>"""

new_hero = """             <div className="pay-hero-diagram">
                <svg width="100%" height="100%" style={{position:'absolute', top:0, left:0, zIndex: 0}}>
                   <ellipse cx="50%" cy="50%" rx="45%" ry="35%" fill="none" stroke="#2563eb" strokeWidth="1.5" strokeDasharray="6 6" />
                </svg>
                {/* Simulated Workflow using absolute positioning */}
                <img src="/images/import-export/modules/factory.png" className="pay-hd-item" style={{top:'40%', left:'0%'}} alt="Exporter"/>"""

content = content.replace(old_hero, new_hero)

with open(jsx_path, 'w') as f:
    f.write(content)

print("Updated Page.jsx")
