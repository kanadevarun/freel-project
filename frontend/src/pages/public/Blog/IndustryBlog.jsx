
import './Blog.css';

export default function IndustryBlog() {
  const articles = [
    {
      id: 1,
      tag: 'Rate Trends',
      tagColor: { bg: '#F0FDFA', text: '#0F766E' },
      title: 'Q2 2026 Freight Rate Forecast: Air & Sea from India',
      desc: 'Analysis of current rate dynamics: Red Sea diversions pushing sea rates up 18%, while air cargo capacity additions are easing prices on key lanes.',
      date: 'Apr 25, 2026',
      readTime: '10 min read',
      emoji: '📈'
    },
    {
      id: 2,
      tag: 'Trade Data',
      tagColor: { bg: '#FAF5FF', text: '#7E22CE' },
      title: 'India\'s Top 10 Export Commodities by HSN: 2026 Update',
      desc: 'Breaking down the $450B export basket: where the volumes are, which ports dominate, and emerging corridors to watch.',
      date: 'Apr 19, 2026',
      readTime: '8 min read',
      emoji: '🌍'
    },
    {
      id: 3,
      tag: 'Regulatory',
      tagColor: { bg: '#EFF6FF', text: '#1D4ED8' },
      title: 'New CBIC Guidelines for DG Cargo: What Shippers Must Know',
      desc: 'Summary of the April 2026 CBIC circular on dangerous goods documentation, MSDS requirements, and port-specific rules.',
      date: 'Apr 15, 2026',
      readTime: '6 min read',
      emoji: '📜'
    },
    {
      id: 4,
      tag: 'Shipping',
      tagColor: { bg: '#FFFBEB', text: '#B45309' },
      title: 'JNPT vs Mundra: The Battle for India\'s Busiest Container Port',
      desc: 'Comparing infrastructure, transit times, charges, and vessel schedules between India\'s two mega-ports.',
      date: 'Apr 10, 2026',
      readTime: '7 min read',
      emoji: '🚢'
    },
    {
      id: 5,
      tag: 'Road',
      tagColor: { bg: '#ECFDF5', text: '#047857' },
      title: 'E-Way Bill 2.0: How It Changes Inter-State Road Freight',
      desc: 'The new integrated e-way bill system and its impact on transit times, compliance costs, and transporter operations.',
      date: 'Apr 3, 2026',
      readTime: '5 min read',
      emoji: '🚛'
    },
    {
      id: 6,
      tag: 'Air Cargo',
      tagColor: { bg: '#EEF2FF', text: '#4338CA' },
      title: 'India\'s Air Cargo Growth: New Freighter Routes in 2026',
      desc: 'Emirates, Qatar, and IndiGo CarGo expanding capacity from DEL, BOM, and MAA — what it means for shippers.',
      date: 'Mar 26, 2026',
      readTime: '6 min read',
      emoji: '✈️'
    }
  ];

  return (
    <div className="bg-ui-surface min-h-screen">
      <section className="blog-hero" style={{ maxWidth: '800px' }}>
        <div className="blog-tag" style={{ backgroundColor: '#F0FDFA', color: '#0F766E', borderColor: '#CCFBF1' }}>
          <span className="text-sm font-semibold uppercase tracking-wider">📊 Industry</span>
        </div>
        <h1 className="blog-hero-title">
          <span className="text-gradient bg-gradient-to-r from-brand-teal to-blue-600">Industry</span> Insights
        </h1>
        <p className="blog-hero-sub">
          Logistics trends, trade data, rate forecasts, and regulatory updates for Indian freight. Stay ahead of the curve.
        </p>
      </section>

      <section className="blog-grid col-3">
        {articles.map(article => (
          <article key={article.id} className="industry-card group">
            <div className="flex items-start justify-between mb-6">
              <span className="industry-icon-wrapper group-hover:scale-110 transition-transform">{article.emoji}</span>
              <span 
                className="industry-tag"
                style={{ backgroundColor: article.tagColor.bg, color: article.tagColor.text }}
              >
                {article.tag}
              </span>
            </div>
            <h2 className="industry-title group-hover:text-brand-teal transition-colors">
              {article.title}
            </h2>
            <p className="text-ui-gray leading-relaxed mb-6 flex-grow">
              {article.desc}
            </p>
            <div className="flex items-center justify-between text-sm text-ui-gray border-t border-ui-border pt-5 mt-auto">
              <span className="font-medium text-brand-navy">{article.date}</span>
              <span className="flex items-center gap-1">⏱️ {article.readTime}</span>
            </div>
          </article>
        ))}
      </section>

      <div className="text-center pb-16">
        <button className="px-8 py-3 rounded-full bg-white border border-ui-border text-brand-navy font-medium hover:bg-ui-surface hover:text-brand-teal transition-colors shadow-sm">
          Load More Articles
        </button>
      </div>

      <section className="bg-gradient-to-br from-brand-navy to-slate-800 py-20 px-6 border-t border-ui-border">
        <div className="max-w-3xl mx-auto text-center">
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">Stay Informed, Ship Smarter</h2>
          <p className="text-slate-300 text-lg mb-8">Get weekly insights, rate updates, and regulatory news delivered directly to your inbox.</p>
          
          <form className="flex flex-col sm:flex-row gap-3 justify-center max-w-lg mx-auto" onSubmit={(e) => e.preventDefault()}>
            <input 
              type="email" 
              placeholder="Enter your work email" 
              required
              className="flex-grow px-5 py-3 rounded-full bg-slate-800 border border-slate-700 text-white placeholder-slate-400 focus:outline-none focus:border-brand-teal focus:ring-1 focus:ring-brand-teal transition-colors"
            />
            <button 
              type="submit"
              className="px-8 py-3 rounded-full text-white font-medium bg-gradient-to-r from-brand-teal to-blue-500 hover:from-teal-400 hover:to-blue-400 shadow-md transition-all whitespace-nowrap"
            >
              Subscribe →
            </button>
          </form>
        </div>
      </section>
    </div>
  );
}
