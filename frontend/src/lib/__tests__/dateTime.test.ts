import {JobInfo} from '@dash-frontend/api/pps';
import {
  SECONDS_IN_MINUTE,
  SECONDS_IN_HOUR,
  SECONDS_IN_DAY,
  getDurationToNow,
  formatDurationFromSeconds,
  formatDurationFromSecondsToNow,
  getStandardDateFromUnixSeconds,
  getISOStringFromUnix,
  getUnixSecondsFromISOString,
  calculateJobTotalRuntime,
  calculateJobSetTotalRuntime,
} from '@dash-frontend/lib/dateTime';

import {InternalJobSet} from '../types';

describe('getDurationToNow', () => {
  it('returns the correct duration', () => {
    const unixSeconds = Date.now() / 1000 - SECONDS_IN_HOUR; // one hour ago
    const result = getDurationToNow(unixSeconds);
    expect(result).toBe('1 hour');
  });

  it('adds "ago" suffix when addSuffix parameter is true', () => {
    const unixSeconds = Date.now() / 1000 - SECONDS_IN_HOUR; // one hour ago
    const result = getDurationToNow(unixSeconds, true);
    expect(result).toBe('1 hour ago');
  });
});

describe('formatDurationFromSeconds', () => {
  it('returns N/A when input is undefined', () => {
    const result = formatDurationFromSeconds(undefined);
    expect(result).toBe('N/A');
  });

  it('returns 0 when input is 0', () => {
    const result = formatDurationFromSeconds(0);
    expect(result).toBe('0 s');
  });

  it('returns the correct format when duration longer than 1 hour', () => {
    const seconds =
      SECONDS_IN_DAY + 6 * SECONDS_IN_HOUR + 13 * SECONDS_IN_MINUTE + 35;
    const result = formatDurationFromSeconds(seconds);
    expect(result).toBe('1 d 6 h 13 mins');
  });

  it('returns the correct format when duration is shorter than 1 hour', () => {
    const seconds = 13 * SECONDS_IN_MINUTE + 35;
    const result = formatDurationFromSeconds(seconds);
    expect(result).toBe('13 mins 35 s');
  });

  it('returns the correct format when duration is shorter than 1 second', () => {
    const seconds = 0.1234567;
    const result = formatDurationFromSeconds(seconds);
    expect(result).toBe('0.1235 s');
  });
});

describe('formatDurationFromSecondsToNow', () => {
  it('returns the correct format compared to current time', () => {
    const pastTimeSeconds = Date.now() / 1000 - 2 * SECONDS_IN_HOUR; // two hours ago
    const result = formatDurationFromSecondsToNow(pastTimeSeconds);
    expect(result).toBe('2 h 0 min');
  });
});

describe('getStandardDate', () => {
  it('returns the date in the correct format', () => {
    const unixSeconds = Date.now() / 1000 - 2 * SECONDS_IN_HOUR; // two hours ago
    const result = getStandardDateFromUnixSeconds(unixSeconds);
    // This regular expression matches a string that has the following pattern:
    // A capital letter followed by two lowercase letters, which represents an abbreviation of a month name (e.g. Jan, Feb, Mar, etc.).
    // A space character.
    // One or two digits, representing the day of the month (e.g. 1, 12, 31).
    // A comma followed by a space character.
    // Four digits representing the year (e.g. 2022, 2023).
    // A semicolon followed by a space character.
    // One or two digits representing the hour in 24-hour format (e.g. 0, 1, 23).
    // A colon character.
    // Two digits representing the minute (e.g. 00, 30).
    expect(result).toMatch(/[A-Z][a-z]{2} \d{1,2}, \d{4}; \d{1,2}:\d{2}/);
  });
});

describe('getISOStringFromUnix', () => {
  it('returns an ISO string given unix seconds', () => {
    expect(getISOStringFromUnix(1690221506)).toBe('2023-07-24T17:58:26Z');
  });
});

describe('getTimeFromISOString', () => {
  it('returns unix seconds given an ISO string', () => {
    expect(getUnixSecondsFromISOString('2023-07-24T17:58:38Z')).toBe(
      1690221518,
    );
    expect(getUnixSecondsFromISOString('2023-09-27T08:20:55.707925Z')).toBe(
      1695802855,
    );
  });
});

describe('calculateJobTotalRuntime', () => {
  const pastTimeSeconds = Date.now() / 1000 - 2 * SECONDS_IN_HOUR; // two hours ago
  const cases: [Pick<JobInfo, 'finished' | 'started'>, string][] = [
    [{started: undefined, finished: undefined}, 'N/A'],
    [
      {started: getISOStringFromUnix(pastTimeSeconds), finished: undefined},
      'N/A',
    ],
    [{started: undefined, finished: '2023-07-12T01:49:09.000Z'}, 'N/A'],
    [
      {
        started: '2023-07-12T01:32:29.000Z',
        finished: '2023-07-12T01:49:09.000Z',
      },
      '16 mins 40 s',
    ],
  ];
  test.each(cases)('%p returns %p', (input, result) => {
    expect(calculateJobTotalRuntime(input)).toContain(result);
  });
});

describe('calculateJobSetTotalRuntime', () => {
  const pastTimeSeconds = Date.now() / 1000 - 2 * SECONDS_IN_HOUR; // two hours ago
  const cases: [Pick<InternalJobSet, 'finished' | 'started'>, string][] = [
    [{started: undefined, finished: undefined}, 'In Progress'],
    [
      {started: getISOStringFromUnix(pastTimeSeconds), finished: undefined},
      '2 h 0 min - In Progress',
    ],
    [{started: undefined, finished: '2023-07-12T01:49:09.000Z'}, 'In Progress'],
    [
      {
        started: '2023-07-12T01:32:29.000Z',
        finished: '2023-07-12T01:49:09.000Z',
      },
      '16 mins 40 s',
    ],
  ];
  test.each(cases)('%p returns %p', (input, result) => {
    expect(calculateJobSetTotalRuntime(input)).toContain(result);
  });
});
