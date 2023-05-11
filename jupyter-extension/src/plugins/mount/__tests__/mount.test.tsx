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
import {ServiceManager} from '@jupyterlab/services';
import {StateDB} from '@jupyterlab/statedb';
import {CommandRegistry} from '@lumino/commands';
import {Widget} from '@lumino/widgets';
import {mockedRequestAPI} from 'utils/testUtils';
import {MountPlugin} from '../mount';
import * as requestAPI from '../../../handler';
import {waitFor} from '@testing-library/react';
import {Mount, MountSettings} from '../types';

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
  let settings: MountSettings;
  let docManager: IDocumentManager;
  let docRegistry: DocumentRegistry;
  let manager: ServiceManager;
  let factory: IFileBrowserFactory;
  let fileBrowser: FileBrowser;
  let restorer: ILayoutRestorer;
  let widgetTracker: ILabShell;
  const mockRequestAPI = requestAPI as jest.Mocked<typeof requestAPI>;
  beforeEach(() => {
    mockRequestAPI.requestAPI.mockImplementation(mockedRequestAPI(items));

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
        commit: null,
        repo: 'images',
        glob: '/*',
        mode: 'ro',
        state: 'mounted',
        status: '',
        mountpoint: '/pfs',
        how_many_commits_behind: 0,
        actual_mounted_commit: '1a2b3c',
        latest_commit: '1a2b3c',
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
    const plugin = new MountPlugin(
      app,
      settings,
      docManager,
      factory,
      restorer,
      widgetTracker,
    );

    await plugin.ready;

    jest.runAllTimers();
    await waitFor(() => {
      expect(plugin.unmountedRepos).toHaveLength(2);
      expect(plugin.unmountedRepos[0].repo).toBe('images');
      expect(plugin.unmountedRepos[1].repo).toBe('data');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(2);
    });
    jest.runAllTimers();
    await waitFor(() => {
      expect(plugin.unmountedRepos).toHaveLength(1);
      expect(plugin.unmountedRepos[0].repo).toBe('data');
      expect(plugin.mountedRepos).toHaveLength(1);
      expect(plugin.mountedRepos[0].name).toBe('images');
      expect(mockRequestAPI.requestAPI).toHaveBeenCalledTimes(3);
    });
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
    expect(plugin.layout.orientation).toBe('vertical');
    expect(plugin.layout.widgets).toHaveLength(9);
    expect(plugin.layout.widgets[0]).toBeInstanceOf(ReactWidget);
    expect(plugin.layout.widgets[1]).toBeInstanceOf(ReactWidget);
    expect(plugin.layout.widgets[2]).toBeInstanceOf(ReactWidget);
    expect(plugin.layout.widgets[3]).toBeInstanceOf(ReactWidget);
    expect(plugin.layout.widgets[4]).toBeInstanceOf(ReactWidget);
    expect(plugin.layout.widgets[5]).toBeInstanceOf(FileBrowser);
    expect(plugin.layout.widgets[6]).toBeInstanceOf(ReactWidget);
    expect(plugin.layout.widgets[7]).toBeInstanceOf(ReactWidget);
    expect(plugin.layout.widgets[8]).toBeInstanceOf(ReactWidget);
  });

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
  });
});
