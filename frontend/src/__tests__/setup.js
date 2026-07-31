/**
 * Vitest global setup file.
 * Loaded before all tests.
 *
 * Note: This project requires Node 22+/26 with --no-webstorage to prevent
 * Node's experimental Web Storage API from interfering with jsdom's
 * localStorage. This is configured in vitest.config.js via execArgv.
 */
import '@testing-library/jest-dom';

// Reset localStorage between tests to prevent bleed.
afterEach(() => {
  localStorage.clear();
});
