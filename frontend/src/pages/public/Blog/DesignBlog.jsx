
import './Blog.css';

export default function DesignBlog() {
  const posts = [
    {
      id: 1,
      title: 'Crafting the Ocean Teal Design System for Freel',
      desc: 'How we developed our core design system to bring trust, clarity, and modern aesthetics to logistics software.',
      tag: 'Design System',
      tagColor: { bg: '#FDF2F8', text: '#BE185D', border: '#FCE7F3' },
      date: 'May 10, 2026',
      readTime: '8 min read',
      author: 'Design Team',
      emoji: '🎨',
      imgClass: 'design-1'
    },
    {
      id: 2,
      title: 'Designing for Transporters: Accessibility on the Road',
      desc: 'Our approach to creating high-contrast, easy-to-tap interfaces for truck drivers operating in harsh sunlight.',
      tag: 'UX / Accessibility',
      tagColor: { bg: '#F5F3FF', text: '#6D28D9', border: '#EDE9FE' },
      date: 'May 2, 2026',
      readTime: '6 min read',
      author: 'UX Research',
      emoji: '🚚',
      imgClass: 'design-2'
    },
    {
      id: 3,
      title: 'Data Visualization in Logistics: Making Sense of 10,000+ Shipments',
      desc: 'A look into how we designed the interactive maps and charts that power the Freel live tracking dashboard.',
      tag: 'Data Viz',
      tagColor: { bg: '#F0FDFA', text: '#0F766E', border: '#CCFBF1' },
      date: 'Apr 28, 2026',
      readTime: '9 min read',
      author: 'Design Team',
      emoji: '📈',
      imgClass: 'design-3'
    }
  ];

  return (
    <div className="bg-ui-surface min-h-screen">
      <section className="blog-hero">
        <div className="blog-tag" style={{ backgroundColor: '#FDF2F8', color: '#BE185D', borderColor: '#FCE7F3' }}>
          <span className="text-lg">✨</span> Design at Freel
        </div>
        <h1 className="blog-hero-title">
          The <span className="text-gradient bg-gradient-to-r from-pink-500 to-orange-400">Design</span> Blog
        </h1>
        <p className="blog-hero-sub">
          Exploring user experience, interfaces, and the aesthetic decisions behind our platform.
        </p>
      </section>

      <section className="blog-grid" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))' }}>
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
        <div className="bg-gradient-to-br from-pink-50 to-orange-50 border border-pink-100 rounded-3xl p-12 text-center relative overflow-hidden">
          <div className="absolute top-0 right-0 -mr-20 -mt-20 w-64 h-64 rounded-full bg-pink-500 opacity-5 blur-3xl"></div>
          <div className="relative z-10 max-w-2xl mx-auto">
            <h3 className="text-3xl font-extrabold text-brand-navy mb-4">Join our Design Team</h3>
            <p className="text-lg text-ui-gray mb-8">Help us shape the future of logistics through beautiful, functional design.</p>
            <button className="bg-pink-600 text-white font-bold px-8 py-3.5 rounded-full hover:bg-pink-700 transition-all shadow-md">
              View Open Roles →
            </button>
          </div>
        </div>
      </section>
    </div>
  );
}
