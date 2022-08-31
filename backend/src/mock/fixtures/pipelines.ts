import {
  CronInput,
  Egress,
  Input,
  PFSInput,
  Pipeline,
  PipelineInfo,
  PipelineState,
  SchedulingSpec,
  Transform,
  JobState,
  ParallelismSpec,
} from '@pachyderm/node-pachyderm';
import {
  ObjectStorageEgress,
  SQLDatabaseEgress,
} from '@pachyderm/node-pachyderm/dist/proto/pfs/pfs_pb';

import {DAGS} from './loadLimits';

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
        .setParallelismSpec(new ParallelismSpec().setConstant(8))
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
        .setSchedulingSpec(schedulingSpec)
        .setTransform(
          new Transform()
            .setCmdList(['sh'])
            .setImage('v4tech/imagemagick')
            .setStdinList([
              'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
            ]),
        ),
    )
    .setReason(
      'datum 64b95f0fe1a787b6c26ec7ede800be6f2b97616f3224592d91cbfe1cfccd00a1 failed',
    )
    .setState(PipelineState.PIPELINE_FAILURE),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('edges'))
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(new Input().setPfs(new PFSInput().setRepo('images')))
        .setDescription('Very cool edges description')
        .setOutputBranch('master')
        .setTransform(
          new Transform()
            .setCmdList(['python3', './edges.py'])
            .setImage('pachyderm/opencv'),
        ),
    )
    .setState(PipelineState.PIPELINE_RUNNING),
];

const egress = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('egress_s3'))
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setParallelismSpec(new ParallelismSpec().setConstant(8))
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(new PFSInput().setRepo('edges')),
            new Input().setPfs(new PFSInput().setRepo('images')),
          ]),
        )
        .setDescription('a pipeline with egress to an s3 bucket')
        .setOutputBranch('master')
        .setEgress(new Egress().setUrl('https://egress.com'))
        .setS3Out(true)
        .setSchedulingSpec(schedulingSpec)
        .setTransform(
          new Transform()
            .setCmdList(['sh'])
            .setImage('v4tech/imagemagick')
            .setStdinList([
              'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
            ]),
        ),
    )
    .setState(PipelineState.PIPELINE_FAILURE),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('egress_sql'))
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setParallelismSpec(new ParallelismSpec().setConstant(8))
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(new PFSInput().setRepo('edges')),
            new Input().setPfs(new PFSInput().setRepo('images')),
          ]),
        )
        .setDescription('a pipeline with egress to an sql database')
        .setOutputBranch('master')
        .setEgress(
          new Egress().setSqlDatabase(
            new SQLDatabaseEgress()
              .setUrl(
                'snowflake://pachyderm@WHMUWUD-CJ80657/PACH_DB/PUBLIC?warehouse=COMPUTE_WH',
              )
              .setFileFormat(
                new SQLDatabaseEgress.FileFormat().setType(
                  SQLDatabaseEgress.FileFormat.Type.CSV,
                ),
              ),
          ),
        )
        .setS3Out(true)
        .setSchedulingSpec(schedulingSpec)
        .setTransform(
          new Transform()
            .setCmdList(['sh'])
            .setImage('v4tech/imagemagick')
            .setStdinList([
              'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
            ]),
        ),
    )
    .setState(PipelineState.PIPELINE_FAILURE),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('egress_object'))
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setParallelismSpec(new ParallelismSpec().setConstant(8))
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(new PFSInput().setRepo('edges')),
            new Input().setPfs(new PFSInput().setRepo('images')),
          ]),
        )
        .setDescription('a pipeline with egress to object storage')
        .setOutputBranch('master')
        .setEgress(
          new Egress().setObjectStorage(
            new ObjectStorageEgress().setUrl(
              'object://pachyderm@WHMUWUD-CJ80657/PACH_DB/PUBLIC?warehouse=COMPUTE_WH',
            ),
          ),
        )
        .setS3Out(true)
        .setSchedulingSpec(schedulingSpec)
        .setTransform(
          new Transform()
            .setCmdList(['sh'])
            .setImage('v4tech/imagemagick')
            .setStdinList([
              'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
            ]),
        ),
    )
    .setState(PipelineState.PIPELINE_FAILURE),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('edges'))
    .setLastJobState(JobState.JOB_CREATED)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(new Input().setPfs(new PFSInput().setRepo('images')))
        .setDescription('Very cool edges description')
        .setOutputBranch('master')
        .setTransform(
          new Transform()
            .setCmdList(['python3', './edges.py'])
            .setImage('pachyderm/opencv'),
        ),
    )
    .setState(PipelineState.PIPELINE_RUNNING),
];

const customerTeam = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('likelihoods'))
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(new Input().setPfs(new PFSInput().setRepo('training')))

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('joint_call'))
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_RUNNING)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(new Input().setPfs(new PFSInput().setRepo('raw_data')))

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('model'))
    .setLastJobState(JobState.JOB_SUCCESS)
    .setState(PipelineState.PIPELINE_PAUSED)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(
          new Input().setCrossList([
            new Input().setPfs(new PFSInput().setRepo('split')),
            new Input().setPfs(
              new PFSInput().setRepo(
                'parameters_pachyderm_version_alternate_replicant',
              ),
            ),
          ]),
        )

        .setOutputBranch('master'),
    ),

  new PipelineInfo()
    .setPipeline(new Pipeline().setName('test'))
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setPipeline(
      new Pipeline().setName('detect_pachyderm_repo_version_alternate'),
    )
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details()
        .setInput(new Input().setCron(new CronInput().setRepo('cron')))
        .setOutputBranch('master'),
    ),
];

const traitDiscovery = [
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_orfs_blastdb'))
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('orfs')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_refseqfasta'))
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('reference_sequences')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_search'))
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_search')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_atg_fasta'))
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('atgs')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_completeness'))
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_candidates')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_candidate_fasta'))
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_candidates')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_candidate_bam'))
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_group_candidate_bam')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_contig_candidates'))
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_quality')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_hmmscan'))
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_clustering')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_geneclass_data'))
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput().setRepo('pachy_group_geneclass_data'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_patent_check'))
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_patent_search')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_promo_data'))
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_group_promo_data')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_group_promo_clstr'))
    .setLastJobState(JobState.JOB_SUCCESS)
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
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_group_promo_clstr')),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_promotion'))
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(
          new PFSInput().setRepo('pachy_trait_promoclstr_filter'),
        ),
      ),
    ),
  new PipelineInfo()
    .setPipeline(new Pipeline().setName('pachy_trait_atgs'))
    .setLastJobState(JobState.JOB_SUCCESS)
    .setDetails(
      new PipelineInfo.Details().setInput(
        new Input().setPfs(new PFSInput().setRepo('pachy_trait_promotion')),
      ),
    ),
];

const getLoadPipelines = (count: number) => {
  return [...new Array(count).keys()].map((i) => {
    return new PipelineInfo()
      .setPipeline(new Pipeline().setName(`load-pipeline-${i}`))
      .setDetails(
        new PipelineInfo.Details().setInput(
          new Input().setPfs(new PFSInput().setRepo(`load-repo-${i}`)),
        ),
      );
  });
};

const pipelines: {[projectId: string]: PipelineInfo[]} = {
  '1': tutorial,
  '2': customerTeam,
  '3': cron,
  '4': customerTeam,
  '5': egress,
  '6': [],
  '7': traitDiscovery,
  '8': [],
  '9': getLoadPipelines(DAGS),
  default: [...tutorial, ...customerTeam],
};

export default pipelines;
