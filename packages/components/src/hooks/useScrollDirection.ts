import throttle from 'lodash/throttle';
import {useState, useEffect} from 'react';

export enum ScrollDirections {
  UP,
  DOWN,
  NONE,
}

const useScrollDirection = (pollingInterval = 200) => {
  const [position, setPostion] = useState({
    scrollY: document.body.getBoundingClientRect().top,
    scrollDirection: ScrollDirections.NONE,
  });
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    const listener = () => {
      if (loaded) {
        setPostion((prev) => {
          const boundingClientTop = -document.body.getBoundingClientRect().top;
          const currentScrollY = boundingClientTop <= 0 ? 0 : boundingClientTop;
          let direction = prev.scrollDirection;
          if (
            !(
              prev.scrollY === 0 &&
              prev.scrollDirection === ScrollDirections.NONE
            )
          ) {
            if (prev.scrollY > currentScrollY) {
              direction = ScrollDirections.UP;
            } else if (prev.scrollY < currentScrollY) {
              direction = ScrollDirections.DOWN;
            } else {
              direction = ScrollDirections.NONE;
            }
          }

          return {
            scrollY: currentScrollY,
            scrollDirection: direction,
          };
        });
      }
    };

    const throttleWrapper = throttle(listener, pollingInterval);
    window.addEventListener('scroll', throttleWrapper);
    return () => {
      window.removeEventListener('scroll', throttleWrapper);
    };
  }, [loaded, pollingInterval]);

  // This useEffect is needed to block the first scroll event that gets fired on page load.
  useEffect(() => {
    setLoaded(true);
  }, []);

  return position.scrollDirection;
};

export default useScrollDirection;
