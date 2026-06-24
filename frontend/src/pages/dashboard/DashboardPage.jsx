import { useAuth } from '../../context/AuthContext';
import './DashboardPage.css';

export default function DashboardPage() {
  const { user, org, memberRole } = useAuth();

  return (
    <div className="dashboard-page">
      <div className="dashboard-header">
        <h1>Welcome to Freel, {user?.email?.split('@')[0] || 'User'}</h1>
        <p>Your logistics workspace is ready.</p>
      </div>

      <div className="dashboard-info-card">
        <div className="info-row">
          <span className="info-label">Workspace:</span>
          <span className="info-value">{org?.name || 'Your Organization'}</span>
        </div>
        <div className="info-row">
          <span className="info-label">Role:</span>
          <span className="info-value">{memberRole || 'OWNER'}</span>
        </div>
      </div>

      <div className="dashboard-modules-section">
        <h3>Coming Soon:</h3>
        <div className="modules-grid">
          <div className="module-card">📦 Shipments</div>
          <div className="module-card">📋 RFQs</div>
          <div className="module-card">💰 Freight Rates</div>
          <div className="module-card">📍 Tracking</div>
          <div className="module-card">📄 Documents</div>
          <div className="module-card">📊 Trade Intelligence</div>
        </div>
      </div>
    </div>
  );
}
