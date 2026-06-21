import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import {
  Ship, Anchor, Shield, FileText, Globe,
  Play, Calendar, Users, Factory, Warehouse,
  Box, Lightbulb
} from 'lucide-react';
import './Page.css';
import { motion } from 'framer-motion';
import SmartSelector from './SmartSelector';
import RealShipmentSimulator from './RealShipmentSimulator';
import IncotermsMatrix from './IncotermsMatrix';
import RealTradeMistakes from './RealTradeMistakes';

/* ─────────────────────────────────────────
   DATA
───────────────────────────────────────── */
const JOURNEY = [
  { id: 1, label: 'Factory',          emoji: '🏭' },
  { id: 2, label: 'Export',           emoji: '🚛' },
  { id: 3, label: 'Origin Port',      emoji: '🏗️' },
  { id: 4, label: 'Ocean',            emoji: '🚢' },
  { id: 5, label: 'Destination Port', emoji: '🏗️' },
  { id: 6, label: 'Import',           emoji: '🛡️' },
  { id: 7, label: 'Warehouse',        emoji: '🏬' },
];

// riskAt  = transfer point index (1–7), the dot position on the bar
// costTo  = seller pays cost beyond risk (for C-terms)
const INCOTERMS = [
  // ── Sea Only ──────────────────────────────────────────────────────────────
  { code:'FAS', name:'Free Alongside Ship',       type:'sea',       riskAt:3,  seller:'Deliver goods alongside vessel at origin port',          buyer:'Loading, ocean freight, import clearance'  },
  { code:'FOB', name:'Free On Board',             type:'sea',       riskAt:4,  seller:'Export clearance + load goods on board vessel',         buyer:'Ocean freight, insurance, import clearance' },
  { code:'CFR', name:'Cost and Freight',          type:'sea',       riskAt:4,  costTo:5, seller:'Pays freight to destination port; risk at loading', buyer:'Bears risk from loading; handles import'   },
  { code:'CIF', name:'Cost, Insurance & Freight', type:'sea',       riskAt:4,  costTo:5, insurance:true, seller:'Pays freight + minimum insurance to destination',  buyer:'Bears risk from loading; handles import'   },
  // ── Any Mode ──────────────────────────────────────────────────────────────
  { code:'EXW', name:'Ex Works',                  type:'universal', riskAt:1,  seller:'Only makes goods available at their premises',           buyer:'All transport, export, ocean, import'       },
  { code:'FCA', name:'Free Carrier',              type:'universal', riskAt:3,  seller:'Delivers to named carrier at named place',               buyer:'Main carriage, insurance, import clearance' },
  { code:'CPT', name:'Carriage Paid To',          type:'universal', riskAt:4,  costTo:5, seller:'Pays carriage to named destination',               buyer:'Bears risk from first carrier; handles import' },
  { code:'CIP', name:'Carriage & Insurance Paid', type:'universal', riskAt:4,  costTo:5, insurance:true, seller:'Pays carriage + all-risk insurance to destination', buyer:'Bears risk from first carrier; handles import' },
  { code:'DAP', name:'Delivered At Place',        type:'universal', riskAt:6,  seller:'All transport to named destination, ready to unload',    buyer:'Unloading + import customs duties'          },
  { code:'DPU', name:'Delivered at Place Unloaded',type:'universal',riskAt:6.5,seller:'Delivers AND unloads at destination terminal',          buyer:'Import customs duties only'                 },
  { code:'DDP', name:'Delivered Duty Paid',       type:'universal', riskAt:7,  seller:'Entire journey door-to-door including import duties',   buyer:'Simply receives goods — zero hassle'        },
];

const STEPS = 7; // total journey steps

/* ─────────────────────────────────────────
   PAGE
───────────────────────────────────────── */
export default function IncotermsPage() {
  const [hovered, setHovered] = useState(null);

  useEffect(() => { window.scrollTo(0, 0); }, []);

  const seaTerms       = INCOTERMS.filter(t => t.type === 'sea');
  const universalTerms = INCOTERMS.filter(t => t.type === 'universal');

  return (
    <div className="inc-page">

      {/* ══════════════════════════════════════
          HERO
      ══════════════════════════════════════ */}
      <section className="inc-hero">
        <div className="inc-hero-bg">
          <img src="/images/hero_port_aerial.png" alt="" className="inc-hero-img" />
          <div className="inc-hero-overlay" />
          <div className="inc-hero-dots" />
        </div>

        <div className="inc-hero-grid">
          {/* Left */}
          <div className="inc-hero-left">
            <motion.div
              initial={{ opacity:0, y:20 }} animate={{ opacity:1, y:0 }} transition={{ duration:0.5 }}
              className="inc-hero-badge">
              <Globe size={14} /> TRADE INTELLIGENCE
            </motion.div>
            <motion.h1
              initial={{ opacity:0, y:20 }} animate={{ opacity:1, y:0 }} transition={{ duration:0.5, delay:0.1 }}
              className="inc-hero-title">
              Master Global Trade<br/>
              <span className="inc-blue-grad">Without Guesswork</span>
            </motion.h1>
            <motion.p
              initial={{ opacity:0, y:20 }} animate={{ opacity:1, y:0 }} transition={{ duration:0.5, delay:0.2 }}
              className="inc-hero-sub">
              Learn who pays, who ships, and where risk transfers across every international shipment.
            </motion.p>
            <motion.div
              initial={{ opacity:0, y:20 }} animate={{ opacity:1, y:0 }} transition={{ duration:0.5, delay:0.3 }}
              className="inc-hero-actions">
              <a href="#glance" className="inc-btn-primary">Explore Incoterms →</a>
              <button className="inc-btn-glass"><Play size={16} /> Watch Guide</button>
            </motion.div>
            <motion.div
              initial={{ opacity:0, y:20 }} animate={{ opacity:1, y:0 }} transition={{ duration:0.5, delay:0.4 }}
              className="inc-hero-stats">
              {[
                { icon:<FileText size={20} color="#4f46e5"/>, val:'11', lbl:'Incoterms' },
                { icon:<Calendar size={20} color="#10b981"/>, val:'2020', lbl:'Rules' },
                { icon:<Globe size={20} color="#14b8a6"/>, val:'190+', lbl:'Countries' },
                { icon:<Users size={20} color="#8b5cf6"/>, val:'Used', lbl:'Worldwide' },
              ].map((s,i) => (
                <div key={i} className="inc-stat">
                  {i>0 && <div className="inc-stat-div"/>}
                  {s.icon}
                  <div><strong>{s.val}</strong><span>{s.lbl}</span></div>
                </div>
              ))}
            </motion.div>
          </div>

          {/* Right */}
          <div className="inc-hero-right">
            <motion.div
              initial={{ opacity:0, scale:0.95 }} animate={{ opacity:1, scale:1 }}
              transition={{ duration:0.8, delay:0.3, type:'spring' }}
              className="inc-journey-card">
              {[
                { label:'Mumbai Factory',    sub:'Production Begins',    cls:'node-blue',   Icon: ()=><Factory size={18}/> },
                { label:'Port Terminal',     sub:'Export Clearance',     cls:'node-green',  Icon: ()=><Anchor size={18}/> },
                { label:'Ocean Freight',     sub:'International Shipping',cls:'node-blue',  Icon: ()=><Ship size={18}/> },
                { label:'Hamburg Port',      sub:'Import Clearance',     cls:'node-green',  Icon: ()=><Anchor size={18}/> },
                { label:'Customer Warehouse',sub:'Goods Delivered',      cls:'node-orange', Icon: ()=><Warehouse size={18}/> },
              ].map((n,i,arr) => (
                <div key={i}>
                  <div className="inc-jnode">
                    <div className={`inc-jnode-icon ${n.cls}`}><n.Icon/></div>
                    <div><strong>{n.label}</strong><span>{n.sub}</span></div>
                  </div>
                  {i < arr.length-1 && <div className="inc-jdiv"/>}
                </div>
              ))}
              <div className="inc-jfooter">
                <div className="inc-jpill">One Shipment. <strong>Many Rules.</strong></div>
              </div>
            </motion.div>

            {[
              { cls:'inc-fc-tl bbot-purple', icon:<Box size={18}/>, iconCls:'bg-purple text-purple', code:'EXW', sub:'Buyer Handles Everything', delay:0.5 },
              { cls:'inc-fc-tr bbot-green',  icon:<Ship size={18}/>,icon:'', iconCls:'bg-green text-green',   code:'FOB', sub:'Risk Transfers On Board',   delay:0.7 },
              { cls:'inc-fc-bl bbot-orange', icon:<Shield size={18}/>,iconCls:'bg-orange text-orange', code:'CIF', sub:'Insurance Included',       delay:0.9 },
              { cls:'inc-fc-br bbot-blue',   icon:<Box size={18}/>, iconCls:'bg-blue text-blue',     code:'DDP', sub:'Delivered Duty Paid',        delay:1.1 },
            ].map((f,i) => (
              <motion.div key={i} className={`inc-float ${f.cls}`}
                initial={{ opacity:0, x: i%2===0 ? -20 : 20 }}
                animate={{ opacity:1, x:0 }}
                transition={{ duration:0.8, delay:f.delay }}>
                <div className="inc-float-top">
                  <div className={`inc-float-icon ${f.iconCls}`}>{f.icon}</div>
                  <strong>{f.code}</strong>
                </div>
                <span>{f.sub}</span>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ══════════════════════════════════════
          11 RULES AT A GLANCE
      ══════════════════════════════════════ */}
      <section className="inc-glance" id="glance">
        <div className="inc-glance-inner">

          {/* Header */}
          <div className="inc-g-header">
            <div className="inc-g-eyebrow"><Globe size={13}/> THE 11 RULES AT A GLANCE</div>
            <h2 className="inc-g-title">The 11 Rules At A Glance</h2>
            <p className="inc-g-sub">One visual map showing who owns the journey.</p>
          </div>

          {/* ── Top Journey Strip ── */}
          <div className="inc-journey-strip">
            {JOURNEY.map((step, i) => {
              let dot = 'dot-neutral';
              if (hovered) {
                if (step.id <= hovered.riskAt) dot = 'dot-seller';
                else if (hovered.costTo && step.id <= hovered.costTo) dot = 'dot-cost';
                else dot = 'dot-buyer';
              }
              return (
                <div key={step.id} className="inc-js-node">
                  <div className={`inc-js-icon ${hovered && step.id <= hovered.riskAt ? 'js-active-seller' : hovered && hovered.costTo && step.id <= hovered.costTo ? 'js-active-cost' : hovered ? 'js-active-buyer' : ''}`}>
                    <span className="inc-js-emoji">{step.emoji}</span>
                  </div>
                  {i < JOURNEY.length - 1 && (
                    <div className={`inc-js-arrow ${hovered && step.id < hovered.riskAt ? 'arr-seller' : hovered && hovered.costTo && step.id < hovered.costTo ? 'arr-cost' : hovered ? 'arr-buyer' : ''}`}>→</div>
                  )}
                  <div className="inc-js-label">{step.label}</div>
                  <div className={`inc-js-dot ${dot}`}></div>
                </div>
              );
            })}
          </div>

          {/* ── Legend ── */}
          <div className="inc-legend">
            <div className="inc-leg-item"><div className="inc-leg-dot leg-seller"/><span>Seller Responsibility</span></div>
            <div className="inc-leg-item"><div className="inc-leg-dot leg-buyer"/><span>Buyer Responsibility</span></div>
          </div>

          {/* ── Matrix ── */}
          <div className="inc-matrix">

            {/* Left sidebar */}
            <div className="inc-matrix-sidebar">
              <div className="inc-sb-card">
                <div className="inc-sb-icon">🌊</div>
                <div className="inc-sb-title">SEA FREIGHT ONLY</div>
                <div className="inc-sb-desc">Used only for sea and inland waterway transport.</div>
              </div>
              <div className="inc-sb-spacer"/>
              <div className="inc-sb-card inc-sb-universal">
                <div className="inc-sb-icon">🚚</div>
                <div className="inc-sb-title">ALL TRANSPORT MODES</div>
                <div className="inc-sb-desc">Can be used for any mode of transport.</div>
              </div>
            </div>

            {/* Right rows */}
            <div className="inc-matrix-rows">
              {/* Sea terms */}
              <div className="inc-rows-group inc-rows-group--sea">
                {seaTerms.map((term, i) => (
                  <IncotermBarRow
                    key={term.code}
                    term={term}
                    rowIndex={i}
                    isHovered={hovered?.code === term.code}
                    onEnter={() => setHovered(term)}
                    onLeave={() => setHovered(null)}
                  />
                ))}
              </div>

              {/* Divider */}
              <div className="inc-matrix-sep"/>

              {/* Universal terms */}
              <div className="inc-rows-group inc-rows-group--uni">
                {universalTerms.map((term, i) => (
                  <IncotermBarRow
                    key={term.code}
                    term={term}
                    rowIndex={i + seaTerms.length}
                    isHovered={hovered?.code === term.code}
                    onEnter={() => setHovered(term)}
                    onLeave={() => setHovered(null)}
                  />
                ))}
              </div>
            </div>

          </div>

          {/* ── How to Read ── */}
          <div className="inc-how-to-read">
            <Lightbulb size={16} color="#6366f1"/>
            <strong>How to read:</strong>
            <span>The point where the blue line ends is where risk transfers from seller to buyer.</span>
          </div>

        </div>
      </section>

      {/* ══════════════════════════════════════
          SMART SELECTOR (Configurator)
      ══════════════════════════════════════ */}
      <SmartSelector />

      {/* ══════════════════════════════════════
          REAL SHIPMENT SIMULATOR
      ══════════════════════════════════════ */}
      <RealShipmentSimulator />

      {/* ══════════════════════════════════════
          COMPARISON MATRIX
      ══════════════════════════════════════ */}
      <IncotermsMatrix />

      {/* ══════════════════════════════════════
          REAL TRADE MISTAKES
      ══════════════════════════════════════ */}
      <RealTradeMistakes />

      {/* ── CTA ── */}
      <section className="inc-cta">
        <div className="inc-cta-overlay"/>
        <div className="inc-cta-body">
          <h2>Ready To Master Global Trade?</h2>
          <p>Join thousands of logistics professionals building smarter supply chains.</p>
          <div className="inc-cta-btns">
            <Link to="/knowledge" className="inc-cbtn-primary">Explore Trade Intelligence</Link>
            <Link to="/contact"   className="inc-cbtn-secondary">Contact Experts</Link>
          </div>
        </div>
      </section>

    </div>
  );
}

/* ─────────────────────────────────────────
   INCOTERM BAR ROW
───────────────────────────────────────── */
function IncotermBarRow({ term, rowIndex = 0, isHovered, onEnter, onLeave }) {
  // Use a 0-6 scale for 7 points so left:0 is Factory and left:100% is Warehouse.
  const segments = STEPS - 1; 

  // The end point of the blue bar is costTo (if defined), else riskAt.
  const endPoint = term.costTo || term.riskAt;
  
  const sellerPct = ((endPoint - 1) / segments) * 100;
  const buyerPct  = 100 - sellerPct;

  // If costTo is defined, we have an intermediate dot for risk transfer.
  const riskLeft = term.costTo ? ((term.riskAt - 1) / segments) * 100 : null;

  // Flip tooltip below for first 3 rows so it doesn't hide behind journey strip
  const tooltipBelow = rowIndex < 3;

  return (
    <div
      className={`inc-bar-row ${isHovered ? 'inc-bar-row--on' : ''}`}
      onMouseEnter={onEnter}
      onMouseLeave={onLeave}
    >
      {/* Code */}
      <div className="inc-br-code">{term.code}</div>

      {/* Name */}
      <div className="inc-br-name">{term.name}</div>

      {/* Bar */}
      <div className="inc-br-bar-wrap">
        <div className="inc-br-bar">
          {/* Seller segment (solid blue) */}
          {sellerPct > 0 && (
            <div className="inc-br-seg inc-br-seller" style={{ width: `${sellerPct}%` }}/>
          )}
          {/* Buyer segment (solid green) */}
          {buyerPct > 0 && (
            <div className="inc-br-seg inc-br-buyer" style={{ width: `${buyerPct}%` }}/>
          )}

          {/* Intermediate dot (C-terms) */}
          {term.costTo && (
            <div
              className={`inc-br-dot inc-br-dot--risk ${term.insurance ? 'has-insurance' : ''}`}
              style={{ left: `${riskLeft}%` }}
              title={`Risk transfers at ${JOURNEY[Math.floor(term.riskAt) - 1]?.label}`}
            >
              {term.insurance && <Shield size={10} color="#fff" />}
            </div>
          )}

          {/* Final end dot (marks boundary between seller and buyer) */}
          <div
            className="inc-br-dot inc-br-dot--main"
            style={{ left: `${sellerPct}%` }}
          />
        </div>

        {/* Hover tooltip — flips below for top rows */}
        {isHovered && (
          <div className={`inc-tooltip ${tooltipBelow ? 'inc-tooltip--below' : ''}`}>
            <div className="inc-tt-code">{term.code}</div>
            <div className="inc-tt-grid">
              <div className="inc-tt-col">
                <div className="inc-tt-label seller">Seller</div>
                <div className="inc-tt-text">{term.seller}</div>
              </div>
              <div className="inc-tt-vline"/>
              <div className="inc-tt-col">
                <div className="inc-tt-label buyer">Buyer</div>
                <div className="inc-tt-text">{term.buyer}</div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Mode pill */}
      <div className={`inc-br-pill ${term.type === 'sea' ? 'pill-sea' : 'pill-any'}`}>
        {term.type === 'sea' ? '🌊 Sea Only' : '🚚 Any Mode'}
      </div>
    </div>
  );
}
