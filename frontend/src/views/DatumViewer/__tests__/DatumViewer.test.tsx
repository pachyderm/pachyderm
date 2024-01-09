import {
  render,
  waitFor,
  within,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  DatumInfo,
  DatumState,
  InspectDatumRequest,
  InspectJobRequest,
  JobInfo,
} from '@dash-frontend/api/pps';
import {
  mockGetEnterpriseInfoInactive,
  mockGetSpoutPipeline,
  mockGetServicePipeline,
  mockEmptyGetAuthorize,
  mockGetVersionInfo,
  mockGetMontagePipeline,
  mockInspectJobMontage5C,
  buildJob,
  mockGetMontageJob5C,
  mockGetMontageJobs,
  mockGetJob5CDatum05,
  mockGetJob5CDatums,
  mockGetJob5CDatumCH,
  buildDatum,
} from '@dash-frontend/mocks';
import {mockEmptyGetLogs} from '@dash-frontend/mocks/logs';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import {
  PipelineDatumViewer as PipelineDatumViewerComponent,
  JobDatumViewer as JobDatumViewerComponent,
} from '../DatumViewer';

const PipelineDatumViewer = withContextProviders(() => {
  return <PipelineDatumViewerComponent />;
});
const JobDatumViewer = withContextProviders(() => {
  return <JobDatumViewerComponent />;
});

describe('Datum Viewer', () => {
  const server = setupServer();

  beforeAll(() => server.listen());

  beforeEach(() => {
    server.resetHandlers();
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetVersionInfo());
    server.use(mockGetMontagePipeline());
    server.use(mockGetMontageJob5C());
    server.use(mockInspectJobMontage5C());
    server.use(mockGetMontageJobs());
    server.use(mockGetJob5CDatums());
    server.use(mockGetJob5CDatum05());
    server.use(mockEmptyGetLogs());
    server.use(mockGetEnterpriseInfoInactive());
  });

  afterAll(() => server.close());

  describe('on close', () => {
    it('should route user back to pipeline view on close', async () => {
      server.use(
        rest.post<InspectJobRequest, Empty, JobInfo>(
          '/api/pps_v2.API/InspectJob',
          async (req, res, ctx) => {
            const body = await req.json();
            if (body?.job?.id === '23b9af7d5d4343219bc8e02ff44cd55a') {
              return res(ctx.json(buildJob({})));
            }
          },
        ),
      );
      window.history.replaceState(
        {},
        '',
        '/lineage/default/pipelines/montage/jobs/23b9af7d5d4343219bc8e02ff44cd55a/logs',
      );
      render(<PipelineDatumViewer />);

      await click(await screen.findByTestId('SidePanel__closeModal'));

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/lineage/default/pipelines/montage',
        ),
      );
    });

    it('should route user back to job view on close', async () => {
      server.use(
        rest.post<InspectJobRequest, Empty, JobInfo>(
          '/api/pps_v2.API/InspectJob',
          async (req, res, ctx) => {
            const body = await req.json();
            if (body?.job?.id === '23b9af7d5d4343219bc8e02ff44cd55a') {
              return res(ctx.json(buildJob({})));
            }
          },
        ),
      );
      window.history.replaceState(
        {},
        '',
        '/project/default/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
      );

      render(<JobDatumViewer />);

      await click(await screen.findByTestId('SidePanel__closeModal'));

      await waitFor(() =>
        expect(window.location.pathname).toBe('/project/default/jobs/subjobs'),
      );
    });
  });

  describe('Right Panel', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/project/default/jobs/5c1aa9bc87dd411ba5a1be0c80a3ebc2/pipeline/montage/logs/datum/05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7',
      );
    });

    it('should render datum details', async () => {
      render(<JobDatumViewer />);

      expect(await screen.findByText('Success')).toBeVisible();

      const runtimeDropDown = await screen.findByText('Cumulative Time: 6 s');
      expect(runtimeDropDown).toBeVisible();
      userEvent.click(runtimeDropDown);
      expect(await screen.findByText('1 s')).toBeVisible();
      expect(await screen.findByText('1 kB')).toBeVisible();
      expect(await screen.findByText('2 s')).toBeVisible();
      expect(await screen.findByText('3 s')).toBeVisible();
      expect(await screen.findByText('3 kB')).toBeVisible();
    });

    it('should render the root key of input spec', async () => {
      render(<JobDatumViewer />);
      const codeSpec = await screen.findByTestId(
        'ConfigFilePreview__codeElement',
      );

      expect(
        await within(codeSpec).findAllByText((node) => node.includes('edges')),
      ).toHaveLength(2); // Allow code element to load

      const deepCloneCodeSpec = codeSpec.cloneNode(true);
      // Remove the blinking cusror if it is present
      const cursors = (deepCloneCodeSpec as HTMLElement).querySelectorAll(
        '.cm-cursor-primary',
      );
      cursors.forEach((cursor) => cursor.remove());

      expect(deepCloneCodeSpec).toMatchSnapshot();
    });

    it('should render N/A when the runtime data is not available', async () => {
      server.use(
        rest.post<InspectDatumRequest, Empty, DatumInfo>(
          '/api/pps_v2.API/InspectDatum',
          (_req, res, ctx) =>
            res(
              ctx.json(
                buildDatum({
                  datum: {
                    id: '05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7',
                    job: {
                      id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
                    },
                  },
                  state: DatumState.SUCCESS,
                }),
              ),
            ),
        ),
      );
      render(<JobDatumViewer />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      const runtimeDropDown = screen.getByRole('button', {
        name: /cumulative time/i,
      });
      expect(runtimeDropDown).toBeVisible();
      userEvent.click(runtimeDropDown);
      await screen.findByText('Download'); // Allow dropdown to finish loading
      expect(await screen.findAllByText('N/A')).toHaveLength(3);
    });

    it('should render a skipped datums details', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/default/jobs/5c1aa9bc87dd411ba5a1be0c80a3ebc2/pipeline/montage/logs/datum/ch3db37fa4594a00ebf1dc972f81b58de642cd0cfca811e1b5bd6a2bb292a8e0',
      );
      server.use(mockGetJob5CDatumCH());
      render(<JobDatumViewer />);
      expect(await screen.findByText('Skipped')).toBeVisible();
      expect(
        await screen.findByText(
          'This datum has been successfully processed in a previous job, has not changed since then, and therefore, it was skipped in the current job.',
        ),
      ).toBeVisible();
      expect(await screen.findByText('Previous Job')).toBeVisible();
      expect(
        await screen.findByText('14291af7da4a4143b8ae12eba16d4661'),
      ).toBeVisible();
      expect(await screen.findByText('Cumulative Time: 3 s')).toBeVisible();
    });
  });

  describe('Spout Pipeline', () => {
    beforeEach(() => {
      server.use(mockGetSpoutPipeline());
      window.history.replaceState(
        {},
        '',
        '/lineage/default/pipelines/montage/logs',
      );
    });

    it('should hide the left panel', async () => {
      render(<PipelineDatumViewer />);
      await screen.findByTestId('SidePanel__right');
      await waitFor(() => {
        expect(screen.queryByTestId('SidePanel__left')).not.toBeInTheDocument();
      });
    });

    it('should show pipeline info in the right panel', async () => {
      server.use(mockGetSpoutPipeline());
      render(<PipelineDatumViewer />);

      expect(await screen.findByText('Running')).toBeVisible();
      expect(screen.getByLabelText('Pipeline Type')).toHaveTextContent('Spout');
    });
  });

  describe('Service Pipeline', () => {
    beforeEach(() => {
      server.use(mockGetServicePipeline());
      window.history.replaceState(
        {},
        '',
        '/lineage/default/pipelines/montage/logs',
      );
    });

    it('should hide the left panel', async () => {
      render(<PipelineDatumViewer />);
      await screen.findByTestId('SidePanel__right');
      await waitFor(() => {
        expect(screen.queryByTestId('SidePanel__left')).not.toBeInTheDocument();
      });
    });

    it('should show pipeline info in the right panel', async () => {
      server.use(mockGetServicePipeline());
      render(<PipelineDatumViewer />);

      expect(await screen.findByText('Running')).toBeVisible();
      expect(screen.getByLabelText('Pipeline Type')).toHaveTextContent(
        'Service',
      );
    });
  });
});
