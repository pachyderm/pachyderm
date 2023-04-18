import {
  ILayoutRestorer,
  JupyterFrontEnd,
  JupyterFrontEndPlugin,
} from '@jupyterlab/application';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {IFileBrowserFactory} from '@jupyterlab/filebrowser';
import {INotebookTracker} from '@jupyterlab/notebook';
import {ISettingRegistry} from '@jupyterlab/settingregistry';

import {MountPlugin} from './mount';
import {IMountPlugin, MountSettings} from './types';

const PLUGIN_ID = 'jupyterlab-pachyderm:mount';

const mount: JupyterFrontEndPlugin<IMountPlugin> = {
  id: PLUGIN_ID,
  autoStart: true,
  requires: [
    IDocumentManager,
    IFileBrowserFactory,
    ILayoutRestorer,
    INotebookTracker,
    ISettingRegistry,
  ],
  activate: (
    app: JupyterFrontEnd,
    manager: IDocumentManager,
    factory: IFileBrowserFactory,
    restorer: ILayoutRestorer,
    tracker: INotebookTracker,
    settingRegistry: ISettingRegistry,
  ): IMountPlugin => {
    const settings: MountSettings = {defaultPipelineImage: ''};
    const loadSettings = (registry: ISettingRegistry.ISettings): void => {
      settings.defaultPipelineImage = registry.get('defaultPipelineImage')
        .composite as string;
    };
    void Promise.all([settingRegistry.load(PLUGIN_ID), app.restored]).then(
      ([registry]) => {
        loadSettings(registry);
        registry.changed.connect(loadSettings);
      },
    );
    return new MountPlugin(app, settings, manager, factory, restorer, tracker);
  },
};

export default mount;
