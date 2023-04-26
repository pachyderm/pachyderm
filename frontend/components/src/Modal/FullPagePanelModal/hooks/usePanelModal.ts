import {useContext} from 'react';

import ModalContext from '../contexts/ModalContext';

const usePanelModal = () => {
  const {show, onHide, onShow, leftOpen, rightOpen, setLeftOpen, setRightOpen} =
    useContext(ModalContext);

  return {
    show,
    leftOpen,
    rightOpen,
    setLeftOpen,
    setRightOpen,
    onHide,
    onShow,
  };
};

export default usePanelModal;
