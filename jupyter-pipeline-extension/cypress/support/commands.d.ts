declare namespace Cypress {
  interface Chainable {
    resetApp(): void;
    isAppReady(): Chainable<any>;
    jupyterlabCommand(command: string): Chainable<any>;
    unmountAllRepos(): void;
    openMountPlugin(): void;
  }
}
