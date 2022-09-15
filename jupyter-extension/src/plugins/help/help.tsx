import React from 'react';
import {JupyterFrontEnd} from '@jupyterlab/application';
import {IMainMenu} from '@jupyterlab/mainmenu';
import {Dialog, showDialog} from '@jupyterlab/apputils';
import {PachydermLogo} from '../../utils/components/Svgs';

namespace CommandIDs {
  export const openDocs = 'jupyterlab-pachyderm:open-docs';
  export const contactSupport = 'jupyterlab-pachyderm:contact-support';
}

export const init = (app: JupyterFrontEnd, mainMenu?: IMainMenu): void => {
  if (mainMenu) {
    app.commands.addCommand(CommandIDs.openDocs, {
      label: 'Pachyderm Docs',
      execute: () => {
        window.open('https://docs.pachyderm.com/latest/getting_started/');
      },
    });

    app.commands.addCommand(CommandIDs.contactSupport, {
      label: 'Contact Pachyderm Support',
      execute: () => {
        const title = (
          <div className="pachyderm-help-title">
            <PachydermLogo className="pachyderm-help-title-logo" />{' '}
            <h2>Pachyderm Support</h2>
          </div>
        );
        const body = (
          <div className="pachyderm-help-body">
            <span>
              Chat with us on{' '}
              <a
                href="https://slack.pachyderm.io"
                className="pachyderm-help-link"
              >
                Slack
              </a>
            </span>
            <span>
              Email us at{' '}
              <a
                href="mailto:support@pachyderm.com"
                className="pachyderm-help-link"
              >
                support@pachyderm.com
              </a>
            </span>
          </div>
        );

        return showDialog({
          title,
          body,
          buttons: [
            Dialog.createButton({
              label: 'Dismiss',
              className: 'jp-About-button jp-mod-reject jp-mod-styled',
            }),
          ],
        });
      },
    });

    const helpMenu = mainMenu.helpMenu;
    helpMenu.addGroup(
      [
        {
          command: CommandIDs.openDocs,
        },
        {
          command: CommandIDs.contactSupport,
        },
      ],
      20,
    );
  }
};
