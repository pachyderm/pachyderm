import {
  ResourceType,
  Permission,
  GetRoleBindingResponse,
} from '@dash-frontend/api/auth';
import {ALLOWED_PRINCIPAL_TYPES} from '@dash-frontend/constants/rbac';

import {
  MappedRoleBindings,
  Principal,
  Roles,
  UserTableRoles,
} from '../hooks/useRolesModal';

// create mapping of Record<string, string[]> for {principal: roles[]},
// filtering out ignored roles and principals
export const reduceAndFilterRoleBindings = (
  roleBindings?: GetRoleBindingResponse,
  roleFilter?: (role: string | null) => boolean,
) => {
  const mappedRoleBindings: MappedRoleBindings = Object.keys(
    roleBindings?.binding?.entries || {},
  ).map((principal) => {
    const roles = Object.keys(
      roleBindings?.binding?.entries?.[principal].roles || {},
    ).map((role) => role);

    return {
      principal,
      roles,
    };
  });
  const reducedAndFilteredRoleBindings = mappedRoleBindings.reduce(
    (acc: Record<string, string[]>, role) => {
      const principalType = role?.principal.split(':')[0];
      if (
        role &&
        principalType &&
        ALLOWED_PRINCIPAL_TYPES.includes(principalType)
      ) {
        let filteredRoles = role.roles;
        if (roleFilter) {
          filteredRoles = filteredRoles.filter(roleFilter);
        }
        if (filteredRoles.length > 0) {
          acc[role.principal] = filteredRoles as string[];
        }
      }
      return acc;
    },
    {},
  );

  return reducedAndFilteredRoleBindings;
};

export const mergeConcat = (a: string[], b: string[]) => {
  if (Array.isArray(a) && Array.isArray(b)) {
    return a.concat(b);
  }
};

type mapTableRolesInput = {
  unlockedRoles: Record<Principal, Roles>;
  lockedRoles: Record<Principal, Roles>;
  higherRoles: Record<Principal, Roles>;
  deletedPrincipals: Principal[];
};

// map unlocked and locked roles to Record<string, {unlockedRoles: string[], lockedRoles: string[]}>
// For example: repo roles are unlocked/editable if they are on the repo resource level, but locked
// if they are inherited from the project resource level
export const mapTableRoles = ({
  unlockedRoles,
  lockedRoles,
  higherRoles,
  deletedPrincipals,
}: mapTableRolesInput) => {
  const userTableRoles: UserTableRoles = {};
  const sortedChipRoles: UserTableRoles = {};

  Object.keys(unlockedRoles).forEach((principal) => {
    userTableRoles[principal] = {
      lockedRoles: [],
      unlockedRoles: [],
      higherRoles: [],
    };
    userTableRoles[principal].unlockedRoles = unlockedRoles[principal];
  });

  Object.keys(lockedRoles).forEach((principal) => {
    if (!userTableRoles[principal]) {
      userTableRoles[principal] = {
        lockedRoles: [],
        unlockedRoles: [],
        higherRoles: [],
      };
    }
    userTableRoles[principal].lockedRoles = lockedRoles[principal];
  });

  Object.keys(higherRoles).forEach((principal) => {
    if (!userTableRoles[principal]) {
      userTableRoles[principal] = {
        lockedRoles: [],
        unlockedRoles: [],
        higherRoles: [],
      };
    }
    userTableRoles[principal].higherRoles = higherRoles[principal];
  });

  deletedPrincipals.forEach((principal) => {
    if (!userTableRoles[principal]) {
      userTableRoles[principal] = {
        lockedRoles: [],
        unlockedRoles: [],
        higherRoles: [],
      };
    }
  });

  Object.keys(userTableRoles)
    .sort()
    .forEach((principal) => {
      sortedChipRoles[principal] = userTableRoles[principal];
    });

  return sortedChipRoles;
};

export const getProjectPermissionQueries = (resourceName: string) => ({
  permissions: [
    Permission.PROJECT_MODIFY_BINDINGS,
    Permission.PROJECT_CREATE,
    Permission.PROJECT_DELETE,
  ],
  resource: {type: ResourceType.PROJECT, name: resourceName},
});

export const getRepoPermissionQueries = (resourceName: string) => ({
  permissions: [
    Permission.REPO_MODIFY_BINDINGS,
    Permission.REPO_DELETE,
    Permission.REPO_WRITE,
  ],
  resource: {type: ResourceType.PROJECT, name: resourceName},
});

export const getPermissionQueries = (
  resourceName: string,
  resourceType: ResourceType,
) => {
  if (resourceType === ResourceType.PROJECT) {
    return getProjectPermissionQueries(resourceName);
  }
  if (resourceType === ResourceType.REPO) {
    return getRepoPermissionQueries(resourceName);
  }
  return {};
};
