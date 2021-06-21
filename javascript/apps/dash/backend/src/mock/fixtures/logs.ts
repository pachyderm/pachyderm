import {LogMessage} from '@pachyderm/proto/pb/pps/pps_pb';

import {timestampFromObject} from '@dash-backend/grpc/builders/protobuf';

const tutorial = [
  new LogMessage()
    .setPipelineName('edges')
    .setJobId('23b9af7d5d4343219bc8e02ff4acd33a')
    .setUser(false)
    .setMessage('started datum task')
    .setTs(
      timestampFromObject({
        seconds: 1614126189,
        nanos: 0,
      }),
    ),
  new LogMessage()
    .setPipelineName('edges')
    .setJobId('23b9af7d5d4343219bc8e02ff4acd33a')
    .setUser(false)
    .setMessage('finished datum task')
    .setTs(
      timestampFromObject({
        seconds: 1614126190,
        nanos: 0,
      }),
    ),
  new LogMessage()
    .setPipelineName('montage')
    .setJobId('23b9af7d5d4343219bc8e02ff4acd33a')
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
    .setJobId('23b9af7d5d4343219bc8e02ff4acd33a')
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
    .setJobId('23b9af7d5d4343219bc8e02ff4acd33a')
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
    .setJobId('23b9af7d5d4343219bc8e02ff4acd33a')
    .setUser(false)
    .setMessage('finished datum task')
    .setTs(
      timestampFromObject({
        seconds: 1616533106,
        nanos: 0,
      }),
    ),
];

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

export const pipelineAndJobLogs: {[projectId: string]: LogMessage[]} = {
  '1': tutorial,
  '2': [],
  '3': [],
  '4': [],
  '5': [],
  '6': [],
  '7': [],
  default: [...tutorial],
};
