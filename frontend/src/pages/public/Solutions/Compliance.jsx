import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { CheckCircle2, CheckCircle, Zap, ShieldCheck, Globe, UploadCloud, ScanSearch, FileText, UserCheck, Calculator, AlertTriangle, TrendingUp, Clock, History, FileSearch, Activity, Package, Cpu } from 'lucide-react';
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

/* ─── Animated Number Component ─── */
function AnimatedNumber({ value, duration = 2 }) {
  const [count, setCount] = useState(0);
  const nodeRef = useRef(null);
  const inView = true; // Simplified for basic IntersectionObserver usage or we can just animate on mount if inside a WhileInView wrapper

  useEffect(() => {
    let startTimestamp = null;
    const step = (timestamp) => {
      if (!startTimestamp) startTimestamp = timestamp;
      const progress = Math.min((timestamp - startTimestamp) / (duration * 1000), 1);
      // easeOutQuart
      const ease = 1 - Math.pow(1 - progress, 4);
      setCount(Math.floor(ease * value));
      if (progress < 1) {
        window.requestAnimationFrame(step);
      }
    };
    const observer = new IntersectionObserver(([entry]) => {
      if (entry.isIntersecting) {
        window.requestAnimationFrame(step);
        observer.disconnect();
      }
    });
    if (nodeRef.current) observer.observe(nodeRef.current);
    return () => observer.disconnect();
  }, [value, duration]);

  return <span ref={nodeRef}>{count}</span>;
}

export default function Compliance() {
  // State for sequential animation of activity feed
  const [showActivity, setShowActivity] = useState(0);

  useEffect(() => {
    const timer1 = setTimeout(() => setShowActivity(1), 1000);
    const timer2 = setTimeout(() => setShowActivity(2), 2000);
    const timer3 = setTimeout(() => setShowActivity(3), 3000);
    return () => { clearTimeout(timer1); clearTimeout(timer2); clearTimeout(timer3); };
  }, []);

  return (
    <div className="bg-white min-h-screen">
      {/* ═══ HERO ═══ */}
      <section className="relative pt-[120px] pb-12 overflow-hidden bg-[#F8FAFC]">
        {/* Subtle Background Map Pattern (Placeholder) */}
        <div className="absolute inset-0 opacity-[0.03] pointer-events-none bg-[radial-gradient(#1e293b_1px,transparent_1px)] [background-size:24px_24px]"></div>
        
        <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-10 items-center">
            
            {/* LEFT CONTENT */}
            <motion.div 
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6 }}
              className="pr-0 lg:pr-10"
            >
              <div className="inline-flex items-center px-4 py-1.5 mb-6 text-sm font-bold tracking-widest uppercase rounded-full bg-teal-50 border border-teal-100 shadow-sm">
                <span className="text-brand-teal mr-2">🌐</span>
                <span className="text-brand-teal">
                  GLOBAL TRADE COMPLIANCE
                </span>
              </div>
              
              <h1 className="text-[38px] md:text-[48px] lg:text-[56px] font-[800] text-brand-navy leading-[1.1] tracking-tight mb-6">
                Stop Compliance Issues
                <span className="block text-transparent bg-clip-text bg-gradient-to-r from-teal-400 to-blue-600 mt-1">
                  Before They Stop Shipments.
                </span>
              </h1>
              
              <p className="text-[20px] lg:text-[22px] text-slate-500 max-w-[600px] mb-10 leading-[1.6]">
                Verify HSN codes, automate KYC, validate hazardous goods, and calculate customs duties before cargo ever reaches customs.
              </p>
              
              <div className="flex flex-wrap items-center gap-4 mb-10">
                <Link to="/signup" className="px-8 py-4 text-white font-semibold text-lg bg-gradient-to-r from-teal-400 to-blue-600 rounded-full shadow-lg shadow-teal-500/30 hover:shadow-teal-500/50 transition-all flex items-center justify-center hover:-translate-y-0.5">
                  Get Started <span className="ml-2">→</span>
                </Link>
                <button className="px-8 py-4 text-brand-navy font-semibold text-lg bg-white border border-slate-300 rounded-full hover:bg-slate-50 transition-all flex items-center justify-center hover:-translate-y-0.5">
                  <span className="mr-2 text-brand-indigo">▶</span> See Compliance In Action
                </button>
              </div>
              
              <div className="flex flex-wrap items-center gap-x-6 gap-y-4 text-sm font-semibold text-slate-700">
                <div className="flex items-center gap-2">
                  <div className="w-5 h-5 rounded-full bg-green-500 text-white flex items-center justify-center text-xs">✓</div>
                  HSN Validation
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-5 h-5 rounded-full bg-green-500 text-white flex items-center justify-center text-xs">✓</div>
                  Automated KYC
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-5 h-5 rounded-full bg-green-500 text-white flex items-center justify-center text-xs">✓</div>
                  MSDS Analysis
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-5 h-5 rounded-full bg-green-500 text-white flex items-center justify-center text-xs">✓</div>
                  Duty Calculation
                </div>
              </div>
            </motion.div>

            {/* RIGHT SIDE VISUAL */}
            <motion.div
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8, delay: 0.2 }}
              className="relative w-full h-[550px] lg:h-[650px] flex items-center justify-center mt-10 lg:mt-0"
            >
              {/* Floating Micro Cards */}
              <motion.div 
                animate={{ y: [-5, 5, -5] }} 
                transition={{ duration: 6, repeat: Infinity, ease: "easeInOut" }}
                className="absolute -top-6 -right-4 lg:-right-10 z-30 bg-white px-4 py-3 rounded-xl shadow-xl shadow-slate-200/50 border border-slate-100 flex items-center gap-3"
              >
                <div className="w-8 h-8 rounded-full bg-green-500 text-white flex items-center justify-center font-bold">✓</div>
                <div>
                  <div className="text-sm font-bold text-brand-navy">IEC Verified</div>
                  <div className="text-xs text-slate-500 font-medium">● Active</div>
                </div>
              </motion.div>

              <motion.div 
                animate={{ y: [5, -5, 5] }} 
                transition={{ duration: 7, repeat: Infinity, ease: "easeInOut", delay: 1 }}
                className="absolute -top-6 -left-4 lg:-left-12 z-30 bg-white px-4 py-3 rounded-xl shadow-xl shadow-slate-200/50 border border-slate-100 flex items-center gap-3"
              >
                <div className="w-8 h-8 rounded-full bg-green-500 text-white flex items-center justify-center font-bold">✓</div>
                <div>
                  <div className="text-sm font-bold text-brand-navy">GST Valid</div>
                  <div className="text-xs text-slate-500 font-medium">Verified</div>
                </div>
              </motion.div>

              <motion.div 
                animate={{ y: [-4, 6, -4] }} 
                transition={{ duration: 5.5, repeat: Infinity, ease: "easeInOut", delay: 2 }}
                className="absolute -bottom-6 -left-4 lg:-left-10 z-30 bg-white px-4 py-3 rounded-xl shadow-xl shadow-slate-200/50 border border-slate-100 flex items-center gap-3"
              >
                <div className="w-8 h-8 rounded-full bg-green-500 text-white flex items-center justify-center font-bold">✓</div>
                <div>
                  <div className="text-sm font-bold text-brand-navy">Customs Cleared</div>
                  <div className="text-xs text-slate-500 font-medium">No Holds</div>
                </div>
              </motion.div>

              <motion.div 
                animate={{ y: [4, -6, 4] }} 
                transition={{ duration: 6.5, repeat: Infinity, ease: "easeInOut", delay: 1.5 }}
                className="absolute -bottom-6 -right-4 lg:-right-8 z-30 bg-white px-4 py-3 rounded-xl shadow-xl shadow-slate-200/50 border border-slate-100 flex items-center gap-3"
              >
                <div className="w-8 h-8 rounded-full bg-green-500 text-white flex items-center justify-center font-bold">✓</div>
                <div>
                  <div className="text-sm font-bold text-brand-navy">Documentation Complete</div>
                  <div className="text-xs text-slate-500 font-medium">All Set</div>
                </div>
              </motion.div>

              {/* DASHBOARD MOCKUP CONTAINER */}
              <div className="w-full lg:w-[105%] h-full bg-white rounded-[24px] shadow-[0_30px_80px_-20px_rgba(0,0,0,0.15)] border border-slate-100 overflow-hidden flex z-20 relative ml-0 lg:ml-4">
                
                {/* Dashboard Sidebar */}
                <div className="w-[64px] bg-[#111827] flex flex-col items-center py-6 shrink-0 border-r border-slate-800">
                  <div className="text-white text-2xl font-bold mb-10">F</div>
                  <div className="space-y-8 text-slate-500">
                    <div className="w-8 h-8 flex items-center justify-center rounded hover:text-white cursor-pointer"><span className="text-lg">⊞</span></div>
                    <div className="w-8 h-8 flex items-center justify-center rounded hover:text-white cursor-pointer"><span className="text-lg">📄</span></div>
                    <div className="w-8 h-8 flex items-center justify-center rounded hover:text-white cursor-pointer"><span className="text-lg">👤</span></div>
                    <div className="w-8 h-8 flex items-center justify-center rounded hover:text-white cursor-pointer"><span className="text-lg">🛡️</span></div>
                    <div className="w-8 h-8 flex items-center justify-center rounded hover:text-white cursor-pointer"><span className="text-lg">📊</span></div>
                    <div className="w-8 h-8 flex items-center justify-center rounded hover:text-white cursor-pointer"><span className="text-lg">⚙️</span></div>
                  </div>
                </div>

                {/* Dashboard Main Content */}
                <div className="flex-1 flex flex-col bg-white overflow-hidden">
                  
                  {/* Top Status Indicators */}
                  <div className="px-6 py-4 flex items-center gap-4 border-b border-slate-100">
                    <div className="flex items-center gap-2 bg-green-50 px-3 py-1.5 rounded-lg border border-green-100">
                      <div className="w-5 h-5 rounded-full bg-green-500 text-white flex items-center justify-center text-[10px] font-bold">✓</div>
                      <span className="text-xs font-bold text-slate-800">HSN Verified</span>
                    </div>
                    <div className="flex items-center gap-2 bg-green-50 px-3 py-1.5 rounded-lg border border-green-100">
                      <div className="w-5 h-5 rounded-full bg-green-500 text-white flex items-center justify-center text-[10px] font-bold">✓</div>
                      <span className="text-xs font-bold text-slate-800">MSDS Validated</span>
                    </div>
                    <div className="flex items-center gap-2 bg-green-50 px-3 py-1.5 rounded-lg border border-green-100 hidden sm:flex">
                      <div className="w-5 h-5 rounded-full bg-green-500 text-white flex items-center justify-center text-[10px] font-bold">✓</div>
                      <span className="text-xs font-bold text-slate-800">Customer KYC Approved</span>
                    </div>
                    <div className="flex items-center gap-2 bg-green-50 px-3 py-1.5 rounded-lg border border-green-100 hidden md:flex">
                      <div className="w-5 h-5 rounded-full bg-green-500 text-white flex items-center justify-center text-[10px] font-bold">✓</div>
                      <span className="text-xs font-bold text-slate-800">Customs Ready</span>
                    </div>
                  </div>

                  {/* Dashboard Grid */}
                  <div className="flex-1 p-6 grid grid-cols-1 md:grid-cols-3 gap-6 overflow-hidden">
                    
                    {/* Left Column */}
                    <div className="md:col-span-2 space-y-6 flex flex-col">
                      
                      {/* Shipment Overview Box */}
                      <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-sm">
                        <h3 className="text-sm font-bold text-brand-navy mb-4">Shipment Overview</h3>
                        <div className="space-y-3 text-sm">
                          <div className="flex justify-between border-b border-slate-100 pb-2">
                            <span className="text-slate-500">Shipment:</span>
                            <span className="font-semibold text-slate-800">Industrial Chemicals</span>
                          </div>
                          <div className="flex justify-between border-b border-slate-100 pb-2">
                            <span className="text-slate-500">Origin:</span>
                            <span className="font-semibold text-slate-800 flex items-center gap-2">🇮🇳 India</span>
                          </div>
                          <div className="flex justify-between border-b border-slate-100 pb-2">
                            <span className="text-slate-500">Destination:</span>
                            <span className="font-semibold text-slate-800 flex items-center gap-2">🇩🇪 Germany</span>
                          </div>
                          <div className="flex justify-between border-b border-slate-100 pb-2">
                            <span className="text-slate-500">HSN Code:</span>
                            <span className="font-semibold text-green-600">2905.11</span>
                          </div>
                          <div className="flex justify-between border-b border-slate-100 pb-2">
                            <span className="text-slate-500">Duty (Est.):</span>
                            <span className="font-semibold text-slate-800">12%</span>
                          </div>
                          <div className="flex justify-between pt-1 items-center">
                            <span className="text-slate-500">Status:</span>
                            <span className="bg-green-100 text-green-700 px-2 py-0.5 rounded text-xs font-bold border border-green-200 flex items-center gap-1">
                              <span className="w-1.5 h-1.5 rounded-full bg-green-500"></span> Ready For Export
                            </span>
                          </div>
                        </div>
                      </div>

                      {/* Recent Activity Box */}
                      <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-sm flex-1">
                        <h3 className="text-sm font-bold text-brand-navy mb-4">Recent Activity</h3>
                        <div className="space-y-3">
                          <div className="flex items-center justify-between text-xs">
                            <div className="flex items-center gap-2">
                              <div className="w-4 h-4 rounded-full bg-green-500 text-white flex items-center justify-center text-[8px]">✓</div>
                              <span className="text-slate-600">HSN code matched successfully</span>
                            </div>
                            <span className="text-slate-400">2 min ago</span>
                          </div>
                          <div className="flex items-center justify-between text-xs">
                            <div className="flex items-center gap-2">
                              <div className="w-4 h-4 rounded-full bg-green-500 text-white flex items-center justify-center text-[8px]">✓</div>
                              <span className="text-slate-600">MSDS document validated</span>
                            </div>
                            <span className="text-slate-400">5 min ago</span>
                          </div>
                          <div className="flex items-center justify-between text-xs">
                            <div className="flex items-center gap-2">
                              <div className="w-4 h-4 rounded-full bg-green-500 text-white flex items-center justify-center text-[8px]">✓</div>
                              <span className="text-slate-600">Customer KYC verification completed</span>
                            </div>
                            <span className="text-slate-400">8 min ago</span>
                          </div>
                          <div className="flex items-center justify-between text-xs">
                            <div className="flex items-center gap-2">
                              <div className="w-4 h-4 rounded-full bg-green-500 text-white flex items-center justify-center text-[8px]">✓</div>
                              <span className="text-slate-600">Customs duty calculated</span>
                            </div>
                            <span className="text-slate-400">10 min ago</span>
                          </div>
                        </div>
                      </div>

                    </div>

                    {/* Right Column */}
                    <div className="space-y-6 flex flex-col">
                      
                      {/* Compliance Score Box */}
                      <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-sm flex flex-col items-center">
                        <h3 className="text-sm font-bold text-brand-navy w-full text-left mb-4">Compliance Score</h3>
                        <div className="relative w-28 h-28 mb-3">
                          <svg className="w-full h-full -rotate-90 drop-shadow-sm" viewBox="0 0 36 36">
                            <path className="text-slate-100" strokeWidth="4" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                            <path className="text-green-500" strokeWidth="4" strokeDasharray="98, 100" strokeLinecap="round" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                          </svg>
                          <div className="absolute inset-0 flex items-center justify-center flex-col">
                            <span className="text-[28px] font-[800] text-brand-navy">98<span className="text-base text-slate-400">%</span></span>
                          </div>
                        </div>
                        <div className="text-green-600 font-bold text-sm mb-1">Excellent</div>
                        <div className="text-[10px] text-slate-500 text-center">All compliance checks passed</div>
                      </div>

                      {/* Documents Box */}
                      <div className="bg-white rounded-xl border border-slate-200 p-5 shadow-sm flex-1">
                        <h3 className="text-sm font-bold text-brand-navy mb-4">Documents</h3>
                        <div className="space-y-3">
                          <div className="flex items-center justify-between text-xs">
                            <div className="flex items-center gap-2">
                              <span className="text-slate-400">📄</span>
                              <span className="text-slate-700">Commercial Invoice</span>
                            </div>
                            <span className="bg-green-50 text-green-600 px-2 py-0.5 rounded border border-green-100">Verified</span>
                          </div>
                          <div className="flex items-center justify-between text-xs">
                            <div className="flex items-center gap-2">
                              <span className="text-slate-400">📄</span>
                              <span className="text-slate-700">Packing List</span>
                            </div>
                            <span className="bg-green-50 text-green-600 px-2 py-0.5 rounded border border-green-100">Verified</span>
                          </div>
                          <div className="flex items-center justify-between text-xs">
                            <div className="flex items-center gap-2">
                              <span className="text-slate-400">📄</span>
                              <span className="text-slate-700">MSDS Certificate</span>
                            </div>
                            <span className="bg-green-50 text-green-600 px-2 py-0.5 rounded border border-green-100">Verified</span>
                          </div>
                          <div className="flex items-center justify-between text-xs">
                            <div className="flex items-center gap-2">
                              <span className="text-slate-400">📄</span>
                              <span className="text-slate-700">Certificate of Origin</span>
                            </div>
                            <span className="bg-green-50 text-green-600 px-2 py-0.5 rounded border border-green-100">Verified</span>
                          </div>
                        </div>
                      </div>

                    </div>
                  </div>
                </div>
              </div>
            </motion.div>
          </div>
        </div>
      </section>

      {/* ═══ FEATURE STRIP (Below Hero) ═══ */}
      <section className="py-12 bg-white border-b border-ui-border">
        <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-10">
            <div className="flex items-start gap-5 group">
              <div className="w-14 h-14 bg-gradient-to-br from-teal-50 to-blue-50 rounded-full flex items-center justify-center shrink-0 border border-teal-100/50 shadow-sm shadow-teal-100/50 group-hover:scale-105 transition-transform duration-300">
                <CheckCircle2 className="w-6 h-6 text-teal-600" />
              </div>
              <div className="pt-1">
                <h4 className="text-[17px] font-bold text-brand-navy mb-1.5">100% Accurate</h4>
                <p className="text-sm text-slate-500 leading-relaxed pr-4">Real-time validation with global databases.</p>
              </div>
            </div>
            <div className="flex items-start gap-5 group">
              <div className="w-14 h-14 bg-gradient-to-br from-teal-50 to-blue-50 rounded-full flex items-center justify-center shrink-0 border border-teal-100/50 shadow-sm shadow-teal-100/50 group-hover:scale-105 transition-transform duration-300">
                <Zap className="w-6 h-6 text-teal-600" />
              </div>
              <div className="pt-1">
                <h4 className="text-[17px] font-bold text-brand-navy mb-1.5">Instant Results</h4>
                <p className="text-sm text-slate-500 leading-relaxed pr-4">Get compliance results in seconds, not hours.</p>
              </div>
            </div>
            <div className="flex items-start gap-5 group">
              <div className="w-14 h-14 bg-gradient-to-br from-teal-50 to-blue-50 rounded-full flex items-center justify-center shrink-0 border border-teal-100/50 shadow-sm shadow-teal-100/50 group-hover:scale-105 transition-transform duration-300">
                <ShieldCheck className="w-6 h-6 text-teal-600" />
              </div>
              <div className="pt-1">
                <h4 className="text-[17px] font-bold text-brand-navy mb-1.5">Secure & Compliant</h4>
                <p className="text-sm text-slate-500 leading-relaxed pr-4">Enterprise-grade security and data protection.</p>
              </div>
            </div>
            <div className="flex items-start gap-5 group">
              <div className="w-14 h-14 bg-gradient-to-br from-teal-50 to-blue-50 rounded-full flex items-center justify-center shrink-0 border border-teal-100/50 shadow-sm shadow-teal-100/50 group-hover:scale-105 transition-transform duration-300">
                <Globe className="w-6 h-6 text-teal-600" />
              </div>
              <div className="pt-1">
                <h4 className="text-[17px] font-bold text-brand-navy mb-1.5">Global Coverage</h4>
                <p className="text-sm text-slate-500 leading-relaxed pr-4">Supports 220+ countries and global regulations.</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ═══ AUTOMATED COMPLIANCE ENGINE (NEW WORKFLOW) ═══ */}
      <section className="pt-[100px] pb-[80px] bg-white relative overflow-hidden">
        {/* Subtle Grid Background */}
        <div className="absolute inset-0 opacity-[0.02] pointer-events-none bg-[linear-gradient(to_right,#808080_1px,transparent_1px),linear-gradient(to_bottom,#808080_1px,transparent_1px)] bg-[size:40px_40px]"></div>

        {/* GLASSMORPHISM FLOATING BADGES (Moved outside overflow container) */}
        <motion.div animate={{ y: [-4, 4, -4] }} transition={{ duration: 6, repeat: Infinity, ease: "easeInOut" }} className="absolute top-[20%] left-[5%] z-30 bg-white/90 backdrop-blur-md px-3 py-2 rounded-lg shadow-xl border border-slate-100 items-center gap-2 hidden xl:flex">
          <CheckCircle2 className="w-4 h-4 text-green-500" />
          <span className="text-xs font-bold text-brand-navy">GST Verified</span>
        </motion.div>
        <motion.div animate={{ y: [4, -4, 4] }} transition={{ duration: 5, repeat: Infinity, ease: "easeInOut", delay: 1 }} className="absolute top-[35%] left-[10%] z-30 bg-white/90 backdrop-blur-md px-3 py-2 rounded-lg shadow-xl border border-slate-100 items-center gap-2 hidden xl:flex">
          <CheckCircle2 className="w-4 h-4 text-blue-500" />
          <span className="text-xs font-bold text-brand-navy">Customs Rules Checked</span>
        </motion.div>
        <motion.div animate={{ y: [-3, 5, -3] }} transition={{ duration: 5.5, repeat: Infinity, ease: "easeInOut", delay: 2 }} className="absolute top-[25%] right-[8%] z-30 bg-white/90 backdrop-blur-md px-3 py-2 rounded-lg shadow-xl border border-slate-100 items-center gap-2 hidden xl:flex">
          <CheckCircle2 className="w-4 h-4 text-purple-500" />
          <span className="text-xs font-bold text-brand-navy">MSDS Approved</span>
        </motion.div>
        <motion.div animate={{ y: [5, -5, 5] }} transition={{ duration: 6, repeat: Infinity, ease: "easeInOut", delay: 0.5 }} className="absolute top-[40%] right-[5%] z-30 bg-white/90 backdrop-blur-md px-3 py-2 rounded-lg shadow-xl border border-slate-100 items-center gap-2 hidden xl:flex">
          <CheckCircle2 className="w-4 h-4 text-teal-500" />
          <span className="text-xs font-bold text-brand-navy">Duty Estimated</span>
        </motion.div>
        <motion.div animate={{ y: [-5, 5, -5] }} transition={{ duration: 6.5, repeat: Infinity, ease: "easeInOut", delay: 1.5 }} className="absolute top-[10%] right-[15%] z-30 bg-white/90 backdrop-blur-md px-3 py-2 rounded-lg shadow-xl border border-slate-100 items-center gap-2 hidden xl:flex">
          <CheckCircle2 className="w-4 h-4 text-blue-600" />
          <span className="text-xs font-bold text-brand-navy">IEC Active</span>
        </motion.div>
        <motion.div animate={{ y: [3, -3, 3] }} transition={{ duration: 5.2, repeat: Infinity, ease: "easeInOut", delay: 2.5 }} className="absolute top-[55%] left-[2%] z-30 bg-white/90 backdrop-blur-md px-3 py-2 rounded-lg shadow-xl border border-slate-100 items-center gap-2 hidden xl:flex">
          <CheckCircle2 className="w-4 h-4 text-green-600" />
          <span className="text-xs font-bold text-brand-navy">Export Eligible</span>
        </motion.div>

        <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          
          {/* HEADER */}
          <div className="text-center mb-10 max-w-[900px] mx-auto flex flex-col items-center">
            <div className="inline-flex items-center px-4 py-1.5 mb-4 text-[11px] font-bold tracking-widest uppercase rounded-full bg-teal-50 border border-teal-200/50 shadow-sm text-brand-teal">
              Automated Compliance Engine
            </div>
            <h2 className="text-[32px] md:text-[40px] lg:text-[48px] font-[800] text-brand-navy leading-[1.1] tracking-tight mb-4">
              From Shipment Data<br/>
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand-indigo to-brand-teal">To Customs Ready.</span>
            </h2>
            <p className="text-lg lg:text-xl text-slate-500 max-w-[760px] leading-[1.6]">
              Upload shipment details once. Freel automatically validates HSN codes, analyzes MSDS documents, verifies customer credentials, calculates duties, and ensures customs readiness before booking.
            </p>
          </div>

          {/* HORIZONTAL WORKFLOW CONTAINER */}
          <div className="relative w-full mt-10 overflow-x-auto pb-10 hide-scrollbar">
            <div className="min-w-[1200px] w-full bg-white rounded-[32px] border border-slate-100 shadow-[0_20px_60px_-15px_rgba(0,0,0,0.05)] p-6 lg:p-10 relative z-20">
              
              {/* TIMELINE PROGRESS BAR */}
              <div className="relative w-full h-1 bg-slate-100 rounded-full mb-12 mt-4 max-w-[95%] mx-auto">
                <motion.div 
                  initial={{ width: "0%" }}
                  whileInView={{ width: "100%" }}
                  viewport={{ once: true, margin: "-100px" }}
                  transition={{ duration: 4, ease: "easeInOut" }}
                  className="absolute top-0 left-0 h-full bg-gradient-to-r from-teal-400 via-blue-500 to-green-500 rounded-full shadow-[0_0_10px_rgba(45,212,191,0.5)]"
                />
                
                {/* TIMELINE NODES */}
                <div className="absolute top-1/2 -translate-y-1/2 left-0 w-full flex justify-between px-[5%]">
                  {[1, 2, 3, 4, 5, 6].map((step, idx) => (
                    <motion.div 
                      key={step}
                      initial={{ scale: 0, backgroundColor: "#f1f5f9" }}
                      whileInView={{ scale: 1, backgroundColor: step === 6 ? "#22c55e" : "#0ea5e9" }}
                      viewport={{ once: true, margin: "-100px" }}
                      transition={{ delay: idx * 0.7, duration: 0.3 }}
                      className="w-8 h-8 rounded-full border-4 border-white flex items-center justify-center shadow-sm z-10"
                    >
                      <span className="text-[10px] font-bold text-white">0{step}</span>
                    </motion.div>
                  ))}
                </div>
              </div>

              {/* 6-COLUMN HORIZONTAL GRID */}
              <motion.div 
                initial="hidden"
                whileInView="visible"
                viewport={{ once: true, margin: "-100px" }}
                variants={{
                  hidden: { opacity: 0 },
                  visible: { opacity: 1, transition: { staggerChildren: 0.6 } }
                }}
                className="grid grid-cols-6 gap-4 w-full"
              >
                {/* STEP 1: Upload */}
                <motion.div variants={{ hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0 } }} className="bg-slate-50 border border-slate-100 rounded-2xl p-4 flex flex-col h-[280px]">
                  <div className="flex flex-col items-center mb-4 text-center">
                    <div className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 mb-2"><UploadCloud className="w-5 h-5"/></div>
                    <span className="font-bold text-brand-navy text-xs">Upload Shipment</span>
                  </div>
                  <div className="flex-1 bg-white border border-slate-200 rounded-xl p-3 flex flex-col justify-center space-y-2">
                    <div className="flex items-center gap-2 text-[10px] text-slate-600"><FileText className="w-3 h-3 text-slate-400"/> Commercial Invoice.pdf</div>
                    <div className="flex items-center gap-2 text-[10px] text-slate-600"><FileText className="w-3 h-3 text-slate-400"/> Packing List.pdf</div>
                    <div className="flex items-center gap-2 text-[10px] text-slate-600"><FileText className="w-3 h-3 text-slate-400"/> Product Information</div>
                  </div>
                  <div className="mt-3 flex justify-center">
                    <div className="px-2 py-1 bg-blue-50 text-blue-600 text-[9px] font-bold rounded border border-blue-100">Uploaded</div>
                  </div>
                </motion.div>

                {/* STEP 2: HSN */}
                <motion.div variants={{ hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0 } }} className="bg-slate-50 border border-slate-100 rounded-2xl p-4 flex flex-col h-[280px]">
                  <div className="flex flex-col items-center mb-4 text-center">
                    <div className="w-10 h-10 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-600 mb-2"><ScanSearch className="w-5 h-5"/></div>
                    <span className="font-bold text-brand-navy text-xs">HSN Validation</span>
                  </div>
                  <div className="flex-1 bg-white border border-slate-200 rounded-xl p-3 flex flex-col justify-center">
                    <div className="text-lg font-[900] text-brand-navy mb-2 text-center">2905.11</div>
                    <div className="space-y-1.5">
                      <motion.div variants={{ hidden: { opacity: 0, x: -10 }, visible: { opacity: 1, x: 0, transition: { delay: 0.8 } } }} className="flex items-center gap-1.5 text-[9px] text-slate-600"><CheckCircle2 className="w-3 h-3 text-green-500"/> GST Rate Found</motion.div>
                      <motion.div variants={{ hidden: { opacity: 0, x: -10 }, visible: { opacity: 1, x: 0, transition: { delay: 1.0 } } }} className="flex items-center gap-1.5 text-[9px] text-slate-600"><CheckCircle2 className="w-3 h-3 text-green-500"/> Export Rules</motion.div>
                      <motion.div variants={{ hidden: { opacity: 0, x: -10 }, visible: { opacity: 1, x: 0, transition: { delay: 1.2 } } }} className="flex items-center gap-1.5 text-[9px] text-slate-600"><CheckCircle2 className="w-3 h-3 text-green-500"/> Verified</motion.div>
                    </div>
                  </div>
                  <div className="mt-3 flex justify-center">
                    <div className="px-2 py-1 bg-green-50 text-green-600 text-[9px] font-bold rounded border border-green-100 flex items-center gap-1"><CheckCircle2 className="w-3 h-3"/> Matched</div>
                  </div>
                </motion.div>

                {/* STEP 3: MSDS */}
                <motion.div variants={{ hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0 } }} className="bg-slate-50 border border-slate-100 rounded-2xl p-4 flex flex-col h-[280px]">
                  <div className="flex flex-col items-center mb-4 text-center">
                    <div className="w-10 h-10 rounded-full bg-amber-100 flex items-center justify-center text-amber-600 mb-2"><FileText className="w-5 h-5"/></div>
                    <span className="font-bold text-brand-navy text-xs">MSDS Analysis</span>
                  </div>
                  <div className="flex-1 bg-white border border-slate-200 rounded-xl p-3 flex flex-col justify-center">
                    <div className="text-[10px] font-bold text-slate-700 mb-2 text-center truncate">Methanol.pdf</div>
                    <div className="space-y-1.5 text-[9px]">
                      <div className="flex justify-between"><span className="text-slate-500">UN No:</span><span className="font-bold text-slate-800">1230</span></div>
                      <div className="flex justify-between"><span className="text-slate-500">Class:</span><span className="font-bold text-red-500">3 (Flam)</span></div>
                      <div className="flex justify-between"><span className="text-slate-500">Pack Grp:</span><span className="font-bold text-slate-800">II</span></div>
                    </div>
                  </div>
                  <div className="mt-3 flex justify-center">
                    <div className="px-2 py-1 bg-green-50 text-green-600 text-[9px] font-bold rounded border border-green-100 flex items-center gap-1"><CheckCircle2 className="w-3 h-3"/> Validated</div>
                  </div>
                </motion.div>

                {/* STEP 4: KYC */}
                <motion.div variants={{ hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0 } }} className="bg-slate-50 border border-slate-100 rounded-2xl p-4 flex flex-col h-[280px]">
                  <div className="flex flex-col items-center mb-4 text-center">
                    <div className="w-10 h-10 rounded-full bg-purple-100 flex items-center justify-center text-purple-600 mb-2"><UserCheck className="w-5 h-5"/></div>
                    <span className="font-bold text-brand-navy text-xs">Customer KYC</span>
                  </div>
                  <div className="flex-1 bg-white border border-slate-200 rounded-xl p-3 flex flex-col justify-center space-y-2">
                    <div className="text-[9px] flex justify-between border-b border-slate-50 pb-1"><span className="text-slate-500">IEC:</span><span className="font-bold text-slate-800 truncate max-w-[60px]">03120X</span></div>
                    <div className="text-[9px] flex justify-between border-b border-slate-50 pb-1"><span className="text-slate-500">GST:</span><span className="font-bold text-slate-800 truncate max-w-[60px]">27AADC</span></div>
                    <div className="text-[9px] flex justify-between pb-1"><span className="text-slate-500">Entity:</span><span className="font-bold text-slate-800 truncate max-w-[60px]">Bharat</span></div>
                    <div className="flex flex-col gap-1 mt-1 pt-1 border-t border-slate-100">
                       <motion.span variants={{ hidden: { opacity: 0 }, visible: { opacity: 1, transition: { delay: 2.0 } } }} className="text-[8px] bg-green-50 text-green-600 px-1 rounded font-bold text-center">IEC Active</motion.span>
                       <motion.span variants={{ hidden: { opacity: 0 }, visible: { opacity: 1, transition: { delay: 2.2 } } }} className="text-[8px] bg-green-50 text-green-600 px-1 rounded font-bold text-center">GST Valid</motion.span>
                    </div>
                  </div>
                  <div className="mt-3 flex justify-center">
                    <div className="px-2 py-1 bg-green-50 text-green-600 text-[9px] font-bold rounded border border-green-100 flex items-center gap-1"><CheckCircle2 className="w-3 h-3"/> Verified</div>
                  </div>
                </motion.div>

                {/* STEP 5: Duty */}
                <motion.div variants={{ hidden: { opacity: 0, y: 20 }, visible: { opacity: 1, y: 0 } }} className="bg-slate-50 border border-slate-100 rounded-2xl p-4 flex flex-col h-[280px]">
                  <div className="flex flex-col items-center mb-4 text-center">
                    <div className="w-10 h-10 rounded-full bg-teal-100 flex items-center justify-center text-teal-600 mb-2"><Calculator className="w-5 h-5"/></div>
                    <span className="font-bold text-brand-navy text-xs">Duty Calculation</span>
                  </div>
                  <div className="flex-1 bg-white border border-slate-200 rounded-xl p-3 flex flex-col justify-center">
                    <div className="space-y-1.5 text-[9px]">
                      <div className="flex justify-between"><span className="text-slate-500">Value:</span><span className="font-bold text-slate-800">$45K</span></div>
                      <div className="flex justify-between"><span className="text-slate-500">Basic(7.5%):</span><span className="font-bold text-slate-800">$3.3K</span></div>
                      <div className="flex justify-between"><span className="text-slate-500">IGST(18%):</span><span className="font-bold text-slate-800">$8.7K</span></div>
                      <div className="flex justify-between pt-1 border-t border-slate-100 mt-1"><span className="text-slate-700 font-bold">Landed:</span><span className="font-bold text-brand-navy">$57K</span></div>
                    </div>
                  </div>
                  <div className="mt-3 flex justify-center">
                    <div className="px-2 py-1 bg-teal-50 text-teal-600 text-[9px] font-bold rounded border border-teal-100 flex items-center gap-1">Calculated</div>
                  </div>
                </motion.div>

                {/* STEP 6: Export Ready (SUMMARY) */}
                <motion.div 
                  variants={{ hidden: { opacity: 0, scale: 0.95 }, visible: { opacity: 1, scale: 1, transition: { delay: 3.5, duration: 0.5 } } }} 
                  className="bg-gradient-to-br from-green-400 to-teal-500 rounded-2xl p-[3px] shadow-[0_0_30px_rgba(34,197,94,0.4)] h-[280px] scale-105 z-20"
                >
                  <div className="bg-white w-full h-full rounded-[13px] p-4 flex flex-col items-center relative overflow-hidden">
                    <div className="absolute top-0 right-0 w-20 h-20 bg-green-50 rounded-bl-full -mr-6 -mt-6 pointer-events-none"></div>
                    
                    <div className="flex flex-col items-center mb-3 text-center">
                      <CheckCircle className="w-8 h-8 text-green-500 mb-1" />
                      <span className="font-[800] text-sm text-brand-navy">Export Ready</span>
                      <span className="text-[9px] font-semibold text-green-600">✓ Shipment Approved</span>
                    </div>
                    
                    <div className="flex-1 w-full bg-slate-50 rounded-lg p-2 flex flex-col justify-center space-y-1.5">
                      <div className="flex items-center gap-1 text-[9px] text-slate-600"><CheckCircle2 className="w-3 h-3 text-green-500"/> HSN Verified</div>
                      <div className="flex items-center gap-1 text-[9px] text-slate-600"><CheckCircle2 className="w-3 h-3 text-green-500"/> MSDS Approved</div>
                      <div className="flex items-center gap-1 text-[9px] text-slate-600"><CheckCircle2 className="w-3 h-3 text-green-500"/> KYC Approved</div>
                      <div className="flex items-center gap-1 text-[9px] text-slate-600"><CheckCircle2 className="w-3 h-3 text-green-500"/> Duty Calculated</div>
                    </div>

                    <div className="mt-3 flex items-center justify-between w-full px-2">
                      <span className="text-[10px] text-slate-500 font-bold uppercase tracking-wider">Score</span>
                      <div className="relative w-12 h-12">
                        <svg className="w-full h-full -rotate-90" viewBox="0 0 36 36">
                          <path className="text-slate-100" strokeWidth="4" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                          <motion.path 
                            initial={{ strokeDasharray: "0, 100" }}
                            whileInView={{ strokeDasharray: "98, 100" }}
                            transition={{ duration: 1.5, delay: 4.0, ease: "easeOut" }}
                            viewport={{ once: true }}
                            className="text-green-500" strokeWidth="4" strokeLinecap="round" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" 
                          />
                        </svg>
                        <div className="absolute inset-0 flex items-center justify-center">
                          <span className="text-[11px] font-[900] text-brand-navy">98<span className="text-[8px] text-slate-400">%</span></span>
                        </div>
                      </div>
                    </div>
                  </div>
                </motion.div>

              </motion.div>

            </div>
          </div>

          {/* BOTTOM SUMMARY PANEL */}
          <motion.div 
            initial={{ opacity: 0, y: 30 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: 1.5, duration: 0.8 }}
            className="w-full mt-8 bg-white rounded-[24px] border border-slate-200 shadow-xl shadow-slate-200/50 p-8 lg:p-10 flex flex-col lg:flex-row items-center justify-between gap-10 relative z-20"
          >
            {/* Left: Container Image & Shipment Details */}
            <div className="flex-[1.5] w-full flex flex-col sm:flex-row items-center gap-8 border-b lg:border-b-0 lg:border-r border-slate-200 pb-8 lg:pb-0 lg:pr-10 min-w-0">
              <img src="/shipping-container.png" alt="Shipping Container" className="w-48 h-auto object-contain drop-shadow-xl hidden sm:block shrink-0" />
              
              <div className="flex-1 space-y-5 w-full min-w-0">
                <div className="flex justify-between items-center text-sm gap-4">
                  <span className="text-slate-500 font-medium whitespace-nowrap">Shipment</span>
                  <span className="font-bold text-brand-navy text-right">Industrial Chemicals</span>
                </div>
                <div className="flex justify-between items-center text-sm gap-4">
                  <span className="text-slate-500 font-medium whitespace-nowrap">Origin</span>
                  <span className="font-bold text-slate-800 flex items-center justify-end gap-2 whitespace-nowrap"><span>🇮🇳</span> India</span>
                </div>
                <div className="flex justify-between items-center text-sm gap-4">
                  <span className="text-slate-500 font-medium whitespace-nowrap">Destination</span>
                  <span className="font-bold text-slate-800 flex items-center justify-end gap-2 whitespace-nowrap"><span>🇩🇪</span> Germany</span>
                </div>
                <div className="flex justify-between items-center text-sm pt-3 border-t border-slate-100 gap-4">
                  <span className="text-slate-500 font-medium whitespace-nowrap">HSN Code</span>
                  <span className="font-bold text-brand-teal text-base whitespace-nowrap">2905.11</span>
                </div>
              </div>
            </div>

            {/* Center: Compliance Checklist */}
            <div className="flex-[1.2] w-full border-b lg:border-b-0 lg:border-r border-slate-200 pb-8 lg:pb-0 lg:pr-10 lg:pl-4">
              <h4 className="text-xs tracking-wider uppercase font-bold text-slate-400 mb-6">Compliance Checklist</h4>
              <div className="space-y-4">
                <div className="flex items-center gap-3 text-sm font-semibold text-slate-700"><CheckCircle className="w-5 h-5 text-green-500 shrink-0"/> HSN Verified</div>
                <div className="flex items-center gap-3 text-sm font-semibold text-slate-700"><CheckCircle className="w-5 h-5 text-green-500 shrink-0"/> MSDS Approved</div>
                <div className="flex items-center gap-3 text-sm font-semibold text-slate-700"><CheckCircle className="w-5 h-5 text-green-500 shrink-0"/> KYC Approved</div>
                <div className="flex items-center gap-3 text-sm font-semibold text-slate-700"><CheckCircle className="w-5 h-5 text-green-500 shrink-0"/> Duty Calculated</div>
                <div className="flex items-center gap-3 text-sm font-bold text-brand-navy mt-4 pt-4 border-t border-slate-100"><CheckCircle className="w-5 h-5 text-brand-teal shrink-0"/> Customs Ready</div>
              </div>
            </div>

            {/* Right: Compliance Score */}
            <div className="flex-[1.1] w-full flex flex-col sm:flex-row items-center gap-8 lg:pl-4">
              <div className="flex flex-col items-center">
                <h4 className="text-xs tracking-wider uppercase font-bold text-slate-400 mb-5">Compliance Score</h4>
                <div className="relative w-32 h-32 shrink-0">
                  <svg className="w-full h-full -rotate-90 drop-shadow-sm" viewBox="0 0 36 36">
                    <path className="text-slate-100" strokeWidth="3" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                    <motion.path 
                      initial={{ strokeDasharray: "0, 100" }}
                      whileInView={{ strokeDasharray: "98, 100" }}
                      transition={{ duration: 1.5, delay: 2.0, ease: "easeOut" }}
                      viewport={{ once: true }}
                      className="text-green-500 drop-shadow-[0_4px_8px_rgba(34,197,94,0.4)]" strokeWidth="3" strokeLinecap="round" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" 
                    />
                  </svg>
                  <div className="absolute inset-0 flex items-center justify-center">
                    <span className="text-4xl font-[900] text-brand-navy tracking-tighter">98<span className="text-lg text-slate-400 font-bold">%</span></span>
                  </div>
                </div>
              </div>
              <div className="flex flex-col gap-3 mt-4 sm:mt-10">
                <span className="px-3 py-1.5 bg-green-50 text-green-600 text-xs font-bold rounded-full w-fit border border-green-200">Excellent</span>
                <p className="text-sm text-slate-500 leading-relaxed font-medium">All compliance checks passed. Shipment is ready for export.</p>
              </div>
            </div>
          </motion.div>

        </div>

        
        {/* CSS Animation Keyframes for specific custom effects */}
        <style dangerouslySetInnerHTML={{__html: `
          @keyframes scan {
            0% { transform: translateY(0); }
            50% { transform: translateY(60px); }
            100% { transform: translateY(0); }
          }
        `}} />
      </section>

      {/* ═══ COMPLIANCE OPERATIONS CENTER ═══ */}
      <section className="relative py-[120px] bg-white overflow-hidden">
        {/* Subtle Background Grid Pattern */}
        <div className="absolute inset-0 pointer-events-none opacity-[0.4] bg-[linear-gradient(to_right,#e2e8f0_1px,transparent_1px),linear-gradient(to_bottom,#e2e8f0_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_50%,#000_70%,transparent_100%)]"></div>
        
        <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          
          {/* Section Header */}
          <div className="flex flex-col items-center text-center mb-20">
            <div className="inline-flex items-center px-4 py-1.5 rounded-full bg-teal-500/10 border border-teal-500/20 text-brand-teal text-xs font-bold tracking-widest uppercase mb-10">
              Compliance Operations Center
            </div>
            <h2 className="text-4xl md:text-5xl font-bold text-brand-navy leading-tight tracking-tight max-w-[700px] mb-6">
              Everything Compliance.<br/>One Workspace.
            </h2>
            <p className="text-[20px] md:text-[22px] text-slate-500 leading-[1.7] max-w-[720px]">
              Manage HSN validation, MSDS analysis, customer verification, duty calculations, and audit trails from a single platform.
            </p>
          </div>

          {/* Main 2-Column Layout */}
          <div className="flex flex-col xl:flex-row items-center gap-20">
            
            {/* LEFT: HERO DASHBOARD */}
            <motion.div 
              initial={{ opacity: 0, y: 40 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.8 }}
              className="w-full xl:w-[60%] relative"
            >
              {/* Floating Glassmorphism Badges */}
              <motion.div animate={{ y: [-4, 4, -4] }} transition={{ duration: 7, repeat: Infinity, ease: "easeInOut" }} className="absolute -top-6 -left-6 z-30 bg-white/90 backdrop-blur-md px-3 py-2 rounded-lg shadow-[0_10px_30px_rgba(15,23,42,0.08)] border border-white/50 flex items-center gap-2 hidden lg:flex">
                <CheckCircle2 className="w-4 h-4 text-green-500" />
                <span className="text-xs font-bold text-slate-700">GST Verified</span>
              </motion.div>
              <motion.div animate={{ y: [4, -4, 4] }} transition={{ duration: 6, repeat: Infinity, ease: "easeInOut", delay: 1 }} className="absolute top-[20%] -right-8 z-30 bg-white/90 backdrop-blur-md px-3 py-2 rounded-lg shadow-[0_10px_30px_rgba(15,23,42,0.08)] border border-white/50 flex items-center gap-2 hidden lg:flex">
                <CheckCircle2 className="w-4 h-4 text-blue-500" />
                <span className="text-xs font-bold text-slate-700">IEC Active</span>
              </motion.div>
              <motion.div animate={{ y: [-3, 5, -3] }} transition={{ duration: 8, repeat: Infinity, ease: "easeInOut", delay: 2 }} className="absolute bottom-[25%] -left-8 z-30 bg-white/90 backdrop-blur-md px-3 py-2 rounded-lg shadow-[0_10px_30px_rgba(15,23,42,0.08)] border border-white/50 flex items-center gap-2 hidden lg:flex">
                <CheckCircle2 className="w-4 h-4 text-purple-500" />
                <span className="text-xs font-bold text-slate-700">MSDS Approved</span>
              </motion.div>
              <motion.div animate={{ y: [5, -5, 5] }} transition={{ duration: 7.5, repeat: Infinity, ease: "easeInOut", delay: 3 }} className="absolute -bottom-6 right-[10%] z-30 bg-white/90 backdrop-blur-md px-3 py-2 rounded-lg shadow-[0_10px_30px_rgba(15,23,42,0.08)] border border-white/50 flex items-center gap-2 hidden lg:flex">
                <CheckCircle2 className="w-4 h-4 text-teal-500" />
                <span className="text-xs font-bold text-slate-700">Customs Ready</span>
              </motion.div>

              {/* Dashboard Container */}
              <div className="bg-white border border-slate-200 rounded-[28px] shadow-[0_20px_60px_rgba(15,23,42,0.08)] overflow-hidden min-h-[550px] flex flex-col relative z-20">
                
                {/* Dashboard Header */}
                <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-lg bg-brand-navy flex items-center justify-center shadow-inner">
                      <ShieldCheck className="w-4 h-4 text-white" />
                    </div>
                    <span className="font-bold text-brand-navy text-sm">Compliance Intelligence</span>
                  </div>
                  <div className="flex items-center gap-2 bg-white px-3 py-1.5 rounded-full border border-slate-200 shadow-sm">
                    <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div>
                    <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Live Status</span>
                    <span className="text-[10px] text-slate-400 ml-1">Updated 12 sec ago</span>
                  </div>
                </div>

                <div className="p-6 flex-1 flex flex-col bg-[#F8FAFC]/50">
                  {/* KPI Strip */}
                  <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
                    {[
                      { label: 'Verified Shipments', value: '846', trend: '+12%', color: 'text-green-500', bg: 'bg-green-50' },
                      { label: 'Compliance Score', value: '99.2%', trend: '+0.4%', color: 'text-teal-500', bg: 'bg-teal-50' },
                      { label: 'Pending Reviews', value: '14', trend: 'Needs Action', color: 'text-orange-500', bg: 'bg-orange-50' },
                      { label: 'Active Alerts', value: '3', trend: 'Critical', color: 'text-red-500', bg: 'bg-red-50' }
                    ].map((kpi, i) => (
                      <motion.div 
                        initial={{ opacity: 0, y: 20 }}
                        whileInView={{ opacity: 1, y: 0 }}
                        viewport={{ once: true }}
                        transition={{ delay: 0.1 * i, duration: 0.5 }}
                        whileHover={{ y: -6, boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.05)' }}
                        key={i} 
                        className="bg-white p-4 rounded-xl border border-slate-200 flex flex-col justify-between transition-all cursor-default shadow-sm"
                      >
                        <span className="text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-2">{kpi.label}</span>
                        <div className="flex items-end justify-between">
                          <span className="text-2xl font-[900] text-slate-800">{kpi.value}</span>
                          <span className={`text-[10px] font-bold px-2 py-0.5 rounded flex items-center gap-1 ${kpi.color} ${kpi.bg}`}>
                            {i < 2 && <TrendingUp className="w-3 h-3" />}
                            {kpi.trend}
                          </span>
                        </div>
                      </motion.div>
                    ))}
                  </div>

                  {/* Main Dashboard Area */}
                  <div className="flex flex-col lg:flex-row gap-6 flex-1">
                    
                    {/* Shipment Table */}
                    <div className="flex-[1.5] bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden flex flex-col">
                      <div className="px-5 py-4 border-b border-slate-100 flex items-center justify-between">
                        <span className="text-xs font-bold text-brand-navy uppercase tracking-wider">Active Shipments</span>
                        <button className="text-[10px] text-brand-teal font-bold hover:underline">View All</button>
                      </div>
                      <div className="overflow-x-auto">
                        <table className="w-full text-left border-collapse">
                          <thead>
                            <tr className="bg-slate-50/80 text-[10px] uppercase tracking-wider text-slate-400 font-bold">
                              <th className="px-5 py-3">Shipment</th>
                              <th className="px-5 py-3">HSN</th>
                              <th className="px-4 py-3 text-center">Docs</th>
                              <th className="px-5 py-3 text-right">Status</th>
                            </tr>
                          </thead>
                          <tbody className="text-xs text-slate-600 font-medium">
                            <tr className="border-b border-slate-50 hover:bg-slate-50 transition-colors group cursor-default">
                              <td className="px-5 py-4 font-bold text-slate-800">Chemical Export</td>
                              <td className="px-5 py-4">2905.11</td>
                              <td className="px-4 py-4 text-center text-green-500 font-bold">✓ ✓</td>
                              <td className="px-5 py-4 text-right">
                                <span className="inline-block px-2.5 py-1 bg-green-50 text-green-600 rounded-md font-bold text-[10px] group-hover:bg-green-100 transition-colors">Approved</span>
                              </td>
                            </tr>
                            <tr className="border-b border-slate-50 hover:bg-slate-50 transition-colors group cursor-default">
                              <td className="px-5 py-4 font-bold text-slate-800">Industrial Equipment</td>
                              <td className="px-5 py-4">8428.90</td>
                              <td className="px-4 py-4 text-center text-orange-500 font-bold">✓ !</td>
                              <td className="px-5 py-4 text-right">
                                <span className="inline-block px-2.5 py-1 bg-orange-50 text-orange-600 rounded-md font-bold text-[10px] group-hover:bg-orange-100 transition-colors">Review</span>
                              </td>
                            </tr>
                            <tr className="border-b border-slate-50 hover:bg-slate-50 transition-colors group cursor-default">
                              <td className="px-5 py-4 font-bold text-slate-800">Food Ingredients</td>
                              <td className="px-5 py-4">2106.90</td>
                              <td className="px-4 py-4 text-center text-green-500 font-bold">✓ ✓</td>
                              <td className="px-5 py-4 text-right">
                                <span className="inline-block px-2.5 py-1 bg-green-50 text-green-600 rounded-md font-bold text-[10px] group-hover:bg-green-100 transition-colors">Approved</span>
                              </td>
                            </tr>
                            <tr className="hover:bg-slate-50 transition-colors group cursor-default">
                              <td className="px-5 py-4 font-bold text-slate-800">Medical Supplies</td>
                              <td className="px-5 py-4 text-red-500 font-bold">Pending</td>
                              <td className="px-4 py-4 text-center text-red-500 font-bold">! !</td>
                              <td className="px-5 py-4 text-right">
                                <span className="inline-block px-2.5 py-1 bg-red-50 text-red-600 rounded-md font-bold text-[10px] group-hover:bg-red-100 transition-colors">Attention</span>
                              </td>
                            </tr>
                          </tbody>
                        </table>
                      </div>
                    </div>

                    {/* Right Panel */}
                    <div className="flex-1 flex flex-col gap-4">
                      {/* Alerts */}
                      <div className="bg-white border border-slate-200 rounded-xl shadow-sm p-4">
                        <div className="flex items-center gap-2 mb-3">
                          <AlertTriangle className="w-4 h-4 text-orange-500" />
                          <span className="text-xs font-bold text-slate-700 uppercase tracking-wider">Active Alerts</span>
                        </div>
                        <ul className="space-y-2">
                          <li className="text-[11px] font-medium text-slate-600 flex items-center gap-2">
                            <div className="w-1.5 h-1.5 rounded-full bg-red-500"></div> Missing MSDS Document
                          </li>
                          <li className="text-[11px] font-medium text-slate-600 flex items-center gap-2">
                            <div className="w-1.5 h-1.5 rounded-full bg-orange-500"></div> HSN Classification Review
                          </li>
                        </ul>
                      </div>

                      {/* Score & Activity */}
                      <div className="bg-white border border-slate-200 rounded-xl shadow-sm p-5 flex-1 flex flex-col items-center justify-center relative overflow-hidden">
                        <span className="absolute top-4 left-4 text-[10px] font-bold text-slate-400 uppercase tracking-wider">Score</span>
                        <div className="relative w-24 h-24 mb-4">
                          <svg className="w-full h-full -rotate-90 drop-shadow-sm" viewBox="0 0 36 36">
                            <path className="text-slate-100" strokeWidth="4" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                            <motion.path 
                              initial={{ strokeDasharray: "0, 100" }}
                              whileInView={{ strokeDasharray: "98, 100" }}
                              transition={{ duration: 2, ease: "easeOut", delay: 0.5 }}
                              viewport={{ once: true }}
                              className="text-brand-teal" strokeWidth="4" strokeLinecap="round" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" 
                            />
                          </svg>
                          <div className="absolute inset-0 flex flex-col items-center justify-center">
                            <span className="text-2xl font-[900] text-slate-800 leading-none">98<span className="text-[10px] text-slate-400">%</span></span>
                          </div>
                        </div>

                        <div className="w-full mt-2 pt-4 border-t border-slate-100">
                          <div className="flex items-center gap-2 mb-3">
                            <Activity className="w-3 h-3 text-brand-teal" />
                            <span className="text-[10px] font-bold text-slate-500 uppercase tracking-wider">Live Activity</span>
                          </div>
                          <div className="space-y-2">
                            <motion.div animate={{ opacity: showActivity >= 1 ? 1 : 0, x: showActivity >= 1 ? 0 : -10 }} transition={{ duration: 0.5 }} className="flex justify-between items-center text-[10px]">
                              <span className="font-semibold text-slate-700 flex items-center gap-1.5"><CheckCircle2 className="w-3 h-3 text-green-500"/> HSN Validated</span>
                              <span className="text-slate-400">2 min</span>
                            </motion.div>
                            <motion.div animate={{ opacity: showActivity >= 2 ? 1 : 0, x: showActivity >= 2 ? 0 : -10 }} transition={{ duration: 0.5 }} className="flex justify-between items-center text-[10px]">
                              <span className="font-semibold text-slate-700 flex items-center gap-1.5"><CheckCircle2 className="w-3 h-3 text-green-500"/> MSDS Approved</span>
                              <span className="text-slate-400">8 min</span>
                            </motion.div>
                            <motion.div animate={{ opacity: showActivity >= 3 ? 1 : 0, x: showActivity >= 3 ? 0 : -10 }} transition={{ duration: 0.5 }} className="flex justify-between items-center text-[10px]">
                              <span className="font-semibold text-slate-700 flex items-center gap-1.5"><CheckCircle2 className="w-3 h-3 text-green-500"/> KYC Verified</span>
                              <span className="text-slate-400">12 min</span>
                            </motion.div>
                          </div>
                        </div>

                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </motion.div>

            {/* RIGHT: CAPABILITY BLOCKS */}
            <div className="w-full xl:w-[40%] flex flex-col gap-10">
              
              {/* Block 1 */}
              <div className="group border-l-2 border-slate-200 pl-8 hover:border-brand-teal transition-all duration-300 hover:translate-x-2">
                <div className="w-10 h-10 rounded-full bg-slate-50 flex items-center justify-center mb-4 text-brand-navy group-hover:bg-teal-50 group-hover:text-brand-teal transition-colors">
                  <ShieldCheck className="w-5 h-5" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 mb-2">Real-Time Validation</h3>
                <p className="text-slate-500 text-base leading-relaxed">
                  Every shipment is checked against customs and export rules before booking.
                </p>
              </div>

              {/* Block 2 */}
              <div className="group border-l-2 border-slate-200 pl-8 hover:border-brand-teal transition-all duration-300 hover:translate-x-2">
                <div className="w-10 h-10 rounded-full bg-slate-50 flex items-center justify-center mb-4 text-brand-navy group-hover:bg-teal-50 group-hover:text-brand-teal transition-colors">
                  <FileSearch className="w-5 h-5" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 mb-2">Document Intelligence</h3>
                <p className="text-slate-500 text-base leading-relaxed">
                  Extract and validate information directly from invoices and MSDS documents.
                </p>
              </div>

              {/* Block 3 */}
              <div className="group border-l-2 border-slate-200 pl-8 hover:border-brand-teal transition-all duration-300 hover:translate-x-2">
                <div className="w-10 h-10 rounded-full bg-slate-50 flex items-center justify-center mb-4 text-brand-navy group-hover:bg-teal-50 group-hover:text-brand-teal transition-colors">
                  <UserCheck className="w-5 h-5" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 mb-2">Customer Verification</h3>
                <p className="text-slate-500 text-base leading-relaxed">
                  Verify GST, IEC, PAN and KYC records automatically.
                </p>
              </div>

              {/* Block 4 */}
              <div className="group border-l-2 border-slate-200 pl-8 hover:border-brand-teal transition-all duration-300 hover:translate-x-2">
                <div className="w-10 h-10 rounded-full bg-slate-50 flex items-center justify-center mb-4 text-brand-navy group-hover:bg-teal-50 group-hover:text-brand-teal transition-colors">
                  <History className="w-5 h-5" />
                </div>
                <h3 className="text-xl font-bold text-slate-900 mb-2">Audit Trail</h3>
                <p className="text-slate-500 text-base leading-relaxed">
                  Every compliance action is logged and available for review.
                </p>
              </div>

            </div>
          </div>
        </div>
      </section>

      {/* ═══ BUSINESS IMPACT (BENTO GRID) ═══ */}
      <section className="relative py-[120px] bg-white overflow-hidden">
        {/* Subtle Background Grid Pattern */}
        <div className="absolute inset-0 pointer-events-none opacity-[0.3] bg-[linear-gradient(to_right,#e2e8f0_1px,transparent_1px),linear-gradient(to_bottom,#e2e8f0_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_50%,#000_70%,transparent_100%)]"></div>
        
        <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          
          {/* Section Header */}
          <div className="flex flex-col items-center text-center mb-20">
            <div className="inline-flex items-center px-4 py-1.5 rounded-full bg-[#14B8A6]/10 border border-[#14B8A6]/15 text-[#14B8A6] text-xs font-bold tracking-widest uppercase mb-[40px]">
              Business Impact
            </div>
            <h2 className="text-4xl md:text-5xl font-bold text-[#0F172A] leading-tight tracking-tight max-w-[850px] mb-[24px]">
              Move Faster Through Customs.<br/>Ship With Confidence.
            </h2>
            <p className="text-[22px] text-[#64748B] leading-[1.7] max-w-[700px]">
              Reduce compliance bottlenecks, prevent shipment delays, and eliminate manual verification work.
            </p>
          </div>

          {/* Bento Grid Layout */}
          <div className="grid grid-cols-1 lg:grid-cols-4 lg:grid-rows-2 gap-6 h-auto lg:h-[420px]">
            
            {/* Left Hero Card (Col 1-2, Row 1-2) */}
            <motion.div 
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              whileHover={{ y: -8, boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.15)' }}
              transition={{ duration: 0.5 }}
              className="lg:col-span-2 lg:row-span-2 bg-white rounded-[32px] border border-[#E2E8F0] shadow-[0_20px_60px_rgba(15,23,42,0.08)] p-10 flex flex-col items-center justify-center relative overflow-hidden group"
            >
              <div className="absolute inset-0 bg-teal-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-700"></div>
              {/* Subtle green/teal glow */}
              <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-64 h-64 bg-teal-400/20 blur-[80px] rounded-full pointer-events-none"></div>
              
              <div className="relative w-48 h-48 mb-8">
                <svg className="w-full h-full -rotate-90" viewBox="0 0 36 36">
                  <path className="text-slate-100" strokeWidth="3" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                  <motion.path 
                    initial={{ strokeDasharray: "0, 100" }}
                    whileInView={{ strokeDasharray: "98, 100" }}
                    transition={{ duration: 2, ease: "easeOut", delay: 0.2 }}
                    viewport={{ once: true }}
                    className="text-[#14B8A6]" strokeWidth="3" strokeLinecap="round" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" 
                  />
                </svg>
                <div className="absolute inset-0 flex flex-col items-center justify-center">
                  <span className="text-[56px] font-[800] text-[#0F172A] leading-none tracking-tighter">
                    <AnimatedNumber value={98} />%
                  </span>
                </div>
              </div>

              <h3 className="text-2xl font-bold text-[#0F172A] mb-2 text-center relative z-10">Compliance Success Rate</h3>
              <p className="text-[#64748B] text-center font-medium relative z-10">Every shipment validated before booking.</p>

              {/* Minimal Illustration */}
              <div className="absolute bottom-6 right-8 flex items-center gap-2 opacity-50">
                <div className="w-2 h-2 rounded-full bg-teal-500"></div>
                <div className="h-[2px] w-8 bg-slate-200"></div>
                <div className="w-2 h-2 rounded-full bg-teal-500"></div>
                <div className="h-[2px] w-8 bg-slate-200"></div>
                <CheckCircle2 className="w-5 h-5 text-teal-500" />
              </div>
            </motion.div>

            {/* Top Right Card (Col 3, Row 1) */}
            <motion.div 
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              whileHover={{ y: -8, boxShadow: '0 20px 40px -10px rgba(0, 0, 0, 0.1)' }}
              transition={{ duration: 0.5, delay: 0.1 }}
              className="bg-white rounded-[32px] border border-[#E2E8F0] shadow-sm p-8 flex flex-col justify-center"
            >
              <Zap className="w-6 h-6 text-[#14B8A6] mb-4" />
              <div className="text-[40px] font-[800] text-[#0F172A] leading-none mb-2 tracking-tighter"><AnimatedNumber value={75} />%</div>
              <h3 className="text-lg font-bold text-[#0F172A] leading-tight mb-1">Faster Document Processing</h3>
              <p className="text-sm text-[#64748B]">Documents verified automatically.</p>
            </motion.div>

            {/* Middle Right Card (Col 4, Row 1) */}
            <motion.div 
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              whileHover={{ y: -8, boxShadow: '0 20px 40px -10px rgba(0, 0, 0, 0.1)' }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="bg-white rounded-[32px] border border-[#E2E8F0] shadow-sm p-8 flex flex-col justify-center"
            >
              <TrendingUp className="w-6 h-6 text-[#14B8A6] mb-4" />
              <div className="text-[40px] font-[800] text-[#0F172A] leading-none mb-2 tracking-tighter"><AnimatedNumber value={3} />x</div>
              <h3 className="text-lg font-bold text-[#0F172A] leading-tight mb-1">Review Capacity</h3>
              <p className="text-sm text-[#64748B]">Process more shipments without increasing workload.</p>
            </motion.div>

            {/* Bottom Right Card (Col 3, Row 2) */}
            <motion.div 
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              whileHover={{ y: -8, boxShadow: '0 20px 40px -10px rgba(0, 0, 0, 0.1)' }}
              transition={{ duration: 0.5, delay: 0.3 }}
              className="bg-white rounded-[32px] border border-[#E2E8F0] shadow-sm p-8 flex flex-col justify-center"
            >
              <ShieldCheck className="w-6 h-6 text-[#14B8A6] mb-4" />
              <div className="text-[40px] font-[800] text-[#0F172A] leading-none mb-2 tracking-tighter">Zero</div>
              <h3 className="text-lg font-bold text-[#0F172A] leading-tight mb-1">Manual Validation</h3>
              <p className="text-sm text-[#64748B]">Compliance checks handled automatically.</p>
            </motion.div>

            {/* Extra Small Card (Col 4, Row 2) */}
            <motion.div 
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              whileHover={{ y: -8, boxShadow: '0 20px 40px -10px rgba(0, 0, 0, 0.1)' }}
              transition={{ duration: 0.5, delay: 0.4 }}
              className="bg-white rounded-[32px] border border-[#E2E8F0] shadow-sm p-8 flex flex-col justify-center relative overflow-hidden group"
            >
              {/* Subtle green glow */}
              <div className="absolute bottom-0 right-0 w-32 h-32 bg-teal-400/10 blur-[40px] rounded-full pointer-events-none group-hover:bg-teal-400/20 transition-colors"></div>
              
              <CheckCircle className="w-6 h-6 text-[#14B8A6] mb-6" />
              <h3 className="text-xl font-bold text-[#0F172A] leading-tight mb-2 relative z-10">Customs Ready</h3>
              <p className="text-sm text-[#64748B] relative z-10">Before shipment booking.</p>
            </motion.div>

          </div>
        </div>
      </section>

      {/* ═══ WHY FREEL (EMBEDDED WORKFLOW) ═══ */}
      <section className="bg-slate-50 py-[120px] overflow-hidden">
        <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
          
          {/* Section Header */}
          <div className="flex flex-col items-center text-center mb-24">
            <div className="inline-flex items-center px-4 py-1.5 rounded-full bg-slate-200/50 border border-slate-300 text-slate-700 text-xs font-bold tracking-widest uppercase mb-[40px]">
              Why Freel
            </div>
            <h2 className="text-4xl md:text-5xl font-bold text-[#0F172A] leading-tight tracking-tight max-w-[900px] mb-[24px]">
              Compliance Built Into Logistics.<br/>Not Added On Later.
            </h2>
            <p className="text-[22px] text-[#64748B] leading-[1.7] max-w-[800px]">
              Most compliance platforms live outside the shipping workflow. Freel validates, verifies, calculates, and clears shipments directly inside the operating system.
            </p>
          </div>

          {/* Main Visual: Horizontal Workflow */}
          <div className="relative flex flex-col xl:flex-row items-center justify-between gap-12 xl:gap-0 max-w-[1100px] mx-auto mb-24">
            
            {/* Connecting Lines (Desktop only, absolute) */}
            <div className="hidden xl:block absolute top-1/2 left-0 right-0 h-[2px] -translate-y-1/2 z-0 pointer-events-none">
              <svg width="100%" height="100%" className="overflow-visible">
                <motion.line 
                  x1="15%" y1="0" x2="85%" y2="0" 
                  stroke="#CBD5E1" strokeWidth="2" strokeDasharray="8 8"
                />
                {/* Animated Data Flow */}
                <motion.line 
                  x1="15%" y1="0" x2="85%" y2="0" 
                  stroke="#14B8A6" strokeWidth="3" strokeDasharray="8 8"
                  animate={{ strokeDashoffset: [0, -100] }}
                  transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
                />
              </svg>
            </div>

            {/* NODE 1: Shipment Created */}
            <motion.div 
              whileHover={{ y: -5 }}
              className="relative z-10 w-full max-w-[320px] xl:w-[280px] bg-white rounded-2xl shadow-[0_10px_40px_-10px_rgba(0,0,0,0.08)] border border-slate-200 p-6 flex flex-col"
            >
              <div className="w-12 h-12 bg-slate-100 rounded-full flex items-center justify-center mb-4">
                <Package className="w-6 h-6 text-slate-600" />
              </div>
              <h3 className="text-lg font-bold text-slate-900 mb-4">Shipment Created</h3>
              <div className="space-y-3 mb-6">
                <div className="flex items-center gap-2">
                  <FileText className="w-4 h-4 text-slate-400" />
                  <span className="text-sm font-medium text-slate-600">Commercial Invoice</span>
                </div>
                <div className="flex items-center gap-2">
                  <FileText className="w-4 h-4 text-slate-400" />
                  <span className="text-sm font-medium text-slate-600">Packing List</span>
                </div>
                <div className="flex items-center gap-2">
                  <FileText className="w-4 h-4 text-slate-400" />
                  <span className="text-sm font-medium text-slate-600">Product Data</span>
                </div>
              </div>
              <div className="mt-auto inline-flex items-center w-max px-3 py-1 rounded-full bg-slate-100 text-slate-600 text-xs font-bold uppercase tracking-wider">
                Received
              </div>
            </motion.div>

            {/* NODE 2: Compliance Engine */}
            <div className="relative z-10 w-[300px] h-[300px] flex items-center justify-center my-12 xl:my-0">
              {/* Outer Orbiting Badges */}
              {[
                { text: "HSN Verified" },
                { text: "MSDS Approved" },
                { text: "Customer Verified" },
                { text: "Duty Calculated" },
                { text: "Rules Passed" }
              ].map((badge, i) => {
                const startRotation = i * 72;
                return (
                  <motion.div key={i} animate={{ rotate: [startRotation, startRotation + 360] }} transition={{ duration: 25, repeat: Infinity, ease: "linear" }} className="absolute w-[380px] h-[380px] pointer-events-none">
                    <motion.div animate={{ rotate: [-startRotation, -(startRotation + 360)] }} transition={{ duration: 25, repeat: Infinity, ease: "linear" }} className="absolute top-0 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white/90 backdrop-blur-md px-3 py-1.5 rounded-lg shadow-lg border border-slate-100 flex items-center gap-2 w-max pointer-events-auto">
                      <CheckCircle2 className="w-4 h-4 text-teal-500" />
                      <span className="text-xs font-bold text-slate-700">{badge.text}</span>
                    </motion.div>
                  </motion.div>
                )
              })}

              {/* Core Engine Orb */}
              <div className="absolute inset-4 bg-gradient-to-br from-teal-400 to-blue-600 rounded-full shadow-[0_0_80px_rgba(20,184,166,0.4)] flex flex-col items-center justify-center text-white z-10">
                <Cpu className="w-10 h-10 mb-2 opacity-90" />
                <span className="font-bold text-lg leading-tight text-center">AI Compliance<br/>Engine</span>
              </div>
              
              {/* Rotating Borders */}
              <motion.div 
                animate={{ rotate: 360 }} 
                transition={{ duration: 10, repeat: Infinity, ease: "linear" }}
                className="absolute inset-0 rounded-full border-2 border-dashed border-teal-400/50"
              />
              <motion.div 
                animate={{ rotate: -360 }} 
                transition={{ duration: 15, repeat: Infinity, ease: "linear" }}
                className="absolute -inset-4 rounded-full border border-blue-400/30"
              />
            </div>

            {/* NODE 3: Customs Ready */}
            <motion.div 
              whileHover={{ y: -5 }}
              className="relative z-10 w-full max-w-[320px] xl:w-[280px] bg-white rounded-2xl shadow-[0_10px_40px_-10px_rgba(20,184,166,0.15)] border border-teal-100 p-6 flex flex-col"
            >
              <div className="w-12 h-12 bg-teal-50 rounded-full flex items-center justify-center mb-4">
                <CheckCircle className="w-6 h-6 text-teal-500" />
              </div>
              <h3 className="text-lg font-bold text-slate-900 mb-4">Customs Ready</h3>
              <div className="space-y-3 mb-6">
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-teal-500" />
                  <span className="text-sm font-medium text-slate-700">Export Ready</span>
                </div>
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-teal-500" />
                  <span className="text-sm font-medium text-slate-700">Compliance Cleared</span>
                </div>
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-teal-500" />
                  <span className="text-sm font-medium text-slate-700">Documents Verified</span>
                </div>
              </div>
              <div className="mt-auto inline-flex items-center w-max px-3 py-1 rounded-full bg-teal-50 text-teal-600 text-xs font-bold uppercase tracking-wider border border-teal-100">
                Cleared
              </div>
            </motion.div>

          </div>

          {/* Bottom Outcome Blocks */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-[1000px] mx-auto border-t border-slate-200 pt-16">
            <div className="flex flex-col items-center text-center">
              <Zap className="w-6 h-6 text-teal-500 mb-4" />
              <h4 className="text-xl font-bold text-slate-900 mb-2">Instant Validation</h4>
              <p className="text-slate-500">Every shipment is checked before booking.</p>
            </div>
            <div className="flex flex-col items-center text-center">
              <Cpu className="w-6 h-6 text-teal-500 mb-4" />
              <h4 className="text-xl font-bold text-slate-900 mb-2">Automated Compliance</h4>
              <p className="text-slate-500">No manual document reviews.</p>
            </div>
            <div className="flex flex-col items-center text-center">
              <ShieldCheck className="w-6 h-6 text-teal-500 mb-4" />
              <h4 className="text-xl font-bold text-slate-900 mb-2">Customs Readiness</h4>
              <p className="text-slate-500">Shipments leave fully verified.</p>
            </div>
          </div>

        </div>
      </section>

      {/* ═══ GRAND FINALE CTA (COMPLIANCE OS) ═══ */}
      <section className="relative min-h-[700px] flex flex-col justify-center py-32 bg-[linear-gradient(135deg,#020617_0%,#071A3D_50%,#0F172A_100%)] overflow-hidden">
        
        {/* Floating Background Elements */}
        <div className="absolute inset-0 pointer-events-none overflow-hidden">
          <motion.div animate={{ y: [0, -30, 0], opacity: [0.03, 0.06, 0.03] }} transition={{ duration: 10, repeat: Infinity, ease: "easeInOut" }} className="absolute top-[20%] left-[10%]">
            <ShieldCheck className="w-32 h-32 text-white" />
          </motion.div>
          <motion.div animate={{ y: [0, 30, 0], opacity: [0.03, 0.08, 0.03] }} transition={{ duration: 12, repeat: Infinity, ease: "easeInOut" }} className="absolute bottom-[20%] right-[10%]">
            <Globe className="w-40 h-40 text-white" />
          </motion.div>
          <motion.div animate={{ y: [-20, 20, -20], opacity: [0.02, 0.05, 0.02] }} transition={{ duration: 15, repeat: Infinity, ease: "easeInOut" }} className="absolute top-[40%] right-[25%]">
            <FileText className="w-24 h-24 text-white" />
          </motion.div>
          <motion.div animate={{ y: [20, -20, 20], opacity: [0.03, 0.06, 0.03] }} transition={{ duration: 14, repeat: Infinity, ease: "easeInOut" }} className="absolute bottom-[30%] left-[25%]">
            <CheckCircle2 className="w-28 h-28 text-white" />
          </motion.div>
        </div>

        <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 relative z-10 w-full">
          
          {/* Section Header */}
          <div className="flex flex-col items-center text-center mb-16">
            <div className="text-[#14B8A6] text-xs font-bold tracking-[0.2em] uppercase mb-8">
              Compliance Operating System
            </div>
            <h2 className="text-[56px] md:text-[72px] font-black text-white leading-[1.05] tracking-tight mb-8">
              Stop Guessing.<br/>Start Shipping.
            </h2>
            <p className="text-[20px] md:text-[24px] text-white/75 leading-[1.7] max-w-[850px]">
              Every shipment should know exactly what regulations apply, what documents are required, what duties will be charged, and whether it is customs ready. Freel makes that happen automatically.
            </p>
          </div>

          {/* Glowing Status Card */}
          <motion.div 
            initial={{ opacity: 0, y: 40 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.8 }}
            className="relative w-full max-w-[700px] mx-auto bg-white/5 backdrop-blur-xl border border-white/10 rounded-3xl p-8 md:p-10 mb-16 shadow-[0_0_100px_rgba(20,184,166,0.15)] overflow-hidden group"
          >
            {/* Subtle glow effect behind the card */}
            <div className="absolute inset-0 bg-teal-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-1000 pointer-events-none"></div>

            {/* Card Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between border-b border-white/10 pb-6 mb-8">
              <div className="mb-4 sm:mb-0">
                <h3 className="text-white/80 font-bold tracking-widest text-xs uppercase mb-1">COMPLIANCE STATUS</h3>
                <p className="text-white font-medium text-lg">Industrial Chemicals</p>
              </div>
              <div className="flex gap-6">
                <div className="flex flex-col">
                  <span className="text-white/40 text-[10px] uppercase tracking-wider mb-1">Origin</span>
                  <span className="text-white font-medium text-sm">India</span>
                </div>
                <div className="w-[1px] bg-white/10"></div>
                <div className="flex flex-col">
                  <span className="text-white/40 text-[10px] uppercase tracking-wider mb-1">Destination</span>
                  <span className="text-white font-medium text-sm">Germany</span>
                </div>
              </div>
            </div>

            {/* Validation Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-y-5 gap-x-8 mb-10">
              {[
                "HSN Verified", 
                "MSDS Approved", 
                "IEC Verified", 
                "GST Verified", 
                "Duty Calculated", 
                "Customs Rules Checked"
              ].map((item, i) => (
                <motion.div 
                  key={item} 
                  initial={{ opacity: 0, x: -10 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: i * 0.1, duration: 0.5 }}
                  className="flex items-center gap-3"
                >
                  <div className="w-6 h-6 rounded-full bg-[#14B8A6]/20 flex items-center justify-center border border-[#14B8A6]/30">
                    <CheckCircle2 className="w-3.5 h-3.5 text-[#14B8A6]" />
                  </div>
                  <span className="text-white/90 font-medium text-sm">{item}</span>
                </motion.div>
              ))}
            </div>

            {/* Footer Status */}
            <div className="bg-[#14B8A6]/10 border border-[#14B8A6]/20 rounded-2xl p-6 flex flex-col sm:flex-row items-center justify-between relative overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-r from-teal-500/0 via-teal-500/5 to-teal-500/0 -translate-x-full animate-[shimmer_3s_infinite]"></div>
              
              <div className="flex items-center gap-4 mb-4 sm:mb-0 relative z-10">
                <div className="relative flex items-center justify-center">
                  <div className="absolute inset-0 bg-[#14B8A6] rounded-full animate-ping opacity-20 duration-1000"></div>
                  <CheckCircle className="w-8 h-8 text-[#14B8A6] relative z-10" />
                </div>
                <span className="text-2xl font-black text-white tracking-tight">CUSTOMS READY</span>
              </div>
              
              <div className="flex flex-col items-center sm:items-end relative z-10">
                <span className="text-3xl font-black text-[#14B8A6]">98%</span>
                <span className="text-[#14B8A6]/70 text-[10px] font-bold uppercase tracking-widest mt-1">Compliance Score</span>
              </div>
            </div>
          </motion.div>

          {/* CTA Buttons */}
          <div className="flex flex-col sm:flex-row justify-center items-center gap-4 mb-16">
            <Link to="/signup" className="w-full sm:w-auto px-8 py-4 bg-gradient-to-r from-teal-500 to-blue-600 text-white text-lg font-bold rounded-full shadow-[0_10px_30px_rgba(20,184,166,0.3)] hover:shadow-[0_15px_40px_rgba(20,184,166,0.5)] transition-all hover:-translate-y-1 text-center">
              Get Started →
            </Link>
            <Link to="/contact" className="w-full sm:w-auto px-8 py-4 bg-white/5 backdrop-blur-sm border border-white/20 text-white text-lg font-bold rounded-full hover:bg-white/10 transition-all text-center">
              Talk To Compliance Expert
            </Link>
          </div>

          {/* Trust Bar */}
          <div className="flex flex-wrap justify-center items-center gap-x-8 gap-y-4 pt-12 border-t border-white/10 max-w-[1000px] mx-auto">
            {[
              "HSN Classification", 
              "MSDS Analysis", 
              "Automated KYC", 
              "Duty Calculation", 
              "Customs Readiness"
            ].map(item => (
              <div key={item} className="flex items-center gap-2">
                <CheckCircle2 className="w-4 h-4 text-[#14B8A6] opacity-90" />
                <span className="text-white/60 text-sm font-medium tracking-wide">{item}</span>
              </div>
            ))}
          </div>

        </div>
      </section>

    </div>
  );
}
