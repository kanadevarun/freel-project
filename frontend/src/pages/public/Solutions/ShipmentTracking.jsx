import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js';
import { Line, Bar } from 'react-chartjs-2';
import '../../../styles/Solutions.css';

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

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

/* ─── Data ─── */
const modes = [
  {
    icon: '🚛', title: 'Road Tracking', desc: 'GPS-based tracking with live location updates every 30 seconds.',
    features: ['Live truck location on India map', 'Speed, halt & deviation alerts', 'ETA prediction via AI', 'Digital LR with QR code', 'E-POD with photo proof'],
    color: 'var(--color-brand-teal)', hoverClass: 'hover:-translate-y-2'
  },
  {
    icon: '✈️', title: 'Air Tracking', desc: 'FlightAware integration for real-time flight position tracking.',
    features: ['Live flight position on globe', 'AWB milestone updates', 'Departure → In Air → Arrived', 'Customs clearance status', 'Last-mile delivery tracking'],
    color: '#3B82F6', hoverClass: 'hover:-translate-y-2 relative overflow-hidden'
  },
  {
    icon: '🚢', title: 'Sea Tracking', desc: 'AIS vessel tracking with container-level milestone updates.',
    features: ['Vessel position on ocean map', 'B/L and container tracking', 'Port congestion alerts', 'Transshipment updates', 'POD/POL with ETA updates'],
    color: 'var(--color-brand-indigo)', hoverClass: 'hover:-translate-y-2 relative overflow-hidden group'
  },
];

const featureData = {
  whatsapp: {
    icon: '📱', title: 'Automated WhatsApp Alerts',
    desc: 'Keep your clients and internal teams informed effortlessly. Set up triggered notifications sent directly to WhatsApp for key milestones: dispatch, border crossings, arrival, and final delivery.',
    Visual: () => (
      <div className="bg-slate-50 p-4 rounded-lg border border-gray-100 font-mono text-sm text-slate-700 flex flex-col gap-2">
        <div className="bg-white p-2 rounded shadow-sm self-start max-w-[80%] border-l-4 border-brand-teal">
          <span className="font-bold text-xs text-gray-500 block mb-1">System Msg</span>
          Shipment #FR-8892 is out for delivery. ETA: 14:30. Track here: freel.co/t/8892
        </div>
      </div>
    )
  },
  pod: {
    icon: '📸', title: 'Instant Proof of Delivery',
    desc: 'Eliminate paperwork delays. Drivers can upload photos of the delivered goods and signed waybills directly via a mobile link, updating the dashboard instantly to trigger invoicing.',
    Visual: () => (
      <div className="flex items-center justify-center h-32 bg-slate-100 rounded-lg border-2 border-dashed border-gray-300">
        <div className="text-center text-slate-500">
          <span className="block text-2xl mb-1">📄➡️💻</span> Digital Document Sync
        </div>
      </div>
    )
  },
  exceptions: {
    icon: '🔔', title: 'Proactive Exception Alerts',
    desc: 'Don\'t wait for a customer call. Our AI monitors route deviations, unscheduled halts, and port congestion, raising alerts immediately so your team can intervene and mitigate delays.',
    Visual: () => (
      <div className="bg-red-50 text-red-700 p-4 rounded-lg border border-red-200">
        <strong className="flex items-center"><span className="mr-2">⚠️</span> Route Deviation Detected</strong>
        <span className="block text-sm mt-1">Truck MH-12-AB-1234 is 15km off intended path. ETA impacted.</span>
      </div>
    )
  },
  sharing: {
    icon: '🔗', title: 'Secure Client Sharing',
    desc: 'Generate secure, white-labeled tracking links with one click. Decide exactly what information (location, documents, pricing) your customer can see without giving them full system access.',
    Visual: () => (
      <div className="flex items-center p-3 bg-slate-50 rounded border border-gray-200 font-mono text-sm overflow-hidden">
        <span className="text-gray-400 mr-2 select-none">Link:</span>
        <span className="text-brand-teal truncate">https://track.freel.co/share/v2/xyz</span>
        <button className="ml-auto bg-white border shadow-sm px-2 py-1 rounded text-xs hover:bg-gray-50">Copy</button>
      </div>
    )
  },
  timeline: {
    icon: '📋', title: 'Unified Timeline View',
    desc: 'See the complete story of a shipment across multiple legs. From factory pickup to port loading, ocean transit, and final mile delivery—all presented in a clean, scrollable chronological timeline.',
    Visual: () => (
      <div className="border-l-2 border-brand-teal/30 ml-3 pl-4 space-y-4 py-2">
        <div className="relative">
          <div className="absolute -left-[23px] top-1 w-3 h-3 bg-brand-teal rounded-full"></div>
          <span className="text-xs text-gray-500">10:00 AM</span> 
          <div className="font-medium">Picked up at Origin</div>
        </div>
        <div className="relative">
          <div className="absolute -left-[23px] top-1 w-3 h-3 bg-white border-2 border-brand-teal rounded-full"></div>
          <span className="text-xs text-gray-500">Est. Tomorrow</span> 
          <div className="font-medium text-gray-400">Port Arrival</div>
        </div>
      </div>
    )
  },
  analytics: {
    icon: '📊', title: 'Tracking Analytics',
    desc: 'Turn tracking data into business intelligence. Analyze lane performance, carrier reliability, and average dwell times to optimize your supply chain network.',
    Visual: () => (
      <>
        <div className="h-24 flex items-end justify-between gap-2 border-b border-gray-200 pb-2 px-2">
          <div className="w-1/6 bg-brand-teal/20 h-[40%] rounded-t"></div>
          <div className="w-1/6 bg-brand-teal/40 h-[60%] rounded-t"></div>
          <div className="w-1/6 bg-brand-teal/60 h-[30%] rounded-t"></div>
          <div className="w-1/6 bg-brand-teal h-[90%] rounded-t shadow-sm"></div>
          <div className="w-1/6 bg-brand-teal/80 h-[70%] rounded-t"></div>
        </div>
        <div className="text-center text-xs text-gray-400 mt-1">Carrier Performance Trend</div>
      </>
    )
  }
};

/* ═══════════════════════════════════════════ */
/*        SHIPMENT TRACKING PAGE              */
/* ═══════════════════════════════════════════ */
export default function ShipmentTracking() {
  const [activeFeature, setActiveFeature] = useState('whatsapp');
  const [fadeState, setFadeState] = useState('in');
  const [chartType, setChartType] = useState('performance');

  const feat = featureData[activeFeature];

  const handleFeatureClick = (key) => {
    if (key === activeFeature) return;
    setFadeState('out');
    setTimeout(() => {
      setActiveFeature(key);
      setFadeState('in');
    }, 300);
  };

  // Chart Data
  const performanceData = {
    labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
    datasets: [{
      label: 'On-Time Delivery %',
      data: [85, 88, 86, 92, 94, 96],
      borderColor: '#00BFA5',
      backgroundColor: 'rgba(0, 191, 165, 0.1)',
      borderWidth: 3,
      pointBackgroundColor: '#ffffff',
      pointBorderColor: '#00BFA5',
      pointBorderWidth: 2,
      pointRadius: 5,
      fill: true,
      tension: 0.4
    }]
  };

  const volumeData = {
    labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
    datasets: [{
      label: 'Shipments Tracked',
      data: [1200, 1500, 1400, 1800, 2200, 2500],
      backgroundColor: '#6366f1',
      borderRadius: 4
    }]
  };

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'top',
        labels: { font: { family: "'Outfit', sans-serif" }, usePointStyle: true }
      },
      tooltip: {
        backgroundColor: 'rgba(15, 23, 42, 0.9)',
        titleFont: { family: "'Outfit', sans-serif", size: 13 },
        bodyFont: { family: "'Outfit', sans-serif", size: 14 },
        padding: 10,
        cornerRadius: 8,
        displayColors: false
      }
    },
    scales: {
      y: {
        beginAtZero: chartType === 'volume',
        grid: { color: '#f1f5f9', drawBorder: false },
        ticks: { font: { family: "'Outfit', sans-serif" }, color: '#64748b' }
      },
      x: {
        grid: { display: false, drawBorder: false },
        ticks: { font: { family: "'Outfit', sans-serif" }, color: '#64748b' }
      }
    },
    interaction: {
      mode: 'index',
      intersect: false,
    },
  };

  return (
    <div className="bg-white min-h-screen">
      {/* ═══ HERO ═══ */}
      <section className="relative pt-32 pb-16 lg:pt-48 lg:pb-24 overflow-hidden">
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[800px] bg-brand-teal opacity-5 rounded-full blur-3xl -z-10"></div>
        <div className="absolute top-20 right-0 w-[400px] h-[400px] bg-brand-indigo opacity-5 rounded-full blur-3xl -z-10"></div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10 text-center">
          <div className="inline-flex items-center justify-center px-4 py-1.5 mb-6 text-sm font-semibold text-brand-teal bg-brand-teal/10 rounded-full border border-brand-teal/20">
            <span className="mr-2">📍</span> Real-Time Visibility
          </div>
          <h1 className="text-5xl md:text-6xl lg:text-7xl font-extrabold text-brand-navy tracking-tight mb-6 leading-tight">
            Global <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand-teal to-blue-600">Shipment Tracking</span>
          </h1>
          <p className="mt-4 text-xl md:text-2xl text-ui-gray max-w-3xl mx-auto font-light leading-relaxed">
            Track trucks via <strong className="font-medium text-brand-navy">GPS</strong>, flights via <strong className="font-medium text-brand-navy">FlightAware</strong>, and vessels via <strong className="font-medium text-brand-navy">AIS</strong> — all on one unified dashboard with a modern card-based UI.
          </p>

          <div className="mt-10 flex flex-col sm:flex-row justify-center gap-4">
            <a href="#demo-dashboard" className="px-8 py-4 text-white font-semibold text-lg bg-gradient-to-r from-brand-teal to-blue-500 rounded-full shadow-lg shadow-teal-500/30 hover:shadow-teal-500/50 hover:-translate-y-1 transition-all flex items-center justify-center">
              See Live Demo
            </a>
            <a href="#modes" className="px-8 py-4 text-brand-navy font-semibold text-lg bg-white border border-gray-200 rounded-full hover:bg-gray-50 hover:border-gray-300 transition-all flex items-center justify-center">
              Explore Modes ↓
            </a>
          </div>
        </div>
      </section>

      {/* ═══ STAT BAR ═══ */}
      <Reveal>
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 -mt-8 relative z-20">
          <div className="bg-white rounded-2xl border border-ui-border p-6 flex flex-wrap justify-center divide-y md:divide-y-0 md:divide-x divide-gray-100" style={{ boxShadow: '0 10px 40px -10px rgba(0,0,0,0.05)' }}>
            {[
              { icon: '🚛 ✈️ 🚢', title: '3 Modes', sub: 'Road + Air + Sea' },
              { icon: '⏱️', title: 'Real-time', sub: 'GPS / AIS / FlightAware' },
              { icon: '📸', title: 'Instant POD', sub: 'Digital Proof of Delivery' },
              { icon: '📱', title: 'Auto-Alerts', sub: 'WhatsApp & Email' },
            ].map((s, i) => (
              <div key={i} className="flex-1 min-w-[200px] text-center p-4">
                <div className="text-3xl mb-2">{s.icon}</div>
                <div className="font-bold text-brand-navy text-lg">{s.title}</div>
                <div className="text-sm text-ui-gray mt-1">{s.sub}</div>
              </div>
            ))}
          </div>
        </div>
      </Reveal>

      {/* ═══ TRACKING BY MODE ═══ */}
      <section id="modes" className="py-24 bg-white relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <Reveal>
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold text-brand-navy mb-4">Tracking by Transport Mode</h2>
              <p className="text-lg text-ui-gray max-w-2xl mx-auto">
                Purpose-built integrations for every leg of your supply chain, delivering milestone accuracy.
              </p>
            </div>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {modes.map((m, i) => (
              <Reveal key={i} delay={`reveal-delay-${i + 1}`}>
                <div className={`bg-ui-surface rounded-[20px] p-8 border border-ui-border transition-all duration-300 hover:shadow-xl group ${m.hoverClass}`}>
                  {i === 1 && <div className="absolute top-0 right-0 w-32 h-32 bg-blue-500/5 rounded-full blur-2xl -mr-10 -mt-10"></div>}
                  {i === 2 && <div className="absolute bottom-0 left-0 w-full h-1 bg-gradient-to-r from-brand-teal to-blue-500 transform scale-x-0 group-hover:scale-x-100 transition-transform origin-left duration-500"></div>}
                  
                  <div className="w-16 h-16 bg-white rounded-2xl shadow-sm flex items-center justify-center text-3xl mb-6 group-hover:scale-110 transition-transform relative z-10">
                    {m.icon}
                  </div>
                  <h3 className="text-2xl font-bold text-brand-navy mb-3 relative z-10">{m.title}</h3>
                  <p className="text-ui-gray mb-6 text-sm relative z-10">{m.desc}</p>
                  
                  <ul className="space-y-3 text-slate-700 relative z-10">
                    {m.features.map((f, j) => (
                      <li key={j} className="flex items-start text-sm font-medium">
                        <span style={{ color: m.color }} className="mr-2 font-bold">✓</span>
                        <span>{f}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ PLATFORM FEATURES (Interactive) ═══ */}
      <section className="py-24 bg-slate-50 border-y border-ui-border">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col md:flex-row gap-12 items-start">
            
            {/* Left: Feature Selection */}
            <div className="w-full md:w-1/2">
              <Reveal>
                <div className="mb-8">
                  <p className="text-brand-teal font-bold uppercase tracking-wider text-sm mb-2">Platform Capabilities</p>
                  <h2 className="text-3xl md:text-4xl font-bold text-brand-navy">Unified Tracking Features</h2>
                  <p className="mt-4 text-ui-gray">Select a feature below to see how Freel empowers your logistics operations beyond simple location dots.</p>
                </div>
                
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  {Object.entries(featureData).map(([key, f]) => (
                    <button
                      key={key}
                      onClick={() => handleFeatureClick(key)}
                      className={`flex items-center p-4 bg-white rounded-xl text-left transition-all shadow-sm
                        ${activeFeature === key ? 'border-2 border-brand-teal' : 'border border-gray-200 hover:border-brand-teal/50 hover:bg-slate-50'}
                      `}
                    >
                      <span className="text-2xl mr-3">{f.icon}</span> 
                      <span className="font-semibold text-brand-navy">{f.title.split(' ').slice(0, 2).join(' ')}</span>
                    </button>
                  ))}
                </div>
              </Reveal>
            </div>

            {/* Right: Dynamic Content Display */}
            <div className="w-full md:w-1/2">
              <Reveal>
                <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-8 h-full min-h-[400px] flex flex-col justify-center relative overflow-hidden">
                  <div className="absolute top-0 right-0 w-64 h-64 bg-brand-teal opacity-5 rounded-full blur-3xl -z-10 translate-x-1/2 -translate-y-1/2"></div>
                  
                  <div className={`transition-opacity duration-300 ${fadeState === 'in' ? 'opacity-100' : 'opacity-0'}`}>
                    <div className="text-5xl mb-6">{feat.icon}</div>
                    <h3 className="text-2xl font-bold text-brand-navy mb-4">{feat.title}</h3>
                    <p className="text-ui-gray text-lg leading-relaxed mb-8">{feat.desc}</p>
                    <feat.Visual />
                  </div>
                </div>
              </Reveal>
            </div>

          </div>
        </div>
      </section>

      {/* ═══ INTERACTIVE DATA VISUALIZATION ═══ */}
      <section id="demo-dashboard" className="py-24 bg-white relative">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <Reveal>
            <div className="text-center mb-12">
              <span className="bg-indigo-100 text-brand-indigo text-xs font-bold px-3 py-1 rounded-full uppercase tracking-wide">
                Live Dashboard Demo
              </span>
              <h2 className="text-3xl md:text-4xl font-bold text-brand-navy mt-4 mb-4">Tracking Analytics at a Glance</h2>
              <p className="text-lg text-ui-gray max-w-2xl mx-auto">
                Make data-driven decisions. Monitor shipment volumes, exception rates, and on-time performance across all transport modes.
              </p>
            </div>
          </Reveal>

          <Reveal>
            <div className="relative w-full max-w-[800px] mx-auto h-[60vh] max-h-[450px] bg-white rounded-2xl p-4 shadow-lg border border-ui-border">
              <div className="absolute top-4 right-4 z-10 flex gap-2">
                <button 
                  onClick={() => setChartType('volume')}
                  className={`px-3 py-1 text-sm rounded transition font-medium ${chartType === 'volume' ? 'bg-brand-teal text-white shadow-sm' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'}`}
                >
                  Volume
                </button>
                <button 
                  onClick={() => setChartType('performance')}
                  className={`px-3 py-1 text-sm rounded transition font-medium ${chartType === 'performance' ? 'bg-brand-teal text-white shadow-sm' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'}`}
                >
                  Performance
                </button>
              </div>
              
              <div style={{ position: 'relative', height: '100%', width: '100%', paddingTop: '40px' }}>
                {chartType === 'performance' ? (
                  <Line options={chartOptions} data={performanceData} />
                ) : (
                  <Bar options={chartOptions} data={volumeData} />
                )}
              </div>
            </div>
            <p className="text-center text-sm text-gray-500 mt-6">
              Interactive demo: Toggle between metrics using the buttons on the chart.
            </p>
          </Reveal>
        </div>
      </section>

      {/* ═══ CTA SECTION ═══ */}
      <section className="py-12 pb-24">
        <Reveal>
          <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="bg-brand-navy rounded-[2rem] p-10 md:p-16 text-center relative overflow-hidden shadow-2xl">
              <div className="absolute top-0 left-0 w-full h-full overflow-hidden z-0">
                <div className="absolute -top-24 -right-24 w-96 h-96 bg-brand-teal rounded-full mix-blend-screen filter blur-3xl opacity-20"></div>
                <div className="absolute -bottom-24 -left-24 w-96 h-96 bg-brand-indigo rounded-full mix-blend-screen filter blur-3xl opacity-20"></div>
              </div>

              <div className="relative z-10">
                <h2 className="text-3xl md:text-5xl font-bold text-white mb-6">Track Every Shipment in Real-Time</h2>
                <p className="text-slate-300 text-lg mb-10 max-w-2xl mx-auto font-light">
                  Join modern logistics teams using Freel to unify their visibility, reduce ETA calls by 70%, and improve customer satisfaction.
                </p>
                <Link to="/contact" className="px-8 py-4 text-brand-navy font-bold text-lg bg-white rounded-full shadow-lg hover:shadow-xl hover:scale-105 transition-all outline-none flex items-center justify-center mx-auto group w-fit">
                  Start Tracking <span className="ml-2 group-hover:translate-x-1 transition-transform">→</span>
                </Link>
              </div>
            </div>
          </div>
        </Reveal>
      </section>
    </div>
  );
}
