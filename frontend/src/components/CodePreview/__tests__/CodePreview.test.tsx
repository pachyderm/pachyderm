import {render, screen} from '@testing-library/react';
import React from 'react';

import CodePreview from '../CodePreview';

describe('Code Preview', () => {
  it('should support rendering from source', async () => {
    render(<CodePreview source="Hello World" />);

    expect(
      await screen.findByText((text) => text.includes('Hello World')),
    ).toBeInTheDocument();
  });

  it('should render different style modes', async () => {
    render(
      <CodePreview source="Modes" hideGutter hideLineNumbers fullHeight />,
    );

    const wrapper = screen.getByTestId('CodePreview__wrapper');

    expect(wrapper).toHaveClass('hideGutter');
    expect(wrapper).toHaveClass('hideLineNumbers');
    expect(wrapper).toHaveClass('fullHeight');
  });
});
