import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import PublicLayout from './layouts/PublicLayout/PublicLayout';
import Landing from './pages/public/Landing/Landing';
import Services from './pages/public/Services/Services';
import AirFreight from './pages/public/Services/AirFreight';
import SeaFreight from './pages/public/Services/SeaFreight';
import RoadTransport from './pages/public/Services/RoadTransport';
import CustomsBrokerage from './pages/public/Services/CustomsBrokerage';
import Solutions from './pages/public/Solutions/Solutions';
import RFQLanding from './pages/public/Solutions/RFQLanding';
import RateComparison from './pages/public/Solutions/RateComparison';
import ShipmentTracking from './pages/public/Solutions/ShipmentTracking';
import Compliance from './pages/public/Solutions/Compliance';

// Trade Intelligence - Guides
import IncotermsPage from './pages/public/trade-intelligence/guides/incoterms/Page';
import AirFreightPage from './pages/public/trade-intelligence/guides/air-freight/Page';
// Other Trade Intelligence routes map directly to ComingSoonPage with props.

// Trade Intelligence - Coming Soon
import ComingSoonPage from './pages/public/trade-intelligence/coming-soon/ComingSoonPage';

import BlogIndex from './pages/public/Blog/BlogIndex';
import EngineeringBlog from './pages/public/Blog/EngineeringBlog';
import DesignBlog from './pages/public/Blog/DesignBlog';
import IndustryBlog from './pages/public/Blog/IndustryBlog';
import About from './pages/public/About/About';
import Contact from './pages/public/Contact/Contact';
import Platform from './pages/public/Platform/Platform';
import ScrollToTop from './components/ScrollToTop';
import { Link } from 'react-router-dom';
import './App.css';


/**
 * App.jsx — Root component with routing.
 *
 * All public pages are wrapped in PublicLayout (Navbar + Footer).
 * Each Route maps a URL path to a page component.
 *
 * Phase 2 note: /login and /signup redirect to /contact until auth is built.
 */


/** Branded placeholder for pages coming in Phase 2/3 */
function PlaceholderPage({ title, emoji, note = 'This page is being built and will be available soon.' }) {
  return (
    <section className="section-padding radial-glow-top">
      <div className="container-sm text-center">
        <div style={{ fontSize: '4rem', marginBottom: '24px' }}>{emoji}</div>
        <div className="section-label section-label-teal" style={{ marginBottom: '16px' }}>Coming Soon</div>
        <h1 style={{ fontSize: '2rem', fontWeight: 800, color: '#1E293B', marginBottom: '12px' }}>{title}</h1>
        <p style={{ color: '#64748B', maxWidth: '420px', margin: '0 auto 32px', lineHeight: 1.7 }}>{note}</p>
        <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', flexWrap: 'wrap' }}>
          <Link to="/" className="btn-primary" style={{ textDecoration: 'none' }}>← Back to Home</Link>
          <Link to="/contact" className="btn-secondary" style={{ textDecoration: 'none' }}>Contact Us</Link>
        </div>
      </div>
    </section>
  );
}

/** 404 Not Found page */
function NotFoundPage() {
  return (
    <section className="section-padding">
      <div className="container-sm text-center">
        <div style={{ fontSize: '5rem', fontWeight: 900, color: '#E2E8F0', marginBottom: '8px', lineHeight: 1 }}>404</div>
        <div className="section-label section-label-slate" style={{ marginBottom: '16px' }}>Page Not Found</div>
        <h1 style={{ fontSize: '1.75rem', fontWeight: 800, color: '#1E293B', marginBottom: '12px' }}>
          We couldn't find that page.
        </h1>
        <p style={{ color: '#64748B', maxWidth: '400px', margin: '0 auto 32px', lineHeight: 1.7 }}>
          The URL you visited doesn't exist. It may have moved or been renamed.
        </p>
        <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', flexWrap: 'wrap' }}>
          <Link to="/" className="btn-primary" style={{ textDecoration: 'none' }}>← Go Home</Link>
          <Link to="/services" className="btn-secondary" style={{ textDecoration: 'none' }}>View Services</Link>
        </div>
      </div>
    </section>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <ScrollToTop />
      <Routes>
        {/* All public pages wrapped in PublicLayout (Navbar + Footer) */}
        <Route element={<PublicLayout />}>
          <Route path="/" element={<Landing />} />

          {/* Services */}
          <Route path="/services" element={<Services />} />
          <Route path="/services/air-freight" element={<AirFreight />} />
          <Route path="/services/sea-freight" element={<SeaFreight />} />
          <Route path="/services/road-transport" element={<RoadTransport />} />
          <Route path="/services/customs" element={<CustomsBrokerage />} />
          <Route path="/services/rail-freight" element={<ComingSoonPage title="Rail Freight" icon="🚆" category="Services" />} />
          <Route path="/services/trade-finance" element={<ComingSoonPage title="Trade Finance" icon="🏦" category="Services" />} />
          <Route path="/services/insurance" element={<ComingSoonPage title="Insurance" icon="🛡️" category="Services" />} />
          <Route path="/services/documentation" element={<ComingSoonPage title="Documentation" icon="📄" category="Services" />} />
          <Route path="/coverage" element={<ComingSoonPage title="Coverage Map" icon="🌍" category="Resources" />} />

          {/* Solutions */}
          <Route path="/solutions" element={<Solutions />} />
          <Route path="/solutions/rfq" element={<RFQLanding />} />
          <Route path="/solutions/rate-comparison" element={<RateComparison />} />
          <Route path="/solutions/tracking" element={<ShipmentTracking />} />
          <Route path="/solutions/compliance" element={<Compliance />} />
          <Route path="/solutions/procurement" element={<ComingSoonPage title="Procurement" icon="🛒" category="Solutions" />} />
          <Route path="/solutions/route" element={<ComingSoonPage title="Route Optimization" icon="🛣️" category="Solutions" />} />
          <Route path="/solutions/analytics" element={<ComingSoonPage title="Analytics" icon="📊" category="Solutions" />} />
          <Route path="/solutions/reporting" element={<ComingSoonPage title="Reporting" icon="📑" category="Solutions" />} />
          <Route path="/solutions/api" element={<ComingSoonPage title="API Access" icon="🔗" category="Integrations" />} />
          <Route path="/solutions/erp" element={<ComingSoonPage title="ERP Integration" icon="🔄" category="Integrations" />} />
          <Route path="/solutions/webhooks" element={<ComingSoonPage title="Webhooks" icon="📡" category="Integrations" />} />
          <Route path="/solutions/edi" element={<ComingSoonPage title="EDI Support" icon="🧩" category="Integrations" />} />

          {/* Blog */}
          <Route path="/blog" element={<BlogIndex />} />
          <Route path="/blog/engineering" element={<EngineeringBlog />} />
          <Route path="/blog/design" element={<DesignBlog />} />
          <Route path="/blog/industry" element={<IndustryBlog />} />

          {/* Company */}
          <Route path="/about" element={<About />} />
          <Route path="/contact" element={<Contact />} />
          <Route path="/platform" element={<Platform />} />

          {/* Phase 2 placeholders — redirect auth routes to contact for now */}
          <Route path="/login" element={<Navigate to="/contact" replace />} />
          <Route path="/signup" element={<Navigate to="/contact" replace />} />

          {/* Phase 3 placeholders */}
          <Route path="/products" element={<PlaceholderPage title="Products" emoji="📦" note="Our full product suite page is coming soon. In the meantime, explore our Services." />} />
          <Route path="/partners" element={<PlaceholderPage title="Partners" emoji="🤝" note="Our partner program is launching soon. Contact us to express interest." />} />
          <Route path="/resources" element={<PlaceholderPage title="Resources" emoji="📚" note="Help center, API docs, and weight calculators are coming soon." />} />
          <Route path="/track" element={<PlaceholderPage title="Track Order" emoji="🔍" note="Live shipment tracking is available on the dashboard after signing up." />} />

          {/* Trade Intelligence - Guides */}
          <Route path="/knowledge" element={<ComingSoonPage title="Trade Intelligence Hub" icon="🧠" category="Knowledge Base" />} />
          <Route path="/knowledge/incoterms" element={<IncotermsPage />} />
          <Route path="/knowledge/air-freight" element={<AirFreightPage />} />
          <Route path="/knowledge/sea-freight" element={<ComingSoonPage title="Sea Freight Guide" icon="🚢" category="Guides" />} />
          <Route path="/knowledge/customs" element={<ComingSoonPage title="Customs Clearance Guide" icon="🛡️" category="Guides" />} />
          <Route path="/knowledge/documentation" element={<ComingSoonPage title="Documentation Guide" icon="📄" category="Guides" />} />
          <Route path="/knowledge/import-export" element={<ComingSoonPage title="Import & Export Basics" icon="🌐" category="Guides" />} />

          {/* Trade Intelligence - Calculators */}
          <Route path="/tools/cbm-calculator" element={<ComingSoonPage title="CBM Calculator" icon="📦" category="Calculators" />} />
          <Route path="/tools/volumetric-weight" element={<ComingSoonPage title="Volumetric Weight" icon="⚖️" category="Calculators" />} />
          <Route path="/tools/duty-calculator" element={<ComingSoonPage title="Duty Calculator" icon="🧮" category="Calculators" />} />
          <Route path="/tools/transit-time" element={<ComingSoonPage title="Transit Time Estimator" icon="⏱️" category="Calculators" />} />
          <Route path="/tools/freight-cost" element={<ComingSoonPage title="Freight Cost Calculator" icon="💰" category="Calculators" />} />
          <Route path="/tools/container-load" element={<ComingSoonPage title="Container Load Planner" icon="🏗️" category="Calculators" />} />

          {/* Trade Intelligence - References */}
          <Route path="/reference/container-sizes" element={<ComingSoonPage title="Container Sizes Guide" icon="📐" category="References" />} />
          <Route path="/reference/ports" element={<ComingSoonPage title="Port Directory" icon="⚓" category="References" />} />
          <Route path="/reference/airports" element={<ComingSoonPage title="Airport Directory" icon="🛫" category="References" />} />
          <Route path="/reference/hsn-codes" element={<ComingSoonPage title="HSN / HS Codes" icon="🔢" category="References" />} />
          <Route path="/reference/dangerous-goods" element={<ComingSoonPage title="Dangerous Goods Guide" icon="⚠️" category="References" />} />
          <Route path="/reference/trade-profiles" element={<ComingSoonPage title="Country Trade Profiles" icon="🗺️" category="References" />} />

          {/* Trade Intelligence - Insights */}
          <Route path="/insights/trends" element={<ComingSoonPage title="Logistics Trends" icon="📈" category="Insights" />} />
          <Route path="/insights/market-updates" element={<ComingSoonPage title="Market Updates" icon="📰" category="Insights" />} />
          <Route path="/insights/news" element={<ComingSoonPage title="Trade News" icon="🗞️" category="Insights" />} />
          <Route path="/insights/reports" element={<ComingSoonPage title="Industry Reports" icon="📋" category="Insights" />} />
          <Route path="/insights/benchmarks" element={<ComingSoonPage title="Logistics Benchmarks" icon="🏆" category="Insights" />} />
          <Route path="/insights/cases" element={<ComingSoonPage title="Case Studies" icon="🤝" category="Insights" />} />

          {/* Trade Intelligence - Coming Soon Fallback */}
          <Route path="/trade-intelligence/coming-soon" element={<ComingSoonPage />} />

          {/* 404 fallback */}
          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
