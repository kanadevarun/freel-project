import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { forgotPassword } from '../../../services/authService';

export default function ForgotPasswordPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      await forgotPassword({ email });
      setIsSuccess(true);
      // Wait a moment before redirecting to reset page
      setTimeout(() => {
        navigate('/reset-password', { state: { email } });
      }, 2500);
    } catch (err) {
      // For security, forgotPassword rarely throws unless it's a network error
      console.error('Network error during forgot password:', err);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="forgot-page animate-fade-in-up">
      <div className="auth-form-header">
        <h2>Reset your password</h2>
        <p>Enter your work email address and we'll send you a 6-digit code to reset your password.</p>
      </div>

      {isSuccess ? (
        <div className="auth-error-banner" style={{ backgroundColor: '#DCFCE7', borderColor: '#86EFAC', color: '#166534' }}>
          <span>✅</span>
          Reset code sent! Redirecting you to enter it...
        </div>
      ) : (
        <form onSubmit={handleSubmit}>
          <div className="auth-field">
            <label htmlFor="email">Work Email</label>
            <input
              id="email"
              type="email"
              className="auth-input"
              placeholder="name@company.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>

          <button type="submit" className="auth-submit-btn" disabled={isLoading || !email}>
            {isLoading ? <div className="auth-spinner" /> : 'Send Reset Code'}
          </button>

          <p className="auth-footer-link">
            Remember your password? <Link to="/login">Log In</Link>
          </p>
        </form>
      )}
    </div>
  );
}
