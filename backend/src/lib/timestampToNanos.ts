import {Timestamp} from 'google-protobuf/google/protobuf/timestamp_pb';

const timestampToNanos = (timestamp: Timestamp.AsObject) => {
  return timestamp.seconds * 1e9 + timestamp.nanos;
};

export default timestampToNanos;
