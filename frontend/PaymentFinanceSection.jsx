const PAY_METHODS = [
  { name: 'Advance Payment', desc: 'Payment before shipment.', riskLabel: 'Low Risk', riskLevel: 'low', img: '/images/import-export/documents_new/doc_invoice.png' },
  { name: 'Open Account', desc: 'Payment after delivery.', riskLabel: 'High Risk', riskLevel: 'high', img: '/images/import-export/documents_new/doc_customs-declaration.png' },
  { name: 'Letter of Credit', desc: 'Bank guarantee of payment.', riskLabel: 'Low-Medium Risk', riskLevel: 'low-medium', img: '/images/import-export/incoterms_new/doc_2.png' },
  { name: 'Documentary Collection', desc: 'Documents against payment.', riskLabel: 'Medium Risk', riskLevel: 'medium', img: '/images/import-export/documents_new/doc_packing-list.png' },
  { name: 'Consignment', desc: 'Payment after sale.', riskLabel: 'High Risk', riskLevel: 'high', img: '/images/import-export/incoterms_new/hero_1.png' }
];

const PAY_LC_STEPS = [
  { label: 'Buyer', img: '/images/import-export/participants/importer.png' },
  { label: "Importer's Bank", img: '/images/import-export/participants/customs.png' },
  { label: 'Letter of Credit', img: '/images/import-export/incoterms_new/doc_2.png' },
  { label: "Exporter's Bank", img: '/images/import-export/participants/customs.png' },
  { label: 'Exporter Ships Goods', img: '/images/import-export/transport/ship.png' },
  { label: 'Documents Submitted', img: '/images/import-export/documents/bill-of-lading.png' },
  { label: 'Bank Verifies', img: '/images/import-export/incoterms_new/hero_4.png' }, // placeholder
  { label: 'Payment Released', img: '/images/import-export/incoterms_new/doc_0.png' } // placeholder
];

const PAY_ECOSYSTEM = [
  { label: 'Importer', img: '/images/import-export/participants/importer.png' },
  { label: 'Exporter', img: '/images/import-export/participants/exporter.png' },
  { label: "Importer's Bank", img: '/images/import-export/participants/customs.png' },
  { label: "Exporter's Bank", img: '/images/import-export/participants/customs.png' },
  { label: 'Insurance Company', img: '/images/import-export/documents_new/doc_insurance.png' },
  { label: 'Freight Forwarder', img: '/images/import-export/participants/freight-forwarder.png' }
];

const PAY_MISTAKES = [
  'Sending goods before payment',
  'Ignoring currency fluctuation',
  'Wrong beneficiary details',
  'Incorrect LC documents',
  'Missing SWIFT codes'
];

const PAY_STATS = [
  { val: '80%', desc: 'Global trade uses banks' },
  { val: '200+', desc: 'Countries connected by SWIFT' },
  { val: '$5T+', desc: 'Trade finance annually' },
  { val: 'Millions', desc: 'International payments daily' }
];

function PaymentFinanceSection() {
  return (
    <section className="pay-section" id="payment-finance">
      <div className="pay-inner">
        {/* HERO SECTION */}
        <div className="pay-hero">
          <div className="pay-hero-left">
            <motion.div className="pay-badge" initial={{opacity:0, y:12}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
              <div className="pay-badge-num">08</div>
              <div className="pay-badge-text">PAYMENT & FINANCE</div>
            </motion.div>
            <motion.h2 initial={{opacity:0, y:16}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.1}}>
              Master International<br/><span>Trade Payments</span>
            </motion.h2>
            <motion.p initial={{opacity:0, y:16}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.2}}>
              Understand how buyers and sellers exchange money securely across borders.
            </motion.p>
            <motion.p initial={{opacity:0, y:16}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.3}} style={{marginTop:16}}>
              Learn payment methods, trade finance, banks, letters of credit, and international risk.
            </motion.p>
          </div>
          <div className="pay-hero-right">
             <div className="pay-hero-diagram">
                {/* Simulated Workflow using absolute positioning */}
                <img src="/images/import-export/modules/factory.png" className="pay-hd-item" style={{top:'40%', left:'0%'}} alt="Exporter"/>
                <div className="pay-hd-label" style={{top:'70%', left:'0%'}}>Exporter</div>
                
                <img src="/images/import-export/documents_new/doc_invoice.png" className="pay-hd-item" style={{top:'20%', left:'20%', height: 40}} alt="Invoice"/>
                <div className="pay-hd-label" style={{top:'35%', left:'20%'}}>Invoice</div>

                <img src="/images/import-export/participants/customs.png" className="pay-hd-item" style={{top:'10%', left:'40%'}} alt="Bank"/>
                <div className="pay-hd-label" style={{top:'40%', left:'40%'}}>Exporter's Bank</div>

                <div className="pay-hd-swift" style={{top:'20%', left:'60%'}}>SWIFT</div>

                <img src="/images/import-export/participants/customs.png" className="pay-hd-item" style={{top:'10%', left:'80%'}} alt="Bank"/>
                <div className="pay-hd-label" style={{top:'40%', left:'80%'}}>Importer's Bank</div>

                <img src="/images/import-export/participants/importer.png" className="pay-hd-item" style={{top:'40%', left:'95%'}} alt="Buyer"/>
                <div className="pay-hd-label" style={{top:'70%', left:'95%'}}>Buyer</div>

                <img src="/images/import-export/transport/ship.png" className="pay-hd-item" style={{top:'80%', left:'20%', height: 80}} alt="Ship"/>
                <img src="/images/import-export/transport/airplane.png" className="pay-hd-item" style={{top:'75%', left:'50%', height: 60}} alt="Plane"/>
                <img src="/images/import-export/transport/truck.png" className="pay-hd-item" style={{top:'85%', left:'80%', height: 50}} alt="Truck"/>
             </div>
          </div>
        </div>

        {/* 01 & 02 ROW */}
        <div className="pay-row-2col">
          {/* 01 WHY PAYMENTS MATTER */}
          <div className="pay-card">
            <div className="pay-section-title">
              <div className="pay-st-num">01</div>
              WHY PAYMENTS MATTER
            </div>
            <div className="pay-wm-flow">
              <div className="pay-wm-item">
                <img src="/images/import-export/transport/truck.png" alt=""/>
                <span>Goods Move</span>
              </div>
              <ArrowRight size={16} color="#94a3b8"/>
              <div className="pay-wm-item">
                <img src="/images/import-export/documents_new/doc_bill-of-lading.png" alt=""/>
                <span>Documents Move</span>
              </div>
              <ArrowRight size={16} color="#94a3b8"/>
              <div className="pay-wm-item pay-wm-money">
                <div className="pay-money-icon">$</div>
                <span>Money Moves</span>
              </div>
            </div>
            <div className="pay-wm-question">
              <div className="pay-wm-qicon">?</div>
              <div className="pay-wm-qtext">
                <strong>Key Question</strong>
                <span>When should payment happen?</span>
              </div>
              <div className="pay-wm-qrisk">
                 <div>Payment Risk</div>
                 <div>Currency Risk</div>
              </div>
            </div>
          </div>

          {/* 02 PAYMENT METHODS OVERVIEW */}
          <div className="pay-card pay-card-wide">
            <div className="pay-section-title">
              <div className="pay-st-num">02</div>
              PAYMENT METHODS OVERVIEW
            </div>
            <div className="pay-methods-grid">
              {PAY_METHODS.map((m, i) => (
                <div key={i} className="pay-method-item">
                  <img src={m.img} alt=""/>
                  <strong>{m.name}</strong>
                  <p>{m.desc}</p>
                  <div className={`pay-risk-badge pay-risk-${m.riskLevel}`}>{m.riskLabel}</div>
                </div>
              ))}
            </div>
            <div className="pay-risk-legend">
               <span>Risk Level:</span>
               <span><div className="pay-rl-dot" style={{background:'#22c55e'}}/> Low</span>
               <span><div className="pay-rl-dot" style={{background:'#eab308'}}/> Low-Medium</span>
               <span><div className="pay-rl-dot" style={{background:'#f97316'}}/> Medium</span>
               <span><div className="pay-rl-dot" style={{background:'#ef4444'}}/> High</span>
            </div>
          </div>
        </div>

        {/* 03 & 04 ROW */}
        <div className="pay-row-2col">
          {/* 03 PAYMENT RISK MATRIX */}
          <div className="pay-card">
            <div className="pay-section-title">
              <div className="pay-st-num">03</div>
              PAYMENT RISK MATRIX
            </div>
            <div className="pay-matrix">
               <div className="pay-mx-y">High<br/><br/><br/>Seller Protection<br/><br/><br/>Low</div>
               <div className="pay-mx-content">
                  <div className="pay-mx-point" style={{top:'10%', left:'10%'}}>Advance Payment</div>
                  <div className="pay-mx-point" style={{top:'20%', left:'40%'}}>Letter of Credit</div>
                  <div className="pay-mx-point" style={{top:'50%', left:'40%'}}>Documentary Collection</div>
                  <div className="pay-mx-point" style={{top:'60%', left:'80%'}}>Open Account</div>
                  <div className="pay-mx-point" style={{top:'85%', left:'70%'}}>Consignment</div>
               </div>
               <div className="pay-mx-x">Low <span style={{marginLeft:40, marginRight:40}}>Buyer Convenience</span> High</div>
            </div>
          </div>

          {/* 04 LETTER OF CREDIT EXPLAINED */}
          <div className="pay-card pay-card-wide">
            <div className="pay-section-title">
              <div className="pay-st-num">04</div>
              LETTER OF CREDIT EXPLAINED
            </div>
            <div className="pay-lc-timeline">
               <div className="pay-lc-line"></div>
               {PAY_LC_STEPS.map((s, i) => (
                 <div key={i} className="pay-lc-step">
                   <div className="pay-lc-num">{i+1}</div>
                   <img src={s.img} alt=""/>
                   <span>{s.label}</span>
                 </div>
               ))}
            </div>
            <div className="pay-lc-footer">
              The bank acts as a trusted intermediary and guarantees payment to the exporter once all terms and documents are verified.
            </div>
          </div>
        </div>

        {/* 05 & 06 ROW */}
        <div className="pay-row-2col">
          {/* 05 TRADE FINANCE ECOSYSTEM */}
          <div className="pay-card">
            <div className="pay-section-title">
              <div className="pay-st-num">05</div>
              TRADE FINANCE ECOSYSTEM
            </div>
            <div className="pay-eco-grid">
               {PAY_ECOSYSTEM.map((s, i) => (
                 <div key={i} className="pay-eco-item">
                   <img src={s.img} alt=""/>
                   <span>{s.label}</span>
                 </div>
               ))}
            </div>
            <div className="pay-eco-footer">
              Multiple parties work together to move goods and secure payments.
            </div>
          </div>

          {/* 06 PAYMENT TIMELINE */}
          <div className="pay-card pay-card-wide">
            <div className="pay-section-title">
              <div className="pay-st-num">06</div>
              PAYMENT TIMELINE (COMPARE METHODS)
            </div>
            <div className="pay-pt-top">
               {['Purchase Order', 'Invoice', 'Production', 'Shipment', 'Documents', 'Customs', 'Delivery', 'Payment Released'].map((l, i) => (
                 <div key={i} className="pay-pt-node">
                   <div className="pay-pt-icon"></div>
                   <span>{l}</span>
                 </div>
               ))}
            </div>
            <div className="pay-pt-bars">
               <div className="pay-pt-row">
                 <div className="pay-pt-label">Advance Payment</div>
                 <div className="pay-pt-bar" style={{left: '0%', width: '25%', background: '#22c55e'}}></div>
               </div>
               <div className="pay-pt-row">
                 <div className="pay-pt-label">Letter of Credit</div>
                 <div className="pay-pt-bar" style={{left: '25%', width: '60%', background: '#eab308'}}></div>
               </div>
               <div className="pay-pt-row">
                 <div className="pay-pt-label">Documentary Collection</div>
                 <div className="pay-pt-bar" style={{left: '50%', width: '35%', background: '#f97316'}}></div>
               </div>
               <div className="pay-pt-row">
                 <div className="pay-pt-label">Open Account</div>
                 <div className="pay-pt-bar" style={{left: '85%', width: '15%', background: '#ef4444'}}></div>
               </div>
            </div>
          </div>
        </div>

        {/* 07 & 08 ROW */}
        <div className="pay-row-2col">
          {/* 07 CURRENCY & EXCHANGE */}
          <div className="pay-card">
            <div className="pay-section-title">
              <div className="pay-st-num">07</div>
              CURRENCY & EXCHANGE
            </div>
            <div className="pay-curr-inner">
               <div className="pay-curr-left">
                  <div className="pay-curr-title">Common Currencies</div>
                  <div className="pay-curr-icons">
                    <div className="pay-curr-ic">$<br/><span>USD</span></div>
                    <div className="pay-curr-ic">€<br/><span>EUR</span></div>
                    <div className="pay-curr-ic">£<br/><span>GBP</span></div>
                    <div className="pay-curr-ic">¥<br/><span>JPY</span></div>
                    <div className="pay-curr-ic">د.إ<br/><span>AED</span></div>
                  </div>
               </div>
               <div className="pay-curr-right">
                  <div className="pay-curr-title">Key Considerations</div>
                  <ul>
                    <li>Exchange Rate Fluctuations</li>
                    <li>FX Risk Management</li>
                    <li>Bank Charges & Fees</li>
                    <li>SWIFT Transfer</li>
                    <li>Currency Hedging</li>
                  </ul>
               </div>
            </div>
          </div>

          {/* 08 COMMON MISTAKES */}
          <div className="pay-card pay-card-wide">
            <div className="pay-section-title">
              <div className="pay-st-num">08</div>
              COMMON MISTAKES
            </div>
            <div className="pay-mistakes-grid">
               {PAY_MISTAKES.map((m, i) => (
                 <div key={i} className="pay-mistake-item">
                   <div className="pay-mistake-warn">!</div>
                   <div className="pay-mistake-img"></div>
                   <p>{m}</p>
                 </div>
               ))}
            </div>
          </div>
        </div>

        {/* 09 & 10 ROW */}
        <div className="pay-row-2col">
          {/* 09 REAL WORLD EXAMPLE */}
          <div className="pay-card pay-card-wide">
            <div className="pay-section-title">
              <div className="pay-st-num">09</div>
              REAL WORLD EXAMPLE
            </div>
            <div style={{fontSize:'0.8rem', color:'#64748b', marginBottom:16}}>Export from India to Germany (FOB Mumbai)</div>
            <div className="pay-rw-flow">
               {/* Simplified representation */}
               <div className="pay-rw-step"><img src="/images/import-export/participants/exporter.png" alt=""/>Indian Exporter</div>
               <ArrowRight size={16} color="#94a3b8"/>
               <div className="pay-rw-step"><img src="/images/import-export/participants/importer.png" alt=""/>German Buyer</div>
               <ArrowRight size={16} color="#94a3b8"/>
               <div className="pay-rw-step"><img src="/images/import-export/transport/port.png" alt=""/>FOB Mumbai</div>
               <ArrowRight size={16} color="#94a3b8"/>
               <div className="pay-rw-step"><div className="pay-rw-dollar">$40,000</div>Order Value</div>
               <ArrowRight size={16} color="#94a3b8"/>
               <div className="pay-rw-step"><img src="/images/import-export/incoterms_new/doc_2.png" alt=""/>Letter of Credit</div>
               <ArrowRight size={16} color="#94a3b8"/>
               <div className="pay-rw-step"><img src="/images/import-export/transport/ship.png" alt=""/>Shipment</div>
               <ArrowRight size={16} color="#94a3b8"/>
               <div className="pay-rw-step"><img src="/images/import-export/documents/bill-of-lading.png" alt=""/>Documents Submitted</div>
               <ArrowRight size={16} color="#94a3b8"/>
               <div className="pay-rw-step"><img src="/images/import-export/participants/customs.png" alt=""/>Bank Verifies Documents</div>
               <ArrowRight size={16} color="#94a3b8"/>
               <div className="pay-rw-step"><div className="pay-rw-dollar">$</div>Payment Released</div>
            </div>
          </div>

          {/* 10 BY THE NUMBERS */}
          <div className="pay-card">
            <div className="pay-section-title">
              <div className="pay-st-num">10</div>
              BY THE NUMBERS
            </div>
            <div className="pay-stats-grid">
               {PAY_STATS.map((s, i) => (
                 <div key={i} className="pay-stat-item">
                   <div className="pay-stat-val">{s.val}</div>
                   <div className="pay-stat-desc">{s.desc}</div>
                 </div>
               ))}
            </div>
          </div>
        </div>

        {/* FINAL CTA */}
        <motion.div className="pay-final-cta" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
          <div className="pay-fc-left">
            <div className="pay-fc-icon">🎓</div>
            <div className="pay-fc-text">
              <h3>Great progress!</h3>
              <p>You now understand how international payments work.</p>
            </div>
          </div>
          <button className="pay-fc-btn">
            Next Module: <strong>Packaging & Labelling</strong> <ArrowRight size={16} />
          </button>
        </motion.div>

      </div>
    </section>
  );
}
