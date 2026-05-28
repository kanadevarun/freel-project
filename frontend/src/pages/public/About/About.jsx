import { useEffect, useRef } from 'react';
import { Link } from 'react-router-dom';
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
      {/* ═══ HERO ═══ */}
      <section className="about-hero relative pt-32 pb-24 lg:pt-48 lg:pb-32 overflow-hidden">
        {/* Animated Background Blobs */}
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[800px] bg-brand-teal opacity-[0.07] rounded-full blur-3xl -z-10 animate-pulse"></div>
        <div className="absolute -top-20 right-0 w-[500px] h-[500px] bg-brand-indigo opacity-[0.05] rounded-full blur-3xl -z-10 animate-pulse" style={{animationDelay: '1s'}}></div>
        
        <Reveal>
          <div className="max-w-4xl mx-auto text-center px-4 relative z-10">
            <div className="inline-flex items-center justify-center px-4 py-1.5 mb-8 text-sm font-semibold text-brand-indigo bg-brand-indigo/10 rounded-full border border-brand-indigo/20 tracking-wide uppercase">
              Our Mission
            </div>
            <h1 className="text-5xl md:text-6xl lg:text-7xl font-extrabold text-brand-navy tracking-tight mb-8 leading-[1.15]">
              Digitizing India's <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand-teal to-blue-600">
                $200B Logistics Market
              </span>
            </h1>
            <p className="text-xl md:text-2xl text-ui-gray max-w-3xl mx-auto font-light leading-relaxed">
              Born from the frustration of managing global trade via WhatsApp groups and Excel sheets, Freel is India's first multi-modal logistics operating system.
            </p>
          </div>
        </Reveal>
      </section>

      {/* ═══ STATS BAR (Elevated) ═══ */}
      <Reveal delay="reveal-delay-1">
        <section className="max-w-6xl mx-auto px-4 sm:px-6 relative z-20 -mt-12 mb-20">
          <div className="bg-white/80 backdrop-blur-xl border border-white rounded-[2rem] p-8 md:p-12 shadow-[0_20px_60px_-15px_rgba(0,0,0,0.05)] grid grid-cols-2 md:grid-cols-4 gap-8 md:gap-4 divide-x-0 md:divide-x divide-gray-100">
            <div className="text-center group">
              <div className="text-4xl md:text-5xl font-extrabold text-transparent bg-clip-text bg-gradient-to-br from-brand-teal to-brand-indigo mb-2 group-hover:scale-110 transition-transform duration-300">500+</div>
              <div className="text-sm font-bold uppercase tracking-wider text-ui-gray">Verified Vendors</div>
            </div>
            <div className="text-center group">
              <div className="text-4xl md:text-5xl font-extrabold text-transparent bg-clip-text bg-gradient-to-br from-brand-teal to-brand-indigo mb-2 group-hover:scale-110 transition-transform duration-300">150+</div>
              <div className="text-sm font-bold uppercase tracking-wider text-ui-gray">Global Ports</div>
            </div>
            <div className="text-center group">
              <div className="text-4xl md:text-5xl font-extrabold text-transparent bg-clip-text bg-gradient-to-br from-brand-teal to-brand-indigo mb-2 group-hover:scale-110 transition-transform duration-300">10K+</div>
              <div className="text-sm font-bold uppercase tracking-wider text-ui-gray">Trade Lanes</div>
            </div>
            <div className="text-center group">
              <div className="text-4xl md:text-5xl font-extrabold text-transparent bg-clip-text bg-gradient-to-br from-brand-teal to-brand-indigo mb-2 group-hover:scale-110 transition-transform duration-300">24/7</div>
              <div className="text-sm font-bold uppercase tracking-wider text-ui-gray">Automated Ops</div>
            </div>
          </div>
        </section>
      </Reveal>

      {/* ═══ TEAM SECTION ═══ */}
      <section className="py-24 bg-slate-50 relative border-y border-ui-border overflow-hidden">
        <div className="absolute top-1/2 left-0 w-full h-[500px] bg-gradient-to-b from-transparent via-brand-teal/5 to-transparent -translate-y-1/2 -z-10 skew-y-3"></div>
        <div className="max-w-6xl mx-auto px-4 sm:px-6">
          <Reveal>
            <div className="text-center mb-16">
              <span className="text-brand-indigo font-bold tracking-widest uppercase text-sm mb-3 block">The Leadership</span>
              <h2 className="text-4xl font-extrabold text-brand-navy">Meet Our Founders</h2>
              <p className="text-ui-gray text-lg mt-4 max-w-2xl mx-auto">Built by logistics natives for logistics professionals. We know the pain because we lived it.</p>
            </div>
          </Reveal>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-12 max-w-4xl mx-auto">
            <Reveal delay="reveal-delay-1">
              <div className="team-card group relative">
                <div className="absolute inset-0 bg-gradient-to-br from-brand-teal/5 to-brand-indigo/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500 rounded-3xl -z-10"></div>
                <div className="team-avatar-container relative mx-auto mb-8">
                  <div className="absolute inset-0 bg-gradient-to-tr from-brand-teal to-brand-indigo rounded-full blur-lg opacity-40 group-hover:opacity-70 group-hover:scale-110 transition-all duration-500"></div>
                  <div className="team-avatar relative z-10 w-40 h-40 mx-auto rounded-full border-4 border-white shadow-xl bg-white flex items-center justify-center overflow-hidden">
                    <svg viewBox="0 0 100 100" style={{ width: '100%', height: '100%' }}>
                      <defs>
                        <linearGradient id="ceoGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                          <stop offset="0%" stopColor="var(--color-brand-teal)" />
                          <stop offset="100%" stopColor="var(--color-brand-indigo)" />
                        </linearGradient>
                      </defs>
                      <circle cx="50" cy="50" r="50" fill="url(#ceoGrad)" opacity="0.15" />
                      <circle cx="50" cy="50" r="44" fill="url(#ceoGrad)" />
                      {/* Female CEO Stylized Avatar */}
                      <path d="M50 25 C40 25, 38 38, 50 38 C62 38, 60 25, 50 25 Z" fill="#ffffff" />
                      <path d="M50 38 C42 38, 30 45, 30 58 C30 65, 70 65, 70 58 C70 45, 58 38, 50 38 Z" fill="#ffffff" opacity="0.9" />
                      <circle cx="50" cy="31" r="9" fill="#ffffff" />
                      <path d="M42 48 L50 56 L58 48" stroke="var(--color-brand-indigo)" strokeWidth="2.5" strokeLinecap="round" fill="none" />
                      {/* Gold Star lapel pins */}
                      <circle cx="50" cy="56" r="1.5" fill="#f59e0b" />
                    </svg>
                  </div>
                </div>
                <h3 className="text-2xl font-bold text-brand-navy mb-1 group-hover:text-brand-teal transition-colors">Vaidanshi Sagar</h3>
                <div className="text-brand-indigo font-medium tracking-wide uppercase text-sm mb-4">CEO & Co-Founder</div>
                <p className="text-ui-gray text-sm px-4">Former freight forwarding executive obsessed with bringing transparency and automation to global supply chains.</p>
              </div>
            </Reveal>

            <Reveal delay="reveal-delay-2">
              <div className="team-card group relative">
                <div className="absolute inset-0 bg-gradient-to-br from-brand-indigo/5 to-blue-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500 rounded-3xl -z-10"></div>
                <div className="team-avatar-container relative mx-auto mb-8">
                  <div className="absolute inset-0 bg-gradient-to-tr from-brand-indigo to-blue-500 rounded-full blur-lg opacity-40 group-hover:opacity-70 group-hover:scale-110 transition-all duration-500"></div>
                  <div className="team-avatar relative z-10 w-40 h-40 mx-auto rounded-full border-4 border-white shadow-xl bg-white flex items-center justify-center overflow-hidden">
                    <svg viewBox="0 0 100 100" style={{ width: '100%', height: '100%' }}>
                      <defs>
                        <linearGradient id="baGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                          <stop offset="0%" stopColor="var(--color-brand-indigo)" />
                          <stop offset="100%" stopColor="#3b82f6" />
                        </linearGradient>
                      </defs>
                      <circle cx="50" cy="50" r="50" fill="url(#baGrad)" opacity="0.15" />
                      <circle cx="50" cy="50" r="44" fill="url(#baGrad)" />
                      {/* Male Analyst Stylized Avatar */}
                      <path d="M50 25 C41 25, 40 37, 50 37 C60 37, 59 25, 50 25 Z" fill="#ffffff" />
                      <path d="M50 37 C43 37, 32 44, 32 57 C32 64, 68 64, 68 57 C68 44, 57 37, 50 37 Z" fill="#ffffff" opacity="0.9" />
                      <circle cx="50" cy="30" r="8.5" fill="#ffffff" />
                      <path d="M44 46 L50 53 L56 46" stroke="#3b82f6" strokeWidth="2.5" strokeLinecap="round" fill="none" />
                      {/* Analyst specs accent */}
                      <rect x="42" y="27" width="16" height="4" rx="1.5" fill="#3b82f6" opacity="0.3" />
                    </svg>
                  </div>
                </div>
                <h3 className="text-2xl font-bold text-brand-navy mb-1 group-hover:text-brand-indigo transition-colors">Yash Sagar</h3>
                <div className="text-brand-teal font-medium tracking-wide uppercase text-sm mb-4">Business Analyst</div>
                <p className="text-ui-gray text-sm px-4">Operations wizard mapping out the intricate workflows of the logistics world into a seamless digital experience.</p>
              </div>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ═══ OUR STORY (Split Layout) ═══ */}
      <section className="py-32 bg-white relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
            <Reveal>
              <div className="relative">
                <div className="absolute -left-12 -top-12 text-[150px] font-serif text-brand-teal/10 leading-none select-none">"</div>
                <h2 className="text-3xl md:text-5xl font-extrabold text-brand-navy leading-tight mb-8 relative z-10">
                  What if comparing freight rates was as simple as searching flights on <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand-teal to-blue-600">Google?</span>
                </h2>
                <div className="w-24 h-2 bg-brand-teal rounded-full"></div>
              </div>
            </Reveal>
            
            <div className="space-y-6 text-lg text-slate-600">
              <Reveal delay="reveal-delay-1">
                <p>
                  Logistics in India is a $200 billion industry — yet most of it runs on phone calls, WhatsApp forwards, and spreadsheets. Rate requests take days. Tracking requires constant follow-ups. Compliance documentation is a nightmare.
                </p>
              </Reveal>
              <Reveal delay="reveal-delay-2">
                <p>
                  We started Freel because we saw this gap firsthand. As people who worked in freight forwarding, we knew the pain of juggling 15 vendor WhatsApp groups, manually creating rate comparison sheets, and losing track of shipments across modes.
                </p>
              </Reveal>
              <Reveal delay="reveal-delay-3">
                <p>
                  Today, we serve shippers, freight forwarders, and carriers — all on the same platform, each with their own tailored dashboard experience. We believe logistics technology should be simple, transparent, and accessible to every business in India.
                </p>
              </Reveal>
            </div>
          </div>
        </div>
      </section>

      {/* ═══ OUR VALUES (Bento Grid) ═══ */}
      <section className="py-24 bg-brand-navy text-white relative overflow-hidden">
        <div className="absolute top-0 right-0 w-[600px] h-[600px] bg-brand-teal opacity-10 rounded-full blur-[100px] -z-10 translate-x-1/2 -translate-y-1/2"></div>
        <div className="absolute bottom-0 left-0 w-[600px] h-[600px] bg-brand-indigo opacity-10 rounded-full blur-[100px] -z-10 -translate-x-1/2 translate-y-1/2"></div>
        
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <Reveal>
            <div className="text-center mb-16">
              <span className="text-brand-teal font-bold tracking-widest uppercase text-sm mb-3 block">Our Values</span>
              <h2 className="text-4xl md:text-5xl font-extrabold text-white">What Drives Us</h2>
            </div>
          </Reveal>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-5xl mx-auto auto-rows-[250px]">
            <Reveal className="md:col-span-2">
              <div className="bg-white/5 border border-white/10 rounded-3xl p-8 h-full flex flex-col justify-end relative overflow-hidden group hover:bg-white/10 transition-colors">
                <div className="absolute top-8 right-8 text-5xl opacity-50 group-hover:scale-110 group-hover:opacity-100 transition-all duration-300">⚡</div>
                <h3 className="text-2xl font-bold mb-3">Speed Over Perfection</h3>
                <p className="text-slate-300 text-sm max-w-md">In logistics, every hour costs money. We optimize for speed — in our product, our decisions, and our support response times.</p>
              </div>
            </Reveal>
            
            <Reveal delay="reveal-delay-1">
              <div className="bg-white/5 border border-white/10 rounded-3xl p-8 h-full flex flex-col justify-end relative overflow-hidden group hover:bg-white/10 transition-colors">
                <div className="absolute top-8 right-8 text-5xl opacity-50 group-hover:scale-110 group-hover:opacity-100 transition-all duration-300">🔍</div>
                <h3 className="text-xl font-bold mb-3">Radical Transparency</h3>
                <p className="text-slate-300 text-sm">No hidden markups. No black-box pricing. See exactly what you're paying and why.</p>
              </div>
            </Reveal>
            
            <Reveal delay="reveal-delay-2">
              <div className="bg-gradient-to-br from-brand-teal/20 to-brand-teal/5 border border-brand-teal/20 rounded-3xl p-8 h-full flex flex-col justify-end relative overflow-hidden group hover:bg-brand-teal/20 transition-colors">
                <div className="absolute top-8 right-8 text-5xl opacity-50 group-hover:scale-110 group-hover:opacity-100 transition-all duration-300">🤝</div>
                <h3 className="text-xl font-bold mb-3">Both Sides of the Table</h3>
                <p className="text-slate-300 text-sm">We serve both shippers and carriers on the same platform. Features create value for everyone.</p>
              </div>
            </Reveal>
            
            <Reveal className="md:col-span-2" delay="reveal-delay-3">
              <div className="bg-white/5 border border-white/10 rounded-3xl p-8 h-full flex flex-col justify-end relative overflow-hidden group hover:bg-white/10 transition-colors">
                <div className="absolute top-8 right-8 text-5xl opacity-50 group-hover:scale-110 group-hover:opacity-100 transition-all duration-300">🧠</div>
                <h3 className="text-2xl font-bold mb-3">Simplicity is Hard</h3>
                <p className="text-slate-300 text-sm max-w-md">Logistics is inherently complex. Our job is to absorb that complexity so our users don't have to. Every feature should feel obvious.</p>
              </div>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ═══ WHY US ═══ */}
      <section id="why-us" className="py-24 bg-ui-surface">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <Reveal>
            <div className="text-center mb-16">
              <span className="text-brand-indigo font-bold tracking-widest uppercase text-sm mb-3 block">Why Freel</span>
              <h2 className="text-4xl font-extrabold text-brand-navy mb-4">What Makes Us Different</h2>
              <p className="text-ui-gray text-lg max-w-2xl mx-auto">
                We're not another logistics marketplace. We're an operating system that replaces 5+ tools you're already using.
              </p>
            </div>
          </Reveal>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {[
              { num: '01', title: 'Multi-Modal, Single Platform', desc: 'Road, Air, and Sea freight — all managed from one dashboard. No more switching between vendor portals.', tags: ['Road', 'Air', 'Sea'] },
              { num: '02', title: 'Real-Time Intelligence', desc: 'Compare 500+ vendor rates instantly. No more emailing vendors for quotes and waiting 2 days.', tags: ['500+ Vendors', 'Instant Quotes'] },
              { num: '03', title: 'One Platform, Every Role', desc: 'Whether you\'re a shipper, forwarder, or carrier — your dashboard adapts to your workflow.', tags: ['Shippers', 'Forwarders', 'Carriers'] },
              { num: '04', title: 'Built-In Compliance', desc: 'HSN lookup, MSDS verification, KYC management — compliance is integrated into every flow.', tags: ['HSN', 'MSDS', 'KYC'] },
              { num: '05', title: 'CRM + Discovery', desc: 'Find new clients, manage leads, and track revenue — all without leaving Freel.', tags: ['Lead Discovery', 'Pipeline CRM'] },
              { num: '06', title: 'Transparent Pricing', desc: 'Zero hidden markups. We earn through a simple SaaS subscription. Every rate is direct.', tags: ['No Middlemen', 'Direct Rates'] }
            ].map((feature, i) => (
              <Reveal key={i} delay={`reveal-delay-${(i % 3) + 1}`}>
                <div className="bg-white rounded-3xl p-8 border border-ui-border transition-all duration-300 hover:border-brand-teal/50 hover:bg-slate-50 group relative overflow-hidden h-full flex flex-col">
                  <div className="text-6xl font-black text-slate-100 group-hover:text-brand-teal/5 transition-colors mb-6 -ml-2 select-none">{feature.num}</div>
                  <h3 className="text-xl font-bold text-brand-navy mb-3">{feature.title}</h3>
                  <p className="text-slate-600 mb-8 flex-grow leading-relaxed">{feature.desc}</p>
                  <div className="flex flex-wrap gap-2 mt-auto">
                    {feature.tags.map((tag, j) => (
                      <span key={j} className="px-3 py-1 bg-slate-50 border border-slate-100 rounded-full text-xs font-semibold text-slate-500 group-hover:border-brand-teal/20 group-hover:bg-brand-teal/5 group-hover:text-brand-teal transition-colors">
                        {tag}
                      </span>
                    ))}
                  </div>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ CTA ═══ */}
      <section className="py-24 bg-white">
        <Reveal>
          <div className="max-w-5xl mx-auto px-4 sm:px-6">
            <div className="bg-brand-navy rounded-[3rem] p-12 md:p-20 text-center relative overflow-hidden shadow-2xl border border-white/10">
              <div className="absolute top-0 left-0 w-full h-full overflow-hidden z-0">
                <div className="absolute -top-32 -right-32 w-96 h-96 bg-brand-teal rounded-full mix-blend-screen filter blur-[80px] opacity-30 animate-pulse"></div>
                <div className="absolute -bottom-32 -left-32 w-96 h-96 bg-brand-indigo rounded-full mix-blend-screen filter blur-[80px] opacity-30 animate-pulse" style={{animationDelay: '1s'}}></div>
              </div>

              <div className="relative z-10">
                <h2 className="text-4xl md:text-5xl font-extrabold text-white mb-6 tracking-tight">Ready to Simplify Your Logistics?</h2>
                <p className="text-slate-300 text-xl mb-10 max-w-2xl mx-auto font-light">
                  Join 500+ businesses already using Freel to manage their entire supply chain from one elegant dashboard.
                </p>
                <div className="flex flex-col sm:flex-row justify-center gap-4">
                  <Link to="/contact" className="px-8 py-4 text-brand-navy font-bold text-lg bg-white rounded-full shadow-lg hover:scale-105 transition-transform outline-none flex items-center justify-center">
                    Start Free Trial <span className="ml-2">→</span>
                  </Link>
                  <Link to="/contact" className="px-8 py-4 text-white font-bold text-lg bg-white/10 border border-white/20 rounded-full hover:bg-white/20 transition-colors flex items-center justify-center backdrop-blur-sm">
                    Contact Sales
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </Reveal>
      </section>
    </div>
  );
}
