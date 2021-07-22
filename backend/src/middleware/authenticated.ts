import {AuthenticationError} from 'apollo-server-errors';

import {Context, UnauthenticatedContext} from '@dash-backend/lib/types';
import {ResolverFn} from '@graphqlTypes';

const authenticated = <Parent, Args, Info>(
  target: ResolverFn<Parent, Args, Context, Info>,
): ResolverFn<Parent, Args, UnauthenticatedContext, Info> => {
  return (_args, _info, {account, ...rest}, _resolveInfo) => {
    if (!account) throw new AuthenticationError('User is not authenticated');

    return target(_args, _info, {account, ...rest}, _resolveInfo);
  };
};

export default authenticated;
