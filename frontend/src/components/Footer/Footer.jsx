import { Link } from 'react-router-dom';
import LogisticsHQLogo from '../Brand/LogisticsHQLogo';
import './Footer.css';

const footerData = {
  services: [
    { label: 'Air Freight', path: '/services/air-freight' },
    { label: 'Sea Freight', path: '/services/sea-freight' },
    { label: 'Road Transport', path: '/services/road-transport' },
    { label: 'Customs', path: '/services/customs' },
  ],
  solutions: [
    { label: 'RFQ Management', path: '/solutions/rfq' },
    { label: 'Rate Comparison', path: '/solutions/rate-comparison' },
    { label: 'Shipment Tracking', path: '/solutions/tracking' },
    { label: 'Compliance', path: '/solutions/compliance' },
  ],
  company: [
    { label: 'About Us', path: '/about' },
    { label: 'Get Started', path: '/signup' },
    { label: 'Resources', path: '/blog' },
    { label: 'Contact', path: '/contact' },
  ],
  resources: [
    { label: 'Weight Calculator', path: '/resources' },
    { label: 'Help Center', path: '/resources' },
    { label: 'API Docs', path: '/resources' },
    { label: 'Blog', path: '/blog' },
  ],
};

export default function Footer() {
  return (
    <footer className="footer" id="site-footer">
      <div className="footer-container">
        <div className="footer-grid">
          {/* Brand Column */}
          <div className="footer-brand">
            <LogisticsHQLogo variant="footer" linkTo="/" className="footer-logo" />
            <p className="footer-tagline">
              The multi-modal logistics operating system. Your Freight Command Center — Road • Air • Sea.
            </p>
          </div>

          {/* Services */}
          <div>
            <h4 className="footer-heading">Services</h4>
            <ul className="footer-link-list">
              {footerData.services.map((link, i) => (
                <li key={i}><Link to={link.path} className="footer-link">{link.label}</Link></li>
              ))}
            </ul>
          </div>

          {/* Solutions */}
          <div>
            <h4 className="footer-heading">Solutions</h4>
            <ul className="footer-link-list">
              {footerData.solutions.map((link, i) => (
                <li key={i}><Link to={link.path} className="footer-link">{link.label}</Link></li>
              ))}
            </ul>
          </div>

          {/* Company */}
          <div>
            <h4 className="footer-heading">Company</h4>
            <ul className="footer-link-list">
              {footerData.company.map((link, i) => (
                <li key={i}><Link to={link.path} className="footer-link">{link.label}</Link></li>
              ))}
            </ul>
          </div>

          {/* Resources */}
          <div>
            <h4 className="footer-heading">Resources</h4>
            <ul className="footer-link-list">
              {footerData.resources.map((link, i) => (
                <li key={i}><Link to={link.path} className="footer-link">{link.label}</Link></li>
              ))}
            </ul>
          </div>
        </div>

        {/* Bottom Row */}
        <div className="footer-bottom">
          <div className="footer-copyright">
            © {new Date().getFullYear()} LogisticsHQ. All rights reserved.
          </div>
          <div className="footer-socials">
            <a href="https://twitter.com" target="_blank" rel="noopener noreferrer" className="social-link">X (Twitter)</a>
            <a href="https://linkedin.com" target="_blank" rel="noopener noreferrer" className="social-link">In (LinkedIn)</a>
            <a href="https://facebook.com" target="_blank" rel="noopener noreferrer" className="social-link">Fb (Facebook)</a>
          </div>
        </div>
      </div>
    </footer>
  );
}
