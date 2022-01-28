import {useState} from 'react';

const useImageProcessing = () => {
  const [exitSurveyOpen, setExitSurveyOpen] = useState(false);

  const onTutorialComplete = () => {
    setExitSurveyOpen(true);
  };

  return {
    exitSurveyOpen,
    onTutorialComplete,
  };
};

export default useImageProcessing;
