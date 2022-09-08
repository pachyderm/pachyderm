import noop from 'lodash/noop';
import {createContext} from 'react';
import {ModalProps as BootstrapModalProps} from 'react-bootstrap/Modal';
export interface IModalContext
  extends Omit<BootstrapModalProps, 'show' | 'onHide' | 'onShow'> {
  show: boolean;
  onHide?: () => void;
  onShow?: () => void;
}

export default createContext<IModalContext>({
  show: false,
  onHide: noop,
  onShow: noop,
});
