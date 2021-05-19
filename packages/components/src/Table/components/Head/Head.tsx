import classNames from 'classnames';
import React, {HTMLAttributes, useEffect, useRef, useState} from 'react';

import styles from './Head.module.css';

export interface HeadProps extends HTMLAttributes<HTMLTableSectionElement> {
  sticky?: boolean;
  screenReaderOnly?: boolean;
}

const Head: React.FC<HeadProps> = ({
  children,
  className,
  sticky,
  screenReaderOnly = false,
  ...rest
}) => {
  const [isStuck, setIsStuck] = useState(false);
  const ref = useRef<HTMLTableSectionElement>(null);

  useEffect(() => {
    const currentRef = ref.current;
    let observer: IntersectionObserver | null = null;
    if (!screenReaderOnly && currentRef) {
      observer = new IntersectionObserver(
        ([e]) => {
          // We are only concerned with the top value of the two bounding
          // rectangles, as the header is not "stuck" when the viewport
          // is constrainted by the x-axis.
          setIsStuck(e.boundingClientRect.top < e.intersectionRect.top);
        },
        {threshold: [1]},
      );
      observer.observe(currentRef);
      return () => {
        if (observer && currentRef) observer.unobserve(currentRef);
      };
    }
  }, [screenReaderOnly]);

  const classes = classNames(className, {
    [styles.sticky]: sticky,
    [styles.stuck]: isStuck,
  });

  return (
    <thead {...rest} className={classes} ref={ref}>
      {children}
    </thead>
  );
};

export default Head;
