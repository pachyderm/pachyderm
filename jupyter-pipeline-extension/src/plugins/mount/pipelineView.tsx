import React from 'react';
import {SplitPanel, Widget} from '@lumino/widgets';

import Pipeline from './components/Pipeline/Pipeline';
import {ReactWidget} from '@jupyterlab/apputils';
import {PpsMetadata} from './types';
import {
    ListMountsResponse,
  } from './types';

export default function CreatePipelineView(): Widget {
    var panel = new SplitPanel();
    panel.orientation = 'vertical';
    panel.spacing = 0;
    panel.title.caption = 'Publish';
    panel.title.label = 'Publish';

    panel.id = 'publish';

    var blah = { defaultPipelineImage: ''};

    var saveNotebookMetadata = (metadata: PpsMetadata): void => {};
    var saveNotebookToDisk = async (): Promise<string | null> => {
      return Promise.resolve("disk");
    };

    var context = undefined;
    var pipeline = ReactWidget.create(
          <Pipeline
            ppsContext={context}
            settings={blah}
            saveNotebookMetadata={saveNotebookMetadata}
            saveNotebookToDisk={saveNotebookToDisk}
          />
    );
    panel.addWidget(pipeline);
    
    return panel
}