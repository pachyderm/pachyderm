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

Cypress.Commands.add('login', (
  email=Cypress.env('AUTH_EMAIL'),
  password=Cypress.env('AUTH_PASSWORD')
) => {
  cy.visit('/');
  cy.findByLabelText("Email", { timeout: 12000 }).type(email);
  cy.findByLabelText('Password').type(password);

  return cy.findByLabelText('Log In').click()
});

Cypress.Commands.add('logout', () => {
  cy.clearCookies();
  cy.clearLocalStorage();

  return cy.request('https://hub-e2e-testing.us.auth0.com/logout').visit('/');
});

Cypress.Commands.add('authenticatePachctl', () => {
  if (!window.localStorage.getItem('id-token')) {
    return cy.login()
    .waitUntil(() => window.localStorage.getItem('id-token')).then((idToken) => {
      cy.exec(`echo ${idToken} | pachctl auth login --id-token`)
    });
  } else {
    return cy.exec(`echo ${window.localStorage.getItem('id-token')} | pachctl auth login --id-token`)
  }

})

Cypress.Commands.add("setupProject", (projectTemplate) => {
  if (projectTemplate === "error-opencv") {
    return cy
      .exec("jq -r .pachyderm version.json")
      .then((res) => {
        cy.exec("pachctl create repo images")
          // invalid image to trigger an error state
          .exec(
            `echo "gibberish" | pachctl put file images@master:badImage.png`
          )
          .exec(
            `pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/v${res.stdout}/examples/opencv/edges.json`
          )
          .exec(
            `echo '${JSON.stringify({
              pipeline: {
                name: "montage",
              },
              description:
                "A pipeline that combines images from the `images` and `edges` repositories into a montage.",
              input: {
                cross: [
                  {
                    pfs: {
                      glob: "/",
                      repo: "images",
                    },
                  },
                  {
                    pfs: {
                      glob: "/",
                      repo: "edges",
                    },
                  },
                ],
              },
              transform: {
                cmd: ["sh"],
                image: "dpokidov/imagemagick:7.0.10-58",
                stdin: [
                  "montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png",
                ],
              },
              egress: {
                URL: "s3://test",
              },
            })}' | pachctl create pipeline`
          );
      });
  }

  return cy.exec('jq -r .pachyderm version.json').then(res => {
        cy.exec('pachctl create repo images')
          .exec('pachctl put file images@master:liberty.png -f http://imgur.com/46Q8nDz.png')
          .exec(`pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/v${res.stdout}/examples/opencv/edges.json`);
      });
});

Cypress.Commands.add('deleteReposAndPipelines', () => {
  return cy
    .exec('pachctl delete pipeline --all --force')
    .exec('pachctl delete repo --all --force')
})
