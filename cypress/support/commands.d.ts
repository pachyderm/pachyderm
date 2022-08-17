declare namespace Cypress {
  interface Chainable {
    login(email?: string, password?: string): Chainable<any>;
    logout(): Chainable<any>;
    authenticatePachctl(): Chainable<any>;
    setupProject(projectTemplate?: string): Chainable<any>;
    deleteReposAndPipelines(): Chainable<any>;
    isInViewport(element: () => Cypress.Chainable<JQuery<HTMLElement>>): Chainable<any>;
  }
}
