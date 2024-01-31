import {useEffect, useRef, useState} from 'react';

const useCurrentWidthAndHeight = () => {
  const ref = useRef<HTMLDivElement>(null);

  const [width, setWidth] = useState(0);
  const [height, setHeight] = useState(0);

  useEffect(() => {
    if (!ref.current) {
      return;
    }

    const resizeObserver = new ResizeObserver(() => {
      if (ref.current?.offsetWidth !== width) {
        setWidth(ref.current?.offsetWidth || 0);
      }
      if (ref.current?.offsetHeight !== height) {
        setHeight(ref.current?.offsetHeight || 0);
      }
    });

    resizeObserver.observe(ref.current);

    return () => {
      resizeObserver.disconnect();
    };
  }, [height, ref, width]);

  return {
    ref,
    width,
    height,
  };
};

export default useCurrentWidthAndHeight;
