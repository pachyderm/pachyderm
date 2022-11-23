import {useState, useRef, useEffect} from 'react';

const useStickyState = () => {
  const [elementIsStuck, setElementIsStuck] = useState(false);
  const stickyRef = useRef(null);

  useEffect(() => {
    const currentRef = stickyRef.current;
    let observer: IntersectionObserver | null = null;
    if (currentRef) {
      observer = new IntersectionObserver(
        ([e]) => {
          setElementIsStuck(
            e.intersectionRatio < 1 &&
              e.boundingClientRect.width > e.intersectionRect.width,
          );
        },
        {threshold: [1]},
      );
      observer.observe(currentRef);
      return () => {
        if (observer && currentRef) observer.unobserve(currentRef);
      };
    }
  }, []);

  return {
    stickyRef,
    elementIsStuck,
  };
};

export default useStickyState;
