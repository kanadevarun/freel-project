import { useRBAC } from '../context/RBACContext';

/**
 * PermissionGate — Renders children only if the user has the required permission
 * or matches one of the required roles.
 *
 * @param {string} module - The RBAC module (e.g., 'LEADS', 'RFQS')
 * @param {string} action - The RBAC action (e.g., 'READ', 'CREATE')
 * @param {string|string[]} role - Require specific role(s) (optional fallback)
 * @param {ReactNode} fallback - Rendered if access is denied (optional)
 * @param {ReactNode} children - The protected content
 */
export default function PermissionGate({ module, action, role, fallback = null, children }) {
  const { can, hasRole, hasAnyRole, memberRole } = useRBAC();

  // If no member role exists (not logged in), block access
  if (!memberRole) {
    return fallback;
  }

  // 1. Permission Check
  if (module && action) {
    if (can(module, action)) {
      return children;
    }
  }

  // 2. Role Check
  if (role) {
    if (Array.isArray(role)) {
      if (hasAnyRole(role)) {
        return children;
      }
    } else {
      if (hasRole(role)) {
        return children;
      }
    }
  }

  // Access denied
  return fallback;
}

