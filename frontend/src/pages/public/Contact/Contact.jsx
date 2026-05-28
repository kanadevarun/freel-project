import { useState } from 'react';
import './Contact.css';

export default function Contact() {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    phone: '',
    company: '',
    spend: 'Under ₹5 Lakhs',
    mode: 'sea', // 'sea', 'air', 'road', 'customs'
    subject: '',
    message: ''
  });

  const [submitted, setSubmitted] = useState(false);

  const handleChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleModeSelect = (modeValue) => {
    setFormData({ ...formData, mode: modeValue });
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    setSubmitted(true);
  };

  const resetForm = () => {
    setFormData({
      name: '',
      email: '',
      phone: '',
      company: '',
      spend: 'Under ₹5 Lakhs',
      mode: 'sea',
      subject: '',
      message: ''
    });
    setSubmitted(false);
  };

  return (
    <div className="bg-white min-h-screen">
      {/* ═══ HERO ═══ */}
      <section className="contact-hero">
        <h1 className="contact-hero-title">
          Let's Transform Your <span className="text-gradient bg-gradient-to-r from-brand-teal to-brand-indigo">Supply Chain</span>
        </h1>
        <p className="contact-hero-sub">
          Book freight, integrate our APIs, or consult our logistics architects to secure automated, margin-protected global trade lanes.
        </p>
      </section>

      {/* ═══ CONTACT SECTION ═══ */}
      <section className="contact-section">
        
        {/* Contact Info */}
        <div>
          <h2 className="text-3xl font-extrabold text-brand-navy mb-8" style={{ color: 'var(--color-brand-navy)', fontSize: '1.85rem', fontWeight: 800, marginBottom: '24px' }}>Connect with our Officers</h2>
          
          <div className="contact-info-grid">
            
            <div className="contact-info-card">
              <div className="contact-icon">📧</div>
              <div className="contact-info-content">
                <h3>Direct Inquiries</h3>
                <p>For high-volume client setups or API integration support.</p>
                <a href="mailto:hello@freel.in">hello@freel.in</a>
              </div>
            </div>

            <div className="contact-info-card">
              <div className="contact-icon">📞</div>
              <div className="contact-info-content">
                <h3>Global Trade Desk</h3>
                <p>Consult with our lead pricing architects for bulk bid contracts.</p>
                <a href="tel:+912269938810">+91 22 6993 8810</a>
              </div>
            </div>

            <div className="contact-info-card">
              <div className="contact-icon">📍</div>
              <div className="contact-info-content">
                <h3>Command Center</h3>
                <p>Freel Technologies Pvt. Ltd.<br/>Executive Wing, MIDC Andheri East<br/>Mumbai 400093, Maharashtra, India</p>
                <a href="https://maps.google.com" target="_blank" rel="noopener noreferrer">View Location Map →</a>
              </div>
            </div>

          </div>
        </div>

        {/* Contact Form Wrapper */}
        <div>
          <div className="contact-form-wrapper" style={{ position: 'relative', overflow: 'hidden' }}>
            
            {!submitted ? (
              <>
                <h2 className="contact-form-title">Launch Consultation</h2>
                <p className="contact-form-desc">Provide your shipment profiles to receive dynamic contract pricing quotes within 2 hours.</p>
                
                <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                  
                  {/* Name & Work Email Row */}
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }} className="form-row-split">
                    <div className="form-group" style={{ marginBottom: 0 }}>
                      <label className="form-label" htmlFor="name">Full Name</label>
                      <input 
                        type="text" 
                        id="name" 
                        name="name" 
                        className="form-input" 
                        placeholder="Vaidanshi Sagar" 
                        value={formData.name}
                        onChange={handleChange}
                        required 
                      />
                    </div>

                    <div className="form-group" style={{ marginBottom: 0 }}>
                      <label className="form-label" htmlFor="email">Work Email</label>
                      <input 
                        type="email" 
                        id="email" 
                        name="email" 
                        className="form-input" 
                        placeholder="sagar@company.com" 
                        value={formData.email}
                        onChange={handleChange}
                        required 
                      />
                    </div>
                  </div>

                  {/* Phone & Company Row */}
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }} className="form-row-split">
                    <div className="form-group" style={{ marginBottom: 0 }}>
                      <label className="form-label" htmlFor="phone">Phone Number</label>
                      <input 
                        type="tel" 
                        id="phone" 
                        name="phone" 
                        className="form-input" 
                        placeholder="+91 98765 43210" 
                        value={formData.phone}
                        onChange={handleChange}
                        required 
                      />
                    </div>

                    <div className="form-group" style={{ marginBottom: 0 }}>
                      <label className="form-label" htmlFor="company">Company Name</label>
                      <input 
                        type="text" 
                        id="company" 
                        name="company" 
                        className="form-input" 
                        placeholder="Logistics Global Ltd" 
                        value={formData.company}
                        onChange={handleChange}
                        required 
                      />
                    </div>
                  </div>

                  {/* Mode of Interest Selector */}
                  <div className="form-group" style={{ marginBottom: 0 }}>
                    <label className="form-label">Primary Mode of Interest</label>
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '8px', background: '#f8fafc', padding: '6px', borderRadius: '12px', border: '1px solid var(--color-ui-border)' }}>
                      {[
                        { id: 'sea', label: '🚢 Ocean' },
                        { id: 'air', label: '✈️ Air' },
                        { id: 'road', label: '🚚 Road' },
                        { id: 'customs', label: '🛃 Customs' }
                      ].map((modeItem) => (
                        <button
                          key={modeItem.id}
                          type="button"
                          onClick={() => handleModeSelect(modeItem.id)}
                          style={{
                            padding: '10px 4px',
                            borderRadius: '8px',
                            border: 'none',
                            fontSize: '0.75rem',
                            fontWeight: 700,
                            cursor: 'pointer',
                            transition: 'all 0.3s ease',
                            background: formData.mode === modeItem.id ? 'white' : 'transparent',
                            color: formData.mode === modeItem.id ? 'var(--color-brand-navy)' : '#64748b',
                            boxShadow: formData.mode === modeItem.id ? '0 4px 10px rgba(0,0,0,0.06)' : 'none'
                          }}
                        >
                          {modeItem.label}
                        </button>
                      ))}
                    </div>
                  </div>

                  {/* Monthly Spend & Subject Row */}
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }} className="form-row-split">
                    <div className="form-group" style={{ marginBottom: 0 }}>
                      <label className="form-label" htmlFor="spend">Monthly Freight Spend</label>
                      <select 
                        id="spend" 
                        name="spend" 
                        className="form-input" 
                        style={{ height: '49px', background: 'var(--color-ui-surface)', appearance: 'none', cursor: 'pointer' }}
                        value={formData.spend}
                        onChange={handleChange}
                        required
                      >
                        <option value="Under ₹5 Lakhs">Under ₹5 Lakhs</option>
                        <option value="₹5 Lakhs - ₹20 Lakhs">₹5L - ₹20L</option>
                        <option value="₹20 Lakhs - ₹1 Crore">₹20L - ₹1Cr</option>
                        <option value="Above ₹1 Crore">Above ₹1 Crore</option>
                      </select>
                    </div>

                    <div className="form-group" style={{ marginBottom: 0 }}>
                      <label className="form-label" htmlFor="subject">Subject</label>
                      <input 
                        type="text" 
                        id="subject" 
                        name="subject" 
                        className="form-input" 
                        placeholder="e.g. Schedule Integration Demo" 
                        value={formData.subject}
                        onChange={handleChange}
                        required 
                      />
                    </div>
                  </div>

                  {/* Message */}
                  <div className="form-group" style={{ marginBottom: 0 }}>
                    <label className="form-label" htmlFor="message">Logistics Requirements</label>
                    <textarea 
                      id="message" 
                      name="message" 
                      className="form-textarea" 
                      style={{ minHeight: '90px' }}
                      placeholder="Describe your active trade lanes, average container sizes, or key compliance challenges..." 
                      value={formData.message}
                      onChange={handleChange}
                      required 
                    ></textarea>
                  </div>

                  <button type="submit" className="submit-btn" style={{ marginTop: '8px' }}>Send Request →</button>
                </form>
              </>
            ) : (
              /* GORGEOUS LEAD CONFIRMED SUCCESS STATE CARD */
              <div style={{ textAlign: 'center', padding: '24px 0', display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <div style={{
                  width: '72px',
                  height: '72px',
                  borderRadius: '50%',
                  background: 'rgba(0,191,165,0.08)',
                  border: '2px solid var(--color-brand-teal)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '2.5rem',
                  color: 'var(--color-brand-teal)',
                  marginBottom: '20px',
                  animation: 'pulse 2s infinite'
                }} className="success-pulse">
                  ✓
                </div>
                
                <h2 style={{ fontSize: '1.75rem', fontWeight: 800, color: 'var(--color-brand-navy)', marginBottom: '8px' }}>Request Dispatched</h2>
                <p style={{ color: 'var(--color-ui-gray)', fontSize: '0.9rem', maxWidth: '380px', margin: '0 auto 24px', lineHeight: 1.6 }}>
                  Our automated trade matching engine has allocated your profile to our high-volume logistics specialists.
                </p>

                {/* Simulated Captured Data Telemetry */}
                <div style={{ width: '100%', background: '#f8fafc', border: '1px solid var(--color-ui-border)', borderRadius: '16px', padding: '16px', textAlign: 'left', marginBottom: '24px', fontSize: '0.8rem', display: 'flex', flexDirection: 'column', gap: '8px' }}>
                  <div style={{ borderBottom: '1px dashed #e2e8f0', paddingBottom: '6px', fontWeight: 700, color: 'var(--color-brand-navy)', textTransform: 'uppercase', fontSize: '0.65rem' }}>
                    📡 Captured Telemetry Summary
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: '#64748b' }}>Account Representative:</span>
                    <span style={{ fontWeight: 700, color: 'var(--color-brand-navy)' }}>{formData.name}</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: '#64748b' }}>Organization:</span>
                    <span style={{ fontWeight: 700, color: 'var(--color-brand-navy)' }}>{formData.company}</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: '#64748b' }}>Target Mode:</span>
                    <span style={{ fontWeight: 700, color: 'var(--color-brand-teal)', textTransform: 'capitalize' }}>{formData.mode} freight</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: '#64748b' }}>Estimated Spend:</span>
                    <span style={{ fontWeight: 700 }}>{formData.spend}</span>
                  </div>
                </div>

                <div style={{ background: 'rgba(90,79,207,0.06)', border: '1px solid rgba(90,79,207,0.15)', padding: '12px 16px', borderRadius: '12px', fontSize: '0.75rem', color: 'var(--color-brand-indigo)', fontWeight: 600, width: '100%', marginBottom: '24px' }}>
                  ⏱️ <b>Response Time Commitment:</b> A pricing officer will call or email you at <b>{formData.email}</b> within 120 minutes.
                </div>

                <button 
                  type="button" 
                  onClick={resetForm} 
                  style={{
                    background: 'white',
                    border: '1px solid var(--color-ui-border)',
                    padding: '12px 24px',
                    borderRadius: '999px',
                    color: 'var(--color-brand-navy)',
                    fontSize: '0.85rem',
                    fontWeight: 700,
                    cursor: 'pointer',
                    transition: 'all 0.3s'
                  }}
                  className="solutions-cta-btn"
                >
                  ← Submit Another Inquiry
                </button>
              </div>
            )}

          </div>
        </div>

      </section>
    </div>
  );
}
