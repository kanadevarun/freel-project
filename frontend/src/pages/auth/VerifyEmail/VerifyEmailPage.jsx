import { useState, useRef, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { confirmEmail, resendCode } from '../../../services/authService';
import { useAuth } from '../../../context/AuthContext';
import './VerifyEmailPage.css';

export default function VerifyEmailPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const { login } = useAuth();
  
  const email = location.state?.email || '';

  // State
  const [code, setCode] = useState(['', '', '', '', '', '']);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [countdown, setCountdown] = useState(60);

  // Refs for auto-focusing next input
  const inputRefs = useRef([]);

  // Redirect to signup if no email is in state
  useEffect(() => {
    if (!email) {
      navigate('/signup', { replace: true });
    }
  }, [email, navigate]);

  // Resend countdown timer
  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000);
      return () => clearTimeout(timer);
    }
  }, [countdown]);

  // Handle individual digit input
  const handleChange = (index, value) => {
    // Only allow numbers
    if (!/^[0-9]*$/.test(value)) return;

    const newCode = [...code];
    newCode[index] = value;
    setCode(newCode);
    setError('');

    // Auto-advance to next input
    if (value && index < 5) {
      inputRefs.current[index + 1].focus();
    }
  };

  // Handle backspace to go to previous input
  const handleKeyDown = (index, e) => {
    if (e.key === 'Backspace' && !code[index] && index > 0) {
      inputRefs.current[index - 1].focus();
    }
  };

  // Handle pasting a full 6-digit code
  const handlePaste = (e) => {
    e.preventDefault();
    const pastedData = e.clipboardData.getData('text/plain').trim();
    if (!/^[0-9]{6}$/.test(pastedData)) return;

    const digits = pastedData.split('');
    setCode(digits);
    inputRefs.current[5].focus();
  };

  const handleResend = async () => {
    if (countdown > 0) return;
    try {
      await resendCode({ email });
      setCountdown(60);
      setError('');
    } catch (err) {
      setError(err.message || 'Failed to resend code.');
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const fullCode = code.join('');
    if (fullCode.length !== 6) {
      setError('Please enter all 6 digits.');
      return;
    }

    setIsLoading(true);
    setError('');

    try {
      // Confirm OTP and pass the context data to the mock backend
      const { user, org, memberRole } = await confirmEmail({ 
        email, 
        code: fullCode,
        fullName: location.state?.fullName,
        companyName: location.state?.companyName,
        orgType: location.state?.orgType
      });
      
      // Log user in globally
      await login(user, org, memberRole);
      
      // Go to onboarding
      navigate('/onboarding');
    } catch (err) {
      setError(err.message || 'Verification failed. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="verify-page animate-fade-in-up">
      <div className="auth-form-header">
        <div className="verify-icon">📧</div>
        <h2>Check your inbox</h2>
        <p>
          We've sent a 6-digit verification code to<br />
          <strong>{email}</strong>
        </p>
      </div>

      {error && (
        <div className="auth-error-banner">
          <span>⚠️</span>
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="verify-form">
        <div className="verify-code-inputs">
          {code.map((digit, idx) => (
            <input
              key={idx}
              ref={(el) => (inputRefs.current[idx] = el)}
              type="text"
              inputMode="numeric"
              maxLength={1}
              value={digit}
              onChange={(e) => handleChange(idx, e.target.value)}
              onKeyDown={(e) => handleKeyDown(idx, e)}
              onPaste={idx === 0 ? handlePaste : undefined}
              className={`verify-input ${error ? 'error' : ''}`}
            />
          ))}
        </div>

        <button type="submit" className="auth-submit-btn" disabled={isLoading || code.includes('')}>
          {isLoading ? <div className="auth-spinner" /> : 'Verify Email →'}
        </button>

        <div className="verify-footer">
          <p>
            Didn't receive it?{' '}
            {countdown > 0 ? (
              <span className="verify-countdown">Resend in {countdown}s</span>
            ) : (
              <button type="button" onClick={handleResend} className="verify-resend-btn">
                Resend Code
              </button>
            )}
          </p>
          <p>
            <button type="button" onClick={() => navigate('/signup')} className="verify-change-email">
              Change email address
            </button>
          </p>
        </div>
      </form>
    </div>
  );
}
