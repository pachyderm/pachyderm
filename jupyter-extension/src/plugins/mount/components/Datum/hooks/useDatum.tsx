import YAML from 'yaml'

import {requestAPI} from '../../../../../handler';
import {useEffect, useState} from 'react';
import {ServerConnection} from '@jupyterlab/services';
import {DatumsResponse} from 'plugins/mount/types';

export type useDatumResponse = {
  loading: boolean;
  shouldShowCycler: boolean;
  currentDatumId: string;
  currentDatumIdx: number;
  setCurrentDatumIdx: (idx: number) => void;
  numDatums: number;
  inputSpec: string;
  setInputSpec: (input: string) => void;
  callMountDatums: () => Promise<void>;
  callUnmountAll: () => Promise<void>;
  errorMessage: string;
};

export const useDatum = (
  showDatum: boolean,
  keepMounted: boolean,
  refresh: () => void,
  pollRefresh: () => Promise<void>,
  currentDatumInfo?: DatumsResponse,
): useDatumResponse => {
  const [loading, setLoading] = useState(false);
  const [shouldShowCycler, setShouldShowCycler] = useState(false);
  const [currentDatumId, setCurrentDatumId] = useState('');
  const [currentDatumIdx, setCurrentDatumIdx] = useState(-1);
  const [numDatums, setNumDatums] = useState(-1);
  const [inputSpec, setInputSpec] = useState('');
  const [errorMessage, setErrorMessage] = useState('');

  useEffect(() => {
    if (showDatum && currentDatumIdx !== -1) {
      callShowDatum();
    }
  }, [currentDatumIdx, showDatum]);

  useEffect(() => {
    if (showDatum && !keepMounted) {
      callUnmountAll();
    }
    if (keepMounted && currentDatumInfo) {
      setShouldShowCycler(true);
      setCurrentDatumIdx(currentDatumInfo.curr_idx);
      setNumDatums(currentDatumInfo.num_datums);
      setInputSpec(YAML.stringify(currentDatumInfo.input, null, 2));
    }
  }, [showDatum]);

  const callMountDatums = async () => {
    setLoading(true);
    setErrorMessage('');

    try {
      let input;
      try {
        input = YAML.parse(inputSpec)
      } catch (e) {
        if (e instanceof YAML.YAMLParseError) {
          input = JSON.parse(inputSpec)
        } else {
          throw e
        }
      }

      const res = await requestAPI<any>('_mount_datums', 'PUT', {
        input: input,
      });
      refresh();
      setCurrentDatumId(res.id);
      setCurrentDatumIdx(res.idx);
      setNumDatums(res.num_datums);
      setShouldShowCycler(true);
      setInputSpec(YAML.stringify(input, null, 2));
    } catch (e) {
      console.log(e);
      if (e instanceof SyntaxError) {
        setErrorMessage('Poorly formatted input spec- must be either YAML or JSON');
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
      const res = await requestAPI<any>(
        `_show_datum?idx=${currentDatumIdx}`,
        'PUT',
      );
      refresh();
      setCurrentDatumId(res.id);
    } catch (e) {
      console.log(e);
    }

    setLoading(false);
  };

  const callUnmountAll = async () => {
    setLoading(true);

    try {
      refresh();
      await requestAPI<any>('_unmount_all', 'PUT');
      refresh();
      await pollRefresh();
      setCurrentDatumId('');
      setCurrentDatumIdx(-1);
      setNumDatums(0);
      setShouldShowCycler(false);
    } catch (e) {
      console.log(e);
    }

    setLoading(false);
  };

  return {
    loading,
    shouldShowCycler,
    currentDatumId,
    currentDatumIdx,
    setCurrentDatumIdx,
    numDatums,
    inputSpec,
    setInputSpec,
    callMountDatums,
    callUnmountAll,
    errorMessage,
  };
};
