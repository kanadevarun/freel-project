import React from 'react';
import { User, Clock, ArrowRight, CheckCircle2 } from 'lucide-react';
import './EssentialDocuments.css';

const DOCUMENTS = [
  {
    id: 'commercial-invoice',
    title: 'Commercial Invoice',
    category: 'Commercial',
    categoryColor: 'green',
    difficulty: 'Beginner',
    difficultyColor: 'green',
    preparedBy: 'Exporter',
    usedAt: 'Before Shipment',
    image: '/images/documentation/card-commercial-invoice.png'
  },
  {
    id: 'packing-list',
    title: 'Packing List',
    category: 'Commercial',
    categoryColor: 'green',
    difficulty: 'Beginner',
    difficultyColor: 'green',
    preparedBy: 'Exporter',
    usedAt: 'Before Shipment',
    image: '/images/documentation/card-packing-list.png'
  },
  {
    id: 'bill-of-lading',
    title: 'Bill of Lading',
    category: 'Transport',
    categoryColor: 'blue',
    difficulty: 'Intermediate',
    difficultyColor: 'blue',
    preparedBy: 'Carrier',
    usedAt: 'During Shipment',
    image: '/images/documentation/card-bill-of-lading.png'
  },
  {
    id: 'certificate-of-origin',
    title: 'Certificate of Origin',
    category: 'Customs',
    categoryColor: 'purple',
    difficulty: 'Beginner',
    difficultyColor: 'green',
    preparedBy: 'Exporter / Chamber',
    usedAt: 'Before Shipment',
    image: '/images/documentation/card-certificate-origin.png'
  },
  {
    id: 'insurance-certificate',
    title: 'Insurance Certificate',
    category: 'Insurance',
    categoryColor: 'orange',
    difficulty: 'Intermediate',
    difficultyColor: 'blue',
    preparedBy: 'Insurance Company',
    usedAt: 'During Shipment',
    image: '/images/documentation/card-insurance.png'
  },
  {
    id: 'customs-declaration',
    title: 'Customs Declaration',
    category: 'Customs',
    categoryColor: 'purple',
    difficulty: 'Advanced',
    difficultyColor: 'orange',
    preparedBy: 'Importer / Customs Broker',
    usedAt: 'At Destination',
    image: '/images/documentation/card-customs-declaration.png'
  },
  {
    id: 'letter-of-credit',
    title: 'Letter of Credit',
    category: 'Finance',
    categoryColor: 'orange',
    difficulty: 'Intermediate',
    difficultyColor: 'blue',
    preparedBy: 'Bank (on Buyer\'s Request)',
    usedAt: 'Before Shipment',
    image: '/images/documentation/card-letter-credit.png'
  },
  {
    id: 'air-waybill',
    title: 'Air Waybill',
    category: 'Transport',
    categoryColor: 'blue',
    difficulty: 'Beginner',
    difficultyColor: 'green',
    preparedBy: 'Airline',
    usedAt: 'During Shipment',
    image: '/images/documentation/card-air-waybill.png'
  },
  {
    id: 'export-license',
    title: 'Export License',
    category: 'Customs',
    categoryColor: 'purple',
    difficulty: 'Advanced',
    difficultyColor: 'orange',
    preparedBy: 'Government Authority',
    usedAt: 'Before Shipment',
    image: '/images/documentation/card-export-license.png'
  }
];

const LEGEND_ITEMS = [
  { name: 'Commercial', color: 'green' },
  { name: 'Transport', color: 'blue' },
  { name: 'Finance', color: 'orange' },
  { name: 'Insurance', color: 'orange' },
  { name: 'Customs', color: 'purple' }
];

export default function EssentialDocuments() {
  return (
    <section className="ed-section">
      <div className="ed-container">
        
        {/* ── HEADER ──────────────────────────────────────── */}
        <div className="ed-header-grid">
          <div className="ed-header-content">
            <div className="ed-badge">
              <span className="ed-badge-num">01</span>
              <span className="ed-badge-text">ESSENTIAL TRADE DOCUMENTS</span>
            </div>
            
            <h2 className="ed-title">
              Essential Trade<br />
              <span className="ed-title-highlight">Documents</span><br />
              Every Exporter<br />
              Should Know
            </h2>
            
            <p className="ed-subtitle">
              These documents form the backbone of every international shipment.<br />
              Learn what each document does, who prepares it, and when it is required.
            </p>
          </div>
          
          <div className="ed-header-illustration">
            {/* The subtle logistics illustration background will be applied via CSS to this container */}
          </div>
        </div>

        {/* ── GRID ────────────────────────────────────────── */}
        <div className="ed-grid">
          {DOCUMENTS.map((doc) => (
            <div key={doc.id} className="ed-card">
              
              <div className="ed-card-image-wrapper">
                <img src={doc.image} alt={doc.title} className="ed-card-image" />
              </div>
              
              <div className="ed-card-body">
                <div className="ed-card-content">
                  <div className="ed-card-header-row">
                    <h3 className="ed-card-title">{doc.title}</h3>
                    <span className={`ed-category-badge ed-bg-${doc.categoryColor}`}>
                      {doc.category}
                    </span>
                  </div>
                  
                  <div className="ed-difficulty">
                    <span className={`ed-dot ed-dot-${doc.difficultyColor}`}></span>
                    {doc.difficulty}
                  </div>
                  
                  <div className="ed-metadata-grid">
                    <div className="ed-meta-item">
                      <User size={15} className="ed-meta-icon" strokeWidth={2} />
                      <div className="ed-meta-text">
                        <span className="ed-meta-label">Prepared By</span>
                        <span className="ed-meta-value">{doc.preparedBy}</span>
                      </div>
                    </div>
                    <div className="ed-meta-item">
                      <Clock size={15} className="ed-meta-icon" strokeWidth={2} />
                      <div className="ed-meta-text">
                        <span className="ed-meta-label">Used At</span>
                        <span className="ed-meta-value">{doc.usedAt}</span>
                      </div>
                    </div>
                  </div>
                </div>
                
                <div className="ed-card-footer">
                  <button className="ed-explore-btn">
                    Explore Document <ArrowRight size={14} className="ed-explore-icon" />
                  </button>
                </div>
              </div>

            </div>
          ))}
        </div>

        {/* ── LEGEND ──────────────────────────────────────── */}
        <div className="ed-legend">
          {LEGEND_ITEMS.map((item) => (
            <div key={item.name} className="ed-legend-item">
              <span className={`ed-legend-icon ed-legend-bg-${item.color}`}>
                <CheckCircle2 size={12} className="ed-legend-check" />
              </span>
              <span className="ed-legend-text">{item.name}</span>
            </div>
          ))}
        </div>
        
      </div>
    </section>
  );
}
