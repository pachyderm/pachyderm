import React, {useState, useCallback} from 'react';

import ProgressBarContext from './contexts/ProgressBarContext';

const ProgressBar = ({
  children,
  vertical = false,
}: {
  children?: React.ReactNode;
  vertical?: boolean;
}) => {
  const [visited, setVisited] = useState<string[]>([]);
  const [completed, setCompleted] = useState<string[]>([]);

  const visit = useCallback(
    (id: string) => {
      if (!visited.includes(id)) {
        setVisited((prevState) => [...prevState, id]);
      }
    },
    [visited, setVisited],
  );

  const complete = useCallback(
    (id: string) => {
      if (!completed.includes(id)) {
        setCompleted((prevState) => [...prevState, id]);
      }
    },
    [setCompleted, completed],
  );

  const clear = () => {
    setCompleted([]);
    setVisited([]);
  };

  const contextValue = {
    visited,
    visit,
    completed,
    complete,
    clear,
    isVertical: vertical,
  };

  return (
    <ProgressBarContext.Provider value={contextValue}>
      {children}
    </ProgressBarContext.Provider>
  );
};

export default ProgressBar;
