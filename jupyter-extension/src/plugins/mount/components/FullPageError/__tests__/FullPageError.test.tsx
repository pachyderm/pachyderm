import React from 'react';
import {render} from '@testing-library/react';

import FullPageError from '../FullPageError';
import {HealthCheck} from 'plugins/mount/types';

describe('full page error', () => {
  it('should display the passed in error message', async () => {
    const healthCheck: HealthCheck = {
      status: 'UNHEALTHY',
      message: 'There was an error retrieving mounts.',
    };

    const {getByTestId} = render(<FullPageError healthCheck={healthCheck} />);
    const message = getByTestId('FullPageError__message');
    expect(message).toHaveTextContent('There was an error retrieving mounts.');
  });
});
