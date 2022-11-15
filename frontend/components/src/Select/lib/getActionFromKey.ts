// https://github.com/microsoft/sonder-ui/blob/master/src/shared/utils.ts

import {Keys} from '../../lib/types';

import {MenuActions} from './types';

const getActionFromKey = (e: React.KeyboardEvent, menuOpen: boolean) => {
  const {key, altKey} = e;
  const openKeys = ['ArrowDown', 'ArrowUp', 'Enter', ' ', 'Home', 'End'];

  if (!menuOpen && openKeys.includes(key)) {
    return MenuActions.Open;
  }

  if (menuOpen) {
    if (key === Keys.Down && !altKey) {
      return MenuActions.Next;
    } else if (key === Keys.Up && altKey) {
      return MenuActions.CloseSelect;
    } else if (key === Keys.Up) {
      return MenuActions.Previous;
    } else if (key === Keys.Home) {
      return MenuActions.First;
    } else if (key === Keys.End) {
      return MenuActions.Last;
    } else if (key === Keys.PageUp) {
      return MenuActions.PageUp;
    } else if (key === Keys.PageDown) {
      return MenuActions.PageDown;
    } else if (key === Keys.Escape) {
      return MenuActions.Close;
    } else if (key === Keys.Enter) {
      return MenuActions.CloseSelect;
    } else if (key === Keys.Space) {
      return MenuActions.Space;
    }
  }

  // eslint-disable-next-line consistent-return
  return;
};

export default getActionFromKey;
