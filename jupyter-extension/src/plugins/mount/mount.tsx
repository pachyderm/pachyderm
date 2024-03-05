import React from 'react';
import {
  ILabShell,
  ILayoutRestorer,
  JupyterFrontEnd,
} from '@jupyterlab/application';
import {ReactWidget, showErrorMessage, UseSignal} from '@jupyterlab/apputils';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {DocumentRegistry} from '@jupyterlab/docregistry';
import {FileBrowser, IFileBrowserFactory} from '@jupyterlab/filebrowser';
import {INotebookModel, NotebookPanel} from '@jupyterlab/notebook';
import {Contents} from '@jupyterlab/services';
import {settingsIcon} from '@jupyterlab/ui-components';
import {Signal} from '@lumino/signaling';
import {SplitPanel, Panel, TabPanel, Widget} from '@lumino/widgets';

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

  // Screens
  private _panel: TabPanel;
  private _loader: ReactWidget;
  private _fullPageError: ReactWidget;
  private _configScreen: ReactWidget;
  private _pipelineScreen: ReactWidget;
  private _exploreScreen: Panel;
  private _datumScreen: SplitPanel;

  private _pfsBrowser: FileBrowser;
  private _datumBrowser: FileBrowser;
  private _poller: PollMounts;
  private _widgetTracker: ILabShell;
  private _readyPromise: Promise<void> = Promise.resolve();

  private _keepMounted = false;
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
    this._poller.mountedSignal.connect(this.verifyBrowserPath);
    this._poller.mountedSignal.connect(this.refresh);
    this._poller.mountedSignal.connect(this.saveMountedReposList);
    this._poller.unmountedSignal.connect(this.refresh);

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
        this._panel.currentWidget = this._exploreScreen;
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
    );

    this._exploreScreen = new Panel();
    this._exploreScreen.addWidget(
      ReactWidget.create(
        <UseSignal signal={this._poller.mountedSignal}>
          {(_, mounted) => (
            <UseSignal signal={this._poller.unmountedSignal}>
              {(_, unmounted) => (
                <Explore
                  mounted={mounted || this._poller.mounted}
                  unmounted={unmounted || this._poller.unmounted}
                  updateData={this._poller.updateData}
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

    this._datumBrowser = createCustomFileBrowser(
      app,
      manager,
      factory,
      'view_datum',
      'test',
      'datum',
    );

    this._datumScreen = new SplitPanel({orientation: 'vertical'});
    this._datumScreen.addWidget(
      ReactWidget.create(
        <UseSignal signal={this._saveInputSpecSignal}>
          {(_, repoViewInputSpec) => (
            <Datum
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

  setCurrentView(widget: Widget) {
    this._panel.currentWidget = widget;
  }
}
