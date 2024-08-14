import graphlib from 'graphlib';
import dag from 'random-dag';

const opts = {
  max_per_rank: 6, // how 'fat' the DAG should be
  min_per_rank: 1,
  max_ranks: 5, // how 'tall' the DAG should be
  min_ranks: 1,
  probability: 0.3, // chance of having an edge
};

dag.graphlib(opts, (_err, graph) => {
  console.log(
    `nodes: ${graph.nodes().length} edges: ${graph.edges().length}
    
vvv COPY BELOW vvv
    `,
  );

  const edges = graph.edges();
  const pipelinesInOrder = graphlib.alg.components(graph);

  // create input repos
  const inputs = graph.sources();
  inputs.forEach((input) => {
    console.log(`pachctl create repo ${input}`);
  });

  // create pipelines in order of layers
  pipelinesInOrder.forEach((pipelineGroup) => {
    pipelineGroup.forEach((pipeline) => {
      const inputs = edges.filter((edge) => edge.w === pipeline);

      if (inputs.length === 1) {
        console.log(
          `echo '{"pipeline":{"name":"${pipeline}"},"input":{"pfs": {"glob": "/","repo": "${inputs[0].v}"}},"transform": {"cmd": ["sh"], "stdin": ["sleep 0"]}}' | pachctl create pipeline`,
        );
      } else if (inputs.length > 1) {
        const cross = inputs
          .map((input) => `{"pfs": {"glob": "/","repo": "${input.v}"}}`)
          .join(',');
        console.log(
          `echo '{"pipeline":{"name":"${pipeline}"},"input":{"cross":[${cross}]},"transform": {"cmd": ["sh"], "stdin": ["sleep 0"]}}' | pachctl create pipeline`,
        );
      }
    });
  });
});
