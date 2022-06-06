// eslint-disable-next-line @typescript-eslint/no-explicit-any
const removeGeneratedSuffixes = (object: any): any => {
  if (typeof object !== 'object') return object;
  if (Array.isArray(object)) return object.map(removeGeneratedSuffixes);
  const suffixes = ['List', 'Map'];
  return Object.keys(object as Record<string, unknown>).reduce(
    (memo: Record<string, unknown>, key: string) => {
      let nextKey = key;
      suffixes.forEach((suffix) => {
        if (key.endsWith(suffix)) {
          nextKey = key.slice(0, key.length - suffix.length);
        }
      });
      memo[nextKey] = removeGeneratedSuffixes(object[key]);
      return memo;
    },
    {},
  );
};

export default removeGeneratedSuffixes;
