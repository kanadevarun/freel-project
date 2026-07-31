import React, { useState, useEffect } from 'react';
import { useAuth } from '../../../context/AuthContext';
import { dashboardService } from '../../../services/dashboardService';
import toast from 'react-hot-toast';

import MissionControlLayout from '../../../components/dashboard/MissionControl/MissionControlLayout';
import StatCardWidget from '../../../components/dashboard/MissionControl/StatCardWidget';
import ApprovalQueueWidget from '../../../components/dashboard/MissionControl/ApprovalQueueWidget';
import GlobalTimelineWidget from '../../../components/dashboard/MissionControl/GlobalTimelineWidget';
import AIWorkforceWidget from '../../../components/dashboard/MissionControl/AIWorkforceWidget';

import './DashboardHome.css';

export default function DashboardHome() {
  const { user } = useAuth();
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchMissionControl = async () => {
      try {
        const response = await dashboardService.getMissionControl();
        setData(response.data);
      } catch (err) {
        console.error("Failed to fetch mission control data", err);
        toast.error("Failed to load Mission Control");
      } finally {
        setLoading(false);
      }
    };
    fetchMissionControl();
  }, []);

  return (
    <div className="dashboard-home">
      <div className="dashboard-welcome">
        <h1>Welcome back, {user?.full_name || user?.email?.split('@')[0]}</h1>
        <p>Mission Control: Here is what requires your attention today.</p>
      </div>

      {loading ? (
        <div className="dashboard-loading">
          <div className="spinner"></div>
          <p>Loading Mission Control...</p>
        </div>
      ) : data ? (
        <MissionControlLayout>
          <div className="mc-stats-strip">
            <StatCardWidget 
              title="Total Revenue" 
              value={data.stats?.total_revenue || 0} 
              prefix="$" 
              trend={5.2} 
            />
            <StatCardWidget 
              title="Open RFQs" 
              value={data.stats?.open_rfqs || 0} 
            />
            <StatCardWidget 
              title="Open Leads" 
              value={data.stats?.open_leads || 0} 
            />
            <StatCardWidget 
              title="Active Shipments" 
              value={data.stats?.active_shipments || 0} 
            />
          </div>

          <div className="mc-main-queue">
            <ApprovalQueueWidget queue={data.approval_queue} />
          </div>

          <div className="mc-side-timeline">
            <GlobalTimelineWidget timeline={[]} />
          </div>

          <div className="mc-ai-strip">
            <AIWorkforceWidget status={data.ai_status} />
          </div>
        </MissionControlLayout>
      ) : (
        <div className="dashboard-error">
          <p>Failed to load dashboard data. Please refresh.</p>
        </div>
      )}
    </div>
  );
}
