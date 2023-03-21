import {Branch, Project, Repo, RepoInfo} from '@dash-backend/proto';
import {timestampFromObject} from '@dash-backend/proto/builders/protobuf';

import {DAGS} from './loadLimits';

const tutorial = [
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('montage')
        .setType('user')
        .setProject(new Project().setName('Solar-Panel-Data-Sorting')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614136189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('edges')
        .setType('user')
        .setProject(new Project().setName('Solar-Panel-Data-Sorting')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('images')
        .setType('user')
        .setProject(new Project().setName('Solar-Panel-Data-Sorting')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614116189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

// clone of tutorial and customer team
const defaultRepos = [
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('montage')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614136189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('edges')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('images')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614116189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),

  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('samples')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1615326189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('likelihoods')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1615526189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('reference')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1615426189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('training')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1615026189, nanos: 0}))
    .setBranchesList([
      new Branch().setName('master'),
      new Branch().setName('develop'),
      new Branch().setName('test'),
    ]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('models')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1615126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('joint_call')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1615626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('raw_data')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1615626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('split')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1614726189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('parameters_pachyderm_version_alternate_replicant')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1614626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('model')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1614526189, nanos: 0}))
    .setBranchesList([new Branch().setName('develop')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('test')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1614426189, nanos: 0})),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('select')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1614326189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('detect_pachyderm_repo_version_alternate')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1614226189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('images')
        .setType('user')
        .setProject(new Project().setName('default')),
    )
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

// clone of tutorial
const openCVTutorial = [
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('montage')
        .setType('user')
        .setProject(new Project().setName('OpenCV-Tutorial')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614136189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('edges')
        .setType('user')
        .setProject(new Project().setName('OpenCV-Tutorial')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('images')
        .setType('user')
        .setProject(new Project().setName('OpenCV-Tutorial')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614116189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

const egress = [
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('egress_sql')
        .setType('user')
        .setProject(new Project().setName('Egress-Examples')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614136189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('egress_object')
        .setType('user')
        .setProject(new Project().setName('Egress-Examples')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614136189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('egress_s3')
        .setType('user')
        .setProject(new Project().setName('Egress-Examples')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614136189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('edges')
        .setType('user')
        .setProject(new Project().setName('Egress-Examples')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('images')
        .setType('user')
        .setProject(new Project().setName('Egress-Examples')),
    )
    .setDetails(new RepoInfo.Details().setSizeBytes(1000))
    .setCreated(timestampFromObject({seconds: 1614116189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

const customerTeam = [
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('samples')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1615326189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('likelihoods')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1615526189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('reference')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1615426189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('training')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1615026189, nanos: 0}))
    .setBranchesList([
      new Branch().setName('master'),
      new Branch().setName('develop'),
      new Branch().setName('test'),
    ]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('models')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1615126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('joint_call')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1615626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('raw_data')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1615626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('split')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1614726189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('parameters_pachyderm_version_alternate_replicant')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1614626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('model')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1614526189, nanos: 0}))
    .setBranchesList([new Branch().setName('develop')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('test')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1614426189, nanos: 0})),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('select')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1614326189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('detect_pachyderm_repo_version_alternate')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1614226189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('images')
        .setType('user')
        .setProject(new Project().setName('Data-Cleaning-Process')),
    )
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

// clone of customerTeam
const solarPricePredictionModal = [
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('samples')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1615326189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('likelihoods')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1615526189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('reference')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1615426189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('training')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1615026189, nanos: 0}))
    .setBranchesList([
      new Branch().setName('master'),
      new Branch().setName('develop'),
      new Branch().setName('test'),
    ]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('models')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1615126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('joint_call')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1615626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('raw_data')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1615626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('split')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1614726189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('parameters_pachyderm_version_alternate_replicant')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1614626189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('model')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1614526189, nanos: 0}))
    .setBranchesList([new Branch().setName('develop')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('test')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1614426189, nanos: 0})),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('select')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1614326189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('detect_pachyderm_repo_version_alternate')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1614226189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('images')
        .setType('user')
        .setProject(new Project().setName('Solar-Price-Prediction-Modal')),
    )
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

const cron = [
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('cron')
        .setType('user')
        .setProject(
          new Project().setName('Solar-Power-Data-Logger-Team-Collab'),
        ),
    )
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setDescription('cron job')
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858))
    .setSizeBytesUpperBound(621858),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('processor')
        .setType('user')
        .setProject(
          new Project().setName('Solar-Power-Data-Logger-Team-Collab'),
        ),
    )
    .setCreated(timestampFromObject({seconds: 1614226189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')]),
];

const traitDiscovery = [
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('assembly_bam_files')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('orfs')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126199, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('reference_sequences')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126200, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('atgs')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126205, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('inter_pro_scan')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126210, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('custom_hmms')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126215, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('patent_databases')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126220, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_orfs_blastdb')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126225, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_refseqfasta')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126230, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_search')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126235, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_candidates')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126240, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_atg_fasta')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126245, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_completeness')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126250, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_candidate_fasta')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126255, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_group_candidate_bam')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126260, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_clustering')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126265, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_quality_downselect')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126270, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_domainscan')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126280, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_quality')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126285, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_neighbors')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126290, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_quality_check')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126295, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_hmmscan')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126300, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_promotion_status')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126305, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_group_geneclass_data')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126310, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_patent_search')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126315, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_geneclass')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126320, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_patent_check')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126325, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_group_promo_data')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126330, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_promotionfilter')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126335, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_group_promo_clstr')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126340, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_promoclstr_filter')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126345, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_promotion')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126350, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_group_contig_candidates')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126360, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('pachy_trait_atgs')
        .setType('user')
        .setProject(new Project().setName('Trait-Discovery')),
    )
    .setCreated(timestampFromObject({seconds: 1614126360, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
];

const multiProjectPipelineA = [
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('Node_1')
        .setType('user')
        .setProject(new Project().setName('Multi-Project-Pipeline-A')),
    )
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
];

const multiProjectPipelineB = [
  new RepoInfo()
    .setRepo(
      new Repo()
        .setName('Node_1')
        .setType('user')
        .setProject(new Project().setName('Multi-Project-Pipeline-B')),
    )
    .setCreated(timestampFromObject({seconds: 1614126189, nanos: 0}))
    .setBranchesList([new Branch().setName('master')])
    .setDetails(new RepoInfo.Details().setSizeBytes(621858)),
];

const getLoadRepos = (count: number) => {
  return [...new Array(count).keys()].reduce((repos: RepoInfo[], i) => {
    const now = Math.floor(new Date().getTime() / 1000);
    repos.push(
      new RepoInfo()
        .setRepo(new Repo().setName(`load-repo-${i}`).setType('user'))
        .setCreated(timestampFromObject({seconds: now, nanos: 0}))
        .setBranchesList([new Branch().setName('master')])
        .setDetails(
          new RepoInfo.Details().setSizeBytes(Math.floor(Math.random() * 1000)),
        ),
    );
    repos.push(
      new RepoInfo()
        .setRepo(new Repo().setName(`load-pipeline-${i}`).setType('user'))
        .setCreated(timestampFromObject({seconds: now, nanos: 0}))
        .setBranchesList([new Branch().setName('master')])
        .setDetails(
          new RepoInfo.Details().setSizeBytes(Math.floor(Math.random() * 1000)),
        ),
    );
    return repos;
  }, []);
};

const repos: {[projectId: string]: RepoInfo[]} = {
  'Solar-Panel-Data-Sorting': tutorial,
  'Data-Cleaning-Process': customerTeam,
  'Solar-Power-Data-Logger-Team-Collab': cron,
  'Solar-Price-Prediction-Modal': solarPricePredictionModal,
  'Egress-Examples': egress,
  'Empty-Project': [],
  'Trait-Discovery': traitDiscovery,
  'OpenCV-Tutorial': openCVTutorial,
  'Load-Project': getLoadRepos(DAGS),
  default: defaultRepos,
  'Multi-Project-Pipeline-A': multiProjectPipelineA,
  'Multi-Project-Pipeline-B': multiProjectPipelineB,
};

export default repos;
