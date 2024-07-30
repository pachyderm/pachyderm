//return a jq filter wrapped in select()
export const jqSelect = (filter: string) => {
  return `select(${filter})`;
};

// return a filter that checks if the selector is in the list of given values
export const jqIn = (selector: string, values: (string | number | null)[]) => {
  return `${selector} | IN(${values
    .map((val) => (typeof val === 'string' ? `"${val}"` : val))
    .join(', ')})`;
};

// return a series of piped filters, or a single filter if there is only one
export const jqCombine = (root: string, ...filters: string[]) => {
  if (root === '' && filters.length === 1) {
    return filters[0];
  }
  return [root, ...filters].join(' | ');
};

export const jqAnd = (...args: string[]) => {
  if (args.length === 1) return args[0];

  return args.join(' and ');
};
