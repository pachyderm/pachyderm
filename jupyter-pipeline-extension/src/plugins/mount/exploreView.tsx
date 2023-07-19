import React from 'react';
import {SplitPanel, Widget} from '@lumino/widgets';

import Pipeline from './components/Pipeline/Pipeline';
import {ReactWidget} from '@jupyterlab/apputils';
import SortableList from './components/SortableList/SortableList';

import {
    ListMountsResponse,
  } from './types';

export default function CreateExplore(): Widget {
    var panel = new SplitPanel();
    panel.orientation = 'vertical';
    panel.spacing = 0;
    panel.title.caption = 'Explore';
    panel.title.label = 'Explore';

    panel.id = 'explore';

    
    var open = (path: string): void => {}; //TODO See open method in mount.tsx
    var updateData = (data: ListMountsResponse): void => {};

    var mountedList = ReactWidget.create(
            <div className="pachyderm-mount-base">
              <SortableList
                open={open}
                items={[]}
                updateData={updateData}
                mountedItems={[]}
                type={'mounted'}
                projects={[]}
              />
            </div>
      );
    mountedList.addClass('pachyderm-mount-react-wrapper');

    var unmountedList = ReactWidget.create(

                <div className="pachyderm-mount-base">
                  <div className="pachyderm-mount-base-title">
                    Unmounted Repositories
                  </div>
                  <SortableList
                    open={open}
                    items={[]}
                    updateData={updateData}
                    mountedItems={[]}
                    type={'unmounted'}
                    projects={[]}
                  />
                </div>
      );
      unmountedList.addClass('pachyderm-mount-react-wrapper');

    panel.addWidget(mountedList);
    //panel.addWidget(this._unmountedList);
    //panel.addWidget(this._mountBrowser);

    return panel;
}