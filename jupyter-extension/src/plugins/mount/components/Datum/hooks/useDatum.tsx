import YAML from 'yaml';
import {JSONObject} from '@lumino/coreutils';
import {useEffect, useState, useMemo, RefObject} from 'react';
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
  inputSpec: string;
  setInputSpec: (input: string) => void;
  callMountDatums: () => Promise<void>;
  callNextDatum: () => Promise<void>;
  callPrevDatum: () => Promise<void>;
  callUnmountAll: () => Promise<void>;
  errorMessage: string;
  saveInputSpec: () => void;
  initialInputSpec: JSONObject;
};

export const useDatum = (
  showDatum: boolean,
  keepMounted: boolean,
  setKeepMounted: (keep: boolean) => void,
  open: (path: string) => void,
  pollRefresh: () => Promise<void>,
  repoViewInputSpec: CrossInputSpec | PfsInput,
  currentDatumInfo?: CurrentDatumResponse,
): useDatumResponse => {
  const [loading, setLoading] = useState(false);
  const [shouldShowCycler, setShouldShowCycler] = useState(false);
  const [currDatum, setCurrDatum] = useState<MountDatumResponse>({
    id: '',
    idx: -1,
    num_datums: 0,
    all_datums_received: false,
  });
  const [inputSpec, setInputSpec] = useState('');
  const [initialInputSpec, setInitialInputSpec] = useState({});
  const [errorMessage, setErrorMessage] = useState('');
  const [datumViewInputSpec, setDatumViewInputSpec] = useState<
    string | JSONObject
  >({});

  useEffect(() => {
    if (showDatum) {
      if (!keepMounted) {
        callUnmountAll();
      }

      // Executes when browser reloaded; resume at currently mounted datum
      if (keepMounted && currentDatumInfo) {
        setShouldShowCycler(true);
        setCurrDatum({
          id: '',
          idx: currentDatumInfo.idx,
          num_datums: currentDatumInfo.num_datums,
          all_datums_received: currentDatumInfo.all_datums_received,
        });
        setInputSpec(inputSpecObjToText(currentDatumInfo.input));
        setKeepMounted(false);
      }
      // Pre-populate input spec from mounted repos
      else {
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
      }
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

  const inputSpecTextToObj = (): JSONObject => {
    let spec = {};
    try {
      spec = JSON.parse(inputSpec);
    } catch (e) {
      if (e instanceof SyntaxError) {
        spec = YAML.parse(inputSpec);
      } else {
        throw e;
      }
    }
    return spec;
  };

  const inputSpecObjToText = (specObj: JSONObject): string => {
    if (Object.keys(specObj).length === 0) {
      return '';
    }

    try {
      JSON.parse(inputSpec);
      return JSON.stringify(specObj, null, 2);
    } catch {
      return YAML.stringify(specObj, null, 2);
    }
  };

  const callMountDatums = async () => {
    setLoading(true);
    setErrorMessage('This could take a few minutes...');
    setShouldShowCycler(false);

    try {
      const spec = inputSpecTextToObj();
      const res = await requestAPI<MountDatumResponse>('datums/_mount', 'PUT', {
        input: spec,
      });
      open('');
      setCurrDatum(res);
      setShouldShowCycler(true);
      setInputSpec(inputSpecObjToText(spec));
      setErrorMessage('');
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

  const callNextDatum = async () => {
    setLoading(true);
    setErrorMessage('');

    try {
      const res = await requestAPI<MountDatumResponse>('datums/_next', 'PUT');
      open('');
      setCurrDatum(res);
    } catch (e) {
      console.log(e);
      if (e instanceof ServerConnection.ResponseError) {
        setCurrDatum({
          id: currDatum.id,
          idx: currDatum.idx,
          num_datums: currDatum.num_datums,
          all_datums_received: true,
        });
        setErrorMessage('Reached final datum');
      }
    }

    setLoading(false);
  };

  const callPrevDatum = async () => {
    setLoading(true);
    setErrorMessage('');

    try {
      const res = await requestAPI<MountDatumResponse>('datums/_prev', 'PUT');
      open('');
      setCurrDatum(res);
    } catch (e) {
      console.log(e);
    }

    setLoading(false);
  };

  const callUnmountAll = async () => {
    setLoading(true);

    try {
      open('');
      await requestAPI<ListMountsResponse>('_unmount_all', 'PUT');
      open('');
      await pollRefresh();
      setCurrDatum({
        id: '',
        idx: -1,
        num_datums: 0,
        all_datums_received: false,
      });
      setShouldShowCycler(false);
    } catch (e) {
      console.log(e);
    }

    setErrorMessage('');
    setLoading(false);
  };

  return {
    loading,
    shouldShowCycler,
    currDatum,
    inputSpec,
    setInputSpec,
    callMountDatums,
    callNextDatum,
    callPrevDatum,
    callUnmountAll,
    errorMessage,
    saveInputSpec,
    initialInputSpec,
  };
};

export default function isVisible(ref: RefObject<HTMLDivElement>): boolean {
  const [isIntersecting, setIntersecting] = useState(false);
  const observer = useMemo(
    () =>
      new IntersectionObserver(([entry]) =>
        setIntersecting(entry.isIntersecting),
      ),
    [ref],
  );

  useEffect(() => {
    if (ref.current) {
      observer.observe(ref.current);
      // Remove the observer as soon as the component is unmounted
      return () => {
        observer.disconnect();
      };
    }
  });

  return isIntersecting;
}
