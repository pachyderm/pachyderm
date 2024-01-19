import {JupyterFrontEnd} from '@jupyterlab/application';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {Menu} from '@lumino/widgets';
import {
  IFileBrowserFactory,
  BreadCrumbs,
  DirListing,
  FileBrowser,
} from '@jupyterlab/filebrowser';
import {Clipboard, showErrorMessage} from '@jupyterlab/apputils';
import {CommandRegistry} from '@lumino/commands';
import {each} from '@lumino/algorithm';

import {MountDrive} from './mountDrive';
import {MOUNT_BROWSER_PREFIX} from './mount';
import {Paging} from './paging';
import {requestAPI} from '../../handler';

const createCustomFileBrowser = (
  app: JupyterFrontEnd,
  manager: IDocumentManager,
  factory: IFileBrowserFactory,
  path: string,
  downloadPath: string,
  nameSuffix: string,
): FileBrowser => {
  const drive = new MountDrive(app.docRegistry, path, nameSuffix);
  manager.services.contents.addDrive(drive);

  const browser = factory.createFileBrowser(
    'jupyterlab-pachyderm-browser-' + nameSuffix,
    {
      driveName: drive.name,
      state: null,
      refreshInterval: 10000,
    },
  );

  const filenameSearcher = browser.node
    .getElementsByClassName('jp-FileBrowser-filterBox')
    .item(0);
  if (filenameSearcher) {
    browser.node.removeChild(filenameSearcher);
  }

  const toolbar = browser.node
    .getElementsByClassName('jp-FileBrowser-toolbar')
    .item(0);
  if (toolbar) {
    browser.node.removeChild(toolbar);
  }

  try {
    const widgets = browser.layout.widgets || [];
    const breadCrumbs = widgets.find(
      (element) => element instanceof BreadCrumbs,
    );
    if (breadCrumbs) {
      breadCrumbs.node
        .querySelector('svg[data-icon="ui-components:folder"]')
        ?.replaceWith('/ pfs');
      const homeElement = breadCrumbs.node
        .getElementsByClassName('jp-BreadCrumbs-home')
        .item(0);
      if (homeElement) {
        homeElement.className = 'jp-BreadCrumbs-item';
      }
    }
    const dirListing = widgets.find((element) => element instanceof DirListing);
    if (dirListing) {
      const commands = new CommandRegistry();
      commands.addCommand('file-open', {
        label: 'Open',
        icon: 'fa fa-folder',
        mnemonic: 0,
        execute: () => {
          each(browser.selectedItems(), (item) => {
            manager.openOrReveal(item.path);
          });
        },
      });

      commands.addCommand('copy-path', {
        label: 'Copy Path',
        icon: 'fa fa-file',
        mnemonic: 0,
        execute: () => {
          each(browser.selectedItems(), (item) => {
            Clipboard.copyToSystem(
              item.path.replace(MOUNT_BROWSER_PREFIX + nameSuffix, '/pfs/'),
            );
          });
        },
      });

      commands.addCommand('file-download', {
        label: 'Download',
        icon: 'fa fa-download',
        mnemonic: 0,
        execute: () => {
          each(browser.selectedItems(), (item) => {
            // Unfortunately, copying between drives is not implemented:
            // https://github.com/jupyterlab/jupyterlab/blob/main/packages/services/src/contents/index.ts#L916
            // so we need to perform this logic within our implementation of the extension
            // Once this is implemented, we should be able to just write something along the lines of
            // manager.copy(item.path, "/home/jovyan/extension-wd")
            const itemPath = item.path.replace(
              MOUNT_BROWSER_PREFIX + nameSuffix + ':',
              '',
            );
            requestAPI(
              'download/' + downloadPath + '/' + itemPath,
              'PUT',
            ).catch((e) => {
              showErrorMessage('Download Error', e.response.statusText);
            });
          });
        },
      });

      const menu = new Menu({commands});
      menu.addItem({command: 'file-open'});
      menu.addItem({command: 'copy-path'});
      menu.addItem({command: 'file-download'});

      const browserContent = dirListing.node.getElementsByClassName(
        'jp-DirListing-content',
      )[0];

      // Connect the MountDrive loading signal to mark the browser content as loading.
      drive.loading.connect(async (_, loading) => {
        if (loading) {
          browserContent.setAttribute('loading', 'true');
        } else {
          browserContent.setAttribute('loading', 'false');
        }
      });

      browserContent.addEventListener('contextmenu', (event: any) => {
        event.stopPropagation();
        event.preventDefault();
        const x = event.clientX;
        const y = event.clientY;
        menu.open(x, y);
      });
    }
  } catch (e) {
    console.log('Failed to edit default browser.');
  }

  const paging = new Paging({
    browser_model: browser.model,
    page_model: drive.model,
  });
  browser.model.pathChanged.connect(async (_, __) => {
    // If the FileBrowser path has changed, then the MountDrive has reset the pagination model.
    // Therefore, the pagination widget needs to be re-rendered.
    paging.update();
  });
  browser.layout.addWidget(paging);

  return browser;
};

export default createCustomFileBrowser;
