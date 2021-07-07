import {Branch, Repo, RepoInfo} from '@pachyderm/proto/pb/pfs/pfs_pb';

import {timestampFromObject} from '@dash-backend/grpc/builders/protobuf';

const tutorial = [
  new RepoInfo()
    .setRepo(new Repo().setName('montage').setType('user'))
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614136189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('edges').setType('user'))
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('images').setType('user'))
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614116189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

const customerTeam = [
  new RepoInfo()
    .setRepo(new Repo().setName('samples').setType('user'))
    .setCreated(timestampFromObject({seconds: 1615326189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('likelihoods').setType('user'))
    .setCreated(timestampFromObject({seconds: 1615526189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('reference').setType('user'))
    .setCreated(timestampFromObject({seconds: 1615426189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('training').setType('user'))
    .setCreated(timestampFromObject({seconds: 1615026189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('models').setType('user'))
    .setCreated(timestampFromObject({seconds: 1615126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('joint_call').setType('user'))
    .setCreated(timestampFromObject({seconds: 1615626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('raw_data').setType('user'))
    .setCreated(timestampFromObject({seconds: 1615626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('split').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614726189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('parameters').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('model').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614526189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('test').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614426189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('select').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614326189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('detect').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614226189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(new Repo().setName('images').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

const cron = [
  new RepoInfo()
    .setRepo(new Repo().setName('cron').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('processor').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614226189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

const traitDiscovery = [
  new RepoInfo()
    .setRepo(new Repo().setName('assembly_bam_files').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('orfs').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126199, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('reference_sequences').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126200, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('atgs').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126205, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('inter_pro_scan').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126210, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('custom_hmms').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126215, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('patent_databases').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126220, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_orfs_blastdb').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126225, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_refseqfasta').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126230, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_search').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126235, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_candidates').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126240, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_atg_fasta').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126245, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_completeness').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126250, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_candidate_fasta').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126255, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_group_candidate_bam').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126260, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_clustering').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126265, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo().setName('pachy_trait_quality_downselect').setType('user'),
    )
    .setCreated(timestampFromObject({seconds: 1614126270, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_domainscan').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126280, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_quality').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126285, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_neighbors').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126290, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_quality_check').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126295, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_hmmscan').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126300, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_promotion_status').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126305, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_group_geneclass_data').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126310, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_patent_search').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126315, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_geneclass').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126320, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_patent_check').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126325, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_group_promo_data').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126330, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_promotionfilter').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126335, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_group_promo_clstr').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126340, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo().setName('pachy_trait_promoclstr_filter').setType('user'),
    )
    .setCreated(timestampFromObject({seconds: 1614126345, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_promotion').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126350, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo().setName('pachy_group_contig_candidates').setType('user'),
    )
    .setCreated(timestampFromObject({seconds: 1614126360, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(new Repo().setName('pachy_trait_atgs').setType('user'))
    .setCreated(timestampFromObject({seconds: 1614126360, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
];

const repos: {[projectId: string]: RepoInfo[]} = {
  '1': tutorial,
  '2': customerTeam,
  '3': cron,
  '4': customerTeam,
  '5': tutorial,
  '6': [],
  '7': traitDiscovery,
  default: [...tutorial, ...customerTeam],
};

export default repos;
