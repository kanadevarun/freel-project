import re

file_path = '/Users/varun.kanade/go/src/freel/freel-project/frontend/src/pages/public/trade-intelligence/guides/import-export-basics/Page.jsx'
with open(file_path, 'r') as f:
    content = f.read()

new_jsx = """const INCO_CARDS = [
  { name: 'EXW', title: 'Ex Works', desc: 'Seller makes goods available at their premises.', risk: 'Risk at Seller\\'s Premises', icons: ['doc_0.png', 'timeline_2.png', 'timeline_4.png'], img: 'any_0.png' },
  { name: 'FCA', title: 'Free Carrier', desc: 'Seller delivers to carrier or another nominated party.', risk: 'Risk at Carrier', icons: ['timeline_2.png', 'timeline_4.png'], img: 'any_1.png' },
  { name: 'CPT', title: 'Carriage Paid To', desc: 'Seller pays for carriage to the destination.', risk: 'Risk at Carrier', icons: ['timeline_2.png', 'timeline_4.png', 'timeline_5.png'], img: 'any_2.png' },
  { name: 'CIP', title: 'Carriage & Insurance Paid To', desc: 'Seller pays for carriage and insurance.', risk: 'Risk at Carrier', icons: ['timeline_2.png', 'timeline_4.png', 'timeline_5.png'], img: 'any_3.png' },
  { name: 'DAP', title: 'Delivered At Place', desc: 'Seller delivers goods to the named place of destination.', risk: 'Risk at Named Place', icons: ['timeline_2.png', 'timeline_4.png'], img: 'any_4.png' },
  { name: 'DPU', title: 'Delivered at Place Unloaded', desc: 'Seller delivers and unloads at destination.', risk: 'Risk at Named Place', icons: ['timeline_2.png', 'timeline_4.png', 'timeline_5.png'], img: 'any_5.png' },
  { name: 'DDP', title: 'Delivered Duty Paid', desc: 'Seller delivers, clears customs and pays all duties.', risk: 'Risk at Buyer\\'s Premises', icons: ['timeline_2.png', 'timeline_4.png', 'timeline_5.png'], img: 'any_6.png' },
  { name: 'FAS', title: 'Free Alongside Ship', desc: 'Seller places goods alongside the ship at port.', risk: 'Risk alongside Ship', icons: ['timeline_5.png'], img: 'sea_0.png' },
  { name: 'FOB', title: 'Free On Board', desc: 'Seller delivers goods on board the vessel.', risk: 'Risk on Board Ship', icons: ['timeline_5.png'], img: 'sea_1.png' },
  { name: 'CFR', title: 'Cost & Freight', desc: 'Seller pays for cost and freight to destination.', risk: 'Risk on Board Ship', icons: ['timeline_5.png'], img: 'sea_2.png' },
  { name: 'CIF', title: 'Cost, Insurance & Freight', desc: 'Seller pays cost, freight and insurance.', risk: 'Risk on Board Ship', icons: ['timeline_5.png'], img: 'sea_3.png' }
];

const INCO_TIMELINE = [
  { label: 'FACTORY', img: 'timeline_0.png' },
  { label: 'LOADING', img: 'timeline_1.png' },
  { label: 'TRUCK', img: 'timeline_2.png' },
  { label: 'EXPORT CUSTOMS', img: 'timeline_3.png' },
  { label: 'PORT', img: 'timeline_4.png' },
  { label: 'SHIP', img: 'timeline_5.png' },
  { label: 'IMPORT PORT', img: 'timeline_6.png' },
  { label: 'IMPORT CUSTOMS', img: 'timeline_3.png' },
  { label: 'WAREHOUSE', img: 'timeline_7.png' },
  { label: 'BUYER', img: 'timeline_8.png' }
];

const INCO_DEEP = [
  {
    name: 'EXW – Ex Works', img: 'any_0.png',
    seller: 'Makes goods available.',
    buyer: 'All transport, export, import, duties and delivery.',
    bestFor: 'Experienced importers.',
    example: 'Buyer arranges pickup from seller\\'s factory in India.'
  },
  {
    name: 'FOB – Free On Board', img: 'sea_1.png',
    seller: 'Delivers goods on board vessel at port.',
    buyer: 'Main transport, insurance, import clearance.',
    bestFor: 'Ocean freight shipments.',
    example: 'Seller ships goods from Mumbai Port.'
  },
  {
    name: 'CIF – Cost, Insurance & Freight', img: 'sea_3.png',
    seller: 'Pays for cost, freight and insurance.',
    buyer: 'Import clearance, duties, delivery.',
    bestFor: 'Buyers who want seller to arrange main transport.',
    example: 'Seller ships and insures goods to Hamburg.'
  },
  {
    name: 'DDP – Delivered Duty Paid', img: 'any_6.png',
    seller: 'Everything including duties and delivery.',
    buyer: 'Only accepts delivery.',
    bestFor: 'Buyers who want zero hassle.',
    example: 'Goods delivered to buyer\\'s warehouse in Germany.'
  }
];

const INCO_STATS = [
  { val: '11', title: 'Official Incoterms®', desc: 'in Incoterms 2020', icon: 'doc_0.png' },
  { val: '200+', title: 'Countries Using ICC Rules', desc: 'A global standard.', icon: 'timeline_8.png' },
  { val: '90%', title: 'International Contracts', desc: 'Reference Incoterms', icon: 'doc_2.png' },
  { val: 'Millions', title: 'of Shipments Governed', desc: 'Every Year', icon: 'doc_3.png' }
];

const INCO_MISTAKES = [
  'Using FOB for air freight shipments.',
  'Thinking CIF includes customs clearance.',
  'Assuming DDP always means cheapest.',
  'Ignoring insurance requirements.',
  'Not understanding risk transfer points.'
];

function IncotermsSection() {
  const BPATH = '/images/import-export/incoterms_new/';
  
  return (
    <section className="inco-section" id="incoterms-2020">
      <div className="inco-inner">
        {/* HERO */}
        <div className="inco-hero-row">
          <div className="inco-hero-header">
            <motion.div className="inco-badge" initial={{opacity:0, y:12}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
              <div className="inco-badge-num">07</div>
              <div className="inco-badge-text">INCOTERMS® 2020</div>
            </motion.div>
            <motion.h2 initial={{opacity:0, y:16}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.1}}>
              Master Global Trade<br/>with <span>Incoterms® 2020</span>
            </motion.h2>
            <motion.p initial={{opacity:0, y:16}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.2}}>
              Know exactly who pays, who ships, and who takes the risk at every stage of international trade.
            </motion.p>
          </div>
          
          <motion.div className="inco-hero-journey" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{duration:0.6}}>
            
            <div className="inco-hj-nodes">
              <div className="inco-hj-node" style={{left: '0%'}}>
                <div className="inco-hj-label">FACTORY</div>
                <img src={`${BPATH}hero_0.png`} className="inco-hj-img" alt="Factory" />
              </div>
              <div className="inco-hj-node" style={{left: '16.66%'}}>
                <div className="inco-hj-label">EXPORT<br/>WAREHOUSE</div>
                <img src={`${BPATH}hero_1.png`} className="inco-hj-img" alt="Export Warehouse" />
              </div>
              <div className="inco-hj-node" style={{left: '33.33%'}}>
                <div className="inco-hj-label">EXPORT<br/>PORT</div>
                <img src={`${BPATH}hero_3.png`} className="inco-hj-img" alt="Export Port" />
              </div>
              <div className="inco-hj-node" style={{left: '50%'}}>
                <div className="inco-hj-label">CARGO<br/>SHIP</div>
                <img src={`${BPATH}hero_4.png`} className="inco-hj-img" alt="Cargo Ship" style={{transform:'scale(1.2) translateY(-10px)'}} />
              </div>
              <div className="inco-hj-node" style={{left: '66.66%'}}>
                <div className="inco-hj-label">IMPORT<br/>PORT</div>
                <img src={`${BPATH}hero_5.png`} className="inco-hj-img" alt="Import Port" />
              </div>
              <div className="inco-hj-node" style={{left: '83.33%'}}>
                <div className="inco-hj-label">IMPORT<br/>WAREHOUSE</div>
                <img src={`${BPATH}timeline_7.png`} className="inco-hj-img" alt="Import Warehouse" />
              </div>
              <div className="inco-hj-node" style={{left: '100%'}}>
                <div className="inco-hj-label">BUYER</div>
                <img src={`${BPATH}timeline_8.png`} className="inco-hj-img" alt="Buyer" />
              </div>
              {/* Trucks linking them */}
              <img src={`${BPATH}hero_2.png`} className="inco-hj-truck" style={{left: '8%', bottom: 0}} alt="" />
              <img src={`${BPATH}hero_2.png`} className="inco-hj-truck" style={{left: '25%', bottom: 0}} alt="" />
              <img src={`${BPATH}hero_2.png`} className="inco-hj-truck" style={{left: '75%', bottom: 0}} alt="" />
            </div>

            {/* Responsibility Bars */}
            <div className="inco-resp-bars-hero">
              <div className="inco-rb-row">
                <div className="inco-rb-label" style={{color: '#3b82f6'}}>SELLER'S RESPONSIBILITY</div>
                <div className="inco-rb-line inco-rb-seller" style={{width: '100%'}}></div>
              </div>
              <div className="inco-rb-row">
                <div className="inco-rb-label" style={{color: '#22c55e'}}>BUYER'S RESPONSIBILITY</div>
                <div className="inco-rb-line inco-rb-buyer" style={{width: '83.33%', marginLeft: '16.66%'}}></div>
              </div>
              <div className="inco-rb-row">
                <div className="inco-rb-label" style={{color: '#f59e0b'}}>RISK TRANSFER POINT</div>
                <div className="inco-rb-line inco-rb-risk" style={{width: '50%'}}>
                  <MapPin className="inco-rb-pin" size={20} style={{right:-10, top: -14, fill: '#fff'}} />
                </div>
              </div>
              <div className="inco-rb-row">
                <div className="inco-rb-label" style={{color: '#8b5cf6'}}>COST TRANSFER POINT</div>
                <div className="inco-rb-line inco-rb-cost" style={{width: '66.66%'}}>
                  <div className="inco-rb-pin-cost" style={{right:-10}}>$</div>
                </div>
              </div>
            </div>
          </motion.div>
        </div>

        {/* EXPLORE INCOTERMS */}
        <div className="inco-explore">
          <div className="inco-section-title">
            <div className="inco-st-num">1</div>
            Explore Incoterms® 2020
          </div>
          <p className="inco-section-desc">Incoterms are international rules that define the responsibilities<br/>of buyers and sellers in global trade.</p>
          
          <div className="inco-cards-container">
            <div className="inco-sea-label">SEA & INLAND WATERWAY ONLY</div>
            <div className="inco-explore-grid">
              {INCO_CARDS.map((card, i) => (
                <motion.div key={i} className="inco-card" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay: i*0.03}}>
                  <h4>{card.name}</h4>
                  <img src={`${BPATH}${card.img}`} className="inco-card-img" alt={card.name} />
                  <div className="inco-card-name">{card.title}</div>
                  <div className="inco-card-desc">{card.desc}</div>
                  <div className="inco-card-risk">{card.risk}</div>
                  <div className="inco-card-icons">
                    {card.icons.map((icon, idx) => (
                      <img key={idx} src={`${BPATH}${icon}`} style={{height: 14}} alt=""/>
                    ))}
                  </div>
                </motion.div>
              ))}
            </div>
          </div>
        </div>

        {/* TIMELINE */}
        <div className="inco-timeline">
          <div className="inco-section-title">
            <div className="inco-st-num">2</div>
            Responsibility Timeline
          </div>
          <p className="inco-section-desc">Click on any Incoterm above to see how responsibility, risk and cost transfer across the journey.</p>
          
          <motion.div className="inco-tl-container" initial={{opacity:0, y:30}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
            <div className="inco-tl-flow">
              {INCO_TIMELINE.map((step, i) => (
                <div key={i} className="inco-tl-step">
                  <img src={`${BPATH}${step.img}`} className="inco-tl-img" alt={step.label} />
                  <div className="inco-tl-label">{step.label}</div>
                </div>
              ))}
            </div>

            <div className="inco-resp-bars-tl">
              <div className="inco-rb-row">
                <img src={`${BPATH}doc_0.png`} className="inco-rb-icon" alt="" />
                <div className="inco-rb-label">Seller's Responsibility</div>
                <div className="inco-rb-line inco-rb-seller" style={{width: '66.66%'}}></div>
                <img src={`${BPATH}timeline_8.png`} className="inco-rb-icon" alt="" style={{marginLeft: 8}}/>
                <div className="inco-rb-label" style={{marginLeft: 8}}>Buyer's Responsibility</div>
                <div className="inco-rb-line inco-rb-buyer" style={{flex: 1}}></div>
              </div>
              <div className="inco-rb-row">
                <MapPin size={16} color="#f59e0b" className="inco-rb-icon" fill="#fff" />
                <div className="inco-rb-label">Risk Transfer Point</div>
                <div className="inco-rb-line inco-rb-risk" style={{width: '66.66%'}}>
                  <MapPin className="inco-rb-pin" size={20} style={{right:-10, top:-14, fill:'#fff'}} />
                </div>
              </div>
              <div className="inco-rb-row">
                <div className="inco-rb-icon-cost">$</div>
                <div className="inco-rb-label">Cost Transfer Point</div>
                <div className="inco-rb-line inco-rb-cost" style={{width: '88.88%'}}>
                  <div className="inco-rb-pin-cost" style={{right:-10}}>$</div>
                </div>
              </div>
              
              <div className="inco-tl-legend">
                <span>Legend:</span>
                <span className="inco-leg-item"><div className="inco-leg-color" style={{background:'#3b82f6'}}></div> Blue = Seller</span>
                <span className="inco-leg-item"><div className="inco-leg-color" style={{background:'#22c55e'}}></div> Green = Buyer</span>
                <span className="inco-leg-item"><div className="inco-leg-color" style={{background:'#f59e0b'}}></div> Orange = Risk Transfer</span>
                <span className="inco-leg-item"><div className="inco-leg-color" style={{background:'#8b5cf6'}}></div> Purple = Cost Transfer</span>
              </div>
            </div>
          </motion.div>
        </div>

        {/* DEEP DIVE */}
        <div className="inco-section-title">
          <div className="inco-st-num">3</div>
          Deep Dive: Popular Incoterms
        </div>
        <p className="inco-section-desc">The most commonly used Incoterms explained in detail.</p>
        
        <div className="inco-deep">
          <div className="inco-deep-grid">
            {INCO_DEEP.map((dd, i) => (
              <motion.div key={i} className="inco-dd-card" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay: i*0.1}}>
                <div className="inco-dd-header">
                  <div className="inco-dd-title">{dd.name}</div>
                  <img src={`${BPATH}${dd.img}`} className="inco-dd-img" alt={dd.name} />
                </div>
                <div className="inco-dd-list">
                  <div className="inco-dd-item">
                    <div className="inco-dd-bullet">• <strong>What Seller Does</strong></div>
                    <p>{dd.seller}</p>
                  </div>
                  <div className="inco-dd-item">
                    <div className="inco-dd-bullet">• <strong>What Buyer Does</strong></div>
                    <p>{dd.buyer}</p>
                  </div>
                  <div className="inco-dd-item">
                    <div className="inco-dd-bullet">• <strong>Best For</strong></div>
                    <p>{dd.bestFor}</p>
                  </div>
                  <div className="inco-dd-item">
                    <div className="inco-dd-bullet">• <strong>Example</strong></div>
                    <p>{dd.example}</p>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* BOTTOM ROWS */}
        <div className="inco-bottom-row">
          <motion.div className="inco-box inco-rw" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
            <div className="inco-section-title">
              <div className="inco-st-num">4</div>
              Real World Example
            </div>
            <p className="inco-section-desc" style={{marginBottom:32}}>A furniture exporter in Vietnam sells goods to a buyer in Germany<br/>under FOB Ho Chi Minh Port.</p>
            
            <div className="inco-rw-flow">
              <div className="inco-rw-step">
                <img src={`${BPATH}hero_0.png`} className="inco-rw-img" alt="Factory" />
                <div className="inco-rw-label">Factory<br/>(Vietnam)</div>
              </div>
              <ArrowRight size={16} color="#94a3b8" />
              <div className="inco-rw-step">
                <img src={`${BPATH}hero_2.png`} className="inco-rw-img" alt="Truck" style={{transform:'scale(0.8)'}}/>
                <div className="inco-rw-label">Truck to<br/>Port</div>
              </div>
              <ArrowRight size={16} color="#94a3b8" />
              <div className="inco-rw-step">
                <img src={`${BPATH}hero_3.png`} className="inco-rw-img" alt="HCM Port" />
                <div className="inco-rw-label">Ho Chi Minh<br/>Port</div>
              </div>
              <ArrowRight size={16} color="#94a3b8" />
              <div className="inco-rw-step">
                <img src={`${BPATH}hero_4.png`} className="inco-rw-img" alt="Ship" />
                <div className="inco-rw-label">Ocean Freight</div>
              </div>
              <ArrowRight size={16} color="#94a3b8" />
              <div className="inco-rw-step">
                <img src={`${BPATH}hero_5.png`} className="inco-rw-img" alt="Hamburg" />
                <div className="inco-rw-label">Hamburg Port</div>
              </div>
              <ArrowRight size={16} color="#94a3b8" />
              <div className="inco-rw-step">
                <img src={`${BPATH}hero_2.png`} className="inco-rw-img" alt="Truck" style={{transform:'scale(0.8)'}}/>
                <div className="inco-rw-label">Truck to<br/>Warehouse</div>
              </div>
            </div>
            
            <div className="inco-rw-bars">
              <div className="inco-rw-bar-row">
                <div className="inco-rw-seller">
                  <div className="inco-rw-bar-label">Seller's Responsibility</div>
                  <div className="inco-rw-bar-desc">Up to goods loaded on ship</div>
                </div>
                <div className="inco-rw-buyer">
                  <div className="inco-rw-bar-label">Buyer's Responsibility</div>
                  <div className="inco-rw-bar-desc">From this point onwards</div>
                </div>
              </div>
              <MapPin className="inco-rw-pin" size={24} fill="#fff" color="#f59e0b" />
            </div>
          </motion.div>

          <div className="inco-bottom-col2">
            <motion.div className="inco-box" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.1}}>
              <div className="inco-section-title">
                <div className="inco-st-num">5</div>
                Common Mistakes
              </div>
              <div className="inco-mistake-list">
                {INCO_MISTAKES.map((msg, i) => (
                  <div key={i} className="inco-mistake-item">
                    <div className="inco-mistake-x">✕</div>
                    <span>{msg}</span>
                  </div>
                ))}
              </div>
            </motion.div>

            <motion.div className="inco-box inco-tip" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay:0.2}}>
              <div className="inco-section-title">
                <div className="inco-st-num">6</div>
                Smart Tip
              </div>
              <p>Think of Incoterms like splitting the bill during a road trip.</p>
              <p>Each person takes responsibility for different parts of the journey.</p>
              <p>The goods always reach the same destination, but Incoterms decide who pays, who manages and who takes the risk along the way.</p>
              
              <div className="inco-tip-roadtrip">
                <img src={`${BPATH}road_trip.png`} alt="Road Trip" />
              </div>
            </motion.div>
          </div>
        </div>

        {/* STATS */}
        <div className="inco-stats-row">
          <div className="inco-section-title" style={{marginRight: 40}}>
            <div className="inco-st-num">7</div>
            By The Numbers
          </div>
          <div className="inco-stats-grid">
            {INCO_STATS.map((stat, i) => (
              <motion.div key={i} className="inco-stat-box" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}} transition={{delay: i*0.1}}>
                <img src={`${BPATH}${stat.icon}`} className="inco-stat-icon" alt="" />
                <div className="inco-stat-text">
                  <div className="inco-stat-val">{stat.val}</div>
                  <div className="inco-stat-title">{stat.title}</div>
                  <div className="inco-stat-desc">{stat.desc}</div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* FINAL CTA */}
        <motion.div className="inco-final-cta" initial={{opacity:0, y:20}} whileInView={{opacity:1, y:0}} viewport={{once:true}}>
          <div className="inco-fc-left">
            <div className="inco-fc-icon"><img src={`${BPATH}doc_1.png`} style={{width: 32}} alt="" /></div>
            <div className="inco-fc-text">
              <h3>Great! You now understand Incoterms® 2020.</h3>
              <p>Next, let's learn how international payments and trade finance work.</p>
            </div>
          </div>
          <button className="inco-fc-btn">
            Next Module:<br/><strong>Payment & Finance</strong> <ArrowRight size={16} />
          </button>
        </motion.div>

      </div>
    </section>
  );
}"""

pattern = re.compile(r'const INCO_ANY_MODE = \[.*?</section>\n\s*\);\n}', re.DOTALL)
new_content = pattern.sub(new_jsx, content)

with open(file_path, 'w') as f:
    f.write(new_content)
