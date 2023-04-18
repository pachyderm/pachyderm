import React from 'react';
import {render} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {Contents} from '@jupyterlab/services';

import * as requestAPI from '../../../../../handler';
import {mockedRequestAPI} from 'utils/testUtils';
import Pipeline from '../Pipeline';
jest.mock('../../../../../handler');
import {MountSettings, PpsContext} from '../../../types';

describe('PPS screen', () => {
  let setShowPipeline = jest.fn();
  const saveNotebookMetaData = jest.fn();

  const testNotebookName = 'NotARealNotebook.ipynb';
  const notebookModel = {name: testNotebookName} as Contents.IModel;
  const settings: MountSettings = {defaultPipelineImage: 'DefaultImage:Tag'};

  const mockRequestAPI = requestAPI as jest.Mocked<typeof requestAPI>;
  beforeEach(() => {
    setShowPipeline = jest.fn();
    mockRequestAPI.requestAPI.mockImplementation(mockedRequestAPI({}));
  });

  describe('spec preview', () => {
    it('proper preview', async () => {
      const ppsContext = {config: null, notebookModel};
      const {getByTestId, findByTestId} = render(
        <Pipeline
          ppsContext={ppsContext}
          settings={settings}
          setShowPipeline={setShowPipeline}
          saveNotebookMetadata={saveNotebookMetaData}
        />,
      );

      const valueCurrentNotebook = await findByTestId(
        'Pipeline__currentNotebookValue',
      );
      expect(valueCurrentNotebook).toHaveTextContent(testNotebookName);

      const inputPipelineName = await findByTestId(
        'Pipeline__inputPipelineName',
      );
      userEvent.type(inputPipelineName, 'ThisPipelineIsNamedFred');

      const inputImageName = await findByTestId('Pipeline__inputImageName');
      expect(inputImageName).toHaveValue(settings.defaultPipelineImage);
      userEvent.clear(inputImageName);
      userEvent.type(inputImageName, 'ThisImageIsNamedLucy');

      const inputRequirements = await findByTestId(
        'Pipeline__inputRequirements',
      );
      userEvent.type(inputRequirements, './requirements.txt');

      const inputInputSpec = await findByTestId('Pipeline__inputSpecInput');
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
        `name: ThisPipelineIsNamedFred
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

  describe('no notebook', () => {
    const ppsContext = {config: null, notebookModel: null};
    it('currentNotebook is None', async () => {
      const {findByTestId} = render(
        <Pipeline
          ppsContext={ppsContext}
          settings={settings}
          setShowPipeline={setShowPipeline}
          saveNotebookMetadata={saveNotebookMetaData}
        />,
      );

      const valueCurrentNotebook = await findByTestId(
        'Pipeline__currentNotebookValue',
      );
      expect(valueCurrentNotebook).toHaveTextContent('None');
    });
  });
});
