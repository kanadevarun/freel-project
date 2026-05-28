import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import '../../../styles/Solutions.css';

/* ─── Scroll Reveal ─── */
function useReveal() {
  const ref = useRef(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      ([e]) => { if (e.isIntersecting) { el.classList.add('visible'); obs.unobserve(el); } },
      { threshold: 0.15 }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, []);
  return ref;
}

function Reveal({ children, className = '', delay = '' }) {
  const ref = useReveal();
  return <div ref={ref} className={`reveal ${delay} ${className}`}>{children}</div>;
}

export default function Compliance() {
  const [activeTab, setActiveTab] = useState('hsn');

  const tabs = [
    { id: 'hsn', label: 'HSN Lookup', icon: '🔍' },
    { id: 'msds', label: 'MSDS Manager', icon: '📋' },
    { id: 'kyc', label: 'Customer KYC', icon: '🛡️' },
    { id: 'customs', label: 'Customs Norms', icon: '🌐' }
  ];

  return (
    <div className="bg-white min-h-screen">
      {/* ═══ HERO ═══ */}
      <section className="relative pt-32 pb-16 lg:pt-48 lg:pb-24 overflow-hidden">
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[800px] bg-brand-teal opacity-[0.04] rounded-full blur-3xl -z-10"></div>
        
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10 text-center">
          <div className="inline-flex items-center justify-center px-4 py-1.5 mb-6 text-sm font-semibold text-brand-teal bg-brand-teal/10 rounded-full border border-brand-teal/20 uppercase tracking-wide">
            <span className="mr-2">🛡️</span> Risk & Compliance
          </div>
          <h1 className="text-5xl md:text-6xl lg:text-7xl font-extrabold text-brand-navy tracking-tight mb-6 leading-tight">
            Automate Your <br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand-teal to-brand-indigo">Global Trade Compliance</span>
          </h1>
          <p className="mt-4 text-xl md:text-2xl text-ui-gray max-w-3xl mx-auto font-light leading-relaxed">
            HSN code lookups, MSDS verification, KYC management, and customs duty calculators — built directly into your shipping flow.
          </p>

          <div className="mt-10 flex flex-col sm:flex-row justify-center gap-4">
            <Link to="/signup" className="px-8 py-4 text-white font-semibold text-lg bg-gradient-to-r from-brand-teal to-brand-indigo rounded-full shadow-lg shadow-teal-500/30 hover:shadow-teal-500/50 hover:-translate-y-1 transition-all flex items-center justify-center">
              Try it Free
            </Link>
          </div>
        </div>
      </section>

      {/* ═══ FEATURE HIGHLIGHTS ═══ */}
      <section className="py-24 bg-slate-50 border-y border-ui-border">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <Reveal>
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold text-brand-navy mb-4">A Complete Compliance Center</h2>
              <p className="text-lg text-ui-gray max-w-2xl mx-auto">
                Avoid customs delays, penalties, and cargo rejections by validating everything before the shipment even gets booked.
              </p>
            </div>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
            <Reveal delay="reveal-delay-1">
              <div className="bg-white p-8 rounded-2xl border border-ui-border shadow-sm h-full flex flex-col items-start">
                <div className="w-12 h-12 bg-blue-50 text-blue-600 rounded-xl flex items-center justify-center text-2xl mb-6">🔍</div>
                <h3 className="text-xl font-bold text-brand-navy mb-3">HSN Code Engine</h3>
                <p className="text-slate-600 text-sm">Instantly find the right 8-digit HSN code for your product. We pull live GST rates, export incentives, and basic customs duties automatically.</p>
              </div>
            </Reveal>
            
            <Reveal delay="reveal-delay-2">
              <div className="bg-white p-8 rounded-2xl border border-ui-border shadow-sm h-full flex flex-col items-start">
                <div className="w-12 h-12 bg-red-50 text-red-600 rounded-xl flex items-center justify-center text-2xl mb-6">⚠️</div>
                <h3 className="text-xl font-bold text-brand-navy mb-3">Hazmat & MSDS</h3>
                <p className="text-slate-600 text-sm">Upload Material Safety Data Sheets (MSDS). Our AI flags hazardous classes (DG), UN numbers, and alerts carriers immediately.</p>
              </div>
            </Reveal>
            
            <Reveal delay="reveal-delay-3">
              <div className="bg-white p-8 rounded-2xl border border-ui-border shadow-sm h-full flex flex-col items-start">
                <div className="w-12 h-12 bg-green-50 text-green-600 rounded-xl flex items-center justify-center text-2xl mb-6">📋</div>
                <h3 className="text-xl font-bold text-brand-navy mb-3">Document Wallet</h3>
                <p className="text-slate-600 text-sm">Store IEC, GST certificates, AD code letters, and LUTs centrally. Never ask your clients to email the same documents twice.</p>
              </div>
            </Reveal>

            <Reveal delay="reveal-delay-4">
              <div className="bg-white p-8 rounded-2xl border border-ui-border shadow-sm h-full flex flex-col items-start">
                <div className="w-12 h-12 bg-purple-50 text-purple-600 rounded-xl flex items-center justify-center text-2xl mb-6">🛡️</div>
                <h3 className="text-xl font-bold text-brand-navy mb-3">Automated KYC</h3>
                <p className="text-slate-600 text-sm">Onboard new shippers and vendors with one-click PAN and GST verification via government APIs. 100% fraud prevention.</p>
              </div>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ═══ INTERACTIVE DASHBOARD VISUALIZATION ═══ */}
      <section className="py-24 bg-white">
        <div className="max-w-6xl mx-auto px-4 sm:px-6">
          <Reveal>
            <div className="flex flex-col md:flex-row justify-between items-end mb-8">
              <div>
                <h2 className="text-3xl font-bold text-brand-navy mb-2">Live Compliance Checks</h2>
                <p className="text-ui-gray">See how Freel validates data in real-time.</p>
              </div>
              <div className="flex space-x-2 bg-slate-100 p-1 rounded-lg mt-4 md:mt-0">
                {tabs.map(tab => (
                  <button 
                    key={tab.id}
                    onClick={() => setActiveTab(tab.id)}
                    className={`px-4 py-2 rounded-md text-sm font-semibold transition-all ${activeTab === tab.id ? 'bg-white shadow-sm text-brand-teal' : 'text-slate-500 hover:text-brand-navy'}`}
                  >
                    {tab.icon} {tab.label}
                  </button>
                ))}
              </div>
            </div>
          </Reveal>

          <Reveal>
            <div className="bg-[#0f172a] rounded-2xl p-6 shadow-2xl border border-slate-700 overflow-hidden relative min-h-[400px]">
              {/* Window Controls */}
              <div className="flex space-x-2 mb-6">
                <div className="w-3 h-3 rounded-full bg-red-500"></div>
                <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
                <div className="w-3 h-3 rounded-full bg-green-500"></div>
              </div>

              {/* Dynamic Content */}
              <div className="font-mono text-sm text-slate-300">
                {activeTab === 'hsn' && (
                  <div className="space-y-4 animate-fade-in">
                    <div className="flex items-center space-x-4 mb-8">
                      <div className="bg-slate-800 p-2 rounded flex-1">
                        <span className="text-slate-500">Query:</span> "Methanol Industrial Grade"
                      </div>
                      <div className="bg-brand-teal text-white px-4 py-2 rounded font-sans font-bold">Search API</div>
                    </div>
                    
                    <div className="bg-slate-800 border border-slate-700 rounded-lg p-4 flex items-start space-x-4">
                      <div className="text-2xl mt-1">🧪</div>
                      <div className="flex-1">
                        <div className="text-white font-bold text-lg mb-1">HSN: 2905.11</div>
                        <div className="text-slate-400 mb-3">Methanol (Methyl Alcohol)</div>
                        <div className="grid grid-cols-3 gap-4 border-t border-slate-700 pt-3">
                          <div>
                            <span className="block text-xs text-slate-500">Basic Duty</span>
                            <span className="text-brand-teal font-bold">7.5%</span>
                          </div>
                          <div>
                            <span className="block text-xs text-slate-500">IGST</span>
                            <span className="text-white">18%</span>
                          </div>
                          <div>
                            <span className="block text-xs text-slate-500">SWS</span>
                            <span className="text-white">10%</span>
                          </div>
                        </div>
                      </div>
                      <div className="bg-red-500/20 text-red-400 px-3 py-1 rounded text-xs border border-red-500/30">
                        HAZARDOUS (Class 3)
                      </div>
                    </div>
                  </div>
                )}

                {activeTab === 'msds' && (
                  <div className="space-y-4 animate-fade-in">
                    <div className="bg-slate-800 border-l-4 border-yellow-500 p-4 rounded mb-4">
                      <div className="flex items-center text-yellow-400 font-bold mb-2">
                        <span className="mr-2">⚠️</span> MSDS Parsing Complete
                      </div>
                      <div className="text-slate-400">Document: <span className="text-white">Methanol_MSDS_TataChem.pdf</span></div>
                    </div>
                    
                    <table className="w-full text-left border-collapse">
                      <thead>
                        <tr className="border-b border-slate-700 text-slate-500">
                          <th className="py-2">Field</th>
                          <th className="py-2">Extracted Value</th>
                          <th className="py-2">Status</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr className="border-b border-slate-800">
                          <td className="py-3">UN Number</td>
                          <td className="py-3 text-white font-bold">1230</td>
                          <td className="py-3"><span className="text-green-400">✓ Validated</span></td>
                        </tr>
                        <tr className="border-b border-slate-800">
                          <td className="py-3">Proper Shipping Name</td>
                          <td className="py-3 text-white">METHANOL</td>
                          <td className="py-3"><span className="text-green-400">✓ Validated</span></td>
                        </tr>
                        <tr className="border-b border-slate-800">
                          <td className="py-3">Hazard Class</td>
                          <td className="py-3 text-red-400 font-bold">Class 3 (Flammable Liquid)</td>
                          <td className="py-3"><span className="text-green-400">✓ Alert Set</span></td>
                        </tr>
                        <tr>
                          <td className="py-3">Packing Group</td>
                          <td className="py-3 text-white">II</td>
                          <td className="py-3"><span className="text-green-400">✓ Validated</span></td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                )}

                {activeTab === 'kyc' && (
                  <div className="space-y-6 animate-fade-in">
                    <div className="flex space-x-4">
                      <div className="flex-1 bg-slate-800 p-4 rounded-lg border border-slate-700">
                        <div className="text-xs text-slate-500 mb-1">Verify GSTIN</div>
                        <div className="flex">
                          <input type="text" value="27AADCB2230M1Z2" readOnly className="bg-slate-900 border border-slate-700 px-3 py-2 rounded-l text-white flex-1" />
                          <button className="bg-brand-teal text-white px-4 rounded-r font-sans text-sm font-bold">Verify</button>
                        </div>
                      </div>
                    </div>

                    <div className="bg-green-500/10 border border-green-500/30 p-6 rounded-lg">
                      <div className="flex items-center text-green-400 font-bold text-lg mb-4">
                        <span className="mr-2">✓</span> Active & Verified (Government API)
                      </div>
                      <div className="grid grid-cols-2 gap-y-4">
                        <div>
                          <div className="text-xs text-slate-500">Legal Name</div>
                          <div className="text-white">BHARAT CHEMICALS PVT LTD</div>
                        </div>
                        <div>
                          <div className="text-xs text-slate-500">State</div>
                          <div className="text-white">Maharashtra (27)</div>
                        </div>
                        <div>
                          <div className="text-xs text-slate-500">Taxpayer Type</div>
                          <div className="text-white">Regular</div>
                        </div>
                        <div>
                          <div className="text-xs text-slate-500">Filing Status</div>
                          <div className="text-white">Up to date</div>
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                {activeTab === 'customs' && (
                  <div className="flex items-center justify-center h-64 flex-col text-center animate-fade-in">
                    <div className="text-6xl mb-4">🌐</div>
                    <div className="text-xl text-white font-bold mb-2">Global Customs Duty Calculator</div>
                    <p className="text-slate-400 max-w-md">Input your origin, destination, and HSN code to instantly calculate landed costs including all duties and taxes.</p>
                  </div>
                )}
              </div>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ═══ CTA SECTION ═══ */}
      <section className="py-24 bg-ui-surface">
        <Reveal>
          <div className="max-w-4xl mx-auto px-4 sm:px-6 text-center">
            <h2 className="text-3xl md:text-5xl font-bold text-brand-navy mb-6">Ship with Confidence.</h2>
            <p className="text-ui-gray text-lg mb-10">
              Stop relying on manual checks and PDF reading. Let our operating system handle the compliance so you can focus on growth.
            </p>
            <Link to="/contact" className="inline-flex items-center gap-2 px-8 py-4 bg-brand-navy text-white font-bold text-lg rounded-full hover:bg-brand-teal transition-colors shadow-lg hover:shadow-xl hover:-translate-y-1">
              Talk to an Expert <span>→</span>
            </Link>
          </div>
        </Reveal>
      </section>

    </div>
  );
}
