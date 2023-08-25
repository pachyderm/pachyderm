import {ApolloLink, FetchResult, Observable} from '@apollo/client';
import {addBreadcrumb, Breadcrumb, configureScope, Scope} from '@sentry/react';

const sentryLink = () =>
  new ApolloLink((operation, forward) => {
    const breadcrumb: Breadcrumb = {
      category: operation.query.definitions[0].kind,
      data: {
        query: operation.query.loc?.source.body || '',
      },
      level: 'log',
      message: operation.operationName,
    };

    configureScope((scope: Scope) => {
      scope.setTransactionName(operation.operationName);
      scope.setFingerprint(['{{default}}', '{{transaction}}']);
    });

    return new Observable<FetchResult>((observer) => {
      const subscription = forward(operation).subscribe({
        next: (result: FetchResult) => {
          observer.next(result);
        },
        complete: () => {
          addBreadcrumb(breadcrumb);
          observer.complete();
        },
        error: (error: unknown) => {
          if (breadcrumb.data) breadcrumb.data.error = JSON.stringify(error);
          breadcrumb.level = 'error';
          breadcrumb.type = 'error';
          addBreadcrumb(breadcrumb);
          return observer.error(error);
        },
      });

      return () => {
        if (subscription) subscription.unsubscribe();
      };
    });
  });

export default sentryLink;
