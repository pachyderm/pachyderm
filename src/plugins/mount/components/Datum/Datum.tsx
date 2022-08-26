import React from 'react';
import {closeIcon} from '@jupyterlab/ui-components';
import { useDatum } from './hooks/useDatum';
import {caretLeftIcon, caretRightIcon} from '@jupyterlab/ui-components';

type DatumProps = {
  showDatum: boolean;
  setShowDatum: (shouldShow: boolean) => void;
  refresh: () => void;
};

const Datum: React.FC<DatumProps> = ({
  showDatum,
  setShowDatum,
  refresh,
}) => {
  const {
    loading,
    shouldShowCycler,
    currentDatumId,
    currentDatumIdx,
    setCurrentDatumIdx,
    numDatums,
    inputSpec,
    setInputSpec,
    callMountDatums,
    callShowDatum,
    callUnmountAll,
  } = useDatum(
    showDatum,
    refresh
  )

  return (
    <>
      <div className="pachyderm-mount-datum-base">
        <div className="pachyderm-mount-datum-back">
          <button
            data-testid="Datum__back"
            className="pachyderm-button-link"
            onClick={() => {
              callUnmountAll()
              setShowDatum(false)
            }}
          >
            Back{' '}
            <closeIcon.react
              tag="span"
              className="pachyderm-mount-icon-padding"
            />
          </button>
        </div>

        <div className="pachyderm-mount-datum-heading">
          Test Datums
          <span className="pachyderm-mount-datum-subheading">
            Input spec
            <input
              className="pachyderm-input"
              data-testid="Datum__inputSpecInput"
              name="pachd"
              value={inputSpec}
              onInput={(e: any) => {
                setInputSpec(e.target.value);
              }}
              disabled={loading}
              placeholder='{\n\t"pfs": {\n\t\t"repo": "images", \n\t\t"glob": "/*"\n\t}\n}'
            ></input>
            <button
              data-testid="Datum__mountDatums"
              className="pachyderm-button-link"
              onClick={callMountDatums}
            >
              Mount Datums
            </button>
          </span>

          {
            shouldShowCycler && (
              <div className="pachyderm-mount-datum-cycler-title">
                Datum
                <button
                  className="pachyderm-button-link"
                  disabled={currentDatumIdx <= 0}
                  onClick={() => {
                    if (currentDatumIdx >= 1) {
                      setCurrentDatumIdx(currentDatumIdx-1);
                    }
                  }}
                  >
                  <caretLeftIcon.react
                    tag="span"
                    className="pachyderm-mount-datum-left"
                  />
                </button>
                {"(" + (currentDatumIdx+1) + "/" + numDatums + ")"}
                <button
                  className="pachyderm-button-link"
                  disabled={currentDatumIdx >= numDatums-1}
                  onClick={() => {
                    if (currentDatumIdx < numDatums-1) {
                      setCurrentDatumIdx(currentDatumIdx+1);
                    }
                  }}
                  >
                  <caretRightIcon.react
                    tag="span"
                    className="pachyderm-mount-datum-right"
                  />
                </button>
              </div>
            )
          }
        </div>
      </div>
    </>
  )
}

export default Datum;