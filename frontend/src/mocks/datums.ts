import {
  Datum,
  DatumState,
  DatumsQuery,
  mockDatumQuery,
  mockDatumsQuery,
} from '@graphqlTypes';
import merge from 'lodash/merge';

export const MOCK_EMPTY_DATUMS: DatumsQuery = {
  datums: {
    items: [],
    cursor: null,
    hasNextPage: false,
    __typename: 'PageableDatum',
  },
};

export const buildDatum = (datum: Partial<Datum>): Datum => {
  const defaultDatum = {
    id: 'default',
    requestedJobId: 'default',
    state: DatumState.SUCCESS,
    jobId: null,
    downloadTimestamp: null,
    uploadTimestamp: null,
    processTimestamp: null,
    downloadBytes: null,
    uploadBytes: null,
    __typename: 'Datum',
  };

  return merge(defaultDatum, datum);
};

export const mockEmptyDatumsQuery = () =>
  mockDatumsQuery((_req, res, ctx) => {
    return res(ctx.data(MOCK_EMPTY_DATUMS));
  });

export const JOB_5C_DATUM_05: Datum = buildDatum({
  id: '05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7',
  jobId: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
  requestedJobId: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
  state: DatumState.SUCCESS,
  downloadTimestamp: {
    seconds: 2,
    nanos: 1469709,
    __typename: 'Timestamp',
  },
  uploadTimestamp: {
    seconds: 3,
    nanos: 791625,
    __typename: 'Timestamp',
  },
  processTimestamp: {
    seconds: 1,
    nanos: 55435418,
    __typename: 'Timestamp',
  },
  downloadBytes: 1000,
  uploadBytes: 3000,
  __typename: 'Datum',
});

const JOB_5C_DATUM_0B: Datum = buildDatum({
  id: '0b9f1328700f60a8edbd2fd510bd65c7e03c7d5ca167c4f58119354da25920e0',
  jobId: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
  requestedJobId: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
  state: DatumState.SUCCESS,
  downloadTimestamp: {
    seconds: 0,
    nanos: 1832000,
    __typename: 'Timestamp',
  },
  uploadTimestamp: {
    seconds: 0,
    nanos: 540500,
    __typename: 'Timestamp',
  },
  processTimestamp: {
    seconds: 1,
    nanos: 2623209,
    __typename: 'Timestamp',
  },
  downloadBytes: 18,
  uploadBytes: 18,
  __typename: 'Datum',
});

const JOB_5C_DATUM_6F: Datum = buildDatum({
  id: '6F6529b0e3459b521225aa3921946a9dae38aa7dd39d53554a8bf4f6bc1d7987',
  jobId: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
  requestedJobId: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
  state: DatumState.FAILED,
  downloadTimestamp: {
    seconds: 0,
    nanos: 1048750,
    __typename: 'Timestamp',
  },
  uploadTimestamp: null,
  processTimestamp: {
    seconds: 0,
    nanos: 507925209,
    __typename: 'Timestamp',
  },
  downloadBytes: 20,
  uploadBytes: 0,
  __typename: 'Datum',
});

const JOB_5C_DATUM_CH: Datum = buildDatum({
  id: 'ch3db37fa4594a00ebf1dc972f81b58de642cd0cfca811e1b5bd6a2bb292a8e0',
  jobId: '14291af7da4a4143b8ae12eba16d4661',
  requestedJobId: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
  state: DatumState.SKIPPED,
  downloadTimestamp: {
    seconds: 0,
    nanos: 8496916,
    __typename: 'Timestamp',
  },
  uploadTimestamp: {
    seconds: 0,
    nanos: 2798417,
    __typename: 'Timestamp',
  },
  processTimestamp: {
    seconds: 2,
    nanos: 532379792,
    __typename: 'Timestamp',
  },
  downloadBytes: 58644,
  uploadBytes: 22783,
  __typename: 'Datum',
});

const MOCK_JOB_5C_DATUMS: Datum[] = [
  JOB_5C_DATUM_05,
  JOB_5C_DATUM_0B,
  JOB_5C_DATUM_6F,
  JOB_5C_DATUM_CH,
];

export const mockGetJob5CDatums = () =>
  mockDatumsQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        datums: {
          items: MOCK_JOB_5C_DATUMS,
          cursor: null,
          hasNextPage: false,
          __typename: 'PageableDatum',
        },
      }),
    );
  });

export const mockGetJob5CDatum05 = () =>
  mockDatumQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        datum: JOB_5C_DATUM_05,
      }),
    );
  });

export const mockGetJob5CDatumCH = () =>
  mockDatumQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        datum: JOB_5C_DATUM_CH,
      }),
    );
  });

type generateDatumsArgs = {
  n: number;
  jobId?: string;
};

export const generatePagingDatums = ({
  n,
  jobId = '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
}: generateDatumsArgs) => {
  const datums: Datum[] = [];

  for (let i = 0; i < n; i++) {
    datums.push({
      id: `${i.toString()}a${'0'.repeat(63 - i.toString().length)}`,
      requestedJobId: jobId,
      state: DatumState.SUCCESS,
      jobId: jobId,
      downloadTimestamp: null,
      uploadTimestamp: null,
      processTimestamp: null,
      downloadBytes: null,
      uploadBytes: null,
      __typename: 'Datum',
    });
  }
  return datums;
};
