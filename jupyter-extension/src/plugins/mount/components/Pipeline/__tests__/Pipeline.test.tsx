import React from 'react';
import {render} from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import * as requestAPI from '../../../../../handler';
import {mockedRequestAPI} from 'utils/testUtils';
import Pipeline from '../Pipeline';
jest.mock('../../../../../handler');
import {PPS_VERSION} from '../hooks/usePipeline';
import {
  Pipeline as PipelineType,
  PpsContext,
  PpsMetadata,
} from '../../../types';

describe('PPS screen', () => {
  let setShowPipeline = jest.fn();
  const saveNotebookMetaData = jest.fn();
  const metadata: PpsMetadata = {
    version: PPS_VERSION,
    config: {
      pipeline: {name: ''} as PipelineType,
      image: '',
      requirements: '',
      input_spec: {},
    },
  };
  const context: PpsContext = {metadata: metadata, notebookModel: null};

  const mockRequestAPI = requestAPI as jest.Mocked<typeof requestAPI>;

  beforeEach(() => {
    setShowPipeline = jest.fn();
    mockRequestAPI.requestAPI.mockImplementation(mockedRequestAPI({}));
  });

  describe('spec preview', () => {
    it('proper preview', async () => {
      const {getByTestId, findByTestId} = render(
        <Pipeline
          ppsContext={context}
          setShowPipeline={setShowPipeline}
          saveNotebookMetadata={saveNotebookMetaData}
        />,
      );

      const inputPipelineName = await findByTestId(
        'Pipeline__inputPipelineName',
      );
      userEvent.type(inputPipelineName, 'test_project/ThisPipelineIsNamedFred');

      const inputImageName = await findByTestId('Pipeline__inputImageName');
      userEvent.type(inputImageName, 'ThisImageIsNamedLucy');

      const inputRequirements = await findByTestId(
        'Pipeline__inputRequirements',
      );
      userEvent.type(inputRequirements, './requirements.txt');

      const inputInputSpec = await findByTestId('Pipeline__inputSpecInput');
      userEvent.clear(inputInputSpec);
      userEvent.type(
        inputInputSpec,
        `pfs:
  repo: data
  branch: primary
  glob: /`,
      );
      // expect(setInputSpec).toHaveBeenCalled();

      const specPreview = getByTestId('Pipeline__specPreview');
      expect(specPreview).toHaveValue(
        `pipeline:
  name: ThisPipelineIsNamedFred
  project: test_project
transform:
  image: ThisImageIsNamedLucy
input:
  pfs:
    repo: data
    branch: primary
    glob: /
`,
      );
    });
  });
});
