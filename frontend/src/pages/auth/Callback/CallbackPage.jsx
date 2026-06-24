import { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';

export default function CallbackPage() {
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    // In the future, this is where we'd parse the authorization code from
    // the URL query params, exchange it for tokens via the backend,
    // and then redirect to the dashboard.
    
    // For now, it's just a placeholder that redirects to login if accessed.
    const timer = setTimeout(() => {
      navigate('/login');
    }, 2000);
    
    return () => clearTimeout(timer);
  }, [navigate, location]);

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      height: '100vh',
      fontFamily: 'Outfit, sans-serif',
      backgroundColor: '#F8FAFC'
    }}>
      <div className="auth-spinner" style={{ width: '40px', height: '40px', border: '3px solid rgba(15, 23, 42, 0.1)', borderTopColor: '#0F172A', borderRadius: '50%', animation: 'spin 1s linear infinite' }} />
      <style>
        {`@keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }`}
      </style>
      <h2 style={{ marginTop: '24px', color: '#0F172A' }}>Authenticating...</h2>
      <p style={{ color: '#64748B', marginTop: '8px' }}>Please wait while we complete your login.</p>
    </div>
  );
}
