import {render} from '@testing-library/react';
import noop from 'lodash/noop';
import React from 'react';

import {withContextProviders, click} from '@dash-frontend/testHelpers';

import IntroductionModalComponent from '../IntroductionModal';

describe('IntroductionModal', () => {
  const IntroductionModal = withContextProviders(({projectId}) => {
    return <IntroductionModalComponent projectId={projectId} onClose={noop} />;
  });

  afterEach(() => {
    localStorage.removeItem('pachyderm-console-6');
  });

  it('should set an active tutorial on confirm', async () => {
    const {findByTestId} = render(<IntroductionModal projectId="6" />);
    expect(localStorage.getItem('pachyderm-console-6')).toBe(null);

    await click(await findByTestId('ModalFooter__confirm'));
    await click(await findByTestId('ModalFooter__confirm'));

    const settings = localStorage.getItem('pachyderm-console-6');
    expect(settings).not.toBeNull();
    expect(JSON.parse(settings || '')['active_tutorial']).toBe(
      'image-processing',
    );
  });
});
