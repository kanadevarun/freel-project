import React, { useEffect } from 'react';
import { Link } from 'react-router-dom';
import EssentialDocuments from './EssentialDocuments';
import DocumentCategories from './DocumentCategories';
import DocumentJourney from './DocumentJourney';
import './Page.css';

export default function DocumentationGuidePage() {
  useEffect(() => { window.scrollTo(0, 0); }, []);

  return (
    <div className="dg-page">

      {/* ── BREADCRUMB ──────────────────────────────────── */}
      <nav className="dg-breadcrumb">
        <div className="dg-bc-inner">
          <Link to="/">Home</Link>
          <span className="dg-bc-sep">›</span>
          <Link to="/knowledge">Trade Intelligence</Link>
          <span className="dg-bc-sep">›</span>
          <span className="dg-bc-active">Documentation</span>
        </div>
      </nav>

      {/* ── FULL IMAGE HERO SECTION ─────────────────────── */}
      <section className="dg-hero-full-image">
         <img 
            src="/images/documentation/hero-section-documentation-picture-cropped.png" 
            alt="Documentation Hero"
         />
      </section>

      {/* ── ESSENTIAL DOCUMENTS SECTION ─────────────────── */}
      <EssentialDocuments />

      {/* ── DOCUMENT CATEGORIES SECTION ─────────────────── */}
      <DocumentCategories />

      {/* ── DOCUMENT JOURNEY SECTION ────────────────────── */}
      <DocumentJourney />

    </div>
  );
}
