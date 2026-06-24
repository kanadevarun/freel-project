import { useState, useRef, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { verifyEmail, resendCode } from '../../../services/authService';
import { useAuth } from '../../../context/AuthContext';
import AuthMessage from '../../../components/auth/AuthMessage/AuthMessage';
import './VerifyEmailPage.css';

export default function VerifyEmailPage() {
  const location = useLocation();
  const navigate = useNavigate();
  
  const email = location.state?.email || '';

  // State
  const [code, setCode] = useState(['', '', '', '', '', '']);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [isSuccess, setIsSuccess] = useState(false);
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
    if (error) setError(null);

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
      setError(null);
    } catch (err) {
      setError(err.message ? err : { message: 'Failed to resend code.' });
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
    setError(null);

    try {
      // Confirm OTP via backend
      await verifyEmail({ 
        email, 
        code: fullCode,
      });
      
      setIsSuccess(true);
      
      // Delay navigation so user sees the success state
      setTimeout(() => {
        navigate('/login', { state: { email, verified: true } });
      }, 2500);
    } catch (err) {
      setError(err.message ? err : { message: 'Verification failed. Please try again.' });
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

      <AuthMessage 
        type="error" 
        message={error?.message} 
        actionLink={error?.action === 'SIGNUP' ? '/signup' : null}
        actionText={error?.action === 'SIGNUP' ? 'Sign Up Instead' : null}
      />

      {isSuccess ? (
        <div className="auth-error-banner" style={{ backgroundColor: '#DCFCE7', borderColor: '#86EFAC', color: '#166534', margin: '20px 0' }}>
          <span>✅</span>
          Email verified successfully! Redirecting to login...
        </div>
      ) : (
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
      )}
    </div>
  );
}
