import React, {useEffect} from 'react';
import {closeIcon} from '@jupyterlab/ui-components';
import {useDatum} from './hooks/useDatum';
import {caretLeftIcon, caretRightIcon} from '@jupyterlab/ui-components';

type DatumProps = {
  showDatum: boolean;
  setShowDatum: (shouldShow: boolean) => void;
  keepMounted: boolean;
  setKeepMounted: (keep: boolean) => void;
  currentDatumInfo: any;
  refresh: () => void;
  pollRefresh: () => Promise<void>;
};

const placeholderText = `{
  "pfs": {
    "repo": "images",
    "branch": "dev",
    "glob": "/*",
  }
}`;

const Datum: React.FC<DatumProps> = ({
  showDatum,
  setShowDatum,
  keepMounted,
  setKeepMounted,
  currentDatumInfo,
  refresh,
  pollRefresh,
}) => {
  const {
    loading,
    shouldShowCycler,
    setShouldShowCycler,
    currentDatumId,
    currentDatumIdx,
    setCurrentDatumIdx,
    numDatums,
    setNumDatums,
    inputSpec,
    setInputSpec,
    callMountDatums,
    callUnmountAll,
  } = useDatum(showDatum, refresh, pollRefresh);

  useEffect(() => {
    if (showDatum && !keepMounted) {
      callUnmountAll();
    }
    if (keepMounted) {
      setShouldShowCycler(true);
      setCurrentDatumIdx(currentDatumInfo['curr_idx']);
      setNumDatums(currentDatumInfo['num_datums']);
      setInputSpec(JSON.stringify(currentDatumInfo['input'], null, 2));
    }
  }, [showDatum]);

  return (
    <div className="pachyderm-mount-datum-base">
      <div className="pachyderm-mount-datum-back">
        <button
          data-testid="Datum__back"
          className="pachyderm-button-link"
          onClick={async () => {
            await callUnmountAll();
            setKeepMounted(false);
            setShowDatum(false);
          }}
        >
          Back{' '}
          <closeIcon.react
            tag="span"
            className="pachyderm-mount-icon-padding"
          />
        </button>
      </div>

      <span className="pachyderm-mount-datum-subheading">Test Datums</span>

      <div className="pachyderm-mount-datum-input-wrapper">
        <label className="pachyderm-mount-datum-label" htmlFor="inputSpec">
          Input spec
        </label>
        <textarea
          className="pachyderm-input"
          data-testid="Datum__inputSpecInput"
          style={{minHeight: '200px'}}
          name="inputSpec"
          value={inputSpec}
          onInput={(e: any) => {
            setInputSpec(e.target.value);
          }}
          disabled={loading}
          placeholder={placeholderText}
        ></textarea>
        <button
          data-testid="Datum__mountDatums"
          className="pachyderm-button-link"
          onClick={callMountDatums}
          style={{padding: '0.5rem'}}
        >
          Mount Datums
        </button>
        {shouldShowCycler && (
          <div className="pachyderm-mount-datum-cycler">
            Datum
            <div style={{display: 'flex'}}>
              <button
                className="pachyderm-button-link"
                disabled={currentDatumIdx <= 0}
                onClick={() => {
                  if (currentDatumIdx >= 1) {
                    setCurrentDatumIdx(currentDatumIdx - 1);
                  }
                }}
              >
                <caretLeftIcon.react
                  tag="span"
                  className="pachyderm-mount-datum-left"
                />
              </button>
              {'(' + (currentDatumIdx + 1) + '/' + numDatums + ')'}
              <button
                className="pachyderm-button-link"
                disabled={currentDatumIdx >= numDatums - 1}
                onClick={() => {
                  if (currentDatumIdx < numDatums - 1) {
                    setCurrentDatumIdx(currentDatumIdx + 1);
                  }
                }}
              >
                <caretRightIcon.react
                  tag="span"
                  className="pachyderm-mount-datum-right"
                />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Datum;
