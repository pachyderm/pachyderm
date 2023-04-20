import YAML from 'yaml';
import {useEffect, useState} from 'react';
import {ServerConnection} from '@jupyterlab/services';

import {
  CreatePipelineResponse,
  MountSettings,
  Pipeline,
  PpsContext,
  PpsMetadata,
} from '../../../types';
import {requestAPI} from '../../../../../handler';
import {ReadonlyJSONObject} from '@lumino/coreutils';

export const PPS_VERSION = 'v1.0.0';

export type usePipelineResponse = {
  loading: boolean;
  pipeline: Pipeline;
  setPipeline: (input: string) => void;
  imageName: string;
  setImageName: (input: string) => void;
  inputSpec: string;
  setInputSpec: (input: string) => void;
  requirements: string;
  setRequirements: (input: string) => void;
  callCreatePipeline: () => Promise<void>;
  callSavePipeline: () => void;
  currentNotebook: string;
  errorMessage: string;
  responseMessage: string;
};

export const usePipeline = (
  ppsContext: PpsContext | undefined,
  settings: MountSettings,
  saveNotebookMetaData: (metadata: PpsMetadata) => void,
): usePipelineResponse => {
  const [loading, setLoading] = useState(false);
  const [pipeline, setPipeline] = useState({name: ''} as Pipeline);
  const [imageName, setImageName] = useState('');
  const [inputSpec, setInputSpec] = useState('');
  const [requirements, setRequirements] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const [responseMessage, setResponseMessage] = useState('');
  const [currentNotebook, setCurrentNotebook] = useState('None');

  const setPipelineFromString = (input: string) => {
    if (input === '') {
      setPipeline({name: ''} as Pipeline);
      return;
    }
    const parts = splitAtFirstSlash(input);
    if (parts.length === 1) {
      setPipeline({name: input} as Pipeline);
    } else {
      setPipeline({
        name: parts[1],
        project: {name: parts[0]},
      } as Pipeline);
    }
  };

  useEffect(() => {
    setImageName(
      ppsContext?.metadata?.config.image ?? settings.defaultPipelineImage,
    );
    setPipeline(
      ppsContext?.metadata?.config.pipeline ?? ({name: ''} as Pipeline),
    );
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
    setCurrentNotebook(ppsContext?.notebookModel?.name ?? 'None');
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
        pipeline: pipeline,
        image: imageName,
        requirements: requirements,
        input_spec: inputSpecJson,
      },
    };
    saveNotebookMetaData(ppsMetadata);
  };

  return {
    loading,
    pipeline,
    setPipeline: setPipelineFromString,
    imageName,
    setImageName,
    inputSpec,
    setInputSpec,
    requirements,
    setRequirements,
    callCreatePipeline,
    callSavePipeline,
    currentNotebook,
    errorMessage,
    responseMessage,
  };
};

/*
splitAtFirstSlash splits a string into two components if it contains a backslash.
  For example test/name => [test, name]. If the text does not contain a backslash
  then text is returned as a one element array, name => [name].
 */
export const splitAtFirstSlash = (text: string): string[] => {
  return text.split(/\/(.*)/s, 2);
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
