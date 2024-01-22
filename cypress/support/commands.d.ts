declare namespace Cypress {
  interface Chainable {
    viewAllLandingPageProjects(): Chainable<any>;
    login(email?: string, password?: string): Chainable<any>;
    logout(): Chainable<any>;
    multiLineExec(stringInputs: string): Chainable<any>;
    setupProject(projectTemplate?: string): Chainable<any>;
    deleteReposAndPipelines(): Chainable<any>;
    isInViewport(
      element: () => Cypress.Chainable<JQuery<HTMLElement>>,
    ): Chainable<any>;
  }
}
