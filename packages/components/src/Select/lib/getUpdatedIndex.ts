import {MenuActions} from './types';

const getUpdatedIndex = (
  max: number,
  action: MenuActions,
  current?: number,
) => {
  switch (action) {
    case MenuActions.First:
      return 0;
    case MenuActions.Last:
      return max;
    case MenuActions.Previous:
      return current !== undefined ? Math.max(0, current - 1) : max;
    case MenuActions.Next:
      return current !== undefined ? Math.min(max, current + 1) : 0;
    default:
      return current;
  }
};

export default getUpdatedIndex;
