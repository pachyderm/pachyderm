import YAML from 'yaml';
import {useEffect, useState} from 'react';
import {ServerConnection} from '@jupyterlab/services';

import {CreatePipelineResponse, PpsContext, PpsMetadata} from '../../../types';
import {requestAPI} from '../../../../../handler';
import {ReadonlyJSONObject} from '@lumino/coreutils';

export const PPS_VERSION = 'v1.0.0';

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
  saveNotebookMetaData: (metadata: PpsMetadata) => void,
): usePipelineResponse => {
  const [loading, setLoading] = useState(false);
  const [pipelineName, setPipelineName] = useState('');
  const [imageName, setImageName] = useState('');
  const [inputSpec, setInputSpec] = useState('');
  const [requirements, setRequirements] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const [responseMessage, setResponseMessage] = useState('');

  useEffect(() => {
    setImageName(ppsContext?.metadata?.config.image ?? '');
    setPipelineName(ppsContext?.metadata?.config.pipeline_name ?? '');
    setRequirements(ppsContext?.metadata?.config.requirements ?? '');
    setResponseMessage('');
    if (ppsContext?.metadata?.config.input_spec) {
      try {
        setInputSpec(YAML.stringify(ppsContext.metadata.config.input_spec));
      } catch (_e) {
        setInputSpec('');
        setErrorMessage('error parsing input spec'); // This error might confuse user.
      }
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
          setErrorMessage('Error creating pipeline');
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
    setErrorMessage('');

    let inputSpecJson;
    try {
      inputSpecJson = parseInputSpec(inputSpec);
    } catch (e) {
      // TODO: More helpful error reporting.
      setErrorMessage('error parsing input spec -- saving aborted');
      return;
    }

    const ppsMetadata: PpsMetadata = {
      version: PPS_VERSION,
      config: {
        pipeline_name: pipelineName,
        image: imageName,
        requirements: requirements,
        input_spec: inputSpecJson,
      },
    };
    saveNotebookMetaData(ppsMetadata);
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

/*
parseInputSpec attempts to convert the entry within the InputSpec text area
  into a JSON serializable format. Throws an error if not possible
 */
const parseInputSpec = (spec: string): ReadonlyJSONObject => {
  let input;
  try {
    input = YAML.parse(spec);
  } catch (e) {
    if (e instanceof YAML.YAMLParseError) {
      input = JSON.parse(spec);
    } else {
      throw e;
    }
  }
  return input as ReadonlyJSONObject;
};
