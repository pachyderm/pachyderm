import {formatDistanceToNowStrict, fromUnixTime, format} from 'date-fns';

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
    formatString += `${secondsLeft} s`;
  }

  return formatString.trim();
};

// same format as above, but using the seconds between a point in time
// in the past and the current time
export const formatDurationFromSecondsToNow = (pastTimeSeconds: number) => {
  const unixNow = Math.ceil(Date.now() / 1000);
  return formatDurationFromSeconds(unixNow - pastTimeSeconds);
};

// Mar 20, 2023; 13:49
export const getStandardDate = (unixSeconds: number) => {
  if (unixSeconds === -1) return '-';
  return format(fromUnixTime(unixSeconds), STANDARD_DATE_FORMAT);
};
