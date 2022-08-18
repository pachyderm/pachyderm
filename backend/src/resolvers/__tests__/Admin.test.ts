import {GET_ADMIN_INFO_QUERY} from '@dash-frontend/queries/GetAdminInfoQuery';

import {executeQuery} from '@dash-backend/testHelpers';
import {GetAdminInfoQuery} from '@graphqlTypes';

describe('Admin Resolver', () => {
  describe('adminInfo', () => {
    it('should return admin information', async () => {
      const {data, errors = []} = await executeQuery<GetAdminInfoQuery>(
        GET_ADMIN_INFO_QUERY,
      );
      expect(errors.length).toBe(0);
      expect(data?.adminInfo.clusterId).toEqual(
        '838327d0f477415799b6da3706d89310',
      );
    });
  });
});
