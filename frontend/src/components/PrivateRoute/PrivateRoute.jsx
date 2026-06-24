import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { onboardingStorage } from '../../utils/onboardingStorage';

/**
 * PrivateRoute — wrapper for authenticated pages.
 * If user is not logged in, redirects to /login.
 * Also checks if onboarding is completed.
 */
export default function PrivateRoute() {
  const { isAuthenticated, org, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <div style={{ height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <div className="auth-spinner" style={{ borderColor: '#00BFA5', borderTopColor: 'transparent' }} />
      </div>
    );
  }

  if (!isAuthenticated) {
    // Redirect to login but save the attempted URL
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // Check onboarding state
  const isComplete = onboardingStorage.isOnboardingCompleted();

  // Force onboarding if not completed (unless they are already on the onboarding page)
  if (!isComplete && location.pathname !== '/onboarding') {
    return <Navigate to="/onboarding" replace />;
  }

  // Prevent completed users from manually revisiting onboarding
  if (isComplete && location.pathname === '/onboarding') {
    return <Navigate to="/dashboard" replace />;
  }

  return <Outlet />;
}
