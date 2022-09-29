import {requestAPI} from '../../../../../handler';
import {useEffect, useState} from 'react';
import {ServerConnection} from '@jupyterlab/services';
import {
  CurrentDatumResponse,
  ListMountsResponse,
  MountDatumResponse,
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
};

export const useDatum = (
  showDatum: boolean,
  keepMounted: boolean,
  refresh: (path: string) => void,
  pollRefresh: () => Promise<void>,
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
  const [errorMessage, setErrorMessage] = useState('');

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
      setInputSpec(JSON.stringify(currentDatumInfo.input, null, 2));
    }
  }, [showDatum]);

  const callMountDatums = async () => {
    setLoading(true);
    setErrorMessage('');

    try {
      const res = await requestAPI<MountDatumResponse>('_mount_datums', 'PUT', {
        input: JSON.parse(inputSpec),
      });
      refresh('');
      setCurrIdx(0);
      setCurrDatum(res);
      setShouldShowCycler(true);
      setInputSpec(JSON.stringify(JSON.parse(inputSpec), null, 2));
    } catch (e) {
      console.log(e);
      if (e instanceof ServerConnection.ResponseError) {
        setErrorMessage('Bad data in input spec');
      } else if (e instanceof SyntaxError) {
        setErrorMessage('Poorly formatted json input spec');
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
  };
};
