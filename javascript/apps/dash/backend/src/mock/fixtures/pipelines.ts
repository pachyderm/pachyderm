import {
  CronInput,
  Egress,
  Input,
  PFSInput,
  Pipeline,
  PipelineInfo,
  PipelineState,
  SchedulingSpec,
} from '@pachyderm/proto/pb/pps/pps_pb';

// Need to define this up here, as the node selector
// map is a mutable set that can't be initialized with
// values
const schedulingSpec = new SchedulingSpec();
schedulingSpec.setPriorityClassName('high-priority');
schedulingSpec.getNodeSelectorMap().set('disktype', 'ssd');

const tutorial = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('montage'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('edges')),
        new Input().setPfs(new PFSInput().setRepo('images')),
      ]),
    )
    .setDescription('Not my favorite pipeline')
    .setState(PipelineState.PIPELINE_FAILURE)
    .setOutputBranch('master')
    .setCacheSize('64M')
    .setEgress(new Egress().setUrl('https://egress.com'))
    .setS3Out(true)
    .setSchedulingSpec(schedulingSpec),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('edges'))
    .setInput(new Input().setPfs(new PFSInput().setRepo('images')))
    .setDescription('Very cool edges description')
    .setState(PipelineState.PIPELINE_RUNNING)
    .setOutputBranch('master')
    .setCacheSize('64M'),
];

const customerTeam = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('likelihoods'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('samples')),
        new Input().setPfs(new PFSInput().setRepo('reference')),
      ]),
    )
    .setState(PipelineState.PIPELINE_STANDBY)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('models'))
    .setInput(new Input().setPfs(new PFSInput().setRepo('training')))
    .setState(PipelineState.PIPELINE_RUNNING)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('joint_call'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('reference')),
        new Input().setPfs(new PFSInput().setRepo('likelihoods')),
      ]),
    )
    .setState(PipelineState.PIPELINE_FAILURE)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('split'))
    .setInput(new Input().setPfs(new PFSInput().setRepo('raw_data')))
    .setState(PipelineState.PIPELINE_RUNNING)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('model'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('split')),
        new Input().setPfs(new PFSInput().setRepo('parameters')),
      ]),
    )
    .setState(PipelineState.PIPELINE_PAUSED)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('test'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('split')),
        new Input().setPfs(new PFSInput().setRepo('model')),
      ]),
    )
    .setState(PipelineState.PIPELINE_RUNNING)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('select'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('test')),
        new Input().setPfs(new PFSInput().setRepo('model')),
      ]),
    )
    .setState(PipelineState.PIPELINE_RUNNING)
    .setOutputBranch('master')
    .setCacheSize('64M'),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('detect'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('model')),
        new Input().setPfs(new PFSInput().setRepo('images')),
      ]),
    )
    .setState(PipelineState.PIPELINE_RUNNING)
    .setOutputBranch('master')
    .setCacheSize('64M'),
];

const cron = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('processor'))
    .setInput(new Input().setCron(new CronInput().setRepo('cron')))
    .setOutputBranch('master')
    .setCacheSize('64M'),
];

const traitDiscovery = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_orfs_blastdb'))
    .setInput(new Input().setPfs(new PFSInput().setRepo('orfs'))),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_refseqfasta'))
    .setInput(
      new Input().setPfs(new PFSInput().setRepo('reference_sequences')),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_search'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('pachy_orfs_blastdb')),
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_refseqfasta')),
      ]),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_candidates'))
    .setInput(new Input().setPfs(new PFSInput().setRepo('pachy_trait_search'))),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_atg_fasta'))
    .setInput(new Input().setPfs(new PFSInput().setRepo('atgs'))),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_completeness'))
    .setInput(
      new Input().setPfs(new PFSInput().setRepo('pachy_trait_candidates')),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_candidate_fasta'))
    .setInput(
      new Input().setPfs(new PFSInput().setRepo('pachy_trait_candidates')),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_candidate_bam'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('assembly_bam_files')),
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_candidates')),
      ]),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_clustering'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('pachy_atg_fasta')),
        new Input().setPfs(
          new PFSInput().setRepo('pachy_trait_candidates_fasta'),
        ),
      ]),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_quality_downselect'))
    .setInput(
      new Input().setPfs(new PFSInput().setRepo('pachy_group_candidate_bam')),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_contig_candidates'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_candidates')),
        new Input().setPfs(new PFSInput().setRepo('pachy_atg_fasta')),
      ]),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_quality'))
    .setInput(
      new Input().setPfs(
        new PFSInput().setRepo('pachy_trait_quality_downselect'),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_neighbors'))
    .setInput(
      new Input().setPfs(
        new PFSInput().setRepo('pachy_group_contig_candidates'),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_domainscan'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(
          new PFSInput().setRepo('pachy_trait_candidate_fasta'),
        ),
        new Input().setPfs(new PFSInput().setRepo('inter_pro_scan')),
      ]),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_quality_check'))
    .setInput(
      new Input().setPfs(new PFSInput().setRepo('pachy_trait_quality')),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_hmmscan'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(
          new PFSInput().setRepo('pachy_trait_candidate_fasta'),
        ),
        new Input().setPfs(new PFSInput().setRepo('custom_hmms')),
      ]),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_promotion_status'))
    .setInput(
      new Input().setPfs(new PFSInput().setRepo('pachy_trait_clustering')),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_geneclass_data'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_domainscan')),
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_hmmscan')),
      ]),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_patent_search'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(
          new PFSInput().setRepo('pachy_trait_candidate_fasta'),
        ),
        new Input().setPfs(new PFSInput().setRepo('patent_databases')),
      ]),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_geneclass'))
    .setInput(
      new Input().setPfs(new PFSInput().setRepo('pachy_group_geneclass_data')),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_patent_check'))
    .setInput(
      new Input().setPfs(new PFSInput().setRepo('pachy_patent_search')),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_promo_data'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_quality_check')),
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_completeness')),
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_neighbors')),
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_candidates')),
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_geneclass')),
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_patent_check')),
      ]),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_promotionfilter'))
    .setInput(
      new Input().setPfs(new PFSInput().setRepo('pachy_group_promo_data')),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_promo_clstr'))
    .setInput(
      new Input().setCrossList([
        new Input().setPfs(
          new PFSInput().setRepo('pachy_trait_promotionfilter'),
        ),
        new Input().setPfs(
          new PFSInput().setRepo('pachy_trait_promotion_status'),
        ),
      ]),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_promoclstr_filter'))
    .setInput(
      new Input().setPfs(new PFSInput().setRepo('pachy_group_promo_clstr')),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_promotion'))
    .setInput(
      new Input().setPfs(
        new PFSInput().setRepo('pachy_trait_promoclstr_filter'),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_atgs'))
    .setInput(
      new Input().setPfs(new PFSInput().setRepo('pachy_trait_promotion')),
    ),
];

const pipelines: {[projectId: string]: PipelineInfo[]} = {
  '1': tutorial,
  '2': customerTeam,
  '3': cron,
  '4': customerTeam,
  '5': tutorial,
  '6': [],
  '7': traitDiscovery,
  default: [...tutorial, ...customerTeam],
};

export default pipelines;
