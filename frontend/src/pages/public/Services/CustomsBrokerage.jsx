import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';

/* ─── Scroll Reveal Hook ─── */
function useReveal() {
  const ref = useRef(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      ([e]) => {
        if (e.isIntersecting) {
          el.classList.add('opacity-100', 'translate-y-0');
          el.classList.remove('opacity-0', 'translate-y-8');
          obs.unobserve(el);
        }
      },
      { threshold: 0.15, rootMargin: '0px 0px -50px 0px' }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, []);
  return ref;
}

function Reveal({ children, className = '', delay = '' }) {
  const ref = useReveal();
  return (
    <div ref={ref} className={`transition-all duration-1000 opacity-0 translate-y-8 ${delay} ${className}`}>
      {children}
    </div>
  );
}

const chaStats = [
  { value: '50+', label: 'Ports Integrated', desc: 'Active customs clearances at all major ICDs, Airports & Sea Ports' },
  { value: '< 12 Hrs', label: 'Avg. Air Clearance', desc: 'Rapid clearance and customs assessment turnaround' },
  { value: '99.8%', label: 'First-Time-Right', desc: 'Minimizing amendment queries and query memos' },
  { value: 'AEO Certified', label: 'Trust Status', desc: 'Secure supply chain trust status by CBIC / Customs' },
];

const customsCapabilities = [
  { icon: '📜', title: 'Bill of Entry & Shipping Bills', desc: 'Direct electronic filing of import and export bills via ICEGATE. Optimized for speed and tariff compliance.', color: 'from-amber-400 to-amber-600' },
  { icon: '🏛️', title: 'In-House CHA Network', desc: 'Direct liaisons with Custom House Agents at major air complexes (IGI, BOM, BLR) and sea ports (JNPT, Mundra, Chennai).', color: 'from-blue-400 to-blue-600' },
  { icon: '📦', title: 'Bonded Warehousing', desc: 'Seamless cargo shifting to public or private bonded warehouses under Section 59 / 60 bonds, deferring duty payment.', color: 'from-purple-400 to-purple-600' },
  { icon: '🛡️', title: 'AEO / Authorized Shippers', desc: 'Fast-track green channel clearances for clients under AEO Tier-1 & Tier-2 programs, reducing port examination frequency.', color: 'from-teal-400 to-teal-600' },
  { icon: '💰', title: 'Duty Drawback & Refund claims', desc: 'Automated processing and tracking of Drawback (DBK) claims, MEIS/RODTEP schemes, and GST refund reconciliations.', color: 'from-green-400 to-green-600' },
  { icon: '✈️', title: 'Direct Port Delivery (DPD)', desc: 'Facilitating immediate container transit directly from ship-to-factory for pre-approved corporate importers, saving port demurrage.', color: 'from-indigo-400 to-indigo-600' },
];

const calculatorCommodities = [
  {
    name: 'Lithium-Ion Battery Cells',
    hsn: '8507.60.00',
    bcd: 10,  // Basic Customs Duty %
    sws: 10,  // Social Welfare Surcharge (10% of BCD)
    igst: 18, // Integrated GST %
    incentive: 'RODTEP Active (1.2%)',
  },
  {
    name: 'Active Pharmaceutical Ingredients (APIs)',
    hsn: '2933.39.19',
    bcd: 7.5,
    sws: 10,
    igst: 18,
    incentive: 'PLI Scheme Applicable',
  },
  {
    name: 'Crystalline Solar PV Modules',
    hsn: '8541.43.00',
    bcd: 40,  // BCD under anti-dumping/domestic tariff safeguard
    sws: 10,
    igst: 18,
    incentive: 'Safeguard Duty Included',
  },
];

export default function CustomsBrokerage() {
  const [commodityIdx, setCommodityIdx] = useState(0);
  const [assessableValue, setAssessableValue] = useState(1000000); // 10 Lakhs default

  const current = calculatorCommodities[commodityIdx];
  
  // Calculations
  const bcdAmount = (assessableValue * current.bcd) / 100;
  const swsAmount = (bcdAmount * current.sws) / 100;
  const subtotalForIgst = assessableValue + bcdAmount + swsAmount;
  const igstAmount = (subtotalForIgst * current.igst) / 100;
  const totalDuties = bcdAmount + swsAmount + igstAmount;
  const totalLandedCost = assessableValue + totalDuties;

  return (
    <div className="bg-white min-h-screen overflow-hidden">
      {/* ═══ HERO ═══ */}
      <section className="relative pt-32 pb-32 lg:pt-48 lg:pb-48 bg-gradient-to-br from-indigo-50/20 via-slate-50 to-teal-50/20 text-slate-900 overflow-hidden">
        {/* Port Operations Watermark Background */}
        <div className="absolute inset-0 z-0">
          <img
            src="/assets/port_operations.png"
            alt="Port Customs Operations Terminal"
            className="w-full h-full object-cover object-center opacity-[0.06] scale-105 mix-blend-multiply animate-[pulse_20s_ease-in-out_infinite_alternate]"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-white via-white/40 to-slate-50/80"></div>
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-7xl h-full pointer-events-none opacity-40">
            <div className="absolute -top-[10%] right-[-5%] w-[450px] h-[450px] rounded-full bg-brand-indigo/10 blur-3xl"></div>
            <div className="absolute bottom-[-10%] left-[-5%] w-[550px] h-[550px] rounded-full bg-brand-teal/10 blur-3xl"></div>
          </div>
        </div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 relative z-10">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
            <Reveal>
              <div>
                <div className="inline-flex items-center justify-center px-4 py-1.5 mb-6 text-xs font-bold text-brand-indigo bg-brand-indigo/10 rounded-full border border-brand-indigo/20 tracking-widest uppercase">
                  <span className="mr-2">📜</span> Customs CHA Services
                </div>
                <h1 className="text-5xl md:text-6xl lg:text-7xl font-extrabold tracking-tight mb-6 leading-tight text-slate-900">
                  Friction-Free <br />
                  <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand-indigo to-brand-teal">
                    Customs Clearance
                  </span>
                </h1>
                <p className="text-xl text-slate-600 max-w-2xl font-light leading-relaxed mb-10">
                  End-to-end Customs House Brokerage (CHA) in India. Experienced filings under all export-import schemes, direct port delivery (DPD) agreements, and integrated duty assessment models.
                </p>

                <div className="flex flex-wrap gap-4">
                  <Link
                    to="/contact"
                    className="px-8 py-4 bg-brand-indigo hover:bg-indigo-700 text-white font-bold rounded-full transition-colors shadow-[0_0_20px_rgba(90,79,207,0.3)]"
                  >
                    Consult Customs CHA Expert
                  </Link>
                  <a
                    href="#duty-calculator-demo"
                    className="px-8 py-4 bg-white/80 border border-slate-200 text-slate-700 font-bold rounded-full hover:bg-slate-100 hover:border-slate-300 transition-colors shadow-sm backdrop-blur-md"
                  >
                    Landed Cost Calculator
                  </a>
                </div>
              </div>
            </Reveal>

            {/* Visual Hero Element - Simulated Bill of Entry Status Card */}
            <Reveal delay="delay-200">
              <div className="relative h-[400px] bg-white/80 rounded-3xl border border-slate-200/80 backdrop-blur-md p-6 shadow-2xl flex flex-col justify-between overflow-hidden">
                <div className="absolute -top-32 -right-32 w-64 h-64 bg-brand-teal/10 rounded-full blur-[80px]"></div>

                <div className="flex justify-between items-center relative z-10">
                  <div>
                    <div className="text-sm text-slate-500 mb-1">ICEGATE Portal</div>
                    <div className="text-slate-800 font-bold text-xl flex items-center gap-2">
                      <span className="w-2.5 h-2.5 rounded-full bg-green-500 animate-pulse"></span> OOC Granted
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm text-slate-500 mb-1">Bill of Entry No.</div>
                    <div className="text-slate-800 font-bold text-xl font-mono">BOE-5892019</div>
                  </div>
                </div>

                <div className="relative z-10 py-8">
                  <div className="flex justify-between text-3xl font-black mb-2 text-slate-800">
                    <span>MUNDRA SEA</span>
                    <span className="text-brand-indigo">📜</span>
                    <span>FACTORY DPD</span>
                  </div>

                  <div className="relative h-2 bg-slate-200 rounded-full mt-4">
                    <div className="absolute top-0 left-0 h-full bg-gradient-to-r from-brand-indigo to-brand-teal rounded-full w-[100%]">
                      <div className="absolute -right-2 -top-1 w-4 h-4 bg-green-500 border-2 border-white rounded-full shadow-[0_0_10px_rgba(0,0,0,0.15)]"></div>
                    </div>
                  </div>
                  <div className="flex justify-between text-xs text-slate-500 mt-2 font-mono">
                    <span>Customs Port: INMUN1</span>
                    <span>Out of Charge (OOC) Issued</span>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 relative z-10">
                  <div className="bg-slate-50/80 border border-slate-100 rounded-xl p-3">
                    <div className="text-xs text-slate-500">Duty Assessment</div>
                    <div className="text-md font-bold text-slate-800">₹4,23,590 Paid</div>
                  </div>
                  <div className="bg-slate-50/80 border border-slate-100 rounded-xl p-3">
                    <div className="text-xs text-slate-500">CHA Facilitator</div>
                    <div className="text-md font-bold text-green-600">Freel CHA Tier-1</div>
                  </div>
                </div>
              </div>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ═══ STATS BAR (Floating) ═══ */}
      <Reveal>
        <section className="max-w-6xl mx-auto px-4 sm:px-6 relative z-20 -mt-10">
          <div className="bg-white rounded-[2rem] p-8 shadow-[0_20px_50px_-10px_rgba(0,0,0,0.06)] border border-slate-100 grid grid-cols-2 md:grid-cols-4 gap-6 divide-x-0 md:divide-x divide-slate-100">
            {chaStats.map((s, i) => (
              <div key={i} className="text-center group">
                <div className="text-3xl md:text-4xl font-black text-transparent bg-clip-text bg-gradient-to-br from-brand-indigo to-brand-teal mb-2 group-hover:scale-105 transition-transform">
                  {s.value}
                </div>
                <div className="text-sm font-bold uppercase tracking-wider text-slate-700 mb-1">{s.label}</div>
                <div className="text-xs text-slate-400 px-2">{s.desc}</div>
              </div>
            ))}
          </div>
        </section>
      </Reveal>

      {/* ═══ CUSTOMS CAPABILITIES GRID ═══ */}
      <section className="py-32 bg-white relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="text-center mb-20">
            <span className="text-brand-indigo font-bold tracking-widest uppercase text-sm mb-3 block">
              Operational Customs
            </span>
            <h2 className="text-4xl md:text-5xl font-extrabold text-brand-navy mb-4 leading-tight">
              CHA Services for High-Volume Importers
            </h2>
            <p className="text-slate-500 text-lg max-w-2xl mx-auto">
              From automated customs tariff classifications to bonded storage logistics, we provide direct on-the-ground support at every major trade gate.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {customsCapabilities.map((v, i) => (
              <Reveal key={i} delay={`delay-${(i % 3) * 100}`}>
                <div className="bg-slate-50/50 rounded-[2rem] p-8 border border-slate-200/60 hover:border-brand-indigo/60 hover:shadow-2xl hover:bg-white transition-all group h-full flex flex-col justify-between">
                  <div>
                    <div className={`w-14 h-14 rounded-2xl bg-gradient-to-br ${v.color} flex items-center justify-center text-3xl text-white shadow-lg mb-6 group-hover:scale-110 transition-transform`}>
                      {v.icon}
                    </div>
                    <h3 className="text-xl font-bold text-slate-800 mb-3 group-hover:text-brand-indigo transition-colors">
                      {v.title}
                    </h3>
                    <p className="text-slate-500 text-sm leading-relaxed mb-6">{v.desc}</p>
                  </div>
                  <div className="inline-flex items-center text-xs font-bold text-brand-indigo group-hover:translate-x-1.5 transition-transform cursor-pointer">
                    Scheme Checklist Available <span className="ml-1">→</span>
                  </div>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ INTERACTIVE LIVE DUTY & LANDED COST CALCULATOR WIDGET ═══ */}
      <section id="duty-calculator-demo" className="py-32 bg-brand-navy text-white relative overflow-hidden">
        <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxwYXRoIGQ9Ik0zNiAzNHYtNGgtdjRoLTR2NGgtdjRoNHY0aDR2LTRoNHptMC0xMnYtNGgtdjRoLTR2NGgtdjRoNHY0aDR2LTRoNHptLTExIDEwaC00djRoLTR2LTRoLTR2LTRoNHYtNGg0djRoNHY0em0tMTEgMTBoLTR2NGgtdjRoLTR2LTRoNHYtNGg0djRoNHY0em0xMSAwaC00djRoLTR2LTRoLTR2LTRoNHYtNGg0djRoNHY0em0xMSAwaC00djRoLTR2LTRoLTR2LTRoNHYtNGg0djRoNHY0eiIgZmlsbD0iI2ZmZmZmZiIgZmlsbC1vcGFjaXR5PSIwLjAyIi8+PC9nPjwvc3ZnPg==')] opacity-50 z-0"></div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 relative z-10">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            
            <Reveal>
              <div>
                <span className="text-brand-teal font-bold tracking-widest uppercase text-sm mb-3 block">
                  Landed Cost Simulator
                </span>
                <h2 className="text-4xl md:text-5xl font-extrabold mb-6">Interactive Duty Calculator</h2>
                <p className="text-slate-300 text-lg mb-8 leading-relaxed">
                  Indian Customs utilizes multi-layered assessments (Basic Duty, SWS, IGST). Select a commodity and adjust the CIF value below to estimate duties and landed cost in real-time.
                </p>

                {/* Commodity Selection Buttons */}
                <div className="space-y-4 mb-8">
                  {calculatorCommodities.map((item, idx) => (
                    <button
                      key={idx}
                      onClick={() => setCommodityIdx(idx)}
                      className={`w-full text-left p-5 rounded-2xl border transition-all ${
                        commodityIdx === idx
                          ? 'bg-white/10 border-brand-teal shadow-[0_0_15px_rgba(0,191,165,0.2)]'
                          : 'bg-transparent border-white/10 hover:bg-white/5'
                      }`}
                    >
                      <div className="flex justify-between items-center">
                        <span className="font-bold text-white text-md">{item.name}</span>
                        <span className="text-xs font-mono bg-white/10 text-brand-teal px-2.5 py-1 rounded-full">
                          HSN {item.hsn}
                        </span>
                      </div>
                      <div className="text-xs text-slate-400 mt-2 font-mono">
                        Basic Duty: {item.bcd}% • IGST: {item.igst}% • {item.incentive}
                      </div>
                    </button>
                  ))}
                </div>

                {/* CIF / Assessable Value Slider */}
                <div className="bg-white/5 border border-white/10 rounded-2xl p-6">
                  <div className="flex justify-between text-xs font-bold font-mono text-slate-300 mb-3">
                    <span>ASSESSABLE CIF VALUE (INR)</span>
                    <span className="text-brand-teal text-sm">₹{assessableValue.toLocaleString('en-IN')}</span>
                  </div>
                  <input
                    type="range"
                    min="100000"
                    max="10000000"
                    step="100000"
                    value={assessableValue}
                    onChange={(e) => setAssessableValue(Number(e.target.value))}
                    className="w-full h-2 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-brand-teal"
                  />
                  <div className="flex justify-between text-[10px] font-mono text-slate-500 mt-2">
                    <span>₹1 Lakh</span>
                    <span>₹50 Lakhs</span>
                    <span>₹1 Crore</span>
                  </div>
                </div>
              </div>
            </Reveal>

            {/* Simulated Duty Assessment Receipt Sheet */}
            <Reveal delay="delay-200">
              <div className="bg-white rounded-[2.5rem] p-8 text-slate-800 shadow-2xl border border-slate-200 relative overflow-hidden">
                <div className="absolute top-0 right-0 w-32 h-32 bg-brand-indigo/5 rounded-full blur-3xl pointer-events-none"></div>

                <div className="relative z-10">
                  <div className="flex justify-between items-center border-b border-slate-200 pb-4 mb-6">
                    <div>
                      <div className="font-mono text-slate-400 text-xs uppercase">Customs Assessment Draft</div>
                      <div className="text-slate-800 font-bold text-lg font-mono">BOE_CALC_TEMP</div>
                    </div>
                    <div className="text-xs bg-slate-100 text-brand-indigo px-3 py-1.5 rounded-full font-bold uppercase tracking-wide">
                      Assessed Draft
                    </div>
                  </div>

                  <div className="space-y-4">
                    <div className="flex justify-between text-sm text-slate-500 font-mono">
                      <span>Commodity Profile:</span>
                      <span className="text-slate-800 font-bold text-right max-w-[200px] truncate">{current.name}</span>
                    </div>
                    <div className="flex justify-between text-sm text-slate-500 font-mono border-b border-slate-100 pb-2">
                      <span>HS Tariff Code:</span>
                      <span className="text-slate-800 font-bold font-mono">{current.hsn}</span>
                    </div>

                    <div className="flex justify-between text-sm text-slate-600 font-medium mt-4">
                      <span>Assessable CIF Value (A)</span>
                      <span className="font-mono">₹{assessableValue.toLocaleString('en-IN')}</span>
                    </div>

                    <div className="flex justify-between text-sm text-slate-600">
                      <span>Basic Customs Duty (B) <span className="text-slate-400 font-mono">[{current.bcd}%]</span></span>
                      <span className="font-mono text-slate-800">₹{bcdAmount.toLocaleString('en-IN')}</span>
                    </div>

                    <div className="flex justify-between text-sm text-slate-600">
                      <span>Social Welfare Surcharge (C) <span className="text-slate-400 font-mono">[10% of BCD]</span></span>
                      <span className="font-mono text-slate-800">₹{swsAmount.toLocaleString('en-IN')}</span>
                    </div>

                    <div className="flex justify-between text-sm text-slate-600">
                      <span>Integrated GST (D) <span className="text-slate-400 font-mono">[{current.igst}% of A+B+C]</span></span>
                      <span className="font-mono text-slate-800">₹{igstAmount.toLocaleString('en-IN')}</span>
                    </div>

                    <div className="border-t border-slate-200/80 pt-4 mt-6">
                      <div className="flex justify-between text-sm font-bold text-slate-800 uppercase">
                        <span>Total Customs Duties</span>
                        <span className="text-brand-indigo font-mono">₹{totalDuties.toLocaleString('en-IN')}</span>
                      </div>
                      <div className="flex justify-between text-xs text-slate-400 font-mono mt-1">
                        <span>Landed Cost overhead</span>
                        <span>+ {((totalDuties / assessableValue) * 100).toFixed(1)}%</span>
                      </div>
                    </div>

                    <div className="bg-slate-50 rounded-2xl p-4 border border-slate-100 mt-6 flex justify-between items-center">
                      <div>
                        <div className="text-xs text-slate-500 font-bold uppercase tracking-wider">Estimated Landed Cost</div>
                        <div className="text-[10px] text-slate-400 font-mono">CIF Value + Total Duties</div>
                      </div>
                      <div className="text-2xl font-black text-brand-navy font-mono">
                        ₹{totalLandedCost.toLocaleString('en-IN')}
                      </div>
                    </div>
                  </div>

                  <div className="mt-6 flex gap-3">
                    <button className="flex-1 py-3.5 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-bold rounded-xl transition-all font-mono">
                      📥 Export PDF Quote
                    </button>
                    <Link
                      to="/contact"
                      className="flex-1 py-3.5 bg-brand-indigo text-white text-xs font-bold rounded-xl hover:bg-indigo-700 transition-all font-mono text-center flex items-center justify-center"
                    >
                      Book Customs Entry
                    </Link>
                  </div>
                </div>
              </div>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ═══ CTA SECTION ═══ */}
      <section className="py-24 bg-white text-center">
        <Reveal>
          <div className="max-w-3xl mx-auto px-4">
            <h2 className="text-4xl md:text-5xl font-black text-brand-navy mb-6">Need Customs Representation?</h2>
            <p className="text-lg text-slate-600 mb-10">
              Avoid cargo port delays. Let our licensed in-house Custom House Agents handle your electronic Bill of Entry filings, duty assessments, and port examinations under direct port delivery.
            </p>
            <Link
              to="/contact"
              className="inline-flex items-center gap-2 px-10 py-5 bg-gradient-to-r from-brand-indigo to-indigo-700 text-white font-bold text-lg rounded-full hover:shadow-[0_10px_30px_rgba(90,79,207,0.3)] transition-all hover:-translate-y-1"
            >
              Request Customs Consultation <span>📜</span>
            </Link>
          </div>
        </Reveal>
      </section>
    </div>
  );
}
