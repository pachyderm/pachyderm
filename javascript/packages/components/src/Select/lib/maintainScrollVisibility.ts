const maintainScrollVisibility = (
  activeElement: HTMLElement | null,
  scrollParent: HTMLElement | null,
) => {
  if (!activeElement || !scrollParent) return;
  const {offsetHeight, offsetTop} = activeElement;
  const {offsetHeight: parentOffsetHeight, scrollTop} = scrollParent;

  const isAbove = offsetTop < scrollTop;
  const isBelow = offsetTop + offsetHeight > scrollTop + parentOffsetHeight;

  if (isAbove) {
    scrollParent.scrollTo(0, offsetTop);
  } else if (isBelow) {
    scrollParent.scrollTo(0, offsetTop - parentOffsetHeight + offsetHeight);
  }
};

export default maintainScrollVisibility;
