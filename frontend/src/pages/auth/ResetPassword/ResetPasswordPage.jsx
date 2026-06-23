import { useState, useEffect } from 'react';
import { useLocation, useNavigate, Link } from 'react-router-dom';
import { resetPassword } from '../../../services/authService';

export default function ResetPasswordPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const email = location.state?.email || '';

  const [formData, setFormData] = useState({ code: '', newPassword: '' });
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [successMsg, setSuccessMsg] = useState('');

  useEffect(() => {
    if (!email) {
      navigate('/forgot-password', { replace: true });
    }
  }, [email, navigate]);

  const handleChange = (e) => {
    setFormData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
    if (error) setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      await resetPassword({
        email,
        code: formData.code,
        newPassword: formData.newPassword,
      });
      
      setSuccessMsg('Password reset successfully! You can now log in.');
      
      // Auto redirect to login
      setTimeout(() => {
        navigate('/login');
      }, 3000);
      
    } catch (err) {
      setError(err.message || 'Failed to reset password. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="reset-page animate-fade-in-up">
      <div className="auth-form-header">
        <h2>Choose a new password</h2>
        <p>Enter the 6-digit code sent to <strong>{email}</strong> and your new password.</p>
      </div>

      {error && (
        <div className="auth-error-banner">
          <span>⚠️</span>
          {error}
        </div>
      )}

      {successMsg ? (
        <div className="auth-error-banner" style={{ backgroundColor: '#DCFCE7', borderColor: '#86EFAC', color: '#166534' }}>
          <span>✅</span>
          {successMsg}
        </div>
      ) : (
        <form onSubmit={handleSubmit}>
          <div className="auth-field">
            <label htmlFor="code">Reset Code (6 digits)</label>
            <input
              id="code"
              name="code"
              type="text"
              inputMode="numeric"
              maxLength={6}
              className="auth-input"
              placeholder="123456"
              value={formData.code}
              onChange={handleChange}
              required
            />
          </div>

          <div className="auth-field">
            <label htmlFor="newPassword">New Password</label>
            <div className="auth-input-wrapper">
              <input
                id="newPassword"
                name="newPassword"
                type={showPassword ? 'text' : 'password'}
                className="auth-input"
                placeholder="••••••••••••"
                value={formData.newPassword}
                onChange={handleChange}
                required
              />
              <button
                type="button"
                className="auth-input-toggle"
                onClick={() => setShowPassword(!showPassword)}
              >
                {showPassword ? '👁️' : '👁️‍🗨️'}
              </button>
            </div>
          </div>

          <button type="submit" className="auth-submit-btn" disabled={isLoading || formData.code.length < 6 || !formData.newPassword}>
            {isLoading ? <div className="auth-spinner" /> : 'Update Password'}
          </button>
        </form>
      )}
      
      <p className="auth-footer-link">
        <Link to="/login">← Back to Log In</Link>
      </p>
    </div>
  );
}
