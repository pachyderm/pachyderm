import {render, screen} from '@testing-library/react';
import React, {useCallback} from 'react';

import {click} from '@dash-frontend/testHelpers';

import ProgressBar, {useProgressBar} from '../';

describe('ProgressBar', () => {
  const Buttons: React.FC = () => {
    const {visitStep, completeStep} = useProgressBar();

    const complete = useCallback(() => completeStep('1'), [completeStep]);
    const visit = useCallback(() => visitStep('1'), [visitStep]);

    return (
      <>
        <button onClick={complete}>Complete</button>
        <button onClick={visit}>Visit</button>
      </>
    );
  };

  const ProgressBarComponent: React.FC<{onClick?: () => void}> = ({
    onClick,
  }) => {
    return (
      <ProgressBar>
        <ProgressBar.Container>
          <ProgressBar.Step onClick={onClick} id="1">
            Step 1
          </ProgressBar.Step>
        </ProgressBar.Container>
        <Buttons />;
      </ProgressBar>
    );
  };

  it('should indicate a visited step', async () => {
    render(<ProgressBarComponent />);

    const unvisitedStepWrapper = await screen.findByTestId(
      'ProgressBarStep__div',
    );

    expect(unvisitedStepWrapper).not.toHaveClass('visited');

    const visitButton = await screen.findByText('Visit');
    await click(visitButton);

    const visitedStepWrapper = await screen.findByTestId(
      'ProgressBarStep__div',
    );

    expect(visitedStepWrapper).toHaveClass('visited');
  });

  it('should indicate a completed step', async () => {
    render(<ProgressBarComponent />);

    const incomplete = screen.queryByTestId(
      'ProgressBarStep__successCheckmark',
    );

    expect(incomplete).not.toBeInTheDocument();

    const completeButton = await screen.findByText('Complete');
    await click(completeButton);

    await screen.findByTestId('ProgressBarStep__successCheckmark');
  });

  it('should disable onClick when it has not been visited yet', async () => {
    const clickMock = jest.fn();
    render(<ProgressBarComponent onClick={clickMock} />);

    const unVisitedButton = await screen.findByTestId('ProgressBarStep__group');
    await click(unVisitedButton);

    expect(clickMock).not.toHaveBeenCalled();

    const visitButton = await screen.findByText('Visit');
    await click(visitButton);

    const visitedButton = await screen.findByTestId('ProgressBarStep__group');

    await click(visitedButton);
    expect(clickMock).toHaveBeenCalled();
  });
});
