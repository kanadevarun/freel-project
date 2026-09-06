import React, { useState, useEffect } from 'react';
import { 
  CheckCircle2, 
  CreditCard, 
  CalendarDays,
  Users,
  Mail,
  FileText,
  Truck,
  Link2,
  Database,
  ArrowRight,
  AlertCircle
} from 'lucide-react';
import { subscriptionAPI } from '../../../services/api/subscription';
import ConfirmModal from './ConfirmModal';
import InvoiceHistoryModal from './InvoiceHistoryModal';
import { Link } from 'react-router-dom';
import './SubscriptionPage.css';
import { toast } from 'react-hot-toast';

export default function SubscriptionPage() {
  const [workspace, setWorkspace] = useState(null);
  const [plans, setPlans] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  
  const [billingCycle, setBillingCycle] = useState('monthly'); // 'monthly' | 'annual'
  
  // Modal states
  const [isConfirmModalOpen, setIsConfirmModalOpen] = useState(false);
  const [selectedPlan, setSelectedPlan] = useState(null);
  const [isChangingPlan, setIsChangingPlan] = useState(false);
  const [planPreview, setPlanPreview] = useState(null);
  const [addonConfigs, setAddonConfigs] = useState([]);
  const [isUpdatingAddon, setIsUpdatingAddon] = useState(false);
  const [isInvoiceModalOpen, setIsInvoiceModalOpen] = useState(false);

  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get('success')) {
      toast.success('Subscription completed successfully!');
      // Clean up URL
      window.history.replaceState({}, document.title, window.location.pathname);
    } else if (urlParams.get('canceled')) {
      toast.error('Checkout was canceled.');
      window.history.replaceState({}, document.title, window.location.pathname);
    }
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setIsLoading(true);
      const [workspaceData, plansData, addonsData] = await Promise.all([
        subscriptionAPI.getWorkspace(),
        subscriptionAPI.getPlans(),
        subscriptionAPI.getAddonConfigs()
      ]);
      // Handle cases where plansData is wrapped in { data: [...] }
      const plansArray = Array.isArray(plansData) ? plansData : (plansData?.data || []);
      const addonsArray = Array.isArray(addonsData) ? addonsData : (addonsData?.data || []);
      setWorkspace(workspaceData);
      setPlans(plansArray);
      setAddonConfigs(addonsArray);
      
      // Init billing cycle from current sub if exists
      if (workspaceData?.subscription?.billing_cycle) {
        setBillingCycle(workspaceData.subscription.billing_cycle);
      }
    } catch (err) {
      console.error(err);
      toast.error('Failed to load subscription data');
    } finally {
      setIsLoading(false);
    }
  };

  const handlePlanSelect = async (plan) => {
    // If it's already the current plan and cycle matches, do nothing
    if (
      workspace?.subscription?.plan_id === plan.id && 
      workspace?.subscription?.billing_cycle === billingCycle
    ) {
      return;
    }
    try {
      setIsLoading(true);
      const preview = await subscriptionAPI.previewPlanChange(plan.id, billingCycle);
      setPlanPreview(preview);
      setSelectedPlan(plan);
      setIsConfirmModalOpen(true);
    } catch (err) {
      toast.error('Failed to preview plan change');
    } finally {
      setIsLoading(false);
    }
  };

  const confirmPlanChange = async () => {
    try {
      setIsChangingPlan(true);
      
      // If no active provider subscription, initiate checkout
      if (!workspace?.subscription?.provider_subscription_id) {
        const data = await subscriptionAPI.checkout(selectedPlan.id, billingCycle);
        if (data.url) {
          window.location.href = data.url;
          return;
        }
      }

      await subscriptionAPI.changePlan(selectedPlan.id, billingCycle);
      toast.success(`Successfully subscribed to ${selectedPlan.name} plan`);
      setIsConfirmModalOpen(false);
      await fetchData(); // Refresh data
    } catch (err) {
      toast.error(err.response?.data?.error || 'Failed to change plan');
    } finally {
      setIsChangingPlan(false);
    }
  };

  if (isLoading) {
    return (
      <div className="subscription-page">
        <div className="settings-page-header">
          <div>
            <h1 className="settings-page-title">Subscription</h1>
            <p className="settings-page-subtitle">Manage your plan, usage, and billing details.</p>
          </div>
        </div>
        <div className="subscription-skeleton-grid">
          <div className="subscription-skeleton-card" style={{ height: '300px' }}></div>
          <div className="subscription-skeleton-card" style={{ height: '300px' }}></div>
          <div className="subscription-skeleton-card" style={{ height: '300px' }}></div>
        </div>
      </div>
    );
  }

  const { subscription, current_plan, customer, payment_method, usage, addons, invoices } = workspace || {};
  
  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const handleReactivate = async () => {
    try {
      setIsLoading(true);
      await subscriptionAPI.reactivateSubscription();
      toast.success('Subscription reactivated successfully!');
      fetchData();
    } catch (err) {
      toast.error('Failed to reactivate subscription');
    } finally {
      setIsLoading(false);
    }
  };

  const handleManageBilling = async () => {
    try {
      setIsLoading(true);
      const res = await subscriptionAPI.createCustomerPortal();
      if (res.url) {
        window.location.href = res.url;
      }
    } catch (err) {
      toast.error('Failed to open billing portal');
    } finally {
      setIsLoading(false);
    }
  };

  const handleUpdateAddon = async (addonConfigId, quantity) => {
    try {
      setIsUpdatingAddon(true);
      await subscriptionAPI.updateAddons(addonConfigId, quantity);
      toast.success('Add-on updated successfully');
      await fetchData();
    } catch (err) {
      toast.error('Failed to update add-on');
    } finally {
      setIsUpdatingAddon(false);
    }
  };

  const renderMetricItem = (metricKey, Icon, label, suffix = '') => {
    const item = usage?.find(u => u.metric_name === metricKey);
    const used = item ? item.current_usage : 0;
    const isUnlimited = item?.unlimited;
    const limit = isUnlimited ? 'Unlimited' : (item?.limit_amount || 0);
    const pct = isUnlimited ? 0 : (item?.percentage || 0);
    
    let color = '#4F46E5'; // Indigo
    let showWarning = false;
    let warningText = '';

    if (!isUnlimited && limit > 0) {
      if (pct >= 100) {
        color = '#EF4444'; // Red
        showWarning = true;
        warningText = `You've reached your ${label} limit. Upgrade your plan to continue.`;
      } else if (pct >= 80) {
        color = '#F59E0B'; // Amber
        showWarning = true;
        warningText = `${pct}% of your ${label} limit used.`;
      }
    }

    return (
      <div className="usage-metric-item" key={metricKey}>
        <div className="usage-metric-header">
          <div className="usage-metric-label">
            <Icon size={16} /> {label}
          </div>
          <div className="usage-metric-values">
            {used}{suffix} / {limit}{suffix && !isUnlimited ? ` ${suffix}` : ''}
          </div>
        </div>
        {!isUnlimited && (
          <div className="usage-progress-bar">
            <div className="usage-progress-fill" style={{ width: `${pct}%`, background: color }}></div>
          </div>
        )}
        {showWarning && (
          <div style={{ fontSize: '0.75rem', color: color, marginTop: '4px', fontWeight: 500 }}>
            {warningText}
          </div>
        )}
      </div>
    );
  };

  // Plan features rendering
  const renderFeatures = (featuresJson) => {
    try {
      // features might be a stringified array if parsed from JSONRawMessage
      let list = Array.isArray(featuresJson) ? featuresJson : JSON.parse(featuresJson);
      return list.map((feat, idx) => (
        <li key={idx} className="plan-feature-item">
          <CheckCircle2 size={16} className="plan-feature-icon" />
          <span>{feat}</span>
        </li>
      ));
    } catch {
      return null;
    }
  };

  return (
    <div className="subscription-page">
      <div className="settings-page-header">
        <div>
          <h1 className="settings-page-title">Subscription & Billing</h1>
          <p className="settings-page-subtitle">Manage your plan, usage, and billing details.</p>
        </div>
        
        {subscription?.status === 'past_due' && (
          <div className="subscription-warning-banner" style={{ background: '#FEF2F2', border: '1px solid #FCA5A5', color: '#991B1B', padding: '12px 16px', borderRadius: '8px', marginBottom: '24px', display: 'flex', alignItems: 'center', gap: '12px' }}>
            <AlertCircle size={20} />
            <span style={{fontWeight: 500}}>Your last payment failed. Please update your payment method to avoid service interruption.</span>
          </div>
        )}

        <div className="subscription-header-actions">
           <div className="subscription-period-badge">
             <CalendarDays size={16} />
             {subscription?.current_period_start ? 
                (subscription?.cancel_at_period_end ? `Cancels on ${formatDate(subscription.current_period_end)}` : `${formatDate(subscription.current_period_start)} - ${formatDate(subscription.current_period_end)}`) : 
                'No active subscription'
             }
           </div>
        </div>
      </div>

      <div className="subscription-top-grid">
        
        {/* CURRENT PLAN */}
        <div className="subscription-card current-plan-card">
          <div className="subscription-card-header">
            <h3>Current Plan</h3>
            <span className={`subscription-status-badge ${subscription?.status === 'active' ? 'active' : 'inactive'}`}>
              {subscription?.status || 'Inactive'}
            </span>
          </div>
          
          <div className="current-plan-hero">
            <div className="plan-icon-wrapper">
              {current_plan?.name === 'Growth' ? <div className="plan-crown">👑</div> : 
               current_plan?.name === 'Professional' ? <div className="plan-star">⭐</div> : 
               <div className="plan-rocket">🚀</div>}
            </div>
            <div className="plan-hero-details">
              <h4>{current_plan?.name || 'No active plan'}</h4>
              <div className="plan-price">
                {current_plan ? (
                  <>
                    <span className="price-val">
                      ${subscription?.billing_cycle === 'annual' ? current_plan?.price_annual / 12 || 0 : current_plan?.price_monthly || 0}
                    </span>
                    <span className="price-period">/ month</span>
                  </>
                ) : (
                  <span className="price-val" style={{fontSize: '1rem', color: '#64748B'}}>Select a plan below to get started.</span>
                )}
              </div>
              <p className="next-billing-text">
                {subscription?.current_period_end ? `Next billing date: ${formatDate(subscription?.current_period_end)}` : ''}
              </p>
            </div>
          </div>

          <div className="plan-features-list-wrapper">
            <p className="plan-features-title">Features included:</p>
            <ul className="plan-features-list">
              {current_plan && renderFeatures(current_plan.features)}
            </ul>
          </div>

          {subscription?.cancel_at_period_end ? (
             <button className="btn-manage-plan" onClick={handleReactivate} disabled={isLoading}>
                {isLoading ? 'Processing...' : 'Reactivate Plan'}
             </button>
          ) : (
             <button className="btn-manage-plan" onClick={() => document.getElementById('available-plans-section').scrollIntoView({behavior: 'smooth'})}>
                Manage Plan
             </button>
          )}
        </div>

        {/* USAGE & LIMITS */}
        <div className="subscription-card usage-card">
          <div className="subscription-card-header">
            <h3>Usage & Limits</h3>
            <span className="renews-text">{subscription?.current_period_end ? `Renews on ${formatDate(subscription.current_period_end)}` : ''}</span>
          </div>
          
          <div className="usage-metrics-list">
            {renderMetricItem('team_members', Users, 'Team Members')}
            {renderMetricItem('ai_email_processing', Mail, 'AI Emails Processed')}
            {renderMetricItem('rfqs', FileText, 'RFQs')}
            {renderMetricItem('shipments', Truck, 'Shipments')}
            {renderMetricItem('carrier_connections', Link2, 'Carrier Connections')}
            {renderMetricItem('storage_gb', Database, 'Storage Used', ' GB')}
          </div>

          <Link to="#" className="view-details-link">View full usage details <ArrowRight size={14} /></Link>
        </div>

        {/* BILLING DETAILS */}
        <div className="subscription-card billing-card">
          <div className="subscription-card-header">
            <h3>Billing Details</h3>
          </div>
          
          <div className="billing-details-list">
            <div className="billing-detail-row">
              <span className="billing-label">Billing Status</span>
              <span className={`billing-status-text ${subscription?.status === 'active' ? 'active' : ''}`}>{subscription?.status === 'active' ? 'Active' : 'Inactive'}</span>
            </div>
            <div className="billing-detail-row">
              <span className="billing-label">Billing Cycle</span>
              <span className="billing-value" style={{textTransform: 'capitalize'}}>{subscription?.billing_cycle || 'Not active'}</span>
            </div>
            <div className="billing-detail-row">
              <span className="billing-label">Current Period</span>
              <span className="billing-value">
                {subscription?.current_period_start ? `${formatDate(subscription.current_period_start)} - ${formatDate(subscription.current_period_end)}` : 'Not active'}
              </span>
            </div>
            <div className="billing-detail-row">
              <span className="billing-label">Payment Method</span>
              <span className="billing-value payment-method-val">
                {payment_method ? (
                  <>
                    <CreditCard size={14} style={{marginRight: '6px', color: '#64748B'}} />
                    <span style={{textTransform: 'uppercase', marginRight: '6px', fontWeight: 600}}>{payment_method.card_brand}</span>
                    •••• {payment_method.card_last4}
                    <span style={{color: '#94A3B8', marginLeft: '8px', fontSize: '0.8rem'}}>Expires {payment_method.exp_month}/{payment_method.exp_year}</span>
                  </>
                ) : (
                  <span style={{color: '#64748B'}}>Not available</span>
                )}
              </span>
            </div>
            <div className="billing-detail-row" style={{borderBottom: 'none'}}>
              <span className="billing-label">Amount</span>
              <span className="billing-value" style={{fontWeight: 700}}>
                {current_plan ? `$${subscription?.billing_cycle === 'annual' ? current_plan?.price_annual : current_plan?.price_monthly}` : '-'}
              </span>
            </div>
          </div>

          <button className="btn-manage-billing" onClick={handleManageBilling} disabled={isLoading}>
             {subscription?.status === 'past_due' ? 'Retry Payment' : 'Manage Billing & Payment'}
          </button>
          <p className="billing-helper-text">Update payment method, view invoices and billing history.</p>
        </div>
      </div>

      <div className="subscription-bottom-grid">
        {/* LEFT COLUMN: AVAILABLE PLANS */}
        <div id="available-plans-section" className="available-plans-section">
          <div className="plans-section-header">
            <h2>Available Plans</h2>
            
            <div className="billing-toggle">
              <button 
                className={`toggle-btn ${billingCycle === 'monthly' ? 'active' : ''}`}
                onClick={() => setBillingCycle('monthly')}
              >
                Monthly
              </button>
              <button 
                className={`toggle-btn ${billingCycle === 'annual' ? 'active' : ''}`}
                onClick={() => setBillingCycle('annual')}
              >
                Annual <span className="save-badge">Save 20%</span>
              </button>
            </div>
          </div>

          <div className="plans-grid">
            {Array.isArray(plans) && plans.map(plan => {
              const isCurrent = subscription?.plan_id === plan.id && subscription?.billing_cycle === billingCycle;
              const price = billingCycle === 'annual' ? (plan.price_annual / 12).toFixed(0) : plan.price_monthly;
              
              return (
                <div key={plan.id} className={`plan-card ${isCurrent ? 'current-plan' : ''} ${plan.name === 'Growth' ? 'popular-plan' : ''}`}>
                  {plan.name === 'Growth' && <div className="popular-badge">Most Popular</div>}
                  
                  <div className="plan-card-header">
                    <div className="plan-icon">
                      {plan.name === 'Growth' ? '👑' : plan.name === 'Professional' ? '⭐' : '🚀'}
                    </div>
                    <h3>{plan.name}</h3>
                  </div>
                  
                  <div className="plan-card-price">
                    <span className="price-currency">$</span>
                    <span className="price-amount">{price}</span>
                    <span className="price-period">/ month</span>
                  </div>
                  
                  <p className="plan-card-desc">{plan.description}</p>
                  
                  <ul className="plan-card-features">
                    {renderFeatures(plan.features)}
                  </ul>
                  
                  <button 
                    className={`plan-card-btn ${isCurrent ? 'btn-current' : 'btn-select'}`}
                    onClick={() => handlePlanSelect(plan)}
                    disabled={isCurrent}
                  >
                    {isCurrent ? '✓ Current Plan' : 'Choose Plan'}
                  </button>
                </div>
              );
            })}
            
            {/* Custom Enterprise Card (Static UI) */}
            <div className="plan-card enterprise-plan">
               <div className="plan-card-header">
                  <div className="plan-icon">🏢</div>
                  <h3>Enterprise</h3>
                </div>
                <div className="plan-card-price" style={{alignItems: 'center', minHeight: '44px'}}>
                    <span className="price-amount" style={{fontSize: '1.5rem'}}>Custom</span>
                </div>
                <p className="plan-card-desc">Tailored for enterprise operations.</p>
                <ul className="plan-card-features">
                    <li className="plan-feature-item"><CheckCircle2 size={16} className="plan-feature-icon" /><span>Custom limits</span></li>
                    <li className="plan-feature-item"><CheckCircle2 size={16} className="plan-feature-icon" /><span>Dedicated account manager</span></li>
                    <li className="plan-feature-item"><CheckCircle2 size={16} className="plan-feature-icon" /><span>Advanced integrations</span></li>
                    <li className="plan-feature-item"><CheckCircle2 size={16} className="plan-feature-icon" /><span>SLA & uptime guarantee</span></li>
                    <li className="plan-feature-item"><CheckCircle2 size={16} className="plan-feature-icon" /><span>Onboarding & training</span></li>
                </ul>
                <button className="plan-card-btn btn-outline">Contact Sales</button>
            </div>
          </div>
        </div>

        {/* RIGHT COLUMN: ADDONS & INVOICES */}
        <div className="subscription-right-column">
          {/* ADD-ONS */}
          <div className="subscription-card addons-card">
            <div className="subscription-card-header">
              <h3>Add-ons</h3>
            </div>
            
            <div className="addons-list">
               {addonConfigs.map(config => {
                 const currentAddon = addons?.find(a => a.addon_config_id === config.id);
                 const quantity = currentAddon ? currentAddon.quantity : 0;
                 
                 return (
                   <div className="addon-row" key={config.id}>
                     <span className="addon-name">
                        {config.name}
                        <br/>
                        <span style={{fontSize: '0.75rem', color: '#64748B'}}>{config.description}</span>
                     </span>
                     <span className="addon-price">${config.unit_price} {config.pricing_model === 'per_unit' ? '/ unit' : '/ mo'}</span>
                     
                     {config.pricing_model === 'per_unit' ? (
                       <select 
                         className="addon-select" 
                         value={quantity}
                         disabled={isUpdatingAddon}
                         onChange={(e) => handleUpdateAddon(config.id, parseInt(e.target.value))}
                       >
                         {[0, 1, 2, 3, 4, 5, 10].map(val => (
                           <option key={val} value={val}>{val}</option>
                         ))}
                       </select>
                     ) : (
                       <div className="toggle-switch">
                         <input 
                           type="checkbox" 
                           id={`addon-${config.id}`} 
                           checked={quantity > 0}
                           disabled={isUpdatingAddon}
                           onChange={(e) => handleUpdateAddon(config.id, e.target.checked ? 1 : 0)}
                         />
                         <label htmlFor={`addon-${config.id}`}></label>
                       </div>
                     )}
                   </div>
                 );
               })}
               {addonConfigs.length === 0 && <div className="addon-row" style={{color: '#64748B'}}>No add-ons available.</div>}
            </div>
            <Link to="#" className="view-details-link" style={{marginTop: '16px', display: 'inline-block'}}>View all add-ons <ArrowRight size={14} /></Link>
          </div>

          {/* INVOICES */}
          <div className="subscription-card invoices-card">
            <div className="subscription-card-header" style={{borderBottom: 'none', paddingBottom: '0'}}>
              <h3>Invoices</h3>
              <button 
                className="view-details-link" 
                style={{paddingTop: 0, background: 'none', border: 'none', cursor: 'pointer', fontSize: '0.875rem'}}
                onClick={() => setIsInvoiceModalOpen(true)}
              >
                View all invoices <ArrowRight size={14} />
              </button>
            </div>
            
            <table className="invoices-table">
              <thead>
                <tr>
                  <th>DATE</th>
                  <th>INVOICE #</th>
                  <th>AMOUNT</th>
                  <th>STATUS</th>
                </tr>
              </thead>
              <tbody>
                {invoices && invoices.length > 0 ? (
                  invoices.slice(0, 3).map(inv => (
                    <tr key={inv.id}>
                      <td>{formatDate(inv.issued_at)}</td>
                      <td>{inv.number}</td>
                      <td>${inv.amount_due?.toFixed(2)}</td>
                      <td><span className={`invoice-status ${inv.status.toLowerCase()}`}>{inv.status}</span></td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan="4" className="empty-state-cell" style={{textAlign: 'center', padding: '24px 0', color: '#64748B'}}>No recent invoices</td>
                  </tr>
                )}
              </tbody>
            </table>
            <p className="billing-helper-text" style={{marginTop: '16px', display: 'flex', justifyContent: 'space-between', width: '100%'}}>
              <span>All amounts are in USD</span>
              <span>Need help? <Link to="#">Contact Support</Link></span>
            </p>
          </div>
        </div>
      </div>

      <div className="enterprise-contact-banner">
        <div className="banner-icon">i</div>
        <div className="banner-content">
          <h4>Need a custom plan for your enterprise?</h4>
          <p>Contact our sales team to discuss your requirements.</p>
        </div>
        <button className="btn-outline" style={{background: '#fff'}}>Contact Sales</button>
      </div>

      <ConfirmModal
        isOpen={isConfirmModalOpen}
        title="Change Subscription Plan"
        message={`Are you sure you want to change your plan? Your new billing rate will be applied immediately.`}
        confirmText="Confirm Change"
        confirmStyle="primary"
        isLoading={isChangingPlan}
        previewData={planPreview}
        onConfirm={confirmPlanChange}
        onCancel={() => setIsConfirmModalOpen(false)}
      />

      <InvoiceHistoryModal 
        isOpen={isInvoiceModalOpen}
        onClose={() => setIsInvoiceModalOpen(false)}
        invoices={invoices}
      />
    </div>
  );
}
