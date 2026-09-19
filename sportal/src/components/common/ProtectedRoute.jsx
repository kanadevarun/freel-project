import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { LoadingSpinner } from './LoadingSpinner';

export function ProtectedRoute({ requiredPermission, children }) {
  const { isAuthenticated, isBooting, hasPermission } = useAuth();
  const location = useLocation();

  if (isBooting) {
    return (
      <div className="flex h-screen w-full items-center justify-center bg-slate-50">
        <div className="flex flex-col items-center gap-3">
          <LoadingSpinner size="lg" />
          <p className="text-sm font-medium text-slate-500">Verifying SPortal session...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (requiredPermission && !hasPermission(requiredPermission)) {
    return <Navigate to="/unauthorized" state={{ requiredPermission }} replace />;
  }

  return children ? children : <Outlet />;
}
