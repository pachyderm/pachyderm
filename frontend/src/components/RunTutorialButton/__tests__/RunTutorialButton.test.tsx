import {render} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import RunTutorialButtonComponent from '../RunTutorialButton';

describe('RunTutorialButton', () => {
  const RunTutorialButton = withContextProviders(() => {
    return <RunTutorialButtonComponent />;
  });

  afterEach(() => {
    window.localStorage.removeItem('pachyderm-console-default');
  });

  it('should start the tutorial when clicked', async () => {
    expect(
      window.localStorage.getItem('pachyderm-console-default'),
    ).toBeFalsy();

    const {getByText} = render(<RunTutorialButton />);

    getByText('Run Tutorial').click();

    expect(window.location.pathname).toBe('/lineage/default');
    expect(window.localStorage.getItem('pachyderm-console-default')).toEqual(
      `{"active_tutorial":"image-processing"}`,
    );
  });

  it('should say resume when tutorial progress is available', async () => {
    window.localStorage.setItem(
      'pachyderm-console-default',
      '{"tutorial_progress":{"image-processing":{"story":3,"task":0}}}',
    );

    const {findByText} = render(<RunTutorialButton />);

    expect(await findByText('Resume Tutorial')).toBeInTheDocument();
  });
});
