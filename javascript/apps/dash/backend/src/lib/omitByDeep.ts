const omitByDeep = (
  value: unknown,
  omitByFn: (val: unknown, key?: string) => boolean,
): unknown => {
  if (Array.isArray(value) && !omitByFn(value, undefined)) {
    return value.map((i) => omitByDeep(i, omitByFn));
  } else if (typeof value === 'object' && value !== null) {
    return Object.entries(value).reduce((newObject, [key, val]) => {
      if (omitByFn(val, key)) return newObject;

      return Object.assign(
        newObject,
        // TS doesn't allow for indexing here, even though we've previously
        // checked that value is an object (and key is a key on value).
        {[key]: omitByDeep((value as Record<string, unknown>)[key], omitByFn)},
      );
    }, {});
  }

  return value;
};

export default omitByDeep;
