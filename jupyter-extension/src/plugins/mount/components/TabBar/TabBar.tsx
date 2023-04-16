import React from 'react';
// import {Display} from 'plugins/mount/types';
import {Display} from '../../types';
import {settingsIcon} from '@jupyterlab/ui-components';

type TabBarProps = {
  display?: Display;
  setDisplay: (display: Display) => void;
  reposStatus?: number;
};

const TabBar: React.FC<TabBarProps> = ({display, setDisplay, reposStatus}) => {
  const hasConnection = reposStatus !== 200;
  const tabClass = (target: Display): string => {
    if (display === target) {
      return 'pachyderm-tab-fg';
    }
    if (!hasConnection) {
      return 'pachyderm-tab-inactive';
    }
    return 'pachyderm-tab-bg';
  };
  return (
    <div id="tabs" className="pachyderm-tabbar">
      <div
        className={tabClass(Display.Explore)}
        style={{width: '30%'}}
        onClick={
          hasConnection ? async () => setDisplay(Display.Explore) : undefined
        }
      >
        <span className="pachyderm-tab-text">&#10122; Explore</span>
      </div>
      <div
        className={tabClass(Display.Test)}
        style={{width: '30%'}}
        onClick={
          hasConnection ? async () => setDisplay(Display.Test) : undefined
        }
      >
        <span className="pachyderm-tab-text">&#10123; Test</span>
      </div>
      <div
        className={tabClass(Display.Publish)}
        style={{width: '30%'}}
        onClick={
          hasConnection ? async () => setDisplay(Display.Publish) : undefined
        }
      >
        <span className="pachyderm-tab-text">&#10124; Publish</span>
      </div>
      <div
        className={tabClass(Display.Settings)}
        style={{width: '10%'}}
        onClick={
          hasConnection ? async () => setDisplay(Display.Settings) : undefined
        }
      >
        <settingsIcon.react
          tag="span"
          className="pachyderm-mount-icon-padding"
        />
      </div>
    </div>
  );
};

export default TabBar;
