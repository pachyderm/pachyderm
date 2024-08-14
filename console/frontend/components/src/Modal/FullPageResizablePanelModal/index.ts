import ModalBody from './components/Body';
import ModalPanel from './components/Panel/Panel';
import ModalPanelResizeHandle from './components/PanelResizeHandle/PanelResizeHandle';
import FullPageResizablePanelModal from './FullPageResizablePanelModal';

export default Object.assign(FullPageResizablePanelModal, {
  Body: ModalBody,
  Panel: ModalPanel,
  PanelResizeHandle: ModalPanelResizeHandle,
});
