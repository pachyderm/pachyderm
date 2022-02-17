declare namespace Cypress {
  interface Chainable {
    login(email?: string, password?: string): Chainable<any>;
    logout(): Chainable<any>;
    authenticatePachctl(): Chainable<any>;
    setupProject(): Chainable<any>;
    deleteReposAndPipelines(): Chainable<any>;
  }
}
