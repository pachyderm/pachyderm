import React from 'react';
import {

    JupyterFrontEnd,
    JupyterFrontEndPlugin,
  } from '@jupyterlab/application';
  
  import {SplitPanel, Widget, TabPanel} from '@lumino/widgets';
  import {mountLogoIcon} from '../../utils/icons';

  import Pipeline from './components/Pipeline';
  import {PpsContext, PpsMetadata, PipelineSettings} from './types';
  import {ReactWidget} from '@jupyterlab/apputils';
  
  const PLUGIN_ID = 'jupyterlab-pachyderm:pipeline';
  
  const pipeline: JupyterFrontEndPlugin<void> = {
    id: PLUGIN_ID,
    autoStart: true,
    requires: [

    ],
    activate: (
      app: JupyterFrontEnd,

    ) => {

        var tabPanel = new TabPanel();
        tabPanel.id = "pachyderm-extension";
        tabPanel.title.icon = mountLogoIcon;

        var publishPanel = new SplitPanel();
        publishPanel.orientation = 'vertical';
        publishPanel.spacing = 0;
        
        publishPanel.title.caption = 'Pachyderm Pipeline';
        publishPanel.title.label = "Publish";
        publishPanel.id = 'pachyderm-pipeline';
        var pipelineWidget = new Widget();

        var blah = { defaultPipelineImage: ''};
        var context = undefined;
        var pipeline = ReactWidget.create(
              <Pipeline
                ppsContext={context}
                settings={blah}
                saveNotebookMetadata={saveNotebookMetadata}
                saveNotebookToDisk={saveNotebookToDisk}
              />
        );
        pipeline.addClass('pachyderm-mount-pipeline-wrapper');
        publishPanel.addWidget(pipeline);
        
        tabPanel.addWidget(publishPanel);
        console.log("titles",tabPanel.tabBar.titles);
        //app.shell.add(tabPanel, 'left', {rank: 100});

    },
  };
  
  function saveNotebookMetadata(metadata: PpsMetadata): void {};
  async function saveNotebookToDisk(): Promise<string | null> {
    return Promise.resolve("disk");
  }

  export default pipeline;