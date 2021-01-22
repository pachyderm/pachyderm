import {act} from '@testing-library/react';
import userEvent, {TargetElement, ITypeOpts} from '@testing-library/user-event';

export const click: typeof userEvent.click = (...args) => {
  act(() => userEvent.click(...args));
};

export const paste: typeof userEvent.paste = (...args) => {
  act(() => userEvent.paste(...args));
};

export const type = async (
  element: TargetElement,
  text: string,
  userOpts?: ITypeOpts,
) => {
  await act(async () => {
    await userEvent.type(element, text, userOpts);
  });
};
