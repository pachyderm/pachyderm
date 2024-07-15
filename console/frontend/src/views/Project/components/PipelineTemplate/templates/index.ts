// import EdgesTemplate from './edges.jsonnet?raw';

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

//Import static templates with the '?raw' flag and add them here
export const Templates: string[] = [];
