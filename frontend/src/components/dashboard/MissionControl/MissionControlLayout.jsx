import React from 'react';
import './MissionControlLayout.css';

export default function MissionControlLayout({ children }) {
  // We expect children to be structured widgets
  return (
    <div className="mission-control-layout">
      {children}
    </div>
  );
}
