import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { forgotPassword } from '../../../services/authService';
import AuthMessage from '../../../components/auth/AuthMessage/AuthMessage';

export default function ForgotPasswordPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [error, setError] = useState(null);
  const [formErrors, setFormErrors] = useState({});

  const handleChange = (e) => {
    setEmail(e.target.value);
    if (error) setError(null);
    if (formErrors.email) setFormErrors({});
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    try {
      await forgotPassword({ email });
      setIsSuccess(true);
      // Wait a moment before redirecting to reset page
      setTimeout(() => {
        navigate('/reset-password', { state: { email } });
      }, 4000);
    } catch (err) {
      setError(err.message ? err : { message: 'An error occurred. Please try again.' });
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

      <AuthMessage 
        type="error" 
        message={error?.message} 
      />

      {isSuccess ? (
        <div className="auth-error-banner" style={{ backgroundColor: '#DCFCE7', borderColor: '#86EFAC', color: '#166534' }}>
          <span>✅</span>
          If an account exists for this email, we’ve sent a 6-digit reset code. Please check your inbox and spam folder.
        </div>
      ) : (
        <form onSubmit={handleSubmit}>
          <div className="auth-field">
            <label htmlFor="email">Work Email</label>
            <input
              id="email"
              type="email"
              className={`auth-input ${formErrors.email ? 'input-error' : ''}`}
              placeholder="name@company.com"
              value={email}
              onChange={handleChange}
            />
            {formErrors.email && <span className="field-error-text">{formErrors.email}</span>}
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
