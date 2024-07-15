import {useEffect, useState} from 'react';

const useBreakpoint = (breakpoint: number) => {
  const [isMobile, setMobile] = useState(window.innerWidth <= breakpoint);

  useEffect(() => {
    const handleResize = () => {
      return setMobile(window.innerWidth <= breakpoint);
    };

    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, [breakpoint]);

  return isMobile;
};

export default useBreakpoint;
