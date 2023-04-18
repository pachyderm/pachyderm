import {
  ILabShell,
  ILayoutRestorer,
  JupyterFrontEnd,
  JupyterFrontEndPlugin,
} from '@jupyterlab/application';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {IFileBrowserFactory} from '@jupyterlab/filebrowser';
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
    ILabShell,
    ISettingRegistry,
  ],
  activate: (
    app: JupyterFrontEnd,
    manager: IDocumentManager,
    factory: IFileBrowserFactory,
    restorer: ILayoutRestorer,
    widgetTracker: ILabShell,
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
    return new MountPlugin(
      app,
      settings,
      manager,
      factory,
      restorer,
      widgetTracker,
    );
  },
};

export default mount;
