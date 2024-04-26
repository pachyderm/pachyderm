// @ts-check
// eslint-disable-next-line @typescript-eslint/triple-slash-reference
///<reference path="./commands.d.ts" />

import '@testing-library/cypress/add-commands';
import 'cypress-wait-until';
import {jsonrepair} from 'jsonrepair';
// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add('login', (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add('drag', { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add('dismiss', { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite('visit', (originalFn, url, options) => { ... })

Cypress.Commands.add(
  'login',
  (route = '/', email = 'admin', password = 'password') => {
    cy.visit('/', {timeout: 12000});
    cy.findByRole('textbox', {timeout: 12000}).type(email);
    cy.findByLabelText('Password').type(password);

    cy.findByRole('button', {name: /login/i, timeout: 12000}).click();

    cy.window().then((win) => {
      cy.waitUntil(() => Boolean(win.localStorage.getItem('auth-token')), {
        errorMsg: 'Auth-token was not set in localstorage',
      });
    });

    if (route !== '/') return cy.visit(route);

    return;
  },
);

Cypress.Commands.add('viewAllLandingPageProjects', () => {
  cy.findByRole('tab', {name: /all projects/i}).click({
    force: true,
  });
});

Cypress.Commands.add('logout', () => {
  cy.clearCookies();
  return cy.clearLocalStorage();
});

/**
 * This can be used to naturally exec pachctl commands.
 *
 * It is used as such:
 * cy.multiLineExec(`
 *    echo "pizza" | pachctl auth use-auth-token
 *    pachctl create project root --description "Project made by pach:root. You can see nothing."
 *    pachctl create repo images --project root
 *    pachctl auth set cluster none user:kilgore@kilgore.trout
 *    `);
 */
Cypress.Commands.add('multiLineExec', (stringInput) => {
  const inputs = stringInput
    .split('\n')
    .map((el) => el.trimStart())
    .filter((el) => !!el);
  return inputs.forEach((command) => cy.exec(command));
});

Cypress.Commands.add('setupProject', (projectTemplate) => {
  if (projectTemplate === 'error-opencv') {
    return cy.exec('jq -r .pachReleaseCommit version.json').then((res) => {
      cy.exec('pachctl create repo images')
        // invalid image to trigger an error state
        .exec(`echo "gibberish" | pachctl put file images@master:badImage.png`)
        .exec(
          `pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/${res.stdout}/examples/opencv/edges.pipeline.json`,
        )
        .exec(
          `echo '${JSON.stringify({
            pipeline: {
              name: 'montage',
            },
            description:
              'A pipeline that combines images from the `images` and `edges` repositories into a montage.',
            input: {
              cross: [
                {
                  pfs: {
                    glob: '/',
                    repo: 'images',
                  },
                },
                {
                  pfs: {
                    glob: '/',
                    repo: 'edges',
                  },
                },
              ],
            },
            transform: {
              cmd: ['sh'],
              image: 'dpokidov/imagemagick:7.0.10-58',
              stdin: [
                'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
              ],
            },
            egress: {
              URL: 's3://test',
            },
          })}' | pachctl create pipeline`,
        );
    });
  } else if (projectTemplate === 'error-opencv-with-one-success') {
    return cy.exec('jq -r .pachReleaseCommit version.json').then((res) => {
      cy.exec('pachctl create repo images')
        .exec(
          'pachctl put file images@master:image1.png -f cypress/fixtures/liberty.png',
        )
        .exec(
          `pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/${res.stdout}/examples/opencv/edges.pipeline.json`,
        )
        // invalid image to trigger an error state
        .exec(`echo "gibberish" | pachctl put file images@master:badImage.png`)
        .exec(
          `echo '${JSON.stringify({
            pipeline: {
              name: 'montage',
            },
            description:
              'A pipeline that combines images from the `images` and `edges` repositories into a montage.',
            input: {
              cross: [
                {
                  pfs: {
                    glob: '/',
                    repo: 'images',
                  },
                },
                {
                  pfs: {
                    glob: '/',
                    repo: 'edges',
                  },
                },
              ],
            },
            transform: {
              cmd: ['sh'],
              image: 'dpokidov/imagemagick:7.0.10-58',
              stdin: [
                'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
              ],
            },
            egress: {
              URL: 's3://test',
            },
          })}' | pachctl create pipeline`,
        );
    });
  } else if (projectTemplate === 'file-browser') {
    return cy.exec('jq -r .pachReleaseCommit version.json').then(() => {
      cy.exec('pachctl create repo images')
        .exec(
          'pachctl put file images@master:image1.png -f cypress/fixtures/liberty.png',
        )
        .exec(
          'pachctl put file images@master:image1.png -f cypress/fixtures/AT-AT.png',
        )
        .exec(
          'pachctl put file images@test:image1.png -f cypress/fixtures/liberty.png',
        )
        .exec('pachctl delete file images@test:image1.png')
        .exec(
          'pachctl put file images@test:image1.png -f cypress/fixtures/liberty.png',
        )
        .exec(
          'pachctl put file images@test:image1.png -f cypress/fixtures/AT-AT.png',
        );
    });
  }

  return cy.exec('jq -r .pachReleaseCommit version.json').then((res) => {
    cy.exec('pachctl create repo images')
      .exec(
        'pachctl put file images@master:liberty.png -f cypress/fixtures/liberty.png',
      )
      .exec(
        `pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/${res.stdout}/examples/opencv/edges.pipeline.json`,
      );
  });
});

Cypress.Commands.add('deleteProjectReposAndPipelines', (project: string) => {
  cy.visit('/');
  return (
    cy
      .exec(`pachctl delete pipeline --all --force --project ${project}`, {
        failOnNonZeroExit: false,
      })
      .exec(`pachctl delete repo --all --force --project ${project}`, {
        failOnNonZeroExit: false,
      })
      .exec(`pachctl delete project ${project} -f`, {
        failOnNonZeroExit: false,
      })
  );
});

Cypress.Commands.add('deleteReposAndPipelines', () => {
  cy.visit('/');
  return (
    cy
      // These two commands are only needed for tests that reuse default
      .exec('pachctl delete pipeline --all --force')
      .exec('pachctl delete repo --all --force')
      .deleteProjects()
  );
});

Cypress.Commands.add('deleteProjects', (keepDefaultProject = true) => {
  cy.visit('/');
  return (
    cy
      // Remove all projects except for default
      .exec('pachctl list projects --raw')
      .then((res) => {
        // this gives us an invalid array of objects
        const repairedString = jsonrepair(`[${res.stdout}]`);
        let projects = JSON.parse(repairedString)
          .map((el) => el.project.name)
          .filter((project) => !!project);

        if (keepDefaultProject) {
          projects = projects.filter(
            (project) => project.toLowerCase() !== 'default',
          );
        }

        if (projects.length !== 0) cy.log('Deleting projects', projects);
        else
          cy.log(
            `No projects to delete${
              keepDefaultProject && ' (ignoring default)'
            }.`,
          );

        projects.forEach((project) => {
          // yes will keep entering y. This prompt can ask for a y/N twice.
          cy.exec(`yes | pachctl delete project ${project}`);
        });
      })
  );
});

Cypress.Commands.add('isInViewport', (element) => {
  element().then(($el) => {
    const bottom = Cypress.$(cy.state('window')).height() || 0;
    const rect = $el[0].getBoundingClientRect();

    expect(rect.top).not.to.be.greaterThan(bottom);
    expect(rect.bottom).not.to.be.greaterThan(bottom);
  });
});
