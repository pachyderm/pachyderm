import React from 'react';
import {
  ILabShell,
  ILayoutRestorer,
  JupyterFrontEnd,
} from '@jupyterlab/application';
import {ReactWidget, UseSignal} from '@jupyterlab/apputils';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {DocumentRegistry} from '@jupyterlab/docregistry';
import {FileBrowser, IFileBrowserFactory} from '@jupyterlab/filebrowser';
import {INotebookModel, NotebookPanel} from '@jupyterlab/notebook';
import {Contents} from '@jupyterlab/services';
import {settingsIcon} from '@jupyterlab/ui-components';
import {Signal} from '@lumino/signaling';
import {SplitPanel, TabPanel, Widget} from '@lumino/widgets';

import {mountLogoIcon} from '../../utils/icons';
import {PollMounts} from './pollMounts';
import createCustomFileBrowser from './customFileBrowser';
import {
  AuthConfig,
  IMountPlugin,
  Repos,
  MountedRepo,
  CurrentDatumResponse,
  PfsInput,
  CrossInputSpec,
  PpsMetadata,
  PpsContext,
  MountSettings,
  Repo,
  Branch,
} from './types';
import Config from './components/Config/Config';
import Datum from './components/Datum/Datum';
import Explore from './components/Explore/Explore';
import Pipeline from './components/Pipeline/Pipeline';
import LoadingDots from '../../utils/components/LoadingDots/LoadingDots';
import FullPageError from './components/FullPageError/FullPageError';

export const MOUNT_BROWSER_PREFIX = 'mount-browser-';
export const PFS_MOUNT_BROWSER_NAME = 'mount-browser-pfs:';
export const DATUM_MOUNT_BROWSER_NAME = 'mount-browser-datum:';

export const METADATA_KEY = 'pachyderm_pps';

export class MountPlugin implements IMountPlugin {
  private _app: JupyterFrontEnd<JupyterFrontEnd.IShell, 'desktop' | 'mobile'>;

  // Screens
  private _panel: TabPanel;
  private _loader: ReactWidget;
  private _fullPageError: ReactWidget;
  private _configScreen: ReactWidget;
  private _pipelineScreen: ReactWidget;
  private _exploreScreen: SplitPanel;
  private _datumScreen: SplitPanel;

  private _pfsBrowser: FileBrowser;
  private _datumBrowser: FileBrowser;
  private _poller: PollMounts;
  private _widgetTracker: ILabShell;
  private _readyPromise: Promise<void> = Promise.resolve();

  private _currentDatumInfo: CurrentDatumResponse | undefined;
  private _repoViewInputSpec: CrossInputSpec | PfsInput = {};
  private _saveInputSpecSignal = new Signal<this, CrossInputSpec | PfsInput>(
    this,
  );
  private _ppsContextSignal = new Signal<this, PpsContext>(this);

  constructor(
    app: JupyterFrontEnd,
    settings: MountSettings,
    manager: IDocumentManager,
    factory: IFileBrowserFactory,
    restorer: ILayoutRestorer,
    widgetTracker: ILabShell,
  ) {
    this._app = app;
    this._poller = new PollMounts('PollMounts');
    this._repoViewInputSpec = {};
    this._widgetTracker = widgetTracker;

    // Setup Poller signals.
    this._poller.reposSignal.connect(this.refresh);
    this._poller.mountedRepoSignal.connect(this.refresh);
    this._poller.mountedRepoSignal.connect(this.updateMountedRepoInputSpec);
    this._poller.mountedRepoSignal.connect(
      this.resetDirectoryOnMountedRepoChange,
    );

    // Call this initially to setup the input spec from localStorage if need be.
    this.updateMountedRepoInputSpec();

    // This is used to detect if the config goes bad (pachd address changes)
    this._poller.healthSignal.connect((_, healthCheck) => {
      const status = healthCheck ? healthCheck.status : this._poller.health;
      if (
        status === 'HEALTHY_INVALID_CLUSTER' ||
        status === 'HEALTHY_LOGGED_OUT'
      ) {
        this._panel.tabBar.hide();
        this.setCurrentView(this._configScreen);
      } else if (status === 'UNHEALTHY') {
        this._panel.tabBar.hide();
        this.setCurrentView(this._fullPageError);
      } else {
        this._panel.tabBar.show();
        this.setCurrentView(this._exploreScreen);
      }
    });

    this._readyPromise = this.setup();

    // Instantiate all of the Screens.
    this._configScreen = ReactWidget.create(
      <UseSignal signal={this._poller.healthSignal}>
        {(_, healthCheck) => (
          <UseSignal signal={this._poller.configSignal}>
            {(_, config) => (
              <Config
                updateConfig={this.updateConfig}
                healthCheck={healthCheck ? healthCheck : this._poller.health}
                authConfig={config ? config : this._poller.config}
                refresh={this._poller.refresh}
              />
            )}
          </UseSignal>
        )}
      </UseSignal>,
    );
    this._configScreen.addClass('pachyderm-mount-config-wrapper');
    this._configScreen.title.icon = settingsIcon;
    this._configScreen.title.className = 'pachyderm-config-tab';

    this._pfsBrowser = createCustomFileBrowser(
      app,
      manager,
      factory,
      'pfs',
      'explore',
      'pfs',
      () => {
        this.updateMountedRepo(null, null, null);
      },
    );

    this._exploreScreen = new SplitPanel({orientation: 'vertical'});
    this._exploreScreen.addWidget(
      ReactWidget.create(
        <UseSignal signal={this._poller.reposSignal}>
          {(_, repos) => (
            <UseSignal signal={this._poller.mountedRepoSignal}>
              {(_, mountedRepo) => (
                <Explore
                  repos={repos || this._poller.repos}
                  mountedRepo={mountedRepo || this._poller.mountedRepo}
                  updateMountedRepo={this.updateMountedRepo.bind(this)}
                  mountedRepoUri={this._poller.mountedRepoUri}
                />
              )}
            </UseSignal>
          )}
        </UseSignal>,
      ),
    );
    this._exploreScreen.addWidget(this._pfsBrowser);
    this._exploreScreen.title.label = 'Explore';
    this._exploreScreen.title.className = 'pachyderm-explore-tab';
    this._exploreScreen.widgets[0].addClass('pachyderm-explore-widget');
    this._datumBrowser = createCustomFileBrowser(
      app,
      manager,
      factory,
      'view_datum',
      'test',
      'datum',
      () => {
        this.updateMountedRepo(null, null, null);
      },
    );

    this._datumScreen = new SplitPanel({orientation: 'vertical'});
    this._datumScreen.addWidget(
      ReactWidget.create(
        <UseSignal signal={this._saveInputSpecSignal}>
          {(_, repoViewInputSpec) => (
            <Datum
              executeCommand={(command, options) => {
                this._app.commands.execute(command, options);
              }}
              open={this.openDatum}
              pollRefresh={this._poller.refresh}
              currentDatumInfo={this._currentDatumInfo}
              repoViewInputSpec={
                repoViewInputSpec ? repoViewInputSpec : this._repoViewInputSpec
              }
            />
          )}
        </UseSignal>,
      ),
    );
    this._datumScreen.addWidget(this._datumBrowser);
    this._datumScreen.addClass('pachyderm-mount-datum-wrapper');
    this._datumScreen.title.label = 'Test';
    this._datumScreen.title.className = 'pachyderm-test-tab';

    this._pipelineScreen = ReactWidget.create(
      <UseSignal signal={this._ppsContextSignal}>
        {(_, context) => (
          <Pipeline
            ppsContext={context}
            settings={settings}
            isCurrentWidgetNotebook={this.isCurrentWidgetNotebook}
            saveNotebookMetadata={this.saveNotebookMetadata}
            saveNotebookToDisk={this.saveNotebookToDisk}
          />
        )}
      </UseSignal>,
    );
    this._pipelineScreen.addClass('pachyderm-mount-pipeline-wrapper');
    this._pipelineScreen.title.label = 'Publish';
    this._pipelineScreen.title.className = 'pachyderm-publish-tab';

    this._loader = ReactWidget.create(<LoadingDots />);
    this._loader.addClass('pachyderm-mount-react-wrapper');
    this._loader.title.label = 'Loading';

    this._fullPageError = ReactWidget.create(
      <UseSignal signal={this._poller.healthSignal}>
        {(_, healthCheck) => (
          <FullPageError
            healthCheck={healthCheck ? healthCheck : this._poller.health}
          />
        )}
      </UseSignal>,
    );
    this._fullPageError.addClass('pachyderm-mount-react-wrapper');
    this._fullPageError.title.label = 'Error';

    this._widgetTracker.currentChanged.connect(this.handleWidgetChanged, this);

    // Construct the panel which organizes the screens.
    this._panel = new TabPanel();
    this._panel.title.icon = mountLogoIcon;
    this._panel.title.caption = 'Pachyderm Mount';
    this._panel.id = 'pachyderm-mount';

    this._panel.addWidget(this._exploreScreen);
    this._panel.addWidget(this._datumScreen);
    this._panel.addWidget(this._pipelineScreen);
    this._panel.addWidget(this._configScreen);

    // Add these widgets to the layout, but remove them as tabs.
    this._panel.addWidget(this._loader);
    this._panel.tabBar.removeTab(this._loader.title);
    this._panel.addWidget(this._fullPageError);
    this._panel.tabBar.removeTab(this._fullPageError.title);

    window.addEventListener('resize', () => {
      this._panel.update();
    });

    restorer.add(this._panel, 'jupyterlab-pachyderm');
    app.shell.add(this._panel, 'left', {rank: 100});
  }

  updateMountedRepo = async (
    repo: Repo | null,
    mountedBranch: Branch | null,
    commit: string | null,
  ): Promise<void> => {
    await this._poller.updateMountedRepo(repo, mountedBranch, commit);

    this._pfsBrowser.model.cd('/');
    this._datumBrowser.model.cd('/');

    return Promise.resolve();
  };

  /**
   * Checks if the widget currently in focus is a NotebookPanel.
   * @param widget is an optionally-provided Widget that will be type narrowed.
   */
  isCurrentWidgetNotebook = (
    widget?: Widget | null,
  ): widget is NotebookPanel => {
    widget = widget ?? this._widgetTracker.currentWidget;
    return widget instanceof NotebookPanel;
  };

  getActiveNotebook = (): NotebookPanel | null => {
    const currentWidget = this._widgetTracker.currentWidget;
    return this.isCurrentWidgetNotebook(currentWidget) ? currentWidget : null;
  };

  /**
   * This handles when the focus is switched to another widget.
   * The parameters are automatically passed from the signal when a switch occurs.
   */
  handleWidgetChanged = async (
    _widgetTracker: ILabShell,
    widget: ILabShell.IChangedArgs,
  ): Promise<void> => {
    if (this.isCurrentWidgetNotebook(widget.newValue)) {
      await this.handleNotebookChanged(widget.newValue);
    } else {
      this._ppsContextSignal.emit({metadata: null, notebookModel: null});
    }
    await Promise.resolve();
  };

  /**
   * This handles when a notebook is switched to another notebook.
   */
  handleNotebookChanged = async (notebook: NotebookPanel): Promise<void> => {
    // Set the current notebook and wait for the session to be ready
    await notebook.sessionContext.ready;
    notebook.context.fileChanged.connect(this.handleNotebookReload);
    const context: PpsContext = {
      metadata: this.getNotebookMetadata(notebook),
      notebookModel: notebook.context.contentsModel,
    };
    this._ppsContextSignal.emit(context);
    await Promise.resolve();
  };

  /**
   * This handles when a ContentModel of NotebookPanel is changed.
   * This occurs when a notebook file is saved and reloaded from disk.
   */
  handleNotebookReload = async (
    _docContext: DocumentRegistry.IContext<INotebookModel>,
    model: Omit<Contents.IModel, 'content'>,
  ): Promise<void> => {
    const context: PpsContext = {
      metadata: this.getNotebookMetadata(),
      notebookModel: model,
    };
    this._ppsContextSignal.emit(context);
    await Promise.resolve();
  };

  getNotebookMetadata = (notebook?: NotebookPanel | null): any | null => {
    notebook = notebook ?? this.getActiveNotebook();
    return notebook?.model?.getMetadata(METADATA_KEY);
  };

  saveNotebookMetadata = (metadata: PpsMetadata): void => {
    const currentNotebook = this.getActiveNotebook();

    if (currentNotebook !== null) {
      currentNotebook?.model?.setMetadata(METADATA_KEY, metadata);
      console.log('notebook metadata saved');
    } else {
      console.log('No active notebook');
    }
  };

  /**
   * saveNotebookToDisk saves the active notebook to disk and then returns
   *   the new modified time of the file.
   */
  saveNotebookToDisk = async (): Promise<string | null> => {
    const currentNotebook = this.getActiveNotebook();
    if (currentNotebook !== null) {
      await currentNotebook.context.ready;
      await currentNotebook.context.save();

      // Calling ready ensures the contentsModel is non-null.
      await currentNotebook.context.ready;
      const currentNotebookModel: Contents.IModel = currentNotebook.context
        .contentsModel as Contents.IModel;
      return currentNotebookModel.last_modified;
    } else {
      return null;
    }
  };

  openPFS = (path: string): void => {
    this._app.commands.execute('filebrowser:open-path', {
      path: PFS_MOUNT_BROWSER_NAME + path,
    });
  };

  openDatum = (path: string): void => {
    this._app.commands.execute('filebrowser:open-path', {
      path: DATUM_MOUNT_BROWSER_NAME + path,
    });
  };

  // _data is any because the data passed to this callback is not used. This allows it to be used
  // with any signal.
  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
  refresh = async (_: PollMounts, _data: any): Promise<void> => {
    await this._pfsBrowser.model.refresh();
    await this._datumBrowser.model.refresh();
  };

  updateMountedRepoInputSpec = (): void => {
    this._repoViewInputSpec = this._poller.getMountedRepoInputSpec();
    this._saveInputSpecSignal.emit(this._repoViewInputSpec);
  };

  resetDirectoryOnMountedRepoChange = (): void => {
    this._pfsBrowser.model.cd('');
  };

  updateConfig = (config: AuthConfig): void => {
    this._poller.config = config;
  };

  setup = async (): Promise<void> => {
    await this._poller.refresh();
    this._loader.setHidden(true);
  };

  get mountedRepo(): MountedRepo | null {
    this._poller.poll.tick;
    return this._poller.mountedRepo;
  }

  get repos(): Repos {
    this._poller.poll.tick;
    return this._poller.repos;
  }

  get layout(): TabPanel {
    return this._panel;
  }

  get ready(): Promise<void> {
    return this._readyPromise;
  }

  setCurrentView(widget: Widget): void {
    this._panel.currentWidget = widget;
  }
}
