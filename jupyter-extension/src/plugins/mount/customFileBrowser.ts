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
import {LabIcon} from '@jupyterlab/ui-components';
import {CommandRegistry} from '@lumino/commands';

import {MountDrive} from './mountDrive';
import {MOUNT_BROWSER_PREFIX} from './mount';
import {requestAPI} from '../../handler';
import {IPachydermModel} from './types';

const createCustomFileBrowser = (
  app: JupyterFrontEnd,
  manager: IDocumentManager,
  factory: IFileBrowserFactory,
  path: string,
  downloadPath: string,
  nameSuffix: string,
  unmountRepo: () => void,
): FileBrowser => {
  const id = `jupyterlab-pachyderm-browser-${nameSuffix}`;
  const drive = new MountDrive(
    app.docRegistry,
    path,
    nameSuffix,
    id,
    async () => {
      await browser.model.cd();
    },
    unmountRepo,
  );
  manager.services.contents.addDrive(drive);

  const browser = factory.createFileBrowser(id, {
    driveName: drive.name,
    refreshInterval: 10000,

    // Restoring the state and path after a refresh causes issues with infinite scrolling since it attempts to
    // select and scroll to the file opened on a delay. The file attempting to be selected may not be visible and
    // if it is then it interferes with the user scrolling immediately.
    state: null,

    // Setting this to false fixes an issue where the file browser was making requests
    // to our backend before the backend was capable of handling them, i.e. before the
    // user is connected to their pachd instance. When set to false, the poller that
    // refreshes the file browser contents every `refreshInterval` ms is initiated on
    // the first `cd` call that the file browser handles. This is compatible with how
    // the plugin currently utilizes the file browser.
    auto: false,
    restore: false,
  });

  const toolbar = browser.node
    .getElementsByClassName('jp-FileBrowser-toolbar')
    .item(0);
  if (toolbar) {
    browser.node.removeChild(toolbar);
  }

  try {
    const widgets = browser.widgets[0].layout
      ? Array.from(browser.widgets[0].layout)
      : [];
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
        icon: new LabIcon({
          name: 'folder-icon',
          svgstr:
            '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><!--!Font Awesome Free 6.5.2 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2024 Fonticons, Inc.--><path d="M464 128H272l-64-64H48C21.5 64 0 85.5 0 112v288c0 26.5 21.5 48 48 48h416c26.5 0 48-21.5 48-48V176c0-26.5-21.5-48-48-48z"/></svg>',
        }),
        mnemonic: 0,
        execute: () => {
          for (const item of browser.selectedItems()) {
            manager.openOrReveal(item.path);
          }
        },
      });

      commands.addCommand('copy-path', {
        label: 'Copy Path',
        icon: new LabIcon({
          name: 'file-icon',
          svgstr:
            '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 384 512"><!--!Font Awesome Free 6.5.2 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2024 Fonticons, Inc.--><path d="M0 64C0 28.7 28.7 0 64 0H224V128c0 17.7 14.3 32 32 32H384V448c0 35.3-28.7 64-64 64H64c-35.3 0-64-28.7-64-64V64zm384 64H256V0L384 128z"/></svg>',
        }),
        mnemonic: 0,
        execute: () => {
          for (const item of browser.selectedItems()) {
            Clipboard.copyToSystem(
              item.path.replace(MOUNT_BROWSER_PREFIX + nameSuffix, '/pfs/'),
            );
          }
        },
      });

      commands.addCommand('file-download', {
        label: 'Download',
        icon: new LabIcon({
          name: 'file-download',
          svgstr:
            '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><!--!Font Awesome Free 6.5.2 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2024 Fonticons, Inc.--><path d="M288 32c0-17.7-14.3-32-32-32s-32 14.3-32 32V274.7l-73.4-73.4c-12.5-12.5-32.8-12.5-45.3 0s-12.5 32.8 0 45.3l128 128c12.5 12.5 32.8 12.5 45.3 0l128-128c12.5-12.5 12.5-32.8 0-45.3s-32.8-12.5-45.3 0L288 274.7V32zM64 352c-35.3 0-64 28.7-64 64v32c0 35.3 28.7 64 64 64H448c35.3 0 64-28.7 64-64V416c0-35.3-28.7-64-64-64H346.5l-45.3 45.3c-25 25-65.5 25-90.5 0L165.5 352H64zm368 56a24 24 0 1 1 0 48 24 24 0 1 1 0-48z"/></svg>',
        }),
        mnemonic: 0,
        execute: () => {
          for (const item of browser.selectedItems()) {
            // Unfortunately, copying between drives is not implemented:
            // https://github.com/jupyterlab/jupyterlab/blob/main/packages/services/src/contents/index.ts#L916
            // so we need to perform this logic within our implementation of the extension
            // Once this is implemented, we should be able to just write something along the lines of
            // manager.copy(item.path, "/home/jovyan/extension-wd")
            const itemPath = item.path.replace(
              MOUNT_BROWSER_PREFIX + nameSuffix + ':',
              '',
            );
            requestAPI(`download/${downloadPath}/${itemPath}`, 'PUT').catch(
              (e) => {
                showErrorMessage('Download Error', e.response.statusText);
              },
            );
          }
        },
      });

      // We need to register this as an app command, but because this function is called multiple times we only want to register it once.
      // This command must be registered as an app command to work with notification commandIds
      if (!app.commands.hasCommand('open-pachyderm-sdk')) {
        app.commands.addCommand('open-pachyderm-sdk', {
          execute: () => {
            window
              ?.open('https://docs.pachyderm.com/latest/sdk/', '_blank')
              ?.focus();
          },
        });
      }

      commands.addCommand('open-determined', {
        label: 'Copy Pachyderm File URI',
        icon: new LabIcon({
          name: 'file-link',
          svgstr:
            '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 512"><!--!Font Awesome Free 6.5.2 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2024 Fonticons, Inc.--><path d="M579.8 267.7c56.5-56.5 56.5-148 0-204.5c-50-50-128.8-56.5-186.3-15.4l-1.6 1.1c-14.4 10.3-17.7 30.3-7.4 44.6s30.3 17.7 44.6 7.4l1.6-1.1c32.1-22.9 76-19.3 103.8 8.6c31.5 31.5 31.5 82.5 0 114L422.3 334.8c-31.5 31.5-82.5 31.5-114 0c-27.9-27.9-31.5-71.8-8.6-103.8l1.1-1.6c10.3-14.4 6.9-34.4-7.4-44.6s-34.4-6.9-44.6 7.4l-1.1 1.6C206.5 251.2 213 330 263 380c56.5 56.5 148 56.5 204.5 0L579.8 267.7zM60.2 244.3c-56.5 56.5-56.5 148 0 204.5c50 50 128.8 56.5 186.3 15.4l1.6-1.1c14.4-10.3 17.7-30.3 7.4-44.6s-30.3-17.7-44.6-7.4l-1.6 1.1c-32.1 22.9-76 19.3-103.8-8.6C74 372 74 321 105.5 289.5L217.7 177.2c31.5-31.5 82.5-31.5 114 0c27.9 27.9 31.5 71.8 8.6 103.9l-1.1 1.6c-10.3 14.4-6.9 34.4 7.4 44.6s34.4 6.9 44.6-7.4l1.1-1.6C433.5 260.8 427 182 377 132c-56.5-56.5-148-56.5-204.5 0L60.2 244.3z"/></svg>',
        }),
        mnemonic: 0,
        execute: () => {
          if (navigator.clipboard && window.isSecureContext) {
            let fileUris = '';
            for (const item of browser.selectedItems()) {
              fileUris += `${(item as IPachydermModel).file_uri}\n`;
            }
            navigator.clipboard.writeText(fileUris);
            app.commands.execute('apputils:notify', {
              message: 'Pachyderm File URI copied to clipboard.',
              type: 'success',
              options: {
                autoClose: 10000, // 10 seconds
                actions: [
                  {
                    label: 'Open Pachyderm SDK Docs',
                    commandId: 'open-pachyderm-sdk',
                    displayType: 'link',
                  },
                ],
              },
            });
          } else {
            for (let item of browser.selectedItems()) {
              item = item as IPachydermModel;
              app.commands.execute('apputils:notify', {
                message: (item as IPachydermModel).file_uri,
                type: 'success',
                options: {
                  autoClose: false, // disable autoclose since the user needs to copy the url manually.
                  actions: [
                    {
                      label: 'Open Pachyderm SDK Docs',
                      commandId: 'open-pachyderm-sdk',
                      displayType: 'link',
                    },
                  ],
                },
              });
            }
            // Notifications have around a 400 character restriction. This should likely workaround the problem of too many urls overloading that limit
            app.commands.execute('apputils:notify', {
              message:
                'Pachyderm File URI could not be copied to clipboard due to browser clipboard restrictions.',
              type: 'warning',
              options: {
                autoClose: false, // disable autoclose since the user needs to copy the url manually.
                actions: [
                  {
                    label: 'Open Pachyderm SDK Docs',
                    commandId: 'open-pachyderm-sdk',
                    displayType: 'link',
                  },
                ],
              },
            });
          }
        },
      });

      const menu = new Menu({commands});
      menu.addItem({command: 'file-open'});
      menu.addItem({command: 'copy-path'});
      menu.addItem({command: 'file-download'});
      menu.addItem({command: 'open-determined'});

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
    // This should not log!
    console.log('Failed to edit default browser.');
  }

  return browser;
};

export default createCustomFileBrowser;
