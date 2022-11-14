import {Duration} from 'google-protobuf/google/protobuf/duration_pb';
import {Timestamp} from 'google-protobuf/google/protobuf/timestamp_pb';

export type TimestampObject = {
  seconds: Timestamp.AsObject['seconds'];
  nanos: Timestamp.AsObject['nanos'];
};

export type DurationObject = {
  seconds: Duration.AsObject['seconds'];
  nanos: Duration.AsObject['nanos'];
};

export const timestampFromObject = ({seconds, nanos}: TimestampObject) => {
  const timestamp = new Timestamp();
  timestamp.setSeconds(seconds);
  timestamp.setNanos(nanos);

  return timestamp;
};

export const durationFromObject = ({seconds, nanos}: DurationObject) => {
  const duration = new Duration();
  duration.setSeconds(seconds);
  duration.setNanos(nanos);

  return duration;
};
