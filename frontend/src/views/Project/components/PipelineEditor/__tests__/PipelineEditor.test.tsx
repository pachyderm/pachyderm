import {
  mockGetClusterDefaultsQuery,
  mockCreatePipelineV2Mutation,
  mockProjectQuery,
} from '@graphqlTypes';
import {
  render,
  screen,
  waitForElementToBeRemoved,
  within,
} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockGetVersionInfo,
  mockEmptyProjectDetails,
  mockGetMontagePipeline,
} from '@dash-frontend/mocks';
import {
  mockCreatePipelineRequestSchema,
  mockCreatePipelineRequestSchema404,
  mockCreatePipelineRequestSchemaBadData,
} from '@dash-frontend/mocks/pachyderm';
import {
  withContextProviders,
  click,
  type,
  hover,
} from '@dash-frontend/testHelpers';

import PipelineEditorComponent from '../PipelineEditor';

describe('PipelineEditor', () => {
  const server = setupServer();

  const PipelineEditor = withContextProviders(() => {
    return <PipelineEditorComponent />;
  });

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    window.history.replaceState({}, '', '/lineage/default/create/pipeline');
    server.resetHandlers();
    server.use(mockGetVersionInfo());
    server.use(mockEmptyProjectDetails());
    server.use(mockCreatePipelineRequestSchema);
    server.use(mockGetMontagePipeline());
    server.use(
      mockProjectQuery((_req, res, ctx) => {
        return res(
          ctx.data({
            project: {
              id: 'default',
              description: '',
              createdAt: {nanos: 0, seconds: 0},
            },
          }),
        );
      }),
    );
    server.use(
      mockGetClusterDefaultsQuery((_req, res, ctx) => {
        return res(
          ctx.data({
            getClusterDefaults: {
              clusterDefaultsJson:
                '{\n    "createPipelineRequest": {\n        "resourceRequests": {\n            "cpu": 1,\n            "memory": "128Mi",\n            "disk": "1Gi"\n        }\n    }\n}',
            },
          }),
        );
      }),
    );
    server.use(
      mockCreatePipelineV2Mutation((_req, res, ctx) => {
        return res(
          ctx.data({
            createPipelineV2: {
              effectiveCreatePipelineRequestJson: '{\n  "effectiveSpec": {}\n}',
            },
          }),
        );
      }),
    );
  });

  afterAll(() => server.close());

  it('should display an error if create pipeline dry run fails', async () => {
    server.use(
      mockCreatePipelineV2Mutation((req, res, ctx) => {
        if (req.variables.args.dryRun) {
          return res(
            ctx.errors([
              {
                message: 'Invalid JSON',
                path: ['createPipelineV2'],
              },
            ]),
          );
        }
      }),
    );

    render(<PipelineEditor />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Unable to generate an effective pipeline spec',
    );

    await click(screen.getByRole('button', {name: 'See Full Error'}));

    expect(
      await within(screen.getByRole('dialog')).findByText('Invalid JSON'),
    ).toBeInTheDocument();
  });

  it('should display an error if create pipeline fails', async () => {
    server.use(
      mockCreatePipelineV2Mutation((req, res, ctx) => {
        if (!req.variables.args.dryRun) {
          return res(
            ctx.errors([
              {
                message: 'Invalid JSON',
                path: ['createPipelineV2'],
              },
            ]),
          );
        } else {
          return res(
            ctx.data({
              createPipelineV2: {
                effectiveCreatePipelineRequestJson:
                  '{\n  "effectiveSpec": {}\n}',
              },
            }),
          );
        }
      }),
    );

    render(<PipelineEditor />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(await screen.findByText(`"transform"`)).toBeInTheDocument();

    await click(screen.getByRole('button', {name: /create pipeline/i}));

    expect(
      screen.getByRole('button', {name: /create pipeline/i}),
    ).toBeDisabled();

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Unable to create pipeline',
    );

    // editing the document removes the error and triggers another dry run
    await type(await screen.findByText(`"transform"`), ' ');

    expect(
      screen.getByRole('button', {name: /create pipeline/i}),
    ).toBeEnabled();
  });

  it('should display cluster defaults and effective spec tabs', async () => {
    render(<PipelineEditor />);

    expect(await screen.findByText(`"effectiveSpec"`)).toBeVisible();
    expect(
      await screen.findByText(`"createPipelineRequest"`),
    ).not.toBeVisible();

    await click(screen.getByRole('tab', {name: /cluster defaults/i}));

    expect(await screen.findByText(`"effectiveSpec"`)).not.toBeVisible();
    expect(await screen.findByText(`"createPipelineRequest"`)).toBeVisible();
  });

  it('should allow the user to return to the lineage page', async () => {
    render(<PipelineEditor />);

    await click(screen.getByRole('button', {name: /cancel/i}));

    expect(window.location.pathname).toBe('/lineage/default');
  });

  it('should disable the create button if users have invalid JSON', async () => {
    render(<PipelineEditor />);

    await type(await screen.findByText(`"transform"`), 'badjson');

    expect(await screen.findByRole('alert')).toHaveTextContent('Invalid JSON');

    expect(
      screen.getByRole('button', {name: /create pipeline/i}),
    ).toBeDisabled();
  });

  it('should set the pipeline project based on url', async () => {
    render(<PipelineEditor />);

    await type(await screen.findByText(`"transform"`), ' ');

    expect(screen.getByLabelText('code editor')).toHaveTextContent(
      /"project": { "name": "default" }/,
    );
  });

  it('should allow the user to prettify the editor contents', async () => {
    render(<PipelineEditor />);

    await type(await screen.findByText(`"transform"`), ' ');

    expect(screen.getByLabelText('code editor')).toHaveTextContent(
      /9123456789›⌄⌄⌄ { "pipeline": { "name": /,
    );

    await click(screen.getByRole('button', {name: /prettify/i}));

    expect(screen.getByLabelText('code editor')).toHaveTextContent(
      /9123456789›⌄⌄⌄{ "pipeline": { "name": /,
    );
  });

  it('should allow the user to view single or split view tabs', async () => {
    render(<PipelineEditor />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(
      await screen.findByRole('tabpanel', {name: 'Effective Spec'}),
    ).toBeInTheDocument();

    expect(await screen.findAllByRole('tablist')).toHaveLength(2);
    expect(await screen.findAllByRole('textbox')).toHaveLength(2);

    await click(screen.getByRole('button', {name: 'view single window'}));

    expect(await screen.findAllByRole('tablist')).toHaveLength(1);
    expect(await screen.findAllByRole('textbox')).toHaveLength(1);

    await click(
      screen.getByRole('button', {name: 'view windows side by side'}),
    );

    expect(await screen.findAllByRole('tablist')).toHaveLength(2);
    expect(await screen.findAllByRole('textbox')).toHaveLength(2);
  });

  it.each([
    [
      'should have autocomplete when the network call fails',
      mockCreatePipelineRequestSchema404,
    ],
    [
      'should have autocomplete when the network call is successful',
      mockCreatePipelineRequestSchema,
    ],
    [
      'should have autocomplete when given a bad schema',
      mockCreatePipelineRequestSchemaBadData,
    ],
  ])('%s', async (_, mock) => {
    server.use(mock);
    render(<PipelineEditor />);
    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
    await type(
      await within(screen.getByTestId('CodeEditor__createPipeline')).findByRole(
        'textbox',
      ),
      '{arrowright}"i',
    );

    // Autocomplete text
    expect(await screen.findByText('input')).toBeInTheDocument();
  });

  it('should have the Effective Spec Dynamic Icons extension working', async () => {
    server.use(
      mockGetClusterDefaultsQuery((_req, res, ctx) => {
        return res(
          ctx.data({
            getClusterDefaults: {
              clusterDefaultsJson: JSON.stringify(
                {
                  createPipelineRequest: {
                    resourceRequests: {cpu: 1, memory: '128Mi', disk: '1Gi'},
                    description: 'original',
                  },
                },
                null,
                4,
              ),
            },
          }),
        );
      }),
    );
    server.use(
      mockCreatePipelineV2Mutation((_req, res, ctx) => {
        return res(
          ctx.data({
            createPipelineV2: {
              effectiveCreatePipelineRequestJson: JSON.stringify({
                pipeline: {
                  name: 'whatever',
                },
                transform: {cmd: ['python']},
                description: 'override',
              }),
            },
          }),
        );
      }),
    );

    render(<PipelineEditor />);
    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    await type(
      await within(screen.getByTestId('CodeEditor__createPipeline')).findByText(
        /"transform"/i,
      ),
      `${'{arrowright}'.repeat(500)}${'{arrowleft}'.repeat(
        3,
      )}"cmd": [["python"]{arrowright}{arrowright},"description": "override"`,
    );

    await click(
      screen.getByRole('button', {
        name: /prettify/i,
      }),
    );

    // No strikethrough for arrays
    // No strikethrough for values that match in the user spec and defaults
    expect(
      screen.getAllByTestId('dynamicEffectiveSpecDecorations__userAvatarSVG'),
    ).toHaveLength(3);
    const foo = screen.getAllByTestId('overWrittenValue');
    expect(foo).toHaveLength(1);
    expect(foo[0]).toHaveTextContent('original');

    // Tooltip
    await hover(
      screen.getAllByTestId(
        'dynamicEffectiveSpecDecorations__userAvatarSVG',
      )[0],
    );

    expect(
      await screen.findByRole('tooltip', {
        name: /user provided values that\nmay have overwritten a default/i,
      }),
    ).toBeInTheDocument();
  });

  describe('Update Pipeline', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/lineage/default/update/pipeline/montage',
      );
    });

    it('should allow users to return to the pipeline page', async () => {
      render(<PipelineEditor />);

      await click(screen.getByRole('button', {name: /cancel/i}));

      expect(window.location.pathname).toBe(
        '/lineage/default/pipelines/montage',
      );
    });

    it('should load the current user spec and show a confirmation modal', async () => {
      render(<PipelineEditor />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      // fixes flakyness
      await new Promise((resolve) => setTimeout(resolve, 100));

      expect(await screen.findByText(`"montage"`)).toBeInTheDocument();

      await click(screen.getByRole('button', {name: /update pipeline/i}));

      expect(
        within(await screen.findByRole('dialog')).getByRole('heading', {
          name: 'Update Pipeline',
        }),
      ).toBeInTheDocument();
    });

    it('should close the confirmation modal and display an error if update fails', async () => {
      server.use(
        mockCreatePipelineV2Mutation((req, res, ctx) => {
          if (!req.variables.args.dryRun) {
            return res(
              ctx.errors([
                {
                  message: 'Invalid JSON',
                  path: ['createPipelineV2'],
                },
              ]),
            );
          } else {
            return res(
              ctx.data({
                createPipelineV2: {
                  effectiveCreatePipelineRequestJson:
                    '{\n  "effectiveSpec": {}\n}',
                },
              }),
            );
          }
        }),
      );

      render(<PipelineEditor />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      await click(screen.getByRole('button', {name: /update pipeline/i}));

      const dialog = await screen.findByRole('dialog');
      await click(
        await within(dialog).findByRole('button', {
          name: /update pipeline/i,
        }),
      );

      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();

      expect(screen.getByRole('alert')).toHaveTextContent(
        'Unable to update pipeline',
      );
    });

    it('should warn the user if the pipeline name or project was changed', async () => {
      server.use(
        mockCreatePipelineV2Mutation((_req, res, ctx) => {
          return res(
            ctx.data({
              createPipelineV2: {
                effectiveCreatePipelineRequestJson: `{
                      "pipeline": {
                        "project": {
                          "name": "notdefault"
                        },
                        "name": "notmontage"
                      }
                    }`,
              },
            }),
          );
        }),
      );

      render(<PipelineEditor />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      expect(await screen.findByText(`"notdefault"`)).toBeInTheDocument();

      await click(screen.getByRole('button', {name: /update pipeline/i}));

      const dialog = await screen.findByRole('dialog');

      expect(within(dialog).getAllByText('notdefault')).toHaveLength(2);

      expect(within(dialog).getByText('notmontage')).toBeInTheDocument();

      await click(screen.getByRole('button', {name: /continue anyway/i}));

      expect(
        within(await screen.findByRole('dialog')).getByRole('heading', {
          name: 'Update Pipeline',
        }),
      ).toBeInTheDocument();
    });
  });
});
