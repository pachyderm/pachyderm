import {useCallback, RefObject} from 'react';

import {Keys} from 'lib/types';

import useTabPanel from './useTabPanel';

const useTab = (id: string, ref: RefObject<HTMLAnchorElement>) => {
  const {isShown, selectTab, setActiveTabId} = useTabPanel(id);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === Keys.Right) {
        const nextTab = ref.current?.parentElement?.nextSibling?.firstChild;

        if (nextTab) {
          const nextId = (nextTab as HTMLElement).id;
          setActiveTabId(nextId);
          (nextTab as HTMLElement).focus();
        } else {
          const firstTab =
            ref.current?.parentElement?.parentElement?.firstChild?.firstChild;

          if (firstTab) {
            const nextId = (firstTab as HTMLElement).id;
            setActiveTabId(nextId);
            (firstTab as HTMLElement).focus();
          }
        }
      }

      if (e.key === Keys.Left) {
        const previousTab =
          ref.current?.parentElement?.previousSibling?.firstChild;

        if (previousTab) {
          const nextId = (previousTab as HTMLElement).id;
          setActiveTabId(nextId);
          (previousTab as HTMLElement).focus();
        } else {
          const lastTab =
            ref.current?.parentElement?.parentElement?.lastChild?.firstChild;

          if (lastTab) {
            const nextId = (lastTab as HTMLElement).id;
            setActiveTabId(nextId);
            (lastTab as HTMLElement).focus();
          }
        }
      }

      if (e.key === Keys.Down) {
        const selectedPanel = document.querySelector(
          `[aria-labelledby="${id}"][role="tabpanel"]`,
        );

        if (selectedPanel) {
          (selectedPanel as HTMLElement).focus();
        }
      }
    },
    [id, ref, setActiveTabId],
  );

  return {isShown, selectTab, handleKeyDown};
};

export default useTab;
