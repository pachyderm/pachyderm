import yaml from 'js-yaml';

import {TemplateMetadata} from '../templates/index';

const COMMENT_REGEX = /(?<=\*)[\s\S]*?(?=\*)/;
const regex = new RegExp(COMMENT_REGEX);

const parseTemplateMetadata = (
  templateString?: string,
): TemplateMetadata | null => {
  if (templateString) {
    const commentMatch = regex.exec(templateString)?.at(0);
    if (commentMatch) {
      const rawData = yaml.load(commentMatch) as TemplateMetadata;

      const errors = [];
      if (!rawData.title) {
        errors.push('Metadata missing title');
      }

      if (!rawData.args) {
        errors.push('Metadata missing arguments');
      } else if (Array.isArray(rawData.args)) {
        rawData.args.forEach((arg, index) => {
          if (!arg.name) {
            errors.push(`Argument ${index}: missing name`);
          }
          if (!arg.type) {
            errors.push(`${arg.name || `Argument ${index}`}: missing type`);
          } else if (!(arg.type === 'string' || arg.type === 'number')) {
            errors.push(
              `${arg.name || `Argument ${index}`}: unsupported type ${arg.type}`,
            );
          }
        });
      } else {
        errors.push('Arguments array is not formatted corretly');
      }

      if (errors.length !== 0) {
        throw new Error(errors.join('\n'));
      }

      return rawData;
    }
  }
  return null;
};

export default parseTemplateMetadata;
