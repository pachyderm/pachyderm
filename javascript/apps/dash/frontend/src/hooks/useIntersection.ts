import {useState, useEffect} from 'react';

const useIntersection = (
  element: HTMLElement | null,
  container: HTMLElement | null,
) => {
  const [isIntersected, setIsIntersected] = useState(false);

  // isIntersected is true when any part of "element" is outside of "container"
  useEffect(() => {
    let observer: IntersectionObserver | null = null;
    if (element && container) {
      observer = new IntersectionObserver(
        ([entries]) => {
          setIsIntersected(
            entries.boundingClientRect.top < entries.intersectionRect.top,
          );
        },
        {root: container, rootMargin: '0px', threshold: 1},
      );
      observer.observe(element);
    }
    return () => {
      if (observer && element) {
        observer.unobserve(element);
      }
    };
  });
  return isIntersected;
};

export default useIntersection;
