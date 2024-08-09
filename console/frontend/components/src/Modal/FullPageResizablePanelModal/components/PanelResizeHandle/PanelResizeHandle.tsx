import React from 'react';
import {PanelResizeHandle} from 'react-resizable-panels';

import styles from './PanelResizeHandle.module.css';

const ModalPanelResizeHandle = () => {
  return <PanelResizeHandle className={styles.handle} />;
};

export default ModalPanelResizeHandle;
