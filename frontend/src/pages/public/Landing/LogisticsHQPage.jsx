import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import './LogisticsHQ.css';

/* ════════════════════════════════════════════════════════════════
   LOGISTICSHQ — Marketing Website
   Narrative: Problem → Chaos → What If? → LogisticsHQ
   ════════════════════════════════════════════════════════════════ */

function useReveal() {
  useEffect(() => {
    const els = document.querySelectorAll('.lhq-reveal');
    const obs = new IntersectionObserver(
      (entries) => entries.forEach(e => {
        if (e.isIntersecting) { e.target.classList.add('visible'); obs.unobserve(e.target); }
      }),
      { threshold: 0.1, rootMargin: '0px 0px -40px 0px' }
    );
    els.forEach(el => obs.observe(el));
    return () => obs.disconnect();
  }, []);
}

function Navbar() {
  const [scrolled, setScrolled] = useState(false);
  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 40);
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);
  const scrollTo = (id) => document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' });

  return (
    <nav className={`lhq-navbar${scrolled ? ' scrolled' : ''}`}>
      <a href="#" className="lhq-nav-logo" id="lhq-logo">
        <div className="lhq-nav-logo-icon">⚡</div>
        LogisticsHQ
      </a>
      <div className="lhq-nav-links">
        <button className="lhq-nav-link" onClick={() => scrollTo('lhq-how-it-works')}>How It Works</button>
        <button className="lhq-nav-link" onClick={() => scrollTo('lhq-workspaces')}>Product</button>
        <button className="lhq-nav-link" onClick={() => scrollTo('lhq-ai')}>AI</button>
        <button className="lhq-nav-link" onClick={() => scrollTo('lhq-cta')}>Contact</button>
        <Link to="/contact" className="lhq-nav-link lhq-nav-cta" id="lhq-nav-book-demo">Book a Demo</Link>
      </div>
    </nav>
  );
}

function Hero() {
  return (
    <section className="lhq-hero" id="lhq-hero">
      <div className="lhq-hero-bg">
        <div className="lhq-hero-glow-1" />
        <div className="lhq-hero-glow-2" />
        <div className="lhq-hero-grid" />
      </div>
      <div className="lhq-hero-content">
        <div className="lhq-hero-badge">
          <span className="lhq-hero-badge-dot" />
          <span>For Freight Forwarders</span>
        </div>
        <h1 className="lhq-hero-title">
          Freight forwarding<br />
          wasn't supposed to be<br />
          <span className="lhq-title-accent">this complicated.</span>
        </h1>
        <div className="lhq-hero-chaos-row">
          <span className="lhq-chaos-chip">📧 Customer Emails</span>
          <span className="lhq-chaos-divider">·</span>
          <span className="lhq-chaos-chip">📑 PDF Contracts</span>
          <span className="lhq-chaos-divider">·</span>
          <span className="lhq-chaos-chip">📊 Spreadsheets</span>
          <span className="lhq-chaos-divider">·</span>
          <span className="lhq-chaos-chip">🌐 Carrier Portals</span>
          <span className="lhq-chaos-divider">·</span>
          <span className="lhq-chaos-chip">💬 WhatsApp Quotes</span>
        </div>
        <p className="lhq-hero-sub">
          Too many systems. Too many handoffs. Too much manual work.
          Your team is talented — they shouldn't be spending their day copying data between tools.
        </p>
        <p className="lhq-hero-pivot">There has to be a better way.</p>
        <div className="lhq-hero-cta-row">
          <button className="lhq-btn-primary" id="lhq-hero-see-how"
            onClick={() => document.getElementById('lhq-how-it-works')?.scrollIntoView({ behavior: 'smooth' })}>
            See How It Works
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M7 17l9.2-9.2M17 17V7H7"/></svg>
          </button>
          <Link to="/contact" className="lhq-btn-secondary" id="lhq-hero-book-demo">Book a Demo</Link>
        </div>
      </div>
      <div className="lhq-hero-scroll" aria-hidden="true">
        <span className="lhq-scroll-text">Scroll to explore</span>
        <div className="lhq-scroll-line" />
      </div>
    </section>
  );
}

function EmailSection() {
  return (
    <section className="lhq-email-section" id="lhq-how-it-works">
      <div className="lhq-email-grid">
        <div className="lhq-reveal">
          <p className="lhq-label">The Email That Starts It All</p>
          <div className="lhq-email-window">
            <div className="lhq-email-titlebar">
              <div className="lhq-email-dots">
                <div className="lhq-email-dot lhq-dot-r" />
                <div className="lhq-email-dot lhq-dot-y" />
                <div className="lhq-email-dot lhq-dot-g" />
              </div>
              <div className="lhq-email-url">📧 Inbox — sales@yourff.com</div>
            </div>
            <div className="lhq-email-body">
              <div className="lhq-email-meta">
                <div className="lhq-email-meta-row">From: <span>procurement@acimanufacturing.com</span></div>
                <div className="lhq-email-meta-row">To: <span>sales@yourff.com</span></div>
                <div className="lhq-email-subject">Rate Request — August Shipment</div>
              </div>
              <div className="lhq-email-text">
                <p>Hi Team,</p>
                <p>Please provide your best rate for the following:</p>
                <div className="lhq-email-highlight">
                  <div><span className="lhq-hl-key">Equipment:&nbsp;&nbsp;</span><span className="lhq-hl-val">3 × 40HC</span></div>
                  <div><span className="lhq-hl-key">Origin:&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</span><span className="lhq-hl-val">Nhava Sheva, India</span></div>
                  <div><span className="lhq-hl-key">Destination:&nbsp;</span><span className="lhq-hl-val">Hamburg, Germany</span></div>
                  <div><span className="lhq-hl-key">Commodity:&nbsp;&nbsp;</span><span className="lhq-hl-val">Industrial Electronics</span></div>
                  <div><span className="lhq-hl-key">Target Date:&nbsp;</span><span className="lhq-hl-val">August 20, 2025</span></div>
                </div>
                <p>Please confirm transit time and free days.</p>
                <p style={{ color: '#94A3B8', fontSize: '13px' }}>Regards,<br />Priya Mehta — ACI Manufacturing</p>
              </div>
            </div>
          </div>
        </div>

        <div className="lhq-reveal lhq-delay-2 lhq-email-right">
          <h2 className="lhq-email-step-title">One email. Two very different workflows.</h2>
          <p className="lhq-email-step-desc">
            This is how every shipment begins. What happens next determines how fast your team responds — and whether you win the deal.
          </p>
          <div className="lhq-today-label">⚠ Today's Reality</div>
          <div className="lhq-manual-steps">
            {[
              ['📖', 'Sales reads the email and re-types key details'],
              ['✏️', 'Manually creates a new RFQ form'],
              ['🔍', 'Realises weight and HS code are missing'],
              ['📤', 'Sends follow-up email to customer'],
              ['⏳', 'Waits 18+ hours for customer to reply'],
              ['➡️', 'Updates RFQ manually, forwards to Pricing team'],
            ].map(([icon, text], i) => (
              <div className="lhq-manual-step" key={i}>
                <span className="lhq-manual-step-icon">{icon}</span>
                <span className="lhq-manual-step-text">{text}</span>
              </div>
            ))}
          </div>
          <div className="lhq-ai-label">✨ With LogisticsHQ</div>
          <div className="lhq-ai-steps">
            {[
              'AI reads the email and extracts all shipment details',
              'Detects missing info (weight, HS code) — flags instantly',
              'RFQ created automatically in seconds',
              'Pricing team notified the moment it is ready',
              'Full audit trail from first email to final quote',
            ].map((text, i) => (
              <div className="lhq-ai-step" key={i}>
                <div className="lhq-ai-check">✓</div>
                <span className="lhq-ai-step-text">{text}</span>
              </div>
            ))}
          </div>
          <div className="lhq-insight-box">
            <p>
              <strong>A simple customer request</strong> became a 4-step, multi-day internal workflow before pricing even started. With LogisticsHQ, it becomes a structured RFQ in seconds.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}

function PricingProblem() {
  const portals = [
    {
      icon: '🚢', name: 'Maersk Spot Portal', type: 'Spot Rate',
      steps: ['Enter origin port', 'Enter destination port', 'Select equipment & date', 'View rate — hope it matches your contract terms'],
      pain: '⚠ Rate changes daily. What you quote may not hold tomorrow.'
    },
    {
      icon: '📑', name: 'MSC Contract PDF', type: 'Contract Rate',
      steps: ['Open 47-page PDF', 'Ctrl+F for "Nhava Sheva"', 'Find applicable note quote section', 'Identify surcharges in footnotes'],
      pain: '⚠ Every carrier formats their contract differently.'
    },
    {
      icon: '📊', name: 'Hapag-Lloyd Rate Sheet', type: 'Contract Rate',
      steps: ['Find latest Excel in shared drive', 'Search by port pair', 'Check validity date', 'Confirm rate hasn\'t expired'],
      pain: '⚠ Who updated this? Is this still the active version?'
    },
    {
      icon: '💬', name: 'WhatsApp / Email Quotes', type: 'Note Quote',
      steps: ['Check WhatsApp group messages', 'Search inbox by carrier name', 'Find last rate message', 'Manually copy into the RFQ'],
      pain: '⚠ Completely unstructured. Impossible to audit.'
    },
  ];
  return (
    <section className="lhq-section lhq-section-dark" id="lhq-pricing">
      <div className="lhq-container">
        <div className="lhq-reveal" style={{ textAlign: 'center', marginBottom: '64px' }}>
          <p className="lhq-label">The Pricing Problem</p>
          <h2 className="lhq-section-title lhq-title-dark">Finding a rate is never<br />just finding a rate.</h2>
          <p className="lhq-section-sub lhq-sub-dark" style={{ margin: '0 auto' }}>
            For every RFQ, your pricing team opens at least four different systems. Each one has a different format, a different access flow, and a different problem.
          </p>
        </div>
        <div className="lhq-portals-grid">
          {portals.map((p, i) => (
            <div className="lhq-reveal lhq-portal-card" key={i} style={{ transitionDelay: `${i * 0.1}s` }}>
              <div className="lhq-portal-header">
                <div className="lhq-portal-icon">{p.icon}</div>
                <div>
                  <div className="lhq-portal-name">{p.name}</div>
                  <div className="lhq-portal-type">{p.type}</div>
                </div>
              </div>
              <div className="lhq-portal-steps">
                {p.steps.map((step, j) => (
                  <div className="lhq-portal-step" key={j}>
                    <div className="lhq-portal-step-num">{j + 1}</div>
                    {step}
                  </div>
                ))}
              </div>
              <div className="lhq-portal-pain">{p.pain}</div>
            </div>
          ))}
        </div>
        <div className="lhq-reveal lhq-pp-insight">
          <p className="lhq-pp-insight-text">
            "Every carrier formats their contracts differently.<br />Different surcharge codes. Different footnote structures.<br />This is not a simple search problem."
          </p>
          <p className="lhq-pp-insight-sub">
            LogisticsHQ unifies contract rates, spot rates, and pricing rules into one workflow — so your team stops searching and starts quoting.
          </p>
        </div>
      </div>
    </section>
  );
}

function TeamChaos() {
  const messages = [
    { s: 'SALES', sc: 'lhq-sender-sales', bc: 'lhq-bubble-sales', t: 'Customer needs a quote urgently for RFQ-2025-089. Can you check this ASAP?' },
    { s: 'PRICING', sc: 'lhq-sender-pricing', bc: 'lhq-bubble-pricing', t: "I need complete cargo details first. What's the weight and HS code?" },
    { s: 'SALES', sc: 'lhq-sender-sales', bc: 'lhq-bubble-sales', t: "It's all in the email I forwarded you." },
    { s: 'PRICING', sc: 'lhq-sender-pricing', bc: 'lhq-bubble-pricing', t: 'I have 47 unread emails. Which one?' },
    { s: 'SALES', sc: 'lhq-sender-sales', bc: 'lhq-bubble-sales', t: '3×40HC, Electronics, FOB Nhava Sheva.' },
    { s: 'PRICING', sc: 'lhq-sender-pricing', bc: 'lhq-bubble-pricing', t: 'Free days? Special terms? Which carrier should I check first?' },
    { s: 'OPERATIONS', sc: 'lhq-sender-ops', bc: 'lhq-bubble-ops', t: "Where's the final confirmed RFQ? I need to check vessel availability." },
    { s: '📱 CUSTOMER', sc: 'lhq-sender-customer', bc: 'lhq-bubble-customer', t: 'Hi team, any update on the rate request? We need to confirm this week.' },
  ];
  return (
    <section className="lhq-section lhq-section-surface" id="lhq-team-chaos">
      <div className="lhq-container">
        <div className="lhq-chat-grid">
          <div className="lhq-reveal">
            <div className="lhq-chat-window">
              <div className="lhq-chat-titlebar">
                <span style={{ fontSize: '16px' }}>💬</span>
                <span className="lhq-chat-title-text">RFQ-2025-089 · Team Thread</span>
                <div className="lhq-chat-status-dot" />
              </div>
              <div className="lhq-chat-messages">
                {messages.map((m, i) => (
                  <div className="lhq-chat-msg" key={i}>
                    <span className={`lhq-chat-sender ${m.sc}`}>{m.s}</span>
                    <div className={`lhq-chat-bubble ${m.bc}`}>{m.t}</div>
                  </div>
                ))}
              </div>
            </div>
          </div>
          <div className="lhq-reveal lhq-delay-2">
            <p className="lhq-label lhq-label-amber">Sound Familiar?</p>
            <h2 className="lhq-section-title lhq-title-light" style={{ marginBottom: '20px' }}>Three teams.<br />One shipment.<br />Too many handoffs.</h2>
            <p className="lhq-section-sub lhq-sub-light" style={{ marginBottom: '32px' }}>
              Everyone is working hard. But no one is working on the same version of the same information. Sales doesn't have rates. Pricing doesn't have complete details. Operations doesn't have approval.
            </p>
            <div style={{ background: 'white', borderRadius: '14px', border: '1px solid #E2E8F0', padding: '20px 24px', boxShadow: '0 2px 12px rgba(0,0,0,0.05)' }}>
              <p style={{ fontWeight: 700, fontSize: '15px', color: '#0F172A', marginBottom: '12px' }}>The real cost of this chaos:</p>
              {[
                ['⏱', 'Slower response → lost deals to faster competitors'],
                ['🔁', 'Duplicated effort across Sales, Pricing, Ops'],
                ['❌', 'Errors from manual data re-entry'],
                ['😤', 'Frustrated customers, burnt-out teams'],
              ].map(([icon, text], i) => (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: '10px', padding: '8px 0', borderBottom: i < 3 ? '1px solid #F1F5F9' : 'none', fontSize: '14px', color: '#475569' }}>
                  <span>{icon}</span><span>{text}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

function WhatIf() {
  const steps = [
    { icon: '📧', text: 'Customer email arrives', ai: false },
    { icon: '🧠', text: <><strong>AI reads and understands</strong> the request instantly</>, ai: true },
    { icon: '📋', text: 'RFQ created automatically with all extracted details', ai: true },
    { icon: '🔍', text: <><strong>Missing fields detected</strong> — follow-up flagged to Sales</>, ai: true },
    { icon: '⚡', text: 'Pricing Agent activated', ai: true },
    { icon: '📑', text: 'Contract rates searched across all carriers', ai: true },
    { icon: '🌐', text: 'Spot rates fetched in real-time', ai: true },
    { icon: '📏', text: 'Pricing rules applied, sell price calculated', ai: true },
    { icon: '✅', text: 'Margin validated against your rules', ai: true },
    { icon: '⚠️', text: 'Anomaly flagged → Human review requested', ai: false, interrupt: true },
    { icon: '👤', text: 'Pricing Manager reviews and approves the quote', ai: false },
    { icon: '🎯', text: 'Quotation sent to customer', ai: false },
  ];
  return (
    <section className="lhq-whatif-section" id="lhq-whatif">
      <div className="lhq-whatif-bg" />
      <div className="lhq-whatif-content">
        <div className="lhq-reveal lhq-whatif-header">
          <p className="lhq-whatif-eyebrow">The LogisticsHQ Way</p>
          <h2 className="lhq-whatif-title">What if the workflow<br /><span className="lhq-whatif-accent">could run itself?</span></h2>
        </div>
        <div className="lhq-flow-container lhq-reveal">
          {steps.map((step, i) => (
            <div key={i}>
              {step.interrupt ? (
                <>
                  <div className="lhq-flow-arrow" />
                  <div className="lhq-flow-interrupt">
                    <span style={{ fontSize: '20px' }}>⚠️</span>
                    <span className="lhq-flow-interrupt-text">Human review — AI surfaces the exact decision you need to make</span>
                  </div>
                  <div className="lhq-flow-arrow" />
                </>
              ) : (
                <>
                  {i > 0 && <div className="lhq-flow-arrow" />}
                  <div className={`lhq-flow-step${step.ai ? ' lhq-flow-step-ai' : ''}`}>
                    <span className="lhq-flow-step-icon">{step.icon}</span>
                    <span className="lhq-flow-step-text">{step.text}</span>
                  </div>
                </>
              )}
            </div>
          ))}
        </div>
        <div className="lhq-reveal" style={{ textAlign: 'center', marginTop: '56px' }}>
          <p style={{ fontSize: '16px', color: 'rgba(255,255,255,0.45)', marginBottom: '24px' }}>
            This is not a concept. This is how LogisticsHQ works today.
          </p>
          <Link to="/contact" className="lhq-btn-primary" id="lhq-whatif-cta">
            See It In Action
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
          </Link>
        </div>
      </div>
    </section>
  );
}

function ProductReveal() {
  return (
    <section className="lhq-reveal-section" id="lhq-product">
      <div className="lhq-reveal">
        <p className="lhq-reveal-eyebrow">The Platform</p>
        <h2 className="lhq-reveal-title">Meet <span className="lhq-title-accent">LogisticsHQ.</span></h2>
        <p className="lhq-reveal-sub">
          The operating system for modern freight forwarding. One connected platform for Sales, Pricing, and Operations — with AI running through every workflow.
        </p>
      </div>
      <div className="lhq-reveal lhq-delay-2 lhq-product-mockup">
        <div className="lhq-mockup-bar">
          <div className="lhq-mockup-dots">
            <div className="lhq-mockup-dot" style={{ background: '#FF5F57' }} />
            <div className="lhq-mockup-dot" style={{ background: '#FEBC2E' }} />
            <div className="lhq-mockup-dot" style={{ background: '#28C840' }} />
          </div>
          <div className="lhq-mockup-url">🔒 app.logisticshq.in/dashboard/rfqs/2025-089</div>
        </div>
        <div className="lhq-mockup-body">
          <div className="lhq-mock-sidebar">
            <div className="lhq-mock-sidebar-brand"><span>⚡</span> LogisticsHQ</div>
            {[['🏠','Dashboard',false],['📋','RFQs',true],['📑','Contracts',false],['📊','Rates',false],['👥','Leads',false],['📈','Reports',false]].map(([icon,label,active],i) => (
              <div key={i} className={`lhq-mock-nav-item${active?' active':''}`}><span>{icon}</span> {label}</div>
            ))}
          </div>
          <div className="lhq-mock-main">
            <div className="lhq-mock-header">
              <span className="lhq-mock-rfq-title">RFQ-2025-089 · Pricing Workspace</span>
              <span className="lhq-mock-stage">⚡ PRICING</span>
            </div>
            <div className="lhq-mock-info-grid">
              {[['Route','INNSA → DEHAM'],['Equipment','3 × 40HC'],['Incoterms','FOB'],['AI Status','✓ Draft Ready']].map(([label,val],i) => (
                <div key={i} className="lhq-mock-info-item">
                  <div className="lhq-mock-info-label">{label}</div>
                  <div className="lhq-mock-info-val" style={label==='AI Status'?{color:'#00BFA5'}:{}}>{val}</div>
                </div>
              ))}
            </div>
            {[
              { c:'Maersk', rec:true, buy:'$2,800', sell:'$3,136', margin:'12.0%', transit:'18 days' },
              { c:'MSC', rec:false, buy:'$2,600', sell:'$2,912', margin:'12.0%', transit:'22 days' },
              { c:'CMA CGM', rec:false, buy:'$2,950', sell:'$3,304', margin:'12.0%', transit:'20 days' },
            ].map((q,i) => (
              <div key={i} className="lhq-mock-quote-row" style={{opacity: i===0?1:0.55}}>
                <div className="lhq-mock-carrier">
                  {q.c} {q.rec && <span className="lhq-mock-rec">AI REC</span>}
                  <span style={{fontSize:'12px',color:'rgba(255,255,255,0.3)',fontWeight:400}}>· {q.transit}</span>
                </div>
                <div className="lhq-mock-price-col">
                  <div className="lhq-mock-buy">Buy {q.buy}</div>
                  <div className="lhq-mock-sell">{q.sell}</div>
                  <div className="lhq-mock-margin">{q.margin} margin</div>
                </div>
              </div>
            ))}
            <div style={{display:'flex',gap:'10px',marginTop:'16px'}}>
              <button style={{flex:1,padding:'10px',background:'#00BFA5',color:'#0A0F1E',border:'none',borderRadius:'8px',fontSize:'14px',fontWeight:700,cursor:'pointer',fontFamily:'Outfit,sans-serif'}}>✓ Approve Quote</button>
              <button style={{padding:'10px 16px',background:'rgba(255,255,255,0.07)',color:'rgba(255,255,255,0.6)',border:'1px solid rgba(255,255,255,0.1)',borderRadius:'8px',fontSize:'14px',cursor:'pointer',fontFamily:'Outfit,sans-serif'}}>Edit</button>
            </div>
          </div>
          <div className="lhq-mock-right">
            <div className="lhq-mock-panel-title">AI Activity Log</div>
            {[
              ['lhq-dot-green','Email parsed by AI','2 min ago'],
              ['lhq-dot-teal','RFQ auto-created','2 min ago'],
              ['lhq-dot-teal','Contract rates fetched','1 min ago'],
              ['lhq-dot-teal','Spot rates fetched','1 min ago'],
              ['lhq-dot-teal','Pricing rules applied','58s ago'],
              ['lhq-dot-amber','3 quotes ready for review','Just now'],
              ['lhq-dot-gray','Awaiting approval...',''],
            ].map(([dot,event,time],i) => (
              <div key={i} className="lhq-mock-tl-item">
                <div className={`lhq-mock-tl-dot ${dot}`} />
                <div>
                  <div className="lhq-mock-tl-event">{event}</div>
                  {time && <div className="lhq-mock-tl-time">{time}</div>}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

function TeamWorkspaces() {
  const ws = [
    { cls:'ws-sales', icon:'🤝', title:'Sales Workspace', badge:'live', desc:'AI reads customer emails, extracts requirements, creates RFQs, and flags missing information — automatically.', flow:['Customer email or inquiry arrives','AI extracts shipment requirements','Missing fields flagged instantly','RFQ created with one click','Pricing team notified immediately'] },
    { cls:'ws-pricing', icon:'📊', title:'Pricing Workspace', badge:'live', desc:'Contract rates, spot rates, and AI-powered quote recommendations — in one workspace, with full human control.', flow:['RFQ received from Sales','AI searches contract & spot rates','Pricing rules applied automatically','Margin validated against thresholds','Human approves before quote is sent'] },
    { cls:'ws-ops', icon:'⚙️', title:'Operations Workspace', badge:'dev', desc:'From approved quotation to shipment execution. Carrier booking, documentation workflow, and status tracking.', flow:['Approved quote triggers booking','Carrier confirmation tracked','Documentation workflow initiated','Shipment status monitored','Customer updates automated'] },
    { cls:'ws-finance', icon:'💰', title:'Finance & Profitability', badge:'roadmap', desc:'Customer invoicing, carrier cost reconciliation, and margin analytics across all shipments and lanes.', flow:['Shipment completed → invoice generated','Carrier costs reconciled automatically','Margin calculated per shipment','Profitability by lane and customer','Anomaly detection on cost variances'] },
  ];
  const badges = { live:['lhq-ws-badge-live','● Live'], dev:['lhq-ws-badge-dev','⚙ In Development'], roadmap:['lhq-ws-badge-road','◎ Roadmap'] };
  return (
    <section className="lhq-section lhq-section-light" id="lhq-workspaces">
      <div className="lhq-container">
        <div className="lhq-reveal" style={{textAlign:'center',marginBottom:'56px'}}>
          <p className="lhq-label">The Platform</p>
          <h2 className="lhq-section-title lhq-title-light">Every team. One platform.</h2>
          <p className="lhq-section-sub lhq-sub-light" style={{margin:'0 auto'}}>
            LogisticsHQ is the connected workflow layer across your entire freight-forwarding organisation — from first email to final invoice.
          </p>
        </div>
        <div className="lhq-workspaces-grid">
          {ws.map((w,i) => {
            const [bc,bl] = badges[w.badge];
            return (
              <div key={i} className={`lhq-reveal lhq-workspace-card ${w.cls}`} style={{transitionDelay:`${i*0.1}s`}}>
                <div className="lhq-ws-header">
                  <div className="lhq-ws-icon">{w.icon}</div>
                  <span className={`lhq-ws-badge ${bc}`}>{bl}</span>
                </div>
                <h3 className="lhq-ws-title">{w.title}</h3>
                <p className="lhq-ws-desc">{w.desc}</p>
                <div className="lhq-ws-flow">
                  {w.flow.map((step,j) => <div key={j} className="lhq-ws-flow-step"><div className="lhq-ws-dot" />{step}</div>)}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}

function AISection() {
  const steps = [
    { done:true, text:<>Read RFQ <em>#2025-089</em></> },
    { done:true, text:<>Extracted: <em>INNSA → DEHAM, 40HC, FOB</em></> },
    { done:true, text:<>Contract rates searched — <em>3 carriers matched</em></> },
    { done:true, text:<>Spot rates fetched — <em>Maersk, MSC, CMA CGM</em></> },
    { done:true, text:<>Lane promo applied — <em>INNSA-DEHAM +12%</em></> },
    { done:true, text:<>Sell price calculated — <em>$3,136 / container</em></> },
    { done:true, text:<>Margin validated — <em>12.0% (above 8% floor ✓)</em></> },
    { done:false, warn:true, text:'Awaiting Pricing Manager approval before sending' },
  ];
  return (
    <section className="lhq-section lhq-section-dark" id="lhq-ai">
      <div className="lhq-container">
        <div className="lhq-ai-grid">
          <div>
            <div className="lhq-reveal">
              <p className="lhq-label">Agentic AI</p>
              <h2 className="lhq-section-title lhq-title-dark">AI that works with your team, not around it.</h2>
              <p className="lhq-section-sub lhq-sub-dark">
                LogisticsHQ's AI agents reason through workflows, use your business data, retrieve rates, validate margins — and hand work back to humans when judgment is required.
              </p>
            </div>
            <div className="lhq-ai-principles">
              {[
                ['👁','Transparent','Every AI action is logged. You can always see what the AI did, why it did it, and what data it used.'],
                ['🎛','Controllable','Human review before any quote is finalised. You set the pricing rules. The AI executes. You decide.'],
                ['📋','Auditable','Full activity timeline for every AI action — every rate source, every rule applied, every decision made.'],
              ].map(([icon,title,desc],i) => (
                <div key={i} className={`lhq-reveal lhq-ai-principle`} style={{transitionDelay:`${(i+1)*0.12}s`}}>
                  <div className="lhq-ai-principle-icon">{icon}</div>
                  <div>
                    <div className="lhq-ai-principle-title">{title}</div>
                    <div className="lhq-ai-principle-desc">{desc}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
          <div className="lhq-reveal lhq-delay-2 lhq-ai-panel">
            <div className="lhq-ai-panel-header">
              <span style={{fontSize:'16px'}}>🤖</span>
              <span className="lhq-ai-panel-title">Pricing Agent — Live Run</span>
              <div className="lhq-ai-panel-live">
                <div style={{width:'6px',height:'6px',borderRadius:'50%',background:'#00BFA5'}} />
                Running
              </div>
            </div>
            <div className="lhq-ai-rfq-block">
              <div>RFQ: <span>#2025-089</span></div>
              <div>Route: <span>INNSA → DEHAM</span></div>
              <div>Equipment: <span>40HC × 3</span></div>
              <div>Incoterms: <span>FOB</span></div>
            </div>
            <div className="lhq-ai-steps-list">
              {steps.map((step,i) => (
                <div key={i} className={`lhq-ai-tl-step${step.warn?' lhq-ai-warn':''}`}>
                  <div className={`lhq-ai-tl-icon ${step.warn?'lhq-ai-icon-warn':step.done?'lhq-ai-icon-done':''}`}>
                    {step.warn ? '⚠' : step.done ? '✓' : '○'}
                  </div>
                  <div className="lhq-ai-tl-text">{step.text}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

function HITLSection() {
  return (
    <section className="lhq-section lhq-section-surface" id="lhq-hitl">
      <div className="lhq-container">
        <div className="lhq-hitl-grid">
          <div className="lhq-reveal">
            <div className="lhq-hitl-card">
              <div className="lhq-hitl-card-header">
                <span style={{fontSize:'18px'}}>⚠️</span>
                <span className="lhq-hitl-card-title">Pricing Review Required</span>
                <span className="lhq-hitl-badge">ACTION NEEDED</span>
              </div>
              <div className="lhq-hitl-body">
                {[['Carrier','Maersk — INNSA→DEHAM'],['Buy Price','$9,200 / container'],['Suggested Sell','$9,728']].map(([label,val],i) => (
                  <div key={i} className="lhq-hitl-row">
                    <span className="lhq-hitl-row-label">{label}</span>
                    <span className="lhq-hitl-row-val">{val}</span>
                  </div>
                ))}
                <div className="lhq-hitl-row">
                  <span className="lhq-hitl-row-label">Margin</span>
                  <span className="lhq-hitl-row-val danger">5.7% ← Below 8% floor</span>
                </div>
                <div className="lhq-hitl-reason">
                  <div className="lhq-hitl-reason-label">🤖 AI Reasoning</div>
                  <div className="lhq-hitl-reason-text">
                    Buy price of $9,200 exceeds the standard flag threshold for this lane. Margin of 5.7% falls below the 8% minimum set in your pricing rules. Review contract rates or adjust sell price before sending to customer.
                  </div>
                </div>
                <div className="lhq-hitl-actions">
                  <button className="lhq-hitl-btn lhq-hitl-approve">✓ Approve</button>
                  <button className="lhq-hitl-btn lhq-hitl-edit">✏ Edit Price</button>
                  <button className="lhq-hitl-btn lhq-hitl-reject">✗ Reject</button>
                </div>
              </div>
            </div>
          </div>
          <div className="lhq-reveal lhq-delay-2">
            <p className="lhq-label">Human-in-the-Loop</p>
            <h2 className="lhq-section-title lhq-title-light" style={{marginBottom:'20px'}}>AI handles the work.<br />People handle the judgment.</h2>
            <p className="lhq-section-sub lhq-sub-light" style={{marginBottom:'32px'}}>
              When margins are thin, when a price looks unusual, or when conditions fall outside your rules — LogisticsHQ stops and surfaces the exact decision you need to make.
            </p>
            {[
              ['🧠','AI reasons autonomously through the pricing workflow'],
              ['⚠️','Flags anomalies based on your own pricing rules'],
              ['👤','Human reviews the specific decision — with full context'],
              ['✅','Approves, edits, or rejects — then the workflow continues'],
            ].map(([icon,text],i) => (
              <div key={i} style={{display:'flex',alignItems:'center',gap:'14px',padding:'14px 0',borderBottom:i<3?'1px solid #F1F5F9':'none'}}>
                <span style={{fontSize:'22px'}}>{icon}</span>
                <span style={{fontSize:'15px',color:'#334155',fontWeight:500}}>{text}</span>
              </div>
            ))}
            <div style={{marginTop:'32px',padding:'20px',background:'#F0FDF4',border:'1px solid #BBF7D0',borderRadius:'14px'}}>
              <p style={{fontSize:'14px',color:'#166534',lineHeight:'1.65'}}>
                <strong style={{display:'block',marginBottom:'4px'}}>The result:</strong>
                Your team focuses on decisions that require human judgment — not on work the AI can handle. Speed increases. Errors decrease. Margin visibility improves.
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

function FinalCTA() {
  return (
    <section className="lhq-final-cta" id="lhq-cta">
      <div className="lhq-final-cta-bg" />
      <div className="lhq-reveal lhq-final-cta-content">
        <h2 className="lhq-cta-title">
          Freight forwarding is complex.<br />
          <span className="lhq-title-accent">Your workflow doesn't have to be.</span>
        </h2>
        <p className="lhq-cta-sub">
          LogisticsHQ is in early access for freight-forwarding teams. Book a demo to see how it works for your lanes, your carriers, and your team.
        </p>
        <div className="lhq-cta-row">
          <Link to="/contact" className="lhq-btn-primary" id="lhq-final-book-demo">
            Book a Demo
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
          </Link>
          <button className="lhq-btn-secondary" id="lhq-final-explore"
            onClick={() => document.getElementById('lhq-how-it-works')?.scrollIntoView({ behavior: 'smooth' })}>
            See How It Works
          </button>
        </div>
        <div className="lhq-social-row">
          <a href="https://linkedin.com" target="_blank" rel="noopener noreferrer" className="lhq-social-link" id="lhq-social-linkedin">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><path d="M16 8a6 6 0 0 1 6 6v7h-4v-7a2 2 0 0 0-2-2 2 2 0 0 0-2 2v7h-4v-7a6 6 0 0 1 6-6zM2 9h4v12H2z"/><circle cx="4" cy="4" r="2"/></svg>
            LinkedIn
          </a>
          <span style={{color:'rgba(255,255,255,0.12)'}}>·</span>
          <a href="https://instagram.com" target="_blank" rel="noopener noreferrer" className="lhq-social-link" id="lhq-social-instagram">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="2" y="2" width="20" height="20" rx="5"/><circle cx="12" cy="12" r="4"/><circle cx="17.5" cy="6.5" r="1" fill="currentColor" stroke="none"/></svg>
            Instagram
          </a>
          <span style={{color:'rgba(255,255,255,0.12)'}}>·</span>
          <a href="https://youtube.com" target="_blank" rel="noopener noreferrer" className="lhq-social-link" id="lhq-social-youtube">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><path d="M22.54 6.42a2.78 2.78 0 0 0-1.95-1.96C18.88 4 12 4 12 4s-6.88 0-8.59.46a2.78 2.78 0 0 0-1.95 1.96A29 29 0 0 0 1 12a29 29 0 0 0 .46 5.58A2.78 2.78 0 0 0 3.41 19.6C5.12 20 12 20 12 20s6.88 0 8.59-.46a2.78 2.78 0 0 0 1.95-1.96A29 29 0 0 0 23 12a29 29 0 0 0-.46-5.58z"/><polygon fill="#0A0F1E" points="9.75 15.02 15.5 12 9.75 8.98 9.75 15.02"/></svg>
            YouTube
          </a>
        </div>
      </div>
    </section>
  );
}

function Footer() {
  return (
    <footer className="lhq-footer">
      <div className="lhq-footer-inner">
        <div className="lhq-footer-grid">
          <div>
            <div className="lhq-footer-brand">
              <div className="lhq-footer-logo-icon">⚡</div>
              <span className="lhq-footer-logo-name">LogisticsHQ</span>
            </div>
            <p className="lhq-footer-tagline">The operating system for modern freight forwarding. Connecting Sales, Pricing, and Operations.</p>
          </div>
          <div>
            <div className="lhq-footer-col-title">Product</div>
            {['Sales Workspace','Pricing Workspace','AI Agent','Operations','Finance'].map((l,i) => <a key={i} href="#" className="lhq-footer-link">{l}</a>)}
          </div>
          <div>
            <div className="lhq-footer-col-title">Solutions</div>
            {['RFQ Management','Rate Intelligence','Contract Intelligence','Pricing Agent'].map((l,i) => <a key={i} href="#" className="lhq-footer-link">{l}</a>)}
          </div>
          <div>
            <div className="lhq-footer-col-title">Company</div>
            <Link to="/about" className="lhq-footer-link">About</Link>
            <Link to="/contact" className="lhq-footer-link">Book a Demo</Link>
            <Link to="/contact" className="lhq-footer-link">Contact</Link>
            <Link to="/login" className="lhq-footer-link">Login</Link>
          </div>
        </div>
        <div className="lhq-footer-bottom">
          <span className="lhq-footer-copy">© 2025 LogisticsHQ. All rights reserved.</span>
          <div className="lhq-footer-links">
            <a href="#" className="lhq-footer-link" style={{marginBottom:0,fontSize:'12px'}}>Privacy</a>
            <a href="#" className="lhq-footer-link" style={{marginBottom:0,fontSize:'12px'}}>Terms</a>
          </div>
        </div>
      </div>
    </footer>
  );
}

export default function LogisticsHQPage() {
  useReveal();
  return (
    <div className="lhq-marketing">
      <Navbar />
      <main>
        <Hero />
        <EmailSection />
        <PricingProblem />
        <TeamChaos />
        <WhatIf />
        <ProductReveal />
        <TeamWorkspaces />
        <AISection />
        <HITLSection />
        <FinalCTA />
      </main>
      <Footer />
    </div>
  );
}
