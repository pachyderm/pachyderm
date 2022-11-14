import {
  DatumInfo,
  FileInfo,
  File,
  DatumState,
  Datum,
} from '@dash-backend/proto';

const allDatums = {
  montage: {
    '33b9af7d5d4343219bc8e02ff44cd55a': [
      new DatumInfo()
        .setState(DatumState.SUCCESS)
        .setDatum(new Datum().setId('a'))
        .setDataList([
          new FileInfo()
            .setFile(new File().setPath('/dummyData.csv'))
            .setSizeBytes(5),
        ]),
      new DatumInfo()
        .setState(DatumState.FAILED)
        .setDatum(new Datum().setId('b'))
        .setDataList([
          new FileInfo()
            .setFile(new File().setPath('/failedData.csv'))
            .setSizeBytes(0),
        ]),
    ],
  },
};

export type Datums = {
  [projectId: string]: {[pipelineId: string]: {[jobId: string]: DatumInfo[]}};
};

const datums: Datums = {
  '1': allDatums,
  '2': allDatums,
  '3': allDatums,
  '4': allDatums,
  '5': allDatums,
  default: allDatums,
};

export default datums;
