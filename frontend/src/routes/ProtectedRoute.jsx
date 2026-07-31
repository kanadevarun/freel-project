import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useRBAC } from '../context/RBACContext';

export default function ProtectedRoute({ requiredModule, requiredAction }) {
  const { isBooting, isAuthenticated, onboardingCompleted } = useAuth();
  const { can } = useRBAC();
  const location = useLocation();

  if (isBooting) {
    return (
      <div className="boot-screen">
        <div className="auth-spinner" />
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // Enforce onboarding logic for authenticated users
  if (location.pathname.startsWith('/dashboard') && !onboardingCompleted) {
    return <Navigate to="/onboarding" replace />;
  }

  if (location.pathname.startsWith('/onboarding') && onboardingCompleted) {
    return <Navigate to="/dashboard" replace />;
  }

  // Route-level RBAC check
  if (requiredModule && requiredAction) {
    if (!can(requiredModule, requiredAction)) {
      return <Navigate to="/dashboard/unauthorized" replace />;
    }
  }

  return <Outlet />;
}

