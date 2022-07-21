import React from 'react';
import { Dialog, showDialog } from '@jupyterlab/apputils';
import { PachydermLogoFooterSVG } from '@pachyderm/components';
var CommandIDs;
(function (CommandIDs) {
    CommandIDs.openDocs = 'jupyterlab-pachyderm:open-docs';
    CommandIDs.contactSupport = 'jupyterlab-pachyderm:contact-support';
})(CommandIDs || (CommandIDs = {}));
export const init = (app, mainMenu) => {
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
                const title = (React.createElement("div", { className: "pachyderm-help-title" },
                    React.createElement(PachydermLogoFooterSVG, { className: "pachyderm-help-title-logo" }),
                    ' ',
                    React.createElement("h2", null, "Pachyderm Support")));
                const body = (React.createElement("div", { className: "pachyderm-help-body" },
                    React.createElement("span", null,
                        "Chat with us on",
                        ' ',
                        React.createElement("a", { href: "https://slack.pachyderm.io", className: "pachyderm-help-link" }, "Slack")),
                    React.createElement("span", null,
                        "Email us at",
                        ' ',
                        React.createElement("a", { href: "mailto:support@pachyderm.com", className: "pachyderm-help-link" }, "support@pachyderm.com"))));
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
        helpMenu.addGroup([
            {
                command: CommandIDs.openDocs,
            },
            {
                command: CommandIDs.contactSupport,
            },
        ], 20);
    }
};
