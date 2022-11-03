import ModalBody from './components/Body';
import {LeftPanel, RightPanel} from './components/SidePanel';
import FullPagePanelModal from './FullPagePanelModal';

export default Object.assign(FullPagePanelModal, {
  Body: ModalBody,
  LeftPanel: LeftPanel,
  RightPanel: RightPanel,
});
