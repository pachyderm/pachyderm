import {LogMessage} from '@dash-backend/proto';
import {timestampFromObject} from '@dash-backend/proto/builders/protobuf';

import {LOGS} from './loadLimits';

const tutorial = [
  new LogMessage()
    .setPipelineName('montage')
    .setJobId('33b9af7d5d4343219bc8e02ff44cd55a')
    .setDatumId(
      '13a4f3a7f9c669ea6ef61fe5f74b50122b00bf8fc33d616d861c2b415e81d1de',
    )
    .setUser(false)
    .setMessage('started datum task')
    .setMaster(true)
    .setTs(
      timestampFromObject({
        seconds: 1614126189,
        nanos: 0,
      }),
    ),
  new LogMessage()
    .setPipelineName('montage')
    .setJobId('33b9af7d5d4343219bc8e02ff44cd55a')
    .setDatumId(
      '13a4f3a7f9c669ea6ef61fe5f74b50122b00bf8fc33d616d861c2b415e81d1de',
    )
    .setUser(true)
    .setMessage('finished datum task')
    .setTs(
      timestampFromObject({
        seconds: 1614126190,
        nanos: 0,
      }),
    ),
  new LogMessage()
    .setPipelineName('montage')
    .setJobId('23b9af7d5d4343219bc8e02ff44cd55a')
    .setDatumId(
      '0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
    )
    .setUser(false)
    .setMessage('started datum task')
    .setTs(
      timestampFromObject({
        seconds: 1616533099,
        nanos: 0,
      }),
    ),
  new LogMessage()
    .setPipelineName('montage')
    .setJobId('23b9af7d5d4343219bc8e02ff44cd55a')
    .setDatumId(
      '0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
    )
    .setUser(false)
    .setMessage('beginning to run user code')
    .setTs(
      timestampFromObject({
        seconds: 1616533100,
        nanos: 0,
      }),
    ),
  new LogMessage()
    .setPipelineName('montage')
    .setJobId('23b9af7d5d4343219bc8e02ff44cd55a')
    .setDatumId(
      '0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
    )
    .setUser(true)
    .setMessage(
      '/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.',
    )
    .setTs(
      timestampFromObject({
        seconds: 1616533101,
        nanos: 0,
      }),
    ),
  new LogMessage()
    .setPipelineName('montage')
    .setJobId('23b9af7d5d4343219bc8e02ff44cd55a')
    .setDatumId(
      '0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
    )
    .setUser(false)
    .setMessage('finished datum task')
    .setTs(
      timestampFromObject({
        seconds: 1616533106,
        nanos: 0,
      }),
    ),
  new LogMessage()
    .setPipelineName('montage')
    .setJobId('23b9af7d5d4343219bc8e02ff44cd55a')
    .setDatumId(
      '006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f14f',
    )
    .setUser(false)
    .setMessage('started datum task')
    .setTs(
      timestampFromObject({
        seconds: 1616533200,
        nanos: 0,
      }),
    ),
  new LogMessage()
    .setPipelineName('montage')
    .setJobId('23b9af7d5d4343219bc8e02ff44cd55a')
    .setDatumId(
      '006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f14f',
    )
    .setUser(false)
    .setMessage('finished datum task')
    .setTs(
      timestampFromObject({
        seconds: 1616533220,
        nanos: 0,
      }),
    ),
];

const duplicatedLogs = [...new Array(100).keys()].map((i) =>
  new LogMessage()
    .setPipelineName('likelihoods')
    .setJobId('23b9af7d5d4343219bc8e02ff4acd33a')
    .setUser(false)
    .setMessage(i === 99 ? 'last message' : tutorial[i % 5].getMessage())
    .setTs(
      timestampFromObject({
        seconds: 1616533106 + i * 1000,
        nanos: 0,
      }),
    ),
);

export const workspaceLogs = [
  new LogMessage()
    .setUser(false)
    .setMessage(
      '2021-06-09T18:55:52.14Z INFO auth.API.GetPermissionsForPrincipal {"duration":0.0979148,"request":{"resource":{"type":2,"name":"images"},"principal":"user:peterjfranchina@gmail.com"},"response":{"permissions":[200,204,205,208,211,210,212,213,301,201,206,207,209,214,202,203,129,125,126,127,128,120,121,122,123,124,118,119,131,139,132,133,134,135,136,137,143,144,146,145,100,101,102,103,104,105,109,110,111,112,113,147,140,141,142,114,115,116,117,138],"roles":["clusterAdmin"]}} ',
    ),
  new LogMessage()
    .setUser(false)
    .setMessage(
      '2021-06-09T18:55:53.07Z WARNING pfs-over-HTTP - TLS disabled: could not stat public cert at /pachd-tls-cert/tls.crt: stat /pachd-tls-cert/tls.crt: no such file or directory ',
    ),
  new LogMessage()
    .setUser(false)
    .setMessage(
      '2021-06-09T18:55:54.23Z INFO auth.API.WhoAmI {"duration":0.0011038,"request":{},"response":{"username":"user:peterjfranchina@gmail.com",\n"expiration":"2021-06-09T20:15:12.264132Z"\n}} ',
    ),
  new LogMessage()
    .setUser(false)
    .setMessage(
      '2021-06-09T18:55:55.77Z INFO pps.API.GetLogs {"request":{"tail":40,"since":{"seconds":86400}}} ',
    ),
  new LogMessage()
    .setUser(false)
    .setMessage('PPS master: processing event for "edges"'),
];

const getLoadLogMessages = (count: number) => {
  const now = Math.floor(new Date().getTime() / 1000);
  return [...new Array(count).keys()].map((i) => {
    return new LogMessage()
      .setUser(false)
      .setJobId('0-0')
      .setPipelineName('load-pipeline-0')
      .setMessage(`log message #${i}`)
      .setTs(
        timestampFromObject({
          seconds: now,
          nanos: 0,
        }),
      );
  });
};

export const pipelineAndJobLogs: {[projectId: string]: LogMessage[]} = {
  'Solar-Panel-Data-Sorting': tutorial,
  'Data-Cleaning-Process': duplicatedLogs,
  'Solar-Power-Data-Logger-Team-Collab': [],
  'Solar-Price-Prediction-Modal': [],
  'Egress-Examples': [],
  'Empty-Project': [],
  'Trait-Discovery': [],
  'Load-Project': getLoadLogMessages(LOGS),
  default: [...tutorial],
};
