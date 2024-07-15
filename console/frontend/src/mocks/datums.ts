import merge from 'lodash/merge';
import {rest} from 'msw';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  DatumInfo,
  DatumState,
  InspectDatumRequest,
  ListDatumRequest,
} from '@dash-frontend/api/pps';
import {RequestError} from '@dash-frontend/api/utils/error';

export const buildDatum = (datum: Partial<DatumInfo>): DatumInfo => {
  const defaultDatum: DatumInfo = {
    datum: {
      id: 'default',
      job: {
        id: 'default',
        pipeline: {
          name: 'default',
          project: {
            name: 'default',
          },
        },
      },
    },
    stats: {
      downloadTime: '',
      uploadTime: '',
      processTime: '',
      downloadBytes: '',
      uploadBytes: '',
    },
    state: DatumState.SUCCESS,
    __typename: 'DatumInfo',
  };

  return merge(defaultDatum, datum);
};

export const mockEmptyDatums = () =>
  rest.post<ListDatumRequest, Empty, DatumInfo[]>(
    '/api/pps_v2.API/ListDatum',
    (_req, res, ctx) => {
      return res(ctx.json([]));
    },
  );

export const JOB_5C_DATUM_INFO_05: DatumInfo = buildDatum({
  datum: {
    id: '05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7',
    job: {
      id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
    },
  },
  state: DatumState.SUCCESS,
  stats: {
    downloadTime: '2.00147s',
    uploadTime: '3.000792s',
    processTime: '1.055435s',
    downloadBytes: '1000',
    uploadBytes: '3000',
  },
});

const JOB_5C_DATUM_INFO_0B: DatumInfo = buildDatum({
  datum: {
    id: '0b9f1328700f60a8edbd2fd510bd65c7e03c7d5ca167c4f58119354da25920e0',
    job: {
      id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
    },
  },
  state: DatumState.SUCCESS,
  stats: {
    downloadTime: '0.001832s',
    uploadTime: '0.000541s',
    processTime: '1.002623s',
    downloadBytes: '18',
    uploadBytes: '18',
  },
});

const JOB_5C_DATUM_INFO_6F: DatumInfo = buildDatum({
  datum: {
    id: '6F6529b0e3459b521225aa3921946a9dae38aa7dd39d53554a8bf4f6bc1d7987',
    job: {
      id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
    },
  },
  state: DatumState.FAILED,
  stats: {
    downloadTime: '0.001049s',
    uploadTime: undefined,
    processTime: '0.507925s',
    downloadBytes: '20',
    uploadBytes: '0',
  },
});

const JOB_5C_DATUM_INFO_CH: DatumInfo = buildDatum({
  datum: {
    id: 'ch3db37fa4594a00ebf1dc972f81b58de642cd0cfca811e1b5bd6a2bb292a8e0',
    job: {
      id: '14291af7da4a4143b8ae12eba16d4661',
    },
  },
  state: DatumState.SKIPPED,
  stats: {
    downloadTime: '0.008497s',
    uploadTime: '0.002798s',
    processTime: '2.53238s',
    downloadBytes: '58644',
    uploadBytes: '22783',
  },
});

const MOCK_JOB_5C_DATUMS_INFO: DatumInfo[] = [
  JOB_5C_DATUM_INFO_05,
  JOB_5C_DATUM_INFO_0B,
  JOB_5C_DATUM_INFO_6F,
  JOB_5C_DATUM_INFO_CH,
];

export const mockGetJob5CDatums = () =>
  rest.post<ListDatumRequest, Empty, DatumInfo[]>(
    '/api/pps_v2.API/ListDatum',
    (_req, res, ctx) => {
      return res(ctx.json(MOCK_JOB_5C_DATUMS_INFO));
    },
  );

export const mockGetJob5CDatum05 = () =>
  rest.post<InspectDatumRequest, Empty, DatumInfo | RequestError>(
    '/api/pps_v2.API/InspectDatum',
    async (req, res, ctx) => {
      const body = await req.json();
      if (
        body.datum.id ===
          '05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7' &&
        body.datum.job.id === '5c1aa9bc87dd411ba5a1be0c80a3ebc2'
      ) {
        return res(ctx.json(JOB_5C_DATUM_INFO_05));
      }

      return res(
        ctx.status(400),
        ctx.json({
          code: 2,
          message:
            'datum 00d9ef90cb4294a7115adcca6b7a3029024151f3889e5292601638c91d1cc02b not found in job default/generate-datums@c64db5bdf8544d2e9fd308e6c8df891a',
          details: [],
        }),
      );
    },
  );

export const mockGetJob5CDatumCH = () =>
  rest.post<InspectDatumRequest, Empty, DatumInfo>(
    '/api/pps_v2.API/InspectDatum',
    (_req, res, ctx) => res(ctx.json(JOB_5C_DATUM_INFO_CH)),
  );

type generateDatumsArgs = {
  n: number;
  jobId?: string;
};

export const generatePagingDatums = ({
  n,
  jobId = '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
}: generateDatumsArgs) => {
  const datums: DatumInfo[] = [];

  for (let i = 0; i < n; i++) {
    datums.push({
      datum: {
        id: `${i.toString()}a${'0'.repeat(63 - i.toString().length)}`,
        job: {
          id: jobId,
        },
      },
      state: DatumState.SUCCESS,
    });
  }
  return datums;
};
