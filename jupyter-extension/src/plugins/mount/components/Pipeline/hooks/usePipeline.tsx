import YAML from 'yaml';
import {useEffect, useState} from 'react';
import {ServerConnection} from '@jupyterlab/services';

import {CreatePipelineResponse, PpsContext, SameMetadata} from '../../../types';
import {requestAPI} from '../../../../../handler';

export type usePipelineResponse = {
  loading: boolean;
  pipelineName: string;
  setPipelineName: (input: string) => void;
  imageName: string;
  setImageName: (input: string) => void;
  inputSpec: string;
  setInputSpec: (input: string) => void;
  requirements: string;
  setRequirements: (input: string) => void;
  callCreatePipeline: () => Promise<void>;
  callSavePipeline: () => void;
  errorMessage: string;
  responseMessage: string;
};

export const usePipeline = (
  ppsContext: PpsContext | undefined,
  saveNotebookMetaData: (metadata: SameMetadata) => void,
): usePipelineResponse => {
  const [loading, setLoading] = useState(false);
  const [pipelineName, setPipelineName] = useState('');
  const [imageName, setImageName] = useState('');
  const [inputSpec, setInputSpec] = useState('');
  const [requirements, setRequirements] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const [responseMessage, setResponseMessage] = useState('');

  useEffect(() => {
    setImageName(ppsContext?.config?.environments.default.image_tag ?? '');
    setPipelineName(ppsContext?.config?.metadata.name ?? '');
    setRequirements(ppsContext?.config?.notebook.requirements ?? '');
    setResponseMessage('');
    if (ppsContext?.config?.run.input) {
      const input = JSON.parse(ppsContext.config.run.input); //TODO: Catch errors
      setInputSpec(YAML.stringify(input));
    } else {
      setInputSpec('');
    }
  }, [ppsContext]);

  let callCreatePipeline: () => Promise<void>;
  if (ppsContext?.notebookModel) {
    const notebook = ppsContext.notebookModel;
    callCreatePipeline = async () => {
      setLoading(true);
      setErrorMessage('');
      setResponseMessage('');
      try {
        const response = await requestAPI<CreatePipelineResponse>(
          `pps/_create/${encodeURI(notebook.path)}`,
          'PUT',
          {last_modified_time: notebook.last_modified},
        );
        if (response.message !== null) {
          setResponseMessage(response.message);
        }
      } catch (e) {
        if (e instanceof ServerConnection.ResponseError) {
          setErrorMessage(e.message);
        } else {
          throw e;
        }
      }
      setLoading(false);
    };
  } else {
    // If no notebookModel is defined, we cannot create a pipeline.
    callCreatePipeline = async () => {
      setErrorMessage('Error: No notebook in focus');
    };
  }

  const callSavePipeline = async () => {
    let input: string;
    try {
      input = YAML.parse(inputSpec);
    } catch (e) {
      if (e instanceof YAML.YAMLParseError) {
        input = JSON.parse(inputSpec);
      } else {
        throw e;
      }
    }

    const sameMetadata = {
      apiVersion: 'sameproject.ml/v1alpha1',
      environments: {
        default: {
          image_tag: imageName,
        },
      },
      metadata: {
        name: pipelineName,
        version: '0.0.0',
      },
      notebook: {
        requirements: requirements,
      },
      run: {
        name: pipelineName,
        input: JSON.stringify(input),
      },
    };
    saveNotebookMetaData(sameMetadata);
  };

  return {
    loading,
    pipelineName,
    setPipelineName,
    imageName,
    setImageName,
    inputSpec,
    setInputSpec,
    requirements,
    setRequirements,
    callCreatePipeline,
    callSavePipeline,
    errorMessage,
    responseMessage,
  };
};
