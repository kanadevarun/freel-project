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

          {/* Solutions */}
          <Route path="/solutions" element={<Solutions />} />
          <Route path="/solutions/rfq" element={<RFQLanding />} />
          <Route path="/solutions/rate-comparison" element={<RateComparison />} />
          <Route path="/solutions/tracking" element={<ShipmentTracking />} />
          <Route path="/solutions/compliance" element={<Compliance />} />

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

          {/* 404 fallback */}
          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

