import noop from 'lodash/noop';
import {createContext} from 'react';
import {ModalProps as BootstrapModalProps} from 'react-bootstrap/Modal';
export interface IModalContext
  extends Omit<BootstrapModalProps, 'show' | 'onHide' | 'onShow'> {
  show: boolean;
  leftOpen: boolean;
  rightOpen: boolean;
  hideLeftPanel: boolean;
  setLeftOpen: (val: boolean) => void;
  setRightOpen: (val: boolean) => void;
  onHide?: () => void;
  onShow?: () => void;
}

export default createContext<IModalContext>({
  show: false,
  leftOpen: false,
  rightOpen: false,
  hideLeftPanel: false,
  setLeftOpen: noop,
  setRightOpen: noop,
  onHide: noop,
  onShow: noop,
});
