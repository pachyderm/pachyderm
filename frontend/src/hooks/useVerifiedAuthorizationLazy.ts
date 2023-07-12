import {GetAuthorizeArgs} from '@graphqlTypes';

import {useGetAuthorizeLazyQuery} from '@dash-frontend/generated/hooks';

export const verifyAuthorizationStatus = (bool?: boolean | null) => {
  // If auth is disabled, this will return true because the auth value will be undefined.
  // We use == instead of === because == performs type coercion and treats null and undefined as equal.
  // If bool is either null or undefined, it returns true, else it returns the value of bool.
  return bool == null ? true : bool;
};

// isAuthorizedAction will account for the case that auth is not enabled and return true.
export const useVerifiedAuthorizationLazy = (args: GetAuthorizeArgs) => {
  const [query, {data, loading, ...rest}] = useGetAuthorizeLazyQuery();

  const isAuthorizedAction = verifyAuthorizationStatus(
    data?.getAuthorize?.authorized,
  );

  const isAuthActive = data?.getAuthorize?.authorized != null;

  return {
    checkRolesPermission: () =>
      query({
        variables: {
          args,
        },
      }),
    loading,
    isAuthorizedAction,
    isAuthActive,
    ...data?.getAuthorize,
    ...rest,
  };
};
