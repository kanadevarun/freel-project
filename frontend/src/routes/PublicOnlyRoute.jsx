import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function PublicOnlyRoute() {
  const { isBooting, isAuthenticated } = useAuth();

  if (isBooting) {
    return (
      <div className="boot-screen">
        <div className="auth-spinner" />
      </div>
    );
  }

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  return <Outlet />;
}
