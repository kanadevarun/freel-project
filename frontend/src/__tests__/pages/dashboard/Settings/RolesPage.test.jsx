import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
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

  it('renders without crashing', async () => {
    api.get.mockResolvedValueOnce({ data: [] });
    render(
      <MemoryRouter>
        <RolesPage />
      </MemoryRouter>
    );
    expect(document.body).toBeInTheDocument();
  });
});
