import {TemplateMetadata} from '../../templates';
import parseTemplateMetadata from '../parseTemplateMetadata';

describe('parseTemplateMetadata', () => {
  const cases: [string, string, null | string | TemplateMetadata][] = [
    ['Null check', '', null],
    [
      'No comment block',
      `sdklfjsdf
      sdfklj
      sdfsdfk"
    `,
      null,
    ],
    [
      'Valid yaml',
      `
/*      
title: Test Title
description: "Test Description"
args:
  - name: name
    description: The name of the pipeline.
    type: string
*/
    `,
      {
        args: [
          {
            description: 'The name of the pipeline.',
            name: 'name',
            type: 'string',
          },
        ],
        description: 'Test Description',
        title: 'Test Title',
      },
    ],
    [
      'Optional description',
      `
/*
title: Test Title
args:
  - name: name
    description: The name of the pipeline.
    type: string
*/
        `,
      {
        args: [
          {
            description: 'The name of the pipeline.',
            name: 'name',
            type: 'string',
          },
        ],
        title: 'Test Title',
      },
    ],
  ];

  test.each(cases)('%p', (_, input, result) => {
    expect(parseTemplateMetadata(input)).toStrictEqual(result);
  });

  const errorCases: [string, string, string][] = [
    [
      'Invalid yaml comment',
      `/*      
      sdfklj
      erwkerjwkejr
      */"
    `,
      'Metadata missing title',
    ],
    [
      'Arguments not an array',
      `
/*      
description: "Test Description"
args:
    name: name
    description: The name of the pipeline.
    type: string
*/
    `,
      'Arguments array is not formatted corretly',
    ],
    [
      'No title',
      `
/*      
description: "Test Description"
args:
  - name: name
    description: The name of the pipeline.
    type: string
*/
    `,
      'Metadata missing title',
    ],
    [
      'No arguments',
      `
/*
title: Test Title
description: "Test Description"
*/
            `,
      'Metadata missing arguments',
    ],

    [
      'No argument name',
      `
/*
title: Test Title
description: "Test Description"
args:
  - description: The name of the pipeline.
    type: string
*/
            `,
      'Argument 0: missing name',
    ],
    [
      'No argument type',
      `
/*
title: Test Title
description: "Test Description"
args:
  - name: name
    description: The name of the pipeline.
*/
            `,
      'name: missing type',
    ],
    [
      'Bad argument type',
      `
/*
title: Test Title
description: "Test Description"
args:
  - name: name
    description: The name of the pipeline.
    type: xyz
*/
          `,
      'name: unsupported type xyz',
    ],
    [
      'Multiple errors in one template',
      `
/*
description: "Test Description"
args:
  - description: The name of the pipeline.
    type: xyz
  - description: The name of the pipeline.
    type: string
  - name: source
    description: The name of the pipeline.
    type: string
*/
          `,
      'Metadata missing title\nArgument 0: missing name\nArgument 0: unsupported type xyz\nArgument 1: missing name',
    ],
  ];

  test.each(errorCases)('%p', (_, input, result) => {
    expect(() => parseTemplateMetadata(input)).toThrow(result);
  });
});
