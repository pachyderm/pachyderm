import {closeHoverTooltips} from '@codemirror/view';
import {EditorView} from 'codemirror';

export const closeToolTipsOnType = () =>
  EditorView.updateListener.of((update) => {
    if (
      update.docChanged &&
      update.transactions.some((tr) => tr.isUserEvent('input.type'))
    ) {
      update.view.dispatch({effects: closeHoverTooltips});
    }
  });
