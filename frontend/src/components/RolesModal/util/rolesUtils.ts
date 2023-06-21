import {GetRolesQuery, Permission, ResourceType} from '@graphqlTypes';

import {GET_AUTHORIZE} from '@dash-frontend/queries/GetAuthorize';

import {Principal, Roles, UserTableRoles} from '../hooks/useRolesModal';

// create mapping of Record<string, string[]> for {principal: roles[]}, filtering out ignored roles
export const reduceAndFilterRoleBindings = (
  roles: GetRolesQuery['getRoles'],
  roleFilter?: (role: string | null) => boolean,
) => {
  const a = roles?.roleBindings?.reduce(
    (acc: Record<string, string[]>, role) => {
      if (role) {
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
  return a;
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

export const getProjectPermissionQueries = (resourceName: string) =>
  [
    Permission.PROJECT_MODIFY_BINDINGS,
    Permission.PROJECT_CREATE,
    Permission.PROJECT_DELETE,
  ].map((permission) => ({
    query: GET_AUTHORIZE,
    variables: {
      args: {
        permissionsList: [permission],
        resource: {type: ResourceType.PROJECT, name: resourceName},
      },
    },
  }));

export const getRepoPermissionQueries = (resourceName: string) =>
  [
    Permission.REPO_MODIFY_BINDINGS,
    Permission.REPO_DELETE,
    Permission.REPO_WRITE,
  ].map((permission) => ({
    query: GET_AUTHORIZE,
    variables: {
      args: {
        permissionsList: [permission],
        resource: {type: ResourceType.PROJECT, name: resourceName},
      },
    },
  }));

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
  return [];
};
