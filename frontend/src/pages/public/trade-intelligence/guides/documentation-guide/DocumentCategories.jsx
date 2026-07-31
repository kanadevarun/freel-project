import React from 'react';
import { ArrowRight } from 'lucide-react';
import './DocumentCategories.css';

export default function DocumentCategories() {
  const cards = [
    {
      id: 'commercial',
      title: 'Commercial Documents',
      desc: 'Documents describing the goods being traded.',
      items: ['Commercial Invoice', 'Packing List', 'Proforma Invoice'],
      icon: 'dc-badge-commercial.png?v=10',
      img: 'dc-card-commercial.png?v=10',
      color: '#16a34a' // Green
    },
    {
      id: 'transport',
      title: 'Transport Documents',
      desc: 'Documents used for transporting cargo.',
      items: ['Bill of Lading', 'Air Waybill', 'CMR'],
      icon: 'dc-badge-transport.png?v=10',
      img: 'dc-card-transport.png?v=10',
      color: '#2563eb' // Blue
    },
    {
      id: 'customs',
      title: 'Customs Documents',
      desc: 'Required by customs authorities.',
      items: ['Customs Declaration', 'Import Declaration', 'Export Declaration'],
      icon: 'dc-badge-customs.png?v=10',
      img: 'dc-card-customs.png?v=10',
      color: '#7c3aed' // Purple
    },
    {
      id: 'financial',
      title: 'Financial Documents',
      desc: 'Documents used for international payments.',
      items: ['Letter of Credit', 'Bank Guarantee', 'Collection Order'],
      icon: 'dc-badge-finance.png?v=10',
      img: 'dc-card-financial.png?v=10',
      color: '#d97706' // Orange
    },
    {
      id: 'insurance',
      title: 'Insurance Documents',
      desc: 'Protects cargo during international transport.',
      items: ['Insurance Certificate', 'Cargo Policy', 'Marine Insurance'],
      icon: 'dc-badge-insurance.png?v=10',
      img: 'dc-card-insurance.png?v=10',
      color: '#ea580c' // Dark Orange
    },
    {
      id: 'certificates',
      title: 'Certificates',
      desc: 'Official certificates issued by authorized bodies.',
      items: ['Certificate of Origin', 'Inspection Certificate', 'Phytosanitary Certificate'],
      icon: 'dc-badge-certificates.png?v=10',
      img: 'certificateOfOrigin.png?v=10',
      color: '#16a34a' // Green
    }
  ];

  return (
    <section className="dc-section">
      <div className="dc-container">
        
        {/* TOP ROW: Header, Ecosystem, Stats */}
        <div className="dc-top-grid">
          
          {/* Header */}
          <div className="dc-header-col">
            <div className="ed-badge">
              <span className="ed-badge-num">02</span>
              <span className="ed-badge-text">DOCUMENT CATEGORIES</span>
            </div>
            <h2 className="ed-title">
              Understanding<br/>
              Trade Document<br/>
              <span className="ed-title-highlight">Categories</span>
            </h2>
            <p className="ed-subtitle">
              Every international trade document has a specific purpose.
              Understanding these categories helps you know exactly when each document is required.
            </p>
          </div>

          {/* Ecosystem */}
          <div className="dc-eco-col">
            <div className="dc-eco-wrapper">
              <img src="/images/documentation/dc-ecosystem-full.png?v=10" alt="Trade Documentation Ecosystem" className="dc-eco-full-img" />
            </div>
          </div>

          {/* Stats */}
          <div className="dc-stats-col">
            <div className="dc-stat-card">
              <img src="/images/documentation/dc-stat-categories-icon.png?v=10" alt="Categories" />
              <div className="dc-stat-content">
                <h3>6</h3>
                <p>Major Categories</p>
              </div>
            </div>
            <div className="dc-stat-card">
              <img src="/images/documentation/dc-stat-documents-icon.png?v=10" alt="Documents" />
              <div className="dc-stat-content">
                <h3>30+</h3>
                <p>Trade Documents</p>
              </div>
            </div>
            <div className="dc-stat-card">
              <img src="/images/documentation/dc-stat-countries-icon.png?v=10" alt="Countries" />
              <div className="dc-stat-content">
                <h3>195+</h3>
                <p>Countries</p>
              </div>
            </div>
            <div className="dc-stat-card">
              <img src="/images/documentation/dc-stat-essential-icon.png?v=10" alt="Essential" />
              <div className="dc-stat-content">
                <h3>100%</h3>
                <p>Essential</p>
              </div>
            </div>
          </div>

        </div>

        {/* CARDS GRID */}
        <div className="dc-cards-grid">
          {cards.map(card => (
            <div className="dc-card" key={card.id}>
              {/* Illustration left */}
              <div className="dc-card-illust">
                <img src={`/images/documentation/${card.img}`} alt={card.title} />
              </div>
              
              {/* Content right */}
              <div className="dc-card-content">
                <div className="dc-card-top">
                  <h3>{card.title}</h3>
                  <img src={`/images/documentation/${card.icon}`} alt="" className="dc-card-badge" />
                </div>
                <p className="dc-card-desc">{card.desc}</p>
                
                <div className="dc-card-contains">
                  <span className="dc-contains-label">Contains:</span>
                  <ul>
                    {card.items.map((item, idx) => (
                      <li key={idx}>
                        <span className="dc-bullet" style={{backgroundColor: card.color}}></span>
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>

                <div className="dc-card-cta">
                  <span>View Documents</span>
                  <ArrowRight size={16} className="dc-card-cta-icon" />
                </div>
              </div>
            </div>
          ))}
        </div>



      </div>
    </section>
  );
}
