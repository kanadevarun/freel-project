import { createContext, useContext, useMemo } from 'react';
import { useAuth } from './AuthContext';

/**
 * RBACContext — Role-Based Access Control for the LogisticsHQ frontend.
 *
 * Consumes `role` from AuthContext and provides two helpers:
 *   can(module, action) — returns true if the user has the given permission.
 *                         Permission string format: "MODULE:ACTION" (e.g. "LEADS:READ")
 *   hasRole(name)       — returns true if the user's role name matches.
 *
 * Example usage:
 *   const { can, hasRole } = useRBAC();
 *   if (can('LEADS', 'CREATE')) { ... }
 *   if (hasRole('SUPER_ADMIN'))  { ... }
 */

const RBACContext = createContext(null);

export function RBACProvider({ children }) {
  const { memberRole } = useAuth();

  /**
   * Build a Set of permissions for O(1) lookup.
   * memberRole shape (from backend LoginResponseData):
   *   { name: "SUPER_ADMIN", display_name: "...", permissions: ["LEADS:READ", "LEADS:CREATE", ...] }
   *
   * For backward-compat, memberRole may also be just a plain string role name
   * (legacy sessions stored before role object was added to AuthContext).
   */
  const permissionsSet = useMemo(() => {
    if (!memberRole) return new Set();

    // New format: role object with permissions array
    if (typeof memberRole === 'object' && Array.isArray(memberRole.permissions)) {
      return new Set(memberRole.permissions.map((p) => String(p).toUpperCase()));
    }

    // Legacy format: plain string. No individual permissions available.
    return new Set();
  }, [memberRole]);

  const roleName = useMemo(() => {
    if (!memberRole) return null;
    if (typeof memberRole === 'object' && memberRole.name) return String(memberRole.name).toUpperCase();
    if (typeof memberRole === 'string') return memberRole.toUpperCase();
    return null;
  }, [memberRole]);

  /**
   * can(module, action) — check if the user can perform `action` on `module`.
   *
   * @param {string} module — resource name, e.g. 'LEADS', 'RFQS', 'COMPANIES'
   * @param {string} action — action name, e.g. 'READ', 'CREATE', 'UPDATE', 'DELETE'
   * @returns {boolean}
   */
  function can(module, action) {
    if (!module || !action) return false;
    const key = `${String(module).toUpperCase()}:${String(action).toUpperCase()}`;
    return permissionsSet.has(key);
  }

  /**
   * hasRole(name) — check if the user's role matches the given name.
   *
   * @param {string} name — role name, e.g. 'SUPER_ADMIN', 'SALES', 'FINANCE'
   * @returns {boolean}
   */
  function hasRole(name) {
    if (!name || !roleName) return false;
    return roleName === String(name).toUpperCase();
  }

  /**
   * hasAnyRole(names) — check if the user's role is one of the given names.
   *
   * @param {string[]} names — array of role names
   * @returns {boolean}
   */
  function hasAnyRole(names) {
    if (!Array.isArray(names) || !roleName) return false;
    return names.some((n) => String(n).toUpperCase() === roleName);
  }

  const value = {
    can,
    hasRole,
    hasAnyRole,
    roleName,
    permissionsSet,
  };

  return <RBACContext.Provider value={value}>{children}</RBACContext.Provider>;
}

// eslint-disable-next-line react-refresh/only-export-components
export function useRBAC() {
  const ctx = useContext(RBACContext);
  if (!ctx) throw new Error('useRBAC must be used inside <RBACProvider>');
  return ctx;
}
