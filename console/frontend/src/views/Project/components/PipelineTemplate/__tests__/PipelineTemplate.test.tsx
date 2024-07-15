import {act, render, screen, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {CreatePipelineV2Request} from '@dash-frontend/api/pps';
import {RequestError} from '@dash-frontend/api/utils/error';
import {
  mockGetVersionInfo,
  mockGetEnterpriseInfo,
  mockPipelines,
  mockInspectProject,
  mockCreatePipelineSuccess,
  mockPipelineSummaries,
} from '@dash-frontend/mocks';
import {
  mockBadFormatTemplate,
  mockRenderTemplate,
  mockRenderTemplateError,
  mockSimpleTemplate,
  mockSnowflakeTemplate,
  mockTemplatError,
  mockTextTemplate,
} from '@dash-frontend/mocks/template';
import {withContextProviders} from '@dash-frontend/testHelpers';

import PipelineTemplateComponent from '../PipelineTemplate';

describe('PipelineTemplate', () => {
  jest.useFakeTimers();
  const user = userEvent.setup({advanceTimers: jest.advanceTimersByTime});
  const server = setupServer();

  const PipelineTemplate = withContextProviders(() => {
    return <PipelineTemplateComponent />;
  });

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    window.history.replaceState({}, '', '/lineage/default/create/template');
    server.resetHandlers();
    server.use(mockGetVersionInfo());
    server.use(mockGetEnterpriseInfo());
    server.use(mockPipelines());
    server.use(mockInspectProject());
    server.use(mockCreatePipelineSuccess());
    server.use(mockSimpleTemplate());
    server.use(mockPipelineSummaries());

    // eslint does not like that "render" is in the function name
    // eslint-disable-next-line testing-library/no-render-in-setup
    server.use(mockRenderTemplate());
  });

  afterAll(() => server.close());

  it('should allow the user to return to the lineage page', async () => {
    render(<PipelineTemplate />);
    const cancelButton = await screen.findByRole('button', {
      name: /cancel/i,
    });
    await act(async () => await user.click(cancelButton));
    expect(window.location.pathname).toBe('/lineage/default');
  });

  describe('Upload template from url', () => {
    it('should verify that input is a valid url', async () => {
      render(<PipelineTemplate />);
      const urlInput = await screen.findByRole('textbox');

      await act(() => user.type(urlInput, 'hsdfsdgfgd'));
      act(() => jest.runAllTimers());
      await screen.findByText('Invalid url');
      const continueButton = await screen.findByRole('button', {
        name: /continue/i,
      });
      expect(continueButton).toBeDisabled();
      await act(async () => await user.clear(urlInput));
      await act(
        async () =>
          await user.type(urlInput, 'https://www.hosted-template-simple.com'),
      );
      act(() => jest.runAllTimers());
      await screen.findByText('Template uploaded');
      expect(continueButton).toBeEnabled();
    });

    it('should display an error if there is an issue pulling the template from the url', async () => {
      server.use(mockTemplatError());
      render(<PipelineTemplate />);
      const urlInput = await screen.findByRole('textbox');
      await act(
        async () =>
          await user.type(urlInput, 'https://www.hosted-unknown-error.com'),
      );
      act(() => jest.runAllTimers());
      await screen.findByText('Error retrieving template from url');
      expect(
        await screen.findByRole('button', {name: /continue/i}),
      ).toBeDisabled();
    });

    it('should display an error if there no template metadata', async () => {
      server.use(mockTextTemplate());
      render(<PipelineTemplate />);
      const urlInput = await screen.findByRole('textbox');
      await act(
        async () =>
          await user.type(urlInput, 'https://www.hosted-template-text.com'),
      );
      act(() => jest.runAllTimers());
      await screen.findByText('No template metadata found');
      expect(
        await screen.findByRole('button', {name: /continue/i}),
      ).toBeDisabled();
    });

    it('should display an error modal metadata is not formatted correctly', async () => {
      server.use(mockBadFormatTemplate());
      render(<PipelineTemplate />);
      const urlInput = await screen.findByRole('textbox');
      await act(
        async () =>
          await user.type(
            urlInput,
            'https://www.hosted-template-bad-format.com',
          ),
      );
      act(() => jest.runAllTimers());
      await screen.findByText('Error parsing metadata from template');
      await act(
        async () => await user.click(await screen.findByText('See Full Error')),
      );

      await screen.findByText('Metadata Parsing Error');
      await screen.findByText('Metadata missing title');
      await screen.findByText('Argument 0: missing name');

      await act(async () => await user.click(await screen.findByText('Back')));
      expect(
        await screen.findByRole('button', {name: /continue/i}),
      ).toBeDisabled();
    });

    it('should allow users to upload a template', async () => {
      render(<PipelineTemplate />);
      const urlInput = await screen.findByRole('textbox');
      await act(
        async () =>
          await user.type(urlInput, 'https://www.hosted-template-simple.com'),
      );
      act(() => jest.runAllTimers());
      await screen.findByText('Template uploaded');
      expect(
        await screen.findByRole('button', {name: /continue/i}),
      ).toBeEnabled();
    });
  });

  describe('Template Form', () => {
    const loadTemplate = async (url: string) => {
      const urlInput = await screen.findByRole('textbox');
      await act(async () => await user.type(urlInput, url));
      act(() => jest.runAllTimers());
      await screen.findByText('Template uploaded');

      const continueButton = await screen.findByRole('button', {
        name: /continue/i,
      });
      await act(async () => await user.click(continueButton));
    };

    it('should show pipeline nane and description', async () => {
      jest.useFakeTimers();
      render(<PipelineTemplate />);
      await loadTemplate('https://www.hosted-template-simple.com');
      await screen.findByText('Test Template');
      await screen.findByText('Template description.');
    });

    it('should render form fields, put fields with no default vaule at top and fill in default values', async () => {
      server.use(mockSnowflakeTemplate());

      render(<PipelineTemplate />);
      await loadTemplate('https://www.hosted-template-snowflake.com');
      await screen.findByText('Snowflake Integration');
      await screen.findByText(
        'Creates a cron pipeline that can execute a query against a Snowflake database and return the results in a single output file.',
      );

      const fields = await screen.findAllByTestId(/PipelineTemplateField/i);
      expect(fields).toHaveLength(5);

      await within(fields[0]).findByLabelText('name');
      expect(fields[0]).toHaveTextContent('The name of the pipeline.');
      expect(
        await within(fields[0]).findByRole('textbox', {name: /name/i}),
      ).toHaveValue('');

      await within(fields[1]).findByLabelText('query');
      expect(fields[1]).toHaveTextContent('The query to execute.');
      expect(
        await within(fields[1]).findByRole('textbox', {name: /query/i}),
      ).toHaveValue('');

      await within(fields[2]).findByLabelText('inputrepo');
      expect(fields[2]).toHaveTextContent(
        "The suffix to use on the cron pipeline's input repo.",
      );
      expect(
        await within(fields[2]).findByRole('textbox', {name: /inputrepo/i}),
      ).toHaveValue('tick');

      await within(fields[3]).findByLabelText('spec');
      expect(fields[3]).toHaveTextContent('The cron spec to use.');
      expect(
        await within(fields[3]).findByRole('textbox', {name: /spec/i}),
      ).toHaveValue('@never');

      await within(fields[4]).findByLabelText('count');
      expect(fields[4]).toHaveTextContent('Number input field.');
      expect(
        await within(fields[4]).findByRole('spinbutton', {name: /count/i}),
      ).toHaveValue(4);
    });

    it('should validate form fields', async () => {
      server.use(mockSnowflakeTemplate());
      const user = userEvent.setup({delay: null});
      jest.useFakeTimers();
      render(<PipelineTemplate />);
      await loadTemplate('https://www.hosted-template-snowflake.com');
      await screen.findByText('Snowflake Integration');

      const submitButton = await screen.findByRole('button', {
        name: /create pipeline/i,
      });

      expect(submitButton).toBeDisabled();
      const nameField = await screen.findByRole('textbox', {name: /name/i});
      await act(async () => await user.type(nameField, 'test'));
      const queryField = await screen.findByRole('textbox', {name: /query/i});
      await act(async () => await user.type(queryField, 'user-query'));
      expect(submitButton).toBeEnabled();

      const countField = await screen.findByRole('spinbutton', {
        name: /count/i,
      });
      await act(async () => await user.clear(countField));
      expect(submitButton).toBeDisabled();
      await act(async () => await user.type(countField, 'sdfsdf'));
      expect(submitButton).toBeDisabled();
      await act(async () => await user.type(countField, '4'));

      const specField = await screen.findByRole('textbox', {
        name: /spec/i,
      });
      await act(async () => await user.clear(specField));
      expect(submitButton).toBeDisabled();
      await act(async () => await user.type(specField, 'sdfsdf'));
      expect(submitButton).toBeEnabled();
    });

    it('should show render template error', async () => {
      server.use(mockRenderTemplateError());
      render(<PipelineTemplate />);
      await loadTemplate('https://www.hosted-template-simple.com');
      await screen.findByText('Test Template');

      await act(
        async () =>
          await user.click(
            await screen.findByRole('button', {
              name: /create pipeline/i,
            }),
          ),
      );

      await screen.findByText('Unable to parse template');

      await act(
        async () =>
          await user.click(
            await screen.findByRole('button', {
              name: 'See Full Error',
            }),
          ),
      );
      expect(
        await within(screen.getByRole('dialog')).findByText(
          /Expected token OPERATOR but got/i,
        ),
      ).toBeInTheDocument();
    });

    it('should show create pipeline error', async () => {
      server.use(
        rest.post<CreatePipelineV2Request, Empty, RequestError>(
          '/api/pps_v2.API/CreatePipelineV2',
          async (_req, res, ctx) => {
            return res(
              ctx.status(400),
              ctx.json({
                code: 3,
                message: 'Invalid JSON',
                details: [],
              }),
            );
          },
        ),
      );

      render(<PipelineTemplate />);
      await loadTemplate('https://www.hosted-template-simple.com');
      await screen.findByText('Test Template');

      await act(
        async () =>
          await user.click(
            await screen.findByRole('button', {
              name: /create pipeline/i,
            }),
          ),
      );

      await screen.findByText('Unable to create pipeline');

      await act(
        async () =>
          await user.click(
            await screen.findByRole('button', {
              name: 'See Full Error',
            }),
          ),
      );

      expect(
        await within(screen.getByRole('dialog')).findByText('Invalid JSON'),
      ).toBeInTheDocument();
    });

    it('should create pipeline from template', async () => {
      render(<PipelineTemplate />);
      await loadTemplate('https://www.hosted-template-simple.com');
      await screen.findByText('Test Template');

      await act(
        async () =>
          await user.click(
            await screen.findByRole('button', {
              name: /create pipeline/i,
            }),
          ),
      );

      expect(window.location.pathname).toBe('/lineage/default');
    });
  });
});
