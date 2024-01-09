import {
  render,
  screen,
  waitForElementToBeRemoved,
  within,
} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  GetProjectDefaultsRequest,
  GetProjectDefaultsResponse,
  SetProjectDefaultsRequest,
  SetProjectDefaultsResponse,
} from '@dash-frontend/api/pps';
import {RequestError} from '@dash-frontend/api/utils/error';
import {mockGetEnterpriseInfoInactive} from '@dash-frontend/mocks';
import {
  mockProjectDefaultsSchema,
  mockProjectDefaultsSchema404,
  mockProjectDefaultsSchemaBadData,
} from '@dash-frontend/mocks/pachyderm';
import {withContextProviders, click, type} from '@dash-frontend/testHelpers';

import ProjectConfigComponent from '../ProjectConfig';

describe('ProjectConfig', () => {
  const server = setupServer();

  const ProjectConfig = withContextProviders(() => {
    return <ProjectConfigComponent />;
  });

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    window.history.replaceState({}, '', '/project/default/defaults');
    server.resetHandlers();
    server.use(mockProjectDefaultsSchema404);
    server.use(mockGetEnterpriseInfoInactive());
    server.use(
      rest.post<GetProjectDefaultsRequest, Empty, GetProjectDefaultsResponse>(
        '/api/pps_v2.API/GetProjectDefaults',
        (_req, res, ctx) => {
          return res(
            ctx.json({
              projectDefaultsJson: JSON.stringify(
                {
                  createPipelineRequest: {
                    resourceRequests: {cpu: 1, memory: '128Mi', disk: '1Gi'},
                  },
                },
                null,
                4,
              ),
            }),
          );
        },
      ),
    );
    server.use(
      rest.post<SetProjectDefaultsRequest, Empty, SetProjectDefaultsResponse>(
        '/api/pps_v2.API/SetProjectDefaults',
        (_req, res, ctx) => {
          return res(ctx.json({affectedPipelines: []}));
        },
      ),
    );
  });

  afterAll(() => server.close());

  it('should display an error if set project defaults with regenerate fails', async () => {
    server.use(
      rest.post<
        SetProjectDefaultsRequest,
        Empty,
        RequestError | SetProjectDefaultsResponse
      >('/api/pps_v2.API/SetProjectDefaults', async (req, res, ctx) => {
        const {regenerate} = await req.json();
        if (regenerate === true) {
          return res(
            ctx.status(400),
            ctx.json({
              code: 3,
              message: 'invalid project defaults JSON',
              details: [],
            }),
          );
        } else {
          return res(ctx.json({affectedPipelines: []}));
        }
      }),
    );

    render(<ProjectConfig />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(await screen.findByText(`"resourceRequests"`)).toBeInTheDocument();

    await click(screen.getByRole('button', {name: /Continue/i}));

    expect(screen.getByRole('button', {name: /Continue/i})).toBeDisabled();

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Unable to set project defaults',
    );

    await click(screen.getByRole('button', {name: 'See Full Error'}));

    expect(
      await within(screen.getByRole('dialog')).findByText(
        'invalid project defaults JSON',
      ),
    ).toBeInTheDocument();

    await click(
      within(screen.getByRole('dialog')).getByRole('button', {name: /back/i}),
    );

    // editing the document removes the error and enables continue again
    // insert syntactically correct JSON at the beginning of the editor
    await type(
      await screen.findByText(`"resourceRequests"`),
      '{arrowright}{backspace},',
    );
    await type(
      await screen.findByText(`"resourceRequests"`),
      '{{"invalidKey": "val",}{backspace},',
    );

    expect(screen.getByRole('button', {name: /Continue/i})).toBeEnabled();

    expect(screen.getByRole('alert')).toHaveTextContent(
      'You have unsaved changes',
    );
  });

  it('should display an error if project defaults dry run fails', async () => {
    server.use(
      rest.post<
        SetProjectDefaultsRequest,
        Empty,
        RequestError | SetProjectDefaultsResponse
      >('/api/pps_v2.API/SetProjectDefaults', async (req, res, ctx) => {
        const {dryRun} = await req.json();
        if (dryRun === true) {
          return res(
            ctx.status(400),
            ctx.json({
              code: 3,
              message: 'invalid JSON',
              details: [],
            }),
          );
        }
      }),
    );

    render(<ProjectConfig />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Unable to generate project defaults',
    );

    await click(screen.getByRole('button', {name: 'See Full Error'}));

    expect(
      await within(screen.getByRole('dialog')).findByText('invalid JSON'),
    ).toBeInTheDocument();
  });

  it('should show reasonable editor defaults if project config is {}', async () => {
    server.use(
      rest.post<GetProjectDefaultsRequest, Empty, GetProjectDefaultsResponse>(
        '/api/pps_v2.API/GetProjectDefaults',
        (_req, res, ctx) => {
          return res(
            ctx.json({
              projectDefaultsJson: '{}',
            }),
          );
        },
      ),
    );
    render(<ProjectConfig />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(
      await screen.findByText(`"createPipelineRequest"`),
    ).toBeInTheDocument();
  });

  it('should allow the user to return to the project page', async () => {
    render(<ProjectConfig />);

    await click(screen.getByRole('button', {name: /back/i}));

    expect(window.location.pathname).toBe('/lineage/default');
  });

  it('should disable the continue button if users have invalid JSON', async () => {
    render(<ProjectConfig />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    await type(await screen.findByText(`"resourceRequests"`), 'badjson');

    expect(screen.getByRole('alert')).toHaveTextContent('Invalid JSON');

    expect(screen.getByRole('button', {name: /continue/i})).toBeDisabled();
  });

  it('should display a warning modal if users try to leave with unsaved changes', async () => {
    render(<ProjectConfig />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    // insert syntactically correct JSON at the beginning of the editor
    await type(
      await screen.findByText(`"resourceRequests"`),
      '{arrowright}{backspace},',
    );
    await type(
      await screen.findByText(`"resourceRequests"`),
      '{{"invalidKey": "val",}{backspace},',
    );

    expect(screen.getByRole('alert')).toHaveTextContent(
      'You have unsaved changes',
    );

    await click(screen.getByRole('button', {name: /back/i}));

    expect(
      screen.getByRole('heading', {name: /unsaved changes/i}),
    ).toBeInTheDocument();

    await click(screen.getByRole('button', {name: /leave/i}));

    expect(window.location.pathname).toBe('/lineage/default');
  });

  it('should allow the user to prettify the document contents', async () => {
    render(<ProjectConfig />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
    await type(await screen.findByRole('textbox'), ' ');

    expect(screen.getByLabelText('code editor')).toHaveTextContent(
      '9123456789›⌄⌄⌄ { "createPipelineRequest": { "resourceRequests": { "cpu": 1, "memory": "128Mi", "disk": "1Gi" } }}',
    );

    await click(screen.getByRole('button', {name: /prettify/i}));

    expect(screen.getByLabelText('code editor')).toHaveTextContent(
      '9123456789›⌄⌄⌄{ "createPipelineRequest": { "resourceRequests": { "cpu": 1, "memory": "128Mi", "disk": "1Gi" } }}',
    );
  });

  it.each([
    [
      'should have autocomplete when the network call fails',
      mockProjectDefaultsSchema404,
    ],
    [
      'should have autocomplete when the network call is successful',
      mockProjectDefaultsSchema,
    ],
    [
      'should have autocomplete when given a bad schema',
      mockProjectDefaultsSchemaBadData,
    ],
  ])('%p', async (_, mock) => {
    server.resetHandlers();
    server.use(mockGetEnterpriseInfoInactive());
    server.use(
      rest.post<GetProjectDefaultsRequest, Empty, GetProjectDefaultsResponse>(
        '/api/pps_v2.API/GetProjectDefaults',
        (_req, res, ctx) => {
          return res(ctx.json({}));
        },
      ),
    );
    server.use(
      rest.post<SetProjectDefaultsRequest, Empty, SetProjectDefaultsResponse>(
        '/api/pps_v2.API/SetProjectDefaults',
        (_req, res, ctx) => {
          return res(ctx.json({affectedPipelines: []}));
        },
      ),
    );
    server.use(mock);
    render(<ProjectConfig />);
    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
    await type(screen.getByRole('textbox'), '{arrowright}"c');

    // Autocomplete text
    expect(
      await screen.findByText('createPipelineRequest'),
    ).toBeInTheDocument();
  });
});
