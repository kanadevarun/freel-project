import React from 'react';
import { FileText, Clock, CheckCircle2, AlertTriangle, ArrowUpRight } from 'lucide-react';
import './InvoiceKpiCards.css';

export default function InvoiceKpiCards({ kpiData }) {
  const cards = [
    {
      key: 'totalInvoices',
      title: 'TOTAL INVOICES',
      value: kpiData?.totalInvoices?.displayAmount || '$2.48M',
      count: kpiData?.totalInvoices?.label || '128 Invoices',
      trend: kpiData?.totalInvoices?.trend || '18.6%',
      trendPeriod: kpiData?.totalInvoices?.trendPeriod || 'vs last 7 days',
      icon: <FileText size={20} className="kpi-icon-svg total" />,
      colorClass: 'total'
    },
    {
      key: 'outstanding',
      title: 'OUTSTANDING',
      value: kpiData?.outstanding?.displayAmount || '$1.42M',
      count: kpiData?.outstanding?.label || '86 Invoices',
      trend: kpiData?.outstanding?.trend || '12.4%',
      trendPeriod: kpiData?.outstanding?.trendPeriod || 'vs last 7 days',
      icon: <Clock size={20} className="kpi-icon-svg outstanding" />,
      colorClass: 'outstanding'
    },
    {
      key: 'paidThisMonth',
      title: 'PAID (THIS MONTH)',
      value: kpiData?.paidThisMonth?.displayAmount || '$96,420',
      count: kpiData?.paidThisMonth?.label || '32 Invoices',
      trend: kpiData?.paidThisMonth?.trend || '24.8%',
      trendPeriod: kpiData?.paidThisMonth?.trendPeriod || 'vs last 7 days',
      icon: <CheckCircle2 size={20} className="kpi-icon-svg paid" />,
      colorClass: 'paid'
    },
    {
      key: 'overdue',
      title: 'OVERDUE',
      value: kpiData?.overdue?.displayAmount || '$38,750',
      count: kpiData?.overdue?.label || '14 Invoices',
      trend: kpiData?.overdue?.trend || '8.2%',
      trendPeriod: kpiData?.overdue?.trendPeriod || 'vs last 7 days',
      icon: <AlertTriangle size={20} className="kpi-icon-svg overdue" />,
      colorClass: 'overdue'
    }
  ];

  return (
    <div className="invoice-kpi-grid">
      {cards.map((card) => (
        <div key={card.key} className={`invoice-kpi-card ${card.colorClass}`}>
          <div className="kpi-top-row">
            <div className="kpi-text-meta">
              <span className="kpi-label">{card.title}</span>
              <h3 className="kpi-value">{card.value}</h3>
            </div>
            <div className={`kpi-icon-wrapper ${card.colorClass}`}>
              {card.icon}
            </div>
          </div>
          
          <div className="kpi-bottom-row">
            <span className="kpi-subtext-count">{card.count}</span>
            <div className={`kpi-trend-pill ${card.colorClass}`}>
              <ArrowUpRight size={13} className="trend-arrow" />
              <span>{card.trend}</span>
              <span className="trend-period">{card.trendPeriod}</span>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
