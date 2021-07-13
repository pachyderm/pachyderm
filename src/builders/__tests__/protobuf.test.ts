import {durationFromObject, timestampFromObject} from '../protobuf';

describe('grpc/buidlers/protobuf', () => {
  it('should create Timestamp from an object', () => {
    const timestamp = timestampFromObject({
      seconds: 1614736724,
      nanos: 344218476,
    });

    expect(timestamp.getSeconds()).toBe(1614736724);
    expect(timestamp.getNanos()).toBe(344218476);
  });

  it('should create Duration from an object', () => {
    const duration = durationFromObject({
      seconds: 1614736724,
      nanos: 344218476,
    });

    expect(duration.getSeconds()).toBe(1614736724);
    expect(duration.getNanos()).toBe(344218476);
  });
});
