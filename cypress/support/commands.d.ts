declare namespace Cypress {
  interface Chainable {
    isAppReady(): Chainable<any>;
    jupyterlabCommand(command: string): Chainable<any>;
  }
}
