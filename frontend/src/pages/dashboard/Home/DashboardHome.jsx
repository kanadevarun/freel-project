import React, { useState, useEffect } from 'react';
import { useAuth } from '../../../context/AuthContext';
import { dashboardService } from '../../../services/dashboardService';
import NewFFDashboard from './NewFFDashboard';
import OperationalDashboard from './OperationalDashboard';
import './DashboardHome.css';

export default function DashboardHome() {
  const { user, isBooting, isAuthenticated } = useAuth();
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [authError, setAuthError] = useState(false);
  const [fetchError, setFetchError] = useState(null);

  useEffect(() => {
    if (isBooting) return;

    let isMounted = true;

    const fetchMissionControl = async () => {
      try {
        setLoading(true);
        setAuthError(false);
        setFetchError(null);

        const response = await dashboardService.getMissionControl();
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
  }, []);

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

  // 4. Determine Progressive Dashboard State
  // An organization is considered operational if it has any active RFQs, shipments, revenue, or pending quotes in queue.
  const stats = data?.stats || {};
  const approvalQueue = data?.approval_queue || [];
  const isOperational =
    (stats.open_rfqs && stats.open_rfqs > 0) ||
    (stats.active_shipments && stats.active_shipments > 0) ||
    (stats.total_revenue && stats.total_revenue > 0) ||
    (stats.open_leads && stats.open_leads > 0) ||
    approvalQueue.length > 0;

  if (isOperational) {
    return <OperationalDashboard data={data} />;
  }

  // Default: New Freight Forwarder Onboarding / Empty-State Dashboard
  return <NewFFDashboard />;
}
