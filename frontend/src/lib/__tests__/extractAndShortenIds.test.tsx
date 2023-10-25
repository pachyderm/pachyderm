import {render, screen} from '@testing-library/react';

import extractAndShortenIds from '@dash-frontend/lib/extractAndShortenIds';

describe('extractAndShortenIds', () => {
  it('should extra and shorten a job id', () => {
    const testText = 'Error from job 7c6134fc957c489d8b522776d9eacc86';
    render(extractAndShortenIds(testText));

    expect(
      screen.queryByText('7c6134fc957c489d8b522776d9eacc86'),
    ).not.toBeInTheDocument();
    expect(screen.getByText('7c6134...')).toBeInTheDocument();
  });

  it('should extra and shorten a datum id', () => {
    const testText =
      'Error from datum 5bd54942c533f497a7e7704b23f2b0f905917dae71beebdea7003b18507ffb91';
    render(extractAndShortenIds(testText));

    expect(
      screen.queryByText(
        '5bd54942c533f497a7e7704b23f2b0f905917dae71beebdea7003b18507ffb91',
      ),
    ).not.toBeInTheDocument();
    expect(screen.getByText('5bd549...')).toBeInTheDocument();
  });
});
