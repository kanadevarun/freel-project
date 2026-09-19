import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Shield, Lock, Mail, Eye, EyeOff, AlertCircle, ArrowRight, CheckCircle2, Building2 } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { LoadingSpinner } from '../../components/common/LoadingSpinner';

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, isAuthenticated, isAuthenticating, authError } = useAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [formError, setFormError] = useState('');
  const [isCustomerBlocked, setIsCustomerBlocked] = useState(false);

  // If already authenticated, redirect to intended page or dashboard
  useEffect(() => {
    if (isAuthenticated) {
      const from = location.state?.from?.pathname || '/';
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, navigate, location]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setFormError('');
    setIsCustomerBlocked(false);

    // Client-side validations
    const cleanEmail = email.trim();
    if (!cleanEmail) {
      setFormError('Email address is required.');
      return;
    }
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(cleanEmail)) {
      setFormError('Please enter a valid email address.');
      return;
    }
    if (!password) {
      setFormError('Password is required.');
      return;
    }

    const result = await login(cleanEmail, password);
    if (!result.success) {
      if (result.error && (result.error.toLowerCase().includes('customer organization') || result.code === 'ACCESS_DENIED')) {
        setIsCustomerBlocked(true);
      }
    }
  };

  const fillQuickCredentials = (demoEmail) => {
    setEmail(demoEmail);
    setPassword('Password123!');
    setFormError('');
    setIsCustomerBlocked(false);
  };

  return (
    <div className="flex min-h-screen bg-slate-50">
      {/* Left side: Hero & Architecture Banner */}
      <div className="hidden lg:flex lg:w-1/2 flex-col justify-between bg-[#0B192C] p-12 text-white relative overflow-hidden">
        {/* Ambient background decoration */}
        <div className="absolute -right-24 -top-24 h-96 w-96 rounded-full bg-blue-600/10 blur-3xl pointer-events-none" />
        <div className="absolute -left-24 -bottom-24 h-96 w-96 rounded-full bg-indigo-600/10 blur-3xl pointer-events-none" />

        {/* Top Logo */}
        <div className="relative z-10 flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-600 text-white font-black text-xl shadow-lg shadow-blue-500/20">
            L
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-bold text-lg tracking-tight">LogisticsHQ</span>
              <span className="rounded bg-blue-500/20 px-1.5 py-0.5 text-[10px] font-semibold text-blue-300 uppercase tracking-wider">
                Internal
              </span>
            </div>
            <p className="text-xs text-slate-400">SPortal SaaS Administration Center</p>
          </div>
        </div>

        {/* Middle Content */}
        <div className="relative z-10 my-auto max-w-lg space-y-6">
          <div className="inline-flex items-center gap-2 rounded-full border border-blue-500/30 bg-blue-500/10 px-3.5 py-1.5 text-xs font-medium text-blue-300">
            <Shield className="h-3.5 w-3.5 text-blue-400" />
            Zero-Trust SaaS Control Plane
          </div>
          <h1 className="text-3xl font-extrabold tracking-tight text-white sm:text-4xl leading-tight">
            Centralized Platform Operations & Multi-Tenant Administration
          </h1>
          <p className="text-sm leading-relaxed text-slate-300">
            Authoritative internal workspace for LogisticsHQ staff. Direct management of subscribing freight-forwarding organizations, subscription billing, customer success telemetry, and global AI governance.
          </p>

          <div className="grid grid-cols-2 gap-3 pt-4 text-xs">
            <div className="flex items-center gap-2 rounded-lg border border-white/10 bg-white/5 p-3">
              <CheckCircle2 className="h-4 w-4 text-blue-400 shrink-0" />
              <span className="text-slate-200">Strict Tenant Separation</span>
            </div>
            <div className="flex items-center gap-2 rounded-lg border border-white/10 bg-white/5 p-3">
              <CheckCircle2 className="h-4 w-4 text-blue-400 shrink-0" />
              <span className="text-slate-200">Shared MariaDB Truth</span>
            </div>
            <div className="flex items-center gap-2 rounded-lg border border-white/10 bg-white/5 p-3">
              <CheckCircle2 className="h-4 w-4 text-blue-400 shrink-0" />
              <span className="text-slate-200">Granular Internal RBAC</span>
            </div>
            <div className="flex items-center gap-2 rounded-lg border border-white/10 bg-white/5 p-3">
              <CheckCircle2 className="h-4 w-4 text-blue-400 shrink-0" />
              <span className="text-slate-200">Immutable Audit Ledger</span>
            </div>
          </div>
        </div>

        {/* Bottom Notice */}
        <div className="relative z-10 border-t border-white/10 pt-4 text-xs text-slate-400">
          <p className="font-semibold text-slate-300">Confidential LogisticsHQ Internal System</p>
          <p className="mt-0.5">Unauthorized access or customer account traversal is strictly monitored and logged.</p>
        </div>
      </div>

      {/* Right side: Clean, Light SPortal Login Form */}
      <div className="flex w-full lg:w-1/2 items-center justify-center p-6 sm:p-12">
        <div className="w-full max-w-md space-y-8">
          {/* Header Mobile / Title */}
          <div>
            <div className="lg:hidden flex items-center gap-2 mb-6">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-blue-600 text-white font-black text-lg">
                L
              </div>
              <span className="font-bold text-base text-slate-900">LogisticsHQ SPortal</span>
            </div>

            <div className="flex items-center gap-2 text-xs font-semibold text-blue-700 bg-blue-50 px-3 py-1 rounded-full w-fit mb-3">
              <Building2 className="h-3.5 w-3.5" />
              Internal Staff Sign In
            </div>
            <h2 className="text-2xl font-bold tracking-tight text-slate-900">Sign in to SPortal</h2>
            <p className="mt-1.5 text-sm text-slate-500">
              Enter your LogisticsHQ internal credentials to access the SaaS administration portal.
            </p>
          </div>

          {/* Alert Banners */}
          {isCustomerBlocked && (
            <div className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800 shadow-sm animate-fade-in">
              <div className="flex items-start gap-3">
                <AlertCircle className="h-5 w-5 text-rose-600 shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold text-rose-900">Customer Account Detected</h4>
                  <p className="mt-1 text-xs text-rose-700 leading-relaxed">
                    This account belongs to a customer organization and is not authorized to access internal SPortal administration. Please use{' '}
                    <a href="http://localhost:5173" className="underline font-semibold hover:text-rose-900">
                      CPortal (app.logisticshq.in)
                    </a>{' '}
                    for customer operations.
                  </p>
                </div>
              </div>
            </div>
          )}

          {!isCustomerBlocked && (authError || formError) && (
            <div className="rounded-xl border border-rose-200 bg-rose-50 p-3.5 text-sm text-rose-700 flex items-center gap-2.5 shadow-sm animate-fade-in">
              <AlertCircle className="h-4 w-4 text-rose-600 shrink-0" />
              <span>{formError || authError}</span>
            </div>
          )}

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-xs font-semibold text-slate-700 mb-1.5" htmlFor="sportal-email">
                Internal Email Address
              </label>
              <div className="relative">
                <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5 text-slate-400">
                  <Mail className="h-4 w-4" />
                </div>
                <input
                  id="sportal-email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="name@logisticshq.in or internal user"
                  autoComplete="username"
                  className="w-full rounded-xl border border-slate-200 bg-white py-2.5 pl-10 pr-3.5 text-sm text-slate-900 placeholder:text-slate-400 focus:border-blue-600 focus:outline-none focus:ring-4 focus:ring-blue-600/10 transition-all"
                  disabled={isAuthenticating}
                  required
                />
              </div>
            </div>

            <div>
              <div className="flex items-center justify-between mb-1.5">
                <label className="block text-xs font-semibold text-slate-700" htmlFor="sportal-password">
                  Password
                </label>
              </div>
              <div className="relative">
                <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5 text-slate-400">
                  <Lock className="h-4 w-4" />
                </div>
                <input
                  id="sportal-password"
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••••••"
                  autoComplete="current-password"
                  className="w-full rounded-xl border border-slate-200 bg-white py-2.5 pl-10 pr-10 text-sm text-slate-900 placeholder:text-slate-400 focus:border-blue-600 focus:outline-none focus:ring-4 focus:ring-blue-600/10 transition-all"
                  disabled={isAuthenticating}
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute inset-y-0 right-0 flex items-center pr-3.5 text-slate-400 hover:text-slate-600 transition-colors"
                  tabIndex={-1}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
            </div>

            <button
              id="sportal-login-submit"
              type="submit"
              disabled={isAuthenticating}
              className="flex w-full items-center justify-center gap-2 rounded-xl bg-[#0B192C] py-2.5 text-sm font-semibold text-white shadow-lg shadow-slate-900/10 hover:bg-slate-800 focus:outline-none focus:ring-4 focus:ring-slate-900/20 active:scale-[0.99] transition-all disabled:opacity-70 disabled:cursor-not-allowed cursor-pointer mt-2"
            >
              {isAuthenticating ? (
                <>
                  <LoadingSpinner size="sm" className="text-white" />
                  <span>Authenticating Identity...</span>
                </>
              ) : (
                <>
                  <span>Sign In to Control Center</span>
                  <ArrowRight className="h-4 w-4" />
                </>
              )}
            </button>
          </form>

          {/* Quick Demo Switcher for Verification & Testing */}
          <div className="rounded-xl border border-slate-200 bg-slate-50/80 p-3.5 text-xs text-slate-600 space-y-2">
            <div className="flex items-center justify-between font-semibold text-slate-700">
              <span>Quick Development Switcher:</span>
              <span className="text-[10px] uppercase font-mono tracking-wider text-slate-400">Dev Mode</span>
            </div>
            <div className="grid grid-cols-2 gap-2">
              <button
                id="quick-btn-ceo"
                type="button"
                onClick={() => fillQuickCredentials('ceo@freel-demo.local')}
                className="rounded-lg border border-slate-200 bg-white p-2 text-left hover:border-blue-500 hover:bg-blue-50/40 transition-all cursor-pointer"
              >
                <div className="font-semibold text-slate-800">Super Admin / CEO</div>
                <div className="text-[10px] text-slate-500 truncate">ceo@freel-demo.local</div>
              </button>
              <button
                id="quick-btn-customer"
                type="button"
                onClick={() => fillQuickCredentials('customer@tata-exports.local')}
                className="rounded-lg border border-slate-200 bg-white p-2 text-left hover:border-rose-500 hover:bg-rose-50/40 transition-all cursor-pointer"
              >
                <div className="font-semibold text-rose-700">Customer User (Block)</div>
                <div className="text-[10px] text-slate-500 truncate">customer@tata-exports.local</div>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
