import '@testing-library/cypress/add-commands';
import 'cypress-wait-until';

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

Cypress.Commands.add('login', (email = 'admin', password = 'password') => {
  cy.visit('/', {timeout: 12000});
  cy.findByRole('textbox', {timeout: 12000}).type(email);
  cy.findByLabelText('Password').type(password);

  return cy.findByRole('button', {name: /login/i, timeout: 12000}).click();
});

Cypress.Commands.add('logout', () => {
  cy.clearCookies();
  return cy.clearLocalStorage();
});

Cypress.Commands.add('setupProject', (projectTemplate) => {
  if (projectTemplate === 'error-opencv') {
    return cy.exec('jq -r .pachReleaseCommit version.json').then((res) => {
      cy.exec('pachctl create repo images')
        // invalid image to trigger an error state
        .exec(`echo "gibberish" | pachctl put file images@master:badImage.png`)
        .exec(
          `pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/${res.stdout}/examples/opencv/edges.json`,
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
  } else if (projectTemplate === 'file-browser') {
    return cy.exec('jq -r .pachReleaseCommit version.json').then((res) => {
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
        `pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/${res.stdout}/examples/opencv/edges.json`,
      );
  });
});

Cypress.Commands.add('deleteReposAndPipelines', () => {
  return (
    cy
      .exec('pachctl delete pipeline --all --force')
      .exec('pachctl delete repo --all --force')
      // Remove all projects except for default
      .exec('pachctl list projects')
      .then((res) => {
        const regex = /(\w*)\s*\[/g;
        const projects = [...res.stdout.matchAll(regex)]
          .map((el) => el[1])
          .filter((el) => el.toLowerCase() !== 'default');
        if (projects.length !== 0) cy.log('Deleting projects', projects);
        else cy.log('No projects to delete (ignoring default).');
        projects.forEach((project) => {
          cy.exec(`echo y | pachctl delete project ${project}`);
        });
      })
  );
});

Cypress.Commands.add('isInViewport', (element) => {
  element().then(($el) => {
    const bottom = Cypress.$(cy.state('window')).height();
    const rect = $el[0].getBoundingClientRect();

    expect(rect.top).not.to.be.greaterThan(bottom);
    expect(rect.bottom).not.to.be.greaterThan(bottom);
  });
});
