import { useState, useEffect } from 'react';
import { useLocation, useNavigate, Link } from 'react-router-dom';
import { resetPassword } from '../../../services/authService';
import AuthMessage from '../../../components/auth/AuthMessage/AuthMessage';

export default function ResetPasswordPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const email = location.state?.email || '';

  const [formData, setFormData] = useState({ code: '', newPassword: '', confirmPassword: '' });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [formErrors, setFormErrors] = useState({});
  const [successMsg, setSuccessMsg] = useState('');

  useEffect(() => {
    if (!email) {
      navigate('/forgot-password', { replace: true });
    }
  }, [email, navigate]);

  const handleChange = (e) => {
    setFormData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
    if (error) setError(null);
    if (formErrors[e.target.name]) {
      setFormErrors((prev) => ({ ...prev, [e.target.name]: null }));
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    const errors = {};
    if (!formData.code.trim() || formData.code.length < 6) {
      errors.code = 'Please enter the 6-digit reset code.';
    }
    
    const p = formData.newPassword;
    const strengthScore = [
      p.length >= 8,
      /[A-Z]/.test(p),
      /[0-9]/.test(p),
      /[^A-Za-z0-9]/.test(p)
    ].filter(Boolean).length;
    
    if (!p) {
      errors.newPassword = 'New password is required.';
    } else if (strengthScore < 4) {
      errors.newPassword = 'Password must meet the required strength criteria.';
    }
    
    if (!formData.confirmPassword) {
      errors.confirmPassword = 'Please confirm your new password.';
    } else if (p !== formData.confirmPassword) {
      errors.confirmPassword = 'Passwords do not match.';
    }

    if (Object.keys(errors).length > 0) {
      setFormErrors(errors);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      await resetPassword({
        email,
        code: formData.code,
        new_password: formData.newPassword,
      });
      
      setSuccessMsg('Password reset successfully! You can now log in.');
      
      // Auto redirect to login
      setTimeout(() => {
        navigate('/login', { state: { email } });
      }, 3000);
      
    } catch (err) {
      setError(err.message ? err : { message: 'Failed to reset password. Please try again.' });
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

      <AuthMessage 
        type="error" 
        message={error?.message} 
      />

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
              className={`auth-input ${formErrors.code ? 'input-error' : ''}`}
              placeholder="123456"
              value={formData.code}
              onChange={handleChange}
            />
            {formErrors.code && <span className="field-error-text">{formErrors.code}</span>}
          </div>

          <div className="auth-field">
            <label htmlFor="newPassword">New Password</label>
            <div className="auth-input-wrapper">
                <input
                  id="newPassword"
                  name="newPassword"
                  type={showPassword ? 'text' : 'password'}
                  className={`auth-input ${formErrors.newPassword ? 'input-error' : ''}`}
                  placeholder="••••••••••••"
                  value={formData.newPassword}
                  onChange={handleChange}
                />
                <button
                  type="button"
                  className="auth-input-toggle"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? '👁️' : '👁️‍🗨️'}
                </button>
              </div>
              {formErrors.newPassword && <span className="field-error-text">{formErrors.newPassword}</span>}
            </div>

          <div className="auth-field">
            <label htmlFor="confirmPassword">Confirm Password</label>
            <div className="auth-input-wrapper">
                <input
                  id="confirmPassword"
                  name="confirmPassword"
                  type={showConfirmPassword ? 'text' : 'password'}
                  className={`auth-input ${formErrors.confirmPassword ? 'input-error' : ''}`}
                  placeholder="••••••••••••"
                  value={formData.confirmPassword}
                  onChange={handleChange}
                />
                <button
                  type="button"
                  className="auth-input-toggle"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                >
                  {showConfirmPassword ? '👁️' : '👁️‍🗨️'}
                </button>
              </div>
              {formErrors.confirmPassword && <span className="field-error-text">{formErrors.confirmPassword}</span>}
            </div>

            <button type="submit" className="auth-submit-btn" disabled={isLoading}>
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
