import {render, screen} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import HeaderButtonsComponent from '../HeaderButtons';

describe('HeaderButtons', () => {
  const HeaderButtons = withContextProviders(() => {
    return <HeaderButtonsComponent />;
  });

  afterEach(() => {
    window.localStorage.removeItem('pachyderm-console-Default');
  });

  /* eslint-disable jest/no-disabled-tests */
  /* Tutorial is temporarily disabled because of "Project" Console Support */
  it.skip('should start the tutorial when clicked', async () => {
    expect(
      window.localStorage.getItem('pachyderm-console-default'),
    ).toBeFalsy();

    render(<HeaderButtons />);

    screen.getAllByText('Learn Pachyderm')[0].click();

    expect(await screen.findByText('Start Tutorial')).toBeInTheDocument();

    screen.getByText('Image processing at scale with pachyderm').click();

    expect(window.location.pathname).toBe('/lineage/default');
    expect(window.localStorage.getItem('pachyderm-console-default')).toBe(
      `{"active_tutorial":"image-processing"}`,
    );
  });

  /* Tutorial is temporarily disabled because of "Project" Console Support */
  it.skip('should show progress dots when tutorial progress is available', async () => {
    window.localStorage.setItem(
      'pachyderm-console-default',
      '{"tutorial_progress":{"image-processing":{"story":2,"task":1}}}',
    );

    render(<HeaderButtons />);

    screen.getAllByText('Learn Pachyderm')[0].click();

    expect(screen.queryByText('Start Tutorial')).not.toBeInTheDocument();
    expect(
      await screen.findByTestId('StoryProgressDots__progress'),
    ).toBeInTheDocument();
    expect(await screen.findByText('Story 4 of 5')).toBeInTheDocument();
  });

  /* Tutorial is temporarily disabled because of "Project" Console Support */
  it.skip('should allow the user to delete tutorial progress', async () => {
    window.localStorage.setItem(
      'pachyderm-console-default',
      '{"tutorial_progress":{"image-processing":{"story":3,"task":0}}}',
    );

    render(<HeaderButtons />);

    screen.getAllByText('Learn Pachyderm')[0].click();
    screen.getByTestId('TutorialItem__deleteProgress').click();
    (await screen.findByText('Delete Tutorial Content')).click();

    expect(window.localStorage.getItem('pachyderm-console-default')).toBe(
      `{"tutorial_progress":{"image-processing":null},"active_tutorial":null}`,
    );
  });
  /* eslint-enable jest/no-disabled-tests */
});
