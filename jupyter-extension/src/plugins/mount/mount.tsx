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
  Repo,
  Mount,
  CurrentDatumResponse,
  PfsInput,
  CrossInputSpec,
  PpsMetadata,
  PpsContext,
  MountSettings,
} from './types';
import Config from './components/Config/Config';
import Datum from './components/Datum/Datum';
import Explore from './components/Explore/Explore';
import Pipeline from './components/Pipeline/Pipeline';
import LoadingDots from '../../utils/components/LoadingDots/LoadingDots';
import FullPageError from './components/FullPageError/FullPageError';
import {requestAPI} from '../../handler';

export const MOUNT_BROWSER_PREFIX = 'mount-browser-';
export const PFS_MOUNT_BROWSER_NAME = 'mount-browser-pfs:';
export const DATUM_MOUNT_BROWSER_NAME = 'mount-browser-datum:';

export const METADATA_KEY = 'pachyderm_pps';

export class MountPlugin implements IMountPlugin {
  private _app: JupyterFrontEnd<JupyterFrontEnd.IShell, 'desktop' | 'mobile'>;
  private _loader: ReactWidget;
  private _fullPageError: ReactWidget;
  private _config: ReactWidget;
  private _pipeline: ReactWidget;
  private _explore: ReactWidget;
  private _explorePanel: SplitPanel;

  private _datum: ReactWidget;
  private _datumPanel: SplitPanel;

  private _pfsBrowser: FileBrowser;
  private _datumBrowser: FileBrowser;
  private _poller: PollMounts;
  private _panel: TabPanel;
  private _widgetTracker: ILabShell;

  private _showConfig = false;
  private _showConfigSignal = new Signal<this, boolean>(this);
  private _readyPromise: Promise<void> = Promise.resolve();

  private _showDatum = false;
  private _keepMounted = false;
  private _currentDatumInfo: CurrentDatumResponse | undefined;
  private _showDatumSignal = new Signal<this, boolean>(this);
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

    // This is used to detect if the config goes bad (pachd address changes)
    this._poller.configSignal.connect((_, config) => {
      const status = config.cluster_status;
      if (status === 'INVALID' || status === 'VALID_LOGGED_OUT') {
        this._panel.tabBar.hide();
        this.setCurrentView(this._config);
      } else {
        this._panel.tabBar.show();
        this._panel.currentWidget = this._explorePanel;
      }
    });

    // This is used to detect if the user becomes unauthenticated of there are errors on the server
    this._poller.statusSignal.connect((_, status) => {
      if (status.code === 500) {
        this._panel.tabBar.hide();
        this.setCurrentView(this._fullPageError);
      }

      if (status.code === 401 && !this._showConfig) {
        this._panel.tabBar.hide();
        this.setCurrentView(this._config);
      }
    });

    this._readyPromise = this.setup();

    this._config = ReactWidget.create(
      <UseSignal signal={this._showConfigSignal}>
        {(_, showConfig) => (
          <UseSignal signal={this._poller.configSignal}>
            {(_, authConfig) => (
              <Config
                showConfig={showConfig ? showConfig : this._showConfig}
                updateConfig={this.updateConfig}
                authConfig={authConfig ? authConfig : this._poller.config}
                refresh={this._poller.refresh}
              />
            )}
          </UseSignal>
        )}
      </UseSignal>,
    );
    this._config.addClass('pachyderm-mount-config-wrapper');
    this._config.title.icon = settingsIcon;
    this._config.title.className = 'pachyderm-config-tab';

    this._pfsBrowser = createCustomFileBrowser(
      app,
      manager,
      factory,
      'pfs',
      'explore',
      'pfs',
    );

    this._explore = ReactWidget.create(
      <UseSignal signal={this._poller.mountedSignal}>
        {(_, mounted) => (
          <UseSignal signal={this._poller.unmountedSignal}>
            {(_, unmounted) => (
              <UseSignal signal={this._poller.projectSignal}>
                {(_, projects) => (
                  <Explore
                    mounted={mounted || this._poller.mounted}
                    unmounted={unmounted || this._poller.unmounted}
                    projects={projects || this._poller.projects}
                    openPFS={this.openPFS}
                    updateData={this._poller.updateData}
                  />
                )}
              </UseSignal>
            )}
          </UseSignal>
        )}
      </UseSignal>,
    );

    this._explorePanel = new SplitPanel();
    this._explorePanel.orientation = 'vertical';
    this._explorePanel.addWidget(this._explore);
    this._explorePanel.addWidget(this._pfsBrowser);
    this._explorePanel.title.label = 'Explore';

    this._datumBrowser = createCustomFileBrowser(
      app,
      manager,
      factory,
      'view_datum',
      'test',
      'datum',
    );

    this._datum = ReactWidget.create(
      <UseSignal signal={this._showDatumSignal}>
        {(_, showDatum) => (
          <UseSignal signal={this._saveInputSpecSignal}>
            {(_, repoViewInputSpec) => (
              <Datum
                showDatum={showDatum ? showDatum : this._showDatum}
                open={this.openDatum}
                pollRefresh={this._poller.refresh}
                currentDatumInfo={this._currentDatumInfo}
                repoViewInputSpec={
                  repoViewInputSpec
                    ? repoViewInputSpec
                    : this._repoViewInputSpec
                }
              />
            )}
          </UseSignal>
        )}
      </UseSignal>,
    );
    this._datum.addClass('pachyderm-mount-datum-wrapper');
    this._datumPanel = new SplitPanel();
    this._datumPanel.orientation = 'vertical';
    this._datumPanel.addWidget(this._datum);
    this._datumPanel.addWidget(this._datumBrowser);
    this._datumPanel.title.label = 'Test';

    this._pipeline = ReactWidget.create(
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
    this._pipeline.addClass('pachyderm-mount-pipeline-wrapper');
    this._pipeline.title.label = 'Publish';

    this._loader = ReactWidget.create(<LoadingDots />);
    this._loader.addClass('pachyderm-mount-react-wrapper');
    this._loader.title.label = 'Loading';

    this._fullPageError = ReactWidget.create(
      <UseSignal signal={this._poller.statusSignal}>
        {(_, status) => (
          <FullPageError status={status ? status : this._poller.status} />
        )}
      </UseSignal>,
    );
    this._fullPageError.addClass('pachyderm-mount-react-wrapper');
    this._fullPageError.title.label = 'Error';

    this._poller.mountedSignal.connect(this.verifyBrowserPath);
    this._poller.mountedSignal.connect(this.refresh);
    this._poller.unmountedSignal.connect(this.refresh);

    this._widgetTracker.currentChanged.connect(this.handleWidgetChanged, this);

    this._panel = new TabPanel();
    this._panel.title.icon = mountLogoIcon;
    this._panel.title.caption = 'Pachyderm Mount';
    this._panel.id = 'pachyderm-mount';

    this._panel.addWidget(this._explorePanel);
    this._panel.addWidget(this._datumPanel);
    this._panel.addWidget(this._pipeline);
    this._panel.addWidget(this._config);

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
    model: Contents.IModel,
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
    return notebook?.model?.metadata.get(METADATA_KEY);
  };

  saveNotebookMetadata = (metadata: PpsMetadata): void => {
    const currentNotebook = this.getActiveNotebook();

    if (currentNotebook !== null) {
      currentNotebook?.model?.metadata.set(METADATA_KEY, metadata);
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

  refresh = async (_: PollMounts, _data: Mount[] | Repo[]): Promise<void> => {
    await this._pfsBrowser.model.refresh();
    await this._datumBrowser.model.refresh();
  };

  // Change back to root directory if in a mount that no longer exists
  verifyBrowserPath = (_: PollMounts, mounted: Mount[]): void => {
    if (this._pfsBrowser.model.path === this._pfsBrowser.model.rootPath) {
      return;
    }

    if (!this.isValidBrowserPath(this._pfsBrowser.model.path, mounted)) {
      this.openPFS('');
    }
  };

  isValidBrowserPath = (path: string, mounted: Mount[]): boolean => {
    const currentMountDir = path.split(PFS_MOUNT_BROWSER_NAME)[1].split('/')[0];
    if (currentMountDir === 'out') {
      return true;
    }
    for (let i = 0; i < mounted.length; i++) {
      if (currentMountDir === mounted[i].name) {
        return true;
      }
    }
    return false;
  };

  saveMountedReposList = (): void => {
    const mounted = this._poller.mounted;
    const pfsInputs: PfsInput[] = [];

    for (let i = 0; i < mounted.length; i++) {
      const pfsInput: PfsInput = {
        pfs: {
          name: `${mounted[i].project}_${mounted[i].repo}`,
          ...(mounted[i].branch !== 'master' && {
            name: `${mounted[i].project}_${mounted[i].repo}_${mounted[i].branch}`,
          }),
          repo: mounted[i].repo,
          ...(mounted[i].project !== 'default' && {
            project: mounted[i].project,
          }),
          ...(mounted[i].branch !== 'master' && {branch: mounted[i].branch}),
          glob: '/',
        },
      };
      pfsInputs.push(pfsInput);
    }

    if (mounted.length === 0) {
      // No mounted repos to save
      this._repoViewInputSpec = {};
    } else if (mounted.length === 1) {
      // Single mounted repo to save
      this._repoViewInputSpec = pfsInputs[0];
    } else {
      // Multiple mounted repos to save; use cross
      this._repoViewInputSpec = {
        cross: pfsInputs,
      };
    }

    this._saveInputSpecSignal.emit(this._repoViewInputSpec);
  };

  setKeepMounted = (keep: boolean): void => {
    this._keepMounted = keep;
  };

  updateConfig = (config: AuthConfig): void => {
    this._poller.config = config;
  };

  setup = async (): Promise<void> => {
    await this._poller.refresh();

    if (this._poller.status.code === 500) {
      console.error('poller status 500');
      // this.setShowFullPageError(true);
    } else {
      if (this._poller.status.code === 200) {
        try {
          const res = await requestAPI<CurrentDatumResponse>('datums', 'GET');
          if (res['num_datums'] > 0) {
            this._keepMounted = true;
            this._currentDatumInfo = res;
            this.setCurrentView(this._datumPanel);
          }
        } catch (e) {
          console.log(e);
        }
      }
    }
    this._loader.setHidden(true);
  };

  get mountedRepos(): Mount[] {
    this._poller.poll.tick;
    return this._poller.mounted;
  }

  get unmountedRepos(): Repo[] {
    this._poller.poll.tick;
    return this._poller.unmounted;
  }

  get layout(): TabPanel {
    return this._panel;
  }

  get ready(): Promise<void> {
    return this._readyPromise;
  }

  private setCurrentView(widget: Widget) {
    this._panel.currentWidget = widget;
  }
}
