import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function PublicOnlyRoute() {
  const { isBooting, isAuthenticated, onboardingCompleted } = useAuth();

  if (isBooting) {
    return (
      <div className="boot-screen">
        <div className="auth-spinner" />
      </div>
    );
  }

  if (isAuthenticated) {
    if (onboardingCompleted) {
      return <Navigate to="/dashboard" replace />;
    } else {
      return <Navigate to="/onboarding" replace />;
    }
  }

  return <Outlet />;
}
