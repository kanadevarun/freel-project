import re

file_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/src/pages/public/trade-intelligence/guides/import-export-basics/Page.css'
with open(file_path, 'r') as f:
    content = f.read()

new_css = """/* ════════════════════════════════════════════════════════════════
   SECTION 7: INCOTERMS 2020
   Prefix: inco-
════════════════════════════════════════════════════════════════ */
.inco-section {
  padding: 80px 5%;
  background: #f8fafc;
  border-top: 1px solid #e2e8f0;
  font-family: 'Outfit', sans-serif;
}
.inco-inner {
  max-width: 1280px;
  margin: 0 auto;
}

/* ── Hero ── */
.inco-hero-row {
  margin-bottom: 60px;
}
.inco-hero-header {
  max-width: 800px;
  margin-bottom: 40px;
}
.inco-badge {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  background: #ffffff;
  border: 1px solid #bfdbfe;
  padding: 6px 20px;
  border-radius: 99px;
  margin-bottom: 24px;
}
.inco-badge-num {
  font-size: 0.85rem;
  font-weight: 900;
  color: #ffffff;
  background: #2563eb;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}
.inco-badge-text {
  font-size: 0.85rem;
  font-weight: 800;
  color: #1e3a8a;
  letter-spacing: 1px;
}
.inco-hero-header h2 {
  font-size: clamp(2.5rem, 4vw, 3.8rem);
  font-weight: 900;
  line-height: 1.1;
  color: #0f172a;
  margin-bottom: 16px;
  letter-spacing: -0.02em;
}
.inco-hero-header h2 span {
  background: linear-gradient(90deg, #2563eb, #1d4ed8);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
.inco-hero-header p {
  font-size: 1.15rem;
  color: #475569;
  line-height: 1.6;
}

/* Hero Journey Collage */
.inco-hero-journey {
  width: 100%;
  position: relative;
  background-image: radial-gradient(#cbd5e1 2px, transparent 2px);
  background-size: 24px 24px;
  background-position: center 40px;
  border-radius: 24px;
  margin-top: 20px;
  padding: 40px 0;
}
.inco-hj-nodes {
  position: relative;
  height: 180px;
  width: 100%;
  margin-bottom: 20px;
}
.inco-hj-node {
  position: absolute;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  transform: translateX(-50%);
}
.inco-hj-img {
  height: 90px;
  object-fit: contain;
  mix-blend-mode: multiply;
}
.inco-hj-label {
  font-size: 0.75rem;
  font-weight: 800;
  color: #0f172a;
  text-transform: uppercase;
  text-align: center;
  margin-bottom: 12px;
  background: rgba(255,255,255,0.8);
  padding: 2px 6px;
  border-radius: 4px;
}
.inco-hj-truck {
  position: absolute;
  height: 40px;
  mix-blend-mode: multiply;
  transform: translateX(-50%);
}

/* Hero Responsibility Bars */
.inco-resp-bars-hero {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 0;
  width: 100%;
}
.inco-rb-row {
  display: flex;
  align-items: center;
  gap: 16px;
  position: relative;
}
.inco-rb-label {
  width: 180px;
  font-size: 0.75rem;
  font-weight: 800;
  text-transform: uppercase;
  color: #475569;
  flex-shrink: 0;
}
.inco-rb-line {
  height: 10px;
  border-radius: 5px;
  position: relative;
}
.inco-rb-line::after {
  content: '';
  position: absolute;
  right: -6px;
  top: -3px;
  border-left: 10px solid inherit;
  border-top: 8px solid transparent;
  border-bottom: 8px solid transparent;
}
.inco-rb-seller { background: #3b82f6; }
.inco-rb-seller::after { border-left-color: #3b82f6; }
.inco-rb-buyer { background: #22c55e; }
.inco-rb-buyer::after { border-left-color: #22c55e; }
.inco-rb-risk { background: #f59e0b; }
.inco-rb-risk::after { border-left-color: #f59e0b; }
.inco-rb-cost { background: #8b5cf6; }
.inco-rb-cost::after { border-left-color: #8b5cf6; }

.inco-rb-pin-cost {
  position: absolute;
  top: -10px;
  width: 28px;
  height: 28px;
  background: #8b5cf6;
  color: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 800;
}

/* ── Section Titles ── */
.inco-section-title {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 1.4rem;
  font-weight: 800;
  color: #1e3a8a;
  margin-bottom: 8px;
}
.inco-st-num {
  width: 32px;
  height: 32px;
  background: #2563eb;
  color: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.9rem;
}
.inco-section-desc {
  font-size: 0.95rem;
  color: #475569;
  margin-bottom: 24px;
  margin-left: 48px;
}

/* ── Explore Incoterms Cards ── */
.inco-explore {
  margin-bottom: 60px;
}
.inco-cards-container {
  position: relative;
}
.inco-sea-label {
  position: absolute;
  top: -24px;
  right: 120px;
  font-size: 0.75rem;
  font-weight: 800;
  color: #0ea5e9;
  text-transform: uppercase;
}
.inco-explore-grid {
  display: grid;
  grid-template-columns: repeat(11, 1fr);
  gap: 8px;
  background: #ffffff;
  padding: 16px;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
}
.inco-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  border-right: 1px dashed #cbd5e1;
  padding: 0 4px;
}
.inco-card:last-child {
  border-right: none;
}
.inco-card h4 {
  font-size: 1rem;
  font-weight: 900;
  color: #1e293b;
  margin-bottom: 8px;
}
.inco-card-img {
  height: 48px;
  object-fit: contain;
  margin-bottom: 8px;
  mix-blend-mode: multiply;
}
.inco-card-name {
  font-size: 0.65rem;
  font-weight: 800;
  color: #334155;
  margin-bottom: 6px;
  line-height: 1.2;
}
.inco-card-desc {
  font-size: 0.6rem;
  color: #64748b;
  margin-bottom: 12px;
  line-height: 1.3;
  flex: 1;
}
.inco-card-risk {
  font-size: 0.55rem;
  font-weight: 800;
  color: #d97706;
  border: 1px solid #f59e0b;
  padding: 4px;
  border-radius: 4px;
  margin-bottom: 8px;
  width: 100%;
}
.inco-card-icons {
  display: flex;
  gap: 4px;
  align-items: center;
  height: 20px;
}

/* ── Timeline ── */
.inco-timeline {
  margin-bottom: 60px;
}
.inco-tl-container {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  padding: 32px 40px;
}
.inco-tl-flow {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 40px;
  position: relative;
}
.inco-tl-flow::after {
  content: '';
  position: absolute;
  bottom: 24px;
  left: 20px;
  right: 20px;
  height: 2px;
  background-image: linear-gradient(to right, #cbd5e1 50%, transparent 50%);
  background-size: 12px 2px;
  z-index: 1;
}
.inco-tl-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  z-index: 2;
  background: #ffffff;
  padding: 0 12px;
}
.inco-tl-img {
  height: 50px;
  mix-blend-mode: multiply;
}
.inco-tl-label {
  font-size: 0.65rem;
  font-weight: 800;
  color: #0f172a;
}
.inco-resp-bars-tl {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.inco-rb-icon {
  height: 20px;
  object-fit: contain;
}
.inco-rb-icon-cost {
  width: 20px;
  height: 20px;
  background: #8b5cf6;
  color: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 800;
}
.inco-tl-legend {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 24px;
  margin-top: 24px;
  font-size: 0.8rem;
  font-weight: 700;
  color: #475569;
}
.inco-leg-item {
  display: flex;
  align-items: center;
  gap: 8px;
}
.inco-leg-color {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

/* ── Deep Dive ── */
.inco-deep {
  margin-bottom: 60px;
}
.inco-deep-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.inco-dd-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 24px;
}
.inco-dd-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}
.inco-dd-title {
  font-size: 1.1rem;
  font-weight: 900;
  color: #0f172a;
  max-width: 120px;
}
.inco-dd-img {
  height: 60px;
  mix-blend-mode: multiply;
}
.inco-dd-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.inco-dd-bullet {
  font-size: 0.8rem;
  color: #2563eb;
  margin-bottom: 4px;
}
.inco-dd-bullet strong {
  color: #1e293b;
}
.inco-dd-item p {
  font-size: 0.8rem;
  color: #475569;
  line-height: 1.4;
  margin: 0;
  padding-left: 12px;
}

/* ── Bottom Rows ── */
.inco-bottom-row {
  display: grid;
  grid-template-columns: 1.5fr 1fr;
  gap: 20px;
  margin-bottom: 40px;
}
.inco-box {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  padding: 32px;
}
.inco-bottom-col2 {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.inco-rw-flow {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 32px;
}
.inco-rw-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  text-align: center;
}
.inco-rw-img {
  height: 48px;
  mix-blend-mode: multiply;
}
.inco-rw-label {
  font-size: 0.65rem;
  font-weight: 800;
  color: #1e293b;
}
.inco-rw-bars {
  position: relative;
}
.inco-rw-bar-row {
  display: flex;
  width: 100%;
}
.inco-rw-seller {
  width: 50%;
  background: #3b82f6;
  color: #fff;
  padding: 12px;
  border-radius: 8px 0 0 8px;
}
.inco-rw-buyer {
  width: 50%;
  background: #22c55e;
  color: #fff;
  padding: 12px;
  border-radius: 0 8px 8px 0;
  text-align: right;
}
.inco-rw-bar-label {
  font-size: 0.75rem;
  font-weight: 800;
  margin-bottom: 4px;
}
.inco-rw-bar-desc {
  font-size: 0.65rem;
}
.inco-rw-pin {
  position: absolute;
  top: -12px;
  left: 50%;
  transform: translateX(-50%);
}

.inco-mistake-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.inco-mistake-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 0.85rem;
  color: #1e293b;
  font-weight: 600;
}
.inco-mistake-x {
  width: 20px;
  height: 20px;
  background: #ef4444;
  color: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 800;
  flex-shrink: 0;
}

.inco-tip {
  background: #f0fdf4;
  border-color: #bbf7d0;
}
.inco-tip p {
  font-size: 0.85rem;
  color: #166534;
  margin-bottom: 12px;
  line-height: 1.5;
}
.inco-tip-roadtrip {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
.inco-tip-roadtrip img {
  height: 120px;
  border-radius: 12px;
}

/* ── Stats ── */
.inco-stats-row {
  display: flex;
  align-items: center;
  margin-bottom: 60px;
}
.inco-stats-grid {
  display: flex;
  gap: 16px;
  flex: 1;
}
.inco-stat-box {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 16px;
}
.inco-stat-icon {
  height: 48px;
  mix-blend-mode: multiply;
}
.inco-stat-val {
  font-size: 1.5rem;
  font-weight: 900;
  color: #1e40af;
}
.inco-stat-title {
  font-size: 0.75rem;
  font-weight: 800;
  color: #1e293b;
}
.inco-stat-desc {
  font-size: 0.7rem;
  color: #64748b;
}

/* ── Final CTA ── */
.inco-final-cta {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  padding: 32px 40px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.inco-fc-left {
  display: flex;
  align-items: center;
  gap: 24px;
}
.inco-fc-icon {
  width: 64px;
  height: 64px;
  background: #eff6ff;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.inco-fc-text h3 {
  font-size: 1.4rem;
  font-weight: 900;
  color: #0f172a;
  margin-bottom: 8px;
}
.inco-fc-text p {
  font-size: 0.95rem;
  color: #475569;
}
.inco-fc-btn {
  background: #2563eb;
  color: #fff;
  border: none;
  padding: 16px 32px;
  border-radius: 12px;
  font-size: 1rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 12px;
}
.inco-fc-btn:hover {
  background: #1d4ed8;
}
"""

pattern = re.compile(r'/\* ════════════════════════════════════════════════════════════════\n   SECTION 7: INCOTERMS 2020\n   Prefix: inco-\n════════════════════════════════════════════════════════════════ \*/\n.inco-section \{.*', re.DOTALL)
new_content = pattern.sub(new_css, content)

with open(file_path, 'w') as f:
    f.write(new_content)
