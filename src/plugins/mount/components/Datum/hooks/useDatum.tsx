import {requestAPI} from '../../../../../handler';
import {useEffect, useState} from 'react';

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
};

export const useDatum = (
  showDatum: boolean,
  refresh: () => void,
): useDatumResponse => {
  const [loading, setLoading] = useState(false);
  const [shouldShowCycler, setShouldShowCycler] = useState(false);
  const [currentDatumId, setCurrentDatumId] = useState('');
  const [currentDatumIdx, setCurrentDatumIdx] = useState(-1);
  const [numDatums, setNumDatums] = useState(-1);
  const [inputSpec, setInputSpec] = useState('');

  useEffect(() => {
    if (showDatum && currentDatumIdx !== -1) {
      callShowDatum();
    }
  }, [currentDatumIdx, showDatum]);

  // TODO: out still showing after unmount all
  // useEffect(() => {
  //   if (!showDatum) {
  //     callUnmountAll()
  //   }
  // }, [showDatum])

  const callMountDatums = async () => {
    setLoading(true);

    try {
      await callUnmountAll();
      const res = await requestAPI<any>('_mount_datums', 'PUT', {
        input: JSON.parse(inputSpec),
      });
      refresh();
      setCurrentDatumId(res.id);
      setCurrentDatumIdx(res.idx);
      setNumDatums(res.num_datums);
      setShouldShowCycler(true);
    } catch (e) {
      console.log(e);
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
      await requestAPI<any>('_unmount_all', 'PUT');
      refresh();
      setCurrentDatumId('');
      setCurrentDatumIdx(-1);
      setNumDatums(0);
      setShouldShowCycler(false);
    } catch (e) {
      console.log(e);
    }
    console.log('===> CALLING UMOUNT ALL');
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
  };
};
