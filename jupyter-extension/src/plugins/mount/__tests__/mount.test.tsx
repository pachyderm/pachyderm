import {ReactWidget, WidgetTracker} from '@jupyterlab/apputils';
import {
  ILayoutRestorer,
  JupyterLab,
  LayoutRestorer,
} from '@jupyterlab/application';
import {IDocumentManager, DocumentManager} from '@jupyterlab/docmanager';
import {DocumentRegistry} from '@jupyterlab/docregistry';
import {
  FileBrowser,
  FilterFileBrowserModel,
  IFileBrowserFactory,
} from '@jupyterlab/filebrowser';
import {ServiceManager} from '@jupyterlab/services';
import {StateDB} from '@jupyterlab/statedb';
import {CommandRegistry} from '@lumino/commands';
import {Widget} from '@lumino/widgets';
import {mockedRequestAPI} from 'utils/testUtils';
import {MountPlugin} from '../mount';
import * as requestAPI from '../../../handler';
import {waitFor} from '@testing-library/react';

jest.mock('../../../handler');

const items = {
  unmounted: [
    {
      repo: 'images',
      branches: [],
    },
    {
      repo: 'data',
      branches: [],
    },
  ],
};

describe('mount plugin', () => {
  let app: JupyterLab;
  let docManager: IDocumentManager;
  let docRegistry: DocumentRegistry;
  let manager: ServiceManager;
  let factory: IFileBrowserFactory;
  let fileBrowser: FileBrowser;
  let restorer: ILayoutRestorer;
  const mockRequestAPI = requestAPI as jest.Mocked<typeof requestAPI>;
  beforeEach(() => {
    mockRequestAPI.requestAPI.mockImplementation(mockedRequestAPI(items));

    const opener = {
      open: (widget: Widget) => jest.fn(),
    };

    app = new JupyterLab();
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

    restorer = new LayoutRestorer({
      connector: new StateDB(),
      first: Promise.resolve<void>(void 0),
      registry: new CommandRegistry(),
    });
  });

  it.skip('should poll for mounts', async () => {
    mockRequestAPI.requestAPI
      .mockImplementationOnce(mockedRequestAPI(items)) // call to api from setup function, part of auth flow.
      .mockImplementationOnce(mockedRequestAPI(items))
      .mockImplementationOnce(
        mockedRequestAPI({
          mounted: [
            {
              images: {
                name: 'images',
                repo: 'images',
                branch: 'master',
                state: 'mounted',
              },
            },
          ],
          unmounted: [
            {
              data: {
                repo: 'data',
                branches: [],
              },
            },
          ],
        }),
      );
    const plugin = new MountPlugin(app, docManager, factory, restorer);

    await plugin.ready;

    jest.runAllTimers();
    await waitFor(() => {
      expect(plugin.unmountedRepos.length).toEqual(2);
      expect(plugin.unmountedRepos[0].repo).toEqual('images');
      expect(plugin.unmountedRepos[1].repo).toEqual('data');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(2);
    });
    jest.runAllTimers();
    await waitFor(() => {
      expect(plugin.unmountedRepos.length).toEqual(1);
      expect(plugin.unmountedRepos[0].repo).toEqual('data');
      expect(plugin.mountedRepos.length).toEqual(1);
      expect(plugin.mountedRepos[0].name).toEqual('images');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(3);
    });
  });

  it('should generate the correct layout', async () => {
    const plugin = new MountPlugin(app, docManager, factory, restorer);
    expect(plugin.layout.title.caption).toEqual('Pachyderm Mount');
    expect(plugin.layout.id).toEqual('pachyderm-mount');
    expect(plugin.layout.orientation).toEqual('vertical');
    expect(plugin.layout.widgets.length).toEqual(7);
    expect(plugin.layout.widgets[0]).toBeInstanceOf(ReactWidget);
    expect(plugin.layout.widgets[1]).toBeInstanceOf(ReactWidget);
    expect(plugin.layout.widgets[2]).toBeInstanceOf(ReactWidget);
    expect(plugin.layout.widgets[3]).toBeInstanceOf(FileBrowser);
    expect(plugin.layout.widgets[4]).toBeInstanceOf(ReactWidget);
    expect(plugin.layout.widgets[5]).toBeInstanceOf(ReactWidget);
    expect(plugin.layout.widgets[6]).toBeInstanceOf(ReactWidget);
  });
});
