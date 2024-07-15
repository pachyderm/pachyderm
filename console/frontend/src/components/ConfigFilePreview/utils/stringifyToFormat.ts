import {dump} from 'js-yaml';

export enum Format {
  JSON = 'JSON',
  YAML = 'YAML',
}

const stringifyToFormat = (
  obj: Record<string, unknown> | [],
  format: Format,
) => {
  return format === Format.YAML ? dump(obj) : JSON.stringify(obj, null, 2);
};

export default stringifyToFormat;
