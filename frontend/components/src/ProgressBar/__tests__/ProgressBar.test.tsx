import {render, waitFor} from '@testing-library/react';
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
    const {findByText, findByTestId} = render(<ProgressBarComponent />);

    const unvisitedStepWrapper = await findByTestId('ProgressBarStep__div');

    expect(unvisitedStepWrapper).not.toHaveClass('visited');

    const visitButton = await findByText('Visit');
    await click(visitButton);

    const visitedStepWrapper = await findByTestId('ProgressBarStep__div');

    expect(visitedStepWrapper).toHaveClass('visited');
  });

  it('should indicate a completed step', async () => {
    const {findByText, queryByTestId} = render(<ProgressBarComponent />);

    const incomplete = queryByTestId('ProgressBarStep__successCheckmark');

    expect(incomplete).toBeNull();

    const completeButton = await findByText('Complete');
    await click(completeButton);

    await waitFor(() =>
      expect(queryByTestId('ProgressBarStep__successCheckmark')).not.toBeNull(),
    );
  });

  it('should disable onClick when it has not been visited yet', async () => {
    const clickMock = jest.fn();
    const {findByText, findByTestId} = render(
      <ProgressBarComponent onClick={clickMock} />,
    );

    const unVisitedButton = await findByTestId('ProgressBarStep__group');
    await click(unVisitedButton);

    expect(clickMock).not.toHaveBeenCalled();

    const visitButton = await findByText('Visit');
    await click(visitButton);

    const visitedButton = await findByTestId('ProgressBarStep__group');

    await click(visitedButton);
    expect(clickMock).toHaveBeenCalled();
  });
});
