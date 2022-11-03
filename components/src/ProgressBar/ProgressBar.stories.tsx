import React, {useCallback, useState} from 'react';

import ProgressBar, {useProgressBar} from 'ProgressBar';

export default {title: 'Progress Bar'};

const Buttons: React.FC = () => {
  const {visitStep, completeStep} = useProgressBar();
  const [currentStep, setCurrentStep] = useState(1);
  const complete = useCallback(() => {
    completeStep(currentStep.toString());
    setCurrentStep((prevValue) => prevValue + 1);
  }, [completeStep, currentStep]);
  const visit = useCallback(
    () => visitStep(currentStep.toString()),
    [visitStep, currentStep],
  );

  return (
    <>
      <button onClick={complete}>Complete</button>
      <button onClick={visit}>Visit</button>
    </>
  );
};

export const Default = () => {
  return (
    <ProgressBar>
      <ProgressBar.Container>
        <ProgressBar.Step id="1" nextStepID="2">
          <span>Step 1</span>
        </ProgressBar.Step>
        <ProgressBar.Step id="2" nextStepID="3">
          <span>Step 2</span>
        </ProgressBar.Step>
        <ProgressBar.Step id="3">
          <span>Step 3</span>
        </ProgressBar.Step>
      </ProgressBar.Container>
      <Buttons />
    </ProgressBar>
  );
};

export const Vertical = () => {
  return (
    <div style={{width: '21.5rem'}}>
      <ProgressBar vertical>
        <ProgressBar.Container>
          <ProgressBar.Step id="1" nextStepID="2">
            <span>
              Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
              eiusmod tempor incididunt ut labore et dolore magna aliqua.
            </span>
          </ProgressBar.Step>
          <ProgressBar.Step id="2" nextStepID="3">
            <span>
              Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
              eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut
              enim ad minim veniam, quis nostrud exercitation ullamco laboris
              nisi ut aliquip ex ea commodo consequat.
            </span>
          </ProgressBar.Step>
          <ProgressBar.Step id="3">
            <span>
              Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
              eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut
              enim ad minim veniam, quis nostrud exercitation ullamco laboris
              nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in
              reprehenderit in voluptate velit esse cillum dolore eu fugiat
              nulla pariatur. Excepteur sint occaecat cupidatat non proident,
              sunt in culpa qui officia deserunt mollit anim id est laborum.
            </span>
          </ProgressBar.Step>
        </ProgressBar.Container>
        <Buttons />
      </ProgressBar>
    </div>
  );
};
