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
import {
  IDocumentManager,
  DocumentManager,
  IDocumentWidgetOpener,
} from '@jupyterlab/docmanager';
import {DocumentRegistry, IDocumentWidget} from '@jupyterlab/docregistry';
import {
  FileBrowser,
  FilterFileBrowserModel,
  IFileBrowserFactory,
} from '@jupyterlab/filebrowser';
import {ServerConnection, ServiceManager} from '@jupyterlab/services';
import {StateDB} from '@jupyterlab/statedb';
import {CommandRegistry} from '@lumino/commands';
import {ISignal, Signal} from '@lumino/signaling';
import {SplitPanel} from '@lumino/widgets';
import {MountPlugin} from '../mount';
import * as handler from '../../../handler';
import {MountSettings} from '../types';

jest.mock('../../../handler');

class MockDocumentWidgetOpener implements IDocumentWidgetOpener {
  private _opened = new Signal<IDocumentWidgetOpener, IDocumentWidget>(this);

  open(widget: IDocumentWidget, options?: DocumentRegistry.IOpenOptions): void {
    console.log(
      'Mock open method called with widget:',
      widget,
      'and options:',
      options,
    );
    this._opened.emit(widget);
  }

  get opened(): ISignal<IDocumentWidgetOpener, IDocumentWidget> {
    return this._opened;
  }
}

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
    const opener = new MockDocumentWidgetOpener();

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
      tracker: new WidgetTracker<FileBrowser>({namespace: 'test'}),
    };
    widgetTracker = new LabShell();
    restorer = new LayoutRestorer({
      connector: new StateDB(),
      first: Promise.resolve<void>(void 0),
      registry: new CommandRegistry(),
    });
    jest.useRealTimers();
  });

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
   for connecting to a pachd instance (GET /health returns status HEALTHCHECK_INVALID_CLUSTER),
     Then
   - the plugin will immediately redirect to the CONFIG screen,
   - the poller will not make calls to GET /mounts or GET /projects (these requests would error).
   */
  it('should show config screen when not connected to a cluster', async () => {
    // Mock responses from GET /health
    mockedRequestAPI.mockImplementation(
      (endPoint?: string, method?: string) => {
        if (endPoint === 'health' && (method === 'GET' || method === null)) {
          return Promise.resolve({
            status: 'HEALTHY_INVALID_CLUSTER',
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

    expect(mockedRequestAPI).not.toHaveBeenCalledWith('mounts', 'GET');
    expect(mockedRequestAPI).not.toHaveBeenCalledWith('projects', 'GET');
  });

  it('return from pipeline view to the correct layout', async () => {
    // Mock responses from GET /health
    mockedRequestAPI.mockImplementation(
      (endPoint?: string, method?: string) => {
        if (endPoint === 'health' && (method === 'GET' || method === null)) {
          return Promise.resolve({
            status: 'HEALTHY_NO_AUTH',
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

    expect(
      plugin.layout.currentWidget?.title.className === 'pachyderm-explore-tab',
    );
    plugin.setCurrentView(plugin.layout.widgets[1]);
    expect(
      plugin.layout.currentWidget?.title.className === 'pachyderm-test-tab',
    );
    plugin.setCurrentView(plugin.layout.widgets[2]);
    expect(
      plugin.layout.currentWidget?.title.className === 'pachyderm-publish-tab',
    );
    plugin.setCurrentView(plugin.layout.widgets[0]);
    expect(
      plugin.layout.currentWidget?.title.className === 'pachyderm-explore-tab',
    );
  });
});
