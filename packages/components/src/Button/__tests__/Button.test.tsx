import {render} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {BrowserRouter} from 'react-router-dom';

import {Button} from '../';

describe('Button', () => {
  it('should be disabled when passed disabled is true', () => {
    const clickFunc = jest.fn();

    const {getByText} = render(
      <Button disabled={true} onClick={clickFunc}>
        Test
      </Button>,
    );
    userEvent.click(getByText('Test'));
    expect(clickFunc).not.toHaveBeenCalled();
  });

  it('should be not disabled when passed disabled is false', () => {
    const clickFunc = jest.fn();

    const {getByText} = render(
      <Button disabled={false} onClick={clickFunc}>
        Test
      </Button>,
    );
    userEvent.click(getByText('Test'));
    expect(clickFunc).toHaveBeenCalledTimes(1);
  });

  it('should render an anchor given href', () => {
    const {getByText} = render(<Button href="/cool">Test</Button>);
    const link = getByText('Test');

    expect(link).toHaveAttribute('href', '/cool');
  });

  it('should render a router link given to', () => {
    const {getByText} = render(
      <BrowserRouter>
        <Button to="/cool">Test</Button>
      </BrowserRouter>,
    );
    const link = getByText('Test');

    expect(link).toHaveAttribute('href', '/cool');
  });
});
