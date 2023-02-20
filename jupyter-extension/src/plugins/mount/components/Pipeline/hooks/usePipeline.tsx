import YAML from 'yaml';

import {useEffect, useState} from 'react';

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
  errorMessage: string;
};

export const usePipeline = (): usePipelineResponse => {
  const [loading, setLoading] = useState(false);
  const [pipelineName, setPipelineName] = useState('');
  const [imageName, setImageName] = useState('');
  const [inputSpec, setInputSpec] = useState('');
  const [requirements, setRequirements] = useState('');
  const [errorMessage, setErrorMessage] = useState('');

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
    errorMessage,
  };
};
