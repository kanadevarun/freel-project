import { useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';
import {
  Globe, Play, BookOpen, FileText, Users,
  Package, Truck, Anchor, ShieldCheck, Shield, Warehouse,
  ArrowRight, CheckCircle, CheckCircle2, Clock, Calendar, User,
  AlertTriangle, Lightbulb, GraduationCap, Plane, MapPin
} from 'lucide-react';
import { motion, useAnimation } from 'framer-motion';
import './Page.css';

/* ─────────────────────────────────────────
   STATS
───────────────────────────────────────── */
const STATS = [
  { icon: <BookOpen size={18} />, val: '15+',   lbl: 'Topics Covered',     color: '#2563eb' },
  { icon: <FileText size={18} />, val: '50+',   lbl: 'Guides & Resources', color: '#0891b2' },
  { icon: <Globe size={18} />,    val: '200+',  lbl: 'Countries Traded',   color: '#0284c7' },
  { icon: <Users size={18} />,    val: '100K+', lbl: 'Monthly Learners',   color: '#4f46e5' },
];

/* ─────────────────────────────────────────
   FLOATING DOCUMENT CARDS
───────────────────────────────────────── */
const FLOAT_CARDS = [
  { emoji: '📄', title: 'Commercial Invoice',  sub: 'Document verified',  color: '#2563eb', delay: 0    },
  { emoji: '🚢', title: 'Bill of Lading',      sub: 'B/L Issued',         color: '#0891b2', delay: 0.3  },
  { emoji: '🛃', title: 'Customs Cleared',     sub: 'Port of Rotterdam',   color: '#059669', delay: 0.6  },
  { emoji: '📦', title: 'Packing List',        sub: '42 cartons · 3.2 T',  color: '#7c3aed', delay: 0.9  },
  { emoji: '📍', title: 'Shipment Tracking',   sub: 'ETA in 4 days',      color: '#0284c7', delay: 1.2  },
  { emoji: '🏷', title: 'HS Code: 8471.30',    sub: 'Classification done', color: '#4f46e5', delay: 1.5  },
];

/* ─────────────────────────────────────────
   JOURNEY NODES  (the 13-step trade flow)
───────────────────────────────────────── */
const JOURNEY_NODES = [
  { id: 'factory',   emoji: '🏭', label: 'Supplier',       x: 8,   y: 38 },
  { id: 'pack',      emoji: '📦', label: 'Packaging',      x: 20,  y: 26 },
  { id: 'load',      emoji: '🏗', label: 'Container',      x: 32,  y: 40 },
  { id: 'truck1',    emoji: '🚚', label: 'Export Truck',   x: 44,  y: 28 },
  { id: 'portA',     emoji: '⚓', label: 'Origin Port',    x: 56,  y: 45 },
  { id: 'ship',      emoji: '🚢', label: 'Ocean Freight',  x: 68,  y: 62 },
  { id: 'portB',     emoji: '⚓', label: 'Dest. Port',     x: 80,  y: 46 },
  { id: 'customs',   emoji: '🛃', label: 'Customs',        x: 88,  y: 34 },
  { id: 'warehouse', emoji: '🏬', label: 'Warehouse',      x: 78,  y: 22 },
  { id: 'truck2',    emoji: '🚛', label: 'Local Delivery', x: 66,  y: 18 },
  { id: 'buyer',     emoji: '🏢', label: 'Buyer',          x: 54,  y: 12 },
];

/* Route path pairs (index into JOURNEY_NODES) */
const ROUTES = [
  [0,1],[1,2],[2,3],[3,4],[4,5],[5,6],[6,7],[7,8],[8,9],[9,10],
];

/* ─────────────────────────────────────────
   LEARNING ROADMAP
───────────────────────────────────────── */
const ROADMAP = [
  { step: '01', icon: <Globe size={20}/>,       title: 'Understanding Global Trade',    desc: 'What trade is, why it matters, and how countries interact.' },
  { step: '02', icon: <Users size={20}/>,       title: 'Who Are the Participants?',     desc: 'Exporters, importers, freight forwarders, customs brokers & banks.' },
  { step: '03', icon: <FileText size={20}/>,    title: 'Trade Documents',               desc: 'Commercial invoice, packing list, bill of lading, certificate of origin.' },
  { step: '04', icon: <Truck size={20}/>,       title: 'Transportation Modes',          desc: 'Air, sea, road, rail and multimodal logistics compared.' },
  { step: '05', icon: <ShieldCheck size={20}/>, title: 'Customs & Compliance',          desc: 'Import duties, HS codes, tariffs and country-specific regulations.' },
  { step: '06', icon: <Package size={20}/>,     title: 'Packaging & Labelling',         desc: 'ISPM-15, hazmat rules, weight limits and proper marking.' },
  { step: '07', icon: <Anchor size={20}/>,      title: 'Incoterms 2020',               desc: 'EXW, FOB, CIF, DDP — who pays, who ships, who bears risk.' },
  { step: '08', icon: <Warehouse size={20}/>,   title: 'Payment & Finance',             desc: 'Letters of credit, open account, documentary collections, trade finance.' },
  { step: '09', icon: <CheckCircle2 size={20}/>,title: 'Real Shipment Walkthrough',     desc: 'Follow a container from Shanghai to Rotterdam step by step.' },
];

/* ─────────────────────────────────────────
   WORLD MAP DOTS  (simplified lat/lon grid)
───────────────────────────────────────── */
function WorldMapDots() {
  /* Generate a simple dot grid approximating a world map outline */
  const dots = [];
  /* We draw a 60×30 grid and skip "ocean-only" cells */
  const LAND = new Set([
    /* North America rough band */
    ...Array.from({length:12},(_,i)=>`${10+i},7`),
    ...Array.from({length:14},(_,i)=>`${9+i},8`),
    ...Array.from({length:15},(_,i)=>`${8+i},9`),
    ...Array.from({length:15},(_,i)=>`${8+i},10`),
    ...Array.from({length:14},(_,i)=>`${9+i},11`),
    ...Array.from({length:12},(_,i)=>`${10+i},12`),
    ...Array.from({length:8},(_,i)=>`${12+i},13`),
    /* South America */
    ...Array.from({length:8},(_,i)=>`${14+i},14`),
    ...Array.from({length:9},(_,i)=>`${13+i},15`),
    ...Array.from({length:9},(_,i)=>`${13+i},16`),
    ...Array.from({length:8},(_,i)=>`${14+i},17`),
    ...Array.from({length:6},(_,i)=>`${15+i},18`),
    /* Europe */
    ...Array.from({length:10},(_,i)=>`${27+i},6`),
    ...Array.from({length:12},(_,i)=>`${26+i},7`),
    ...Array.from({length:12},(_,i)=>`${26+i},8`),
    ...Array.from({length:10},(_,i)=>`${27+i},9`),
    /* Africa */
    ...Array.from({length:10},(_,i)=>`${28+i},10`),
    ...Array.from({length:12},(_,i)=>`${27+i},11`),
    ...Array.from({length:12},(_,i)=>`${27+i},12`),
    ...Array.from({length:12},(_,i)=>`${27+i},13`),
    ...Array.from({length:10},(_,i)=>`${28+i},14`),
    ...Array.from({length:8},(_,i)=>`${29+i},15`),
    /* Asia */
    ...Array.from({length:18},(_,i)=>`${36+i},5`),
    ...Array.from({length:20},(_,i)=>`${35+i},6`),
    ...Array.from({length:22},(_,i)=>`${34+i},7`),
    ...Array.from({length:22},(_,i)=>`${34+i},8`),
    ...Array.from({length:20},(_,i)=>`${35+i},9`),
    ...Array.from({length:18},(_,i)=>`${36+i},10`),
    /* Australia */
    ...Array.from({length:8},(_,i)=>`${44+i},15`),
    ...Array.from({length:9},(_,i)=>`${43+i},16`),
    ...Array.from({length:8},(_,i)=>`${44+i},17`),
  ]);

  for (let col = 0; col < 58; col++) {
    for (let row = 0; row < 25; row++) {
      if (LAND.has(`${col},${row}`)) {
        dots.push(
          <circle
            key={`${col}-${row}`}
            cx={col * 14 + 7}
            cy={row * 14 + 7}
            r={1.8}
            fill="#bfdbfe"
            opacity={0.55}
          />
        );
      }
    }
  }
  return <>{dots}</>;
}

/* ─────────────────────────────────────────
   TRADE JOURNEY ILLUSTRATION (SVG)
───────────────────────────────────────── */
function TradeJourneyIllustration() {
  /* Convert percentage coords to SVG viewBox (800×520) */
  const W = 800, H = 480;
  const toX = pct => (pct / 100) * W;
  const toY = pct => (pct / 100) * H;

  /* Build smooth cubic bezier path between consecutive nodes */
  function makePath(nodes) {
    if (nodes.length < 2) return '';
    const pts = nodes.map(n => ({ x: toX(n.x), y: toY(n.y) }));
    let d = `M ${pts[0].x} ${pts[0].y}`;
    for (let i = 1; i < pts.length; i++) {
      const prev = pts[i - 1];
      const curr = pts[i];
      const cx1 = (prev.x + curr.x) / 2;
      const cy1 = prev.y;
      const cx2 = (prev.x + curr.x) / 2;
      const cy2 = curr.y;
      d += ` C ${cx1} ${cy1} ${cx2} ${cy2} ${curr.x} ${curr.y}`;
    }
    return d;
  }

  const pathD = makePath(JOURNEY_NODES);

  return (
    <svg
      viewBox={`0 0 ${W} ${H}`}
      className="ieb-journey-svg"
      aria-hidden="true"
    >
      <defs>
        {/* Route gradient */}
        <linearGradient id="routeGrad" x1="0%" y1="0%" x2="100%" y2="0%">
          <stop offset="0%"   stopColor="#00C2FF" stopOpacity="0.6" />
          <stop offset="50%"  stopColor="#2563EB" stopOpacity="0.8" />
          <stop offset="100%" stopColor="#4f46e5" stopOpacity="0.6" />
        </linearGradient>

        {/* Animated dash for the route */}
        <linearGradient id="dotGrad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%"   stopColor="#00C2FF" />
          <stop offset="100%" stopColor="#2563EB" />
        </linearGradient>

        {/* Glow filter for nodes */}
        <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="3" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>

        {/* Soft glow for moving dot */}
        <filter id="movingGlow">
          <feGaussianBlur stdDeviation="4" result="glow" />
          <feMerge>
            <feMergeNode in="glow"/>
            <feMergeNode in="SourceGraphic"/>
          </feMerge>
        </filter>

        {/* World map dots clip */}
        <clipPath id="mapClip">
          <rect x="0" y="0" width={W} height={H} />
        </clipPath>
      </defs>

      {/* ── World Map Dots ── */}
      <g clipPath="url(#mapClip)" opacity="0.6">
        <WorldMapDots />
      </g>

      {/* ── Background glow blob ── */}
      <ellipse cx="400" cy="280" rx="340" ry="180"
        fill="url(#routeGrad)" opacity="0.04" />

      {/* ── Main Route Path (dashed animated) ── */}
      <path
        d={pathD}
        fill="none"
        stroke="url(#routeGrad)"
        strokeWidth="2"
        strokeDasharray="6 5"
        opacity="0.5"
      >
        <animate
          attributeName="stroke-dashoffset"
          from="0" to="-110"
          dur="3s"
          repeatCount="indefinite"
        />
      </path>

      {/* ── Main Route Path (solid, thinner) ── */}
      <path
        d={pathD}
        fill="none"
        stroke="url(#routeGrad)"
        strokeWidth="1.5"
        opacity="0.25"
      />

      {/* ── Moving Cargo Dot ── */}
      <circle r="6" fill="#00C2FF" filter="url(#movingGlow)" opacity="0.9">
        <animateMotion dur="6s" repeatCount="indefinite" path={pathD} />
      </circle>
      <circle r="3" fill="#ffffff" opacity="0.9">
        <animateMotion dur="6s" repeatCount="indefinite" path={pathD} />
      </circle>

      {/* ── Journey Nodes ── */}
      {JOURNEY_NODES.map((node, i) => {
        const cx = toX(node.x);
        const cy = toY(node.y);
        return (
          <g key={node.id} filter="url(#glow)">
            {/* Pulse ring */}
            <circle cx={cx} cy={cy} r="18" fill="#2563eb" opacity="0.06">
              <animate attributeName="r" values="14;22;14" dur={`${2.5 + i * 0.3}s`} repeatCount="indefinite" />
              <animate attributeName="opacity" values="0.08;0.02;0.08" dur={`${2.5 + i * 0.3}s`} repeatCount="indefinite" />
            </circle>
            {/* Node circle */}
            <circle cx={cx} cy={cy} r="13" fill="white" stroke="#bfdbfe" strokeWidth="1.5" />
            {/* Node background fill */}
            <circle cx={cx} cy={cy} r="11" fill="#eff6ff" />
            {/* Emoji text */}
            <text x={cx} y={cy + 5} textAnchor="middle" fontSize="11" className="ieb-node-emoji">
              {node.emoji}
            </text>
            {/* Label */}
            <text
              x={cx}
              y={cy + 26}
              textAnchor="middle"
              fontSize="7.5"
              fontWeight="700"
              fill="#475569"
              letterSpacing="0.3"
            >
              {node.label}
            </text>
          </g>
        );
      })}

      {/* ── Route segment highlight dots ── */}
      {JOURNEY_NODES.map((node, i) => {
        if (i === 0) return null;
        const prev = JOURNEY_NODES[i - 1];
        const mx = toX((node.x + prev.x) / 2);
        const my = toY((node.y + prev.y) / 2);
        return (
          <circle key={`mid-${i}`} cx={mx} cy={my} r="2.5"
            fill="#00C2FF" opacity="0.35" />
        );
      })}
    </svg>
  );
}

/* ─────────────────────────────────────────
   SECTION 2: UNDERSTANDING IMPORT & EXPORT
───────────────────────────────────────── */
const JOURNEY_STEPS = [
  { n: '1', emoji: '🏭', label: 'Factory',          x: 8,  y: 52 },
  { n: '2', emoji: '🚛', label: 'Truck',             x: 24, y: 40 },
  { n: '3', emoji: '🏗', label: 'Port of Origin',   x: 46, y: 20 },
  { n: '4', emoji: '🚢', label: 'Ocean Freight',    x: 62, y: 56 },
  { n: '5', emoji: '⚓', label: 'Destination Port', x: 22, y: 80 },
  { n: '6', emoji: '🏬', label: 'Warehouse',        x: 42, y: 88 },
  { n: '7', emoji: '🚐', label: 'Local Delivery',   x: 62, y: 80 },
  { n: '8', emoji: '🏢', label: 'Customer',         x: 80, y: 88 },
];

const IMPORT_EXAMPLES = [
  { flag: '🇨🇳', text: 'Electronics from China' },
  { flag: '🇦🇪', text: 'Crude Oil from UAE' },
  { flag: '🇩🇪', text: 'Machinery from Germany' },
];

const EXPORT_EXAMPLES = [
  { emoji: '💊', text: 'Pharmaceuticals' },
  { emoji: '🌾', text: 'Rice' },
  { emoji: '🍃', text: 'Tea' },
  { emoji: '👕', text: 'Textiles' },
  { emoji: '</>', text: 'Software Services' },
];

const GLOBAL_STATS = [
  { icon: '🌐', val: '$32T+',   title: 'Global Trade Value',      desc: 'Total value of goods traded globally every year.' },
  { icon: '🤝', val: '200+',    title: 'Countries Connected',      desc: 'Global reach across every major economy.' },
  { icon: '🚢', val: '80%',     title: 'of Cargo Moves by Sea',    desc: 'Ocean freight remains the backbone of global trade.' },
  { icon: '📦', val: 'Millions',title: 'Containers Shipped',       desc: 'Containers move around the world every year.' },
];

function JourneySVG() {
  const W = 700, H = 380;
  const toX = p => (p / 100) * W;
  const toY = p => (p / 100) * H;

  function curvePath(a, b) {
    const ax = toX(a.x), ay = toY(a.y);
    const bx = toX(b.x), by = toY(b.y);
    const mx = (ax + bx) / 2;
    return `M ${ax} ${ay} C ${mx} ${ay} ${mx} ${by} ${bx} ${by}`;
  }

  return (
    <svg viewBox={`0 0 ${W} ${H}`} className="us2-journey-svg" aria-hidden="true">
      <defs>
        <linearGradient id="us2Grad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%"   stopColor="#00C2FF" />
          <stop offset="100%" stopColor="#2563EB" />
        </linearGradient>
        <filter id="us2Glow">
          <feGaussianBlur stdDeviation="3" result="b"/>
          <feMerge><feMergeNode in="b"/><feMergeNode in="SourceGraphic"/></feMerge>
        </filter>
        <filter id="us2MovGlow">
          <feGaussianBlur stdDeviation="5" result="b"/>
          <feMerge><feMergeNode in="b"/><feMergeNode in="SourceGraphic"/></feMerge>
        </filter>
      </defs>

      {/* ── World-map dot grid ── */}
      {Array.from({ length: 24 }, (_, col) =>
        Array.from({ length: 14 }, (_, row) => (
          <circle
            key={`d-${col}-${row}`}
            cx={col * 30 + 15}
            cy={row * 28 + 14}
            r="1.8"
            fill="#bfdbfe"
            opacity="0.4"
          />
        ))
      )}

      {/* ── Curved route paths ── */}
      {JOURNEY_STEPS.map((step, i) => {
        if (i === JOURNEY_STEPS.length - 1) return null;
        const next = JOURNEY_STEPS[i + 1];
        const d = curvePath(step, next);
        return (
          <path key={`path-${i}`} d={d} fill="none"
            stroke="url(#us2Grad)" strokeWidth="2"
            strokeDasharray="7 5" opacity="0.55"
          >
            <animate attributeName="stroke-dashoffset"
              from="0" to="-120" dur={`${2.5 + i * 0.2}s`}
              repeatCount="indefinite" />
          </path>
        );
      })}

      {/* ── Moving cargo dot ── */}
      {JOURNEY_STEPS.length > 1 && (() => {
        const allPaths = JOURNEY_STEPS.slice(0, -1)
          .map((s, i) => curvePath(s, JOURNEY_STEPS[i + 1]))
          .join(' ');
        return (
          <>
            <circle r="7" fill="#00C2FF" filter="url(#us2MovGlow)" opacity="0.85">
              <animateMotion dur="9s" repeatCount="indefinite" path={allPaths} />
            </circle>
            <circle r="3.5" fill="#fff" opacity="0.95">
              <animateMotion dur="9s" repeatCount="indefinite" path={allPaths} />
            </circle>
          </>
        );
      })()}

      {/* ── Journey Nodes ── */}
      {JOURNEY_STEPS.map((step, i) => {
        const cx = toX(step.x), cy = toY(step.y);
        return (
          <g key={step.n} filter="url(#us2Glow)">
            {/* Soft pulse ring */}
            <circle cx={cx} cy={cy} r="22" fill="#2563eb" opacity="0.04">
              <animate attributeName="r" values="16;26;16" dur={`${3 + i * 0.4}s`} repeatCount="indefinite"/>
              <animate attributeName="opacity" values="0.06;0.01;0.06" dur={`${3 + i * 0.4}s`} repeatCount="indefinite"/>
            </circle>
            {/* White card background */}
            <rect x={cx - 26} y={cy - 26} width="52" height="52"
              rx="12" fill="white"
              stroke="#dbeafe" strokeWidth="1.5"
              style={{ filter: 'drop-shadow(0 4px 12px rgba(37,99,235,0.08))' }}
            />
            {/* Emoji */}
            <text x={cx} y={cy + 2} textAnchor="middle" fontSize="18"
              dominantBaseline="middle" className="ieb-node-emoji">
              {step.emoji}
            </text>
            {/* Step number badge */}
            <circle cx={cx + 18} cy={cy - 18} r="9" fill="url(#us2Grad)" />
            <text x={cx + 18} y={cy - 15} textAnchor="middle"
              fontSize="8" fontWeight="800" fill="white">
              {step.n}
            </text>
            {/* Label below */}
            <text x={cx} y={cy + 38} textAnchor="middle"
              fontSize="9" fontWeight="700" fill="#475569" letterSpacing="0.2">
              {step.label}
            </text>
          </g>
        );
      })}
    </svg>
  );
}

function UnderstandingSection() {
  return (
    <section className="us2-section" id="what-is-import-export">
      <div className="us2-inner">

        {/* ────────── LEFT: heading + SVG map ────────── */}
        <div className="us2-left">
          <motion.div
            className="us2-module-badge"
            initial={{ opacity: 0, x: -16 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.4 }}
          >
            <span className="us2-mb-num">01</span>
            <span className="us2-mb-text">WHAT IS IMPORT &amp; EXPORT?</span>
          </motion.div>

          <motion.h2
            className="us2-heading"
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.08 }}
          >
            Understanding<br />
            <span className="us2-heading-grad">Import &amp; Export</span>
          </motion.h2>

          <motion.p
            className="us2-desc"
            initial={{ opacity: 0, y: 16 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.16 }}
          >
            Everything that moves across international borders follows an import-export
            process. Global trade connects manufacturers, businesses, and consumers worldwide.
          </motion.p>

          {/* Journey Map Image */}
          <motion.div
            className="us2-img-wrap"
            initial={{ opacity: 0, y: 24 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.7, delay: 0.22 }}
          >
            <img
              src="/images/import-export/journey-map.png"
              alt="Global trade journey: Factory → Truck → Port → Ocean Freight → Destination Port → Warehouse → Local Delivery → Customer"
              className="us2-journey-img"
            />
          </motion.div>
        </div>

        {/* ────────── RIGHT: cards + insight banner ────────── */}
        <div className="us2-right">

          {/* Import & Export definition cards */}
          <div className="us2-def-cards">

            {/* IMPORT card */}
            <motion.div
              className="us2-def-card us2-def-card--import"
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.45, delay: 0.1 }}
            >
              <div className="us2-dc-header">
                <span className="us2-dc-icon-badge us2-dc-icon-badge--import">📥</span>
                <h3 className="us2-dc-title us2-dc-title--import">Import</h3>
              </div>
              <p className="us2-dc-def">
                Import means bringing goods into your country from another country.
              </p>
              <div className="us2-dc-examples-label us2-dc-examples-label--import">Examples</div>
              <ul className="us2-dc-list">
                {IMPORT_EXAMPLES.map((ex, i) => (
                  <li key={i} className="us2-dc-item">
                    <span className="us2-dc-flag">{ex.flag}</span>
                    <span>{ex.text}</span>
                  </li>
                ))}
              </ul>
            </motion.div>

            {/* EXPORT card */}
            <motion.div
              className="us2-def-card us2-def-card--export"
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.45, delay: 0.2 }}
            >
              <div className="us2-dc-header">
                <span className="us2-dc-icon-badge us2-dc-icon-badge--export">📤</span>
                <h3 className="us2-dc-title us2-dc-title--export">Export</h3>
              </div>
              <p className="us2-dc-def">
                Export means selling goods from your country to customers in another country.
              </p>
              <div className="us2-dc-examples-label us2-dc-examples-label--export">Examples</div>
              <ul className="us2-dc-list">
                {EXPORT_EXAMPLES.map((ex, i) => (
                  <li key={i} className="us2-dc-item">
                    <span className="us2-dc-emoji">{ex.emoji}</span>
                    <span>{ex.text}</span>
                  </li>
                ))}
              </ul>
            </motion.div>
          </div>

          {/* Insight Banner */}
          <motion.div
            className="us2-insight-banner"
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            <div className="us2-ib-icon">
              <Globe size={28} />
            </div>
            <div className="us2-ib-body">
              <span className="us2-ib-eyebrow">Did you know?</span>
              <p className="us2-ib-text">
                Every international shipment is an{' '}
                <strong className="us2-ib-export">Export</strong> for one country
                and an <strong className="us2-ib-import">Import</strong> for another.
              </p>
            </div>
          </motion.div>
        </div>

      </div>
    </section>
  );
}

/* ─────────────────────────────────────────
   SECTION 3: THE GLOBAL TRADE JOURNEY
───────────────────────────────────────── */
const TRADE_STEPS = [
  {
    n: '1', name: 'Factory',         img: '/images/import-export/modules/factory.png',
    desc: 'Manufacturer prepares the products.',
    time: '1–3 Days',
  },
  {
    n: '2', name: 'Packaging',       img: '/images/import-export/modules/packaging.png',
    desc: 'Products are packed, labeled, and prepared for shipping.',
    time: '1 Day',
  },
  {
    n: '3', name: 'Truck Pickup',    img: '/images/import-export/modules/truck-pickup.png',
    desc: 'Local truck collects the cargo from the factory or warehouse.',
    time: '1 Day',
  },
  {
    n: '4', name: 'Export Customs',  img: '/images/import-export/modules/export-customs.png',
    desc: 'Documents are verified and cargo receives export clearance.',
    time: '1–2 Days',
  },
  {
    n: '5', name: 'Port Terminal',   img: '/images/import-export/modules/port-terminal.png',
    desc: 'Containers enter the terminal and are loaded on to the vessel.',
    time: '1–2 Days',
  },
  {
    n: '6', name: 'Ocean / Air Freight', img: '/images/import-export/modules/ocean-freight.png',
    desc: 'The longest part of the journey. Your cargo moves across oceans.',
    time: '10–20 Days',
  },
  {
    n: '7', name: 'Import Customs',  img: '/images/import-export/modules/import-customs.png',
    desc: 'Destination country verifies shipment, assesses duties, and clears cargo.',
    time: '1–3 Days',
  },
  {
    n: '8', name: 'Warehouse',       img: '/images/import-export/modules/warehouse.png',
    desc: 'Cargo is unloaded, inspected, and stored in the warehouse.',
    time: '1–2 Days',
  },
  {
    n: '9', name: 'Final Delivery',  img: '/images/import-export/modules/final-delivery.png',
    desc: 'Shipment is delivered to the customer. POD completed.',
    time: '1 Day',
  },
];

const TIMELINE_MILESTONES = [
  { day: 'DAY 1',  label: 'Factory',          emoji: '🏭' },
  { day: 'DAY 2',  label: 'Pickup',            emoji: '🚛' },
  { day: 'DAY 4',  label: 'Export Port',       emoji: '⚓' },
  { day: 'DAY 6',  label: 'Vessel Departure',  emoji: '🚢' },
  { day: 'DAY 25', label: 'Destination Port',  emoji: '⚓' },
  { day: 'DAY 27', label: 'Warehouse',         emoji: '🏬' },
  { day: 'DAY 28', label: 'Customer Delivery', emoji: '🏢' },
];

function GlobalTradeJourney() {
  return (
    <section className="gtj-section" id="global-trade-journey">
      <div className="gtj-inner">

        {/* ── Top: header + hero image + status card ── */}
        <div className="gtj-top">

          {/* Left: badge + title + description */}
          <div className="gtj-header">
            <motion.div
              className="gtj-badge"
              initial={{ opacity: 0, x: -16 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4 }}
            >
              <span className="gtj-badge-num">02</span>
              <Globe size={13} />
              <span>GLOBAL TRADE JOURNEY</span>
            </motion.div>

            <motion.h2
              className="gtj-title"
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.06 }}
            >
              The Global Trade Journey
            </motion.h2>
            <motion.p
              className="gtj-subtitle"
              initial={{ opacity: 0, y: 14 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.12 }}
            >
              Follow one shipment across the world in <strong>9 easy steps.</strong>
            </motion.p>
            <motion.p
              className="gtj-desc"
              initial={{ opacity: 0, y: 14 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.18 }}
            >
              Every international shipment follows a predictable journey. Whether you're shipping
              electronics from China, coffee from Brazil, or machinery from Germany,
              the process is remarkably similar.
            </motion.p>
          </div>

          {/* Centre: hero world map + ship image — with status card overlaid */}
          <motion.div
            className="gtj-hero-img-wrap"
            initial={{ opacity: 0, scale: 0.97 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.7, delay: 0.1 }}
          >
            <div className="gtj-hero-img-inner">
              <img
                src="/images/import-export/modules/hero-map.png"
                alt="World map with container ship and cargo plane showing global trade routes"
                className="gtj-hero-img"
              />
            </div>

            {/* Status card overlaid on top-right of the image */}
            <motion.div
              className="gtj-status-card"
              initial={{ opacity: 0, x: 20 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.3 }}
            >
              <div className="gtj-sc-header">
                <span className="gtj-sc-title">Shipment Overview</span>
                <span className="gtj-sc-live"><span className="gtj-sc-dot" />LIVE</span>
              </div>
              <div className="gtj-sc-row">
                <span className="gtj-sc-icon">📦</span>
                <div>
                  <div className="gtj-sc-label">Container Status</div>
                  <div className="gtj-sc-val gtj-sc-val--transit">IN TRANSIT</div>
                </div>
              </div>
              <div className="gtj-sc-row">
                <span className="gtj-sc-icon">📅</span>
                <div>
                  <div className="gtj-sc-label">Estimated Arrival</div>
                  <div className="gtj-sc-val gtj-sc-val--blue">14 DAYS</div>
                  <div className="gtj-sc-sub">Jul 13, 2025</div>
                </div>
              </div>
              <div className="gtj-sc-row">
                <span className="gtj-sc-icon">📍</span>
                <div>
                  <div className="gtj-sc-label">Tracking</div>
                  <div className="gtj-sc-val gtj-sc-val--green">LIVE</div>
                  <div className="gtj-sc-sub">Real-time updates</div>
                </div>
              </div>
              <div className="gtj-sc-divider" />
              <div className="gtj-sc-row">
                <span className="gtj-sc-icon">🏷</span>
                <div>
                  <div className="gtj-sc-label">Container No.</div>
                  <div className="gtj-sc-val gtj-sc-val--mono">MSKU 483920</div>
                </div>
              </div>
              <div className="gtj-sc-row">
                <span className="gtj-sc-icon">🛡</span>
                <div>
                  <div className="gtj-sc-label">Customs Status</div>
                  <div className="gtj-sc-val gtj-sc-val--green">CLEARED ✓</div>
                </div>
              </div>
            </motion.div>
          </motion.div>

        </div>

        {/* ── 9-step cards grid ── */}
        <div className="gtj-steps-wrap">
          {/* Connector line behind the steps */}
          <div className="gtj-connector-track">
            <div className="gtj-connector-line" />
          </div>

          <div className="gtj-steps-grid">
            {TRADE_STEPS.map((step, i) => (
              <motion.div
                key={step.n}
                className="gtj-step-card"
                initial={{ opacity: 0, y: 28 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: i * 0.06 }}
              >
                {/* Number badge */}
                <div className="gtj-step-num">{step.n}</div>

                {/* Image */}
                <div className="gtj-step-img-wrap">
                  <img
                    src={step.img}
                    alt={step.name}
                    className="gtj-step-img"
                  />
                </div>

                {/* Content */}
                <h3 className="gtj-step-name">{step.name}</h3>
                <p className="gtj-step-desc">{step.desc}</p>
                <div className="gtj-step-time">
                  <Clock size={11} />
                  {step.time}
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* ── Typical Shipment Timeline strip ── */}
        <motion.div
          className="gtj-timeline-strip"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.1 }}
        >
          <div className="gtj-tl-label">
            <span>Typical</span>
            <span>Shipment</span>
            <span>Timeline</span>
          </div>
          <div className="gtj-tl-track">
            {TIMELINE_MILESTONES.map((m, i) => (
              <div key={i} className="gtj-tl-item">
                {i > 0 && (
                  <div className="gtj-tl-dash">
                    {i === 3 && <div className="gtj-tl-dash-long" />}
                  </div>
                )}
                <div className="gtj-tl-milestone">
                  <span className="gtj-tl-emoji">{m.emoji}</span>
                  <div className="gtj-tl-day">{m.day}</div>
                  <div className="gtj-tl-name">{m.label}</div>
                </div>
              </div>
            ))}
          </div>
        </motion.div>

        {/* ── CTA Banner ── */}
        <motion.div
          className="gtj-cta-banner"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.1 }}
        >
          <div className="gtj-cta-left">
            <span className="gtj-cta-icon"><Users size={24} /></span>
            <div>
              <strong>Ready to understand the people involved?</strong>
              <span>Learn about the key participants who make global trade possible.</span>
            </div>
          </div>
          <button className="gtj-cta-btn">
            Meet the Key Participants →
          </button>
        </motion.div>

      </div>
    </section>
  );
}

/* ─────────────────────────────────────────
   SECTION 4: MEET THE PARTICIPANTS
───────────────────────────────────────── */
const PARTICIPANTS = [
  {
    n: '01', name: 'Manufacturer', img: '/images/import-export/participants/manufacturer.png',
    role: 'Produces the goods.',
    accent: '#2563eb',
    examples: { label: 'Examples:', list: ['Apple Factory', 'Tata Steel', 'Nike Manufacturing'] },
  },
  {
    n: '02', name: 'Exporter', img: '/images/import-export/participants/exporter.png',
    role: 'Sells goods internationally.',
    accent: '#0891b2',
    examples: { label: 'Responsible for:', list: ['Commercial Invoice', 'Packing List', 'Export Documentation'] },
  },
  {
    n: '03', name: 'Freight Forwarder', img: '/images/import-export/participants/freight-forwarder.png',
    role: 'The logistics coordinator.',
    accent: '#7c3aed',
    examples: { label: 'Books:', list: ['Ships', 'Flights', 'Trucks'] },
    extra: 'Handles documentation and customs coordination.',
  },
  {
    n: '04', name: 'Customs', img: '/images/import-export/participants/customs.png',
    role: 'Government authority.',
    accent: '#dc2626',
    examples: { label: 'Responsible for:', list: ['Checking documentation', 'Import duties', 'Export clearance', 'Regulatory compliance'] },
  },
  {
    n: '05', name: 'Importer', img: '/images/import-export/participants/importer.png',
    role: 'Purchases goods.',
    accent: '#059669',
    examples: { label: 'Responsible for:', list: ['Receives cargo', 'Pays duties', 'Coordinates local delivery'] },
  },
  {
    n: '06', name: 'Local Transport', img: '/images/import-export/participants/local-transport.png',
    role: 'Moves cargo from port to final destination.',
    accent: '#d97706',
    examples: { label: 'Moves cargo from:', list: ['Port', 'Warehouse', 'Retailer', 'Customer'] },
  },
];

const ECOSYSTEM = [
  { icon: <Calendar size={16} strokeWidth={2.5} />, name: 'Manufacturer', sub: 'Produces the goods', img: '/images/import-export/participants/manufacturer.png' },
  { icon: <Package size={16} strokeWidth={2.5} />, name: 'Exporter', sub: 'Sells goods internationally', img: '/images/import-export/participants/exporter.png' },
  { icon: <Globe size={16} strokeWidth={2.5} />, name: 'Freight Forwarder', sub: 'The logistics coordinator', img: '/images/import-export/participants/freight-forwarder.png' },
  { icon: <Anchor size={16} strokeWidth={2.5} />, name: 'Shipping Line / Airline', sub: 'Moves cargo across the world', img: '/images/import-export/participants/shipping-airline.png' },
  { icon: <ShieldCheck size={16} strokeWidth={2.5} />, name: 'Customs', sub: 'Government authority', img: '/images/import-export/participants/customs.png' },
  { icon: <Warehouse size={16} strokeWidth={2.5} />, name: 'Importer', sub: 'Purchases goods and receives cargo', img: '/images/import-export/participants/importer.png' },
  { icon: <Truck size={16} strokeWidth={2.5} />, name: 'Local Transport', sub: 'Moves cargo from port to final destination', img: '/images/import-export/participants/local-transport.png' },
  { icon: <User size={16} strokeWidth={2.5} />, name: 'Retailer / Customer', sub: 'Final delivery to the customer', img: '/images/import-export/participants/retailer-customer.png' },
];

const MTP_STATS = [
  { icon: '👥', val: '8+',      title: 'Key Participants',              desc: 'In every international shipment' },
  { icon: '🌐', val: '195+',    title: 'Countries Trading Daily',       desc: 'Connected through global trade' },
  { icon: '🚢', val: '90%',     title: 'Trade Coordinated by Freight Professionals', desc: 'Ensuring smooth movement worldwide' },
  { icon: '📦', val: 'Millions',title: 'Shipments Managed Every Year',  desc: 'Across industries and continents' },
];

function MeetTheParticipants() {
  return (
    <section className="mtp-section" id="meet-the-participants">
      <div className="mtp-inner">

        {/* ══ TOP: header + 6-card grid ══ */}
        <div className="mtp-top">

          {/* Left: title + ecosystem flow */}
          <div className="mtp-left">
            <motion.div className="mtp-badge"
              initial={{ opacity: 0, x: -16 }} whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }} transition={{ duration: 0.4 }}>
              <span className="mtp-badge-num">03</span>
              <Users size={13} />
              <span>MEET THE PARTICIPANTS</span>
            </motion.div>

            <motion.h2 className="mtp-title"
              initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }} transition={{ duration: 0.5, delay: 0.06 }}>
              Meet the People<br />Behind <span className="mtp-title-grad">Every Shipment</span>
            </motion.h2>

            <motion.p className="mtp-desc"
              initial={{ opacity: 0, y: 14 }} whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }} transition={{ duration: 0.5, delay: 0.12 }}>
              Every international shipment involves multiple businesses working together.
              From manufacturers to customs officials, each participant has a specific
              responsibility that keeps global trade moving.
            </motion.p>

            {/* Ecosystem flow panel */}
            <motion.div className="mtp-ecosystem"
              initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }} transition={{ duration: 0.5, delay: 0.18 }}>
              <div className="mtp-eco-label">THE GLOBAL TRADE ECOSYSTEM</div>
              <div className="mtp-eco-flow">
                <svg className="mtp-eco-dashline" viewBox="0 0 100 800" preserveAspectRatio="none">
                  {/* CSS will draw the connecting dashed path */}
                </svg>
                {ECOSYSTEM.map((e, i) => (
                  <div key={i} className="mtp-eco-item">
                    <div className="mtp-eco-left-part">
                      <div className="mtp-eco-icon">{e.icon}</div>
                      <div className="mtp-eco-text">
                        <strong>{e.name}</strong>
                        <span>{e.sub}</span>
                      </div>
                    </div>
                    <div className="mtp-eco-right-part">
                      <img src={e.img} alt={e.name} className="mtp-eco-3d-img" />
                    </div>
                  </div>
                ))}
              </div>
            </motion.div>
          </div>

          {/* Right Column: 3×2 cards + Insight Banner */}
          <div className="mtp-right">
            <div className="mtp-cards-grid">
              {PARTICIPANTS.map((p, i) => (
                <motion.div key={i} className="mtp-card"
                  style={{ '--p-accent': p.accent }}
                  initial={{ opacity: 0, y: 28 }} whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }} transition={{ duration: 0.4, delay: i * 0.07 }}>

                  <div className="mtp-card-num">{p.n}</div>

                  <div className="mtp-card-img-wrap">
                    <img src={p.img} alt={p.name} className="mtp-card-img" />
                  </div>

                  <div className="mtp-card-body">
                    <h3 className="mtp-card-name">{p.name}</h3>
                    <p className="mtp-card-role">{p.role}</p>
                    <div className="mtp-card-examples-label"
                      style={{ color: p.accent }}>{p.examples.label}</div>
                    <ul className="mtp-card-list">
                      {p.examples.list.map((item, j) => (
                        <li key={j}>
                          <span className="mtp-bullet" style={{ background: p.accent }} />
                          {item}
                        </li>
                      ))}
                    </ul>
                    {p.extra && <p className="mtp-card-extra">{p.extra}</p>}
                  </div>
                </motion.div>
              ))}
            </div>

            {/* ══ Freight Forwarder insight banner ══ */}
            <motion.div className="mtp-insight"
              initial={{ opacity: 0, y: 24 }} whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }} transition={{ duration: 0.55, delay: 0.1 }}>
              <div className="mtp-insight-left">
                <div className="mtp-insight-icon">✈️</div>
                <div className="mtp-insight-text">
                  <p className="mtp-insight-headline">
                    Think of a Freight Forwarder as the{' '}
                    <strong className="mtp-insight-highlight">"Travel Agent" for Cargo.</strong>
                  </p>
                  <p className="mtp-insight-sub">
                    Just like a travel agent books flights, hotels, and visas for people,
                    a freight forwarder arranges ships, trucks, customs documentation,
                    warehousing, and delivery for goods.
                  </p>
                </div>
              </div>
              <div className="mtp-insight-checklist">
                {['Best Routes', 'Best Rates', 'Documentation', 'Customs Support', 'End-to-End Coordination'].map((item, i) => (
                  <div key={i} className="mtp-insight-check">
                    <CheckCircle2 size={15} />
                    {item}
                  </div>
                ))}
              </div>
            </motion.div>
          </div>
        </div>

        {/* ══ Stats strip ══ */}
        <div className="mtp-stats-strip">
          {MTP_STATS.map((s, i) => (
            <motion.div key={i} className="mtp-stat"
              initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }} transition={{ duration: 0.4, delay: i * 0.08 }}>
              <div className="mtp-stat-icon">{s.icon}</div>
              <div>
                <div className="mtp-stat-val">{s.val}</div>
                <div className="mtp-stat-title">{s.title}</div>
                <div className="mtp-stat-desc">{s.desc}</div>
              </div>
            </motion.div>
          ))}
        </div>

        {/* ══ CTA Banner ══ */}
        <motion.div className="mtp-cta"
          initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }} transition={{ duration: 0.5, delay: 0.1 }}>
          <div className="mtp-cta-left">
            <span className="mtp-cta-icon"><Truck size={24} /></span>
            <div>
              <strong>Ready to learn how products actually move?</strong>
              <span>Explore transportation modes and how different cargo travels across the world.</span>
            </div>
          </div>
          <button className="mtp-cta-btn">
            Continue to Transportation Modes →
          </button>
        </motion.div>

      </div>
    </section>
  );
}

/* ─────────────────────────────────────────
   GLOBAL STATS STRIP
───────────────────────────────────────── */
function GlobalStatsStrip() {
  return (
    <section className="gss-section">
      <div className="gss-inner">
        {GLOBAL_STATS.map((s, i) => (
          <motion.div
            key={i}
            className="gss-stat"
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.4, delay: i * 0.1 }}
          >
            <div className="gss-stat-icon">{s.icon}</div>
            <div className="gss-stat-body">
              <div className="gss-stat-val">{s.val}</div>
              <div className="gss-stat-title">{s.title}</div>
              <div className="gss-stat-desc">{s.desc}</div>
            </div>
          </motion.div>
        ))}
      </div>
    </section>
  );
}


/* ─────────────────────────────────────────
   SECTION 4: ESSENTIAL TRADE DOCUMENTS
───────────────────────────────────────── */
const DOC_FLOW = [
  { n: '1', title: 'Manufacturer', desc: 'Produces the goods', img: '/images/import-export/documents_new/participant_manufacturer.png' },
  { n: '2', title: 'Commercial Invoice', desc: 'The official sales invoice\nfrom exporter to importer', img: '/images/import-export/documents_new/doc_invoice.png' },
  { n: '3', title: 'Packing List', desc: "Details of what's inside\nthe shipment", img: '/images/import-export/documents_new/doc_packing-list.png' },
  { n: '4', title: 'Forwarder', desc: 'Documents are reviewed\nand shipment is arranged', img: '/images/import-export/documents_new/participant_forwarder.png' },
  { n: '5', title: 'Bill of Lading', desc: 'Issued by shipping line\nas proof of shipment', img: '/images/import-export/documents_new/transport_ship.png' },
  { n: '6', title: 'Customs', desc: 'Documents submitted\nfor customs clearance', img: '/images/import-export/documents_new/participant_customs.png' },
  { n: '7', title: 'Importer', desc: 'Goods released and\ndelivered to importer', img: '/images/import-export/documents_new/participant_importer.png' }
];

const DOC_CARDS = [
  {
    n: '01', title: 'Commercial Invoice', img: '/images/import-export/documents_new/doc_invoice.png',
    subtitle: 'The official sales bill.',
    accent: '#2563eb',
    details: {
      label: 'Contains:',
      list: ['Seller', 'Buyer', 'Product Description', 'Quantity', 'Unit Price', 'Total Value']
    },
    footer: { label: 'Prepared By', val: 'Exporter' }
  },
  {
    n: '02', title: 'Packing List', img: '/images/import-export/documents_new/doc_packing-list.png',
    subtitle: "What's inside the shipment?",
    accent: '#2563eb',
    details: {
      label: 'Contains:',
      list: ['Number of Packages', 'Weight', 'Dimensions', 'Package Markings', 'Item Description']
    },
    footer: { label: 'Prepared By', val: 'Manufacturer /\nExporter' }
  },
  {
    n: '03', title: 'Bill of Lading (B/L)', img: '/images/import-export/documents_new/doc_bill-of-lading.png',
    subtitle: 'Ownership of cargo.',
    accent: '#2563eb',
    details: {
      label: 'Issued By', val: 'Shipping Line', isIssuedTop: true,
      label2: 'Acts as:',
      list: ['Cargo Receipt', 'Transport Contract', 'Ownership Document']
    }
  },
  {
    n: '04', title: 'Certificate of Origin', img: '/images/import-export/documents_new/doc_certificate-of-origin.png',
    subtitle: 'Which country produced\nthe goods?',
    accent: '#2563eb',
    details: {
      label: 'Issued By', val: 'Authorized Chamber\nof Commerce', isIssuedTop: true,
      label2: 'Used For:',
      list: ['Import Duties', 'Trade Agreements', 'Preferential Benefits']
    }
  },
  {
    n: '05', title: 'Insurance Certificate', img: '/images/import-export/documents_new/doc_insurance.png',
    subtitle: 'Protects cargo during transit.',
    accent: '#2563eb',
    details: {
      label: 'Covers:',
      list: ['Damage', 'Theft', 'Accidents', 'Natural Disasters']
    },
    footer: { label: 'Issued By', val: 'Insurance Company', icon: 'shield' }
  },
  {
    n: '06', title: 'Customs Declaration', img: '/images/import-export/documents_new/doc_customs-declaration.png',
    subtitle: 'Official customs paperwork.',
    accent: '#2563eb',
    details: {
      label: 'Contains:',
      list: ['HS Code', 'Duties & Taxes', 'Cargo Details', 'Country of Origin', 'Value Details']
    },
    footer: { label: 'Submitted To', val: 'Customs Authority', icon: 'check' }
  }
];

const DOC_STATS = [
  { icon: <FileText size={28} color="#2563eb" />, val: '25+', title: 'Common Trade\nDocuments', desc: 'Used in global trade' },
  { icon: <Globe size={28} color="#2563eb" />, val: '195+', title: 'Countries Accept\nStandard Documentation', desc: 'For smooth trade' },
  { icon: <Anchor size={28} color="#2563eb" />, val: '99%', title: 'International Shipments\nRequire a Bill of Lading', desc: 'For legal shipment' },
  { icon: <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#2563eb" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>, val: 'Billions', title: 'Documents Processed\nEvery Year', desc: 'Across the globe' },
];

function EssentialTradeDocuments() {
  return (
    <section className="ed-section" id="essential-documents">
      <div className="ed-inner">
        
        {/* TOP ROW: Header + Hero */}
        <div className="ed-top-row">
          {/* HEADER */}
          <div className="ed-header">
            <motion.div className="ed-badge"
              initial={{ opacity: 0, x: -16 }} whileInView={{ opacity: 1, x: 0 }} viewport={{ once: true }}>
              <span className="ed-badge-num">04</span>
              <span className="ed-badge-text">ESSENTIAL TRADE DOCUMENTS</span>
            </motion.div>

            <motion.h2 className="ed-title"
              initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }}>
              Master the Documents<br />Behind <span className="ed-title-blue">Every Shipment</span>
            </motion.h2>

            <motion.p className="ed-desc"
              initial={{ opacity: 0, y: 14 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }}>
              Every international shipment travels with a complete set<br/>
              of trade documents. Learn what each document does,<br/>
              who prepares it, and why customs authorities require it.
            </motion.p>
          </div>

          {/* HERO COMPOSITE */}
          <motion.div className="ed-hero"
            initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }}>
            <div className="ed-hero-map"></div>
            <div className="ed-hero-items">
              <img src="/images/import-export/documents_new/transport_airplane.png" className="ed-hero-air" alt="Air" />
              <img src="/images/import-export/documents_new/transport_ship.png" className="ed-hero-sea" alt="Sea" />
              <img src="/images/import-export/documents_new/transport_container.png" className="ed-hero-rail" alt="Rail" />
              
              {/* Floating Docs */}
              <img src="/images/import-export/documents_new/doc_invoice.png" className="ed-hero-doc1" alt="Doc" />
              <img src="/images/import-export/documents_new/doc_packing-list.png" className="ed-hero-doc2" alt="Doc" />
              <img src="/images/import-export/documents_new/doc_bill-of-lading.png" className="ed-hero-doc3" alt="Doc" />
              <img src="/images/import-export/documents_new/doc_certificate-of-origin.png" className="ed-hero-doc4" alt="Doc" />
              <img src="/images/import-export/documents_new/doc_customs-declaration.png" className="ed-hero-doc5" alt="Doc" />
            </div>
          </motion.div>
        </div>

        {/* MAIN CONTENT (TWO COLUMNS) */}
        <div className="ed-main-content">
          
          {/* Left: Document Flow */}
          <div className="ed-left">
            <div className="ed-journey-panel">
              <h3 className="ed-journey-title">The Document Flow</h3>
              <div className="ed-journey-flow">
                <div className="ed-journey-line"></div>
                {DOC_FLOW.map((step, i) => (
                  <div key={i} className="ed-journey-step">
                    <div className="ed-journey-num">{step.n}</div>
                    <img src={step.img} alt={step.title} className="ed-journey-img" />
                    <div className="ed-journey-text">
                      <strong>{step.title}</strong>
                      <span style={{ whiteSpace: 'pre-line' }}>{step.desc}</span>
                    </div>
                  </div>
                ))}
              </div>
              <div className="ed-flow-important">
                <div className="ed-fi-icon"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#2563eb" strokeWidth="2"><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/></svg></div>
                <div>
                  <strong>Important</strong>
                  <p>Missing even one document can<br/>cause delays, penalties, or rejection<br/>of your shipment.</p>
                </div>
              </div>
            </div>
          </div>

          {/* Right: 6 Cards Grid */}
          <div className="ed-right">
            <div className="ed-cards-grid">
              {DOC_CARDS.map((card, i) => (
                <div key={i} className="ed-card">
                  <div className="ed-card-num">{card.n}</div>
                  <div className="ed-card-img-wrap">
                    <img src={card.img} alt={card.title} className="ed-card-img" />
                  </div>
                  <div className="ed-card-body">
                    <h4 className="ed-card-title">{card.title}</h4>
                    <p className="ed-card-subtitle" style={{ whiteSpace: 'pre-line' }}>{card.subtitle}</p>
                    
                    {card.details.isIssuedTop && (
                      <div className="ed-card-issued-top">
                        <span className="ed-cit-label">{card.details.label}</span>
                        <div className="ed-cit-val">
                          <Anchor size={14} color="#2563eb" style={{marginRight: '6px'}} />
                          <span style={{ whiteSpace: 'pre-line' }}>{card.details.val}</span>
                        </div>
                      </div>
                    )}
                    
                    <div className="ed-card-details-label">{card.details.label2 || card.details.label}</div>
                    <ul className="ed-card-list">
                      {card.details.list.map((item, j) => (
                        <li key={j}>
                          <span className="ed-bullet"></span>
                          {item}
                        </li>
                      ))}
                    </ul>
                    
                    {card.footer && (
                      <div className="ed-card-footer">
                        <span className="ed-cf-label">{card.footer.label}</span>
                        <div className="ed-cf-val-wrap">
                          {card.footer.icon === 'shield' ? <Shield size={14}/> : card.footer.icon === 'check' ? <CheckCircle size={14}/> : <User size={14}/>}
                          <span className="ed-cf-val" style={{ whiteSpace: 'pre-line' }}>{card.footer.val}</span>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* EXAMPLE STRIP */}
        <div className="ed-example-strip">
          <div className="ed-ex-left">
            <div className="ed-ex-text">
              <h4>Real World Example</h4>
              <p>You are exporting 1,000 smartphones<br/>from India to Germany. You will need<br/>all the essential documents shown<br/>here to ensure smooth customs<br/>clearance and on-time delivery.</p>
            </div>
            <div className="ed-ex-img">
               <img src="/images/import-export/documents_new/transport_container.png" alt="Container"/>
            </div>
          </div>
          <div className="ed-ex-divider"></div>
          <div className="ed-ex-right">
            <div className="ed-ex-text">
              <h4>Documents are the Passport of Cargo.</h4>
              <p>Just like people cannot travel internationally without<br/>passports and visas, cargo cannot cross borders<br/>without the correct trade documents.</p>
            </div>
            <div className="ed-ex-icon">
              <BookOpen size={60} color="#1e3a8a" />
            </div>
          </div>
        </div>

        {/* STATS STRIP */}
        <div className="ed-stats-strip">
          {DOC_STATS.map((s, i) => (
            <div key={i} className="ed-stat">
              <div className="ed-stat-icon-wrap">{s.icon}</div>
              <div className="ed-stat-content">
                <div className="ed-stat-val">{s.val}</div>
                <div className="ed-stat-title" style={{ whiteSpace: 'pre-line' }}>{s.title}</div>
                <div className="ed-stat-desc">{s.desc}</div>
              </div>
            </div>
          ))}
        </div>

        {/* BOTTOM CTA */}
        <div className="ed-bottom-cta">
          <div className="ed-cta-content">
            <div className="ed-cta-icon-wrap"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth="2"><path d="M22 10v6M2 10l10-5 10 5-10 5z"/><path d="M6 12v5c3 3 9 3 12 0v-5"/></svg></div>
            <div>
              <h3>Great! You now understand the paperwork.</h3>
              <p>Next, let's explore the different ways goods move around the world.</p>
            </div>
          </div>
          <button className="ed-cta-btn">Next Module: Transportation Modes →</button>
        </div>

      </div>
    </section>
  );
}

/* ─────────────────────────────────────────
   SECTION 5: TRANSPORTATION MODES
───────────────────────────────────────── */
const TM_STATS = [
  { icon: <img src="/images/import-export/documents_new/transport_ship.png" alt="" style={{width: 32}} />, val: '90%', title: 'of global trade\nmoves by Sea Freight', desc: 'Making it the backbone\nof world commerce.' },
  { icon: <img src="/images/import-export/documents_new/transport_airplane.png" alt="" style={{width: 32}} />, val: '35%', title: 'of shipment value\ntravels by Air Freight', desc: 'Ideal for high-value and\ntime-sensitive goods.' },
  { icon: <img src="/images/import-export/documents_new/transport_truck.png" alt="" style={{width: 32}} />, val: 'Millions', title: 'of trucks complete\nlast-mile deliveries daily', desc: 'Ensuring goods reach\nevery doorstep.' },
  { icon: <Globe size={32} color="#2563eb" />, val: '200+', title: 'countries connected\nthrough multimodal logistics', desc: 'Powering global trade\nevery single day.' }
];

const TM_CARDS = [
  {
    num: '01', title: 'Air Freight', img: '/images/import-export/documents_new/transport_airplane.png',
    desc: 'Fastest international transportation method. Ideal for urgent and high-value shipments.',
    bestFor: ['Electronics', 'Pharmaceuticals', 'Luxury Goods', 'Emergency Cargo'],
    adv: ['Fastest delivery', 'Global connectivity', 'High reliability'],
    lim: ['Highest cost', 'Weight restrictions']
  },
  {
    num: '02', title: 'Sea Freight', img: '/images/import-export/documents_new/transport_ship.png',
    desc: 'Most economical method for moving large volumes internationally.',
    bestFor: ['Heavy Machinery', 'Furniture', 'Bulk Cargo', 'Containers'],
    adv: ['Lowest shipping cost', 'Huge cargo capacity', 'Sustainable'],
    lim: ['Long transit time', 'Weather delays']
  },
  {
    num: '03', title: 'Road Freight', img: '/images/import-export/documents_new/transport_truck.png',
    desc: 'Door-to-door transportation connecting warehouses, ports and customers.',
    bestFor: ['Domestic deliveries', 'Regional distribution', 'Last-mile logistics'],
    adv: ['Flexible routes', 'Fast loading', 'Direct delivery'],
    lim: ['Traffic', 'Distance limitations']
  },
  {
    num: '04', title: 'Rail Freight', img: '/images/import-export/documents_new/transport_container.png',
    desc: 'Efficient transportation across long inland distances.',
    bestFor: ['Bulk cargo', 'Containers', 'Heavy industrial goods'],
    adv: ['Low emissions', 'Reliable schedules', 'Cost efficient'],
    lim: ['Limited network', 'Requires road connection']
  },
  {
    num: '05', title: 'Multimodal Transport',
    img: 'combo',
    desc: 'A shipment uses multiple transportation modes under one logistics plan.',
    adv: ['Better optimization', 'Lower cost', 'Greater flexibility']
  },
  {
    num: '06', title: 'Choosing the Right Mode',
    img: '/images/import-export/documents_new/doc_customs-declaration.png',
    desc: 'Choose transportation based on what matters most.',
    checks: ['Delivery deadline', 'Budget', 'Shipment size', 'Destination', 'Cargo sensitivity', 'Cargo type']
  }
];

function TransportationModes() {
  return (
    <section className="tm-section">
      <div className="tm-inner">
        {/* HERO */}
        <div className="tm-hero-row">
          <div className="tm-hero-left">
            <motion.div className="tm-badge" initial={{opacity:0, y:12}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
              <div className="tm-badge-num">05</div>
              <div className="tm-badge-text">TRANSPORTATION MODES</div>
            </motion.div>
            <motion.h2 initial={{opacity:0, y:16}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.1}}>
              Choose the Right<br/><span>Transportation Mode</span>
            </motion.h2>
            <motion.p initial={{opacity:0, y:16}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.2}}>
              Every international shipment travels through one or more transportation modes depending on speed, cost, cargo type, destination, and urgency.
            </motion.p>
            <motion.p initial={{opacity:0, y:16}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.25}}>
              Understand when to use Air, Sea, Road, Rail, and Multimodal transport in real-world logistics.
            </motion.p>
          </div>
          <div className="tm-hero-right">
            <div className="tm-hero-comp">
              <img src="/images/import-export/documents_new/transport_airplane.png" className="tm-comp-plane" alt="Air" />
              <img src="/images/import-export/documents_new/transport_port.png" className="tm-comp-crane" alt="Crane" />
              <img src="/images/import-export/documents_new/transport_ship.png" className="tm-comp-ship" alt="Ship" />
              <img src="/images/import-export/documents_new/transport_warehouse.png" className="tm-comp-warehouse" alt="Warehouse" />
              <img src="/images/import-export/documents_new/transport_truck.png" className="tm-comp-truck" alt="Truck" />
            </div>
          </div>
        </div>

        {/* MAIN LAYOUT */}
        <div className="tm-main-content">
          
          {/* LEFT: JOURNEY */}
          <div className="tm-journey-panel">
            <h3 className="tm-journey-title">Transportation Journey</h3>
            <div className="tm-journey-flow">
              {[
                {img: 'participant_manufacturer.png', title: 'Factory', desc: 'Goods manufactured.'},
                {img: 'transport_truck.png', title: 'Truck Pickup', desc: 'Cargo collected from factory.'},
                {img: 'transport_warehouse.png', title: 'Warehouse', desc: 'Cargo consolidated.'},
                {img: 'transport_container.png', title: 'Port / Rail Terminal', desc: 'Cargo prepared for international transport.'},
                {img: 'transport_port.png', title: 'Sea / Air Terminal', desc: 'Loaded onto ship or aircraft.'},
                {icon: <Globe size={28} color="#2563eb" strokeWidth={1.5} />, title: 'International Transit', desc: 'Cross-country movement.'},
                {img: 'transport_ship.png', title: 'Destination Port', desc: 'Arrival and unloading.'},
                {img: 'transport_truck.png', title: 'Local Delivery', desc: 'Last mile transport.'},
                {img: 'participant_importer.png', title: 'Customer', desc: 'Shipment delivered.'}
              ].map((step, i) => (
                <div key={i} className="tm-journey-step">
                  <div className="tm-journey-num">{i + 1}</div>
                  {step.img ? (
                    <img src={`/images/import-export/documents_new/${step.img}`} alt={step.title} className="tm-journey-img" />
                  ) : (
                    <div className="tm-journey-img" style={{display:'flex', alignItems:'center', justifyContent:'center'}}>{step.icon}</div>
                  )}
                  <div className="tm-journey-text">
                    <strong>{step.title}</strong>
                    <span>{step.desc}</span>
                  </div>
                </div>
              ))}
            </div>
            
            <div className="tm-flow-important">
              <img src="/images/import-export/documents_new/doc_insurance.png" style={{width: 32, marginTop: -4}} alt="Lightbulb Idea" />
              <div>
                <strong>Important</strong>
                <p>Selecting the right transportation mode can significantly reduce cost, transit time, and supply chain risks.</p>
              </div>
            </div>
          </div>

          {/* RIGHT: 6 CARDS */}
          <div className="tm-cards-grid">
            {TM_CARDS.map((card, i) => (
              <div key={i} className="tm-card">
                <div className="tm-card-num">{card.num}</div>
                <div className="tm-card-img-wrap">
                  {card.img === 'combo' ? (
                    <div style={{position: 'relative', width: 180, height: 110}}>
                      <img src="/images/import-export/documents_new/transport_airplane.png" className="tm-combo-img" style={{width: 80, top: 0, right: 0}} alt="" />
                      <img src="/images/import-export/documents_new/transport_ship.png" className="tm-combo-img" style={{width: 120, bottom: 0, left: 0}} alt="" />
                      <img src="/images/import-export/documents_new/transport_truck.png" className="tm-combo-img" style={{width: 60, bottom: 10, right: -15}} alt="" />
                    </div>
                  ) : (
                    <img src={card.img} className="tm-card-img" alt={card.title} />
                  )}
                </div>
                <div className="tm-card-body">
                  <div className="tm-card-title">{card.title}</div>
                  <div className="tm-card-desc">{card.desc}</div>

                  {card.checks && (
                    <div className="tm-checks">
                      {card.checks.map((chk, idx) => (
                        <div key={idx} className="tm-check-item">
                          <CheckCircle2 size={14} color="#2563eb" /> {chk}
                        </div>
                      ))}
                    </div>
                  )}

                  {card.num === '05' && (
                    <div className="tm-multi-flow">
                      <div className="tm-mf-step"><img src="/images/import-export/documents_new/transport_truck.png" alt=""/> Truck</div>
                      <ArrowRight size={12} color="#94a3b8" />
                      <div className="tm-mf-step"><img src="/images/import-export/documents_new/transport_container.png" alt=""/> Rail</div>
                      <ArrowRight size={12} color="#94a3b8" />
                      <div className="tm-mf-step"><img src="/images/import-export/documents_new/transport_ship.png" alt=""/> Ship</div>
                      <ArrowRight size={12} color="#94a3b8" />
                      <div className="tm-mf-step"><img src="/images/import-export/documents_new/transport_truck.png" alt=""/> Truck</div>
                    </div>
                  )}

                  {card.bestFor && (
                    <div className="tm-card-section">
                      <div className="tm-sec-title best"><img src="/images/import-export/documents_new/doc_certificate-of-origin.png" style={{width: 14}} alt=""/> Best For</div>
                      <div className="tm-badge-list">
                        {card.bestFor.map((item, idx) => <span key={idx} className="tm-pill">{item}</span>)}
                      </div>
                    </div>
                  )}
                  {card.adv && (
                    <div className="tm-card-section">
                      <div className="tm-sec-title adv"><CheckCircle2 size={12}/> Advantages</div>
                      <ul className="tm-card-list">
                        {card.adv.map((item, idx) => <li key={idx}><div className="tm-bullet" style={{background:'#16a34a'}}/>{item}</li>)}
                      </ul>
                    </div>
                  )}
                  {card.lim && (
                    <div className="tm-card-section" style={{marginBottom: 0}}>
                      <div className="tm-sec-title lim">
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
                        Limitations
                      </div>
                      <ul className="tm-card-list">
                        {card.lim.map((item, idx) => <li key={idx}><div className="tm-bullet" style={{background:'#dc2626'}}/>{item}</li>)}
                      </ul>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* BOTTOM BLOCKS */}
        <div className="tm-bottom-blocks">
          <div className="tm-block">
            <div className="tm-block-title">Real Shipment Example</div>
            <div className="tm-block-p">A furniture manufacturer in Vietnam exports products to Germany.</div>
            <div className="tm-shipment-flow">
              {[
                {img: 'participant_manufacturer.png', title: 'Factory', sub: '(Vietnam)'},
                {img: 'transport_truck.png', title: 'Truck', sub: 'Pickup'},
                {img: 'transport_port.png', title: 'Seaport', sub: ''},
                {img: 'transport_ship.png', title: 'Sea Freight', sub: 'to Europe'},
                {img: 'transport_port.png', title: 'Port (Germany)', sub: ''},
                {img: 'transport_truck.png', title: 'Truck', sub: 'Delivery'},
                {img: 'participant_importer.png', title: 'Retailer', sub: '(Germany)'}
              ].map((step, i) => (
                <div key={i} className="tm-sf-step">
                  <img src={`/images/import-export/documents_new/${step.img}`} className="tm-sf-img" alt={step.title} />
                  <div className="tm-sf-text">{step.title}</div>
                  <div className="tm-sf-sub">{step.sub}</div>
                </div>
              ))}
            </div>
            <div className="tm-block-footer">This shipment combines <strong>Road + Sea Freight</strong> to minimize cost.</div>
          </div>
          <div className="tm-block">
            <div className="tm-tip-content">
              <div className="tm-tip-text">
                <div className="tm-block-title"><img src="/images/import-export/documents_new/doc_insurance.png" style={{width:16}} alt=""/> Smart Tip</div>
                <div className="tm-tip-title">There is no "best" transportation mode.</div>
                <div className="tm-tip-desc">The right choice depends on balancing speed, cost, reliability, cargo type, and destination. Modern supply chains often combine multiple modes to achieve the best outcome.</div>
              </div>
              <img src="/images/import-export/documents_new/doc_customs-declaration.png" className="tm-tip-img" alt="Clipboard" />
            </div>
          </div>
        </div>

        {/* STATS */}
        <div className="tm-stats">
          {TM_STATS.map((s, i) => (
            <div key={i} className="tm-stat">
              <div className="tm-stat-val">{s.icon} {s.val}</div>
              <div className="tm-stat-title" style={{whiteSpace: 'pre-line'}}>{s.title}</div>
              <div className="tm-stat-desc" style={{whiteSpace: 'pre-line'}}>{s.desc}</div>
            </div>
          ))}
        </div>

        {/* CTA */}
        <div className="tm-cta">
          <div className="tm-cta-left">
            <div className="tm-cta-icon">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth="2"><path d="M22 10v6M2 10l10-5 10 5-10 5z"/><path d="M6 12v5c3 3 9 3 12 0v-5"/></svg>
            </div>
            <div className="tm-cta-text">
              <h3>Great! You now understand how cargo moves worldwide.</h3>
              <p>Next, let's learn how Customs works and how to clear shipments smoothly.</p>
            </div>
          </div>
          <button className="tm-cta-btn">Next Module: Customs &amp; Compliance →</button>
        </div>

      </div>
    </section>
  );
}

/* ─────────────────────────────────────────
   PAGE
───────────────────────────────────────── */
const CC_TIMELINE = [
  { num: 1, title: 'Shipment Arrives', desc: 'Container reaches destination port.', img: 'cc_sprite_08.png' },
  { num: 2, title: 'Documents Submitted', desc: 'Importer submits required paperwork.', img: 'cc_sprite_09.png' },
  { num: 3, title: 'Document Verification', desc: 'Customs checks invoices and certificates.', img: 'cc_sprite_10.png' },
  { num: 4, title: 'Duty & Tax Assessment', desc: 'Government calculates import duties.', img: 'cc_sprite_11.png' },
  { num: 5, title: 'Inspection', desc: 'If required, shipment is physically inspected.', img: 'cc_sprite_13.png' },
  { num: 6, title: 'Duty Payment', desc: 'Importer pays applicable duties.', img: 'cc_sprite_14.png' },
  { num: 7, title: 'Customs Clearance', desc: 'Shipment officially released.', img: 'cc_sprite_15.png' },
  { num: 8, title: 'Last Mile Delivery', desc: 'Cargo delivered to importer.', img: 'cc_sprite_16.png' }
];

const CC_CARDS = [
  {
    num: '01', title: 'HS Code', img: 'cc_sprite_00.png',
    desc: 'Every internationally traded product has a unique Harmonized System (HS) Code used by customs authorities worldwide.',
    list: ['Product Classification', 'Duty Calculation', 'Trade Statistics']
  },
  {
    num: '02', title: 'Duties & Taxes', img: 'cc_sprite_01.png',
    desc: 'Import duties and taxes are calculated based on product type, value and country of origin.',
    list: ['Import Duty', 'GST / VAT', 'Anti Dumping Duty']
  },
  {
    num: '03', title: 'Customs Inspection', img: 'cc_sprite_02.png',
    desc: 'Some shipments undergo physical inspection before release.',
    list: ['Product Quantity', 'Product Quality', 'Restricted Goods', 'Documentation']
  },
  {
    num: '04', title: 'Restricted Goods', img: 'cc_sprite_04.png',
    desc: 'Some products require special permits or cannot be imported.',
    list: ['Chemicals', 'Medicines', 'Wildlife Products', 'Weapons']
  },
  {
    num: '05', title: 'Customs Broker', img: 'cc_sprite_05.png',
    desc: 'Licensed customs brokers help businesses complete paperwork and clear shipments efficiently.',
    list: ['Documentation', 'Duty Filing', 'Customs Communication']
  },
  {
    num: '06', title: 'Clearance Approved', img: 'cc_sprite_06.png',
    desc: 'After successful clearance the shipment is released for delivery.',
    list: ['Customs Cleared', 'Ready for Delivery']
  }
];

const CC_STATS = [
  { val: '5000+', title: 'HS Product Categories', desc: 'Used for classification of global goods.', icon: 'cc_sprite_20.png' },
  { val: '98%', title: 'Shipments Clear Successfully', desc: 'Most shipments clear without major issues.', icon: 'cc_sprite_15.png' },
  { val: '200+', title: 'Countries Follow HS System', desc: 'A global standard for smooth international trade.', icon: 'cc_sprite_27.png' },
  { val: 'Millions', title: 'Customs Declarations Processed Every Day', desc: 'Keeping global trade moving efficiently.', icon: 'cc_sprite_29.png' }
];

const CC_FLOW = [
  { text: 'Shipment Arrives', img: 'cc_sprite_08.png' },
  { text: 'Documents Submitted', img: 'cc_sprite_09.png' },
  { text: 'Duties Calculated', img: 'cc_sprite_11.png' },
  { text: 'Duties Paid', img: 'cc_sprite_14.png' },
  { text: 'Customs Cleared', img: 'cc_sprite_15.png' },
  { text: 'Delivered to Warehouse', img: 'cc_sprite_16.png' }
];

function CustomsAndCompliance() {
  const BPATH = '/images/import-export/customs/';
  return (
    <section className="cc-section" id="customs-compliance">
      <div className="cc-inner">
        {/* HERO */}
        <div className="cc-hero-row">
          <div className="cc-hero-left">
            <motion.div className="cc-badge" initial={{opacity:0, y:12}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
              <div className="cc-badge-num">06</div>
              <div className="cc-badge-text">CUSTOMS & COMPLIANCE</div>
            </motion.div>
            <motion.h2 initial={{opacity:0, y:16}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.1}}>
              Clear Customs<br/><span>Without Costly Delays</span>
            </motion.h2>
            <motion.p initial={{opacity:0, y:16}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.2}}>
              Every international shipment must pass through Customs before entering or leaving a country. 
              Understand customs clearance, duties & taxes, inspections, HS Codes, and compliance requirements to avoid unnecessary shipment delays.
            </motion.p>
          </div>
          <div className="cc-hero-right">
            <motion.div className="cc-hero-comp" initial={{opacity:0, scale:0.95}} whileInView={{opacity:1, scale:1}} viewport={{once:true}} transition={{duration:0.6}}>
              <img src={`${BPATH}cc_sprite_27.png`} className="cc-hc-img" style={{width: '90%', left: '5%', top: '10%', opacity: 0.2}} alt="" />
              <img src={`${BPATH}cc_sprite_20.png`} className="cc-hc-img" style={{width: 220, left: '10%', top: '20%'}} alt="Customs Building" />
              <img src={`${BPATH}cc_sprite_21.png`} className="cc-hc-img" style={{width: 140, left: '40%', top: '35%', zIndex: 5}} alt="Officer" />
              <img src={`${BPATH}cc_sprite_22.png`} className="cc-hc-img" style={{width: 160, right: '10%', top: '25%', zIndex: 4}} alt="Scanner" />
              <img src={`${BPATH}cc_sprite_28.png`} className="cc-hc-img" style={{width: 70, left: '15%', bottom: '5%', zIndex: 6}} alt="Invoice" />
              <img src={`${BPATH}cc_sprite_29.png`} className="cc-hc-img" style={{width: 70, left: '25%', bottom: '-5%', zIndex: 7}} alt="Bill of Lading" />
              <img src={`${BPATH}cc_sprite_30.png`} className="cc-hc-img" style={{width: 70, left: '35%', bottom: '0%', zIndex: 6}} alt="Certificate" />
              <img src={`${BPATH}cc_sprite_31.png`} className="cc-hc-img" style={{width: 100, right: '5%', bottom: '15%', zIndex: 8}} alt="Approved" />
            </motion.div>
          </div>
        </div>

        {/* MAIN LAYOUT */}
        <div className="cc-main-content">
          {/* Timeline Column */}
          <div className="cc-timeline-col">
            <div className="cc-timeline-card">
              <h3 className="cc-tc-title">Customs Clearance Process</h3>
              <div className="cc-tl-list">
                {CC_TIMELINE.map((item, i) => (
                  <motion.div key={i} className="cc-tl-item" initial={{opacity:0, x:-20}} whileInView={{opacity:1, x:0}} viewport={{once:true}} transition={{delay: i*0.1}}>
                    <div className="cc-tl-num">{item.num}</div>
                    <img src={`${BPATH}${item.img}`} className="cc-tl-img" alt={item.title} />
                    <div className="cc-tl-text">
                      <strong>{item.title}</strong>
                      <span>{item.desc}</span>
                    </div>
                  </motion.div>
                ))}
              </div>
            </div>
            
            <motion.div className="cc-warning-box" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
              <AlertTriangle size={24} color="#1e40af" style={{flexShrink:0}} />
              <div>
                <strong>Important</strong>
                <p>Incomplete or incorrect documentation is one of the biggest reasons shipments get delayed at Customs.</p>
              </div>
            </motion.div>
          </div>

          {/* Cards Grid Column */}
          <div className="cc-cards-col">
            <div className="cc-cards-grid">
              {CC_CARDS.map((card, i) => (
                <motion.div key={i} className="cc-card" initial={{opacity:0, y:30}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay: i*0.1}} whileHover={{y:-4, transition:{duration:0.2}}}>
                  <div className="cc-card-num">{card.num}</div>
                  <div className="cc-card-img-wrap">
                    <img src={`${BPATH}${card.img}`} className="cc-card-img" alt={card.title} />
                  </div>
                  <div className="cc-card-body">
                    <h3 className="cc-card-title">{card.title}</h3>
                    <p className="cc-card-desc">{card.desc}</p>
                    <div className="cc-badge-list">
                      {card.list.map((li, idx) => (
                        <div key={idx} className="cc-badge-item">
                          <CheckCircle2 className="cc-check-icon" />
                          <span>{li}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        </div>

        {/* BOTTOM BLOCKS */}
        <div className="cc-bottom-blocks">
          <motion.div className="cc-block" initial={{opacity:0, y:30}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
            <h3 className="cc-block-title">Real World Example</h3>
            <p className="cc-block-p">A company imports laptops from Japan into India.</p>
            <div className="cc-flow">
              {CC_FLOW.map((f, i) => (
                <div key={i} className="cc-f-step">
                  <img src={`${BPATH}${f.img}`} className="cc-f-img" alt={f.text} />
                  <div className="cc-f-text">{f.text}</div>
                </div>
              ))}
            </div>
          </motion.div>

          <motion.div className="cc-block" initial={{opacity:0, y:30}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.1}}>
            <h3 className="cc-block-title">
              <Lightbulb size={18} color="#f59e0b" /> Smart Tip
            </h3>
            <div className="cc-tip-content">
              <div className="cc-tip-text">
                <p className="cc-tip-desc">
                  <strong>Think of Customs as airport immigration for cargo.</strong><br/>
                  Passengers need passports. Cargo needs trade documents. If everything is correct, Customs lets the shipment continue.
                </p>
              </div>
              <img src={`${BPATH}cc_sprite_00.png`} className="cc-tip-img" alt="Clipboard" />
            </div>
          </motion.div>
        </div>

        {/* STATS */}
        <div className="cc-stats">
          {CC_STATS.map((stat, i) => (
            <motion.div key={i} className="cc-stat" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay: i*0.1}}>
              <div className="cc-stat-val">
                <img src={`${BPATH}${stat.icon}`} className="cc-stat-icon" alt="" />
                {stat.val}
              </div>
              <div className="cc-stat-title">{stat.title}</div>
              <div className="cc-stat-desc">{stat.desc}</div>
            </motion.div>
          ))}
        </div>

        {/* CTA */}
        <motion.div className="cc-cta" initial={{opacity:0, y:30}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
          <div className="cc-cta-left">
            <div className="cc-cta-icon">
              <GraduationCap size={28} />
            </div>
            <div className="cc-cta-text">
              <h3>Great! Your shipment has now cleared Customs.</h3>
              <p>Next, let's learn about Incoterms and responsibilities in international trade.</p>
            </div>
          </div>
          <button className="cc-cta-btn">Next Module: Incoterms 2020 →</button>
        </motion.div>
      </div>
    </section>
  );
}


const INCO_CARDS = [
  { name: 'EXW', title: 'Ex Works', desc: 'Seller makes goods available at their premises.', risk: 'Risk at Seller\'s Premises', icons: ['doc_0.png', 'timeline_2.png', 'timeline_4.png'], img: 'any_0.png' },
  { name: 'FCA', title: 'Free Carrier', desc: 'Seller delivers to carrier or another nominated party.', risk: 'Risk at Carrier', icons: ['timeline_2.png', 'timeline_4.png'], img: 'any_1.png' },
  { name: 'CPT', title: 'Carriage Paid To', desc: 'Seller pays for carriage to the destination.', risk: 'Risk at Carrier', icons: ['timeline_2.png', 'timeline_4.png', 'timeline_5.png'], img: 'any_2.png' },
  { name: 'CIP', title: 'Carriage & Insurance Paid To', desc: 'Seller pays for carriage and insurance.', risk: 'Risk at Carrier', icons: ['timeline_2.png', 'timeline_4.png', 'timeline_5.png'], img: 'any_3.png' },
  { name: 'DAP', title: 'Delivered At Place', desc: 'Seller delivers goods to the named place of destination.', risk: 'Risk at Named Place', icons: ['timeline_2.png', 'timeline_4.png'], img: 'any_4.png' },
  { name: 'DPU', title: 'Delivered at Place Unloaded', desc: 'Seller delivers and unloads at destination.', risk: 'Risk at Named Place', icons: ['timeline_2.png', 'timeline_4.png', 'timeline_5.png'], img: 'any_5.png' },
  { name: 'DDP', title: 'Delivered Duty Paid', desc: 'Seller delivers, clears customs and pays all duties.', risk: 'Risk at Buyer\'s Premises', icons: ['timeline_2.png', 'timeline_4.png', 'timeline_5.png'], img: 'any_6.png' },
  { name: 'FAS', title: 'Free Alongside Ship', desc: 'Seller places goods alongside the ship at port.', risk: 'Risk alongside Ship', icons: ['timeline_5.png'], img: 'sea_0.png' },
  { name: 'FOB', title: 'Free On Board', desc: 'Seller delivers goods on board the vessel.', risk: 'Risk on Board Ship', icons: ['timeline_5.png'], img: 'sea_1.png' },
  { name: 'CFR', title: 'Cost & Freight', desc: 'Seller pays for cost and freight to destination.', risk: 'Risk on Board Ship', icons: ['timeline_5.png'], img: 'sea_2.png' },
  { name: 'CIF', title: 'Cost, Insurance & Freight', desc: 'Seller pays cost, freight and insurance.', risk: 'Risk on Board Ship', icons: ['timeline_5.png'], img: 'sea_3.png' }
];

const INCO_TIMELINE = [
  { label: 'FACTORY', img: 'timeline_0.png' },
  { label: 'LOADING', img: 'timeline_1.png' },
  { label: 'TRUCK', img: 'timeline_2.png' },
  { label: 'EXPORT CUSTOMS', img: 'timeline_3.png' },
  { label: 'PORT', img: 'timeline_4.png' },
  { label: 'SHIP', img: 'timeline_5.png' },
  { label: 'IMPORT PORT', img: 'timeline_6.png' },
  { label: 'IMPORT CUSTOMS', img: 'timeline_3.png' },
  { label: 'WAREHOUSE', img: 'timeline_7.png' },
  { label: 'BUYER', img: 'timeline_8.png' }
];

const INCO_DEEP = [
  {
    name: 'EXW – Ex Works', img: 'any_0.png',
    seller: 'Makes goods available.',
    buyer: 'All transport, export, import, duties and delivery.',
    bestFor: 'Experienced importers.',
    example: 'Buyer arranges pickup from seller\'s factory in India.'
  },
  {
    name: 'FOB – Free On Board', img: 'sea_1.png',
    seller: 'Delivers goods on board vessel at port.',
    buyer: 'Main transport, insurance, import clearance.',
    bestFor: 'Ocean freight shipments.',
    example: 'Seller ships goods from Mumbai Port.'
  },
  {
    name: 'CIF – Cost, Insurance & Freight', img: 'sea_3.png',
    seller: 'Pays for cost, freight and insurance.',
    buyer: 'Import clearance, duties, delivery.',
    bestFor: 'Buyers who want seller to arrange main transport.',
    example: 'Seller ships and insures goods to Hamburg.'
  },
  {
    name: 'DDP – Delivered Duty Paid', img: 'any_6.png',
    seller: 'Everything including duties and delivery.',
    buyer: 'Only accepts delivery.',
    bestFor: 'Buyers who want zero hassle.',
    example: 'Goods delivered to buyer\'s warehouse in Germany.'
  }
];

const INCO_STATS = [
  { val: '11', title: 'Official Incoterms®', desc: 'in Incoterms 2020', icon: 'doc_0.png' },
  { val: '200+', title: 'Countries Using ICC Rules', desc: 'A global standard.', icon: 'timeline_8.png' },
  { val: '90%', title: 'International Contracts', desc: 'Reference Incoterms', icon: 'doc_2.png' },
  { val: 'Millions', title: 'of Shipments Governed', desc: 'Every Year', icon: 'doc_3.png' }
];

const INCO_MISTAKES = [
  'Using FOB for air freight shipments.',
  'Thinking CIF includes customs clearance.',
  'Assuming DDP always means cheapest.',
  'Ignoring insurance requirements.',
  'Not understanding risk transfer points.'
];

function IncotermsSection() {
  const BPATH = '/images/import-export/incoterms_new/';
  
  return (
    <section className="inco-section" id="incoterms-2020">
      <div className="inco-inner">
        {/* HERO */}
        <div className="inco-hero-row">
          <div className="inco-hero-header">
            <motion.div className="inco-badge" initial={{opacity:0, y:12}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
              <div className="inco-badge-num">07</div>
              <div className="inco-badge-text">INCOTERMS® 2020</div>
            </motion.div>
            <motion.h2 initial={{opacity:0, y:16}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.1}}>
              Master Global Trade<br/>with <span>Incoterms® 2020</span>
            </motion.h2>
            <motion.p initial={{opacity:0, y:16}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.2}}>
              Know exactly who pays, who ships, and who takes the risk at every stage of international trade.
            </motion.p>
          </div>
          
          <motion.div className="inco-hero-journey" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{duration:0.6}}>
            
            <div className="inco-hj-nodes">
              <div className="inco-hj-node" style={{left: '0%'}}>
                <div className="inco-hj-label">FACTORY</div>
                <img src={`${BPATH}hero_0.png`} className="inco-hj-img" alt="Factory" />
              </div>
              <div className="inco-hj-node" style={{left: '16.66%'}}>
                <div className="inco-hj-label">EXPORT<br/>WAREHOUSE</div>
                <img src={`${BPATH}hero_1.png`} className="inco-hj-img" alt="Export Warehouse" />
              </div>
              <div className="inco-hj-node" style={{left: '33.33%'}}>
                <div className="inco-hj-label">EXPORT<br/>PORT</div>
                <img src={`${BPATH}hero_3.png`} className="inco-hj-img" alt="Export Port" />
              </div>
              <div className="inco-hj-node" style={{left: '50%'}}>
                <div className="inco-hj-label">CARGO<br/>SHIP</div>
                <img src={`${BPATH}hero_4.png`} className="inco-hj-img" alt="Cargo Ship" style={{transform:'scale(1.2) translateY(-10px)'}} />
              </div>
              <div className="inco-hj-node" style={{left: '66.66%'}}>
                <div className="inco-hj-label">IMPORT<br/>PORT</div>
                <img src={`${BPATH}hero_5.png`} className="inco-hj-img" alt="Import Port" />
              </div>
              <div className="inco-hj-node" style={{left: '83.33%'}}>
                <div className="inco-hj-label">IMPORT<br/>WAREHOUSE</div>
                <img src={`${BPATH}timeline_7.png`} className="inco-hj-img" alt="Import Warehouse" />
              </div>
              <div className="inco-hj-node" style={{left: '100%'}}>
                <div className="inco-hj-label">BUYER</div>
                <img src={`${BPATH}timeline_8.png`} className="inco-hj-img" alt="Buyer" />
              </div>
              {/* Trucks linking them */}
              <img src={`${BPATH}hero_2.png`} className="inco-hj-truck" style={{left: '8%', bottom: 0}} alt="" />
              <img src={`${BPATH}hero_2.png`} className="inco-hj-truck" style={{left: '25%', bottom: 0}} alt="" />
              <img src={`${BPATH}hero_2.png`} className="inco-hj-truck" style={{left: '75%', bottom: 0}} alt="" />
            </div>

            {/* Responsibility Bars */}
            <div className="inco-resp-bars-hero">
              <div className="inco-rb-row">
                <div className="inco-rb-label" style={{color: '#3b82f6'}}>SELLER'S RESPONSIBILITY</div>
                <div className="inco-rb-line inco-rb-seller" style={{width: '100%'}}></div>
              </div>
              <div className="inco-rb-row">
                <div className="inco-rb-label" style={{color: '#22c55e'}}>BUYER'S RESPONSIBILITY</div>
                <div className="inco-rb-line inco-rb-buyer" style={{width: '83.33%', marginLeft: '16.66%'}}></div>
              </div>
              <div className="inco-rb-row">
                <div className="inco-rb-label" style={{color: '#f59e0b'}}>RISK TRANSFER POINT</div>
                <div className="inco-rb-line inco-rb-risk" style={{width: '50%'}}>
                  <MapPin className="inco-rb-pin" size={20} style={{right:-10, top: -14, fill: '#fff'}} />
                </div>
              </div>
              <div className="inco-rb-row">
                <div className="inco-rb-label" style={{color: '#8b5cf6'}}>COST TRANSFER POINT</div>
                <div className="inco-rb-line inco-rb-cost" style={{width: '66.66%'}}>
                  <div className="inco-rb-pin-cost" style={{right:-10}}>$</div>
                </div>
              </div>
            </div>
          </motion.div>
        </div>

        {/* EXPLORE INCOTERMS */}
        <div className="inco-explore">
          <div className="inco-section-title">
            <div className="inco-st-num">1</div>
            Explore Incoterms® 2020
          </div>
          <p className="inco-section-desc">Incoterms are international rules that define the responsibilities<br/>of buyers and sellers in global trade.</p>
          
          <div className="inco-cards-container">
            <div className="inco-sea-label">SEA & INLAND WATERWAY ONLY</div>
            <div className="inco-explore-grid">
              {INCO_CARDS.map((card, i) => (
                <motion.div key={i} className="inco-card" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay: i*0.03}}>
                  <h4>{card.name}</h4>
                  <img src={`${BPATH}${card.img}`} className="inco-card-img" alt={card.name} />
                  <div className="inco-card-name">{card.title}</div>
                  <div className="inco-card-desc">{card.desc}</div>
                  <div className="inco-card-risk">{card.risk}</div>
                  <div className="inco-card-icons">
                    {card.icons.map((icon, idx) => (
                      <img key={idx} src={`${BPATH}${icon}`} style={{height: 14}} alt=""/>
                    ))}
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        </div>

        {/* TIMELINE */}
        <div className="inco-timeline">
          <div className="inco-section-title">
            <div className="inco-st-num">2</div>
            Responsibility Timeline
          </div>
          <p className="inco-section-desc">Click on any Incoterm above to see how responsibility, risk and cost transfer across the journey.</p>
          
          <motion.div className="inco-tl-container" initial={{opacity:0, y:30}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
            <div className="inco-tl-flow">
              {INCO_TIMELINE.map((step, i) => (
                <div key={i} className="inco-tl-step">
                  <img src={`${BPATH}${step.img}`} className="inco-tl-img" alt={step.label} />
                  <div className="inco-tl-label">{step.label}</div>
                </div>
              ))}
            </div>

            <div className="inco-resp-bars-tl">
              <div className="inco-rb-row">
                <img src={`${BPATH}doc_0.png`} className="inco-rb-icon" alt="" />
                <div className="inco-rb-label">Seller's Responsibility</div>
                <div className="inco-rb-line inco-rb-seller" style={{width: '66.66%'}}></div>
                <img src={`${BPATH}timeline_8.png`} className="inco-rb-icon" alt="" style={{marginLeft: 8}}/>
                <div className="inco-rb-label" style={{marginLeft: 8}}>Buyer's Responsibility</div>
                <div className="inco-rb-line inco-rb-buyer" style={{flex: 1}}></div>
              </div>
              <div className="inco-rb-row">
                <MapPin size={16} color="#f59e0b" className="inco-rb-icon" fill="#fff" />
                <div className="inco-rb-label">Risk Transfer Point</div>
                <div className="inco-rb-line inco-rb-risk" style={{width: '66.66%'}}>
                  <MapPin className="inco-rb-pin" size={20} style={{right:-10, top:-14, fill:'#fff'}} />
                </div>
              </div>
              <div className="inco-rb-row">
                <div className="inco-rb-icon-cost">$</div>
                <div className="inco-rb-label">Cost Transfer Point</div>
                <div className="inco-rb-line inco-rb-cost" style={{width: '88.88%'}}>
                  <div className="inco-rb-pin-cost" style={{right:-10}}>$</div>
                </div>
              </div>
              
              <div className="inco-tl-legend">
                <span>Legend:</span>
                <span className="inco-leg-item"><div className="inco-leg-color" style={{background:'#3b82f6'}}></div> Blue = Seller</span>
                <span className="inco-leg-item"><div className="inco-leg-color" style={{background:'#22c55e'}}></div> Green = Buyer</span>
                <span className="inco-leg-item"><div className="inco-leg-color" style={{background:'#f59e0b'}}></div> Orange = Risk Transfer</span>
                <span className="inco-leg-item"><div className="inco-leg-color" style={{background:'#8b5cf6'}}></div> Purple = Cost Transfer</span>
              </div>
            </div>
          </motion.div>
        </div>

        {/* DEEP DIVE */}
        <div className="inco-section-title">
          <div className="inco-st-num">3</div>
          Deep Dive: Popular Incoterms
        </div>
        <p className="inco-section-desc">The most commonly used Incoterms explained in detail.</p>
        
        <div className="inco-deep">
          <div className="inco-deep-grid">
            {INCO_DEEP.map((dd, i) => (
              <motion.div key={i} className="inco-dd-card" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay: i*0.1}}>
                <div className="inco-dd-header">
                  <div className="inco-dd-title">{dd.name}</div>
                  <img src={`${BPATH}${dd.img}`} className="inco-dd-img" alt={dd.name} />
                </div>
                <div className="inco-dd-list">
                  <div className="inco-dd-item">
                    <div className="inco-dd-bullet">• <strong>What Seller Does</strong></div>
                    <p>{dd.seller}</p>
                  </div>
                  <div className="inco-dd-item">
                    <div className="inco-dd-bullet">• <strong>What Buyer Does</strong></div>
                    <p>{dd.buyer}</p>
                  </div>
                  <div className="inco-dd-item">
                    <div className="inco-dd-bullet">• <strong>Best For</strong></div>
                    <p>{dd.bestFor}</p>
                  </div>
                  <div className="inco-dd-item">
                    <div className="inco-dd-bullet">• <strong>Example</strong></div>
                    <p>{dd.example}</p>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* BOTTOM ROWS */}
        <div className="inco-bottom-row">
          <motion.div className="inco-box inco-rw" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
            <div className="inco-section-title">
              <div className="inco-st-num">4</div>
              Real World Example
            </div>
            <p className="inco-section-desc" style={{marginBottom:32}}>A furniture exporter in Vietnam sells goods to a buyer in Germany<br/>under FOB Ho Chi Minh Port.</p>
            
            <div className="inco-rw-flow">
              <div className="inco-rw-step">
                <img src={`${BPATH}hero_0.png`} className="inco-rw-img" alt="Factory" />
                <div className="inco-rw-label">Factory<br/>(Vietnam)</div>
              </div>
              <ArrowRight size={16} color="#94a3b8" />
              <div className="inco-rw-step">
                <img src={`${BPATH}hero_2.png`} className="inco-rw-img" alt="Truck" style={{transform:'scale(0.8)'}}/>
                <div className="inco-rw-label">Truck to<br/>Port</div>
              </div>
              <ArrowRight size={16} color="#94a3b8" />
              <div className="inco-rw-step">
                <img src={`${BPATH}hero_3.png`} className="inco-rw-img" alt="HCM Port" />
                <div className="inco-rw-label">Ho Chi Minh<br/>Port</div>
              </div>
              <ArrowRight size={16} color="#94a3b8" />
              <div className="inco-rw-step">
                <img src={`${BPATH}hero_4.png`} className="inco-rw-img" alt="Ship" />
                <div className="inco-rw-label">Ocean Freight</div>
              </div>
              <ArrowRight size={16} color="#94a3b8" />
              <div className="inco-rw-step">
                <img src={`${BPATH}hero_5.png`} className="inco-rw-img" alt="Hamburg" />
                <div className="inco-rw-label">Hamburg Port</div>
              </div>
              <ArrowRight size={16} color="#94a3b8" />
              <div className="inco-rw-step">
                <img src={`${BPATH}hero_2.png`} className="inco-rw-img" alt="Truck" style={{transform:'scale(0.8)'}}/>
                <div className="inco-rw-label">Truck to<br/>Warehouse</div>
              </div>
            </div>
            
            <div className="inco-rw-bars">
              <div className="inco-rw-bar-row">
                <div className="inco-rw-seller">
                  <div className="inco-rw-bar-label">Seller's Responsibility</div>
                  <div className="inco-rw-bar-desc">Up to goods loaded on ship</div>
                </div>
                <div className="inco-rw-buyer">
                  <div className="inco-rw-bar-label">Buyer's Responsibility</div>
                  <div className="inco-rw-bar-desc">From this point onwards</div>
                </div>
              </div>
              <MapPin className="inco-rw-pin" size={24} fill="#fff" color="#f59e0b" />
            </div>
          </motion.div>

          <div className="inco-bottom-col2">
            <motion.div className="inco-box" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.1}}>
              <div className="inco-section-title">
                <div className="inco-st-num">5</div>
                Common Mistakes
              </div>
              <div className="inco-mistake-list">
                {INCO_MISTAKES.map((msg, i) => (
                  <div key={i} className="inco-mistake-item">
                    <div className="inco-mistake-x">✕</div>
                    <span>{msg}</span>
                  </div>
                ))}
              </div>
            </motion.div>

            <motion.div className="inco-box inco-tip" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.2}}>
              <div className="inco-section-title">
                <div className="inco-st-num">6</div>
                Smart Tip
              </div>
              <p>Think of Incoterms like splitting the bill during a road trip.</p>
              <p>Each person takes responsibility for different parts of the journey.</p>
              <p>The goods always reach the same destination, but Incoterms decide who pays, who manages and who takes the risk along the way.</p>
              
              <div className="inco-tip-roadtrip">
                <img src={`${BPATH}road_trip.png`} alt="Road Trip" />
              </div>
            </motion.div>
          </div>
        </div>

        {/* STATS */}
        <div className="inco-stats-row">
          <div className="inco-section-title" style={{marginRight: 40}}>
            <div className="inco-st-num">7</div>
            By The Numbers
          </div>
          <div className="inco-stats-grid">
            {INCO_STATS.map((stat, i) => (
              <motion.div key={i} className="inco-stat-box" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay: i*0.1}}>
                <img src={`${BPATH}${stat.icon}`} className="inco-stat-icon" alt="" />
                <div className="inco-stat-text">
                  <div className="inco-stat-val">{stat.val}</div>
                  <div className="inco-stat-title">{stat.title}</div>
                  <div className="inco-stat-desc">{stat.desc}</div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* FINAL CTA */}
        <motion.div className="inco-final-cta" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
          <div className="inco-fc-left">
            <div className="inco-fc-icon"><img src={`${BPATH}doc_1.png`} style={{width: 32}} alt="" /></div>
            <div className="inco-fc-text">
              <h3>Great! You now understand Incoterms® 2020.</h3>
              <p>Next, let's learn how international payments and trade finance work.</p>
            </div>
          </div>
          <button className="inco-fc-btn">
            Next Module:<br/><strong>Payment & Finance</strong> <ArrowRight size={16} />
          </button>
        </motion.div>

      </div>
    </section>
  );
}


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
  { text: 'Sending goods before payment', img: '/images/import-export/transport/ship.png' },
  { text: 'Ignoring currency fluctuation', img: '/images/import-export/documents_new/doc_insurance.png' },
  { text: 'Wrong beneficiary details', img: '/images/import-export/documents_new/doc_commercial-invoice.png' },
  { text: 'Incorrect LC documents', img: '/images/import-export/incoterms_new/doc_3.png' },
  { text: 'Missing SWIFT codes', img: '/images/import-export/incoterms_new/doc_0.png' }
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
                <svg width="100%" height="100%" style={{position:'absolute', top:0, left:0, zIndex: 0}}>
                   <ellipse cx="50%" cy="50%" rx="45%" ry="35%" fill="none" stroke="#2563eb" strokeWidth="1.5" strokeDasharray="6 6" />
                </svg>
                {/* Simulated Workflow using absolute positioning */}
                <img src="/images/import-export/modules/factory.png" className="pay-hd-item" style={{top:'40%', left:'0%'}} alt="Exporter"/>
                <div className="pay-hd-label" style={{top:'70%', left:'0%'}}>Exporter</div>
                
                <img src="/images/import-export/documents_new/doc_invoice.png" className="pay-hd-item" style={{top:'20%', left:'20%', height: 70}} alt="Invoice"/>
                <div className="pay-hd-label" style={{top:'35%', left:'20%'}}>Invoice</div>

                <img src="/images/import-export/participants/customs.png" className="pay-hd-item" style={{top:'10%', left:'40%'}} alt="Bank"/>
                <div className="pay-hd-label" style={{top:'40%', left:'40%'}}>Exporter's Bank</div>

                <div className="pay-hd-swift" style={{top:'20%', left:'60%'}}>SWIFT</div>

                <img src="/images/import-export/participants/customs.png" className="pay-hd-item" style={{top:'10%', left:'80%'}} alt="Bank"/>
                <div className="pay-hd-label" style={{top:'40%', left:'80%'}}>Importer's Bank</div>

                <img src="/images/import-export/participants/importer.png" className="pay-hd-item" style={{top:'40%', left:'95%'}} alt="Buyer"/>
                <div className="pay-hd-label" style={{top:'70%', left:'95%'}}>Buyer</div>

                <img src="/images/import-export/transport/ship.png" className="pay-hd-item" style={{top:'80%', left:'20%', height: 120}} alt="Ship"/>
                <img src="/images/import-export/transport/airplane.png" className="pay-hd-item" style={{top:'75%', left:'50%', height: 100}} alt="Plane"/>
                <img src="/images/import-export/transport/truck.png" className="pay-hd-item" style={{top:'85%', left:'80%', height: 90}} alt="Truck"/>
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
                 <div className="pay-pt-bar-wrap"><div className="pay-pt-bar" style={{left: '0%', width: '25%', background: '#22c55e'}}></div></div>
               </div>
               <div className="pay-pt-row">
                 <div className="pay-pt-label">Letter of Credit</div>
                 <div className="pay-pt-bar-wrap"><div className="pay-pt-bar" style={{left: '25%', width: '60%', background: '#eab308'}}></div></div>
               </div>
               <div className="pay-pt-row">
                 <div className="pay-pt-label">Documentary Collection</div>
                 <div className="pay-pt-bar-wrap"><div className="pay-pt-bar" style={{left: '50%', width: '35%', background: '#f97316'}}></div></div>
               </div>
               <div className="pay-pt-row">
                 <div className="pay-pt-label">Open Account</div>
                 <div className="pay-pt-bar-wrap"><div className="pay-pt-bar" style={{left: '85%', width: '15%', background: '#ef4444'}}></div></div>
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
                   <img src={m.img} style={{height:80, marginBottom:16, objectFit: 'contain'}} alt=""/>
                   <p>{m.text}</p>
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


export default function ImportExportBasicsPage() {
  useEffect(() => { window.scrollTo(0, 0); }, []);

  return (
    <div className="ieb-page">

      {/* ══════════════════════════════════════
          BREADCRUMB
      ══════════════════════════════════════ */}
      <nav className="ieb-breadcrumb">
        <div className="ieb-bc-inner">
          <Link to="/">Home</Link>
          <span className="ieb-bc-sep">›</span>
          <Link to="/trade-intelligence">Trade Intelligence</Link>
          <span className="ieb-bc-sep">›</span>
          <span className="ieb-bc-active">Import Export Basics</span>
        </div>
      </nav>

      {/* ══════════════════════════════════════
          HERO
      ══════════════════════════════════════ */}
      <section className="ieb-hero">

        {/* Ambient gradient blobs */}
        <div className="ieb-hero-blob ieb-blob-1" />
        <div className="ieb-hero-blob ieb-blob-2" />

        <div className="ieb-hero-grid">

          {/* ────────────────── LEFT ────────────────── */}
          <div className="ieb-hero-left">

            <motion.div
              className="ieb-badge"
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4 }}
            >
              <Globe size={12} />
              TRADE INTELLIGENCE
            </motion.div>

            <motion.h1
              className="ieb-title"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.08 }}
            >
              Import &amp;&nbsp;<br />Export Basics
            </motion.h1>

            <motion.p
              className="ieb-sub"
              initial={{ opacity: 0, y: 18 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.16 }}
            >
              Your complete guide to global trade fundamentals. Understand the
              process, people, documents, and strategies before shipping your
              first container.
            </motion.p>

            <motion.div
              className="ieb-actions"
              initial={{ opacity: 0, y: 18 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.24 }}
            >
              <a href="#roadmap" className="ieb-btn-primary">
                Start Learning →
              </a>
              <button className="ieb-btn-outline">
                <span className="ieb-play-ring">
                  <Play size={12} fill="currentColor" />
                </span>
                Watch 5 Minute Overview
              </button>
            </motion.div>

            <motion.div
              className="ieb-stats-row"
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.32 }}
            >
              {STATS.map((s, i) => (
                <div key={i} className="ieb-stat-item">
                  {i > 0 && <div className="ieb-stat-sep" />}
                  <span className="ieb-stat-ic" style={{ color: s.color }}>{s.icon}</span>
                  <div>
                    <strong>{s.val}</strong>
                    <span>{s.lbl}</span>
                  </div>
                </div>
              ))}
            </motion.div>

          </div>

          {/* ────────────────── RIGHT ILLUSTRATION ────────────────── */}
          <div className="ieb-hero-right">

            {/* SVG Trade Journey */}
            <motion.div
              className="ieb-illustration-wrap"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.9, delay: 0.1 }}
            >
              <TradeJourneyIllustration />
            </motion.div>

            {/* ── Floating Document Cards ── */}
            {FLOAT_CARDS.map((card, i) => (
              <motion.div
                key={i}
                className={`ieb-float-card ieb-float-${i}`}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.55, delay: 0.4 + card.delay }}
                style={{ '--card-accent': card.color }}
              >
                <span className="ieb-fc-emoji">{card.emoji}</span>
                <div className="ieb-fc-text">
                  <strong>{card.title}</strong>
                  <span>{card.sub}</span>
                </div>
                <div className="ieb-fc-dot" />
              </motion.div>
            ))}

            {/* Central "In Transit" status pill */}
            <motion.div
              className="ieb-transit-pill"
              initial={{ opacity: 0, scale: 0.88 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5, delay: 0.7 }}
            >
              <span className="ieb-pulse-dot" />
              Shipment in Transit · Live
            </motion.div>

          </div>
        </div>
      </section>

      {/* ══════════════════════════════════════
          LEARNING ROADMAP
      ══════════════════════════════════════ */}
      <section className="ieb-roadmap-section" id="roadmap">
        <div className="ieb-roadmap-wrap">

          <motion.div
            className="ieb-roadmap-header"
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
          >
            <div className="ieb-roadmap-eyebrow">
              <Clock size={12} />
              YOUR LEARNING PATH
            </div>
            <h2>Master Global Trade<br /><span>Step by Step</span></h2>
            <p>9 focused modules. Go at your own pace. Start anywhere.</p>
          </motion.div>

          <div className="ieb-roadmap-grid">
            {ROADMAP.map((item, i) => (
              <motion.div
                key={i}
                className="ieb-roadmap-card"
                initial={{ opacity: 0, y: 28 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: i * 0.07 }}
              >
                {/* Connector line */}
                {i < ROADMAP.length - 1 && (
                  <div className="ieb-connector">
                    <div className="ieb-connector-line" />
                    <ArrowRight size={12} className="ieb-connector-arrow" />
                  </div>
                )}

                <div className="ieb-rc-step">{item.step}</div>
                <div className="ieb-rc-icon">{item.icon}</div>
                <h3 className="ieb-rc-title">{item.title}</h3>
                <p className="ieb-rc-desc">{item.desc}</p>
                <button className="ieb-rc-cta">
                  Start Module <ArrowRight size={12} />
                </button>
              </motion.div>
            ))}
          </div>

        </div>
      </section>

      {/* ══════════════════════════════════════
          SECTION 2: UNDERSTANDING IMPORT & EXPORT
      ══════════════════════════════════════ */}
      <UnderstandingSection />

      {/* ══════════════════════════════════════
          GLOBAL STATS STRIP
      ══════════════════════════════════════ */}
      <GlobalStatsStrip />

      {/* ══════════════════════════════════════
          SECTION 3: GLOBAL TRADE JOURNEY
      ══════════════════════════════════════ */}
      <GlobalTradeJourney />

      {/* ══════════════════════════════════════
          SECTION 4: MEET THE PARTICIPANTS
      ══════════════════════════════════════ */}
      <MeetTheParticipants />

      {/* ══════════════════════════════════════
          SECTION 4: ESSENTIAL TRADE DOCUMENTS
      ══════════════════════════════════════ */}
      <EssentialTradeDocuments />

      {/* ══════════════════════════════════════
          SECTION 5: TRANSPORTATION MODES
      ══════════════════════════════════════ */}
      <TransportationModes />

      {/* ══════════════════════════════════════
          SECTION 6: CUSTOMS & COMPLIANCE
      ══════════════════════════════════════ */}
      <CustomsAndCompliance />

      {/* ══════════════════════════════════════
          SECTION 7: INCOTERMS 2020
      ══════════════════════════════════════ */}
      <IncotermsSection />


      {/* ══════════════════════════════════════
          SECTION 8: PAYMENT & FINANCE
      ══════════════════════════════════════ */}
      <PaymentFinanceSection />

    </div>
  );
}
