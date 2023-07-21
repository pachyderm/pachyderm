import {GetVersionInfoQuery} from '@graphqlTypes';
import {graphql} from 'msw';

export const mockGetVersionInfo = () =>
  graphql.query<GetVersionInfoQuery>('getVersionInfo', (_req, res, ctx) => {
    return res(
      ctx.data({
        versionInfo: {
          pachdVersion: {
            major: 0,
            minor: 0,
            micro: 0,
            additional: '',
            gitCommit: '',
            gitTreeModified: '',
            buildDate: '',
            goVersion: '',
            platform: '',
          },
          consoleVersion: 'test',
        },
      }),
    );
  });
