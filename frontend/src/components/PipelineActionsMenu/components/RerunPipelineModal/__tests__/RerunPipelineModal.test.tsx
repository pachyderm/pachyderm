import {render, screen} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {RerunPipelineRequest} from '@dash-frontend/api/pps';
import {RequestError} from '@dash-frontend/api/utils/error';
import {mockGetEnterpriseInfo} from '@dash-frontend/mocks';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import RerunPipelineModalComponent from '../RerunPipelineModal';
const server = setupServer();

describe('RerunPipelineModal', () => {
  const RerunPipelineModal = withContextProviders(({pipelineId}) => {
    return (
      <RerunPipelineModalComponent
        show={true}
        onHide={() => null}
        pipelineId={pipelineId}
      />
    );
  });

  beforeAll(() => {
    server.listen();
    window.history.replaceState('', '', '/lineage/default/pipelines/edges');
  });

  beforeEach(() => {
    server.use(mockGetEnterpriseInfo());
  });

  afterAll(() => server.close());

  it('should display an error message if mutation fails', async () => {
    server.use(
      rest.post<RerunPipelineRequest, Empty, RequestError>(
        '/api/pps_v2.API/RerunPipeline',
        async (_req, res, ctx) => {
          return res(
            ctx.status(400),
            ctx.json({
              message: 'unable to rerun pipeline',
              details: ['rerunPipeline'],
            }),
          );
        },
      ),
    );

    render(<RerunPipelineModal pipelineId="edges" />);

    await click(
      screen.getByRole('radio', {
        name: /process all datums/i,
      }),
    );

    await click(
      screen.getByRole('button', {
        name: /rerun pipeline/i,
      }),
    );

    expect(
      await screen.findByText('unable to rerun pipeline'),
    ).toBeInTheDocument();
  });
});
