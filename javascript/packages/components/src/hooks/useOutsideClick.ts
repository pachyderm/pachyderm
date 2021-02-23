import {RefObject, useEffect} from 'react';

const containsClick = (e: Event) => (ref: RefObject<HTMLElement>) => {
  return ref.current && ref.current.contains(e.target as Node);
};

/**
 * Hook for detecting clicks outside of a DOM node.
 * This is useful when needing to execute a function when someone clicks away from an element.
 * @param ref RefObject(s) to the desired DOM node
 * @param handleOutsideClick Callback executed when the user has clicked away from the ref
 */

const useOutsideClick = (
  ref: RefObject<HTMLElement>,
  handleOutsideClick: (e: Event) => void,
) => {
  useEffect(() => {
    const handleClick = (e: Event) => {
      const detectClick = containsClick(e);

      if (!detectClick(ref)) {
        handleOutsideClick(e);
      }
    };

    document.addEventListener('mousedown', handleClick);

    return () => {
      document.removeEventListener('mousedown', handleClick);
    };
  }, [ref, handleOutsideClick]);
};

export default useOutsideClick;
