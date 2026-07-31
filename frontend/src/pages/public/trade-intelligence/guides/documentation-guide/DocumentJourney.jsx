import React from 'react';
import { Book } from 'lucide-react';
import './DocumentJourney.css';

export default function DocumentJourney() {
  const stages = [
    { id: 1, title: 'Factory', badge: 'Party', bColor: '#eff6ff', bText: '#2563eb', desc: 'Goods are manufactured and ready for export.', img: 'dj-stage-01.png', handler: null },
    { id: 2, title: 'Exporter', badge: 'Party', bColor: '#eff6ff', bText: '#2563eb', desc: 'Exporter prepares shipment and documentation.', img: 'dj-stage-02.png', handler: null },
    { id: 3, title: 'Commercial Invoice', badge: 'Commercial', bColor: '#eff6ff', bText: '#2563eb', desc: 'Declares goods value and terms of sale.', img: 'dj-stage-03.png', handler: 'Prepared by Exporter' },
    { id: 4, title: 'Packing List', badge: 'Commercial', bColor: '#eff6ff', bText: '#2563eb', desc: 'Lists all items, quantities, and packaging details.', img: 'dj-stage-04.png', handler: 'Prepared by Exporter' },
    { id: 5, title: 'Certificate of Origin', badge: 'Certificates', bColor: '#f0fdf4', bText: '#16a34a', desc: 'Certifies the origin of goods for customs purposes.', img: 'dj-stage-05.png', handler: 'Prepared by Exporter' },
    { id: 6, title: 'Truck Pickup', badge: 'Transport', bColor: '#fff7ed', bText: '#ea580c', desc: 'Goods are picked up and moved to the port.', img: 'dj-stage-06.png', handler: 'Handled by Carrier' },
    { id: 7, title: 'Port of Loading', badge: 'Transport', bColor: '#fff7ed', bText: '#ea580c', desc: 'Shipment arrives at port for loading.', img: 'dj-stage-07.png', handler: 'Handled by Carrier' },
    { id: 8, title: 'Bill of Lading', badge: 'Transport', bColor: '#fff7ed', bText: '#ea580c', desc: 'Contract of carriage and receipt of goods.', img: 'dj-stage-08.png', handler: 'Issued by Carrier' },
    { id: 9, title: 'Ocean Freight', badge: 'Transport', bColor: '#fff7ed', bText: '#ea580c', desc: 'Goods are transported to the destination.', img: 'dj-stage-09.png', handler: 'Handled by Carrier' },
    { id: 10, title: 'Insurance Certificate', badge: 'Insurance', bColor: '#f0fdf4', bText: '#16a34a', desc: 'Provides financial protection during transit.', img: 'dj-stage-10.png', handler: 'Issued by Insurer' },
    { id: 11, title: 'Customs Declaration', badge: 'Customs', bColor: '#f3e8ff', bText: '#7c3aed', desc: 'Filed with customs for clearance approval.', img: 'dj-stage-11.png', handler: 'Prepared by Importer' },
    { id: 12, title: 'Importer', badge: 'Party', bColor: '#eff6ff', bText: '#2563eb', desc: 'Goods are cleared by customs and delivered to the importer.', img: 'dj-stage-12.png', handler: null }
  ];

  const legends = [
    { name: 'PARTY', color: '#2563eb', desc: 'People / Entities Involved' },
    { name: 'COMMERCIAL', color: '#2563eb', desc: 'Describes the Goods' },
    { name: 'CERTIFICATES', color: '#16a34a', desc: 'Official Certifications' },
    { name: 'TRANSPORT', color: '#ea580c', desc: 'Movement of Goods' },
    { name: 'INSURANCE', color: '#16a34a', desc: 'Risk Protection' },
    { name: 'CUSTOMS', color: '#7c3aed', desc: 'Regulatory Clearance' }
  ];

  return (
    <section className="dj-section">
      <div className="dj-container">
        
        {/* TOP GRID */}
        <div className="dj-top-grid">
          
          <div className="dj-header-col">
            <div className="ed-badge">
              <span className="ed-badge-num">03</span>
              <span className="ed-badge-text">DOCUMENT JOURNEY</span>
            </div>
            <h2 className="ed-title">
              Follow Every<br/>
              Trade <span className="ed-title-highlight">Document</span><br/>
              Through the<br/>
              Shipment Journey
            </h2>
            <p className="ed-subtitle">
              Every trade document is created at a different stage of an international shipment.<br/><br/>
              Understanding the sequence helps prevent documentation mistakes and customs delays.
            </p>
          </div>

          <div className="dj-scene-col">
            <img src="/images/documentation/dj-decor-worldmap.png?v=1" alt="World Map" className="dj-scene-worldmap" />
            <div className="dj-scene-path"></div>
            <div className="dj-scene-items">
              <img src="/images/documentation/dj-decor-factory.png?v=1" alt="Factory" className="dj-scene-item" />
              <img src="/images/documentation/dj-decor-truck.png?v=1" alt="Truck" className="dj-scene-item dj-scene-item-truck" />
              <img src="/images/documentation/dj-decor-crane.png?v=1" alt="Port Crane" className="dj-scene-item" />
              <img src="/images/documentation/dj-decor-ship.png?v=1" alt="Cargo Ship" className="dj-scene-item dj-scene-item-ship" />
              <img src="/images/documentation/dj-decor-warehouse.png?v=1" alt="Importer Building" className="dj-scene-item" />
            </div>
            <img src="/images/documentation/dj-decor-cloud.png?v=1" alt="Cloud" className="dj-scene-cloud" />
          </div>

        </div>

        {/* TIMELINE */}
        <div className="dj-timeline-wrapper">
          <div className="dj-connector-line"></div>
          <div className="dj-timeline-scroll">
            {stages.map(stage => (
              <div className="dj-stage-card" key={stage.id}>
                <div className="dj-stage-num">{stage.id.toString().padStart(2, '0')}</div>
                <div className="dj-stage-content">
                  <div className="dj-stage-badge" style={{ backgroundColor: stage.bColor, color: stage.bText }}>
                    {stage.badge}
                  </div>
                  <img src={`/images/documentation/${stage.img}?v=1`} alt={stage.title} className="dj-stage-illust" />
                  <h4 className="dj-stage-title">{stage.title}</h4>
                  <p className="dj-stage-desc">{stage.desc}</p>
                  {stage.handler && (
                    <div className="dj-stage-handler">
                      {stage.handler}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* LEGEND */}
        <div className="dj-legend">
          {legends.map(l => (
            <div className="dj-legend-item" key={l.name}>
              <div className="dj-legend-dot" style={{ backgroundColor: l.color }}></div>
              <span className="dj-legend-name" style={{ color: l.color }}>{l.name}</span>
              <span className="dj-legend-desc">{l.desc}</span>
            </div>
          ))}
        </div>

        {/* LEARNING CARDS */}
        <div className="dj-learning-grid">
          
          <div className="dj-learning-card dj-lc-blue">
            <div className="dj-learning-icon-bg">
              <img src="/images/documentation/dj-icon-why.png?v=1" alt="Why" />
            </div>
            <div className="dj-learning-content">
              <h3>Why This Order Matters</h3>
              <p>Creating documents in the wrong sequence can delay customs clearance and shipment release.</p>
            </div>
          </div>

          <div className="dj-learning-card dj-lc-orange">
            <div className="dj-learning-icon-bg">
              <img src="/images/documentation/dj-icon-mistake.png?v=1" alt="Mistake" />
            </div>
            <div className="dj-learning-content">
              <h3>Common Mistake</h3>
              <p>Submitting customs paperwork before receiving the Bill of Lading.</p>
            </div>
          </div>

          <div className="dj-learning-card dj-lc-green">
            <div className="dj-learning-icon-bg">
              <img src="/images/documentation/dj-icon-tip.png?v=1" alt="Tip" />
            </div>
            <div className="dj-learning-content">
              <h3>Pro Tip</h3>
              <p>Most exporters prepare the Commercial Invoice and Packing List together.</p>
            </div>
          </div>

        </div>

        {/* BOTTOM STRIP */}
        <div className="dj-bottom-strip">
          <div className="dj-bs-icon">
            <Book size={20} strokeWidth={2.5} />
          </div>
          <p>Master the document flow to ensure smooth shipments, happy customers, and hassle-free global trade.</p>
        </div>

      </div>
    </section>
  );
}
