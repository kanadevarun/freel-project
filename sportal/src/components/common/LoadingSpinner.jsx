import React from 'react';

export function LoadingSpinner({ size = 'md', message = 'Loading...' }) {
  const sizeClasses = {
    sm: 'w-4 h-4 border-2',
    md: 'w-8 h-8 border-3',
    lg: 'w-12 h-12 border-4',
  };

  return (
    <div className="flex flex-col items-center justify-center p-8 text-slate-500">
      <div
        className={`${sizeClasses[size] || sizeClasses.md} rounded-full border-blue-600 border-t-transparent animate-spin mb-3`}
      ></div>
      {message && <p className="text-sm font-medium text-slate-500">{message}</p>}
    </div>
  );
}
