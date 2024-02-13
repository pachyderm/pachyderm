// Force animation frames to resolve immediately. Necessary for executing deferred code from Lumino Polls
window.requestAnimationFrame = (cb: any) => { return cb(); }; // prettier-ignore

import {ReactWidget, WidgetTracker} from '@jupyterlab/apputils';
import {
  ILabShell,
  ILayoutRestorer,
  JupyterLab,
  LabShell,
  LayoutRestorer,
} from '@jupyterlab/application';
import {IDocumentManager, DocumentManager} from '@jupyterlab/docmanager';
import {DocumentRegistry} from '@jupyterlab/docregistry';
import {
  FileBrowser,
  FilterFileBrowserModel,
  IFileBrowserFactory,
} from '@jupyterlab/filebrowser';
import {ServerConnection, ServiceManager} from '@jupyterlab/services';
import {StateDB} from '@jupyterlab/statedb';
import {CommandRegistry} from '@lumino/commands';
import {SplitPanel, Widget} from '@lumino/widgets';
import {MountPlugin} from '../mount';
import * as handler from '../../../handler';
import {MountSettings} from '../types';

jest.mock('../../../handler');

describe('mount plugin', () => {
  let app: JupyterLab;
  let settings: MountSettings;
  let docManager: IDocumentManager;
  let docRegistry: DocumentRegistry;
  let manager: ServiceManager;
  let factory: IFileBrowserFactory;
  let fileBrowser: FileBrowser;
  let restorer: ILayoutRestorer;
  let widgetTracker: ILabShell;
  const mockedRequestAPI = handler.requestAPI as jest.MockedFunction<
    typeof handler.requestAPI
  >;
  beforeEach(() => {
    mockedRequestAPI.mockClear();
    const opener = {
      open: (widget: Widget) => jest.fn(),
    };

    app = new JupyterLab();
    settings = {defaultPipelineImage: ''};
    docRegistry = new DocumentRegistry();
    manager = new ServiceManager();
    docManager = new DocumentManager({
      registry: docRegistry,
      manager: manager,
      opener: opener,
    });

    const model = new FilterFileBrowserModel({manager: docManager});

    fileBrowser = new FileBrowser({id: 'test', model});
    factory = {
      createFileBrowser: () => fileBrowser,
      defaultBrowser: fileBrowser,
      tracker: new WidgetTracker<FileBrowser>({namespace: 'test'}),
    };
    widgetTracker = new LabShell();
    restorer = new LayoutRestorer({
      connector: new StateDB(),
      first: Promise.resolve<void>(void 0),
      registry: new CommandRegistry(),
    });
  });
  /* TODO: tests must be updated for the new FUSE-less impl
  it('should accept /pfs/out as a valid FileBrowser path', async () => {
    const plugin = new MountPlugin(
      app,
      settings,
      docManager,
      factory,
      restorer,
      widgetTracker,
    );
    const mounts: Mount[] = [
      {
        name: 'default_images',
        project: 'default',
        branch: 'master',
        repo: 'images',
      },
    ];
    expect(
      plugin.isValidBrowserPath('mount-browser:default_images', mounts),
    ).toBe(true);
    expect(
      plugin.isValidBrowserPath('mount-browser:default_images/testdir', mounts),
    ).toBe(true);
    expect(
      plugin.isValidBrowserPath('mount-browser:default_edges', mounts),
    ).toBe(false);
    expect(
      plugin.isValidBrowserPath('mount-browser:default_edges/testdir', mounts),
    ).toBe(false);
    expect(plugin.isValidBrowserPath('mount-browser:out', mounts)).toBe(true);
    expect(plugin.isValidBrowserPath('mount-browser:out/testdir', mounts)).toBe(
      true,
    );
  });
*/

  it('should generate the correct layout', async () => {
    const plugin = new MountPlugin(
      app,
      settings,
      docManager,
      factory,
      restorer,
      widgetTracker,
    );
    expect(plugin.layout.title.caption).toBe('Pachyderm Mount');
    expect(plugin.layout.id).toBe('pachyderm-mount');

    expect(plugin.layout.widgets).toHaveLength(6);
    expect(plugin.layout.widgets[0]).toBeInstanceOf(SplitPanel); // Explore
    expect(plugin.layout.widgets[1]).toBeInstanceOf(SplitPanel); // Datum
    expect(plugin.layout.widgets[2]).toBeInstanceOf(ReactWidget); // Pipeline
    expect(plugin.layout.widgets[3]).toBeInstanceOf(ReactWidget); // Config
    expect(plugin.layout.widgets[4]).toBeInstanceOf(ReactWidget); // Loader
    expect(plugin.layout.widgets[5]).toBeInstanceOf(ReactWidget); // Error
  });

  /*
   This test is checking:
     If
   the plugin is instantiated and the backend does not have any config
   for connecting to a pachd instance (GET /config returns status UNKNOWN),
     Then
   - the plugin will immediately redirect to the CONFIG screen,
   - the poller will not make calls to GET /mounts or GET /projects (these requests would error).
   */
  it('should show config screen when not connected to a cluster', async () => {
    // Mock responses from GET /config
    mockedRequestAPI.mockImplementation(
      (endPoint?: string, method?: string) => {
        if (endPoint === 'config' && (method === 'GET' || method === null)) {
          return Promise.resolve({
            cluster_status: 'UNKNOWN',
            pachd_address: '',
          });
        }
        throw ServerConnection.NetworkError;
      },
    );

    const plugin = new MountPlugin(
      app,
      settings,
      docManager,
      factory,
      restorer,
      widgetTracker,
    );
    await plugin.ready;

    expect(plugin.layout.currentWidget?.title.className).toBe(
      'pachyderm-config-tab', // Easiest way to assert the widget is the config panel.
    );

    expect(mockedRequestAPI).toHaveBeenCalledTimes(1);
    expect(mockedRequestAPI).not.toHaveBeenCalledWith('mounts', 'GET');
    expect(mockedRequestAPI).not.toHaveBeenCalledWith('projects', 'GET');
  });

  /* TODO: tests must be updated for the new FUSE-less impl
  it('return from pipeline view to the correct layout', async () => {
    const plugin = new MountPlugin(
      app,
      settings,
      docManager,
      factory,
      restorer,
      widgetTracker,
    );
    const pipelineSplash = plugin.layout.widgets[3];
    const fileBrowser = plugin.layout.widgets[5];
    expect(fileBrowser).toBeInstanceOf(FileBrowser);

    plugin.setShowConfig(false);
    expect(pipelineSplash.isHidden).toBe(true);
    expect(fileBrowser.isHidden).toBe(false);

    plugin.setShowPipeline(true);
    expect(pipelineSplash.isHidden).toBe(false);
    expect(fileBrowser.isHidden).toBe(true);

    plugin.setShowPipeline(false);
    expect(pipelineSplash.isHidden).toBe(true);
    expect(fileBrowser.isHidden).toBe(false);
  });*/
});
