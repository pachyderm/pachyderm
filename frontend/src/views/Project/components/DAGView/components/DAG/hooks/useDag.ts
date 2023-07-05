const useDag = () => {
  const [, translateX, translateY, scale] =
    document
      .getElementById('Dags')
      ?.getAttribute('transform')
      ?.split(/translate\(|\) scale\(|\)|,/)
      ?.map((s) => parseFloat(s)) || [];
  const svgWidth = document.getElementById('Svg')?.clientWidth || 0;
  const svgHeight = document.getElementById('Svg')?.clientHeight || 0;

  return {translateX, translateY, scale, svgWidth, svgHeight};
};

export default useDag;
