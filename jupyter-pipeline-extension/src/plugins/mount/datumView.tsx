import React from 'react';
import {SplitPanel, Widget} from '@lumino/widgets';

import Pipeline from './components/Pipeline/Pipeline';
import {ReactWidget} from '@jupyterlab/apputils';
import {PpsMetadata} from './types';
import {
    ListMountsResponse,
  } from './types';

export default function CreateDatumView(): Widget {
    var panel = new SplitPanel();
    panel.orientation = 'vertical';
    panel.spacing = 0;
    panel.title.caption = 'Test';
    panel.title.label = 'Test';

    panel.id = 'test';


    //panel.addWidget(myWidget);
    
    return panel
}