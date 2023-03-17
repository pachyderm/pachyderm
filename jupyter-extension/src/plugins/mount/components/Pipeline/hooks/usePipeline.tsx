import YAML from 'yaml';

import {useEffect, useState} from 'react';
import {SameMetadata} from '../../../types';

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
  saveNotebookMetaData: (metadata: any) => void,
): usePipelineResponse => {
  const [loading, setLoading] = useState(false);
  const [pipelineName, setPipelineName] = useState('');
  const [imageName, setImageName] = useState('');
  const [inputSpec, setInputSpec] = useState('');
  const [requirements, setRequirements] = useState('');
  const [errorMessage, setErrorMessage] = useState('');

  useEffect(() => {
    console.log('use effect: metadata: ', metadata);
    setImageName(
      metadata?.environments.default.image_tag
        ? metadata?.environments.default.image_tag
        : '',
    );
    setPipelineName(metadata?.metadata.name ? metadata?.metadata.name : '');
    setRequirements(
      metadata?.notebook.requirements ? metadata?.notebook.requirements : '',
    );
    //TODO Pipeline spec
  }, [metadata]);

  const callCreatePipeline = async () => {
    setLoading(true);
    setErrorMessage('');

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

    let reqsParsed;
    try {
      reqsParsed = YAML.parse(requirements);
    } catch (e) {
      if (e instanceof YAML.YAMLParseError) {
        reqsParsed = JSON.parse(requirements);
      } else {
        throw e;
      }
    }

    setErrorMessage('No action hooked up');
    setLoading(false);
  };

  const callSavePipeline = async () => {
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
        name: pipelineName + ' run',
      },
    };

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
