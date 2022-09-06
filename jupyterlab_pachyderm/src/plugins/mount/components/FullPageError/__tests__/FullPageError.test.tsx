import React from 'react';
import {render} from '@testing-library/react';

import FullPageError from '../FullPageError';
import {ServerStatus} from 'plugins/mount/pollRepos';

describe('full page error', () => {
  it('should display the passed in error message', async () => {
    const status: ServerStatus = {
      code: 500,
      message: '500: There was an error retrieving repos.',
    };

    const {getByTestId} = render(<FullPageError status={status} />);
    const message = getByTestId('FullPageError__message');
    expect(message).toHaveTextContent(
      '500: There was an error retrieving repos.',
    );
  });
});
