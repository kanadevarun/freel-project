import { useState, useEffect } from 'react';
import { Link, useLocation } from 'react-router-dom';
import LogisticsHQLogo from '../Brand/LogisticsHQLogo';
import {
  Plane, Ship, Truck, Train, Scroll, ShieldCheck, Warehouse, FileText,
  Calculator, Timer, Anchor, Globe,
  ClipboardList, Navigation, ShoppingCart, Route, BarChart3, Shield,
  PieChart, FileCheck, Code, Blocks, Network, Puzzle, ArrowRight, Box
} from 'lucide-react';
import './Navbar.css';

const servicesDropdown = [
  { icon: '✈️', title: 'Air Freight', desc: 'Global air cargo via IATA network', path: '/services/air-freight' },
  { icon: '🚢', title: 'Sea Freight', desc: 'FCL/LCL via major shipping lines', path: '/services/sea-freight' },
  { icon: '🚛', title: 'Road Transport', desc: 'FTL/LTL with GPS tracking across India', path: '/services/road-transport' },
  { icon: '📜', title: 'Customs Brokerage', desc: 'End-to-end customs clearance', path: '/services/customs' },
];

const solutionsDropdown = [
  { icon: '📋', title: 'RFQ Management', desc: 'Automate vendor bidding & rate collection', path: '/solutions/rfq' },
  { icon: '📈', title: 'Rate Comparison', desc: 'Compare 500+ vendor rates instantly', path: '/solutions/rate-comparison' },
  { icon: '📍', title: 'Shipment Tracking', desc: 'Real-time GPS, AIS & flight tracking', path: '/solutions/tracking' },
  { icon: '🛡️', title: 'Compliance & KYC', desc: 'HSN, MSDS, and document verification', path: '/solutions/compliance' },
];

const blogsDropdown = [
  { icon: '⚙️', title: 'Engineering Blog', desc: 'Technical deep-dives, architecture & APIs', path: '/blog/engineering' },
  { icon: '🎨', title: 'Design Blog', desc: 'UX patterns, UI systems & product design', path: '/blog/design' },
  { icon: '📊', title: 'Industry Insights', desc: 'Logistics trends, trade data & market analysis', path: '/blog/industry' },
];

function DropdownMenu({ items }) {
  return (
    <div className="dropdown-menu">
      <div className="flex flex-col">
        {items.map((item, i) => (
          <Link key={i} to={item.path} className="dropdown-item">
            <div className="text-xl">{item.icon}</div>
            <div>
              <div className="font-bold text-sm text-brand-navy">{item.title}</div>
              <div className="text-xs text-slate-500 mt-0.5">{item.desc}</div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}

export default function Navbar() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const [mobileDropdown, setMobileDropdown] = useState(null);
  const [scrolled, setScrolled] = useState(false);
  const [megaMenuOpen, setMegaMenuOpen] = useState(false);
  const [servicesOpen, setServicesOpen] = useState(false);
  const [solutionsOpen, setSolutionsOpen] = useState(false);
  const location = useLocation();

  const isLandingPage = location.pathname === '/' || location.pathname.includes('/services/air-freight') || location.pathname.includes('/services/sea-freight') || location.pathname.includes('/services/road-transport') || location.pathname.includes('/solutions/tracking');

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 50);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const toggleMobileDropdown = (name) => {
    setMobileDropdown(mobileDropdown === name ? null : name);
  };

  const navClass = (scrolled || mobileOpen || !isLandingPage) ? 'navbar-scrolled' : '';

  return (
    <nav className={`navbar ${navClass}`} id="main-navbar">
      <div className="navbar-container">
        {/* Logo */}
        <LogisticsHQLogo variant="header" linkTo="/" className="navbar-logo" />

        {/* Desktop Links */}
        <div className="navbar-links">
          {/* Services Dropdown */}
          <div 
            className="mega-dropdown-wrapper nav-dropdown-wrapper"
            onMouseEnter={() => setServicesOpen(true)}
            onMouseLeave={() => setServicesOpen(false)}
          >
            <Link to="/services" className="nav-link">
              Services <span className="text-xs">▼</span>
            </Link>
            
            <div 
              className={`mega-menu-panel ${servicesOpen ? 'open' : ''}`}
              onClick={() => setServicesOpen(false)}
            >
              <div className="mm-layout">
                {/* Left Hero Card */}
                <div className="mm-hero-card services">
                  <div className="mm-hero-img-bg" style={{backgroundImage: "url('/images/hero_cargo_plane_1779795307360.png')"}}></div>
                  <div className="mm-hero-content">
                    <h3>Freight Services</h3>
                    <p>Move cargo globally across air, ocean, road and rail with one logistics operating system.</p>
                    <Link to="/services" className="mm-hero-btn">Explore Services <ArrowRight size={16} /></Link>
                  </div>
                  <div className="mm-hero-stats">
                    <div className="mm-stat"><strong>10M+</strong><span>Shipments</span></div>
                    <div className="mm-stat"><strong>150+</strong><span>Countries</span></div>
                    <div className="mm-stat"><strong>24/7</strong><span>Support</span></div>
                  </div>
                </div>

                {/* Right Columns Grid */}
                <div className="mm-columns">
                  
                  {/* Column 1 */}
                  <div className="mm-col">
                    <div className="mm-col-header"><Box size={16} className="mm-col-icon"/> Core Services</div>
                    
                    <Link to="/services/air-freight" className="mm-item">
                      <div className="mm-item-icon-box"><Plane size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Air Freight</div>
                        <div className="mm-item-desc">Global cargo via IATA network</div>
                      </div>
                    </Link>
                    <Link to="/services/sea-freight" className="mm-item">
                      <div className="mm-item-icon-box"><Ship size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Sea Freight</div>
                        <div className="mm-item-desc">FCL / LCL shipping</div>
                      </div>
                    </Link>
                    <Link to="/services/road-transport" className="mm-item">
                      <div className="mm-item-icon-box"><Truck size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Road Transport</div>
                        <div className="mm-item-desc">FTL / LTL transportation</div>
                      </div>
                    </Link>
                    <Link to="/services/rail-freight" className="mm-item">
                      <div className="mm-item-icon-box"><Train size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Rail Freight</div>
                        <div className="mm-item-desc">Cost-efficient inland freight</div>
                      </div>
                    </Link>
                  </div>

                  {/* Column 2 */}
                  <div className="mm-col">
                    <div className="mm-col-header"><ShieldCheck size={16} className="mm-col-icon" style={{color:'#f59e0b'}}/> Value Added Services</div>
                    
                    <Link to="/services/customs" className="mm-item">
                      <div className="mm-item-icon-box orange"><Scroll size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Customs Brokerage</div>
                        <div className="mm-item-desc">End-to-end customs clearance</div>
                      </div>
                    </Link>
                    <Link to="/services/insurance" className="mm-item">
                      <div className="mm-item-icon-box purple"><ShieldCheck size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Insurance</div>
                        <div className="mm-item-desc">Cargo insurance solutions</div>
                      </div>
                    </Link>
                    <Link to="/services/trade-finance" className="mm-item">
                      <div className="mm-item-icon-box green"><Warehouse size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Trade Finance</div>
                        <div className="mm-item-desc">Freight financing solutions</div>
                      </div>
                    </Link>
                    <Link to="/services/documentation" className="mm-item">
                      <div className="mm-item-icon-box"><FileText size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Documentation</div>
                        <div className="mm-item-desc">Bills of Lading, AWB, Invoices</div>
                      </div>
                    </Link>
                  </div>

                  {/* Column 3 */}
                  <div className="mm-col">
                    <div className="mm-col-header"><BarChart3 size={16} className="mm-col-icon" style={{color:'#8b5cf6'}}/> Tools & Resources</div>
                    
                    <Link to="/tools/freight-cost" className="mm-item">
                      <div className="mm-item-icon-box green"><Calculator size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Freight Cost Calculator</div>
                        <div className="mm-item-desc">Compare live freight rates</div>
                      </div>
                    </Link>
                    <Link to="/tools/transit-time" className="mm-item">
                      <div className="mm-item-icon-box purple"><Timer size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Transit Time Calculator</div>
                        <div className="mm-item-desc">Accurate transit estimates</div>
                      </div>
                    </Link>
                    <Link to="/reference/ports" className="mm-item">
                      <div className="mm-item-icon-box"><Anchor size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Port Directory</div>
                        <div className="mm-item-desc">Global port & terminal data</div>
                      </div>
                    </Link>
                    <Link to="/coverage" className="mm-item">
                      <div className="mm-item-icon-box green"><Globe size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Service Coverage Map</div>
                        <div className="mm-item-desc">Explore our global network</div>
                      </div>
                    </Link>
                  </div>

                </div>
              </div>
              
              {/* Bottom Strip */}
              <div className="mm-bottom-strip">
                <div className="mm-bottom-content">
                  <h4><Globe size={18} style={{color:'#2563eb'}}/> Reliable. Transparent. Efficient.</h4>
                  <p>Move your cargo safely and on time through a unified freight operating platform.</p>
                </div>
                <Link to="/quote" className="mm-bottom-btn">Get A Quote <ArrowRight size={16} style={{display:'inline', verticalAlign:'middle', marginLeft:'4px'}}/></Link>
              </div>
            </div>
          </div>

          {/* Solutions Dropdown */}
          <div 
            className="mega-dropdown-wrapper nav-dropdown-wrapper"
            onMouseEnter={() => setSolutionsOpen(true)}
            onMouseLeave={() => setSolutionsOpen(false)}
          >
            <Link to="/solutions" className="nav-link">
              Solutions <span className="text-xs">▼</span>
            </Link>
            
            <div 
              className={`mega-menu-panel ${solutionsOpen ? 'open' : ''}`}
              onClick={() => setSolutionsOpen(false)}
            >
              <div className="mm-layout">
                {/* Left Hero Card */}
                <div className="mm-hero-card solutions">
                  <div className="mm-hero-img-bg" style={{backgroundImage: "url('/images/hero_logistics_hub_1779795325054.png')"}}></div>
                  <div className="mm-hero-content">
                    <h3>Logistics Solutions</h3>
                    <p>Optimize procurement, visibility, compliance and analytics from a single platform.</p>
                    <Link to="/solutions" className="mm-hero-btn">Explore Solutions <ArrowRight size={16} /></Link>
                  </div>
                  <div className="mm-hero-stats">
                    <div className="mm-stat"><strong>500+</strong><span>Customers</span></div>
                    <div className="mm-stat"><strong>30%</strong><span>Cost Savings</span></div>
                    <div className="mm-stat"><strong>99.9%</strong><span>Uptime</span></div>
                  </div>
                </div>

                {/* Right Columns Grid */}
                <div className="mm-columns">
                  
                  {/* Column 1 */}
                  <div className="mm-col">
                    <div className="mm-col-header"><ClipboardList size={16} className="mm-col-icon" style={{color:'#10b981'}}/> Operations</div>
                    
                    <Link to="/solutions/rfq" className="mm-item">
                      <div className="mm-item-icon-box"><ClipboardList size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">RFQ Management</div>
                        <div className="mm-item-desc">Automated vendor bidding</div>
                      </div>
                    </Link>
                    <Link to="/solutions/tracking" className="mm-item">
                      <div className="mm-item-icon-box orange"><Navigation size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Shipment Tracking</div>
                        <div className="mm-item-desc">Real-time visibility</div>
                      </div>
                    </Link>
                    <Link to="/solutions/procurement" className="mm-item">
                      <div className="mm-item-icon-box green"><ShoppingCart size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Procurement</div>
                        <div className="mm-item-desc">Supplier management</div>
                      </div>
                    </Link>
                    <Link to="/solutions/route" className="mm-item">
                      <div className="mm-item-icon-box purple"><Route size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Route Optimization</div>
                        <div className="mm-item-desc">Reduce transit cost</div>
                      </div>
                    </Link>
                  </div>

                  {/* Column 2 */}
                  <div className="mm-col">
                    <div className="mm-col-header"><BarChart3 size={16} className="mm-col-icon" style={{color:'#3b82f6'}}/> Intelligence</div>
                    
                    <Link to="/solutions/rate-comparison" className="mm-item">
                      <div className="mm-item-icon-box green"><BarChart3 size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Rate Comparison</div>
                        <div className="mm-item-desc">Compare vendor rates</div>
                      </div>
                    </Link>
                    <Link to="/solutions/compliance" className="mm-item">
                      <div className="mm-item-icon-box purple"><Shield size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Compliance</div>
                        <div className="mm-item-desc">HSN & document check</div>
                      </div>
                    </Link>
                    <Link to="/solutions/analytics" className="mm-item">
                      <div className="mm-item-icon-box"><PieChart size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Analytics</div>
                        <div className="mm-item-desc">Data-driven decisions</div>
                      </div>
                    </Link>
                    <Link to="/solutions/reporting" className="mm-item">
                      <div className="mm-item-icon-box orange"><FileCheck size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Reporting</div>
                        <div className="mm-item-desc">Custom reports & exports</div>
                      </div>
                    </Link>
                  </div>

                  {/* Column 3 */}
                  <div className="mm-col">
                    <div className="mm-col-header"><Blocks size={16} className="mm-col-icon" style={{color:'#8b5cf6'}}/> Integrations</div>
                    
                    <Link to="/solutions/api" className="mm-item">
                      <div className="mm-item-icon-box purple"><Code size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">API Access</div>
                        <div className="mm-item-desc">Developer APIs</div>
                      </div>
                    </Link>
                    <Link to="/solutions/erp" className="mm-item">
                      <div className="mm-item-icon-box green"><Blocks size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">ERP Integration</div>
                        <div className="mm-item-desc">SAP, Oracle, Dynamics</div>
                      </div>
                    </Link>
                    <Link to="/solutions/webhooks" className="mm-item">
                      <div className="mm-item-icon-box"><Network size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">Webhooks</div>
                        <div className="mm-item-desc">Real-time data push</div>
                      </div>
                    </Link>
                    <Link to="/solutions/edi" className="mm-item">
                      <div className="mm-item-icon-box orange"><Puzzle size={20} /></div>
                      <div className="mm-item-content">
                        <div className="mm-item-title">EDI Support</div>
                        <div className="mm-item-desc">Legacy system support</div>
                      </div>
                    </Link>
                  </div>

                </div>
              </div>
              
              {/* Bottom Strip */}
              <div className="mm-bottom-strip purple">
                <div className="mm-bottom-content">
                  <h4><Shield size={18} style={{color:'#7c3aed'}}/> Smarter Decisions. Better Outcomes.</h4>
                  <p>Use real-time intelligence to optimize freight procurement, tracking and compliance.</p>
                </div>
                <Link to="/demo" className="mm-bottom-btn">Request Demo <ArrowRight size={16} style={{display:'inline', verticalAlign:'middle', marginLeft:'4px'}}/></Link>
              </div>
            </div>
          </div>

          <Link to="/platform" className="nav-link">Platform</Link>

          {/* Trade Intelligence Mega Menu */}
          <div 
            className="mega-dropdown-wrapper nav-dropdown-wrapper"
            onMouseEnter={() => setMegaMenuOpen(true)}
            onMouseLeave={() => setMegaMenuOpen(false)}
          >
            <Link to="/trade-intelligence/coming-soon" className="nav-link">
              Trade Intelligence <span className="text-xs">▼</span>
            </Link>
            
            <div 
              className={`trade-intelligence-menu ${megaMenuOpen ? 'open' : ''}`}
              onClick={() => setMegaMenuOpen(false)}
            >
              
              {/* Subtle Map Background Pattern */}
              <div className="ti-bg-layer" />

              <div className="ti-main-layout">
                {/* Hero Section */}
                <div className="ti-hero-section">
                  <div className="ti-hero-content">
                    <h3>Trade Intelligence</h3>
                    <p>Learn logistics through visual guides, interactive tools, calculators, and industry knowledge.</p>
                    <Link to="/trade-intelligence/coming-soon" className="ti-hero-btn">Explore Knowledge Hub &rarr;</Link>
                  </div>
                  <div className="ti-hero-stats">
                    <div className="ti-stat"><strong>100+</strong><span>Guides</span></div>
                    <div className="ti-stat"><strong>25+</strong><span>Tools</span></div>
                    <div className="ti-stat"><strong>150+</strong><span>Ports</span></div>
                    <div className="ti-stat"><strong>50K+</strong><span>Monthly Readers</span></div>
                  </div>
                </div>

                {/* Right Columns Grid */}
                <div className="ti-grid-columns">
                  
                  {/* Column 1 */}
                  <div className="ti-col">
                    <div className="ti-col-header">
                      <span className="ti-icon-large">📚</span> Guides
                    </div>
                    
                    <div className="ti-popular-label">Most Popular</div>
                    <div className="ti-popular-group">
                      <Link to="/knowledge/incoterms" className="ti-popular-item">⭐ Incoterms Guide</Link>
                      <Link to="/knowledge/air-freight" className="ti-popular-item">⭐ Air Freight Guide</Link>
                      <Link to="/knowledge/customs" className="ti-popular-item">⭐ Customs Clearance</Link>
                    </div>

                    <Link to="/knowledge/sea-freight" className="ti-item">
                      <span className="ti-item-icon">🚢</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Sea Freight</div>
                        <div className="ti-item-desc">FCL and LCL shipping explained.</div>
                      </div>
                    </Link>
                    <Link to="/knowledge/documentation" className="ti-item">
                      <span className="ti-item-icon">📄</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Documentation</div>
                        <div className="ti-item-desc">Bill of Lading, AWB, Invoices.</div>
                      </div>
                    </Link>
                    <Link to="/knowledge/import-export" className="ti-item">
                      <span className="ti-item-icon">📦</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Import Export Basics</div>
                        <div className="ti-item-desc">Getting started with global trade.</div>
                      </div>
                    </Link>
                  </div>

                  {/* Column 2 */}
                  <div className="ti-col">
                    <div className="ti-col-header">
                      <span className="ti-icon-large">🧮</span> Calculators
                    </div>
                    <Link to="/tools/cbm-calculator" className="ti-item">
                      <span className="ti-item-icon">📐</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">CBM Calculator <span className="ti-pill-live"><span className="ti-pulse"></span>Live Tool</span></div>
                      </div>
                    </Link>
                    <Link to="/tools/volumetric-weight" className="ti-item">
                      <span className="ti-item-icon">⚖️</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Volumetric Weight <span className="ti-pill-live"><span className="ti-pulse"></span>Live Tool</span></div>
                      </div>
                    </Link>
                    <Link to="/tools/duty-calculator" className="ti-item">
                      <span className="ti-item-icon">💰</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Duty Calculator <span className="ti-pill-live"><span className="ti-pulse"></span>Live Tool</span></div>
                      </div>
                    </Link>
                    <Link to="/tools/transit-time" className="ti-item">
                      <span className="ti-item-icon">⏱️</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Transit Time</div>
                      </div>
                    </Link>
                    <Link to="/tools/freight-cost" className="ti-item">
                      <span className="ti-item-icon">💲</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Freight Cost</div>
                      </div>
                    </Link>
                    <Link to="/tools/container-load" className="ti-item">
                      <span className="ti-item-icon">🏗️</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Container Load</div>
                      </div>
                    </Link>
                  </div>

                  {/* Column 3 */}
                  <div className="ti-col">
                    <div className="ti-col-header">
                      <span className="ti-icon-large">🌍</span> References
                    </div>
                    <div className="ti-ref-grid">
                      <Link to="/reference/container-sizes" className="ti-ref-card">
                        <div className="ti-ref-title">Container Sizes</div>
                        <div className="ti-ref-desc">20ft • 40ft</div>
                      </Link>
                      <Link to="/reference/ports" className="ti-ref-card">
                        <div className="ti-ref-title">Port Directory</div>
                        <div className="ti-ref-desc">150+ Ports</div>
                      </Link>
                      <Link to="/reference/airports" className="ti-ref-card">
                        <div className="ti-ref-title">Airports</div>
                        <div className="ti-ref-desc">IATA Codes</div>
                      </Link>
                      <Link to="/reference/hsn-codes" className="ti-ref-card">
                        <div className="ti-ref-title">HSN Codes</div>
                        <div className="ti-ref-desc">HS Database</div>
                      </Link>
                      <Link to="/reference/dangerous-goods" className="ti-ref-card">
                        <div className="ti-ref-title">Dangerous Goods</div>
                        <div className="ti-ref-desc">IMDG Rules</div>
                      </Link>
                      <Link to="/reference/trade-profiles" className="ti-ref-card">
                        <div className="ti-ref-title">Trade Profiles</div>
                        <div className="ti-ref-desc">By Country</div>
                      </Link>
                    </div>
                  </div>

                  {/* Column 4 */}
                  <div className="ti-col">
                    <div className="ti-col-header">
                      <span className="ti-icon-large">📈</span> 
                      <div>
                        Insights
                        <div className="ti-col-subtitle">Updated Daily</div>
                      </div>
                    </div>
                    <Link to="/insights/trends" className="ti-item">
                      <span className="ti-item-icon">📊</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Logistics Trends</div>
                      </div>
                    </Link>
                    <Link to="/insights/reports" className="ti-item">
                      <span className="ti-item-icon">📑</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Supply Chain Reports</div>
                      </div>
                    </Link>
                    <Link to="/insights/market-updates" className="ti-item">
                      <span className="ti-item-icon">⚡</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Market Updates <span className="ti-badge-red">NEW</span></div>
                      </div>
                    </Link>
                    <Link to="/insights/benchmarks" className="ti-item">
                      <span className="ti-item-icon">🎯</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Industry Benchmarks</div>
                      </div>
                    </Link>
                    <Link to="/insights/news" className="ti-item">
                      <span className="ti-item-icon">📰</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Trade News <span className="ti-badge-red">NEW</span></div>
                      </div>
                    </Link>
                    <Link to="/insights/cases" className="ti-item">
                      <span className="ti-item-icon">💼</span>
                      <div className="ti-item-content">
                        <div className="ti-item-title">Case Studies</div>
                      </div>
                    </Link>
                  </div>
                  
                </div>
              </div>
              
              {/* Bottom CTA */}
              <div className="ti-bottom-cta">
                <div className="ti-bottom-content">
                  <h3>Master Logistics Like A Pro</h3>
                  <p>Access visual guides, calculators, trade references, and logistics insights used by modern supply chain teams.</p>
                </div>
                <Link to="/trade-intelligence/coming-soon" className="ti-bottom-btn">Start Learning &rarr;</Link>
              </div>

            </div>
          </div>

          <Link to="/about" className="nav-link">About Us</Link>

          {/* Blogs Dropdown */}
          <div className="nav-dropdown-wrapper">
            <Link to="/blog" className="nav-link">
              Blogs <span className="text-xs">▼</span>
            </Link>
            <DropdownMenu items={blogsDropdown} />
          </div>

          <Link to="/contact" className="nav-link">Contact</Link>
        </div>

        {/* CTA Buttons */}
        <div className="navbar-actions">
          <Link to="/login" className="btn-login">Log In</Link>
          <Link to="/signup" className="btn-start">Get Started →</Link>
          {/* Mobile Hamburger */}
          <button
            className="mobile-toggle"
            onClick={() => setMobileOpen(!mobileOpen)}
            aria-label="Toggle navigation menu"
            id="mobile-menu-toggle"
          >
            <span className={`hamburger-line ${mobileOpen ? 'rotate-45 translate-y-[7px]' : ''}`}></span>
            <span className={`hamburger-line ${mobileOpen ? 'opacity-0' : ''}`}></span>
            <span className={`hamburger-line ${mobileOpen ? '-rotate-45 -translate-y-[7px]' : ''}`}></span>
          </button>
        </div>
      </div>

      {/* Mobile Drawer */}
      {mobileOpen && (
        <div className="mobile-drawer" id="mobile-drawer">
          <div className="mobile-nav-section">
            <button onClick={() => toggleMobileDropdown('services')} className="mobile-nav-trigger">
              Services <span className="text-xs">{mobileDropdown === 'services' ? '▲' : '▼'}</span>
            </button>
            {mobileDropdown === 'services' && (
              <div className="mobile-dropdown-content">
                {servicesDropdown.map((item, i) => (
                  <Link key={i} to={item.path} className="mobile-dropdown-item" onClick={() => setMobileOpen(false)}>
                    <span>{item.icon}</span> {item.title}
                  </Link>
                ))}
              </div>
            )}
          </div>

          <div className="mobile-nav-section">
            <button onClick={() => toggleMobileDropdown('solutions')} className="mobile-nav-trigger">
              Solutions <span className="text-xs">{mobileDropdown === 'solutions' ? '▲' : '▼'}</span>
            </button>
            {mobileDropdown === 'solutions' && (
              <div className="mobile-dropdown-content">
                {solutionsDropdown.map((item, i) => (
                  <Link key={i} to={item.path} className="mobile-dropdown-item" onClick={() => setMobileOpen(false)}>
                    <span>{item.icon}</span> {item.title}
                  </Link>
                ))}
              </div>
            )}
          </div>

          <Link to="/about" className="mobile-nav-link" onClick={() => setMobileOpen(false)}>About Us</Link>
          <Link to="/platform" className="mobile-nav-link" onClick={() => setMobileOpen(false)}>Platform</Link>

          <div className="mobile-nav-section">
            <button onClick={() => toggleMobileDropdown('blogs')} className="mobile-nav-trigger">
              Blogs <span className="text-xs">{mobileDropdown === 'blogs' ? '▲' : '▼'}</span>
            </button>
            {mobileDropdown === 'blogs' && (
              <div className="mobile-dropdown-content">
                {blogsDropdown.map((item, i) => (
                  <Link key={i} to={item.path} className="mobile-dropdown-item" onClick={() => setMobileOpen(false)}>
                    <span>{item.icon}</span> {item.title}
                  </Link>
                ))}
              </div>
            )}
          </div>

          <Link to="/contact" className="mobile-nav-link" onClick={() => setMobileOpen(false)}>Contact</Link>

          <div className="mobile-nav-cta">
            <Link to="/login" className="btn-login-mobile" onClick={() => setMobileOpen(false)}>Log In</Link>
            <Link to="/signup" className="btn-start-mobile" onClick={() => setMobileOpen(false)}>Get Started →</Link>
          </div>
        </div>
      )}
    </nav>
  );
}
