import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Shield, Truck, Factory, Globe2, Eye, EyeOff, CheckCircle2, Circle } from 'lucide-react';
import { register } from '../../../services/authService';
import './SignupPage.css';

const ROLES = [
  {
    id: 'CARRIER',
    icon: <img src="/images/auth-icons/carrier.png" alt="Carrier" width="48" height="48" style={{ objectFit: 'contain' }} />,
    title: 'Carrier',
    desc: 'I own trucks, fleet\nor transport capacity',
  },
  {
    id: 'SHIPPER',
    icon: <img src="/images/auth-icons/shipper.png" alt="Shipper" width="48" height="48" style={{ objectFit: 'contain' }} />,
    title: 'Shipper',
    desc: 'I need to ship cargo\nand goods',
  },
  {
    id: 'FREIGHT_FORWARDER',
    icon: <img src="/images/auth-icons/forwarder.png" alt="Freight Forwarder" width="48" height="48" style={{ objectFit: 'contain' }} />,
    title: 'Freight Forwarder',
    desc: 'I manage logistics\nfor my clients',
  },
];

export default function SignupPage() {
  const navigate = useNavigate();
  
  // State
  const [selectedRole, setSelectedRole] = useState(null);
  const [formData, setFormData] = useState({
    companyName: '',
    fullName: '',
    email: '',
    phone: '',
    password: '',
    confirmPassword: '',
  });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  // Handle inputs
  const handleChange = (e) => {
    setFormData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
    if (error) setError('');
  };

  // Password Validation Logic
  const p = formData.password;
  const validations = {
    length: p.length >= 8,
    uppercase: /[A-Z]/.test(p),
    number: /[0-9]/.test(p),
    special: /[^A-Za-z0-9]/.test(p),
  };

  const strengthScore = Object.values(validations).filter(Boolean).length;
  let strengthLabel = 'Weak';
  let strengthColor = '#EF4444'; // Red
  
  if (strengthScore === 4) {
    strengthLabel = 'Strong';
    strengthColor = '#10B981'; // Green
  } else if (strengthScore >= 2) {
    strengthLabel = 'Medium';
    strengthColor = '#F59E0B'; // Amber
  }
  
  if (p.length === 0) {
    strengthLabel = '';
  }

  // Handle submit
  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!selectedRole) {
      setError('Please select your role first.');
      window.scrollTo(0, 0);
      return;
    }
    if (formData.password !== formData.confirmPassword) {
      setError('Passwords do not match.');
      return;
    }
    if (strengthScore < 4) {
      setError('Please meet all password requirements.');
      return;
    }

    setIsLoading(true);
    setError('');

    try {
      await register({
        orgType: selectedRole,
        fullName: formData.fullName,
        companyName: formData.companyName,
        email: formData.email,
        password: formData.password,
      });
      navigate('/verify-email', { 
        state: { 
          email: formData.email,
          fullName: formData.fullName,
          companyName: formData.companyName,
          orgType: selectedRole
        } 
      });
    } catch (err) {
      setError(err.message || 'Something went wrong. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="signup-page animate-fade-in-up">
      <div className="auth-form-header">
        <h2>Create Your Account</h2>
        <p>Start managing freight in minutes.</p>
      </div>

      {error && (
        <div className="auth-error-banner">
          <span>⚠️</span>
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="signup-form-container">
        
        {/* ── STEP 1: CHOOSE ROLE ── */}
        <div className="signup-step-section">
          <h3 className="signup-step-title">1. Choose Your Role</h3>
          <div className="signup-roles-grid">
            {ROLES.map((role) => (
              <div
                key={role.id}
                className={`signup-role-card ${selectedRole === role.id ? 'selected' : ''}`}
                onClick={() => {
                  setSelectedRole(role.id);
                  setError('');
                }}
              >
                {selectedRole === role.id && (
                  <div className="signup-role-check">
                    <CheckCircle2 size={20} fill="white" color="#14B8A6" />
                  </div>
                )}
                <div className={`signup-role-icon ${selectedRole === role.id ? 'active' : ''}`}>
                  {role.icon}
                </div>
                <div className="signup-role-title">{role.title}</div>
                <div className="signup-role-desc">{role.desc}</div>
              </div>
            ))}
          </div>
        </div>

        {/* ── STEP 2: DETAILS ── */}
        <div className="signup-step-section">
          <h3 className="signup-step-title">2. Company & Account Details</h3>
          
          <div className="signup-form-grid">
            <div className="auth-field">
              <label htmlFor="companyName">Company Name *</label>
              <input
                id="companyName"
                name="companyName"
                type="text"
                className="auth-input"
                placeholder="Enter your company name"
                value={formData.companyName}
                onChange={handleChange}
                required
              />
            </div>

            <div className="auth-field">
              <label htmlFor="fullName">Your Full Name *</label>
              <input
                id="fullName"
                name="fullName"
                type="text"
                className="auth-input"
                placeholder="Enter your full name"
                value={formData.fullName}
                onChange={handleChange}
                required
              />
            </div>

            <div className="auth-field">
              <label htmlFor="email">Business Email *</label>
              <input
                id="email"
                name="email"
                type="email"
                className="auth-input"
                placeholder="name@company.com"
                value={formData.email}
                onChange={handleChange}
                required
              />
            </div>

            <div className="auth-field">
              <label htmlFor="phone">Phone Number <span className="optional">(Optional)</span></label>
              <div className="phone-input-wrapper">
                <div className="phone-country-code">
                  <span>🇮🇳</span>
                  <span>+91</span>
                </div>
                <input
                  id="phone"
                  name="phone"
                  type="tel"
                  className="auth-input phone-input"
                  placeholder="Enter phone number"
                  value={formData.phone}
                  onChange={handleChange}
                />
              </div>
            </div>

            <div className="auth-field">
              <label htmlFor="password">Password *</label>
              <div className="auth-input-wrapper">
                <input
                  id="password"
                  name="password"
                  type={showPassword ? 'text' : 'password'}
                  className="auth-input"
                  placeholder="Create a strong password"
                  value={formData.password}
                  onChange={handleChange}
                  required
                />
                <button
                  type="button"
                  className="auth-input-toggle"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </div>
            </div>

            <div className="auth-field">
              <label htmlFor="confirmPassword">Confirm Password *</label>
              <div className="auth-input-wrapper">
                <input
                  id="confirmPassword"
                  name="confirmPassword"
                  type={showConfirmPassword ? 'text' : 'password'}
                  className="auth-input"
                  placeholder="Confirm your password"
                  value={formData.confirmPassword}
                  onChange={handleChange}
                  required
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
          </div>

          {/* Password Strength Section */}
          <div className="signup-password-section">
            <div className="signup-password-meter-row">
              <span className="signup-password-meter-label">Password Strength</span>
              <div className="signup-password-bars">
                {[1, 2, 3, 4].map(level => (
                  <div 
                    key={level} 
                    className="signup-password-bar" 
                    style={{ 
                      backgroundColor: strengthScore >= level ? strengthColor : '#E2E8F0',
                    }}
                  />
                ))}
              </div>
              <span className="signup-password-status" style={{ color: strengthColor }}>
                {strengthLabel}
              </span>
            </div>

            <div className="signup-password-checklist">
              <div className="checklist-col">
                <span className={`checklist-item ${validations.length ? 'met' : ''}`}>
                  {validations.length ? <CheckCircle2 size={14} /> : <Circle size={14} />} At least 8 characters
                </span>
                <span className={`checklist-item ${validations.uppercase ? 'met' : ''}`}>
                  {validations.uppercase ? <CheckCircle2 size={14} /> : <Circle size={14} />} One uppercase letter
                </span>
              </div>
              <div className="checklist-col">
                <span className={`checklist-item ${validations.number ? 'met' : ''}`}>
                  {validations.number ? <CheckCircle2 size={14} /> : <Circle size={14} />} One number
                </span>
                <span className={`checklist-item ${validations.special ? 'met' : ''}`}>
                  {validations.special ? <CheckCircle2 size={14} /> : <Circle size={14} />} One special character
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* CTA */}
        <button type="submit" className="auth-submit-btn" disabled={isLoading || !selectedRole}>
          {isLoading ? <div className="auth-spinner" /> : (
            <>
              Create Account
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
            </>
          )}
        </button>

        {/* Security Footer */}
        <div className="auth-security-footer">
          <div className="auth-security-badge">
            <Shield size={18} />
            Protected by AWS Cognito
          </div>
          <div className="auth-security-text">
            Enterprise-grade security for your data
          </div>
        </div>

      </form>
    </div>
  );
}
