import {useQuery} from '@tanstack/react-query';

import {
  authorize,
  AuthorizeRequest,
  AuthorizeResponse,
  Permission,
  ResourceType,
} from '@dash-frontend/api/auth';
import {isErrorWithMessage, isUnknown} from '@dash-frontend/api/utils/error';
import getErrorMessage from '@dash-frontend/lib/getErrorMessage';
import queryKeys from '@dash-frontend/lib/queryKeys';

const useAuthorizeCommon = (
  data:
    | AuthorizeResponse
    | {
        satisfied: never[];
        missing: never[];
        principal: string;
        authorized: null;
      }
    | undefined,
) => {
  const authorized = data?.authorized;
  const isAuthActive = authorized !== null && authorized !== undefined;
  const hasAllPermissions = isAuthActive ? authorized : true;
  const satisfied = data?.satisfied || [];

  const hasPermission = (permission: Permission) =>
    !isAuthActive || satisfied.includes(permission);

  const lookups = {
    hasClusterAuthSetConfig: hasPermission(Permission.CLUSTER_AUTH_SET_CONFIG),
    hasProjectCreate: hasPermission(Permission.PROJECT_CREATE),
    hasProjectCreateRepo: hasPermission(Permission.PROJECT_CREATE_REPO),
    hasProjectDelete: hasPermission(Permission.PROJECT_DELETE),
    hasProjectModifyBindings: hasPermission(Permission.PROJECT_MODIFY_BINDINGS),
    hasProjectSetDefaults: hasPermission(Permission.PROJECT_SET_DEFAULTS),
    hasRepoWrite: hasPermission(Permission.REPO_WRITE),
    hasRepoRead: hasPermission(Permission.REPO_READ),
    hasRepoDelete: hasPermission(Permission.REPO_DELETE),
    hasRepoModifyBindings: hasPermission(Permission.REPO_MODIFY_BINDINGS),
  };

  return {
    isAuthActive,
    hasAllPermissions,
    ...lookups,
  };
};

export const useAuthorize = (req: AuthorizeRequest, enabled = true) => {
  const {
    data,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: queryKeys.authorize<AuthorizeRequest>({args: req}),
    queryFn: () => authorize(req),
    enabled,
    throwOnError: (e) =>
      // This error will occur when you auth on a resource that does not exist
      !(
        isUnknown(e) &&
        isErrorWithMessage(e) &&
        e?.message.includes('no role binding exists for ')
      ),
    retry: false,
    refetchInterval: false,
  });

  const permissions = useAuthorizeCommon(data);

  return {
    ...permissions,
    loading,
    error: getErrorMessage(error),
  };
};

export const useAuthorizeLazy = (req: AuthorizeRequest) => {
  const {
    data,
    isLoading: loading,
    error,
    refetch,
  } = useQuery({
    queryKey: queryKeys.authorize<AuthorizeRequest>({args: req}),
    queryFn: () => authorize(req),
    enabled: false,
    retry: false,
    refetchInterval: false,
  });

  const permissions = useAuthorizeCommon(data);

  return {
    ...permissions,
    loading,
    error: getErrorMessage(error),
    checkPermissions: refetch,
  };
};

export {Permission, ResourceType};
