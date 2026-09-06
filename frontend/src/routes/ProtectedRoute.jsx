import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useRBAC } from '../context/RBACContext';

export default function ProtectedRoute({ requiredModule, requiredAction }) {
  const { isBooting, isAuthenticated } = useAuth();
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
  // Route-level RBAC check
  if (requiredModule && requiredAction) {
    if (!can(requiredModule, requiredAction)) {
      return <Navigate to="/dashboard/unauthorized" replace />;
    }
  }

  return <Outlet />;
}

