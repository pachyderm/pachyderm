import {
  Datum,
  DatumInfo,
  DatumState,
  File,
  FileInfo,
  Job,
  ProcessStats,
} from '@dash-backend/proto';
import {durationFromObject} from '@dash-backend/proto/builders/protobuf';

const allDatums = {
  montage: {
    '23b9af7d5d4343219bc8e02ff44cd55a': [
      new DatumInfo()
        .setState(DatumState.SUCCESS)
        .setDatum(
          new Datum().setId(
            '0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
          ),
        )
        .setDataList([
          new FileInfo()
            .setFile(new File().setPath('/dummyData.csv'))
            .setSizeBytes(5),
        ])
        .setStats(
          new ProcessStats()
            .setDownloadBytes(1000)
            .setDownloadTime(durationFromObject({seconds: 1, nanos: 123123123}))
            .setUploadBytes(2000)
            .setUploadTime(durationFromObject({seconds: 2, nanos: 23123123}))
            .setProcessTime(durationFromObject({seconds: 3, nanos: 3123123})),
        ),
      new DatumInfo()
        .setState(DatumState.SUCCESS)
        .setDatum(
          new Datum().setId(
            '006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f14f',
          ),
        )
        .setDataList([
          new FileInfo()
            .setFile(new File().setPath('/failedData.csv'))
            .setSizeBytes(0),
        ]),
      new DatumInfo()
        .setState(DatumState.FAILED)
        .setDatum(
          new Datum().setId(
            '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
          ),
        )
        .setDataList([
          new FileInfo()
            .setFile(new File().setPath('/failedData.csv'))
            .setSizeBytes(0),
        ]),
      new DatumInfo()
        .setState(DatumState.SKIPPED)
        .setDatum(
          new Datum()
            .setId(
              '1112b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
            )
            .setJob(
              new Job().setId(
                '2222b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
              ),
            ),
        )
        .setDataList([
          new FileInfo()
            .setFile(new File().setPath('/dummyData.csv'))
            .setSizeBytes(5),
        ])
        .setStats(
          new ProcessStats()
            .setDownloadBytes(1000)
            .setDownloadTime(durationFromObject({seconds: 1, nanos: 123123123}))
            .setUploadBytes(2000)
            .setUploadTime(durationFromObject({seconds: 2, nanos: 23123123}))
            .setProcessTime(durationFromObject({seconds: 3, nanos: 3123123})),
        ),
    ],
    '33b9af7d5d4343219bc8e02ff44cd55a': [
      new DatumInfo()
        .setState(DatumState.FAILED)
        .setDatum(
          new Datum().setId(
            '13a4f3a7f9c669ea6ef61fe5f74b50122b00bf8fc33d616d861c2b415e81d1de',
          ),
        )
        .setDataList([
          new FileInfo()
            .setFile(new File().setPath('/dummyData.csv'))
            .setSizeBytes(5),
        ]),
      new DatumInfo()
        .setState(DatumState.FAILED)
        .setDatum(
          new Datum().setId(
            '10045ee28dbb460a649a72bf1a1147159737c785f622c0c149ff89d7fcb66747',
          ),
        )
        .setDataList([
          new FileInfo()
            .setFile(new File().setPath('/failedData.csv'))
            .setSizeBytes(0),
        ]),
    ],
    '7798fhje5d4343219bc8e02ff4acd33a': [
      new DatumInfo()
        .setState(DatumState.SUCCESS)
        .setDatum(
          new Datum().setId(
            '123456789dbb460a649a72bf1a1147159737c785f622c0c149ff89d7fcb66747',
          ),
        ),
      new DatumInfo()
        .setState(DatumState.SKIPPED)
        .setDatum(
          new Datum().setId(
            '987654321dbb460a649a72bf1a1147159737c785f622c0c149ff89d7fcb66747',
          ),
        ),
    ],
  },
};

const lotsOfDatums = {
  likelihoods: {
    '23b9af7d5d4343219bc8e02ff4acd33a': [...new Array(100).keys()].map((i) => {
      const id = i.toString();
      return new DatumInfo()
        .setState(DatumState.SUCCESS)
        .setDatum(new Datum().setId(`${id}a${'0'.repeat(63 - id.length)}`));
    }),
  },
};

export type Datums = {
  [projectId: string]: {[pipelineId: string]: {[jobId: string]: DatumInfo[]}};
};

const datums: Datums = {
  'Solar-Panel-Data-Sorting': allDatums,
  'Data-Cleaning-Process': lotsOfDatums,
  'Solar-Power-Data-Logger-Team-Collab': allDatums,
  'Solar-Price-Prediction-Modal': allDatums,
  'Egress-Examples': allDatums,
  default: allDatums,
};

export default datums;
