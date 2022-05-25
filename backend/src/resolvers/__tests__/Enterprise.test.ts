import {GET_ENTERPRISE_INFO_QUERY} from '@dash-frontend/queries/GetEnterpriseInfoQuery';

import {executeQuery} from '@dash-backend/testHelpers';
import {EnterpriseState, GetEnterpriseInfoQuery} from '@graphqlTypes';

describe('Enterprise Resolver', () => {
  describe('enterpriseInfo', () => {
    it('should return information about the enterprise state', async () => {
      const {data, errors = []} = await executeQuery<GetEnterpriseInfoQuery>(
        GET_ENTERPRISE_INFO_QUERY,
      );

      expect(errors.length).toBe(0);
      expect(data?.enterpriseInfo.state).toEqual(EnterpriseState.ACTIVE);
      expect(data?.enterpriseInfo.expiration).toEqual(50596369);
    });
  });
});
