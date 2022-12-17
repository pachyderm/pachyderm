const NANOS = 1e9;

export const timestampToSeconds = (timestamp?: {
  seconds?: number;
  nanos?: number;
}) => {
  if (!timestamp) return 0;
  if (!timestamp.nanos) return timestamp.seconds || 0;
  return +Number((timestamp.seconds || 0) + timestamp.nanos / NANOS).toFixed(2);
};
