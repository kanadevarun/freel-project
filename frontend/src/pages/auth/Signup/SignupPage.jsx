import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Shield, Eye, EyeOff, CheckCircle2, Circle } from 'lucide-react';
import { signup } from '../../../services/authService';
import AuthMessage from '../../../components/auth/AuthMessage/AuthMessage';
import { COUNTRIES, getFlagEmoji } from '../../../utils/countries';
import './SignupPage.css';

// Hardcoded for Phase 1 — only Freight Forwarders onboard now.
// When we add Shippers/Carriers, this becomes a UI selection again.
const ORG_TYPE = 'FREIGHT_FORWARDER';

export default function SignupPage() {
  const navigate = useNavigate();

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
  const [error, setError] = useState(null);
  const [formErrors, setFormErrors] = useState({});
  const [isCountryDropdownOpen, setIsCountryDropdownOpen] = useState(false);
  const [selectedCountry, setSelectedCountry] = useState(COUNTRIES[3]); // Default: India (+91)

  const handleChange = (e) => {
    setFormData((prev) => ({ ...prev, [e.target.name]: e.target.value }));
    if (error) setError(null);
    if (formErrors[e.target.name]) {
      setFormErrors((prev) => ({ ...prev, [e.target.name]: null }));
    }
  };

  // Password strength
  const p = formData.password;
  const validations = {
    length:    p.length >= 8,
    uppercase: /[A-Z]/.test(p),
    number:    /[0-9]/.test(p),
    special:   /[^A-Za-z0-9]/.test(p),
  };
  const strengthScore = Object.values(validations).filter(Boolean).length;
  let strengthLabel = p.length === 0 ? '' : 'Weak';
  let strengthColor = '#EF4444';
  if (strengthScore === 4) { strengthLabel = 'Strong'; strengthColor = '#10B981'; }
  else if (strengthScore >= 2) { strengthLabel = 'Medium'; strengthColor = '#F59E0B'; }

  const validateForm = () => {
    const errors = {};
    if (!formData.companyName.trim()) errors.companyName = 'Company name is required.';
    if (!formData.fullName.trim())    errors.fullName    = 'Full name is required.';
    if (!formData.email.trim()) {
      errors.email = 'Business email is required.';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      errors.email = 'Please enter a valid email address.';
    }
    if (!formData.password) {
      errors.password = 'Password is required.';
    } else if (strengthScore < 4) {
      errors.password = 'Password must meet all strength criteria.';
    }
    if (!formData.confirmPassword) {
      errors.confirmPassword = 'Please confirm your password.';
    } else if (formData.password !== formData.confirmPassword) {
      errors.confirmPassword = 'Passwords do not match.';
    }
    if (formData.phone?.trim()) {
      const digits = formData.phone.replace(/\D/g, '');
      if (digits.length < 6 || digits.length > 15) errors.phone = 'Please enter a valid phone number.';
    }
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validateForm()) { window.scrollTo(0, 0); return; }

    setIsLoading(true);
    setError(null);

    try {
      await signup({
        role:         ORG_TYPE.toLowerCase(),   // always "freight_forwarder" for now
        full_name:    formData.fullName,
        company_name: formData.companyName,
        email:        formData.email,
        password:     formData.password,
      });
      navigate('/verify-email', {
        state: {
          email:       formData.email,
          fullName:    formData.fullName,
          companyName: formData.companyName,
          orgType:     ORG_TYPE,
        },
      });
    } catch (err) {
      setError(err.message ? err : { message: 'Something went wrong. Please try again.' });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="signup-page animate-fade-in-up">
      <div className="auth-form-header">
        <div className="signup-org-badge">
          <img src="/images/auth-icons/forwarder.png" alt="Freight Forwarder" width="28" height="28" />
          <span>Freight Forwarder</span>
        </div>
        <h2>Create Your Account</h2>
        <p>Start finding leads and winning freight deals in minutes.</p>
      </div>

      <AuthMessage
        type="error"
        message={error?.message}
        actionLink={error?.action === 'LOGIN' ? '/login' : null}
        actionText={error?.action === 'LOGIN' ? 'Sign In Instead' : null}
      />

      {Object.keys(formErrors).length > 0 && !error && (
        <div className="form-summary-error">
          Please fix the highlighted fields before continuing.
        </div>
      )}

      <form onSubmit={handleSubmit} className="signup-form-container" noValidate>

        <div className="signup-form-grid">

          {/* Company Name */}
          <div className="auth-field">
            <label htmlFor="companyName">Company Name *</label>
            <input
              id="companyName"
              name="companyName"
              type="text"
              className={`auth-input ${formErrors.companyName ? 'input-error' : ''}`}
              placeholder="e.g. ABC Freight Pvt. Ltd."
              value={formData.companyName}
              onChange={handleChange}
            />
            {formErrors.companyName && <span className="field-error-text">{formErrors.companyName}</span>}
          </div>

          {/* Full Name */}
          <div className="auth-field">
            <label htmlFor="fullName">Your Full Name *</label>
            <input
              id="fullName"
              name="fullName"
              type="text"
              className={`auth-input ${formErrors.fullName ? 'input-error' : ''}`}
              placeholder="Enter your full name"
              value={formData.fullName}
              onChange={handleChange}
            />
            {formErrors.fullName && <span className="field-error-text">{formErrors.fullName}</span>}
          </div>

          {/* Business Email */}
          <div className="auth-field">
            <label htmlFor="email">Business Email *</label>
            <input
              id="email"
              name="email"
              type="email"
              className={`auth-input ${formErrors.email ? 'input-error' : ''}`}
              placeholder="name@company.com"
              value={formData.email}
              onChange={handleChange}
            />
            {formErrors.email && <span className="field-error-text">{formErrors.email}</span>}
          </div>

          {/* Phone */}
          <div className="auth-field">
            <label htmlFor="phone">Phone Number <span className="optional">(Optional)</span></label>
            <div className={`auth-input phone-picker-wrapper ${formErrors.phone ? 'input-error' : ''}`}>
              <div
                className="custom-country-selector"
                onClick={() => setIsCountryDropdownOpen(!isCountryDropdownOpen)}
              >
                <span className="country-flag">{getFlagEmoji(selectedCountry.iso)}</span>
                <span className="country-dial-code">{selectedCountry.code}</span>
                <svg className="country-dropdown-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="m6 9 6 6 6-6"/></svg>
              </div>

              {isCountryDropdownOpen && (
                <>
                  <div className="custom-country-dropdown-backdrop" onClick={() => setIsCountryDropdownOpen(false)} />
                  <div className="custom-country-dropdown-menu">
                    {COUNTRIES.map(country => (
                      <div
                        key={country.iso}
                        className={`custom-country-option ${selectedCountry.iso === country.iso ? 'selected' : ''}`}
                        onClick={() => { setSelectedCountry(country); setIsCountryDropdownOpen(false); }}
                      >
                        <span className="country-flag">{getFlagEmoji(country.iso)}</span>
                        <span className="country-name">{country.label}</span>
                        <span className="country-code-list">{country.code}</span>
                      </div>
                    ))}
                  </div>
                </>
              )}

              <input
                id="phone"
                name="phone"
                type="tel"
                className="custom-phone-input"
                placeholder="Enter phone number"
                value={formData.phone}
                onChange={handleChange}
                onFocus={() => setIsCountryDropdownOpen(false)}
              />
            </div>
            {formErrors.phone && <span className="field-error-text">{formErrors.phone}</span>}
          </div>

          {/* Password */}
          <div className="auth-field">
            <label htmlFor="password">Password *</label>
            <div className="auth-input-wrapper">
              <input
                id="password"
                name="password"
                type={showPassword ? 'text' : 'password'}
                className={`auth-input ${formErrors.password ? 'input-error' : ''}`}
                placeholder="Create a strong password"
                value={formData.password}
                onChange={handleChange}
              />
              <button type="button" className="auth-input-toggle" onClick={() => setShowPassword(!showPassword)}>
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
            {formErrors.password && <span className="field-error-text">{formErrors.password}</span>}
          </div>

          {/* Confirm Password */}
          <div className="auth-field">
            <label htmlFor="confirmPassword">Confirm Password *</label>
            <div className="auth-input-wrapper">
              <input
                id="confirmPassword"
                name="confirmPassword"
                type={showConfirmPassword ? 'text' : 'password'}
                className={`auth-input ${formErrors.confirmPassword ? 'input-error' : ''}`}
                placeholder="Confirm your password"
                value={formData.confirmPassword}
                onChange={handleChange}
              />
              <button type="button" className="auth-input-toggle" onClick={() => setShowConfirmPassword(!showConfirmPassword)}>
                {showConfirmPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
            {formErrors.confirmPassword && <span className="field-error-text">{formErrors.confirmPassword}</span>}
          </div>

          {/* Password Strength */}
          <div className="signup-password-section">
            <div className="signup-password-meter-row">
              <span className="signup-password-meter-label">Password Strength</span>
              <div className="signup-password-bars">
                {[1, 2, 3, 4].map(level => (
                  <div
                    key={level}
                    className="signup-password-bar"
                    style={{ backgroundColor: strengthScore >= level ? strengthColor : '#E2E8F0' }}
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
        <button type="submit" className="auth-submit-btn" disabled={isLoading}>
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
