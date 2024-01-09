import mergeWith from 'lodash/mergeWith';
import {useMemo, useState, useRef, useCallback, useEffect} from 'react';

import {ResourceType, ModifyRoleBindingRequest} from '@dash-frontend/api/auth';
import {
  PROJECT_ROLES,
  REPO_ROLES,
  IGNORED_CLUSTER_ROLES,
} from '@dash-frontend/constants/rbac';
import {useModifyRoleBinding} from '@dash-frontend/hooks/useModifyRoleBinding';
import {useRoleBinding} from '@dash-frontend/hooks/useRoleBinding';

import {
  reduceAndFilterRoleBindings,
  mergeConcat,
  mapTableRoles,
} from '../util/rolesUtils';

export type Principal = string;
export type Roles = string[];

export type MappedRoleBinding = {
  principal: Principal;
  roles: Roles;
};
export type MappedRoleBindings = MappedRoleBinding[];

export type UserTableRoles = Record<
  Principal,
  {
    unlockedRoles: Roles;
    lockedRoles: Roles;
    higherRoles: Roles;
  }
>;

type useRolesModalProps = {
  projectName: string;
  repoName?: string;
  resourceType: ResourceType;
};

const useRolesModal = ({
  projectName,
  repoName,
  resourceType,
}: useRolesModalProps) => {
  const [principalFilterOpen, setPrincipalFilterOpen] = useState(false);
  const [principalFilter, setPrincipalFilter] = useState('');
  const principalFilterRef = useRef<HTMLInputElement>(null);

  const [deletedRoles, setDeletedRoles] = useState<
    Record<Principal, ModifyRoleBindingRequest>
  >({});

  useEffect(() => {
    if (principalFilterOpen && principalFilterRef.current) {
      principalFilterRef.current.focus();
    }
  }, [principalFilterOpen]);

  const openPrincipalFilter = useCallback(() => {
    setPrincipalFilterOpen(true);
  }, [setPrincipalFilterOpen]);

  const exitPrincipalFilter = useCallback(() => {
    if (principalFilterRef.current) {
      principalFilterRef.current.value = '';
    }
    setPrincipalFilter('');
    setPrincipalFilterOpen(false);
  }, [setPrincipalFilter, setPrincipalFilterOpen, principalFilterRef]);

  const resourceName =
    resourceType === ResourceType.REPO && repoName
      ? `${projectName}/${repoName}`
      : projectName;

  const {
    roleBinding: clusterRolesResponse,
    loading: clusterRolesLoading,
    error: clusterRolesError,
  } = useRoleBinding({
    resource: {name: '', type: ResourceType.CLUSTER},
  });

  const {
    roleBinding: projectRolesResponse,
    loading: projectRolesLoading,
    error: projectRolesError,
  } = useRoleBinding({
    resource: {name: projectName, type: ResourceType.PROJECT},
  });

  const {
    roleBinding: repoRolesResponse,
    loading: repoRolesLoading,
    error: repoRolesError,
  } = useRoleBinding(
    {
      resource: {name: resourceName, type: ResourceType.REPO},
    },
    !!repoName,
  );

  const {
    modifyRoleBinding,
    loading: modifyRolesLoading,
    error: modifyRolesError,
  } = useModifyRoleBinding({
    resource: {name: resourceName, type: resourceType},
  });

  const loading =
    clusterRolesLoading || projectRolesLoading || repoRolesLoading;

  const error = clusterRolesError || projectRolesError || repoRolesError;

  const clusterRoles = useMemo(
    () =>
      reduceAndFilterRoleBindings(
        clusterRolesResponse,
        (role) => !!role && !IGNORED_CLUSTER_ROLES.includes(role),
      ),
    [clusterRolesResponse],
  );

  const clusterNotProjectRoles = useMemo(
    () =>
      reduceAndFilterRoleBindings(
        clusterRolesResponse,
        (role) =>
          !!role &&
          !PROJECT_ROLES.includes(role) &&
          !IGNORED_CLUSTER_ROLES.includes(role),
      ),
    [clusterRolesResponse],
  );

  const clusterProjectRoles = useMemo(
    () =>
      reduceAndFilterRoleBindings(
        clusterRolesResponse,
        (role) =>
          !!role &&
          PROJECT_ROLES.includes(role) &&
          !IGNORED_CLUSTER_ROLES.includes(role),
      ),
    [clusterRolesResponse],
  );

  const projectRoles = useMemo(
    () =>
      reduceAndFilterRoleBindings(
        projectRolesResponse,
        (role) => !!role && !IGNORED_CLUSTER_ROLES.includes(role),
      ),
    [projectRolesResponse],
  );

  const projectNotRepoRoles = useMemo(
    () =>
      reduceAndFilterRoleBindings(
        projectRolesResponse,
        (role) =>
          !!role &&
          !REPO_ROLES.includes(role) &&
          !IGNORED_CLUSTER_ROLES.includes(role),
      ),
    [projectRolesResponse],
  );

  const projectRepoRoles = useMemo(
    () =>
      reduceAndFilterRoleBindings(
        projectRolesResponse,
        (role) => !!role && REPO_ROLES.includes(role),
      ),
    [projectRolesResponse],
  );

  const repoRoles = useMemo(
    () => reduceAndFilterRoleBindings(repoRolesResponse),
    [repoRolesResponse],
  );

  const userTableRoles = useMemo(() => {
    return resourceType === ResourceType.REPO
      ? mapTableRoles({
          unlockedRoles: repoRoles || {},
          lockedRoles: projectRepoRoles || {},
          higherRoles:
            mergeWith(
              {},
              clusterRoles || {},
              projectNotRepoRoles || {},
              mergeConcat,
            ) || {},
          deletedPrincipals: Object.keys(deletedRoles) || [],
        })
      : mapTableRoles({
          unlockedRoles: projectRoles || {},
          lockedRoles: clusterProjectRoles || {},
          higherRoles: clusterNotProjectRoles || {},
          deletedPrincipals: Object.keys(deletedRoles) || [],
        });
  }, [
    clusterNotProjectRoles,
    clusterProjectRoles,
    clusterRoles,
    deletedRoles,
    projectNotRepoRoles,
    projectRepoRoles,
    projectRoles,
    repoRoles,
    resourceType,
  ]);

  const filteredUserTableRoles = useMemo(() => {
    const roles: UserTableRoles = {};
    Object.keys(userTableRoles).forEach((user) => {
      if (
        !principalFilter ||
        user.toLowerCase().includes(principalFilter.toLowerCase())
      ) {
        roles[user] = userTableRoles[user];
      }
    });
    return roles;
  }, [principalFilter, userTableRoles]);

  const deleteAllRoles =
    (principal: string, allRoles: string[]) => async () => {
      modifyRoleBinding(
        {
          resource: {
            name: resourceName,
            type: resourceType,
          },
          principal,
          roles: [],
        },
        {
          onSettled: () => {
            const updatedDeletedRoles = {...deletedRoles};

            updatedDeletedRoles[principal] = {
              principal,
              resource: {
                name: resourceName,
                type: resourceType,
              },
              roles: allRoles,
            };
            setDeletedRoles(updatedDeletedRoles);
          },
        },
      );
    };

  const undoDeleteAllRoles = (principal: string) => async () => {
    if (deletedRoles[principal]) {
      modifyRoleBinding(
        {
          ...deletedRoles[principal],
        },
        {
          onSettled: () => {
            const updatedDeletedRoles = {...deletedRoles};

            delete updatedDeletedRoles[principal];
            setDeletedRoles(updatedDeletedRoles);
          },
        },
      );
    }
  };

  const deleteRole =
    (principal: string, allRoles: string[], roleName: string) => async () => {
      modifyRoleBinding({
        resource: {
          name: resourceName,
          type: resourceType,
        },
        principal,
        roles: allRoles.filter((r) => r !== roleName),
      });
    };

  const addSelectedRole =
    (principal: string, allRoles: string[]) => async (role: string) => {
      modifyRoleBinding(
        {
          resource: {
            name: resourceName,
            type: resourceType,
          },
          principal,
          roles: allRoles.concat(role),
        },
        {
          onSettled: () => {
            const updatedDeletedRoles = {...deletedRoles};
            delete updatedDeletedRoles[principal];
            setDeletedRoles(updatedDeletedRoles);
          },
        },
      );
    };

  return {
    deletedRoles,
    setDeletedRoles,
    principalFilterRef,
    principalFilterOpen,
    exitPrincipalFilter,
    openPrincipalFilter,
    setPrincipalFilter,
    userTableRoles: filteredUserTableRoles,
    loading,
    error,
    resourceName,
    deleteAllRoles,
    undoDeleteAllRoles,
    deleteRole,
    addSelectedRole,
    modifyRolesLoading,
    modifyRolesError,
  };
};

export default useRolesModal;
