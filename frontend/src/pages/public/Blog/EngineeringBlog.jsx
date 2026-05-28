
import './Blog.css';

export default function EngineeringBlog() {
  const posts = [
    {
      id: 1,
      title: 'How We Built Real-Time Multi-Modal Tracking with GPS, AIS & FlightAware',
      desc: 'A deep-dive into unifying truck GPS, vessel AIS, and flight tracking into a single dashboard with sub-second updates.',
      tag: 'Tracking',
      tagColor: { bg: '#F0FDFA', text: '#0F766E', border: '#CCFBF1' },
      date: 'Apr 22, 2026',
      readTime: '8 min read',
      author: 'Vaidanshi',
      emoji: '🗺️',
      imgClass: 'engineering-1'
    },
    {
      id: 2,
      title: 'AI Cargo Scanner: From Photo to HS Code in 3 Seconds',
      desc: 'How we trained a vision model to estimate cargo dimensions, weight, and harmonized system codes from a single smartphone photo.',
      tag: 'AI/ML',
      tagColor: { bg: '#FAF5FF', text: '#7E22CE', border: '#F3E8FF' },
      date: 'Apr 18, 2026',
      readTime: '12 min read',
      author: 'Vaidanshi',
      emoji: '🤖',
      imgClass: 'engineering-2'
    },
    {
      id: 3,
      title: 'Scaling RFQ Matching: How We Handle 10,000+ Vendor Responses/Hour',
      desc: 'Our event-driven architecture for real-time RFQ distribution, quote collection, and AI-powered vendor scoring.',
      tag: 'Architecture',
      tagColor: { bg: '#EFF6FF', text: '#1D4ED8', border: '#DBEAFE' },
      date: 'Apr 12, 2026',
      readTime: '10 min read',
      author: 'Engineering Team',
      emoji: '📊',
      imgClass: 'engineering-3'
    },
    {
      id: 4,
      title: 'Role-Based Access in Multi-Tenant Logistics: Our RBAC Design',
      desc: 'How we designed permission systems that work across Clients, Transporters, and Shipping Lines with zero data leakage.',
      tag: 'Security',
      tagColor: { bg: '#FFFBEB', text: '#B45309', border: '#FEF3C7' },
      date: 'Apr 5, 2026',
      readTime: '7 min read',
      author: 'Engineering Team',
      emoji: '🔐',
      imgClass: 'engineering-4'
    },
    {
      id: 5,
      title: 'Making Rate Comparison 10x Faster with Edge Caching',
      desc: 'How we reduced rate comparison latency from 3s to 300ms by implementing edge-cached vendor rate sheets.',
      tag: 'Performance',
      tagColor: { bg: '#ECFDF5', text: '#047857', border: '#D1FAE5' },
      date: 'Mar 28, 2026',
      readTime: '6 min read',
      author: 'Engineering Team',
      emoji: '⚡',
      imgClass: 'engineering-5'
    },
    {
      id: 6,
      title: 'Building the Transporter App: React Native + Offline-First Architecture',
      desc: 'How drivers upload proof-of-delivery photos with spotty network coverage using our offline-first sync engine.',
      tag: 'Mobile',
      tagColor: { bg: '#EEF2FF', text: '#4338CA', border: '#E0E7FF' },
      date: 'Mar 20, 2026',
      readTime: '9 min read',
      author: 'Engineering Team',
      emoji: '📱',
      imgClass: 'engineering-6'
    }
  ];

  return (
    <div className="bg-ui-surface min-h-screen">
      <section className="blog-hero">
        <div className="blog-tag">
          <span className="text-lg">⚙️</span> Engineering at Freel
        </div>
        <h1 className="blog-hero-title">
          The <span className="text-gradient bg-gradient-to-r from-brand-teal to-brand-indigo">Engineering</span> Blog
        </h1>
        <p className="blog-hero-sub">
          How we build the technology behind India's first multi-modal logistics OS.
        </p>
      </section>

      <section className="blog-grid">
        {posts.map(post => (
          <article key={post.id} className="blog-card group">
            <div className={`blog-card-img ${post.imgClass}`}>
              <div className="absolute inset-0 bg-white/10 mix-blend-overlay"></div>
              <span className="blog-card-emoji">{post.emoji}</span>
            </div>
            <div className="blog-card-content">
              <span 
                className="blog-card-tag"
                style={{ 
                  backgroundColor: post.tagColor.bg, 
                  color: post.tagColor.text, 
                  borderColor: post.tagColor.border,
                  borderWidth: '1px'
                }}
              >
                {post.tag}
              </span>
              <h2 className="blog-card-title">{post.title}</h2>
              <p className="blog-card-desc">{post.desc}</p>
              <div className="blog-card-meta">
                <span>{post.date} · {post.readTime}</span>
                <span className="text-brand-navy font-semibold">By {post.author}</span>
              </div>
            </div>
          </article>
        ))}
      </section>

      <section className="max-w-[1100px] mx-auto px-6 pb-20">
        <div className="bg-gradient-to-br from-ui-surface to-white border border-ui-border rounded-3xl p-12 text-center relative overflow-hidden">
          <div className="absolute top-0 right-0 -mr-20 -mt-20 w-64 h-64 rounded-full bg-brand-teal opacity-10 blur-3xl"></div>
          <div className="absolute bottom-0 left-0 -ml-20 -mb-20 w-64 h-64 rounded-full bg-brand-indigo opacity-10 blur-3xl"></div>
          <div className="relative z-10 max-w-2xl mx-auto">
            <h3 className="text-3xl font-extrabold text-brand-navy mb-4">Want to Build With Us?</h3>
            <p className="text-lg text-ui-gray mb-8">We're hiring engineers who love logistics & hard problems.</p>
            <button className="bg-brand-indigo text-white font-bold px-8 py-3.5 rounded-full hover:bg-opacity-90 transition-all shadow-md">
              Join the Team →
            </button>
          </div>
        </div>
      </section>
    </div>
  );
}
