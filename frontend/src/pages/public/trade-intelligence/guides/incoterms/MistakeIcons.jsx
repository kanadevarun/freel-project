import React from 'react';

export const ShipIllustration = ({ className = "" }) => (
  <svg viewBox="0 0 400 300" className={className} width="100%" height="100%" preserveAspectRatio="xMidYMid slice">
    <defs>
      <linearGradient id="skyGrad" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#fde047" />
        <stop offset="50%" stopColor="#fef08a" />
        <stop offset="100%" stopColor="#bae6fd" />
      </linearGradient>
      <linearGradient id="waterGrad" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#0369a1" />
        <stop offset="100%" stopColor="#075985" />
      </linearGradient>
      <linearGradient id="hullGrad" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#1e3a8a" />
        <stop offset="100%" stopColor="#172554" />
      </linearGradient>
    </defs>
    {/* Sky */}
    <rect width="400" height="200" fill="url(#skyGrad)" />
    {/* Sun */}
    <circle cx="80" cy="150" r="30" fill="#fff" opacity="0.8" filter="blur(8px)" />
    <circle cx="80" cy="150" r="20" fill="#fef08a" />
    {/* Water */}
    <rect y="200" width="400" height="100" fill="url(#waterGrad)" />
    {/* Ship Hull */}
    <path d="M40 240 L360 240 L380 180 L20 180 Z" fill="url(#hullGrad)" />
    <rect x="20" y="170" width="360" height="10" fill="#b91c1c" />
    {/* Containers */}
    <rect x="60" y="130" width="40" height="40" fill="#3b82f6" />
    <rect x="105" y="130" width="40" height="40" fill="#10b981" />
    <rect x="150" y="130" width="40" height="40" fill="#f43f5e" />
    <rect x="195" y="130" width="40" height="40" fill="#f59e0b" />
    <rect x="240" y="130" width="40" height="40" fill="#3b82f6" />
    
    <rect x="60" y="90" width="40" height="40" fill="#f43f5e" />
    <rect x="105" y="90" width="40" height="40" fill="#3b82f6" />
    <rect x="150" y="90" width="40" height="40" fill="#f59e0b" />
    <rect x="195" y="90" width="40" height="40" fill="#10b981" />
    
    {/* Ship Bridge */}
    <rect x="300" y="80" width="60" height="90" fill="#e2e8f0" />
    <rect x="290" y="120" width="80" height="10" fill="#94a3b8" />
    <rect x="310" y="90" width="40" height="20" fill="#0f172a" />
    
    {/* Crane Silhouette in background */}
    <path d="M10 200 L10 50 L30 50 L30 200 Z" fill="rgba(15, 23, 42, 0.4)" />
    <path d="M10 70 L120 50 L120 60 L30 80 Z" fill="rgba(15, 23, 42, 0.4)" />
  </svg>
);

export const SuspendedContainer = ({ className = "" }) => (
  <svg viewBox="0 0 200 200" className={className} width="100%" height="100%">
    <defs>
      <linearGradient id="redCont" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#ef4444" />
        <stop offset="100%" stopColor="#b91c1c" />
      </linearGradient>
    </defs>
    {/* Crane hook & cables */}
    <path d="M95 10 L105 10 L105 30 L115 30 L100 50 L85 30 L95 30 Z" fill="#475569" />
    <line x1="100" y1="50" x2="40" y2="100" stroke="#94a3b8" strokeWidth="2" />
    <line x1="100" y1="50" x2="160" y2="100" stroke="#94a3b8" strokeWidth="2" />
    {/* Container */}
    <rect x="30" y="100" width="140" height="60" fill="url(#redCont)" rx="4" />
    {/* Ribs */}
    <rect x="40" y="100" width="8" height="60" fill="rgba(0,0,0,0.15)" />
    <rect x="60" y="100" width="8" height="60" fill="rgba(0,0,0,0.15)" />
    <rect x="80" y="100" width="8" height="60" fill="rgba(0,0,0,0.15)" />
    <rect x="100" y="100" width="8" height="60" fill="rgba(0,0,0,0.15)" />
    <rect x="120" y="100" width="8" height="60" fill="rgba(0,0,0,0.15)" />
    <rect x="140" y="100" width="8" height="60" fill="rgba(0,0,0,0.15)" />
  </svg>
);

export const InvoiceStack = ({ className = "" }) => (
  <svg viewBox="0 0 100 100" className={className} width="100%" height="100%">
    <defs>
      <linearGradient id="board" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stopColor="#3b82f6" />
        <stop offset="100%" stopColor="#1d4ed8" />
      </linearGradient>
      <linearGradient id="coin" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#fde047" />
        <stop offset="100%" stopColor="#ca8a04" />
      </linearGradient>
    </defs>
    {/* Clipboard */}
    <rect x="20" y="15" width="50" height="70" rx="4" fill="url(#board)" transform="rotate(-5 45 50)" />
    <rect x="25" y="20" width="40" height="60" fill="#ffffff" transform="rotate(-5 45 50)" />
    <rect x="35" y="10" width="20" height="8" rx="2" fill="#94a3b8" transform="rotate(-5 45 50)" />
    {/* Lines */}
    <line x1="32" y1="35" x2="58" y2="33" stroke="#cbd5e1" strokeWidth="2" />
    <line x1="30" y1="45" x2="56" y2="43" stroke="#ef4444" strokeWidth="3" />
    <line x1="28" y1="55" x2="54" y2="53" stroke="#cbd5e1" strokeWidth="2" />
    {/* Coins */}
    <ellipse cx="75" cy="75" rx="15" ry="5" fill="url(#coin)" />
    <rect x="60" y="65" width="30" height="10" fill="url(#coin)" />
    <ellipse cx="75" cy="65" rx="15" ry="5" fill="#fef08a" />
    
    <rect x="60" y="55" width="30" height="10" fill="url(#coin)" />
    <ellipse cx="75" cy="55" rx="15" ry="5" fill="#fef08a" />
  </svg>
);

export const AirplaneGraphic = ({ className = "" }) => (
  <svg viewBox="0 0 200 100" className={className} width="100%" height="100%">
    {/* Plane Body */}
    <path d="M20 50 Q100 40 160 50 Q180 50 180 60 Q100 65 20 60 Z" fill="#cbd5e1" />
    {/* Tail */}
    <path d="M30 50 L10 20 L40 20 L50 50 Z" fill="#f59e0b" />
    {/* Wing Front */}
    <path d="M120 55 L90 80 L140 80 L140 55 Z" fill="#94a3b8" />
    {/* Windows */}
    <circle cx="165" cy="53" r="2" fill="#1e293b" />
    <line x1="60" y1="53" x2="140" y2="53" stroke="#1e293b" strokeWidth="2" strokeDasharray="4 2" />
  </svg>
);

export const DoorDelivery = ({ className = "" }) => (
  <svg viewBox="0 0 200 200" className={className} width="100%" height="100%">
    <defs>
      <linearGradient id="doorGrad" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#166534" />
        <stop offset="100%" stopColor="#14532d" />
      </linearGradient>
    </defs>
    {/* Wall */}
    <rect x="20" y="20" width="160" height="180" fill="#f8fafc" />
    {/* Door */}
    <rect x="60" y="40" width="80" height="160" fill="url(#doorGrad)" />
    <rect x="70" y="50" width="25" height="30" fill="#22c55e" opacity="0.3" />
    <rect x="105" y="50" width="25" height="30" fill="#22c55e" opacity="0.3" />
    <circle cx="130" cy="120" r="4" fill="#fbbf24" />
    {/* Boxes */}
    <rect x="30" y="140" width="40" height="40" fill="#d97706" />
    <polygon points="30,140 50,130 90,130 70,140" fill="#f59e0b" />
    <polygon points="70,140 90,130 90,170 70,180" fill="#b45309" />
    <line x1="50" y1="140" x2="50" y2="180" stroke="#b45309" strokeWidth="2" />
    
    <rect x="40" y="100" width="30" height="30" fill="#d97706" />
    <polygon points="40,100 55,90 85,90 70,100" fill="#f59e0b" />
    <polygon points="70,100 85,90 85,120 70,130" fill="#b45309" />
  </svg>
);

export const ShieldDocs = ({ className = "" }) => (
  <svg viewBox="0 0 100 100" className={className} width="100%" height="100%">
    {/* Document */}
    <rect x="20" y="10" width="50" height="70" fill="#ffffff" stroke="#cbd5e1" strokeWidth="2" />
    <rect x="30" y="25" width="30" height="4" fill="#94a3b8" />
    <rect x="30" y="35" width="20" height="4" fill="#94a3b8" />
    <rect x="30" y="45" width="25" height="4" fill="#94a3b8" />
    {/* Shield Overlap */}
    <path d="M45 40 L75 40 L75 60 C75 75 60 85 45 90 C30 85 15 75 15 60 L15 40 Z" fill="#10b981" />
    <path d="M35 60 L45 70 L65 50" fill="none" stroke="#ffffff" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
);

export const BlueShield = ({ className = "" }) => (
  <svg viewBox="0 0 200 200" className={className} width="100%" height="100%">
    <defs>
      <linearGradient id="bigShield" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stopColor="#60a5fa" />
        <stop offset="100%" stopColor="#2563eb" />
      </linearGradient>
    </defs>
    {/* Boxes in background */}
    <rect x="40" y="120" width="60" height="60" fill="#94a3b8" opacity="0.3" />
    <rect x="100" y="100" width="60" height="80" fill="#cbd5e1" opacity="0.3" />
    {/* Big Shield */}
    <path d="M100 20 L160 40 L160 100 C160 140 130 170 100 190 C70 170 40 140 40 100 L40 40 Z" fill="url(#bigShield)" />
    <path d="M70 100 L90 120 L130 80" fill="none" stroke="#ffffff" strokeWidth="12" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
);
