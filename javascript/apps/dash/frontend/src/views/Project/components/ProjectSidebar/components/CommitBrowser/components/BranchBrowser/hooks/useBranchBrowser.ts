import {useCallback, useEffect, useRef, useState} from 'react';
import {useHistory} from 'react-router';

import useUrlState from '@dash-frontend/hooks/useUrlState';
import {repoRoute} from '@dash-frontend/views/Project/utils/routes';

const useBranchBrowser = () => {
  const {dagId, projectId, repoId} = useUrlState();
  const browserRef = useRef<HTMLDivElement>(null);
  const browserHistory = useHistory();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const handleBranchClick = useCallback(
    (branchId: string) => {
      browserHistory.push(
        repoRoute({
          branchId,
          dagId,
          projectId,
          repoId,
        }),
      );

      setIsMenuOpen(false);
    },
    [browserHistory, dagId, projectId, repoId],
  );
  const handleHeaderClick = useCallback(() => {
    setIsMenuOpen(!isMenuOpen);
  }, [isMenuOpen]);

  useEffect(() => {
    const handleClick = (evt: MouseEvent) => {
      if (!browserRef.current?.contains(evt.target as Node)) {
        return setIsMenuOpen(false);
      }
    };

    window.addEventListener('click', handleClick);

    return () => window.removeEventListener('click', handleClick);
  }, []);

  return {browserRef, isMenuOpen, handleBranchClick, handleHeaderClick};
};

export default useBranchBrowser;
