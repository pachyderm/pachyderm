import {rest} from 'msw';

import {Empty} from '@dash-frontend/api/googleTypes';
import {Version} from '@dash-frontend/api/version';

export const mockGetVersionInfo = () =>
  rest.post<Empty, Empty, Version>(
    '/api/versionpb_v2.API/GetVersion',
    (_req, res, ctx) => {
      return res(
        ctx.json({
          major: 0,
          minor: 0,
          micro: 0,
          additional: '',
          gitCommit: '',
          gitTreeModified: '',
          buildDate: '',
          goVersion: '',
          platform: '',
        }),
      );
    },
  );
