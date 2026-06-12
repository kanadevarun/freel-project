import React, { useEffect, useRef, useState } from 'react';
import './WhyChooseFreel.css';

function useOnScreen(ref, rootMargin = '0px') {
  const [isIntersecting, setIntersecting] = useState(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIntersecting(true);
        }
      },
      { rootMargin, threshold: 0.2 }
    );
    if (ref.current) {
      observer.observe(ref.current);
    }
    return () => observer.disconnect();
  }, [ref, rootMargin]);

  return isIntersecting;
}

function AnimatedCounter({ active, target, duration = 1500, suffix = "" }) {
  const [count, setCount] = useState(0);

  useEffect(() => {
    if (active && target > 0) {
      let start = 0;
      const increment = target / (duration / 16);
      const timer = setInterval(() => {
        start += increment;
        if (start >= target) {
          setCount(target);
          clearInterval(timer);
        } else {
          setCount(Math.floor(start));
        }
      }, 16);
      return () => clearInterval(timer);
    } else {
      setCount(target > 0 ? 0 : target);
    }
  }, [active, target, duration]);

  if (typeof target === 'string') {
    return <>{target}{suffix}</>; // Handle 24/7 as string
  }

  return <>{count.toLocaleString()}{suffix}</>;
}

export default function WhyChooseFreel() {
  const sectionRef = useRef(null);
  const isActive = useOnScreen(sectionRef, '-50px');

  const valueBlocks = [
    {
      icon: "🚚",
      target: 500,
      suffix: "+",
      titleSuffix: " Verified Transporters",
      desc: "Every transporter is vetted, audited, and performance monitored."
    },
    {
      icon: "📍",
      target: 0,
      suffix: "",
      titleSuffix: "Real-Time Tracking",
      desc: "Track shipments from pickup to proof-of-delivery."
    },
    {
      icon: "📄",
      target: 100,
      suffix: "%",
      titleSuffix: " Digital Compliance",
      desc: "Instant LR, E-Way Bill, and POD workflows."
    },
    {
      icon: "🛡",
      target: "24/7",
      suffix: "",
      titleSuffix: " Control Tower",
      desc: "Continuous monitoring, exception management, and operational support."
    }
  ];

  return (
    <section ref={sectionRef} className={`wcf-container ${isActive ? 'is-active' : ''}`}>
      <div className="max-w-7xl mx-auto px-6">
        
        {/* Header */}
        <div className="mb-16">
          <span className="wcf-eyebrow">Trusted Operations</span>
          <h2 className="wcf-headline">Why India's Supply Chains Choose Freel</h2>
          <p className="wcf-subheadline">
            Built for reliability, visibility, compliance, and scale — everything modern freight operations need in one network.
          </p>
        </div>

        {/* 2-Column Layout */}
        <div className="grid lg:grid-cols-2 gap-12 lg:gap-20 items-center">
          
          {/* LEFT: Premium Image */}
          <div className="wcf-image-wrapper">
            <img src="/images/logistics_control_tower.png" alt="Logistics Control Tower" className="wcf-image" />
          </div>

          {/* RIGHT: Value Blocks */}
          <div className="flex flex-col">
            {valueBlocks.map((block, idx) => (
              <div key={idx} className="wcf-value-block" style={{ transitionDelay: `${0.2 + (idx * 0.15)}s` }}>
                <div className="wcf-icon">{block.icon}</div>
                <div>
                  <h3 className="wcf-block-title">
                    {block.target > 0 || typeof block.target === 'string' ? (
                      <span className="text-amber-500 mr-1">
                        <AnimatedCounter active={isActive} target={block.target} suffix={block.suffix} />
                      </span>
                    ) : null}
                    {block.titleSuffix}
                  </h3>
                  <p className="wcf-block-desc">{block.desc}</p>
                </div>
              </div>
            ))}
          </div>

        </div>

      </div>
    </section>
  );
}
