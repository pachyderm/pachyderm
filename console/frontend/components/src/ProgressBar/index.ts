import ProgressBarContainer from './components/ProgressBarContainer';
import ProgressBarStep from './components/ProgressBarStep';
import useProgressBar from './hooks/useProgressBar';
import ProgressBar from './ProgressBar';

export {useProgressBar};

export default Object.assign(ProgressBar, {
  Step: ProgressBarStep,
  Container: ProgressBarContainer,
});
