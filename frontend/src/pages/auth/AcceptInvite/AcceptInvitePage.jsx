import { useState, useEffect } from 'react';
import { useSearchParams, Link, useNavigate } from 'react-router-dom';
import './AcceptInvitePage.css';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export default function AcceptInvitePage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get('token');

  const [fullName, setFullName] = useState('');
  const [password, setPassword] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    if (!token) {
      setError('No invitation token found in the URL. Please check your email link.');
    }
  }, [token]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!token) return;
    
    setIsSubmitting(true);
    setError('');

    try {
      const res = await fetch(`${API_BASE_URL}/auth/invite/accept`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          token,
          password,
          full_name: fullName,
        }),
      });

      const data = await res.json();
      
      if (!res.ok) {
        throw new Error(data.message || data.error || 'Failed to accept invitation.');
      }

      setSuccess(true);
    } catch (err) {
      setError(err.message);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (success) {
    return (
      <div className="accept-invite-page">
        <div className="ai-success">
          <div className="ai-success-icon">🎉</div>
          <h2>Welcome to Freel!</h2>
          <p>Your account has been successfully created and linked to your organization.</p>
          <button 
            className="ai-submit-btn" 
            onClick={() => navigate('/login')}
          >
            Log in to your account
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="accept-invite-page">
      <div className="ai-header">
        <h1>Accept Invitation</h1>
        <p>Complete your profile to join your team on Freel.</p>
      </div>

      <div className="ai-card">
        {error && (
          <div className="form-alert error" style={{ marginBottom: '1.5rem' }}>
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label" htmlFor="fullName">Full Name</label>
            <input
              id="fullName"
              type="text"
              className="form-input"
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              placeholder="John Doe"
              required
              disabled={isSubmitting || !token}
            />
          </div>

          <div className="form-group">
            <label className="form-label" htmlFor="password">Create Password</label>
            <input
              id="password"
              type="password"
              className="form-input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              required
              minLength={8}
              disabled={isSubmitting || !token}
            />
            <p style={{ fontSize: '0.8rem', color: 'var(--slate-500)', marginTop: '0.5rem' }}>
              Must be at least 8 characters.
            </p>
          </div>

          <button 
            type="submit" 
            className="ai-submit-btn"
            disabled={isSubmitting || !token || !fullName || password.length < 8}
          >
            {isSubmitting ? 'Joining...' : 'Join Organization'}
          </button>
        </form>
      </div>
      
      <div style={{ textAlign: 'center', marginTop: '1.5rem', fontSize: '0.9rem', color: 'var(--slate-500)' }}>
        Already have an account? <Link to="/login" style={{ color: 'var(--primary)', fontWeight: 600 }}>Log in</Link>
      </div>
    </div>
  );
}
