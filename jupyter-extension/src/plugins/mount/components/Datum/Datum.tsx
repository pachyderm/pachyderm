import React, {useRef} from 'react';
import isVisible, {useDatum} from './hooks/useDatum';
import {caretLeftIcon, caretRightIcon} from '@jupyterlab/ui-components';
import {
  CrossInputSpec,
  CurrentDatumResponse,
  PfsInput,
} from 'plugins/mount/types';

type DatumProps = {
  open: (path: string) => void;
  pollRefresh: () => Promise<void>;
  currentDatumInfo?: CurrentDatumResponse;
  repoViewInputSpec: CrossInputSpec | PfsInput;
};

const placeholderText = `pfs:
  repo: images
  glob: /*
`;

const Datum: React.FC<DatumProps> = ({
  open,
  pollRefresh,
  currentDatumInfo,
  repoViewInputSpec,
}) => {
  const cyclerRef = useRef(null);

  const {
    loading,
    shouldShowCycler,
    shouldShowDownload,
    currDatum,
    inputSpec,
    setInputSpec,
    callMountDatums,
    callNextDatum,
    callPrevDatum,
    callDownloadDatum,
    errorMessage,
    initialInputSpec,
  } = useDatum(open, pollRefresh, repoViewInputSpec, currentDatumInfo);

  return (
    <div className="pachyderm-mount-datum-base">
      <span className="pachyderm-mount-datum-subheading">Test Datums</span>

      <div className="pachyderm-mount-datum-input-wrapper">
        <label className="pachyderm-mount-datum-label" htmlFor="inputSpec">
          Input spec
        </label>
        <textarea
          className="pachyderm-input"
          data-testid="Datum__inputSpecInput"
          style={{minHeight: '200px', resize: 'none'}}
          name="inputSpec"
          value={inputSpec}
          onChange={(e: any) => {
            setInputSpec(e.target.value);
          }}
          disabled={loading}
          placeholder={
            Object.keys(initialInputSpec).length === 0
              ? placeholderText
              : undefined
          }
        ></textarea>
        <span
          className="pachyderm-mount-datum-error"
          data-testid="Datum__errorMessage"
        >
          {errorMessage}
        </span>
        <span
          className="pachyderm-mount-datum-info"
          data-testid="Datum__infoMessage"
        >
          {!isVisible(cyclerRef) &&
            shouldShowCycler &&
            'Drag line below to show datum cycler'}
        </span>
        <div
          className="pachyderm-mount-datum-actions"
          style={{display: 'flex'}}
        >
          <button
            data-testid="Datum__loadDatums"
            className="pachyderm-button-link"
            onClick={callMountDatums}
            style={{padding: '0.5rem'}}
          >
            Load Datums
          </button>
          {shouldShowDownload && (
            <button
              data-testid="Datum__downloadDatum"
              className="pachyderm-button-link"
              onClick={callDownloadDatum}
              style={{padding: '0.5rem'}}
            >
              Download Datum
            </button>
          )}
        </div>
        {shouldShowCycler && (
          <div
            className="pachyderm-mount-datum-cycler"
            data-testid="Datum__cycler"
            ref={cyclerRef}
          >
            Datum
            <div style={{display: 'flex'}}>
              <button
                className="pachyderm-button-link"
                data-testid="Datum__cyclerLeft"
                disabled={currDatum.idx <= 0}
                onClick={callPrevDatum}
              >
                <caretLeftIcon.react
                  tag="span"
                  className="pachyderm-mount-datum-left"
                />
              </button>
              {'(' +
                (currDatum.idx + 1) +
                '/' +
                currDatum.num_datums +
                (currDatum.all_datums_received ? '' : '+') +
                ')'}
              <button
                className="pachyderm-button-link"
                data-testid="Datum__cyclerRight"
                disabled={
                  currDatum.idx >= currDatum.num_datums - 1 &&
                  currDatum.all_datums_received
                }
                onClick={callNextDatum}
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
