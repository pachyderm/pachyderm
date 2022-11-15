import {render} from '@testing-library/react';
import React from 'react';

import {click} from '@dash-frontend/testHelpers';

import {ButtonLink} from '../';

describe('ButtonLink', () => {
  it('should be disabled when passed disabled is true', async () => {
    const clickFunc = jest.fn();

    const {getByText} = render(
      <ButtonLink disabled={true} onClick={clickFunc}>
        Test
      </ButtonLink>,
    );
    await click(getByText('Test'));
    expect(clickFunc).not.toHaveBeenCalled();
  });
  it('should be not disabled when passed disabled is false', async () => {
    const clickFunc = jest.fn();

    const {getByText} = render(
      <ButtonLink disabled={false} onClick={clickFunc}>
        Test
      </ButtonLink>,
    );
    await click(getByText('Test'));
    expect(clickFunc).toHaveBeenCalledTimes(1);
  });
});
