import {render} from '@testing-library/react';
import React from 'react';
import {BrowserRouter} from 'react-router-dom';

import {click} from '@dash-frontend/testHelpers';

import {Button} from '../';

describe('Button', () => {
  it('should be disabled when passed disabled is true', async () => {
    const clickFunc = jest.fn();

    const {getByText} = render(
      <Button disabled={true} onClick={clickFunc}>
        Test
      </Button>,
    );
    await click(getByText('Test'));
    expect(clickFunc).not.toHaveBeenCalled();
  });

  it('should be not disabled when passed disabled is false', async () => {
    const clickFunc = jest.fn();

    const {getByText} = render(
      <Button disabled={false} onClick={clickFunc}>
        Test
      </Button>,
    );
    await click(getByText('Test'));
    expect(clickFunc).toHaveBeenCalledTimes(1);
  });

  it('should render an anchor given href', () => {
    const {getByRole} = render(<Button href="/cool">Test</Button>);
    const link = getByRole('link');

    expect(link).toHaveAttribute('href', '/cool');
  });

  it('should render a router link given to', () => {
    const {getByRole} = render(
      <BrowserRouter>
        <Button to="/cool">Test</Button>
      </BrowserRouter>,
    );
    const link = getByRole('link');

    expect(link).toHaveAttribute('href', '/cool');
  });
});
