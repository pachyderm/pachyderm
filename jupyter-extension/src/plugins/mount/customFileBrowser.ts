import {JupyterFrontEnd} from '@jupyterlab/application';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {Menu} from '@lumino/widgets';
import {
  IFileBrowserFactory,
  BreadCrumbs,
  DirListing,
  FileBrowser,
} from '@jupyterlab/filebrowser';
import {Clipboard} from '@jupyterlab/apputils';
import {CommandRegistry} from '@lumino/commands';
import {each} from '@lumino/algorithm';

import {MountDrive} from './mountDrive';
import {MOUNT_BROWSER_NAME} from './mount';

const createCustomFileBrowser = (
  app: JupyterFrontEnd,
  manager: IDocumentManager,
  factory: IFileBrowserFactory,
): FileBrowser => {
  const drive = new MountDrive(app.docRegistry);
  manager.services.contents.addDrive(drive);

  const browser = factory.createFileBrowser('jupyterlab-pachyderm-browser', {
    driveName: drive.name,
    state: null,
    refreshInterval: 10000,
  });

  try {
    const newLauncher = browser.toolbar.node.childNodes[0];
    const newFolder = browser.toolbar.node.childNodes[1];
    browser.toolbar.node.removeChild(newLauncher);
    browser.toolbar.node.removeChild(newFolder);

    const widgets = browser.layout.widgets || [];

    const breadCrumbs = widgets.find(
      (element) => element instanceof BreadCrumbs,
    );
    if (breadCrumbs) {
      breadCrumbs.node
        .querySelector('svg[data-icon="ui-components:folder"]')
        ?.replaceWith('/ pfs');
      const homeElement = breadCrumbs.node.querySelector(
        'span[title="~/extension-wd"]',
      );
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
              item.path.replace(MOUNT_BROWSER_NAME, '/pfs/'),
            );
          });
        },
      });

      const menu = new Menu({commands});
      menu.addItem({command: 'file-open'});
      menu.addItem({command: 'copy-path'});

      const browserContent = dirListing.node.getElementsByClassName(
        'jp-DirListing-content',
      )[0];

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

  return browser;
};

export default createCustomFileBrowser;
