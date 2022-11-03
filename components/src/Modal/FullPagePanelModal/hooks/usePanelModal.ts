import {useContext} from 'react';

import ModalContext from '../contexts/ModalContext';

const usePanelModal = () => {
  const {show, onHide, onShow, leftOpen, setLeftOpen} =
    useContext(ModalContext);

  return {
    show,
    leftOpen,
    setLeftOpen,
    onHide,
    onShow,
  };
};

export default usePanelModal;
