import React, { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useAuth } from '../../../context/AuthContext';
import { dashboardService } from '../../../services/dashboardService';
import NewFFDashboard from './NewFFDashboard';
import OperationalDashboard from './OperationalDashboard';
import './DashboardHome.css';

export default function DashboardHome() {
  const { user, isBooting, isAuthenticated } = useAuth();
  const [searchParams] = useSearchParams();
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [authError, setAuthError] = useState(false);
  const [fetchError, setFetchError] = useState(null);

  const presetParam = searchParams.get('preset') || 'LAST_7D';
  const startDateParam = searchParams.get('startDate') || '';
  const endDateParam = searchParams.get('endDate') || '';

  useEffect(() => {
    if (isBooting) return;

    let isMounted = true;

    const fetchMissionControl = async () => {
      try {
        setLoading(true);
        setAuthError(false);
        setFetchError(null);

        const response = await dashboardService.getMissionControl({
          preset: presetParam,
          startDate: startDateParam,
          endDate: endDateParam,
        });
        if (isMounted) {
          // Normalize response envelope
          const missionData = response?.data || response || {};
          setData(missionData);
        }
      } catch (err) {
        if (isMounted) {
          console.error('Failed to fetch mission control data', err);
          if (err?.status === 401 || err?.code === 'UNAUTHORIZED') {
            setAuthError(true);
          } else {
            const errorMsg = typeof err === 'string' ? err : err?.message || 'Unable to connect to service';
            setFetchError(typeof errorMsg === 'string' ? errorMsg : JSON.stringify(errorMsg));
          }
        }
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    };

    fetchMissionControl();

    return () => {
      isMounted = false;
    };
  }, [isBooting, presetParam, startDateParam, endDateParam]);

  // 1. Loading State
  if (loading) {
    return (
      <div className="dashboard-loading-state">
        <div className="dashboard-spinner"></div>
        <p className="loading-text">Loading freight workspace...</p>
      </div>
    );
  }

  // 2. Authentication Error (401)
  if (authError) {
    return (
      <div className="dashboard-notice-card auth-error">
        <div className="notice-icon">🔒</div>
        <h3>Session Expired or Unauthorized</h3>
        <p>Please log in again to access your freight workspace.</p>
        <button
          className="btn-notice-action"
          onClick={() => {
            window.location.href = '/login';
          }}
        >
          Go to Login →
        </button>
      </div>
    );
  }

  // 3. Backend Error (500)
  if (fetchError && !data) {
    return (
      <div className="dashboard-notice-card server-error">
        <div className="notice-icon">⚠️</div>
        <h3>Unable to load workspace data</h3>
        <p>{String(fetchError)}</p>
        <button
          className="btn-notice-action secondary"
          onClick={() => window.location.reload()}
        >
          Retry Connection
        </button>
      </div>
    );
  }

  // 4. Determine Progressive Dashboard State driven by real backend operational classification
  const stats = data?.stats || {};
  const isNewUser =
    stats.is_new_user ??
    (stats.account_maturity === 'NEW' ||
      stats.account_maturity === 'LOW_DATA' ||
      ((!stats.total_customers || stats.total_customers <= 1) &&
        (!stats.open_rfqs || stats.open_rfqs === 0) &&
        (!stats.active_shipments || stats.active_shipments === 0) &&
        (!stats.total_invoices || stats.total_invoices === 0) &&
        (!stats.open_leads || stats.open_leads === 0)));

  if (!isNewUser) {
    return <OperationalDashboard data={data} user={user} />;
  }

  // Default: New Freight Forwarder Onboarding / Low-Data Dashboard
  return <NewFFDashboard data={data} user={user} />;
}
