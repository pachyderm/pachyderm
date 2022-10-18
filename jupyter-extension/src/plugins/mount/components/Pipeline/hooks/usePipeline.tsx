import YAML from 'yaml';

import {requestAPI} from '../../../../../handler';
import {useEffect, useState} from 'react';
import {ServerConnection} from '@jupyterlab/services';
import {
  CurrentDatumResponse,
  ListMountsResponse,
  MountDatumResponse,
} from 'plugins/mount/types';

export type usePipelineResponse = {
  loading: boolean;
  inputSpec: string;
  setInputSpec: (input: string) => void;
  requirements: string;
  setRequirements: (input: string) => void;
  secrets: string;
  setSecrets: (input: string) => void;
  callCreatePipeline: () => Promise<void>;
  errorMessage: string;
};

export const usePipeline = (
  showPipeline: boolean,
): usePipelineResponse => {
  const [loading, setLoading] = useState(false);
  const [inputSpec, setInputSpec] = useState('');
  const [requirements, setRequirements] = useState('');
  const [secrets, setSecrets] = useState('');
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

    let secretsParsed;
    try {
      secretsParsed = YAML.parse(secrets);
    } catch (e) {
      if (e instanceof YAML.YAMLParseError) {
        secretsParsed = JSON.parse(secrets);
      } else {
        throw e;
      }
    }

    setErrorMessage('No action hooked up');
    setLoading(false);
  };

  return {
    loading,
    inputSpec,
    setInputSpec,
    requirements,
    setRequirements,
    secrets,
    setSecrets,
    callCreatePipeline,
    errorMessage,
  };
};
