export const UUID_WITHOUT_DASHES_REGEX = new RegExp(
  '^[0-9a-f]{12}4[0-9a-f]{19}$',
);

export const LONG_UUID_WITHOUT_DASHES_REGEX = new RegExp('^[0-9a-f]{64}$');

export const SHORT_OR_LONG_UUID = new RegExp(
  '^[0-9a-f]{12}4[0-9a-f]{19}$|^[0-9a-f]{64}$',
);
