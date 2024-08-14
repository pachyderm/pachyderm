import HuggingFaceDownloader from './huggingFaceDownloader.jsonnet?raw';
import SnowflakeTemplate from './snowflake.jsonnet?raw';

export interface TemplateArg {
  name: string;
  type: 'string' | 'number';
  description?: string;
  default?: string | number;
}

export interface TemplateMetadata {
  title: string;
  description?: string;
  args: TemplateArg[];
}

export interface Template {
  metadata: TemplateMetadata;
  template: string;
}
export const Templates: string[] = [SnowflakeTemplate, HuggingFaceDownloader];
