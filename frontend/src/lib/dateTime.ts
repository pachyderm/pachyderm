import {
  formatDistanceToNowStrict,
  fromUnixTime,
  format,
  parseISO,
  formatISO,
} from 'date-fns';

import {JobInfo} from '@dash-frontend/api/pps';

import {InternalJobSet} from './types';

export const STANDARD_DATE_FORMAT = 'MMM d, yyyy; H:mm';

export const SECONDS_IN_MINUTE = 60;
export const SECONDS_IN_HOUR = 60 * SECONDS_IN_MINUTE;
export const SECONDS_IN_DAY = 24 * SECONDS_IN_HOUR;

// from date-fns // Return the distance between the given date and now in words.
export const getDurationToNow = (unixSeconds: number, addSuffix = false) => {
  return formatDistanceToNowStrict(fromUnixTime(unixSeconds), {addSuffix});
};

// ex: 1 day 6 h 13 mins, 6 h 13 mins 35s
export const formatDurationFromSeconds = (seconds?: number) => {
  if (!seconds && seconds !== 0) return 'N/A';
  if (seconds > 0 && seconds < 1) return seconds.toFixed(4) + ' s';

  let secondsLeft = seconds;
  let formatString = '';

  if (secondsLeft >= SECONDS_IN_DAY) {
    const totalDays = Math.floor(secondsLeft / SECONDS_IN_DAY);
    formatString += `${totalDays} d `;
    secondsLeft -= totalDays * SECONDS_IN_DAY;
  }

  if (secondsLeft >= SECONDS_IN_HOUR) {
    const totalHours = Math.floor(secondsLeft / SECONDS_IN_HOUR);
    formatString += `${totalHours} h `;
    secondsLeft -= totalHours * SECONDS_IN_HOUR;
  }

  if (seconds >= SECONDS_IN_MINUTE) {
    const totalMinutes = Math.floor(secondsLeft / SECONDS_IN_MINUTE);
    formatString += `${totalMinutes} ${totalMinutes > 1 ? 'mins' : 'min'} `;
    secondsLeft -= totalMinutes * SECONDS_IN_MINUTE;
  }

  // only show seconds if the value is smaller than 1 hour
  if (seconds < SECONDS_IN_HOUR) {
    formatString += `${secondsLeft.toFixed(0)} s`;
  }

  return formatString.trim();
};

// same format as above, but using the seconds between a point in time
// in the past and the current time
export const formatDurationFromSecondsToNow = (pastTimeSeconds: number) => {
  const unixNow = Math.ceil(Date.now() / 1000);
  return formatDurationFromSeconds(unixNow - pastTimeSeconds);
};

/**
 * Converts from unix seconds to our display format.
 *
 * Example:
 * 1690221506 => Mar 20, 2023; 13:49
 */
export const getStandardDateFromUnixSeconds = (unixSeconds: number) => {
  if (unixSeconds === -1) return 'N/A';

  return format(fromUnixTime(unixSeconds), STANDARD_DATE_FORMAT);
};

/**
 * Converts from unix seconds to ISO
 *
 * Example:
 * 1690221506 => 2023-09-27T08:20:55.707925Z
 **/
export const getISOStringFromUnix = (unixSeconds: number) => {
  if (unixSeconds === -1) return undefined;

  return formatISO(fromUnixTime(unixSeconds));
};

/**
 * Converts ISO string to our display format.
 *
 * Example:
 * 2023-09-27T08:20:55.707925Z => Sep 09, 2023; 4:20
 **/
export const getStandardDateFromISOString = (date?: string) => {
  if (!date) return 'N/A';

  return format(parseISO(date), STANDARD_DATE_FORMAT);
};

/**
 * Converts ISO string to unix seconds
 *
 * Examples:
 *
 * 2023-09-27T08:20:55Z        => 1695802855
 *
 * 2023-09-27T08:20:55.707925Z => 1695802855
 */
export const getUnixSecondsFromISOString = (date?: string) => {
  return date ? Math.floor(parseISO(date).getTime() / 1000) : 0;
};

export const getDurationToNowFromISOString = (
  date: string,
  addSuffix = false,
) => {
  return formatDistanceToNowStrict(parseISO(date), {addSuffix});
};

/**
 * Used to calculate the total runtime of a job.
 * Response is a formatted string representing
 * the calculated runtime.
 */
export const calculateJobTotalRuntime = (
  jobInfo?: Pick<JobInfo, 'finished' | 'started'>,
  nullCase = 'N/A',
) =>
  jobInfo?.finished && jobInfo?.started
    ? formatDurationFromSeconds(
        getUnixSecondsFromISOString(jobInfo.finished) -
          getUnixSecondsFromISOString(jobInfo.started),
      )
    : nullCase;

/**
 * Used to calculate the total runtime of an internal jobset.
 * Response is a formatted string representing
 * the calculated runtime.
 */
export const calculateJobSetTotalRuntime = (
  internalJobSet?: Pick<InternalJobSet, 'finished' | 'started'>,
) => {
  const startedAt = internalJobSet?.started;
  const finishedAt = internalJobSet?.finished;
  if (finishedAt && startedAt) {
    return formatDurationFromSeconds(
      getUnixSecondsFromISOString(finishedAt) -
        getUnixSecondsFromISOString(startedAt),
    );
  }
  if (startedAt) {
    return `${formatDurationFromSecondsToNow(
      getUnixSecondsFromISOString(startedAt),
    )} - In Progress`;
  }
  return 'In Progress';
};
