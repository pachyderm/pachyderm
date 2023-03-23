import YAML from 'yaml';

import {useEffect, useState} from 'react';
import {CreatePipelineResponse, SameMetadata} from '../../../types';
import {requestAPI} from '../../../../../handler';
import {ReadonlyJSONObject} from '@lumino/coreutils';

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
};

export const usePipeline = (
  metadata: SameMetadata | undefined,
  notebookPath: string | undefined,
  saveNotebookMetaData: (metadata: any) => void,
): usePipelineResponse => {
  const [loading, setLoading] = useState(false);
  const [pipelineName, setPipelineName] = useState('');
  const [imageName, setImageName] = useState('');
  const [inputSpec, setInputSpec] = useState('');
  const [requirements, setRequirements] = useState('');
  const [errorMessage, setErrorMessage] = useState('');

  useEffect(() => {
    setImageName(metadata?.environments.default.image_tag ?? '');
    setPipelineName(metadata?.metadata.name ?? '');
    setRequirements(metadata?.notebook.requirements ?? '');
    if (metadata?.run.input) {
      const input = JSON.parse(metadata?.run.input); //TODO: Catch errors
      setInputSpec(YAML.stringify(input));
    } else {
      setInputSpec('');
    }
  }, [metadata]);

  let input;
  try {
    input = YAML.parse(inputSpec);
  } catch (e) {
    if (e instanceof YAML.YAMLParseError) {
      input = JSON.parse(inputSpec);
    } else {
      throw e;
    }
  }

  const samemeta: SameMetadata = {
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

  const callCreatePipeline = async () => {
    setLoading(true);
    setErrorMessage('');

    const response = await requestAPI<CreatePipelineResponse>(
      `pps/_create/${notebookPath}`,
      'PUT',
      samemeta as ReadonlyJSONObject,
    );
    if (response.error) {
      setErrorMessage(response.error);
    }
    console.log('create pipeline called');
    setLoading(false);
  };

  const callSavePipeline = async () => {
    saveNotebookMetaData(samemeta);
    console.log('save pipeline called');
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
  };
};
