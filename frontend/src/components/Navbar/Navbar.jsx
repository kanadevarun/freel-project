import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
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

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 10);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const toggleMobileDropdown = (name) => {
    setMobileDropdown(mobileDropdown === name ? null : name);
  };

  return (
    <nav className={`navbar ${scrolled ? 'navbar-scrolled' : ''}`} id="main-navbar">
      <div className="navbar-container">
        {/* Logo */}
        <Link to="/" className="navbar-logo">Freel</Link>

        {/* Desktop Links */}
        <div className="navbar-links">
          {/* Services Dropdown */}
          <div className="nav-dropdown-wrapper">
            <Link to="/services" className="nav-link">
              Services <span className="text-xs">▼</span>
            </Link>
            <DropdownMenu items={servicesDropdown} />
          </div>

          {/* Solutions Dropdown */}
          <div className="nav-dropdown-wrapper">
            <Link to="/solutions" className="nav-link">
              Solutions <span className="text-xs">▼</span>
            </Link>
            <DropdownMenu items={solutionsDropdown} />
          </div>

          <Link to="/about" className="nav-link">About Us</Link>
          <Link to="/platform" className="nav-link">Platform</Link>

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
          <Link to="/signup" className="btn-start">Start Free →</Link>
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
            <Link to="/signup" className="btn-start-mobile" onClick={() => setMobileOpen(false)}>Start Free →</Link>
          </div>
        </div>
      )}
    </nav>
  );
}
