type ClusterRoles = 'clusterAdmin';
type ProjectRoles = 'projectViewer' | 'projectWriter' | 'projectOwner';
type RepoRoles = 'repoReader' | 'repoWriter' | 'repoOwner';
export type AllRoles = ClusterRoles | ProjectRoles | RepoRoles;

// Each role has a set of roles which are all roles that inherit the given role.
// In other words, the items in the set are roles that equal or inherit the given role.
const roleHierarchy: Record<AllRoles, Set<AllRoles>> = {
  clusterAdmin: new Set(['clusterAdmin']),
  projectOwner: new Set(['clusterAdmin', 'projectOwner']),
  projectWriter: new Set(['clusterAdmin', 'projectOwner', 'projectWriter']),
  projectViewer: new Set([
    'clusterAdmin',
    'projectOwner',
    'projectWriter',
    'projectViewer',
  ]),
  repoOwner: new Set(['clusterAdmin', 'repoOwner']),
  repoWriter: new Set(['clusterAdmin', 'repoOwner', 'repoWriter']),
  repoReader: new Set([
    'clusterAdmin',
    'repoOwner',
    'repoWriter',
    'repoReader',
  ]),
};

/**
 * This function is intended for internal use because Pachyderm protos do not enumerate userRole types. Use `hasAtLeastRole` instead.
 */
export const safeHasAtLeastRole = (
  requiredRole: AllRoles,
  userRoles?: AllRoles[],
) => {
  if (!userRoles) return true; // auth is not enabled
  if (userRoles.length === 0) return false;
  const sufficientRoles = roleHierarchy[requiredRole];
  return userRoles.some((role) => sufficientRoles.has(role));
};

export const hasAtLeastRole = (
  requiredRole: AllRoles,
  userRoles?: Array<string | null> | null, // This was taken from GQL frontend type ReposWithCommitQuery['repos']['authInfo']['rolesList]
) => safeHasAtLeastRole(requiredRole, userRoles as unknown as AllRoles[]);
