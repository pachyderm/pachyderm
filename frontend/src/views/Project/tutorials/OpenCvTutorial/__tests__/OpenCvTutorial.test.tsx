import {render} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import ProjectTutorial from '../../ProjectTutorial';

describe('Open CV Tutorial', () => {
  beforeEach(() => {
    window.history.replaceState(
      {},
      '',
      '/project/8?view=eyJ0dXRvcmlhbElkIjoib3Blbi1jdiJ9',
    );
  });
  const Tutorial = withContextProviders(() => {
    return <ProjectTutorial />;
  });

  it('should start the tutorial from the url param', async () => {
    const {findByText} = render(<Tutorial />);
    const tutorialTitle = await findByText('Create a pipeline');
    expect(tutorialTitle).toBeInTheDocument();
  });
});
