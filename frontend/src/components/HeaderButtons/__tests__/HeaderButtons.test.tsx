import {render} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import HeaderButtonsComponent from '../HeaderButtons';

describe('HeaderButtons', () => {
  const HeaderButtons = withContextProviders(() => {
    return <HeaderButtonsComponent />;
  });

  afterEach(() => {
    window.localStorage.removeItem('pachyderm-console-default');
  });

  it('should start the tutorial when clicked', async () => {
    expect(
      window.localStorage.getItem('pachyderm-console-default'),
    ).toBeFalsy();

    const {getByText, getAllByText, findByText} = render(<HeaderButtons />);

    getAllByText('Learn Pachyderm')[0].click();

    expect(await findByText('Start Tutorial')).toBeInTheDocument();

    getByText('Image processing at scale with pachyderm').click();

    expect(window.location.pathname).toBe('/lineage/default');
    expect(window.localStorage.getItem('pachyderm-console-default')).toEqual(
      `{"active_tutorial":"image-processing"}`,
    );
  });

  it('should show progress dots when tutorial progress is available', async () => {
    window.localStorage.setItem(
      'pachyderm-console-default',
      '{"tutorial_progress":{"image-processing":{"story":2,"task":1}}}',
    );

    const {getAllByText, findByText, findByTestId, queryByText} = render(
      <HeaderButtons />,
    );

    getAllByText('Learn Pachyderm')[0].click();

    expect(queryByText('Start Tutorial')).not.toBeInTheDocument();
    expect(
      await findByTestId('StoryProgressDots__progress'),
    ).toBeInTheDocument();
    expect(await findByText('Story 4 of 5')).toBeInTheDocument();
  });

  it('should allow the user to delete tutorial progress', async () => {
    window.localStorage.setItem(
      'pachyderm-console-default',
      '{"tutorial_progress":{"image-processing":{"story":3,"task":0}}}',
    );

    const {getByTestId, findByText, getAllByText} = render(<HeaderButtons />);

    getAllByText('Learn Pachyderm')[0].click();
    getByTestId('TutorialItem__deleteProgress').click();
    (await findByText('Delete Tutorial Content')).click();

    expect(window.localStorage.getItem('pachyderm-console-default')).toEqual(
      `{"tutorial_progress":{"image-processing":null},"active_tutorial":null}`,
    );
  });
});
