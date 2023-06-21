import {ResourceType, ModifyRolesArgs} from '@graphqlTypes';
import mergeWith from 'lodash/mergeWith';
import {useMemo, useState, useRef, useCallback, useEffect} from 'react';

import {
  PROJECT_ROLES,
  REPO_ROLES,
  IGNORED_CLUSTER_ROLES,
} from '@dash-frontend/constants/rbac';
import {
  useModifyRolesMutation,
  useGetRolesQuery,
} from '@dash-frontend/generated/hooks';
import {GET_ROLES_QUERY} from '@dash-frontend/queries/GetRolesQuery';

import {
  reduceAndFilterRoleBindings,
  mergeConcat,
  mapTableRoles,
  getPermissionQueries,
} from '../util/rolesUtils';

export type Principal = string;
export type Roles = string[];

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
    Record<Principal, ModifyRolesArgs>
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
    data: clusterRolesResponse,
    loading: clusterRolesLoading,
    error: clusterRolesError,
  } = useGetRolesQuery({
    variables: {
      args: {resource: {name: '', type: ResourceType.CLUSTER}},
    },
  });

  const {
    data: projectRolesResponse,
    loading: projectRolesLoading,
    error: projectRolesError,
  } = useGetRolesQuery({
    variables: {
      args: {resource: {name: projectName, type: ResourceType.PROJECT}},
    },
  });

  const {
    data: repoRolesResponse,
    loading: repoRolesLoading,
    error: repoRolesError,
  } = useGetRolesQuery({
    variables: {
      args: {
        resource: {
          name: resourceName,
          type: ResourceType.REPO,
        },
      },
    },
    skip: !repoName,
  });

  const [
    modifyRolesMutation,
    {loading: modifyRolesLoading, error: modifyRolesError},
  ] = useModifyRolesMutation({
    awaitRefetchQueries: true,
    refetchQueries: [
      {
        query: GET_ROLES_QUERY,
        variables: {
          args: {
            resource: {name: resourceName, type: resourceType},
          },
        },
      },
      ...getPermissionQueries(resourceName, resourceType),
    ],
  });

  const loading =
    clusterRolesLoading || projectRolesLoading || repoRolesLoading;

  const error = clusterRolesError || projectRolesError || repoRolesError;

  const clusterRoles = useMemo(
    () =>
      reduceAndFilterRoleBindings(
        clusterRolesResponse?.getRoles,
        (role) => !!role && !IGNORED_CLUSTER_ROLES.includes(role),
      ),
    [clusterRolesResponse],
  );

  const clusterNotProjectRoles = useMemo(
    () =>
      reduceAndFilterRoleBindings(
        clusterRolesResponse?.getRoles,
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
        clusterRolesResponse?.getRoles,
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
        projectRolesResponse?.getRoles,
        (role) => !!role && !IGNORED_CLUSTER_ROLES.includes(role),
      ),
    [projectRolesResponse],
  );

  const projectNotRepoRoles = useMemo(
    () =>
      reduceAndFilterRoleBindings(
        projectRolesResponse?.getRoles,
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
        projectRolesResponse?.getRoles,
        (role) => !!role && REPO_ROLES.includes(role),
      ),
    [projectRolesResponse],
  );

  const repoRoles = useMemo(
    () => reduceAndFilterRoleBindings(repoRolesResponse?.getRoles),
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
      await modifyRolesMutation({
        variables: {
          args: {
            resource: {
              name: resourceName,
              type: resourceType,
            },
            principal,
            rolesList: [],
          },
        },
        onCompleted: () => {
          const updatedDeletedRoles = {...deletedRoles};
          updatedDeletedRoles[principal] = {
            principal,
            resource: {
              name: resourceName,
              type: resourceType,
            },
            rolesList: allRoles,
          };
          setDeletedRoles(updatedDeletedRoles);
        },
      });
    };

  const undoDeleteAllRoles = (principal: string) => async () => {
    if (deletedRoles[principal]) {
      await modifyRolesMutation({
        variables: {
          args: deletedRoles[principal],
        },
        onCompleted: () => {
          const updatedDeletedRoles = {...deletedRoles};
          delete updatedDeletedRoles[principal];
          setDeletedRoles(updatedDeletedRoles);
        },
      });
    }
  };

  const deleteRole =
    (principal: string, allRoles: string[], roleName: string) => async () => {
      await modifyRolesMutation({
        variables: {
          args: {
            resource: {
              name: resourceName,
              type: resourceType,
            },
            principal,
            rolesList: allRoles.filter((r) => r !== roleName),
          },
        },
      });
    };

  const addSelectedRole =
    (principal: string, allRoles: string[]) => async (role: string) => {
      await modifyRolesMutation({
        variables: {
          args: {
            resource: {
              name: resourceName,
              type: resourceType,
            },
            principal,
            rolesList: allRoles.concat(role),
          },
        },
        onCompleted: () => {
          const updatedDeletedRoles = {...deletedRoles};
          delete updatedDeletedRoles[principal];
          setDeletedRoles(updatedDeletedRoles);
        },
      });
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
