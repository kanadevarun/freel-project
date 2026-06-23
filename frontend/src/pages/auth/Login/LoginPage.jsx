import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Eye, EyeOff } from 'lucide-react';
import { login as apiLogin } from '../../../services/authService';
import { useAuth } from '../../../context/AuthContext';
import './LoginPage.css';

export default function LoginPage() {
  const navigate = useNavigate();
  const { login: setGlobalAuth } = useAuth();

  const [formData, setFormData] = useState({ email: '', password: '' });
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [rememberMe, setRememberMe] = useState(false);

  const handleChange = (e) => {
    setFormData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
    if (error) setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      // Call mock auth service
      const { user, org, memberRole } = await apiLogin(formData);
      
      // Update global context (localStorage)
      await setGlobalAuth(user, org, memberRole);

      // Route based on onboarding status
      if (!org.onboardingCompleted) {
        navigate('/onboarding');
      } else {
        navigate('/dashboard'); // Mock destination
      }
    } catch (err) {
      if (err.code === 'EMAIL_NOT_VERIFIED') {
        // Automatically route to verification if they aren't verified
        navigate('/verify-email', { state: { email: formData.email } });
      } else {
        setError(err.message || 'Incorrect email or password.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleGoogleLogin = () => {
    // Mock google login
    alert("Google OAuth flow initiated...");
  };

  return (
    <div className="login-fullscreen-layout">
      
      {/* ── LEFT PANEL (55%) ── */}
      <div className="login-left-panel">
        <div 
          className="login-left-bg"
          style={{ backgroundImage: "url('/images/authentication/left-background.png')" }}
        />
        <div className="login-left-overlay" />

        <div className="login-left-content">
          <div className="login-left-main">
            <span className="login-badge">GLOBAL FREIGHT PLATFORM</span>
            <h1 className="login-heading">Welcome Back To Freel</h1>
            <p className="login-description">
              Manage shipments, compare freight rates, collaborate with your logistics team, and monitor global cargo movements from a single platform.
            </p>

            <div className="login-trust-metrics">
              <div className="login-metric-card">
                <span className="metric-number">220+</span>
                <span className="metric-label">Countries<br/>Connected</span>
              </div>
              <div className="login-metric-card">
                <span className="metric-number">500+</span>
                <span className="metric-label">Airlines</span>
              </div>
              <div className="login-metric-card">
                <span className="metric-number">150+</span>
                <span className="metric-label">Ocean<br/>Carriers</span>
              </div>
              <div className="login-metric-card">
                <span className="metric-number">24/7</span>
                <span className="metric-label">Shipment<br/>Visibility</span>
              </div>
            </div>
          </div>

          <div className="login-left-bottom">
            <p className="trusted-by-text">Trusted by logistics teams worldwide</p>
            <div className="trusted-logos">
              <span>DHL</span>
              <span>FedEx</span>
              <span>Maersk</span>
              <span>Kuehne+Nagel</span>
              <span>DB Schenker</span>
            </div>
          </div>
        </div>
      </div>

      {/* ── RIGHT PANEL (45%) ── */}
      <div className="login-right-panel">
        <div className="login-card-container">
          
          {/* Logo & Header */}
          <div className="login-card-header">
            <div className="login-logo">
              <span className="login-logo-icon">⚡</span> FREEL
            </div>
            <h2>Welcome Back</h2>
            <p>Sign in to access your freight workspace.</p>
          </div>

          {error && (
            <div className="login-error-banner">
              <span>⚠️</span>
              {error}
            </div>
          )}

          {/* Form */}
          <form className="login-form" onSubmit={handleSubmit}>
            <div className="login-field">
              <label htmlFor="email">Work Email</label>
              <input
                id="email"
                name="email"
                type="email"
                className="login-input"
                placeholder="Enter your email"
                value={formData.email}
                onChange={handleChange}
                required
                autoComplete="email"
              />
            </div>

            <div className="login-field">
              <label htmlFor="password">Password</label>
              <div className="login-input-wrapper">
                <input
                  id="password"
                  name="password"
                  type={showPassword ? 'text' : 'password'}
                  className="login-input"
                  placeholder="Enter password"
                  value={formData.password}
                  onChange={handleChange}
                  required
                  autoComplete="current-password"
                />
                <button
                  type="button"
                  className="login-password-toggle"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </div>
            </div>

            <div className="login-remember-row">
              <label className="login-checkbox-label">
                <input 
                  type="checkbox" 
                  checked={rememberMe}
                  onChange={(e) => setRememberMe(e.target.checked)}
                />
                <span className="checkmark"></span>
                Remember Me
              </label>
              <Link to="/forgot-password" className="login-forgot-link">
                Forgot Password?
              </Link>
            </div>

            <button type="submit" className="login-submit-btn" disabled={isLoading}>
              {isLoading ? <div className="login-spinner" /> : 'Sign In →'}
            </button>
          </form>

          {/* Divider */}
          <div className="login-divider">
            <span>OR</span>
          </div>

          {/* Secondary Auth */}
          <button type="button" className="login-google-btn" onClick={handleGoogleLogin}>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M22.56 12.25C22.56 11.47 22.49 10.72 22.36 10H12V14.26H17.92C17.66 15.63 16.88 16.81 15.71 17.59V20.34H19.28C21.36 18.42 22.56 15.6 22.56 12.25Z" fill="#4285F4"/>
              <path d="M12 23C14.97 23 17.46 22.02 19.28 20.34L15.71 17.59C14.73 18.25 13.48 18.66 12 18.66C9.13 18.66 6.71 16.73 5.84 14.12H2.15V16.98C3.96 20.58 7.7 23 12 23Z" fill="#34A853"/>
              <path d="M5.84 14.12C5.62 13.46 5.49 12.75 5.49 12C5.49 11.25 5.62 10.54 5.84 9.88V7.02H2.15C1.4 8.52 1 10.22 1 12C1 13.78 1.4 15.48 2.15 16.98L5.84 14.12Z" fill="#FBBC05"/>
              <path d="M12 5.34C13.62 5.34 15.07 5.9 16.21 6.99L19.36 3.84C17.46 2.06 14.97 1 12 1C7.7 1 3.96 3.42 2.15 7.02L5.84 9.88C6.71 7.27 9.13 5.34 12 5.34Z" fill="#EA4335"/>
            </svg>
            Continue with Google
          </button>

          {/* Footer Create Account */}
          <p className="login-footer-text">
            Don't have an account? <Link to="/signup" className="create-account-link">Create Account</Link>
          </p>
        </div>

        {/* Security Badges */}
        <div className="login-security-badges">
          <span>🔒 SOC2 Certified</span>
          <span className="badge-dot">•</span>
          <span>🛡️ 256-bit Encryption</span>
          <span className="badge-dot">•</span>
          <span>☁️ AWS Hosted</span>
          <span className="badge-dot">•</span>
          <span>⚡ 99.99% Uptime</span>
        </div>
      </div>

    </div>
  );
}
