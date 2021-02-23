import {render} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import ButtonLink from '../';

describe('ButtonLink', () => {
  it('should be disabled when passed disabled is true', () => {
    const clickFunc = jest.fn();

    const {getByText} = render(
      <ButtonLink disabled={true} onClick={clickFunc}>
        Test
      </ButtonLink>,
    );
    userEvent.click(getByText('Test'));
    expect(clickFunc).not.toHaveBeenCalled();
  });
  it('should be not disabled when passed disabled is false', () => {
    const clickFunc = jest.fn();

    const {getByText} = render(
      <ButtonLink disabled={false} onClick={clickFunc}>
        Test
      </ButtonLink>,
    );
    userEvent.click(getByText('Test'));
    expect(clickFunc).toHaveBeenCalledTimes(1);
  });
});
