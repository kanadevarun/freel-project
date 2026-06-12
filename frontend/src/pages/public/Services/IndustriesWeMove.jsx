import React from 'react';
import './IndustriesWeMove.css';

export default function IndustriesWeMove() {
  const industries = [
    {
      title: "Agriculture",
      desc: "Connecting farms to cities through reliable freight movement.",
      image: "/images/road-freight/Human_figures_in_logistics_hub_202606071633.jpeg"
    },
    {
      title: "Healthcare",
      desc: "Delivering critical medical supplies and cold-chain pharmaceuticals.",
      image: "/images/road-freight/Figures_inspect_trucks_at_border_202606071637.jpeg"
    },
    {
      title: "Retail",
      desc: "Keeping store shelves stocked and inventory flowing continuously.",
      image: "/images/Modern_fulfillment_warehouse_where_human_202606052226.jpeg"
    },
    {
      title: "Manufacturing",
      desc: "Transporting raw materials and finished goods nationwide.",
      image: "/images/Regional_manufacturing_hubs_emerging_across_202606052226.jpeg"
    },
    {
      title: "Infrastructure",
      desc: "Supplying mega construction projects with heavy materials.",
      image: "/images/road-freight/Figures_operating_construction_m…_202606071629.jpeg"
    },
    {
      title: "E-Commerce",
      desc: "Powering the final mile and fulfilling millions of online orders.",
      image: "/images/road-freight/Human_figure_overlooking_highway…_202606071615.jpeg"
    }
  ];

  return (
    <section className="iwm-container">
      <div className="max-w-7xl mx-auto px-6">
        
        <div className="text-center mb-16">
          <h2 className="text-5xl md:text-6xl font-black tracking-tighter mb-6">
            Industries We Move
          </h2>
          <p className="text-xl md:text-2xl text-slate-500 font-medium max-w-2xl mx-auto">
            One logistics network serving every sector of the economy.
          </p>
        </div>

        <div className="iwm-grid">
          {industries.map((industry, idx) => (
            <div key={idx} className="iwm-card">
              <div className="iwm-image-wrapper">
                <img src={industry.image} alt={industry.title} className="iwm-image" />
              </div>
              <div className="iwm-content">
                <h3 className="iwm-title">{industry.title}</h3>
                <p className="iwm-desc">{industry.desc}</p>
              </div>
            </div>
          ))}
        </div>
        
      </div>
    </section>
  );
}
