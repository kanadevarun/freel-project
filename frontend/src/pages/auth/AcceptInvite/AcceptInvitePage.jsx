import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import toast from 'react-hot-toast';
import { Mail, User, Lock, Eye, EyeOff, CheckCircle2, ShieldCheck, Shield, UserPlus } from 'lucide-react';
import { validateInvite, acceptInvite } from '../../../services/authService';
import './AcceptInvitePage.css';

export default function AcceptInvitePage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');
  const navigate = useNavigate();

  const [isLoading, setIsLoading] = useState(false);
  const [isValidating, setIsValidating] = useState(true);
  const [inviteData, setInviteData] = useState(null);
  
  const [fullName, setFullName] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  useEffect(() => {
    const validateToken = async () => {
      if (!token) {
        toast.error('Invalid or missing invitation token.');
        setIsValidating(false);
        return;
      }
      try {
        const data = await validateInvite(token);
        setInviteData(data);
      } catch (err) {
        console.error(err);
        toast.error(err.message || 'This invitation is expired or invalid.');
      } finally {
        setIsValidating(false);
      }
    };

    validateToken();
  }, [token]);

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!fullName) {
      toast.error('Please enter your full name.');
      return;
    }
    if (!password) {
      toast.error('Password is required.');
      return;
    }
    if (password !== confirmPassword) {
      toast.error('Passwords do not match.');
      return;
    }

    setIsLoading(true);

    try {
      await acceptInvite({
        token,
        password,
        full_name: fullName,
      });

      toast.success('Account created successfully! Please log in.');
      navigate('/login');
    } catch (err) {
      console.error(err);
      toast.error(err.message || 'Failed to accept invitation. Please try again.');
      setIsLoading(false);
    }
  };

  if (isValidating) {
    return (
      <div className="auth-card" style={{ textAlign: 'center', padding: '40px' }}>
        <p>Validating invitation...</p>
      </div>
    );
  }

  if (!inviteData) {
    return (
      <div className="auth-card" style={{ textAlign: 'center', padding: '40px' }}>
        <h2>Invitation Invalid</h2>
        <p style={{ color: '#64748B', marginTop: '10px' }}>
          This invitation has expired, has already been used, or is no longer valid.
        </p>
      </div>
    );
  }

  return (
    <div className="auth-card">
      <div className="auth-header">
        <h1>Join LogisticsHQ</h1>
        <p>You've been invited to join</p>
        <span className="org-name-highlight">{inviteData.org_name}</span>
        <div className="verified-badge">
          <CheckCircle2 size={16} />
          <span>Invitation verified</span>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="auth-form">
        <div className="auth-field" style={{ marginBottom: '8px' }}>
          <label style={{ marginBottom: '4px' }}>Email address</label>
          <div className="auth-input-wrapper" style={{ position: 'relative' }}>
            <div className="input-icon" style={{ position: 'absolute', left: '14px', top: '50%', transform: 'translateY(-50%)', color: '#94a3b8', pointerEvents: 'none', display: 'flex', alignItems: 'center' }}>
              <Mail size={18} strokeWidth={2} />
            </div>
            <input
              type="email"
              value={inviteData.email}
              disabled
              className="auth-input"
              style={{ padding: '8px 14px 8px 40px', backgroundColor: '#F1F5F9', color: '#64748B', cursor: 'not-allowed' }}
            />
          </div>
          <span className="input-hint" style={{ fontSize: '0.75rem', color: '#64748B', display: 'block', marginTop: '4px' }}>This email is associated with your invitation.</span>
        </div>

        <div className="auth-field" style={{ marginBottom: '8px' }}>
          <label style={{ marginBottom: '4px' }}>Full name</label>
          <div className="auth-input-wrapper" style={{ position: 'relative' }}>
            <div className="input-icon" style={{ position: 'absolute', left: '14px', top: '50%', transform: 'translateY(-50%)', color: '#94a3b8', pointerEvents: 'none', display: 'flex', alignItems: 'center' }}>
              <User size={18} strokeWidth={2} />
            </div>
            <input
              type="text"
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              placeholder="Enter your full name"
              required
              className="auth-input"
              style={{ padding: '8px 14px 8px 40px' }}
              disabled={isLoading}
            />
          </div>
        </div>

        <div className="auth-field" style={{ marginBottom: '8px' }}>
          <label style={{ marginBottom: '4px' }}>Password</label>
          <div className="auth-input-wrapper" style={{ position: 'relative' }}>
            <div className="input-icon" style={{ position: 'absolute', left: '14px', top: '50%', transform: 'translateY(-50%)', color: '#94a3b8', pointerEvents: 'none', display: 'flex', alignItems: 'center' }}>
              <Lock size={18} strokeWidth={2} />
            </div>
            <input
              type={showPassword ? "text" : "password"}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Create a strong password"
              required
              className="auth-input"
              style={{ padding: '8px 40px 8px 40px' }}
              disabled={isLoading}
              minLength={8}
            />
            <button
              type="button"
              className="auth-input-toggle"
              onClick={() => setShowPassword(!showPassword)}
            >
              {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
            </button>
          </div>
          <span className="input-hint" style={{ fontSize: '0.75rem', color: '#64748B', display: 'block', marginTop: '4px' }}>Use at least 8 characters</span>
        </div>

        <div className="auth-field" style={{ marginBottom: '8px' }}>
          <label style={{ marginBottom: '4px' }}>Confirm password</label>
          <div className="auth-input-wrapper" style={{ position: 'relative' }}>
            <div className="input-icon" style={{ position: 'absolute', left: '14px', top: '50%', transform: 'translateY(-50%)', color: '#94a3b8', pointerEvents: 'none', display: 'flex', alignItems: 'center' }}>
              <Lock size={18} strokeWidth={2} />
            </div>
            <input
              type={showConfirmPassword ? "text" : "password"}
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Re-enter your password"
              required
              className="auth-input"
              style={{ padding: '8px 40px 8px 40px' }}
              disabled={isLoading}
              minLength={8}
            />
            <button
              type="button"
              className="auth-input-toggle"
              onClick={() => setShowConfirmPassword(!showConfirmPassword)}
            >
              {showConfirmPassword ? <EyeOff size={18} /> : <Eye size={18} />}
            </button>
          </div>
        </div>

        <button type="submit" className="auth-submit-btn" disabled={isLoading} style={{ marginTop: '8px', padding: '10px 20px' }}>
          {isLoading ? (
            <div className="auth-spinner" />
          ) : (
            <>
              <UserPlus size={18} />
              Create Account & Join Organization
            </>
          )}
        </button>

        <div className="info-alert-box" style={{ backgroundColor: '#F8FAFC', border: '1px solid #E2E8F0', borderRadius: '8px', padding: '8px 12px', display: 'flex', alignItems: 'flex-start', gap: '8px', marginTop: '12px' }}>
          <div className="info-alert-icon" style={{ color: '#2563EB', backgroundColor: '#EFF6FF', padding: '4px', borderRadius: '50%', display: 'flex' }}>
            <ShieldCheck size={18} />
          </div>
          <p className="info-alert-text" style={{ fontSize: '0.8rem', lineHeight: '1.4', color: '#475569', margin: 0 }}>
            Your account will be securely added to <strong>{inviteData.org_name}</strong> with the role assigned by your organization administrator.
          </p>
        </div>
      </form>

      <div className="secure-footer-links">
        <div className="secure-footer-item">
          <Lock size={14} />
          <span>Secure invitation. This link will expire.</span>
        </div>
        <div className="secure-footer-divider"></div>
        <div className="secure-footer-item">
          <Shield size={14} />
          <span>Your information is safe with us.</span>
        </div>
      </div>
    </div>
  );
}
