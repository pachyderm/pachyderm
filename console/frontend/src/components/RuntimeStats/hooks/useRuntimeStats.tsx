import {useState} from 'react';

export const useRuntimeStats = () => {
  const [runtimeDetailsOpen, setRuntimeDetailsOpen] = useState(false);
  const [runtimeDetailsClosing, setRuntimeDetailsClosing] = useState(false);

  const toggleRunTimeDetailsOpen = () => {
    if (runtimeDetailsOpen) {
      setRuntimeDetailsClosing(true);
      // slide out animation
      setTimeout(() => {
        setRuntimeDetailsOpen(false);
        setRuntimeDetailsClosing(false);
      }, 300);
    } else {
      setRuntimeDetailsOpen(true);
    }
  };
  return {toggleRunTimeDetailsOpen, runtimeDetailsOpen, runtimeDetailsClosing};
};
