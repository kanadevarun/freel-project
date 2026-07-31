import os

jsx_draft_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/PaymentFinanceSection.jsx'
css_draft_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/PaymentFinanceSection.css'
page_jsx_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/src/pages/public/trade-intelligence/guides/import-export-basics/Page.jsx'
page_css_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/src/pages/public/trade-intelligence/guides/import-export-basics/Page.css'

with open(jsx_draft_path, 'r') as f:
    jsx_draft = f.read()
    
with open(css_draft_path, 'r') as f:
    css_draft = f.read()

# Update CSS
with open(page_css_path, 'a') as f:
    f.write('\n\n' + css_draft)

# Update JSX
with open(page_jsx_path, 'r') as f:
    page_jsx = f.read()

# 1. Insert the component code before 'export default function ImportExportBasicsPage()'
insert_pos_1 = page_jsx.find('export default function ImportExportBasicsPage()')
page_jsx = page_jsx[:insert_pos_1] + jsx_draft + '\n\n' + page_jsx[insert_pos_1:]

# 2. Add the <PaymentFinanceSection /> call inside the component before the last closing div.
# Looking for:
#       <IncotermsSection />
# 
#     </div>
#   );
# }

call_to_insert = """
      {/* ══════════════════════════════════════
          SECTION 8: PAYMENT & FINANCE
      ══════════════════════════════════════ */}
      <PaymentFinanceSection />
"""

insert_pos_2 = page_jsx.find('    </div>\n  );\n}')
if insert_pos_2 != -1:
    page_jsx = page_jsx[:insert_pos_2] + call_to_insert + '\n' + page_jsx[insert_pos_2:]

with open(page_jsx_path, 'w') as f:
    f.write(page_jsx)

print("Successfully injected PaymentFinanceSection into Page.jsx and Page.css")
