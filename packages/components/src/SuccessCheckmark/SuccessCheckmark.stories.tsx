import React, {useEffect, useState} from 'react';

import {SuccessCheckmark} from './';

/* eslint-disable-next-line import/no-anonymous-default-export */
export default {title: 'SuccessCheckmark'};

export const Default = () => {
  const [show, setShow] = useState(false);

  useEffect(() => {
    setTimeout(() => {
      setShow(true);
    }, 1000);
  }, []);

  return <SuccessCheckmark show={show} />;
};
