const findFocusableChild = (child: ChildNode | null | undefined) => {
  if (!child) {
    return undefined;
  }

  if (child.nodeName === 'BUTTON') {
    return child as HTMLElement;
  }

  return (child as HTMLElement).querySelector(
    [
      'a:not(:disabled)',
      'button:not(:disabled)',
      'input:not(:disabled)',
      'textarea:not(:disabled)',
      'select:not(:disabled)',
      'details:not(:disabled)',
      '[tabindex]:not([tabindex="-1"]):not(:disabled)',
    ].join(','),
  ) as HTMLElement | null | undefined;
};

export default findFocusableChild;
