import { useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { CheckCircle2, MapPin, ShieldCheck, Zap, BarChart, Ship, Plane, Truck, Activity, AlertCircle, MessageSquare, Grid, FileText, Phone, AlertTriangle, Eye, Folder, Users, Check, ArrowRight } from 'lucide-react';
import './About.css';

/* ─── Scroll Reveal ─── */
function useReveal() {
  const ref = useRef(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      ([e]) => { if (e.isIntersecting) { el.classList.add('visible'); obs.unobserve(el); } },
      { threshold: 0.15, rootMargin: '0px 0px -50px 0px' }
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

export default function About() {
  return (
    <div className="bg-white min-h-screen overflow-hidden">
      
      {/* ═══ SAAS PRODUCT HERO ═══ */}
      <section className="relative bg-[#F8FAFC] pt-[80px] pb-[60px] overflow-hidden flex items-center" style={{ minHeight: '92vh' }}>

        {/* Subtle dot grid */}
        <div className="absolute inset-0 pointer-events-none" style={{ backgroundImage: 'radial-gradient(#cbd5e1 1px, transparent 1px)', backgroundSize: '40px 40px', opacity: 0.035 }}/>

        {/* Blue glow top-right */}
        <div className="absolute pointer-events-none" style={{ top: '-80px', right: '-120px', width: '600px', height: '600px', background: 'radial-gradient(circle, rgba(37,99,235,0.07) 0%, transparent 70%)', borderRadius: '50%' }}/>
        {/* Teal glow bottom-left */}
        <div className="absolute pointer-events-none" style={{ bottom: '-60px', left: '-80px', width: '500px', height: '500px', background: 'radial-gradient(circle, rgba(20,184,166,0.06) 0%, transparent 70%)', borderRadius: '50%' }}/>

        <div className="max-w-[1400px] mx-auto px-6 lg:px-10 relative z-10 w-full">
          <div className="grid grid-cols-1 xl:grid-cols-[44%_56%] gap-12 xl:gap-0 items-center">

            {/* ── LEFT CONTENT ── */}
            <div className="flex flex-col items-start text-left">

              {/* Eyebrow */}
              <motion.div
                initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.6 }}
                className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-white border border-slate-200 shadow-sm mb-7"
              >
                <div className="w-1.5 h-1.5 rounded-full bg-gradient-to-r from-teal-500 to-blue-600"/>
                <span className="text-[11px] font-bold tracking-[0.15em] uppercase text-slate-600">Our Story</span>
              </motion.div>

              {/* Headline */}
              <motion.h1
                initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.7, delay: 0.1 }}
                className="font-black text-[#0F172A] tracking-tight mb-5"
                style={{ fontSize: 'clamp(2.2rem, 4vw, 3.5rem)', lineHeight: 1.08, maxWidth: '520px' }}
              >
                The Operating System{' '}
                <span style={{ display: 'block' }}>
                  For{' '}
                  <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#14B8A6] to-[#2563EB]">
                    Modern Logistics
                  </span>
                </span>
              </motion.h1>

              {/* Subtext */}
              <motion.p
                initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.7, delay: 0.2 }}
                className="text-[#64748B] mb-8"
                style={{ fontSize: '16px', lineHeight: 1.7, maxWidth: '440px' }}
              >
                Freight operations, compliance, shipment visibility, and rate discovery — unified in one intelligent platform built for logistics teams.
              </motion.p>

              {/* CTAs */}
              <motion.div
                initial={{ opacity: 0, y: 16 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.7, delay: 0.3 }}
                className="flex items-center gap-3 mb-7"
              >
                <Link to="/platform"
                  className="flex items-center justify-center px-7 rounded-full font-bold text-white text-[15px] transition-all hover:-translate-y-0.5"
                  style={{ height: '50px', background: 'linear-gradient(135deg,#14B8A6,#2563EB)', boxShadow: '0 8px 28px rgba(37,99,235,0.28)' }}
                >
                  Explore Platform
                </Link>
                <a href="#video"
                  className="flex items-center justify-center px-7 rounded-full font-bold text-slate-700 text-[15px] bg-white border border-slate-200 transition-all hover:-translate-y-0.5 hover:shadow-md"
                  style={{ height: '50px' }}
                >
                  Watch Story →
                </a>
              </motion.div>

              {/* Trust line */}
              <motion.div
                initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.8, delay: 0.5 }}
                className="flex items-center gap-2"
              >
                <div className="w-2 h-2 rounded-full bg-emerald-500"/>
                <span className="text-[13px] text-slate-500 font-medium">Built by logistics operators who lived the problem.</span>
              </motion.div>
            </div>

            {/* ── RIGHT: Premium Dashboard ── */}
            <motion.div
              initial={{ opacity: 0, x: 40 }} animate={{ opacity: 1, x: 0 }} transition={{ duration: 0.9, delay: 0.25 }}
              className="relative"
              style={{ paddingLeft: '40px' }}
            >
              {/* === Floating chips OUTSIDE card === */}

              {/* Top-left chip */}
              <motion.div animate={{ y: [0, -8, 0] }} transition={{ duration: 3.5, repeat: Infinity, ease: 'easeInOut' }}
                className="absolute z-40 flex items-center gap-2 bg-white rounded-2xl border border-slate-100 px-4 py-2.5"
                style={{ top: '-18px', left: '10px', boxShadow: '0 8px 32px rgba(0,0,0,0.10)' }}
              >
                <div className="w-6 h-6 rounded-lg flex items-center justify-center" style={{ background: 'linear-gradient(135deg,#14B8A6,#0EA5E9)' }}>
                  <CheckCircle2 className="w-3.5 h-3.5 text-white" />
                </div>
                <div>
                  <div className="text-[10px] text-slate-400 font-semibold uppercase tracking-wide leading-none">Customs</div>
                  <div className="text-[12px] font-bold text-slate-800">Cleared ✓</div>
                </div>
              </motion.div>

              {/* Top-right chip */}
              <motion.div animate={{ y: [0, -10, 0] }} transition={{ duration: 4.5, repeat: Infinity, ease: 'easeInOut', delay: 1 }}
                className="absolute z-40 flex items-center gap-2 bg-white rounded-2xl border border-slate-100 px-4 py-2.5"
                style={{ top: '32px', right: '-16px', boxShadow: '0 8px 32px rgba(0,0,0,0.10)' }}
              >
                <div className="w-6 h-6 rounded-lg flex items-center justify-center bg-blue-50">
                  <ShieldCheck className="w-3.5 h-3.5 text-blue-600" />
                </div>
                <div>
                  <div className="text-[10px] text-slate-400 font-semibold uppercase tracking-wide leading-none">KYC</div>
                  <div className="text-[12px] font-bold text-slate-800">Verified</div>
                </div>
              </motion.div>

              {/* Bottom-left chip */}
              <motion.div animate={{ y: [0, -7, 0] }} transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut', delay: 0.8 }}
                className="absolute z-40 flex items-center gap-2 bg-white rounded-2xl border border-slate-100 px-4 py-2.5"
                style={{ bottom: '-16px', left: '24px', boxShadow: '0 8px 32px rgba(0,0,0,0.10)' }}
              >
                <div className="w-6 h-6 rounded-lg flex items-center justify-center" style={{ background: 'linear-gradient(135deg,#8B5CF6,#EC4899)' }}>
                  <Activity className="w-3.5 h-3.5 text-white" />
                </div>
                <div>
                  <div className="text-[10px] text-slate-400 font-semibold uppercase tracking-wide leading-none">Live</div>
                  <div className="text-[12px] font-bold text-slate-800">Tracking Active</div>
                </div>
              </motion.div>

              {/* Bottom-right chip */}
              <motion.div animate={{ y: [0, -9, 0] }} transition={{ duration: 5, repeat: Infinity, ease: 'easeInOut', delay: 2 }}
                className="absolute z-40 flex items-center gap-2 bg-white rounded-2xl border border-slate-100 px-4 py-2.5"
                style={{ bottom: '48px', right: '-20px', boxShadow: '0 8px 32px rgba(0,0,0,0.10)' }}
              >
                <div className="w-6 h-6 rounded-lg flex items-center justify-center bg-amber-50">
                  <BarChart className="w-3.5 h-3.5 text-amber-600" />
                </div>
                <div>
                  <div className="text-[10px] text-slate-400 font-semibold uppercase tracking-wide leading-none">Rates</div>
                  <div className="text-[12px] font-bold text-slate-800">Compared ↗</div>
                </div>
              </motion.div>

              {/* === Main Dashboard Card === */}
              <div className="relative z-20 bg-white flex flex-col overflow-hidden"
                style={{
                  borderRadius: '28px',
                  border: '1px solid #E8EDF5',
                  boxShadow: '0 24px 80px -16px rgba(15,23,42,0.16), 0 4px 24px rgba(15,23,42,0.06)',
                  height: '480px',
                }}
              >
                {/* Header bar */}
                <div className="flex items-center justify-between px-6 border-b border-slate-100" style={{ height: '56px', background: '#FAFBFE' }}>
                  <div className="flex items-center gap-2.5">
                    <div className="w-8 h-8 rounded-xl flex items-center justify-center" style={{ background: 'linear-gradient(135deg,#14B8A6,#2563EB)' }}>
                      <Zap className="w-4 h-4 text-white" />
                    </div>
                    <span className="font-bold text-slate-800 text-[15px] tracking-tight">LogisticsHQ OS</span>
                  </div>
                  <div className="flex items-center gap-1.5 bg-emerald-50 px-3 py-1 rounded-full border border-emerald-100">
                    <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"/>
                    <span className="text-[10px] font-bold text-emerald-700 uppercase tracking-wider">Live</span>
                  </div>
                </div>

                {/* Metrics row */}
                <div className="grid grid-cols-4 gap-3 px-5 pt-4 pb-3">
                  {[
                    { label: 'Active Shipments', val: '127', color: '#0F172A' },
                    { label: 'Compliance', val: 'Verified', color: '#14B8A6' },
                    { label: 'Rate Requests', val: '18', color: '#0F172A' },
                    { label: 'ETA Accuracy', val: '98%', color: '#2563EB' },
                  ].map(({ label, val, color }) => (
                    <div key={label} className="rounded-xl p-3 border border-slate-100" style={{ background: '#F8FAFC' }}>
                      <div className="text-[9px] font-bold uppercase tracking-wider text-slate-400 mb-1 leading-tight">{label}</div>
                      <div className="text-[18px] font-black leading-none" style={{ color }}>{val}</div>
                    </div>
                  ))}
                </div>

                {/* Divider */}
                <div className="mx-5 border-t border-slate-100 mb-3"/>

                {/* Shipments */}
                <div className="flex-1 px-5 flex flex-col gap-1 overflow-hidden">
                  <div className="text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-1">Recent Shipments</div>
                  {[
                    { icon: <Ship className="w-4 h-4 text-blue-500"/>, bg: '#EFF6FF', name: 'ACME Imports', mode: 'Sea Freight', eta: 'ETA Jun 24', badge: 'In Transit', badgeColor: '#3B82F6', badgeBg: '#EFF6FF' },
                    { icon: <Plane className="w-4 h-4 text-indigo-500"/>, bg: '#EEF2FF', name: 'Bharat Chemicals', mode: 'Air Freight', eta: 'ETA Jun 21', badge: 'Customs Cleared', badgeColor: '#14B8A6', badgeBg: '#F0FDFA' },
                    { icon: <Truck className="w-4 h-4 text-slate-500"/>, bg: '#F8FAFC', name: 'Global Retail', mode: 'Road Freight', eta: 'Completed', badge: 'Delivered', badgeColor: '#64748B', badgeBg: '#F1F5F9' },
                  ].map(({ icon, bg, name, mode, eta, badge, badgeColor, badgeBg }) => (
                    <div key={name} className="flex items-center justify-between py-2.5 px-3 rounded-xl transition-colors cursor-default hover:bg-slate-50" style={{ border: '1px solid transparent' }}>
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: bg }}>{icon}</div>
                        <div>
                          <div className="text-[13px] font-bold text-slate-800">{name}</div>
                          <div className="text-[11px] text-slate-400 font-medium">{mode}</div>
                        </div>
                      </div>
                      <div className="flex items-center gap-3">
                        <div className="text-[11px] font-bold text-slate-500">{eta}</div>
                        <div className="text-[11px] font-bold px-3 py-1 rounded-full" style={{ color: badgeColor, background: badgeBg }}>{badge}</div>
                      </div>
                    </div>
                  ))}
                </div>

                {/* Bottom module bar */}
                <div className="grid grid-cols-4 border-t border-slate-100 mt-2">
                  {[
                    { icon: <MapPin className="w-4 h-4"/>, label: 'Tracking' },
                    { icon: <ShieldCheck className="w-4 h-4"/>, label: 'Compliance' },
                    { icon: <Activity className="w-4 h-4"/>, label: 'RFQ Engine' },
                    { icon: <BarChart className="w-4 h-4"/>, label: 'Analytics' },
                  ].map(({ icon, label }) => (
                    <div key={label} className="flex flex-col items-center justify-center py-3 gap-1 text-slate-400 hover:text-teal-600 hover:bg-teal-50/40 cursor-pointer transition-colors">
                      {icon}
                      <span className="text-[9px] font-bold uppercase tracking-wide">{label}</span>
                    </div>
                  ))}
                </div>
              </div>
            </motion.div>

          </div>
        </div>
      </section>
      {/* ═══ INDUSTRY PROBLEM SECTION ═══ */}
      <section className="bg-[#F8FAFC] pt-[80px] pb-[80px]">
        <div className="max-w-[1400px] mx-auto px-6">
          
          {/* Section Header */}
          <div className="text-center mb-[60px] flex flex-col items-center">
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#14B8A6] to-[#2563EB] font-bold uppercase mb-4" style={{ fontSize: '13px', letterSpacing: '2px' }}>
              The Industry We Grew Up In
            </span>
            <h2 className="font-[800] text-[#0F172A] mb-5 tracking-tight" style={{ fontSize: 'clamp(28px, 3vw, 42px)', lineHeight: 1.1 }}>
              Logistics Was Never Broken.<br/>
              The Software Around It Was.
            </h2>
            <p className="text-slate-500" style={{ maxWidth: '580px', fontSize: '16px', lineHeight: 1.65 }}>
              Shipments move globally every day. Information doesn't.
              Teams still rely on WhatsApp groups, phone calls, PDFs, and spreadsheets to move cargo.
            </p>
          </div>

          {/* Cards Container */}
          <div className="flex flex-col gap-8">
            
            {/* Card 1: WhatsApp Chaos - Premium SaaS Storytelling */}
            <motion.div 
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-100px" }}
              transition={{ duration: 0.6 }}
              className="bg-white rounded-[28px] border border-slate-100 shadow-sm overflow-hidden flex flex-col lg:flex-row min-h-[380px] mt-10"
            >
              {/* Left Content */}
              <div className="w-full lg:w-[42%] p-10 lg:p-12 flex flex-col justify-center gap-5">
                
                {/* 01 Badge */}
                <div className="flex items-center gap-2">
                  <div className="w-6 h-6 rounded-full bg-emerald-50 flex items-center justify-center">
                    <div className="w-2 h-2 bg-emerald-400 rounded-sm rotate-45"></div>
                  </div>
                  <span className="text-emerald-600 font-bold text-[13px] tracking-wider uppercase">01</span>
                </div>

                <h3 className="font-[800] text-slate-900 leading-[1.1] text-[38px]">
                  WhatsApp Chaos
                </h3>
                
                <p className="text-slate-500 text-[16px] leading-relaxed max-w-[380px]">
                  Rates, updates, approvals—everything buried in endless group chats. Nothing organized. Nothing tracked.
                </p>

                {/* Compact Stat Pills */}
                <div className="flex flex-col gap-2 pt-1">
                  <div className="flex items-center gap-2.5 w-fit px-4 py-2 rounded-full bg-slate-50 border border-slate-200">
                    <span className="text-base">💬</span>
                    <span className="text-slate-700 text-[13px] font-semibold">200+ messages per shipment</span>
                  </div>
                  <div className="flex items-center gap-2.5 w-fit px-4 py-2 rounded-full bg-slate-50 border border-slate-200">
                    <span className="text-base">👥</span>
                    <span className="text-slate-700 text-[13px] font-semibold">15 vendor groups, zero visibility</span>
                  </div>
                  <div className="flex items-center gap-2.5 w-fit px-4 py-2 rounded-full bg-slate-50 border border-slate-200">
                    <span className="text-base">⏳</span>
                    <span className="text-slate-700 text-[13px] font-semibold">4+ hours lost chasing updates daily</span>
                  </div>
                </div>
              </div>

              {/* Right Image - Horizontal Slanted View */}
              <div className="w-full lg:w-[58%] relative min-h-[320px] lg:min-h-full bg-slate-50 border-l border-slate-100 overflow-hidden">
                <img 
                  src="/images/about/whatsapp-stage.png" 
                  alt="WhatsApp Logistics Chaos" 
                  className="absolute top-1/2 left-1/2 w-[155%] max-w-none h-auto -translate-y-1/2 -translate-x-[38%] -rotate-[10deg] origin-center" 
                  style={{ filter: 'drop-shadow(0 20px 40px rgba(0,0,0,0.12))' }}
                />
              </div>
            </motion.div>

          </div>
        </div>
      </section>

      {/* ═══ SPREADSHEET HELL SECTION (02) ═══ */}
      <section className="bg-[#F8FAFC] pb-[80px]">
        <div className="max-w-[1400px] mx-auto px-6">
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: "-80px" }}
            transition={{ duration: 0.65 }}
            style={{
              background: 'white',
              borderRadius: '36px',
              boxShadow: '0 30px 80px rgba(0,0,0,0.06)',
              border: '1px solid rgba(226,232,240,0.8)',
              minHeight: '440px',
              position: 'relative',
              overflow: 'hidden',
              display: 'grid',
              gridTemplateColumns: '40% 60%',
            }}
          >
            {/* ── LEFT CONTENT ── */}
            <div style={{ padding: '44px 44px', display: 'flex', flexDirection: 'column', justifyContent: 'center', gap: '16px', position: 'relative', zIndex: 10 }}>

              {/* 02 Badge */}
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <div style={{ width: '24px', height: '24px', borderRadius: '50%', background: '#EFF6FF', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <div style={{ width: '8px', height: '8px', background: '#3B82F6', borderRadius: '2px', transform: 'rotate(45deg)' }}></div>
                </div>
                <span style={{ color: '#3B82F6', fontWeight: 700, fontSize: '13px', letterSpacing: '2px', textTransform: 'uppercase' }}>02</span>
              </div>

              {/* Headline */}
              <h3 style={{ fontWeight: 800, color: '#0F172A', lineHeight: 1.05, fontSize: 'clamp(38px, 4vw, 56px)', margin: 0 }}>
                Spreadsheet<br/>Hell
              </h3>

              {/* Body */}
              <p style={{ color: '#64748B', fontSize: '16px', lineHeight: 1.7, maxWidth: '360px', margin: 0 }}>
                Every shipment starts with a spreadsheet. Ops teams email vendors, wait for replies, paste quotes into Excel, then manually compare pricing, transit times, and service quality.<br/><br/>
                <strong style={{ color: '#334155' }}>A process that should take minutes often takes days.</strong>
              </p>

              {/* Pain Metric Pills */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', paddingTop: '4px' }}>
                {[
                  { icon: '📊', text: '500+ quotes compared monthly' },
                  { icon: '⏰', text: '2 days average vendor response time' },
                  { icon: '🗂️', text: '12 browser tabs open daily' },
                ].map(({ icon, text }) => (
                  <div key={text} style={{ display: 'flex', alignItems: 'center', gap: '10px', width: 'fit-content', padding: '8px 16px', borderRadius: '999px', background: '#F8FAFC', border: '1px solid #E2E8F0' }}>
                    <span style={{ fontSize: '14px' }}>{icon}</span>
                    <span style={{ color: '#475569', fontSize: '13px', fontWeight: 600 }}>{text}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* ── RIGHT VISUAL ── */}
            <div style={{ position: 'relative', display: 'flex', alignItems: 'center', justifyContent: 'center', overflow: 'hidden' }}>

              {/* Background grid */}
              <div style={{
                position: 'absolute', inset: 0, zIndex: 0,
                backgroundImage: 'linear-gradient(rgba(0,0,0,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(0,0,0,0.04) 1px, transparent 1px)',
                backgroundSize: '40px 40px',
                opacity: 0.3,
              }}/>

              {/* Blue glow */}
              <div style={{
                position: 'absolute', zIndex: 1,
                width: '500px', height: '500px',
                top: '50%', left: '50%',
                transform: 'translate(-50%, -50%)',
                background: 'radial-gradient(circle, rgba(59,130,246,0.13), transparent 70%)',
                filter: 'blur(50px)',
                pointerEvents: 'none',
              }}/>

              {/* Spreadsheet Stack */}
              <div style={{ position: 'relative', zIndex: 10, width: '520px' }}>

                {/* Back sheet */}
                <div style={{
                  position: 'absolute', inset: 0,
                  background: 'white', borderRadius: '16px',
                  border: '1px solid #E2E8F0',
                  boxShadow: '0 8px 30px rgba(0,0,0,0.06)',
                  transform: 'rotate(-4deg)',
                  opacity: 0.25,
                }}/>

                {/* Middle sheet */}
                <div style={{
                  position: 'absolute', inset: 0,
                  background: 'white', borderRadius: '16px',
                  border: '1px solid #E2E8F0',
                  boxShadow: '0 8px 30px rgba(0,0,0,0.06)',
                  transform: 'rotate(2deg)',
                  opacity: 0.5,
                }}/>

                {/* Main Sheet */}
                <div
                  className="group"
                  style={{
                    position: 'relative',
                    background: 'white', borderRadius: '16px',
                    border: '1px solid #E2E8F0',
                    boxShadow: '0 20px 60px rgba(0,0,0,0.10)',
                    overflow: 'hidden',
                    transition: 'transform 0.3s ease, box-shadow 0.3s ease',
                  }}
                  onMouseEnter={e => { e.currentTarget.style.transform = 'scale(1.02)'; e.currentTarget.style.boxShadow = '0 30px 80px rgba(0,0,0,0.14)'; }}
                  onMouseLeave={e => { e.currentTarget.style.transform = 'scale(1)'; e.currentTarget.style.boxShadow = '0 20px 60px rgba(0,0,0,0.10)'; }}
                >
                  {/* Sheet Title Bar */}
                  <div style={{ background: '#F1F5F9', borderBottom: '1px solid #E2E8F0', padding: '10px 16px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <div style={{ display: 'flex', gap: '5px' }}>
                      <div style={{ width: '10px', height: '10px', borderRadius: '50%', background: '#FC3D37' }}/>
                      <div style={{ width: '10px', height: '10px', borderRadius: '50%', background: '#FEBC2E' }}/>
                      <div style={{ width: '10px', height: '10px', borderRadius: '50%', background: '#28C840' }}/>
                    </div>
                    <span style={{ fontSize: '12px', color: '#64748B', fontWeight: 600, marginLeft: '8px' }}>Freight_Rate_Comparison_May_2024.xlsx</span>
                  </div>

                  {/* Tab bar */}
                  <div style={{ background: '#F8FAFC', borderBottom: '1px solid #E2E8F0', padding: '0 16px', display: 'flex', gap: '0' }}>
                    {['Sheet 1', 'Sheet 2', 'Sheet 3'].map((t, i) => (
                      <div key={t} style={{ padding: '6px 14px', fontSize: '11px', fontWeight: i === 0 ? 700 : 500, color: i === 0 ? '#1E293B' : '#94A3B8', borderBottom: i === 0 ? '2px solid #3B82F6' : 'none', cursor: 'default' }}>{t}</div>
                    ))}
                  </div>

                  {/* Table */}
                  <div style={{ overflowX: 'auto' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px' }}>
                      <thead>
                        <tr style={{ background: '#F1F5F9' }}>
                          {['Vendor', 'Rate (USD)', 'Transit', 'Status', 'Last Updated'].map(h => (
                            <th key={h} style={{ padding: '10px 14px', textAlign: 'left', fontWeight: 700, color: '#475569', fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.5px', borderBottom: '1px solid #E2E8F0', whiteSpace: 'nowrap' }}>{h}</th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {/* Best row */}
                        <tr style={{ background: 'rgba(34,197,94,0.05)', border: '2px solid #22c55e', borderLeft: '3px solid #22c55e', transition: 'all 0.3s ease', cursor: 'default' }}
                          onMouseEnter={e => { e.currentTarget.style.background = '#F0FDF4'; e.currentTarget.style.transform = 'translateX(4px)'; }}
                          onMouseLeave={e => { e.currentTarget.style.background = 'rgba(34,197,94,0.05)'; e.currentTarget.style.transform = 'translateX(0)'; }}>
                          <td style={{ padding: '11px 14px', fontWeight: 700, color: '#166534', display: 'flex', alignItems: 'center', gap: '8px' }}>
                            DHL
                            <span style={{ background: '#22c55e', color: 'white', fontSize: '9px', fontWeight: 800, padding: '2px 6px', borderRadius: '4px', letterSpacing: '0.5px' }}>BEST RATE</span>
                          </td>
                          <td style={{ padding: '11px 14px', fontWeight: 700, color: '#166534' }}>$2,400</td>
                          <td style={{ padding: '11px 14px', color: '#166534' }}>18 Days</td>
                          <td style={{ padding: '11px 14px' }}><span style={{ background: '#DCFCE7', color: '#166534', padding: '3px 10px', borderRadius: '999px', fontSize: '11px', fontWeight: 700 }}>Quoted</span></td>
                          <td style={{ padding: '11px 14px', color: '#64748B', fontSize: '11px' }}>2 hrs ago</td>
                        </tr>
                        {/* Normal row */}
                        {[
                          { vendor: 'Maersk', rate: '$2,550', transit: '20 Days', status: 'Quoted', statusColor: '#DCFCE7', statusText: '#166534', updated: '3 hrs ago', rowBg: 'white' },
                        ].map(r => (
                          <tr key={r.vendor} style={{ background: r.rowBg, transition: 'all 0.3s ease', cursor: 'default', borderBottom: '1px solid #F1F5F9' }}
                            onMouseEnter={e => { e.currentTarget.style.background = '#F8FAFC'; e.currentTarget.style.transform = 'translateX(4px)'; }}
                            onMouseLeave={e => { e.currentTarget.style.background = r.rowBg; e.currentTarget.style.transform = 'translateX(0)'; }}>
                            <td style={{ padding: '11px 14px', fontWeight: 600, color: '#1E293B' }}>{r.vendor}</td>
                            <td style={{ padding: '11px 14px', fontWeight: 600, color: '#1E293B' }}>{r.rate}</td>
                            <td style={{ padding: '11px 14px', color: '#475569' }}>{r.transit}</td>
                            <td style={{ padding: '11px 14px' }}><span style={{ background: r.statusColor, color: r.statusText, padding: '3px 10px', borderRadius: '999px', fontSize: '11px', fontWeight: 700 }}>{r.status}</span></td>
                            <td style={{ padding: '11px 14px', color: '#64748B', fontSize: '11px' }}>{r.updated}</td>
                          </tr>
                        ))}
                        {/* Waiting row */}
                        <tr style={{ background: 'rgba(255,180,0,0.08)', transition: 'all 0.3s ease', cursor: 'default', borderBottom: '1px solid #F1F5F9' }}
                          onMouseEnter={e => { e.currentTarget.style.background = 'rgba(255,180,0,0.15)'; e.currentTarget.style.transform = 'translateX(4px)'; }}
                          onMouseLeave={e => { e.currentTarget.style.background = 'rgba(255,180,0,0.08)'; e.currentTarget.style.transform = 'translateX(0)'; }}>
                          <td style={{ padding: '11px 14px', fontWeight: 600, color: '#1E293B' }}>Vendor A</td>
                          <td style={{ padding: '11px 14px', color: '#B45309', fontStyle: 'italic', fontWeight: 600 }}>Waiting…</td>
                          <td style={{ padding: '11px 14px', color: '#94A3B8' }}>—</td>
                          <td style={{ padding: '11px 14px' }}><span style={{ background: '#FEF3C7', color: '#92400E', padding: '3px 10px', borderRadius: '999px', fontSize: '11px', fontWeight: 700 }}>Pending</span></td>
                          <td style={{ padding: '11px 14px', color: '#64748B', fontSize: '11px' }}>2 days ago</td>
                        </tr>
                        {/* More quoted */}
                        <tr style={{ background: 'white', transition: 'all 0.3s ease', cursor: 'default', borderBottom: '1px solid #F1F5F9' }}
                          onMouseEnter={e => { e.currentTarget.style.background = '#F8FAFC'; e.currentTarget.style.transform = 'translateX(4px)'; }}
                          onMouseLeave={e => { e.currentTarget.style.background = 'white'; e.currentTarget.style.transform = 'translateX(0)'; }}>
                          <td style={{ padding: '11px 14px', fontWeight: 600, color: '#1E293B' }}>Vendor B</td>
                          <td style={{ padding: '11px 14px', fontWeight: 600, color: '#1E293B' }}>$2,750</td>
                          <td style={{ padding: '11px 14px', color: '#475569' }}>24 Days</td>
                          <td style={{ padding: '11px 14px' }}><span style={{ background: '#DCFCE7', color: '#166534', padding: '3px 10px', borderRadius: '999px', fontSize: '11px', fontWeight: 700 }}>Quoted</span></td>
                          <td style={{ padding: '11px 14px', color: '#64748B', fontSize: '11px' }}>5 hrs ago</td>
                        </tr>
                        {/* No reply row */}
                        <tr style={{ background: 'rgba(255,0,0,0.06)', transition: 'all 0.3s ease', cursor: 'default' }}
                          onMouseEnter={e => { e.currentTarget.style.background = 'rgba(255,0,0,0.12)'; e.currentTarget.style.transform = 'translateX(4px)'; }}
                          onMouseLeave={e => { e.currentTarget.style.background = 'rgba(255,0,0,0.06)'; e.currentTarget.style.transform = 'translateX(0)'; }}>
                          <td style={{ padding: '11px 14px', fontWeight: 600, color: '#1E293B' }}>Vendor C</td>
                          <td style={{ padding: '11px 14px', color: '#DC2626', fontWeight: 700 }}>No Reply</td>
                          <td style={{ padding: '11px 14px', color: '#94A3B8' }}>—</td>
                          <td style={{ padding: '11px 14px' }}><span style={{ background: '#FEE2E2', color: '#991B1B', padding: '3px 10px', borderRadius: '999px', fontSize: '11px', fontWeight: 700 }}>Waiting</span></td>
                          <td style={{ padding: '11px 14px', color: '#DC2626', fontSize: '11px', fontWeight: 600 }}>3 days ago</td>
                        </tr>
                      </tbody>
                    </table>
                  </div>

                  {/* Footer bar */}
                  <div style={{ padding: '8px 14px', background: '#F8FAFC', borderTop: '1px solid #E2E8F0', display: 'flex', alignItems: 'center', gap: '16px' }}>
                    <span style={{ fontSize: '11px', color: '#94A3B8' }}>5 vendors contacted</span>
                    <span style={{ fontSize: '11px', color: '#94A3B8' }}>•</span>
                    <span style={{ fontSize: '11px', color: '#F59E0B', fontWeight: 600 }}>2 still pending</span>
                    <span style={{ fontSize: '11px', color: '#94A3B8' }}>•</span>
                    <span style={{ fontSize: '11px', color: '#DC2626', fontWeight: 600 }}>1 no reply</span>
                  </div>
                </div>
              </div>

              {/* ── Floating Alert Cards ── */}

              {/* Card 1: Top Left */}
              <motion.div
                animate={{ y: [-10, 10, -10] }}
                transition={{ duration: 6, repeat: Infinity, ease: 'easeInOut' }}
                style={{ position: 'absolute', top: '10%', left: '4%', zIndex: 20, background: 'rgba(255,255,255,0.96)', backdropFilter: 'blur(12px)', borderRadius: '18px', padding: '14px 18px', boxShadow: '0 20px 50px rgba(0,0,0,0.08)', border: '1px solid #F1F5F9', cursor: 'default', transition: 'transform 0.35s ease' }}
                onMouseEnter={e => e.currentTarget.style.transform = 'translateY(-6px) scale(1.03)'}
                onMouseLeave={e => e.currentTarget.style.transform = ''}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
                  <div style={{ width: '8px', height: '8px', borderRadius: '50%', background: '#3B82F6', boxShadow: '0 0 8px rgba(59,130,246,0.5)' }}/>
                  <span style={{ fontWeight: 700, color: '#0F172A', fontSize: '13px' }}>14 Emails Sent</span>
                </div>
                <div style={{ fontSize: '11px', color: '#64748B', marginLeft: '16px' }}>Awaiting 3 vendor replies</div>
              </motion.div>

              {/* Card 2: Top Right */}
              <motion.div
                animate={{ y: [12, -12, 12] }}
                transition={{ duration: 8, repeat: Infinity, ease: 'easeInOut' }}
                style={{ position: 'absolute', top: '12%', right: '2%', zIndex: 20, background: 'rgba(255,255,255,0.96)', backdropFilter: 'blur(12px)', borderRadius: '18px', padding: '14px 18px', boxShadow: '0 20px 50px rgba(0,0,0,0.08)', border: '1px solid #F1F5F9', cursor: 'default', transition: 'transform 0.35s ease' }}
                onMouseEnter={e => e.currentTarget.style.transform = 'translateY(-6px) scale(1.03)'}
                onMouseLeave={e => e.currentTarget.style.transform = ''}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
                  <div style={{ width: '8px', height: '8px', borderRadius: '50%', background: '#F59E0B', boxShadow: '0 0 8px rgba(245,158,11,0.5)' }}/>
                  <span style={{ fontWeight: 700, color: '#0F172A', fontSize: '13px' }}>Still Waiting</span>
                </div>
                <div style={{ fontSize: '11px', color: '#64748B', marginLeft: '16px' }}>3 vendors haven't responded</div>
              </motion.div>

              {/* Card 3: Middle Left */}
              <motion.div
                animate={{ y: [-12, 12, -12] }}
                transition={{ duration: 10, repeat: Infinity, ease: 'easeInOut' }}
                style={{ position: 'absolute', top: '44%', left: '2%', zIndex: 20, background: 'rgba(255,255,255,0.96)', backdropFilter: 'blur(12px)', borderRadius: '18px', padding: '14px 18px', boxShadow: '0 20px 50px rgba(0,0,0,0.08)', border: '1px solid #F1F5F9', cursor: 'default', transition: 'transform 0.35s ease' }}
                onMouseEnter={e => e.currentTarget.style.transform = 'translateY(-6px) scale(1.03)'}
                onMouseLeave={e => e.currentTarget.style.transform = ''}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
                  <div style={{ width: '8px', height: '8px', borderRadius: '50%', background: '#8B5CF6', boxShadow: '0 0 8px rgba(139,92,246,0.5)' }}/>
                  <span style={{ fontWeight: 700, color: '#0F172A', fontSize: '13px' }}>Quote Updated Again</span>
                </div>
                <div style={{ fontSize: '11px', color: '#64748B', marginLeft: '16px' }}>Maersk revised pricing</div>
              </motion.div>

              {/* Card 4: Bottom Right */}
              <motion.div
                animate={{ y: [10, -10, 10] }}
                transition={{ duration: 12, repeat: Infinity, ease: 'easeInOut' }}
                style={{ position: 'absolute', bottom: '14%', right: '2%', zIndex: 20, background: 'rgba(255,255,255,0.96)', backdropFilter: 'blur(12px)', borderRadius: '18px', padding: '14px 18px', boxShadow: '0 20px 50px rgba(0,0,0,0.08)', border: '1px solid #FEE2E2', cursor: 'default', transition: 'transform 0.35s ease' }}
                onMouseEnter={e => e.currentTarget.style.transform = 'translateY(-6px) scale(1.03)'}
                onMouseLeave={e => e.currentTarget.style.transform = ''}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
                  <div style={{ width: '8px', height: '8px', borderRadius: '50%', background: '#DC2626', boxShadow: '0 0 8px rgba(220,38,38,0.5)' }}/>
                  <span style={{ fontWeight: 700, color: '#DC2626', fontSize: '13px' }}>Rate Expires in 6hrs</span>
                </div>
                <div style={{ fontSize: '11px', color: '#64748B', marginLeft: '16px' }}>DHL quote validity ending</div>
              </motion.div>

              {/* Card 5: Bottom Left */}
              <motion.div
                animate={{ y: [-8, 8, -8] }}
                transition={{ duration: 7, repeat: Infinity, ease: 'easeInOut' }}
                style={{ position: 'absolute', bottom: '10%', left: '4%', zIndex: 20, background: 'rgba(255,255,255,0.96)', backdropFilter: 'blur(12px)', borderRadius: '18px', padding: '14px 18px', boxShadow: '0 20px 50px rgba(0,0,0,0.08)', border: '1px solid #F1F5F9', cursor: 'default', transition: 'transform 0.35s ease' }}
                onMouseEnter={e => e.currentTarget.style.transform = 'translateY(-6px) scale(1.03)'}
                onMouseLeave={e => e.currentTarget.style.transform = ''}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
                  <div style={{ width: '8px', height: '8px', borderRadius: '50%', background: '#6B7280', boxShadow: '0 0 8px rgba(107,114,128,0.4)' }}/>
                  <span style={{ fontWeight: 700, color: '#0F172A', fontSize: '13px' }}>Manual Comparison</span>
                </div>
                <div style={{ fontSize: '11px', color: '#64748B', marginLeft: '16px' }}>Copy-pasting into Excel</div>
              </motion.div>

            </div>
          </motion.div>
        </div>
      </section>

      {/* ═══ ENDLESS FOLLOW-UPS SECTION (03) ═══ */}
      <section className="bg-[#F8FAFC] pb-[80px]">
        <div className="max-w-[1400px] mx-auto px-6">
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: "-80px" }}
            transition={{ duration: 0.65 }}
            style={{
              background: 'white',
              borderRadius: '36px',
              boxShadow: '0 30px 80px rgba(15,23,42,0.06)',
              border: '1px solid rgba(226,232,240,0.8)',
              overflow: 'hidden',
              display: 'flex',
              alignItems: 'center',
            }}
          >
            {/* ── LEFT CONTENT (40%) ── */}
            <div style={{ width: '40%', flexShrink: 0, padding: '44px 44px', display: 'flex', flexDirection: 'column', justifyContent: 'center', gap: '18px' }}>

              {/* 03 Badge */}
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <div style={{ width: '24px', height: '24px', borderRadius: '50%', background: '#FFF7ED', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <div style={{ width: '8px', height: '8px', background: '#F97316', borderRadius: '2px', transform: 'rotate(45deg)' }}></div>
                </div>
                <span style={{ color: '#F97316', fontWeight: 700, fontSize: '13px', letterSpacing: '2px', textTransform: 'uppercase' }}>03</span>
              </div>

              {/* Headline */}
              <h3 style={{ fontWeight: 800, color: '#0F172A', lineHeight: 1.05, fontSize: 'clamp(38px, 4vw, 54px)', margin: 0 }}>
                Endless<br/>Follow-Ups
              </h3>

              {/* Body */}
              <div style={{ color: '#64748B', fontSize: '15px', lineHeight: 1.75, maxWidth: '360px' }}>
                <p style={{ marginBottom: '14px', color: '#475569' }}>Every shipment update starts with a phone call.</p>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '5px', paddingLeft: '14px', borderLeft: '2px solid #E2E8F0', marginBottom: '14px' }}>
                  {[
                    'Customers call operations.',
                    'Operations call carriers.',
                    'Carriers call drivers.',
                    'Drivers call warehouses.',
                    'Warehouses call customs.',
                  ].map(line => (
                    <span key={line} style={{ color: '#94A3B8', fontSize: '14px', lineHeight: 1.6 }}>{line}</span>
                  ))}
                </div>
                <p style={{ fontWeight: 600, color: '#1E293B', fontSize: '15px' }}>
                  Everyone is chasing information.<br/>
                  <span style={{ color: '#DC2626' }}>Nobody has visibility.</span>
                </p>
              </div>

              {/* Metric Pills */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', paddingTop: '4px' }}>
                {[
                  { icon: '📞', text: '25+ Status Requests Per Shipment' },
                  { icon: '👥', text: '7 Stakeholders Involved' },
                  { icon: '⏱️', text: '4 Hours Lost Daily' },
                ].map(({ icon, text }) => (
                  <div key={text} style={{
                    display: 'flex', alignItems: 'center', gap: '12px',
                    width: 'fit-content', height: '48px',
                    padding: '0 20px', borderRadius: '999px',
                    background: 'white', border: '1px solid #EAEAEA',
                    boxShadow: '0 2px 8px rgba(0,0,0,0.04)',
                  }}>
                    <span style={{ fontSize: '16px' }}>{icon}</span>
                    <span style={{ color: '#334155', fontSize: '13px', fontWeight: 600 }}>{text}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* ── RIGHT SIDE: VIDEO (60%) ── */}
            <motion.div
              initial={{ scale: 0.96, opacity: 0 }}
              whileInView={{ scale: 1, opacity: 1 }}
              viewport={{ once: true, margin: "-80px" }}
              transition={{ duration: 0.8, ease: 'easeOut' }}
              style={{ flex: 1, padding: '24px 24px 24px 0' }}
            >
              <div style={{
                borderRadius: '24px',
                overflow: 'hidden',
                boxShadow: '0 30px 80px rgba(15,23,42,0.08)',
                background: '#0F172A',
                aspectRatio: '16/9',
              }}>
                <video
                  autoPlay
                  muted
                  loop
                  playsInline
                  style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
                >
                  <source src="/videos/about-us/Shipment_lifecycle_completion_an…_202606191515.mp4" type="video/mp4" />
                </video>
              </div>
            </motion.div>

          </motion.div>
        </div>
      </section>


      {/* ═══ COMPLIANCE NIGHTMARE SECTION (04) ═══ */}
      <section className="bg-[#F8FAFC] pb-[80px]">
        <div className="max-w-[1400px] mx-auto px-6">
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: "-80px" }}
            transition={{ duration: 0.65 }}
            style={{
              background: 'white',
              borderRadius: '36px',
              boxShadow: '0 30px 80px rgba(15,23,42,0.06)',
              border: '1px solid rgba(226,232,240,0.8)',
              overflow: 'hidden',
              display: 'grid',
              gridTemplateColumns: '45% 55%',
              minHeight: '520px',
            }}
          >
            {/* ── LEFT CONTENT ── */}
            <div style={{ padding: '52px 48px', display: 'flex', flexDirection: 'column', justifyContent: 'center', gap: '20px' }}>

              {/* 04 Badge */}
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <div style={{ width: '24px', height: '24px', borderRadius: '50%', background: '#FFF1F2', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <div style={{ width: '8px', height: '8px', background: '#EF4444', borderRadius: '2px', transform: 'rotate(45deg)' }}></div>
                </div>
                <span style={{ color: '#EF4444', fontWeight: 700, fontSize: '13px', letterSpacing: '2px', textTransform: 'uppercase' }}>04</span>
              </div>

              {/* Headline */}
              <h3 style={{ fontWeight: 800, color: '#0F172A', lineHeight: 1.05, fontSize: 'clamp(36px, 3.5vw, 52px)', margin: 0 }}>
                Compliance<br/>Nightmare
              </h3>

              {/* Body */}
              <div style={{ color: '#64748B', fontSize: '15px', lineHeight: 1.75, maxWidth: '400px' }}>
                <p style={{ fontStyle: 'italic', color: '#94A3B8', marginBottom: '10px', fontSize: '14px' }}>
                  One wrong HSN code.<br/>
                  One missing declaration.<br/>
                  One expired certificate.
                </p>
                <p style={{ color: '#475569', marginBottom: '10px' }}>
                  A shipment worth thousands can sit at customs for days while teams scramble to fix paperwork.
                </p>
                <p style={{ color: '#475569' }}>
                  Compliance shouldn't live in spreadsheets, emails, and PDF attachments.
                </p>
              </div>

              {/* Stat Pills */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', paddingTop: '4px' }}>
                {[
                  '30% of shipments require document corrections',
                  'Customs delays can cost thousands per day',
                  'Compliance data scattered across multiple systems',
                ].map((text, i) => (
                  <div key={i} style={{
                    display: 'flex', alignItems: 'center', gap: '10px',
                    width: 'fit-content', height: '44px', padding: '0 16px',
                    borderRadius: '999px', background: '#FFF1F2',
                    border: '1px solid #FECACA',
                  }}>
                    <div style={{ width: '6px', height: '6px', borderRadius: '50%', background: '#EF4444', flexShrink: 0 }}/>
                    <span style={{ color: '#991B1B', fontSize: '13px', fontWeight: 600 }}>{text}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* ── RIGHT VISUAL: Document Review Workspace ── */}
            <div style={{ position: 'relative', overflow: 'hidden', background: '#FAFBFC', borderLeft: '1px solid #F1F5F9' }}>

              {/* Blueprint grid background */}
              <div style={{
                position: 'absolute', inset: 0, zIndex: 0,
                backgroundImage: 'linear-gradient(rgba(59,130,246,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(59,130,246,0.04) 1px, transparent 1px)',
                backgroundSize: '32px 32px',
              }}/>

              {/* Faint stamp watermark */}
              <div style={{
                position: 'absolute', right: '8%', top: '10%', zIndex: 0,
                width: '160px', height: '160px', borderRadius: '50%',
                border: '6px solid rgba(239,68,68,0.07)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                transform: 'rotate(-20deg)',
              }}>
                <div style={{ border: '3px solid rgba(239,68,68,0.07)', borderRadius: '50%', padding: '16px' }}>
                  <span style={{ color: 'rgba(239,68,68,0.08)', fontWeight: 900, fontSize: '11px', letterSpacing: '3px', textTransform: 'uppercase' }}>CUSTOMS</span>
                </div>
              </div>

              {/* ── DOCUMENT CARDS ── */}

              {/* Doc 1: Commercial Invoice — Verified (top-left) */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6, delay: 0.1 }}
                animate={{ y: [-4, 4, -4] }}
                style={{
                  position: 'absolute', top: '8%', left: '5%', zIndex: 5,
                  background: 'white', borderRadius: '14px', padding: '14px 18px',
                  width: '195px',
                  border: '1.5px solid #BBF7D0',
                  boxShadow: '0 8px 24px rgba(0,0,0,0.07)',
                  cursor: 'default',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                  <div>
                    <div style={{ fontSize: '10px', color: '#94A3B8', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: '3px' }}>Commercial Invoice</div>
                    <div style={{ fontSize: '12px', fontWeight: 700, color: '#0F172A' }}>INV-2024-0892</div>
                  </div>
                  <div style={{ background: '#DCFCE7', color: '#166534', fontSize: '10px', fontWeight: 700, padding: '2px 7px', borderRadius: '999px', whiteSpace: 'nowrap' }}>✓ Verified</div>
                </div>
                <div style={{ height: '1px', background: '#F1F5F9', marginBottom: '8px' }}/>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ fontSize: '11px', color: '#94A3B8' }}>Value</span>
                    <span style={{ fontSize: '11px', color: '#1E293B', fontWeight: 600 }}>$48,200</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ fontSize: '11px', color: '#94A3B8' }}>Currency</span>
                    <span style={{ fontSize: '11px', color: '#1E293B', fontWeight: 600 }}>USD</span>
                  </div>
                </div>
              </motion.div>

              {/* Doc 2: Packing List — Verified (top-right area) */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6, delay: 0.2 }}
                animate={{ y: [4, -4, 4] }}
                style={{
                  position: 'absolute', top: '6%', right: '5%', zIndex: 5,
                  background: 'white', borderRadius: '14px', padding: '14px 18px',
                  width: '185px',
                  border: '1.5px solid #BBF7D0',
                  boxShadow: '0 8px 24px rgba(0,0,0,0.07)',
                  cursor: 'default',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                  <div>
                    <div style={{ fontSize: '10px', color: '#94A3B8', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: '3px' }}>Packing List</div>
                    <div style={{ fontSize: '12px', fontWeight: 700, color: '#0F172A' }}>PL-2024-0892</div>
                  </div>
                  <div style={{ background: '#DCFCE7', color: '#166534', fontSize: '10px', fontWeight: 700, padding: '2px 7px', borderRadius: '999px' }}>✓ OK</div>
                </div>
                <div style={{ height: '1px', background: '#F1F5F9', marginBottom: '8px' }}/>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ fontSize: '11px', color: '#94A3B8' }}>Packages</span>
                    <span style={{ fontSize: '11px', color: '#1E293B', fontWeight: 600 }}>24 CTN</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ fontSize: '11px', color: '#94A3B8' }}>Gross Wt</span>
                    <span style={{ fontSize: '11px', color: '#1E293B', fontWeight: 600 }}>480 kg</span>
                  </div>
                </div>
              </motion.div>

              {/* Doc 3: IEC — Warning (left-mid) */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6, delay: 0.3 }}
                animate={{ y: [-6, 6, -6] }}
                style={{
                  position: 'absolute', top: '45%', left: '4%', zIndex: 5,
                  background: 'white', borderRadius: '14px', padding: '14px 18px',
                  width: '190px',
                  border: '1.5px solid #FCA5A5',
                  boxShadow: '0 8px 24px rgba(239,68,68,0.10)',
                  cursor: 'default',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                  <div>
                    <div style={{ fontSize: '10px', color: '#94A3B8', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: '3px' }}>IEC Verification</div>
                    <div style={{ fontSize: '12px', fontWeight: 700, color: '#0F172A' }}>IEC-0810012XX</div>
                  </div>
                  <div style={{ background: '#FEE2E2', color: '#991B1B', fontSize: '10px', fontWeight: 700, padding: '2px 7px', borderRadius: '999px', whiteSpace: 'nowrap' }}>⚠ Failed</div>
                </div>
                <div style={{ height: '1px', background: '#FEE2E2', marginBottom: '8px' }}/>
                <div style={{ fontSize: '11px', color: '#DC2626', fontWeight: 500 }}>IEC validation failed.<br/>Customs hold active.</div>
              </motion.div>

              {/* Doc 4: MSDS Certificate — expired (bottom right) */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.6, delay: 0.35 }}
                animate={{ y: [5, -5, 5] }}
                style={{
                  position: 'absolute', bottom: '8%', right: '4%', zIndex: 5,
                  background: 'white', borderRadius: '14px', padding: '14px 18px',
                  width: '190px',
                  border: '1.5px solid #FDE68A',
                  boxShadow: '0 8px 24px rgba(245,158,11,0.10)',
                  cursor: 'default',
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '8px' }}>
                  <div>
                    <div style={{ fontSize: '10px', color: '#94A3B8', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: '3px' }}>MSDS Certificate</div>
                    <div style={{ fontSize: '12px', fontWeight: 700, color: '#0F172A' }}>MSDS-2024</div>
                  </div>
                  <div style={{ background: '#FEF3C7', color: '#92400E', fontSize: '10px', fontWeight: 700, padding: '2px 7px', borderRadius: '999px', whiteSpace: 'nowrap' }}>⚠ Expired</div>
                </div>
                <div style={{ height: '1px', background: '#FEF3C7', marginBottom: '8px' }}/>
                <div style={{ fontSize: '11px', color: '#B45309', fontWeight: 500 }}>Certificate expired.<br/>Renewal required.</div>
              </motion.div>

              {/* ── CENTRAL DOMINANT WARNING CARD ── */}
              <motion.div
                initial={{ opacity: 0, scale: 0.9 }}
                whileInView={{ opacity: 1, scale: 1 }}
                viewport={{ once: true }}
                transition={{ duration: 0.7, delay: 0.2, ease: 'easeOut' }}
                animate={{ scale: [1, 1.015, 1] }}
                style={{
                  position: 'absolute',
                  top: '50%', left: '50%',
                  transform: 'translate(-50%, -50%) rotate(-2deg)',
                  zIndex: 20,
                  background: 'white',
                  borderRadius: '20px',
                  padding: '22px 26px',
                  width: '250px',
                  border: '2px solid #EF4444',
                  boxShadow: '0 20px 60px rgba(239,68,68,0.18), 0 4px 16px rgba(0,0,0,0.08)',
                  cursor: 'default',
                }}
              >
                {/* Red top bar */}
                <div style={{ background: '#FEF2F2', borderRadius: '10px', padding: '10px 14px', marginBottom: '14px', display: 'flex', alignItems: 'center', gap: '10px' }}>
                  <div style={{ width: '28px', height: '28px', borderRadius: '8px', background: '#EF4444', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                    <span style={{ color: 'white', fontSize: '14px', fontWeight: 900 }}>!</span>
                  </div>
                  <div>
                    <div style={{ fontSize: '9px', color: '#DC2626', fontWeight: 700, letterSpacing: '1.5px', textTransform: 'uppercase' }}>Customs Alert</div>
                    <div style={{ fontSize: '11px', color: '#7F1D1D', fontWeight: 700, marginTop: '1px' }}>Shipment Blocked</div>
                  </div>
                </div>

                <div style={{ fontSize: '15px', fontWeight: 800, color: '#0F172A', letterSpacing: '0.3px', marginBottom: '6px' }}>
                  HSN CODE MISMATCH
                </div>
                <div style={{ fontSize: '12px', color: '#64748B', lineHeight: 1.5, marginBottom: '14px' }}>
                  Shipment cannot proceed until corrected. Contact your compliance team immediately.
                </div>

                <div style={{ height: '1px', background: '#FEE2E2', marginBottom: '12px' }}/>

                {/* Mini checklist */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                  {[
                    { label: 'Commercial Invoice', ok: true },
                    { label: 'HSN Classification', ok: false },
                    { label: 'IEC Certificate', ok: false },
                  ].map(({ label, ok }) => (
                    <div key={label} style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                      <div style={{
                        width: '16px', height: '16px', borderRadius: '50%', flexShrink: 0,
                        background: ok ? '#DCFCE7' : '#FEE2E2',
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                      }}>
                        <span style={{ fontSize: '9px', fontWeight: 800, color: ok ? '#166534' : '#DC2626' }}>{ok ? '✓' : '✕'}</span>
                      </div>
                      <span style={{ fontSize: '11px', color: ok ? '#475569' : '#DC2626', fontWeight: ok ? 400 : 600 }}>{label}</span>
                    </div>
                  ))}
                </div>

                <div style={{ marginTop: '14px', background: '#EF4444', borderRadius: '8px', padding: '8px 12px', textAlign: 'center' }}>
                  <span style={{ color: 'white', fontSize: '11px', fontWeight: 700 }}>Resolve Immediately →</span>
                </div>
              </motion.div>

              {/* Glow behind central card */}
              <div style={{
                position: 'absolute', zIndex: 10,
                width: '320px', height: '320px',
                top: '50%', left: '50%',
                transform: 'translate(-50%, -50%)',
                background: 'radial-gradient(circle, rgba(239,68,68,0.08), transparent 70%)',
                filter: 'blur(40px)',
                pointerEvents: 'none',
              }}/>

            </div>
          </motion.div>
        </div>
      </section>

      {/* ═══ THE BREAKTHROUGH SECTION (05) ═══ */}
      <section className="bg-white py-[160px] relative overflow-hidden">
        
        {/* Background Gradients */}
        <div className="absolute inset-0 pointer-events-none" style={{ backgroundImage: 'radial-gradient(#e2e8f0 1px, transparent 1px)', backgroundSize: '40px 40px', opacity: 0.3 }}/>
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[900px] h-[900px] bg-gradient-to-r from-teal-500/5 to-blue-600/5 rounded-full blur-[100px] pointer-events-none"/>

        <div className="max-w-[1400px] mx-auto px-6 relative z-10 flex flex-col items-center">
          
          {/* Eyebrow */}
          <motion.div
            initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} transition={{ duration: 0.6 }}
            className="text-transparent bg-clip-text bg-gradient-to-r from-[#14B8A6] to-[#2563EB] text-[14px] font-bold tracking-[0.15em] uppercase mb-6"
          >
            The Breakthrough
          </motion.div>

          {/* Headline */}
          <motion.h2
            initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} transition={{ duration: 0.8, delay: 0.1 }}
            className="text-center font-black mb-6 tracking-tight"
            style={{ fontSize: 'clamp(32px, 5vw, 56px)', lineHeight: 1.05 }}
          >
            <div className="text-[#0F172A] mb-1">The Problem Was Never Moving Cargo.</div>
            <div className="text-transparent bg-clip-text bg-gradient-to-r from-[#14B8A6] to-[#2563EB]">It Was Moving Information.</div>
          </motion.h2>

          {/* Subtext */}
          <motion.p
            initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} transition={{ duration: 0.8, delay: 0.2 }}
            className="text-center text-[#64748B] max-w-[700px] mb-[100px]"
            style={{ fontSize: '18px', lineHeight: 1.7 }}
          >
            The real challenge is keeping everyone—from carriers to customers—aligned with accurate data. <strong className="text-slate-800 font-semibold">LogisticsHQ is the operating system connecting them all.</strong>
          </motion.p>

          {/* Ecosystem Visualization */}
          <div className="relative w-full max-w-[1100px] aspect-square md:aspect-[16/10] flex items-center justify-center">
            
            {/* SVG Connections */}
            <svg className="absolute inset-0 w-full h-full pointer-events-none z-0">
              <defs>
                <linearGradient id="lineGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" stopColor="#14B8A6" stopOpacity="0.25" />
                  <stop offset="100%" stopColor="#2563EB" stopOpacity="0.25" />
                </linearGradient>
                <linearGradient id="pulseGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" stopColor="#14B8A6" />
                  <stop offset="100%" stopColor="#2563EB" />
                </linearGradient>
              </defs>

              {[
                {x: '50%', y: '10%'}, {x: '76%', y: '19%'}, {x: '89%', y: '43%'},
                {x: '85%', y: '70%'}, {x: '64%', y: '88%'}, {x: '36%', y: '88%'},
                {x: '15%', y: '70%'}, {x: '11%', y: '43%'}, {x: '24%', y: '19%'}
              ].map((pos, i) => (
                <g key={i}>
                  <line x1="50%" y1="50%" x2={pos.x} y2={pos.y} stroke="url(#lineGrad)" strokeWidth="1.5" />
                  <line x1="50%" y1="50%" x2={pos.x} y2={pos.y} stroke="url(#pulseGrad)" strokeWidth="2.5" strokeDasharray="4 24" className="opacity-80">
                    <animate attributeName="stroke-dashoffset" from="28" to="0" dur={`${1.5 + (i%3)*0.5}s`} repeatCount="indefinite" />
                  </line>
                </g>
              ))}
            </svg>

            {/* Orbiting Nodes Array */}
            {[
              { id: 'customers', label: 'Customers', color: '#0EA5E9', x: '50%', y: '10%' },
              { id: 'operations', label: 'Operations', color: '#6366F1', x: '76%', y: '19%' },
              { id: 'carriers', label: 'Carriers', color: '#8B5CF6', x: '89%', y: '43%' },
              { id: 'warehouses', label: 'Warehouses', color: '#D946EF', x: '85%', y: '70%' },
              { id: 'customs', label: 'Customs', color: '#F43F5E', x: '64%', y: '88%' },
              { id: 'finance', label: 'Finance', color: '#F59E0B', x: '36%', y: '88%' },
              { id: 'compliance', label: 'Compliance', color: '#10B981', x: '15%', y: '70%' },
              { id: 'tracking', label: 'Tracking', color: '#14B8A6', x: '11%', y: '43%' },
              { id: 'rate', label: 'Rate Discovery', color: '#3B82F6', x: '24%', y: '19%' }
            ].map((node, i) => (
              <motion.div
                key={node.id}
                initial={{ opacity: 0, scale: 0.8 }}
                whileInView={{ opacity: 1, scale: 1 }}
                viewport={{ once: true, margin: "-100px" }}
                transition={{ duration: 0.6, delay: 0.3 + i * 0.1, ease: 'easeOut' }}
                className="absolute z-10 flex items-center gap-3 bg-white/95 backdrop-blur-md rounded-full border border-slate-200 shadow-lg px-5 py-3"
                style={{ 
                  left: node.x, top: node.y, 
                  transform: 'translate(-50%, -50%)',
                  boxShadow: '0 12px 32px rgba(15,23,42,0.06)'
                }}
              >
                <div className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: node.color, boxShadow: `0 0 10px ${node.color}` }}></div>
                <div className="text-[13px] font-bold text-slate-700 whitespace-nowrap tracking-wide">{node.label}</div>
              </motion.div>
            ))}

            {/* Micro Animations (Floating Notifications) */}
            {[
              { text: '✓ Customs Cleared', top: '35%', left: '25%', delay: 0 },
              { text: '✓ Invoice Approved', top: '65%', left: '75%', delay: 2 },
              { text: '✓ ETA Updated', top: '25%', left: '65%', delay: 1 },
              { text: '✓ HSN Verified', top: '75%', left: '35%', delay: 3 },
              { text: '✓ Rate Updated', top: '80%', left: '50%', delay: 1.5 },
            ].map((notif, i) => (
              <motion.div
                key={i}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: [0, 1, 1, 0], y: [10, -10, -10, -20] }}
                transition={{ duration: 4, repeat: Infinity, delay: notif.delay, times: [0, 0.2, 0.8, 1] }}
                className="absolute z-20 bg-emerald-50 border border-emerald-100 text-emerald-700 text-[11px] font-bold px-3 py-1.5 rounded-full shadow-sm whitespace-nowrap"
                style={{ top: notif.top, left: notif.left, transform: 'translate(-50%, -50%)' }}
              >
                {notif.text}
              </motion.div>
            ))}

            {/* Center LogisticsHQ Node */}
            <motion.div
              initial={{ opacity: 0, scale: 0.5 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true, margin: "-100px" }}
              transition={{ duration: 1, ease: 'easeOut' }}
              className="absolute z-30 top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full flex flex-col items-center justify-center"
              style={{
                width: '280px', height: '280px',
                background: 'linear-gradient(135deg, #14B8A6 0%, #2563EB 100%)',
                boxShadow: '0 0 100px rgba(37,99,235,0.4), inset 0 0 40px rgba(255,255,255,0.2)',
                border: '1px solid rgba(255,255,255,0.4)'
              }}
            >
              <div className="absolute inset-0 rounded-full animate-pulse opacity-40 pointer-events-none" style={{ border: '2px solid rgba(255,255,255,0.2)', transform: 'scale(1.1)' }}/>
              <div className="absolute inset-0 rounded-full animate-pulse opacity-20 pointer-events-none" style={{ border: '1px solid rgba(255,255,255,0.1)', transform: 'scale(1.25)', animationDelay: '0.5s' }}/>

              <div className="text-white text-[38px] font-black tracking-tight mb-1" style={{ textShadow: '0 4px 20px rgba(0,0,0,0.1)' }}>LogisticsHQ</div>
              <div className="text-teal-100 text-[12px] font-bold uppercase tracking-[0.2em] text-center px-4 leading-relaxed">
                Logistics<br/>Operating System
              </div>
            </motion.div>

          </div>
        </div>
      </section>

      {/* ═══ THE COMPARISON SECTION (06) ═══ */}
      <section className="bg-[#FAFBFC] py-[140px] relative overflow-hidden">
        <div className="max-w-[1500px] mx-auto px-6 relative z-10">
          
          {/* Top Content */}
          <div className="flex flex-col items-center text-center mb-24">
            <motion.div
              initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} transition={{ duration: 0.6 }}
              className="text-transparent bg-clip-text bg-gradient-to-r from-[#14B8A6] to-[#2563EB] text-[14px] font-bold tracking-[0.15em] uppercase mb-6"
            >
              The Impact of Connecting Everything
            </motion.div>

            <motion.h2
              initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} transition={{ duration: 0.8, delay: 0.1 }}
              className="font-black text-[#0F172A] mb-6 tracking-tight"
              style={{ fontSize: 'clamp(32px, 5vw, 56px)', lineHeight: 1.05, maxWidth: '900px' }}
            >
              What Changes When<br/>Companies Switch To <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#14B8A6] to-[#2563EB]">LogisticsHQ</span>?
            </motion.h2>

            <motion.p
              initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} transition={{ duration: 0.8, delay: 0.2 }}
              className="text-[#64748B] max-w-[700px]"
              style={{ fontSize: '20px', lineHeight: 1.8 }}
            >
              From fragmented operations to unified logistics intelligence.
            </motion.p>
          </div>

          {/* Main Comparison Area */}
          <div className="grid grid-cols-1 lg:grid-cols-[1fr_240px_1fr] gap-8 lg:gap-0 items-stretch relative">
            
            {/* Left Column: Before LogisticsHQ */}
            <motion.div
              initial={{ opacity: 0, x: -40 }} whileInView={{ opacity: 1, x: 0 }} viewport={{ once: true }} transition={{ duration: 0.8, delay: 0.3 }}
              className="flex flex-col gap-6 w-full"
            >
              {/* Header */}
              <div className="flex flex-col items-center text-center mb-6">
                <div className="bg-red-500 text-white text-[11px] font-bold tracking-widest uppercase px-5 py-1.5 rounded-full mb-6 shadow-sm">
                  Before LogisticsHQ
                </div>
                <h3 className="text-2xl font-black text-slate-800 tracking-tight">Disconnected. Manual. Error-Prone.</h3>
              </div>

              {/* Cards Container */}
              <div className="flex flex-col gap-4 p-6 lg:p-8 rounded-[32px] border border-red-500/10 bg-white/50 relative overflow-hidden" style={{ boxShadow: '0 20px 40px -20px rgba(239,68,68,0.05)' }}>
                {/* Red subtle gradient bg */}
                <div className="absolute inset-0 bg-gradient-to-br from-red-50/50 to-transparent pointer-events-none"/>

                {/* Left Card 1 */}
                <div className="relative bg-white rounded-2xl p-5 border border-red-500/10 shadow-[0_4px_20px_rgba(239,68,68,0.03)] flex flex-col xl:flex-row items-start xl:items-center gap-5 hover:shadow-[0_8px_30px_rgba(239,68,68,0.08)] transition-all">
                  <div className="w-12 h-12 rounded-xl bg-red-50 flex items-center justify-center shrink-0">
                    <MessageSquare className="w-5 h-5 text-red-500" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-bold text-slate-800 text-[15px] mb-1">WhatsApp Groups</div>
                    <div className="text-slate-500 text-[13px] leading-snug xl:pr-4">Updates buried in endless chats. Nothing is tracked.</div>
                  </div>
                  {/* Mini Visual */}
                  <div className="shrink-0 w-full xl:w-[150px] flex flex-col gap-2 mt-2 xl:mt-0">
                    <div className="flex items-center justify-between bg-slate-50 rounded-lg px-2.5 py-1.5 border border-slate-100">
                      <div className="flex items-center gap-1.5"><div className="w-4 h-4 bg-green-500 rounded-full flex items-center justify-center"><Phone className="w-2 h-2 text-white" /></div><div className="text-[10px] font-semibold text-slate-600">Freight Updates</div></div>
                      <div className="w-4 h-4 bg-red-500 rounded-full flex items-center justify-center text-[8px] font-bold text-white">47</div>
                    </div>
                    <div className="flex items-center justify-between bg-slate-50 rounded-lg px-2.5 py-1.5 border border-slate-100">
                      <div className="flex items-center gap-1.5"><div className="w-4 h-4 bg-green-500 rounded-full flex items-center justify-center"><Phone className="w-2 h-2 text-white" /></div><div className="text-[10px] font-semibold text-slate-600">Mumbai Ops</div></div>
                      <div className="w-4 h-4 bg-red-500 rounded-full flex items-center justify-center text-[8px] font-bold text-white">22</div>
                    </div>
                  </div>
                </div>

                {/* Left Card 2 */}
                <div className="relative bg-white rounded-2xl p-5 border border-red-500/10 shadow-[0_4px_20px_rgba(239,68,68,0.03)] flex flex-col xl:flex-row items-start xl:items-center gap-5 hover:shadow-[0_8px_30px_rgba(239,68,68,0.08)] transition-all">
                  <div className="w-12 h-12 rounded-xl bg-red-50 flex items-center justify-center shrink-0">
                    <Grid className="w-5 h-5 text-red-500" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-bold text-slate-800 text-[15px] mb-1">Spreadsheet Tracking</div>
                    <div className="text-slate-500 text-[13px] leading-snug xl:pr-4">Manual updates, version mismatches, no visibility.</div>
                  </div>
                  {/* Mini Visual */}
                  <div className="shrink-0 w-full xl:w-[150px] bg-slate-50 rounded-lg border border-slate-100 p-2.5 relative overflow-hidden mt-2 xl:mt-0">
                    <div className="text-[9px] font-semibold text-slate-500 mb-2 border-b border-slate-200 pb-1">Shipment_Tracker_Final_v3.xlsx</div>
                    <div className="flex flex-col gap-1.5">
                      <div className="h-1.5 bg-slate-200 rounded w-full flex"><div className="h-full bg-red-200 rounded w-1/3 mr-1"/><div className="h-full bg-slate-300 rounded w-1/4"/></div>
                      <div className="h-1.5 bg-slate-200 rounded w-[80%] flex"><div className="h-full bg-green-200 rounded w-1/2 mr-1"/></div>
                      <div className="h-1.5 bg-slate-200 rounded w-[90%] flex"><div className="h-full bg-orange-200 rounded w-1/4 mr-1"/></div>
                    </div>
                    <div className="absolute bottom-1 right-1 w-3.5 h-3.5 bg-red-500 rounded-full flex items-center justify-center text-[8px] font-bold text-white shadow-sm">!</div>
                  </div>
                </div>

                {/* Left Card 3 */}
                <div className="relative bg-white rounded-2xl p-5 border border-red-500/10 shadow-[0_4px_20px_rgba(239,68,68,0.03)] flex flex-col xl:flex-row items-start xl:items-center gap-5 hover:shadow-[0_8px_30px_rgba(239,68,68,0.08)] transition-all">
                  <div className="w-12 h-12 rounded-xl bg-red-50 flex items-center justify-center shrink-0">
                    <FileText className="w-5 h-5 text-red-500" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-bold text-slate-800 text-[15px] mb-1">Document Chaos</div>
                    <div className="text-slate-500 text-[13px] leading-snug xl:pr-4">Files scattered across emails, drives and folders.</div>
                  </div>
                  {/* Mini Visual */}
                  <div className="shrink-0 w-full xl:w-[150px] flex flex-col gap-1.5 relative mt-2 xl:mt-0">
                    <div className="flex items-center gap-2 text-[10px] text-slate-600 font-medium"><div className="w-4 h-4 bg-red-100 rounded text-[7px] flex items-center justify-center text-red-600 font-bold">PDF</div> Commercial_Invoice.pdf</div>
                    <div className="flex items-center gap-2 text-[10px] text-slate-600 font-medium"><div className="w-4 h-4 bg-red-100 rounded text-[7px] flex items-center justify-center text-red-600 font-bold">PDF</div> Packing_List_Final.pdf</div>
                    <div className="flex items-center gap-2 text-[10px] text-slate-600 font-medium"><div className="w-4 h-4 bg-green-100 rounded text-[7px] flex items-center justify-center text-green-600 font-bold">XLS</div> HSN_Details(1).xlsx</div>
                    <div className="absolute bottom-0 right-2 w-3.5 h-3.5 bg-red-500 rounded-full flex items-center justify-center text-[8px] font-bold text-white shadow-sm">!</div>
                  </div>
                </div>

                {/* Left Card 4 */}
                <div className="relative bg-white rounded-2xl p-5 border border-red-500/10 shadow-[0_4px_20px_rgba(239,68,68,0.03)] flex flex-col xl:flex-row items-start xl:items-center gap-5 hover:shadow-[0_8px_30px_rgba(239,68,68,0.08)] transition-all">
                  <div className="w-12 h-12 rounded-xl bg-red-50 flex items-center justify-center shrink-0">
                    <Phone className="w-5 h-5 text-red-500" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-bold text-slate-800 text-[15px] mb-1">Endless Follow-Ups</div>
                    <div className="text-slate-500 text-[13px] leading-snug xl:pr-4">Status updates take hours of back-and-forth calls.</div>
                  </div>
                  {/* Mini Visual */}
                  <div className="shrink-0 w-full xl:w-[150px] flex flex-col gap-1.5 mt-2 xl:mt-0">
                    <div className="flex items-center justify-between text-[10px] text-slate-600"><div className="flex items-center gap-1.5"><Phone className="w-2.5 h-2.5"/> Customer Call</div><span>2:15 PM</span></div>
                    <div className="flex items-center justify-between text-[10px] text-slate-600"><div className="flex items-center gap-1.5"><Phone className="w-2.5 h-2.5"/> Carrier Call</div><span>2:45 PM</span></div>
                    <div className="flex items-center justify-between text-[10px] text-slate-600"><div className="flex items-center gap-1.5"><Phone className="w-2.5 h-2.5"/> Warehouse Call</div><span>3:20 PM</span></div>
                    <div className="bg-red-500 text-white text-[9px] font-bold rounded px-2 py-0.5 self-end mt-1">18+ calls</div>
                  </div>
                </div>

                {/* Left Card 5 */}
                <div className="relative bg-white rounded-2xl p-5 border border-red-500/10 shadow-[0_4px_20px_rgba(239,68,68,0.03)] flex flex-col xl:flex-row items-start xl:items-center gap-5 hover:shadow-[0_8px_30px_rgba(239,68,68,0.08)] transition-all">
                  <div className="w-12 h-12 rounded-xl bg-red-50 flex items-center justify-center shrink-0">
                    <AlertTriangle className="w-5 h-5 text-red-500" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-bold text-slate-800 text-[15px] mb-1">Compliance Mistakes</div>
                    <div className="text-slate-500 text-[13px] leading-snug xl:pr-4">Errors, rework and customs delays cost time and money.</div>
                  </div>
                  {/* Mini Visual */}
                  <div className="shrink-0 w-full xl:w-[150px] bg-red-50 rounded-xl p-2.5 border border-red-100 flex items-start gap-2.5 mt-2 xl:mt-0">
                    <div className="w-4 h-4 rounded-full bg-red-500 text-white flex items-center justify-center text-[10px] shrink-0 mt-0.5">!</div>
                    <div>
                      <div className="text-[11px] font-bold text-red-700">HSN Mismatch</div>
                      <div className="text-[9px] text-red-500/80">Shipment on hold</div>
                    </div>
                  </div>
                </div>

              </div>
            </motion.div>

            {/* Center Transformation Area */}
            <div className="hidden lg:flex flex-col items-center justify-center relative -mx-8 z-20" style={{ width: '304px' }}>
              {/* Dynamic SVG Curves */}
              <svg className="absolute inset-0 w-full h-[700px] pointer-events-none" style={{ top: '50%', transform: 'translateY(-50%)' }} viewBox="0 0 304 700" preserveAspectRatio="none">
                <defs>
                  <linearGradient id="chaosToCenter" x1="0%" y1="0%" x2="100%" y2="0%">
                    <stop offset="0%" stopColor="#EF4444" stopOpacity="0.8"/>
                    <stop offset="100%" stopColor="#14B8A6" stopOpacity="0.2"/>
                  </linearGradient>
                  <linearGradient id="centerToClarity" x1="0%" y1="0%" x2="100%" y2="0%">
                    <stop offset="0%" stopColor="#14B8A6" stopOpacity="0.2"/>
                    <stop offset="100%" stopColor="#2563EB" stopOpacity="0.8"/>
                  </linearGradient>
                </defs>
                
                {/* Lines from Left to Center */}
                <path d="M 0 100 C 80 100, 80 350, 152 350" fill="none" stroke="url(#chaosToCenter)" strokeWidth="2.5" strokeDasharray="4 12" className="animate-[dash_15s_linear_infinite]" />
                <path d="M 0 225 C 100 225, 100 350, 152 350" fill="none" stroke="url(#chaosToCenter)" strokeWidth="1.5" strokeDasharray="3 8" className="animate-[dash_12s_linear_infinite]" />
                <path d="M 0 350 C 60 350, 60 350, 152 350" fill="none" stroke="url(#chaosToCenter)" strokeWidth="2" strokeDasharray="4 12" className="animate-[dash_18s_linear_infinite]" />
                <path d="M 0 475 C 100 475, 100 350, 152 350" fill="none" stroke="url(#chaosToCenter)" strokeWidth="1.5" strokeDasharray="3 8" className="animate-[dash_12s_linear_infinite]" />
                <path d="M 0 600 C 80 600, 80 350, 152 350" fill="none" stroke="url(#chaosToCenter)" strokeWidth="2.5" strokeDasharray="4 12" className="animate-[dash_15s_linear_infinite]" />

                {/* Lines from Center to Right */}
                <path d="M 152 350 C 224 350, 224 100, 304 100" fill="none" stroke="url(#centerToClarity)" strokeWidth="2.5" strokeDasharray="4 12" className="animate-[dash_15s_linear_infinite]" />
                <path d="M 152 350 C 204 350, 204 225, 304 225" fill="none" stroke="url(#centerToClarity)" strokeWidth="1.5" strokeDasharray="3 8" className="animate-[dash_12s_linear_infinite]" />
                <path d="M 152 350 C 244 350, 244 350, 304 350" fill="none" stroke="url(#centerToClarity)" strokeWidth="2" strokeDasharray="4 12" className="animate-[dash_18s_linear_infinite]" />
                <path d="M 152 350 C 204 350, 204 475, 304 475" fill="none" stroke="url(#centerToClarity)" strokeWidth="1.5" strokeDasharray="3 8" className="animate-[dash_12s_linear_infinite]" />
                <path d="M 152 350 C 224 350, 224 600, 304 600" fill="none" stroke="url(#centerToClarity)" strokeWidth="2.5" strokeDasharray="4 12" className="animate-[dash_15s_linear_infinite]" />

                <style>{`
                  @keyframes dash { to { stroke-dashoffset: -1000; } }
                `}</style>
              </svg>

              {/* Labels */}
              <div className="absolute top-[28%] w-full flex justify-between px-6 font-black text-[14px] tracking-widest uppercase">
                <span className="text-red-500 drop-shadow-sm">Chaos</span>
                <span className="text-blue-600 drop-shadow-sm">Clarity</span>
              </div>

              {/* Center Transformation Circle */}
              <motion.div
                initial={{ opacity: 0, scale: 0.5 }} whileInView={{ opacity: 1, scale: 1 }} viewport={{ once: true }} transition={{ duration: 0.8, delay: 0.6 }}
                className="w-28 h-28 rounded-full flex flex-col items-center justify-center z-10 shadow-[0_0_50px_rgba(37,99,235,0.3)] border-[4px] border-white relative"
                style={{ background: 'linear-gradient(135deg, #14B8A6, #2563EB)' }}
              >
                <div className="absolute inset-0 rounded-full animate-ping opacity-20" style={{ border: '2px solid white' }}/>
                <ArrowRight className="w-10 h-10 text-white" />
              </motion.div>
            </div>

            {/* Right Column: With LogisticsHQ */}
            <motion.div
              initial={{ opacity: 0, x: 40 }} whileInView={{ opacity: 1, x: 0 }} viewport={{ once: true }} transition={{ duration: 0.8, delay: 0.5 }}
              className="flex flex-col gap-6 w-full"
            >
              {/* Header */}
              <div className="flex flex-col items-center text-center mb-6">
                <div className="text-white text-[11px] font-bold tracking-widest uppercase px-5 py-1.5 rounded-full mb-6 shadow-sm" style={{ background: 'linear-gradient(135deg, #14B8A6, #2563EB)' }}>
                  With LogisticsHQ
                </div>
                <h3 className="text-2xl font-black text-slate-800 tracking-tight">Connected. Automated. Reliable.</h3>
              </div>

              {/* Cards Container */}
              <div className="flex flex-col gap-4 p-6 lg:p-8 rounded-[32px] border border-blue-500/10 bg-white/50 relative overflow-hidden" style={{ boxShadow: '0 20px 40px -20px rgba(37,99,235,0.05)' }}>
                {/* Blue subtle gradient bg */}
                <div className="absolute inset-0 bg-gradient-to-bl from-teal-50/50 to-blue-50/50 pointer-events-none"/>

                {/* Right Card 1 */}
                <div className="relative bg-white rounded-2xl p-5 border border-blue-500/10 shadow-[0_4px_20px_rgba(37,99,235,0.03)] flex flex-col xl:flex-row items-start xl:items-center gap-5 hover:shadow-[0_8px_30px_rgba(37,99,235,0.08)] transition-all">
                  <div className="w-12 h-12 rounded-xl bg-teal-50 flex items-center justify-center shrink-0">
                    <Eye className="w-5 h-5 text-teal-600" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-bold text-slate-800 text-[15px] mb-1">One Shared Timeline</div>
                    <div className="text-slate-500 text-[13px] leading-snug xl:pr-4">Real-time visibility for everyone. No information gaps.</div>
                  </div>
                  {/* Mini Visual */}
                  <div className="shrink-0 w-full xl:w-[160px] flex flex-col gap-2 mt-2 xl:mt-0">
                    <div className="text-[10px] font-bold text-slate-800">Shipment FR-2847</div>
                    <div className="flex items-center gap-1 w-full relative">
                      <div className="absolute top-1/2 left-0 right-0 h-[3px] bg-slate-100 -translate-y-1/2 -z-10 rounded-full"/>
                      <div className="absolute top-1/2 left-0 w-3/4 h-[3px] bg-teal-500 -translate-y-1/2 -z-10 rounded-full"/>
                      <div className="w-3 h-3 rounded-full bg-teal-500 border-[2px] border-white flex-shrink-0"></div>
                      <div className="flex-1"></div>
                      <div className="w-3 h-3 rounded-full bg-teal-500 border-[2px] border-white flex-shrink-0"></div>
                      <div className="flex-1"></div>
                      <div className="w-3 h-3 rounded-full bg-teal-500 border-[2px] border-white flex-shrink-0"></div>
                      <div className="flex-1"></div>
                      <div className="w-4 h-4 rounded-full bg-teal-500 flex items-center justify-center border-2 border-white flex-shrink-0"><Check className="w-2.5 h-2.5 text-white"/></div>
                    </div>
                    <div className="flex justify-between text-[7px] text-slate-400 font-semibold mt-0.5">
                      <span>Pick Up</span><span>In Transit</span><span>Customs</span><span>Delivered</span>
                    </div>
                  </div>
                </div>

                {/* Right Card 2 */}
                <div className="relative bg-white rounded-2xl p-5 border border-blue-500/10 shadow-[0_4px_20px_rgba(37,99,235,0.03)] flex flex-col xl:flex-row items-start xl:items-center gap-5 hover:shadow-[0_8px_30px_rgba(37,99,235,0.08)] transition-all">
                  <div className="w-12 h-12 rounded-xl bg-teal-50 flex items-center justify-center shrink-0">
                    <Zap className="w-5 h-5 text-teal-600" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-bold text-slate-800 text-[15px] mb-1">Automated Updates</div>
                    <div className="text-slate-500 text-[13px] leading-snug xl:pr-4">Proactive alerts and notifications keep everyone informed.</div>
                  </div>
                  {/* Mini Visual */}
                  <div className="shrink-0 w-full xl:w-[160px] bg-slate-50 rounded-lg border border-slate-100 p-2.5 relative mt-2 xl:mt-0">
                    <div className="flex justify-between items-center mb-2"><span className="text-[10px] font-bold text-slate-800">Shipment Update</span><span className="text-[8px] text-slate-400">2m ago</span></div>
                    <div className="text-[9px] text-slate-500 pr-4">Your shipment is out for delivery</div>
                    <div className="absolute bottom-2 right-2 w-5 h-5 bg-teal-50 rounded-full flex items-center justify-center border border-teal-100"><Zap className="w-3 h-3 text-teal-600"/></div>
                  </div>
                </div>

                {/* Right Card 3 */}
                <div className="relative bg-white rounded-2xl p-5 border border-blue-500/10 shadow-[0_4px_20px_rgba(37,99,235,0.03)] flex flex-col xl:flex-row items-start xl:items-center gap-5 hover:shadow-[0_8px_30px_rgba(37,99,235,0.08)] transition-all">
                  <div className="w-12 h-12 rounded-xl bg-teal-50 flex items-center justify-center shrink-0">
                    <Folder className="w-5 h-5 text-teal-600" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-bold text-slate-800 text-[15px] mb-1">Centralized Documents</div>
                    <div className="text-slate-500 text-[13px] leading-snug xl:pr-4">All documents organized, version-controlled and easy to access.</div>
                  </div>
                  {/* Mini Visual */}
                  <div className="shrink-0 w-full xl:w-[160px] flex flex-col gap-1.5 mt-2 xl:mt-0">
                    <div className="text-[10px] font-bold text-slate-800 mb-0.5">Documents</div>
                    <div className="flex items-center justify-between text-[10px] text-slate-600 bg-slate-50 px-2 py-1 rounded"><div className="flex items-center gap-1.5"><Folder className="w-3 h-3 text-blue-500"/> Commercial_Invoice.pdf</div><CheckCircle2 className="w-3 h-3 text-teal-500"/></div>
                    <div className="flex items-center justify-between text-[10px] text-slate-600 bg-slate-50 px-2 py-1 rounded"><div className="flex items-center gap-1.5"><Folder className="w-3 h-3 text-blue-500"/> Packing_List.pdf</div><CheckCircle2 className="w-3 h-3 text-teal-500"/></div>
                    <div className="flex items-center justify-between text-[10px] text-slate-600 bg-slate-50 px-2 py-1 rounded"><div className="flex items-center gap-1.5"><Folder className="w-3 h-3 text-blue-500"/> Cert_of_Origin.pdf</div><CheckCircle2 className="w-3 h-3 text-teal-500"/></div>
                  </div>
                </div>

                {/* Right Card 4 */}
                <div className="relative bg-white rounded-2xl p-5 border border-blue-500/10 shadow-[0_4px_20px_rgba(37,99,235,0.03)] flex flex-col xl:flex-row items-start xl:items-center gap-5 hover:shadow-[0_8px_30px_rgba(37,99,235,0.08)] transition-all">
                  <div className="w-12 h-12 rounded-xl bg-teal-50 flex items-center justify-center shrink-0">
                    <Users className="w-5 h-5 text-teal-600" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-bold text-slate-800 text-[15px] mb-1">No More Follow-Ups</div>
                    <div className="text-slate-500 text-[13px] leading-snug xl:pr-4">Real-time status. Fewer calls. More productivity.</div>
                  </div>
                  {/* Mini Visual */}
                  <div className="shrink-0 w-full xl:w-[160px] flex flex-col gap-1 mt-2 xl:mt-0">
                    <div className="text-[10px] font-bold text-slate-800">Follow-up Calls</div>
                    <div className="flex items-end gap-2 bg-slate-50 p-2 rounded-lg border border-slate-100">
                      <div className="text-3xl font-black text-slate-800 leading-none">2<span className="text-[9px] font-medium text-slate-400 ml-0.5">/ shipment</span></div>
                      <svg viewBox="0 0 50 20" className="w-[50px] h-[20px] ml-auto overflow-visible">
                        <path d="M0,5 Q10,5 20,10 T40,18 L50,18" fill="none" stroke="#14B8A6" strokeWidth="2.5" strokeLinecap="round"/>
                      </svg>
                    </div>
                  </div>
                </div>

                {/* Right Card 5 */}
                <div className="relative bg-white rounded-2xl p-5 border border-blue-500/10 shadow-[0_4px_20px_rgba(37,99,235,0.03)] flex flex-col xl:flex-row items-start xl:items-center gap-5 hover:shadow-[0_8px_30px_rgba(37,99,235,0.08)] transition-all">
                  <div className="w-12 h-12 rounded-xl bg-teal-50 flex items-center justify-center shrink-0">
                    <ShieldCheck className="w-5 h-5 text-teal-600" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-bold text-slate-800 text-[15px] mb-1">Built-in Compliance</div>
                    <div className="text-slate-500 text-[13px] leading-snug xl:pr-4">Automated validations. Fewer errors. Faster clearance.</div>
                  </div>
                  {/* Mini Visual */}
                  <div className="shrink-0 w-full xl:w-[160px] flex flex-col justify-between h-auto gap-2 bg-slate-50 p-2.5 rounded-lg border border-slate-100 mt-2 xl:mt-0">
                    <div className="flex justify-between items-center"><div className="text-[10px] font-bold text-slate-800">Compliance Check</div><CheckCircle2 className="w-4 h-4 text-teal-500"/></div>
                    <div className="flex justify-between items-center"><div className="text-[9px] text-slate-500 w-[70px] leading-tight">All documents verified</div><div className="bg-teal-50 text-teal-700 px-2 py-1 rounded text-[9px] font-bold border border-teal-100">Success</div></div>
                  </div>
                </div>

              </div>
            </motion.div>

          </div>

          {/* Bottom Highlight Panel */}
          <motion.div
            initial={{ opacity: 0, y: 40 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} transition={{ duration: 0.8, delay: 0.8 }}
            className="mt-20 relative rounded-[28px] p-8 lg:p-12 flex flex-col md:flex-row items-center justify-center gap-8 overflow-hidden border border-blue-500/10 max-w-[1100px] mx-auto text-center md:text-left"
            style={{ background: 'linear-gradient(135deg, rgba(20,184,166,0.08) 0%, rgba(37,99,235,0.08) 100%)' }}
          >
            <div className="w-16 h-16 rounded-2xl flex items-center justify-center shrink-0 shadow-xl" style={{ background: 'linear-gradient(135deg, #14B8A6, #2563EB)' }}>
              <Zap className="w-8 h-8 text-white" />
            </div>
            <div>
              <h3 className="text-[28px] font-black text-slate-800 mb-2">From Disconnected To Intelligent</h3>
              <p className="text-slate-600 text-[18px] leading-relaxed max-w-[800px]">
                LogisticsHQ brings freight operations, compliance, documents, communication, and visibility together in one unified operating system.
              </p>
            </div>
          </motion.div>

        </div>
      </section>

      {/* ═══ FINAL CTA SECTION (07) - CINEMATIC ═══ */}
      <section className="relative w-full flex justify-center bg-[#0F172A]">
        <motion.div
          initial={{ opacity: 0, y: 40 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} transition={{ duration: 1 }}
          className="w-full min-h-[650px] py-20 relative overflow-hidden flex flex-col justify-center items-center text-center"
        >
          {/* Background Image with Slow Zoom */}
          <div className="absolute inset-0 z-0 overflow-hidden bg-[#0F172A]">
            <style>{`
              @keyframes slowZoom {
                0% { transform: scale(1); }
                50% { transform: scale(1.05); }
                100% { transform: scale(1); }
              }
            `}</style>
            <div 
              className="absolute inset-0 w-full h-full animate-[slowZoom_20s_ease-in-out_infinite]"
              style={{
                backgroundImage: "url('/images/about/final-cta.png')",
                backgroundSize: 'cover',
                backgroundPosition: 'center'
              }}
            />
            {/* Dark Gradient Overlay */}
            <div className="absolute inset-0" style={{ background: 'linear-gradient(180deg, rgba(15,23,42,0.3), rgba(15,23,42,0.65))' }} />
          </div>

          <div className="relative z-10 flex flex-col items-center w-full px-6 max-w-[900px]">
            {/* Top Label */}
            <motion.div
              initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} transition={{ duration: 0.8, delay: 0.2 }}
              className="text-[#14B8A6] text-[14px] font-bold tracking-[0.15em] uppercase mb-6"
            >
              The Future of Freight Operations
            </motion.div>

            {/* Headline */}
            <motion.h2
              initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} transition={{ duration: 0.8, delay: 0.3 }}
              className="font-black text-white mb-10 tracking-tight"
              style={{ fontSize: 'clamp(32px, 4vw, 56px)', lineHeight: 1.05 }}
            >
              The Logistics Industry<br className="hidden md:block" /> Doesn't Need More Tools.<br/>
              It Needs An Operating System.
            </motion.h2>


            {/* Buttons */}
            <motion.div
              initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} transition={{ duration: 0.8, delay: 0.5 }}
              className="flex flex-col sm:flex-row items-center justify-center gap-4 w-full mb-12"
            >
              <Link to="/contact" className="h-[64px] px-[36px] rounded-full flex items-center justify-center text-white font-bold text-lg transition-transform hover:scale-105 duration-300" style={{ background: 'linear-gradient(135deg, #14B8A6, #2563EB)' }}>
                Join LogisticsHQ
              </Link>
              <Link to="/contact" className="h-[64px] px-[36px] rounded-full flex items-center justify-center text-white font-bold text-lg transition-colors hover:bg-white/10 duration-300" style={{ background: 'rgba(255,255,255,0.08)', backdropFilter: 'blur(12px)', WebkitBackdropFilter: 'blur(12px)', border: '1px solid rgba(255,255,255,0.2)' }}>
                Book Demo
              </Link>
            </motion.div>

            {/* Trust Bar */}
            <motion.div
              initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} transition={{ duration: 0.8, delay: 0.6 }}
              className="flex flex-wrap justify-center items-center gap-x-6 gap-y-4 text-white/90 text-sm font-medium tracking-wide"
            >
              <div>500+ Businesses</div>
              <div className="w-[1px] h-4 bg-white/30 hidden sm:block" />
              <div>150+ Ports</div>
              <div className="w-[1px] h-4 bg-white/30 hidden sm:block" />
              <div>10K+ Trade Lanes</div>
              <div className="w-[1px] h-4 bg-white/30 hidden md:block" />
              <div>24/7 Operations</div>
            </motion.div>
          </div>
        </motion.div>
      </section>

    </div>
  );
}
