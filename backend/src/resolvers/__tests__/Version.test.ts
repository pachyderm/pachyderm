import {GET_VERSION_INFO_QUERY} from '@dash-frontend/queries/GetVersionInfoQuery';

import {executeQuery} from '@dash-backend/testHelpers';
import {GetVersionInfoQuery} from '@graphqlTypes';

describe('Version Resolver', () => {
  describe('versionInfo', () => {
    it('should return version information', async () => {
      const {data, errors = []} = await executeQuery<GetVersionInfoQuery>(
        GET_VERSION_INFO_QUERY,
      );
      expect(errors).toHaveLength(0);
      expect(data?.versionInfo.consoleVersion).toBe('unversioned-development');
      expect(data?.versionInfo.pachdVersion).toEqual(
        expect.objectContaining({
          __typename: 'Version',
          major: 2,
          minor: 5,
          micro: 4,
          platform: 'arm64',
          goVersion: 'g01.19',
          gitTreeModified: 'false',
          gitCommit: '143fff8a572a2ec019dce29363f9030b3599e1ff',
          buildDate: '2023-04-13T16:13:48Z',
          additional: 'additional',
        }),
      );
    });
  });
});
