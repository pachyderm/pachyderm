import {_deriveDuration, _processDuration} from '../useDatumDetails';

describe('_deriveDuration', () => {
  it.each([
    [
      [
        {
          seconds: 1,
          nanos: 123123123,
        },
      ],
      '1.12 seconds',
    ],
    [
      [
        {
          seconds: 1,
          nanos: 123123123,
        },
        {
          seconds: 3,
          nanos: 3123123,
        },
        {
          seconds: 2,
          nanos: 23123123,
        },
      ],
      '6.15 seconds',
    ],
    [[], ''],
  ])('calling _deriveDuration with %j should equal %s', (input, expected) => {
    expect(_deriveDuration(input)).toEqual(expected);
  });
});

describe('_processDuration', () => {
  it.each([
    [
      {
        seconds: 1,
        nanos: 123123123,
      },
      '1.12 seconds',
    ],
    [{seconds: 0, nanos: 123123123}, '0.12 seconds'],
    [{seconds: 0, nanos: 0}, '< 0.01 seconds'],
    [null, 'N/A'],
    [undefined, 'N/A'],
  ])('calling _processDuration with %j should equal %s', (input, expected) => {
    expect(_processDuration(input)).toEqual(expected);
  });
});
