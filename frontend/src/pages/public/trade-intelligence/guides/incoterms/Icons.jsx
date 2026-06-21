import React from 'react';

export const FactoryIcon = ({ className = "" }) => (
  <svg viewBox="0 0 64 64" className={className} width="100%" height="100%">
    <defs>
      <linearGradient id="facGrad" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#94a3b8" />
        <stop offset="100%" stopColor="#475569" />
      </linearGradient>
      <linearGradient id="facGrad2" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stopColor="#cbd5e1" />
        <stop offset="100%" stopColor="#94a3b8" />
      </linearGradient>
    </defs>
    <path d="M12 50 L12 28 L24 20 L24 28 L36 20 L36 28 L48 20 L48 50 Z" fill="url(#facGrad)" />
    <rect x="16" y="14" width="6" height="14" fill="url(#facGrad2)" />
    <rect x="28" y="14" width="6" height="14" fill="url(#facGrad2)" />
    <rect x="40" y="14" width="6" height="14" fill="url(#facGrad2)" />
    <rect x="16" y="38" width="8" height="12" fill="#fff" opacity="0.8" />
    <rect x="28" y="38" width="8" height="12" fill="#fff" opacity="0.8" />
    <rect x="40" y="38" width="8" height="12" fill="#fff" opacity="0.8" />
    {/* Smoke */}
    <circle cx="19" cy="8" r="4" fill="#e2e8f0" opacity="0.6" />
    <circle cx="23" cy="5" r="5" fill="#e2e8f0" opacity="0.4" />
    <circle cx="31" cy="8" r="4" fill="#e2e8f0" opacity="0.6" />
    <circle cx="35" cy="5" r="5" fill="#e2e8f0" opacity="0.4" />
    <circle cx="43" cy="8" r="4" fill="#e2e8f0" opacity="0.6" />
  </svg>
);

export const DocumentsIcon = ({ className = "" }) => (
  <svg viewBox="0 0 64 64" className={className} width="100%" height="100%">
    <defs>
      <linearGradient id="docGrad" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stopColor="#ffffff" />
        <stop offset="100%" stopColor="#f1f5f9" />
      </linearGradient>
      <linearGradient id="docShadow" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#e2e8f0" />
        <stop offset="100%" stopColor="#cbd5e1" />
      </linearGradient>
    </defs>
    {/* Back paper */}
    <path d="M16 12 h24 l12 12 v30 a4 4 0 0 1 -4 4 h-32 a4 4 0 0 1 -4 -4 v-38 a4 4 0 0 1 4 -4 z" fill="url(#docShadow)" />
    {/* Front paper */}
    <path d="M12 16 h24 l12 12 v30 a4 4 0 0 1 -4 4 h-32 a4 4 0 0 1 -4 -4 v-38 a4 4 0 0 1 4 -4 z" fill="url(#docGrad)" stroke="#cbd5e1" strokeWidth="1" />
    {/* Fold */}
    <path d="M36 16 v12 h12" fill="#e2e8f0" />
    {/* Lines */}
    <rect x="18" y="28" width="14" height="2" rx="1" fill="#94a3b8" />
    <rect x="18" y="34" width="24" height="2" rx="1" fill="#94a3b8" />
    <rect x="18" y="40" width="20" height="2" rx="1" fill="#94a3b8" />
    {/* Stamp */}
    <circle cx="38" cy="46" r="6" fill="#fecaca" opacity="0.8" />
    <circle cx="38" cy="46" r="4" stroke="#ef4444" strokeWidth="1.5" fill="none" />
  </svg>
);

export const OriginPortIcon = ({ className = "" }) => (
  <svg viewBox="0 0 64 64" className={className} width="100%" height="100%">
    <defs>
      <linearGradient id="crane" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#f59e0b" />
        <stop offset="100%" stopColor="#b45309" />
      </linearGradient>
    </defs>
    {/* Containers */}
    <rect x="12" y="36" width="14" height="8" fill="#3b82f6" />
    <rect x="28" y="36" width="14" height="8" fill="#10b981" />
    <rect x="12" y="46" width="14" height="8" fill="#f43f5e" />
    <rect x="28" y="46" width="14" height="8" fill="#3b82f6" />
    {/* Crane Base & Tower */}
    <rect x="46" y="20" width="6" height="34" fill="url(#crane)" />
    <path d="M42 54 h14 v4 h-14 z" fill="#475569" />
    {/* Crane Arm */}
    <path d="M16 26 L52 20 L52 24 L18 29 Z" fill="url(#crane)" />
    {/* Crane Wire & Hook */}
    <line x1="22" y1="28" x2="22" y2="40" stroke="#94a3b8" strokeWidth="1.5" />
    <path d="M20 40 h4 v2 h-4 z" fill="#475569" />
  </svg>
);

export const ContainerShipIcon = ({ className = "" }) => (
  <svg viewBox="0 0 64 64" className={className} width="100%" height="100%">
    <defs>
      <linearGradient id="shipHull" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#1e40af" />
        <stop offset="100%" stopColor="#172554" />
      </linearGradient>
    </defs>
    {/* Waves Back */}
    <path d="M4 52 Q16 48 32 52 T60 52" fill="none" stroke="#bfdbfe" strokeWidth="2" opacity="0.6" />
    {/* Hull */}
    <path d="M6 46 L50 46 L56 32 L4 32 Z" fill="url(#shipHull)" />
    {/* Cabin */}
    <rect x="44" y="20" width="10" height="12" fill="#cbd5e1" />
    <rect x="46" y="16" width="6" height="4" fill="#94a3b8" />
    {/* Containers */}
    <rect x="10" y="26" width="10" height="6" fill="#3b82f6" />
    <rect x="22" y="26" width="10" height="6" fill="#10b981" />
    <rect x="34" y="26" width="8" height="6" fill="#f43f5e" />
    
    <rect x="14" y="20" width="10" height="6" fill="#f43f5e" />
    <rect x="26" y="20" width="10" height="6" fill="#3b82f6" />
    {/* Waves Front */}
    <path d="M0 56 Q16 52 32 56 T64 56" fill="none" stroke="#60a5fa" strokeWidth="3" opacity="0.8" />
  </svg>
);

export const DestPortIcon = ({ className = "" }) => (
  <svg viewBox="0 0 64 64" className={className} width="100%" height="100%">
    <defs>
      <linearGradient id="craneBlue" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#3b82f6" />
        <stop offset="100%" stopColor="#1d4ed8" />
      </linearGradient>
    </defs>
    {/* Containers */}
    <rect x="22" y="46" width="14" height="8" fill="#10b981" />
    <rect x="38" y="46" width="14" height="8" fill="#f43f5e" />
    <rect x="38" y="36" width="14" height="8" fill="#3b82f6" />
    {/* Crane Tower */}
    <rect x="12" y="20" width="6" height="34" fill="url(#craneBlue)" />
    <path d="M8 54 h14 v4 h-14 z" fill="#475569" />
    {/* Crane Arm */}
    <path d="M48 26 L12 20 L12 24 L46 29 Z" fill="url(#craneBlue)" />
    {/* Crane Wire & Hook */}
    <line x1="42" y1="28" x2="42" y2="40" stroke="#94a3b8" strokeWidth="1.5" />
    <path d="M40 40 h4 v2 h-4 z" fill="#475569" />
  </svg>
);

export const CustomsIcon = ({ className = "" }) => (
  <svg viewBox="0 0 64 64" className={className} width="100%" height="100%">
    <defs>
      <linearGradient id="boxFront" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#fcd34d" />
        <stop offset="100%" stopColor="#d97706" />
      </linearGradient>
    </defs>
    {/* Open Box Back */}
    <polygon points="16,30 48,30 40,46 24,46" fill="#b45309" />
    {/* Goods inside */}
    <rect x="26" y="24" width="12" height="12" fill="#3b82f6" rx="2" />
    {/* Open Box Front */}
    <polygon points="12,34 52,34 44,54 20,54" fill="url(#boxFront)" />
    {/* Box Flaps */}
    <polygon points="12,34 24,24 16,24 8,30" fill="#f59e0b" />
    <polygon points="52,34 40,24 48,24 56,30" fill="#f59e0b" />
    {/* Magnifying Glass / Inspection */}
    <circle cx="44" cy="20" r="8" fill="rgba(255,255,255,0.6)" stroke="#cbd5e1" strokeWidth="2" />
    <line x1="39" y1="25" x2="34" y2="30" stroke="#475569" strokeWidth="3" strokeLinecap="round" />
  </svg>
);

export const WarehouseIcon = ({ className = "" }) => (
  <svg viewBox="0 0 64 64" className={className} width="100%" height="100%">
    <defs>
      <linearGradient id="whGrad" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#f8fafc" />
        <stop offset="100%" stopColor="#e2e8f0" />
      </linearGradient>
    </defs>
    {/* Back building */}
    <path d="M8 28 L32 20 L56 28 V54 H8 Z" fill="#cbd5e1" />
    {/* Front building */}
    <path d="M12 34 L32 26 L52 34 V54 H12 Z" fill="url(#whGrad)" stroke="#94a3b8" strokeWidth="1" />
    {/* Docks */}
    <rect x="20" y="42" width="8" height="12" fill="#475569" />
    <rect x="36" y="42" width="8" height="12" fill="#475569" />
    {/* Roof Trim */}
    <polygon points="10,34 32,25 54,34 52,36 32,28 12,36" fill="#3b82f6" />
    {/* Details */}
    <line x1="24" y1="42" x2="24" y2="54" stroke="#0f172a" strokeWidth="1" opacity="0.3" />
    <line x1="40" y1="42" x2="40" y2="54" stroke="#0f172a" strokeWidth="1" opacity="0.3" />
  </svg>
);

export const MiniDonutChart = ({ percent, colorClass }) => {
  // A simple SVG donut chart. percent is 0-100.
  // Using stroke-dasharray based on circumference of a circle with r=16 (approx 100).
  const c = 2 * Math.PI * 16;
  const dash = (percent / 100) * c;
  
  const strokeColor = colorClass === 'seller' ? '#3b82f6' : '#10b981';
  const bgColor = colorClass === 'seller' ? '#e0e7ff' : '#d1fae5';

  return (
    <svg viewBox="0 0 40 40" width="40" height="40" style={{ transform: 'rotate(-90deg)' }}>
      <circle cx="20" cy="20" r="16" fill="none" stroke={bgColor} strokeWidth="6" />
      <circle 
        cx="20" 
        cy="20" 
        r="16" 
        fill="none" 
        stroke={strokeColor} 
        strokeWidth="6" 
        strokeDasharray={`${dash} ${c}`}
        strokeLinecap="round"
        style={{ transition: 'stroke-dasharray 0.5s ease' }}
      />
    </svg>
  );
};
