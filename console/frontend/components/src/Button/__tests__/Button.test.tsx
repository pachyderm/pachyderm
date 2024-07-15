import {render, screen} from '@testing-library/react';
import React from 'react';
import {BrowserRouter} from 'react-router-dom';

import {click} from '@dash-frontend/testHelpers';

import {Button} from '../';

describe('Button', () => {
  it('should be disabled when passed disabled is true', async () => {
    const clickFunc = jest.fn();

    render(
      <Button disabled={true} onClick={clickFunc}>
        Test
      </Button>,
    );
    await click(screen.getByText('Test'));
    expect(clickFunc).not.toHaveBeenCalled();
  });

  it('should be not disabled when passed disabled is false', async () => {
    const clickFunc = jest.fn();

    render(
      <Button disabled={false} onClick={clickFunc}>
        Test
      </Button>,
    );
    await click(screen.getByText('Test'));
    expect(clickFunc).toHaveBeenCalledTimes(1);
  });

  it('should render an anchor given href', () => {
    render(<Button href="/cool">Test</Button>);
    const link = screen.getByRole('link');

    expect(link).toHaveAttribute('href', '/cool');
  });

  it('should render a router link given to', () => {
    render(
      <BrowserRouter>
        <Button to="/cool">Test</Button>
      </BrowserRouter>,
    );
    const link = screen.getByRole('link');

    expect(link).toHaveAttribute('href', '/cool');
  });
});
