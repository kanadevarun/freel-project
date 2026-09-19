import React from 'react';
import { Link } from 'react-router-dom';
import { Sparkles, Info } from 'lucide-react';

export default function PersonalizationIndicator({ style = {}, notice = "Personalized using your saved preferences" }) {
  return (
    <div 
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '6px',
        padding: '3px 10px',
        borderRadius: '20px',
        background: '#f0f9ff',
        border: '1px solid #e0f2fe',
        color: '#0369a1',
        fontSize: '0.72rem',
        fontWeight: 600,
        ...style
      }}
      title="This response was personalized according to your active AI Memory settings."
    >
      <Sparkles size={12} color="#0284c7" />
      <span>{notice}</span>
      <Link 
        to="/dashboard/settings/memory"
        style={{
          color: '#0284c7',
          display: 'inline-flex',
          alignItems: 'center',
          textDecoration: 'none',
          marginLeft: '2px'
        }}
        title="View or edit AI memory preferences"
      >
        <Info size={12} />
      </Link>
    </div>
  );
}
