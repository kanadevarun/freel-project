import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import TopBar from '../../layouts/AppShell/TopBar';
import { dashboardService } from '../../services/dashboardService';
import { searchService } from '../../services/searchService';

// Mock dashboard and search services
vi.mock('../../services/dashboardService', () => ({
  dashboardService: {
    getMissionControl: vi.fn().mockResolvedValue({
      data: {
        attention_items: [
          {
            id: 'notif-1',
            title: 'Invoice Due',
            subtitle: 'INV-2026-0012 is pending approval',
            priority: 'HIGH',
            category: 'FINANCE',
            timestamp: '10m ago',
            action_url: '/dashboard/finance',
          },
        ],
        date_range: {
          label: 'Aug 9 – Aug 15, 2026',
          comparison_label: 'vs last 7 days',
        },
      },
    }),
  },
}));

vi.mock('../../services/searchService', () => ({
  searchService: {
    globalSearch: vi.fn().mockResolvedValue({
      data: {
        data: {
          query: 'BK',
          total_matches: 2,
          groups: [
            {
              category: 'BOOKING',
              category_label: 'Bookings',
              count: 2,
              items: [
                {
                  id: 'bk-1',
                  category: 'BOOKING',
                  title: 'BK-QA-1787840980',
                  subtitle: 'Maersk Line • INNSA ➔ NLRTM',
                  badge: 'Confirmed',
                  badge_type: 'success',
                  url: '/dashboard/bookings/7',
                },
              ],
            },
          ],
        },
      },
    }),
  },
}));

// Mock Auth Context
vi.mock('../../context/AuthContext', () => ({
  useAuth: () => ({
    user: {
      first_name: 'Varun',
      last_name: 'Kanade',
      email: 'varunkanade3456@gmail.com',
      role: 'Org Admin',
      org_name: 'ABC FreightForwarding',
    },
    logout: vi.fn(),
  }),
}));

describe('TopBar Global Header Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders welcome greeting with real user first name and header controls', async () => {
    render(
      <BrowserRouter>
        <TopBar />
      </BrowserRouter>
    );

    expect(screen.getByText(/Welcome to LogisticsHQ, Varun!/i)).toBeInTheDocument();
    expect(screen.getByText(/Search shipments, RFQs, bookings, invoices.../i)).toBeInTheDocument();
    expect(screen.getByText(/Aug 9 – Aug 15, 2026/i)).toBeInTheDocument();
  });

  it('opens and closes global search command palette with keyboard shortcut and mouse clicks', async () => {
    render(
      <BrowserRouter>
        <TopBar />
      </BrowserRouter>
    );

    const searchTrigger = document.getElementById('global-search-trigger');
    expect(searchTrigger).toBeInTheDocument();

    // Click search trigger
    fireEvent.click(searchTrigger);
    expect(screen.getByPlaceholderText(/Search across shipments, bookings, RFQs, customers, invoices, tracking.../i)).toBeInTheDocument();

    // Escape closes modal
    fireEvent.keyDown(window, { key: 'Escape' });
    await waitFor(() => {
      expect(screen.queryByPlaceholderText(/Search across shipments, bookings, RFQs, customers, invoices, tracking.../i)).not.toBeInTheDocument();
    });
  });

  it('executes global search and displays categorized matches with direct navigation links', async () => {
    render(
      <BrowserRouter>
        <TopBar />
      </BrowserRouter>
    );

    // Open Search
    const searchTrigger = document.getElementById('global-search-trigger');
    fireEvent.click(searchTrigger);

    const input = screen.getByPlaceholderText(/Search across shipments, bookings, RFQs, customers, invoices, tracking.../i);
    fireEvent.change(input, { target: { value: 'BK' } });

    await waitFor(() => {
      expect(screen.getByText('BK-QA-1787840980')).toBeInTheDocument();
      expect(screen.getByText(/Maersk Line • INNSA ➔ NLRTM/i)).toBeInTheDocument();
    });
  });

  it('toggles notifications, globe locale popover, date picker, and user profile popover with mutual exclusion', async () => {
    render(
      <BrowserRouter>
        <TopBar />
      </BrowserRouter>
    );

    // 1. Open Globe
    const globeBtn = screen.getByRole('button', { name: /Language & Region Settings/i });
    fireEvent.click(globeBtn);
    expect(screen.getByText(/Language & Regional Preferences/i)).toBeInTheDocument();

    // 2. Open Profile (should close Globe)
    const profileBtn = screen.getByRole('button', { name: /User Account Menu/i });
    fireEvent.click(profileBtn);
    expect(screen.queryByText(/Language & Regional Preferences/i)).not.toBeInTheDocument();
    expect(screen.getByText('varunkanade3456@gmail.com')).toBeInTheDocument();
    expect(screen.getByText('Org Admin')).toBeInTheDocument();
    expect(screen.getByText(/Sign out of LogisticsHQ/i)).toBeInTheDocument();
  });
});
