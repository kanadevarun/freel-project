import React, { useEffect, useRef, useState } from 'react';
import './EconomyOnWheels.css';

function useOnScreen(ref, rootMargin = '0px') {
  const [isIntersecting, setIntersecting] = useState(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIntersecting(true);
          // Optional: observer.unobserve(entry.target) if we only want it to trigger once
        }
      },
      { rootMargin, threshold: 0.2 }
    );
    if (ref.current) {
      observer.observe(ref.current);
    }
    return () => {
      observer.disconnect();
    };
  }, [ref, rootMargin]);

  return isIntersecting;
}

function Panel({ eyebrow, headline, subtext, image, reverse }) {
  const ref = useRef();
  const isVisible = useOnScreen(ref, '-50px');

  return (
    <div ref={ref} className={`eow-panel max-w-[1800px] mx-auto px-6 lg:px-12 ${reverse ? 'eow-panel-reverse' : ''} ${isVisible ? 'is-visible' : ''}`}>
      <div className="eow-image-container">
        <img src={image} alt={headline} className="eow-image" />
        <div className="eow-image-overlay" />
      </div>
      <div className="eow-text-container">
        <span className="eow-eyebrow">{eyebrow}</span>
        <h3 className="eow-headline">{headline}</h3>
        <p className="eow-subtext">{subtext}</p>
      </div>
    </div>
  );
}

export default function EconomyOnWheels() {
  const panels = [
    {
      eyebrow: "Food & Agriculture",
      headline: "Food Reaches Cities By Road.",
      subtext: "Farms connect to markets through freight networks.",
      image: "/images/road-freight/Human_figures_in_logistics_hub_202606071633.jpeg"
    },
    {
      eyebrow: "Healthcare",
      headline: "Medicines Arrive On Time.",
      subtext: "Critical healthcare supply chains depend on road transport.",
      image: "/images/road-freight/Figures_inspect_trucks_at_border_202606071637.jpeg"
    },
    {
      eyebrow: "Retail & E-Commerce",
      headline: "Every Click Starts A Journey.",
      subtext: "Millions of online orders begin with freight movement.",
      image: "/images/Modern_fulfillment_warehouse_where_human_202606052226.jpeg"
    },
    {
      eyebrow: "Manufacturing",
      headline: "Factories Never Stop Moving.",
      subtext: "Raw materials and finished goods travel nationwide.",
      image: "/images/Regional_manufacturing_hubs_emerging_across_202606052226.jpeg"
    },
    {
      eyebrow: "Construction & Infrastructure",
      headline: "Cities Are Built By Freight.",
      subtext: "Every major project depends on logistics.",
      image: "/images/road-freight/Figures_operating_construction_m…_202606071629.jpeg"
    },
    {
      eyebrow: "National Commerce",
      headline: "One Network.\nMillions Of Deliveries.",
      subtext: "Road freight connects every corner of the economy.",
      image: "/images/road-freight/Human_figure_overlooking_highway…_202606071615.jpeg"
    }
  ];

  return (
    <section className="eow-container">
      <div className="max-w-[1400px] mx-auto px-6 text-center mb-32">
        <div className="inline-block px-6 py-2 rounded-full border border-amber-500/30 bg-amber-500/10 text-amber-400 font-mono text-sm tracking-widest uppercase mb-8 shadow-[0_0_20px_rgba(245,158,11,0.1)]">
          The Economy On Wheels
        </div>
        <h2 className="text-6xl md:text-8xl font-black tracking-tighter mb-8 leading-[1.05] drop-shadow-2xl">
          If Trucks Stop,<br/>
          <span className="text-transparent bg-clip-text bg-gradient-to-r from-amber-400 to-orange-600">Everything Stops.</span>
        </h2>
        <p className="text-xl md:text-2xl text-slate-400 max-w-4xl mx-auto font-medium leading-relaxed">
          From food and medicine to factories, retail shelves, construction projects, and e-commerce deliveries, road freight keeps every sector moving.
        </p>
      </div>

      <div className="flex flex-col gap-10">
        {panels.map((panel, idx) => (
          <Panel key={idx} {...panel} reverse={idx % 2 !== 0} />
        ))}
      </div>
    </section>
  );
}
