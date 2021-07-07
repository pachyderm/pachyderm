import {
  CronInput,
  Egress,
  Input,
  JobState,
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
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(new PFSInput().setRepo('edges')),
            new Input().setPfs(new PFSInput().setRepo('images')),
          ]),
        )
        .setDescription('Not my favorite pipeline')
        .setOutputBranch('master')
        .setEgress(new Egress().setUrl('https://egress.com'))
        .setS3Out(true)
        .setSchedulingSpec(schedulingSpec),
    )
    .setState(PipelineState.PIPELINE_FAILURE),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('edges'))
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(new Input().setPfs(new PFSInput().setRepo('images')))
        .setDescription('Very cool edges description')
        .setOutputBranch('master'),
    )
    .setState(PipelineState.PIPELINE_RUNNING),
];

const customerTeam = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('likelihoods'))
    .setState(PipelineState.PIPELINE_STANDBY)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(new PFSInput().setRepo('samples')),
            new Input().setPfs(new PFSInput().setRepo('reference')),
          ]),
        )

        .setOutputBranch('master'),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('models'))
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(new Input().setPfs(new PFSInput().setRepo('training')))

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('joint_call'))
    .setState(PipelineState.PIPELINE_FAILURE)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(new PFSInput().setRepo('reference')),
            new Input().setPfs(new PFSInput().setRepo('likelihoods')),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('split'))
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(new Input().setPfs(new PFSInput().setRepo('raw_data')))

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('model'))
    .setState(PipelineState.PIPELINE_PAUSED)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(new PFSInput().setRepo('split')),
            new Input().setPfs(new PFSInput().setRepo('parameters')),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('test'))
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(new PFSInput().setRepo('split')),
            new Input().setPfs(new PFSInput().setRepo('model')),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('select'))
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(new PFSInput().setRepo('test')),
            new Input().setPfs(new PFSInput().setRepo('model')),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('detect'))
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(new PFSInput().setRepo('model')),
            new Input().setPfs(new PFSInput().setRepo('images')),
          ]),
        )

        .setOutputBranch('master'),
    ),
];

const cron = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('processor'))
    .setDetails(
      new PipelineInfo.Details()
        .setInput(new Input().setCron(new CronInput().setRepo('cron')))
        .setOutputBranch('master'),
    ),
];

const traitDiscovery = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_orfs_blastdb'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('orfs')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_refseqfasta'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('reference_sequences')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_search'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(new PFSInput().setRepo('pachy_orfs_blastdb')),
          new Input().setPfs(new PFSInput().setRepo('pachy_trait_refseqfasta')),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_candidates'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_search')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_atg_fasta'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('atgs')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_completeness'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_candidates')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_candidate_fasta'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_candidates')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_candidate_bam'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(new PFSInput().setRepo('assembly_bam_files')),
          new Input().setPfs(new PFSInput().setRepo('pachy_trait_candidates')),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_clustering'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(new PFSInput().setRepo('pachy_atg_fasta')),
          new Input().setPfs(
            new PFSInput().setRepo('pachy_trait_candidate_fasta'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_quality_downselect'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_group_candidate_bam')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_contig_candidates'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(new PFSInput().setRepo('pachy_trait_candidates')),
          new Input().setPfs(new PFSInput().setRepo('pachy_atg_fasta')),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_quality'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput().setRepo('pachy_trait_quality_downselect'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_neighbors'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput().setRepo('pachy_group_contig_candidates'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_domainscan'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput().setRepo('pachy_trait_candidate_fasta'),
          ),
          new Input().setPfs(new PFSInput().setRepo('inter_pro_scan')),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_quality_check'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_quality')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_hmmscan'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput().setRepo('pachy_trait_candidate_fasta'),
          ),
          new Input().setPfs(new PFSInput().setRepo('custom_hmms')),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_promotion_status'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_clustering')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_geneclass_data'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(new PFSInput().setRepo('pachy_trait_domainscan')),
          new Input().setPfs(new PFSInput().setRepo('pachy_trait_hmmscan')),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_patent_search'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput().setRepo('pachy_trait_candidate_fasta'),
          ),
          new Input().setPfs(new PFSInput().setRepo('patent_databases')),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_geneclass'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput().setRepo('pachy_group_geneclass_data'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_patent_check'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_patent_search')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_promo_data'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput().setRepo('pachy_trait_quality_check'),
          ),
          new Input().setPfs(
            new PFSInput().setRepo('pachy_trait_completeness'),
          ),
          new Input().setPfs(new PFSInput().setRepo('pachy_trait_neighbors')),
          new Input().setPfs(new PFSInput().setRepo('pachy_trait_candidates')),
          new Input().setPfs(new PFSInput().setRepo('pachy_trait_geneclass')),
          new Input().setPfs(
            new PFSInput().setRepo('pachy_trait_patent_check'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_promotionfilter'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_group_promo_data')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_promo_clstr'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setCrossList([
          new Input().setPfs(
            new PFSInput().setRepo('pachy_trait_promotionfilter'),
          ),
          new Input().setPfs(
            new PFSInput().setRepo('pachy_trait_promotion_status'),
          ),
        ]),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_promoclstr_filter'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_group_promo_clstr')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_promotion'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput().setRepo('pachy_trait_promoclstr_filter'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_atgs'))
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_promotion')),
      ),
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
