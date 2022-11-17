import {render, waitFor} from '@testing-library/react';
import React from 'react';

import {Page} from '../';

describe('Page', () => {
  it('should change the document title', async () => {
    render(<Page title="Hello Page" />);

    await waitFor(() =>
      expect(document.title).toEqual('Hello Page - Pachyderm'),
    );
  });
});
