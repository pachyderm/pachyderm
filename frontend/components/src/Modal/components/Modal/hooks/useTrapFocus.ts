// this implementation is based off the micromodal library
// https://github.com/ghosh/Micromodal/tree/master

import noop from 'lodash/noop';
import React, {useEffect, useCallback} from 'react';

const FOCUSABLE_ELEMENTS = [
  'a[href]',
  'area[href]',
  'input:not([disabled]):not([type="hidden"]):not([aria-hidden])',
  'select:not([disabled]):not([aria-hidden])',
  'textarea:not([disabled]):not([aria-hidden])',
  'button:not([disabled]):not([aria-hidden]):not([tabindex^="-"])',
  'iframe',
  'object',
  'embed',
  '[contenteditable]',
  '[tabindex]:not([tabindex^="-"])',
];

const useTrapFocus = (
  modalRef: React.RefObject<HTMLDivElement>,
  onHide = noop,
) => {
  const getFocusableNodes = useCallback(() => {
    const nodes = modalRef.current?.querySelectorAll<HTMLElement>(
      FOCUSABLE_ELEMENTS.join(', '),
    );
    return nodes ? Array(...nodes) : [];
  }, [modalRef]);

  const tabEvent = useCallback(
    (event: KeyboardEvent) => {
      const focusableNodes = getFocusableNodes();

      if (focusableNodes.length === 0) return;

      const activeElement = document.activeElement as HTMLElement;
      const focusedItemIndex = activeElement
        ? focusableNodes.indexOf(activeElement)
        : 0;

      // backward loop to last element
      if (event.shiftKey && focusedItemIndex === 0) {
        focusableNodes[focusableNodes.length - 1].focus();
        return;
      }

      // forward loop to first element
      if (
        !event.shiftKey &&
        focusableNodes.length > 0 &&
        focusedItemIndex === focusableNodes.length - 1
      ) {
        focusableNodes[0].focus();
        return;
      }

      // tab backwards to next element
      if (event.shiftKey) {
        focusableNodes[focusedItemIndex - 1].focus();
        return;
      }

      // tab forwards to next element
      if (!event.shiftKey) {
        focusableNodes[focusedItemIndex + 1].focus();
        return;
      }
    },
    [getFocusableNodes],
  );

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (['Tab', 'Escape'].includes(event.key)) {
        event.preventDefault();
      }
      if (event.key === 'Escape') {
        onHide();
      }
      if (event.key === 'Tab') {
        tabEvent(event);
      }
    };

    window.addEventListener('keydown', onKeyDown);

    return () => {
      window.removeEventListener('keydown', onKeyDown);
    };
  }, [onHide, tabEvent]);

  return {};
};

export default useTrapFocus;
