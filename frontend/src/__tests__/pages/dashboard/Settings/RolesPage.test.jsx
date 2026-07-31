import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import RolesPage from '../../../../pages/dashboard/Settings/RolesPage';
import api from '../../../../services/api';

// Mock the API module
vi.mock('../../../../services/api', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
  },
}));

describe('RolesPage Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    // Keep API pending to check loading spinner
    api.get.mockImplementation(() => new Promise(() => {}));
    
    const { container } = render(<RolesPage />);
    expect(container.querySelector('.auth-spinner')).toBeInTheDocument();
  });

  it('renders roles and fetches permissions on load', async () => {
    api.get.mockImplementation((url) => {
      if (url === '/api/v1/roles') {
        return Promise.resolve([
          { id: 1, name: 'ADMIN', description: 'Admin role' },
          { id: 2, name: 'SALES', description: 'Sales role' }
        ]);
      }
      if (url === '/api/v1/roles/1/permissions') {
        return Promise.resolve({
          role_id: 1,
          permissions: [
            { resource: 'LEADS', action: 'CREATE' }
          ]
        });
      }
      return Promise.resolve(null);
    });

    render(<RolesPage />);

    // Wait for Roles to load
    await waitFor(() => {
      expect(screen.getByText('ADMIN')).toBeInTheDocument();
      expect(screen.getByText('SALES')).toBeInTheDocument();
    });

    // Wait for Permissions to load
    await waitFor(() => {
      const checkboxes = screen.getAllByRole('checkbox');
      expect(checkboxes.length).toBeGreaterThan(0);
    });
  });

  it('renders empty state if no roles returned', async () => {
    api.get.mockResolvedValueOnce([]);
    
    render(<RolesPage />);

    await waitFor(() => {
      expect(screen.getByText('No Roles Found')).toBeInTheDocument();
    });
  });
});
