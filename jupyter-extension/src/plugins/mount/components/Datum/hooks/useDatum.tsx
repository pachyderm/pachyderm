import YAML from 'yaml';
import {JSONObject} from '@lumino/coreutils';
import {useEffect, useState} from 'react';
import {ServerConnection} from '@jupyterlab/services';
import {isEqual} from 'lodash';

import {requestAPI} from '../../../../../handler';
import {
  CrossInputSpec,
  CurrentDatumResponse,
  ListMountsResponse,
  MountDatumResponse,
  PfsInput,
} from 'plugins/mount/types';

export type useDatumResponse = {
  loading: boolean;
  shouldShowCycler: boolean;
  currDatum: MountDatumResponse;
  currIdx: number;
  setCurrIdx: (idx: number) => void;
  inputSpec: string;
  setInputSpec: (input: string) => void;
  callMountDatums: () => Promise<void>;
  callUnmountAll: () => Promise<void>;
  errorMessage: string;
  saveInputSpec: () => void;
  initialInputSpec: JSONObject;
};

export const useDatum = (
  showDatum: boolean,
  keepMounted: boolean,
  refresh: (path: string) => void,
  pollRefresh: () => Promise<void>,
  repoViewInputSpec: CrossInputSpec | PfsInput,
  currentDatumInfo?: CurrentDatumResponse,
): useDatumResponse => {
  const [loading, setLoading] = useState(false);
  const [shouldShowCycler, setShouldShowCycler] = useState(false);
  const [currIdx, setCurrIdx] = useState(-1);
  const [currDatum, setCurrDatum] = useState<MountDatumResponse>({
    id: '',
    idx: -1,
    num_datums: 0,
  });
  const [inputSpec, setInputSpec] = useState('');
  const [initialInputSpec, setInitialInputSpec] = useState({});
  const [errorMessage, setErrorMessage] = useState('');
  const [datumViewInputSpec, setDatumViewInputSpec] = useState<any>({});
  const [isYaml, setIsYaml] = useState(true);

  useEffect(() => {
    if (showDatum && currIdx !== -1) {
      callShowDatum();
    }
  }, [currIdx, showDatum]);

  useEffect(() => {
    if (showDatum && !keepMounted) {
      callUnmountAll();
    }
    if (keepMounted && currentDatumInfo) {
      setShouldShowCycler(true);
      setCurrIdx(currentDatumInfo.curr_idx);
      setCurrDatum({
        id: '',
        idx: currentDatumInfo.curr_idx,
        num_datums: currentDatumInfo.num_datums,
      });
      setInputSpec(inputSpecObjToText(currentDatumInfo.input));
    }
  }, [showDatum]);

  useEffect(() => {
    if (typeof datumViewInputSpec === 'string') {
      setInputSpec(datumViewInputSpec);
    } else {
      let specToShow = {};
      if (Object.keys(datumViewInputSpec).length === 0) {
        specToShow = repoViewInputSpec;
      } else {
        specToShow = datumViewInputSpec;
      }
      setInputSpec(inputSpecObjToText(specToShow));
      setInitialInputSpec(specToShow);
    }
  }, [showDatum, repoViewInputSpec]);

  const saveInputSpec = (): void => {
    try {
      const inputSpecObj = inputSpecTextToObj();
      if (isEqual(repoViewInputSpec, inputSpecObj)) {
        setDatumViewInputSpec({});
      } else {
        setDatumViewInputSpec(inputSpecObj ? inputSpecObj : {});
      }
    } catch (e) {
      if (e instanceof YAML.YAMLParseError) {
        setDatumViewInputSpec(inputSpec);
      } else {
        throw e;
      }
    }
  };

  let useYaml = isYaml;
  const inputSpecTextToObj = (): JSONObject => {
    let spec = {};
    try {
      spec = JSON.parse(inputSpec);
      useYaml = false;
      setIsYaml(false);
    } catch (e) {
      if (e instanceof SyntaxError) {
        spec = YAML.parse(inputSpec);
        useYaml = true;
        setIsYaml(true);
      } else {
        throw e;
      }
    }
    return spec;
  };

  useEffect(() => {
    useYaml = isYaml;
  }, [isYaml]);

  const inputSpecObjToText = (specObj: JSONObject): string => {
    if (Object.keys(specObj).length === 0) {
      return '';
    }
    if (useYaml) {
      return YAML.stringify(specObj, null, 2);
    }
    return JSON.stringify(specObj, null, 2);
  };

  const callMountDatums = async () => {
    setLoading(true);
    setErrorMessage('');

    try {
      const spec = inputSpecTextToObj();
      const res = await requestAPI<MountDatumResponse>('_mount_datums', 'PUT', {
        input: spec,
      });
      refresh('');
      setCurrIdx(0);
      setCurrDatum(res);
      setShouldShowCycler(true);
      setInputSpec(inputSpecObjToText(spec));
    } catch (e) {
      console.log(e);
      if (e instanceof YAML.YAMLParseError) {
        setErrorMessage(
          'Poorly formatted input spec- must be either YAML or JSON',
        );
      } else if (e instanceof ServerConnection.ResponseError) {
        setErrorMessage('Bad data in input spec');
      } else {
        setErrorMessage('Error mounting datums');
      }
    }

    setLoading(false);
  };

  const callShowDatum = async () => {
    setLoading(true);

    try {
      const res = await requestAPI<MountDatumResponse>(
        `_show_datum?idx=${currIdx}`,
        'PUT',
      );
      refresh('');
      setCurrDatum(res);
    } catch (e) {
      console.log(e);
    }

    setLoading(false);
  };

  const callUnmountAll = async () => {
    setLoading(true);

    try {
      refresh('');
      await requestAPI<ListMountsResponse>('_unmount_all', 'PUT');
      refresh('');
      await pollRefresh();
      setCurrIdx(-1);
      setCurrDatum({id: '', idx: -1, num_datums: 0});
      setShouldShowCycler(false);
    } catch (e) {
      console.log(e);
    }

    setLoading(false);
  };

  return {
    loading,
    shouldShowCycler,
    currDatum,
    currIdx,
    setCurrIdx,
    inputSpec,
    setInputSpec,
    callMountDatums,
    callUnmountAll,
    errorMessage,
    saveInputSpec,
    initialInputSpec,
  };
};
