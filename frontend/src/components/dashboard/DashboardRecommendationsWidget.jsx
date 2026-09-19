import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Compass, ExternalLink, ChevronRight, AlertTriangle, Clock } from 'lucide-react';
import { recommendationService } from '../../services/recommendationService';

export default function DashboardRecommendationsWidget() {
  const [stats, setStats] = useState(null);
  const [topItems, setTopItems] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let isMounted = true;
    Promise.all([
      recommendationService.getStats().catch(() => null),
      recommendationService.listRecommendations({ limit: 3, status: 'new', sort_by: 'priority', sort_dir: 'desc' }).catch(() => null),
    ]).then(([statsRes, listRes]) => {
      if (!isMounted) return;
      if (statsRes) {
        const s = statsRes?.data || statsRes?.stats || statsRes;
        if (s) setStats(s);
      }
      if (listRes) {
        const items = listRes?.data?.items || listRes?.items || listRes?.recommendations || [];
        setTopItems(items);
      }
    }).finally(() => {
      if (isMounted) setLoading(false);
    });

    return () => {
      isMounted = false;
    };
  }, []);

  if (loading || (!stats?.total_active && !stats?.total && topItems.length === 0)) {
    return null;
  }

  return (
    <div style={{
      background: '#ffffff',
      border: '1px solid #e2e8f0',
      borderRadius: '12px',
      padding: '20px',
      marginBottom: '24px',
      boxShadow: '0 1px 3px rgba(15, 23, 42, 0.04)',
    }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '16px',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <Compass size={18} color="#2563eb" />
          <h3 style={{ margin: 0, fontSize: '1rem', fontWeight: 700, color: '#0f172a' }}>
            Urgent AI Recommendations ({stats?.total_active ?? topItems.length})
          </h3>
        </div>
        <Link
          to="/dashboard/recommendations"
          style={{
            fontSize: '0.8125rem',
            fontWeight: 600,
            color: '#2563eb',
            textDecoration: 'none',
            display: 'flex',
            alignItems: 'center',
            gap: '4px',
          }}
        >
          <span>Open Recommendation Center</span>
          <ExternalLink size={13} />
        </Link>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '12px' }}>
        {topItems.map((rec) => (
          <div
            key={rec.id}
            style={{
              background: '#f8fafc',
              border: '1px solid #e2e8f0',
              borderLeft: rec.priority === 'critical' ? '3px solid #dc2626' : rec.priority === 'high' ? '3px solid #d97706' : '3px solid #2563eb',
              borderRadius: '8px',
              padding: '12px 14px',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'space-between',
            }}
          >
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '6px' }}>
                <span style={{
                  fontSize: '0.6875rem',
                  fontWeight: 700,
                  textTransform: 'uppercase',
                  padding: '2px 6px',
                  borderRadius: '4px',
                  background: rec.priority === 'critical' ? '#fef2f2' : rec.priority === 'high' ? '#fffbeb' : '#eff6ff',
                  color: rec.priority === 'critical' ? '#991b1b' : rec.priority === 'high' ? '#92400e' : '#1e40af',
                }}>
                  {rec.priority}
                </span>
                <span style={{ fontSize: '0.75rem', color: '#64748b' }}>{rec.source_reference}</span>
              </div>
              <div style={{ fontSize: '0.875rem', fontWeight: 600, color: '#0f172a', marginBottom: '4px' }}>
                {rec.title}
              </div>
              <div style={{ fontSize: '0.8125rem', color: '#475569', lineHeight: 1.4 }}>
                {rec.description}
              </div>
            </div>

            <div style={{
              marginTop: '10px',
              paddingTop: '8px',
              borderTop: '1px solid #e2e8f0',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}>
              <span style={{ fontSize: '0.75rem', color: '#64748b' }}>Next: {rec.recommended_action}</span>
              <Link
                to="/dashboard/recommendations"
                style={{
                  color: '#2563eb',
                  fontSize: '0.75rem',
                  fontWeight: 600,
                  textDecoration: 'none',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '2px',
                }}
              >
                <span>Review</span>
                <ChevronRight size={13} />
              </Link>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
