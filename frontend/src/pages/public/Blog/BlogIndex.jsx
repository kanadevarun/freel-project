
import { Link } from 'react-router-dom';
import './Blog.css';

export default function BlogIndex() {
  return (
    <div className="bg-ui-surface min-h-screen">
      <section className="blog-index-hero">
        <h1 className="blog-index-hero-title">
          Freel <span className="text-brand-teal">Insights</span>
        </h1>
        <p className="blog-index-hero-sub">
          The latest thoughts on logistics engineering, product design, and industry trends from the team building India's first multi-modal logistics OS.
        </p>
      </section>

      <section className="blog-category-section">
        
        {/* Engineering Section */}
        <div className="mb-20">
          <div className="blog-category-header">
            <h2 className="blog-category-title">
              <span>⚙️</span> Engineering Blog
            </h2>
            <Link to="/blog/engineering" className="blog-view-all">View All →</Link>
          </div>
          <div className="blog-grid" style={{ padding: '0', maxWidth: '100%' }}>
            <article className="blog-card group">
              <div className="blog-card-img engineering-1">
                <span className="blog-card-emoji">🗺️</span>
              </div>
              <div className="blog-card-content">
                <h3 className="blog-card-title">How We Built Real-Time Multi-Modal Tracking</h3>
                <p className="blog-card-desc">A deep-dive into unifying truck GPS, vessel AIS, and flight tracking into a single dashboard.</p>
                <Link to="/blog/engineering" className="text-brand-indigo font-semibold mt-4">Read Article →</Link>
              </div>
            </article>
            <article className="blog-card group">
              <div className="blog-card-img engineering-2">
                <span className="blog-card-emoji">🤖</span>
              </div>
              <div className="blog-card-content">
                <h3 className="blog-card-title">AI Cargo Scanner: From Photo to HS Code</h3>
                <p className="blog-card-desc">How we trained a vision model to estimate cargo dimensions from a single smartphone photo.</p>
                <Link to="/blog/engineering" className="text-brand-indigo font-semibold mt-4">Read Article →</Link>
              </div>
            </article>
          </div>
        </div>

        {/* Industry Section */}
        <div className="mb-20">
          <div className="blog-category-header">
            <h2 className="blog-category-title">
              <span>📊</span> Industry Insights
            </h2>
            <Link to="/blog/industry" className="blog-view-all">View All →</Link>
          </div>
          <div className="blog-grid col-3" style={{ padding: '0', maxWidth: '100%' }}>
            <article className="industry-card group">
              <div className="mb-4 text-4xl">📈</div>
              <h3 className="industry-title group-hover:text-brand-teal">Q2 2026 Freight Rate Forecast</h3>
              <p className="text-ui-gray text-sm flex-grow mb-4">Analysis of current rate dynamics: Red Sea diversions pushing sea rates up 18%...</p>
              <Link to="/blog/industry" className="text-brand-teal font-semibold text-sm">Read Insight →</Link>
            </article>
            <article className="industry-card group">
              <div className="mb-4 text-4xl">🌍</div>
              <h3 className="industry-title group-hover:text-brand-teal">India's Top 10 Export Commodities</h3>
              <p className="text-ui-gray text-sm flex-grow mb-4">Breaking down the $450B export basket: where the volumes are, which ports dominate...</p>
              <Link to="/blog/industry" className="text-brand-teal font-semibold text-sm">Read Insight →</Link>
            </article>
            <article className="industry-card group">
              <div className="mb-4 text-4xl">📜</div>
              <h3 className="industry-title group-hover:text-brand-teal">New CBIC Guidelines for DG Cargo</h3>
              <p className="text-ui-gray text-sm flex-grow mb-4">Summary of the April 2026 CBIC circular on dangerous goods documentation...</p>
              <Link to="/blog/industry" className="text-brand-teal font-semibold text-sm">Read Insight →</Link>
            </article>
          </div>
        </div>

        {/* Design Section */}
        <div>
          <div className="blog-category-header">
            <h2 className="blog-category-title">
              <span>🎨</span> Design Blog
            </h2>
            <Link to="/blog/design" className="blog-view-all">View All →</Link>
          </div>
          <div className="blog-grid" style={{ padding: '0', maxWidth: '100%' }}>
             <article className="blog-card group">
              <div className="blog-card-img design-1">
                <span className="blog-card-emoji">🎨</span>
              </div>
              <div className="blog-card-content">
                <h3 className="blog-card-title">Crafting the Ocean Teal Design System</h3>
                <p className="blog-card-desc">How we developed our core design system to bring trust and clarity to logistics software.</p>
                <Link to="/blog/design" className="text-pink-600 font-semibold mt-4">Read Article →</Link>
              </div>
            </article>
            <article className="blog-card group">
              <div className="blog-card-img design-2">
                <span className="blog-card-emoji">🚚</span>
              </div>
              <div className="blog-card-content">
                <h3 className="blog-card-title">Designing for Transporters</h3>
                <p className="blog-card-desc">Our approach to creating high-contrast, easy-to-tap interfaces for truck drivers.</p>
                <Link to="/blog/design" className="text-pink-600 font-semibold mt-4">Read Article →</Link>
              </div>
            </article>
          </div>
        </div>

      </section>
    </div>
  );
}
