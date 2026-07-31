import React, { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import api from '../../../services/api';
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, LineChart, Line
} from 'recharts';
import './ReportsPage.css';

// Mock charts data for MVP
const funnelData = [
  { name: 'Leads', value: 100 },
  { name: 'RFQs Created', value: 75 },
  { name: 'Quotes Sent', value: 50 },
  { name: 'Shipments Won', value: 20 },
];

const revenueData = [
  { month: 'Jan', revenue: 40000, margin: 8000 },
  { month: 'Feb', revenue: 30000, margin: 6000 },
  { month: 'Mar', revenue: 50000, margin: 10000 },
  { month: 'Apr', revenue: 45000, margin: 9000 },
  { month: 'May', revenue: 60000, margin: 12000 },
  { month: 'Jun', revenue: 75000, margin: 15000 },
];

export default function ReportsPage() {
  const [metrics, setMetrics] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        const response = await api.get('/api/v1/reports/metrics');
        setMetrics(response.data);
      } catch (err) {
        console.error("Failed to load metrics", err);
        toast.error("Failed to load Reports");
      } finally {
        setLoading(false);
      }
    };
    fetchMetrics();
  }, []);

  return (
    <div className="reports-page">
      <div className="reports-header">
        <h1>Business Analytics</h1>
        <p>Monitor your company's performance, conversion rates, and revenue.</p>
      </div>

      {loading ? (
        <div className="reports-loading">
          <div className="spinner"></div>
          <p>Loading analytics...</p>
        </div>
      ) : metrics ? (
        <div className="reports-content">
          <div className="metrics-strip">
            <div className="metric-card">
              <span className="metric-label">Lead to RFQ</span>
              <span className="metric-value">{metrics.lead_conversion}%</span>
            </div>
            <div className="metric-card">
              <span className="metric-label">RFQ to Won</span>
              <span className="metric-value">{metrics.rfq_conversion}%</span>
            </div>
            <div className="metric-card">
              <span className="metric-label">Overall Win Rate</span>
              <span className="metric-value">{metrics.win_rate}%</span>
            </div>
            <div className="metric-card">
              <span className="metric-label">Total Revenue</span>
              <span className="metric-value text-green">${metrics.revenue.toLocaleString()}</span>
            </div>
          </div>

          <div className="charts-grid">
            <div className="chart-card">
              <h3>Conversion Funnel</h3>
              <div className="chart-container">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={funnelData} layout="vertical" margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                    <XAxis type="number" stroke="#94a3b8" />
                    <YAxis dataKey="name" type="category" width={100} stroke="#94a3b8" />
                    <Tooltip cursor={{fill: 'rgba(255,255,255,0.05)'}} contentStyle={{backgroundColor: '#1e293b', border: 'none', color: '#fff'}} />
                    <Bar dataKey="value" fill="#3b82f6" radius={[0, 4, 4, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>

            <div className="chart-card">
              <h3>Revenue Trend (YTD)</h3>
              <div className="chart-container">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={revenueData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                    <XAxis dataKey="month" stroke="#94a3b8" />
                    <YAxis stroke="#94a3b8" />
                    <Tooltip contentStyle={{backgroundColor: '#1e293b', border: 'none', color: '#fff'}} />
                    <Legend />
                    <Line type="monotone" dataKey="revenue" stroke="#10b981" activeDot={{ r: 8 }} strokeWidth={2} />
                    <Line type="monotone" dataKey="margin" stroke="#8b5cf6" strokeWidth={2} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <div className="reports-error">
          <p>Failed to load reports. Please refresh.</p>
        </div>
      )}
    </div>
  );
}
