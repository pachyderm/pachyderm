import {toGQLEnterpriseState} from '@dash-backend/lib/gqlEnumMappers';
import {GetStateResponse} from '@dash-backend/proto';
import {EnterpriseInfo} from '@graphqlTypes';

export const enterpriseInfoToGQLInfo = (
  enterpriseState: GetStateResponse.AsObject,
): EnterpriseInfo => {
  return {
    expiration: enterpriseState.info?.expires?.seconds || -1,
    state: toGQLEnterpriseState(enterpriseState.state),
  };
};
