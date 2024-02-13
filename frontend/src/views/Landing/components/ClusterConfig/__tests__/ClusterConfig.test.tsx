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
  GetClusterDefaultsRequest,
  GetClusterDefaultsResponse,
  SetClusterDefaultsRequest,
  SetClusterDefaultsResponse,
} from '@dash-frontend/api/pps';
import {RequestError} from '@dash-frontend/api/utils/error';
import {
  mockGetEnterpriseInfoInactive,
  mockGetVersionInfo,
} from '@dash-frontend/mocks';
import {
  mockClusterDefaultsSchema,
  mockClusterDefaultsSchema404,
  mockClusterDefaultsSchemaBadData,
} from '@dash-frontend/mocks/pachyderm';
import {withContextProviders, click, type} from '@dash-frontend/testHelpers';

import ClusterConfigComponent from '../ClusterConfig';

describe('ClusterConfig', () => {
  const server = setupServer();

  const ClusterConfig = withContextProviders(() => {
    return <ClusterConfigComponent />;
  });

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    window.history.replaceState({}, '', '/cluster/defaults');
    server.resetHandlers();
    server.use(mockGetVersionInfo());
    server.use(mockClusterDefaultsSchema404);
    server.use(mockGetEnterpriseInfoInactive());
    server.use(
      rest.post<GetClusterDefaultsRequest, Empty, GetClusterDefaultsResponse>(
        '/api/pps_v2.API/GetClusterDefaults',
        (_req, res, ctx) => {
          return res(
            ctx.json({
              clusterDefaultsJson: JSON.stringify(
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
      rest.post<SetClusterDefaultsRequest, Empty, SetClusterDefaultsResponse>(
        '/api/pps_v2.API/SetClusterDefaults',
        (_req, res, ctx) => {
          return res(ctx.json({affectedPipelines: []}));
        },
      ),
    );
  });

  afterAll(() => server.close());

  it('should display an error if set cluster defaults with regenerate fails', async () => {
    server.use(
      rest.post<
        SetClusterDefaultsRequest,
        Empty,
        RequestError | SetClusterDefaultsResponse
      >('/api/pps_v2.API/SetClusterDefaults', async (req, res, ctx) => {
        const {regenerate} = await req.json();
        if (regenerate === true) {
          return res(
            ctx.status(400),
            ctx.json({
              code: 3,
              message: 'invalid cluster defaults JSON',
              details: [],
            }),
          );
        } else {
          return res(ctx.json({affectedPipelines: []}));
        }
      }),
    );

    render(<ClusterConfig />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(await screen.findByText(`"resourceRequests"`)).toBeInTheDocument();

    await click(screen.getByRole('button', {name: /Continue/i}));

    expect(screen.getByRole('button', {name: /Continue/i})).toBeDisabled();

    expect(screen.getByRole('alert')).toHaveTextContent(
      'Unable to set cluster defaults',
    );

    await click(screen.getByRole('button', {name: 'See Full Error'}));

    expect(
      await within(screen.getByRole('dialog')).findByText(
        'invalid cluster defaults JSON',
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

  it('should display an error if create pipeline dry run fails', async () => {
    server.use(
      rest.post<
        SetClusterDefaultsRequest,
        Empty,
        RequestError | SetClusterDefaultsResponse
      >('/api/pps_v2.API/SetClusterDefaults', async (req, res, ctx) => {
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

    render(<ClusterConfig />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Unable to generate cluster defaults',
    );

    await click(screen.getByRole('button', {name: 'See Full Error'}));

    expect(
      await within(screen.getByRole('dialog')).findByText('invalid JSON'),
    ).toBeInTheDocument();
  });

  it('should allow the user to return to the landing page', async () => {
    render(<ClusterConfig />);

    await click(screen.getByRole('button', {name: /back/i}));

    expect(window.location.pathname).toBe('/');
  });

  it('should disable the continue button if users have invalid JSON', async () => {
    render(<ClusterConfig />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    await type(await screen.findByText(`"resourceRequests"`), 'badjson');

    expect(screen.getByRole('alert')).toHaveTextContent('Invalid JSON');

    expect(screen.getByRole('button', {name: /continue/i})).toBeDisabled();
  });

  it('should display a warning modal if users try to leave with unsaved changes', async () => {
    render(<ClusterConfig />);

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

    expect(window.location.pathname).toBe('/');
  });

  it('should allow the user to prettify the document contents', async () => {
    render(<ClusterConfig />);

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
      mockClusterDefaultsSchema404,
    ],
    [
      'should have autocomplete when the network call is successful',
      mockClusterDefaultsSchema,
    ],
    [
      'should have autocomplete when given a bad schema',
      mockClusterDefaultsSchemaBadData,
    ],
  ])('%p', async (_, mock) => {
    server.resetHandlers();
    server.use(mockGetVersionInfo());
    server.use(mockGetEnterpriseInfoInactive());
    server.use(
      rest.post<GetClusterDefaultsRequest, Empty, GetClusterDefaultsResponse>(
        '/api/pps_v2.API/GetClusterDefaults',
        (_req, res, ctx) => {
          return res(ctx.json({}));
        },
      ),
    );
    server.use(
      rest.post<SetClusterDefaultsRequest, Empty, SetClusterDefaultsResponse>(
        '/api/pps_v2.API/SetClusterDefaults',
        (_req, res, ctx) => {
          return res(ctx.json({affectedPipelines: []}));
        },
      ),
    );
    server.use(mock);
    render(<ClusterConfig />);
    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
    await type(screen.getByRole('textbox'), '{arrowright}"c');

    // Autocomplete text
    expect(
      await screen.findByText('createPipelineRequest'),
    ).toBeInTheDocument();
  });
});
