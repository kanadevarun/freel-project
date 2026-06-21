import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import './ComingSoonPage.css';

/**
 * ComingSoonPage — accepts props for context OR reads from URL params
 *
 * Props:
 *  title    – e.g. "Sea Freight Guide"
 *  icon     – emoji e.g. "🚢"
 *  category – e.g. "Guides" | "Calculators" | "References" | "Insights"
 *  backTo   – link href e.g. "/knowledge"
 */
export default function TradeIntelligenceComingSoon({
  title    = 'Coming Soon',
  icon     = '🚀',
  category = 'Trade Intelligence',
  backTo   = '/knowledge',
}) {
  const [notified, setNotified] = useState(false);
  const [email, setEmail]       = useState('');

  useEffect(() => { window.scrollTo(0, 0); }, []);

  const handleNotify = (e) => {
    e.preventDefault();
    if (email.trim()) setNotified(true);
  };

  const upcomingFeatures = [
    { emoji: '📘', label: 'Visual Guides',       desc: 'Interactive deep-dives' },
    { emoji: '🧮', label: 'Smart Calculators',   desc: 'CBM, duty, freight cost' },
    { emoji: '🗺️', label: 'Port & Airport Index',desc: '150+ global locations' },
    { emoji: '📊', label: 'Trade Insights',       desc: 'Daily market intelligence' },
    { emoji: '🛡️', label: 'Compliance Hub',       desc: 'HS codes, dangerous goods' },
    { emoji: '🤝', label: 'Case Studies',         desc: 'Real-world logistics wins' },
  ];

  return (
    <div className="cs-page">

      {/* ── Animated background ── */}
      <div className="cs-bg">
        <div className="cs-bg-grid"/>
        <div className="cs-bg-orb cs-bg-orb--blue"/>
        <div className="cs-bg-orb cs-bg-orb--teal"/>
        <div className="cs-bg-lines">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="cs-bg-line" style={{ '--i': i }}/>
          ))}
        </div>
      </div>

      <div className="cs-inner">

        {/* ── Breadcrumb ── */}
        <motion.div
          className="cs-breadcrumb"
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y:  0  }}
          transition={{ duration: 0.5 }}
        >
          <Link to="/knowledge" className="cs-bc-link">Trade Intelligence</Link>
          <span className="cs-bc-sep">›</span>
          <span className="cs-bc-category">{category}</span>
          <span className="cs-bc-sep">›</span>
          <span className="cs-bc-current">{title}</span>
        </motion.div>

        {/* ── Hero ── */}
        <motion.div
          className="cs-hero"
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y:  0  }}
          transition={{ duration: 0.6, delay: 0.1 }}
        >
          {/* Icon bubble */}
          <div className="cs-icon-ring">
            <div className="cs-icon-inner">
              <span className="cs-icon-emoji">{icon}</span>
            </div>
          </div>

          <div className="cs-badge">
            <span className="cs-badge-dot"/>
            In Development
          </div>

          <h1 className="cs-title">{title}</h1>
          <p className="cs-subtitle">
            We're building this page with the same care and depth as everything else in the
            Freel Trade Intelligence hub. It will be ready soon — check back shortly!
          </p>
        </motion.div>

        {/* ── Notify form ── */}
        <motion.div
          className="cs-notify"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y:  0  }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          {notified ? (
            <div className="cs-notify-success">
              <span>✅</span>
              <span>You're on the list! We'll notify you when <strong>{title}</strong> goes live.</span>
            </div>
          ) : (
            <form className="cs-notify-form" onSubmit={handleNotify}>
              <input
                type="email"
                className="cs-notify-input"
                placeholder="your@email.com"
                value={email}
                onChange={e => setEmail(e.target.value)}
                required
              />
              <button type="submit" className="cs-notify-btn">
                Notify Me →
              </button>
            </form>
          )}
          <p className="cs-notify-hint">No spam. Only a one-time launch email.</p>
        </motion.div>

        {/* ── What's coming across the hub ── */}
        <motion.div
          className="cs-features"
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y:  0  }}
          transition={{ duration: 0.6, delay: 0.3 }}
        >
          <p className="cs-features-label">What's coming across the hub</p>
          <div className="cs-features-grid">
            {upcomingFeatures.map((f, i) => (
              <motion.div
                key={f.label}
                className="cs-feature-card"
                initial={{ opacity: 0, y: 16 }}
                animate={{ opacity: 1, y:  0  }}
                transition={{ duration: 0.4, delay: 0.35 + i * 0.06 }}
              >
                <span className="cs-fc-emoji">{f.emoji}</span>
                <div>
                  <div className="cs-fc-label">{f.label}</div>
                  <div className="cs-fc-desc">{f.desc}</div>
                </div>
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* ── Stats bar ── */}
        <motion.div
          className="cs-stats"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.6, delay: 0.5 }}
        >
          {[
            { value: '11', label: 'Guides Live' },
            { value: '6+', label: 'Calculators' },
            { value: '150+', label: 'Port Profiles' },
            { value: 'Daily', label: 'Market Data' },
          ].map(s => (
            <div key={s.label} className="cs-stat">
              <div className="cs-stat-val">{s.value}</div>
              <div className="cs-stat-label">{s.label}</div>
            </div>
          ))}
        </motion.div>

        {/* ── Actions ── */}
        <motion.div
          className="cs-actions"
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y:  0  }}
          transition={{ duration: 0.5, delay: 0.55 }}
        >
          <Link to={backTo} className="cs-btn-secondary">
            ← Back to {category}
          </Link>
          <Link to="/knowledge/incoterms" className="cs-btn-primary">
            ⭐ See Our Live Guide: Incoterms →
          </Link>
        </motion.div>

      </div>
    </div>
  );
}
