# Variant Discovery with GATK

![alt tag](pipeline.png)

This example illustrates the use of GATK in Pachyderm for
[Germline](https://en.wikipedia.org/wiki/Germline) variant calling and joint
[genotyping](https://en.wikipedia.org/wiki/Genotyping).

GATK is a toolkit provided by the Broad Institute that was created to assist
scientists with variant discovery and genotyping. Each stage of this GATK best
practice pipeline can be scaled individually and is automatically triggered as
data flows into the top of the pipeline. The example follows
[this tutorial](https://drive.google.com/open?id=0BzI1CyccGsZiQ1BONUxfaGhZRGc)
from GATK, which includes more details about the various stages.

## Committing the reference genome

You can retrieve all the input data for this pipeline from the Broad Institute
[here](https://drive.google.com/open?id=0BzI1CyccGsZicE5HNkR6anpLTnM). We will
utilize a b37 human genome reference containing only a subset of chromosome 20,
which we prepared specially for GATK tutorials in order to provide a reasonable
size for download. It is accompanied by its index and sequence dictionary.

Download `GATK_Germline.zip` and unzip it to your local dir.

```sh
wget https://s3-us-west-1.amazonaws.com/pachyderm.io/Examples_Data_Repo/GATK_Germline.zip
```

```sh
$ unzip GATK_Germline.zip
Archive:  GATK_Germline.zip
   creating: data/
  inflating: data/.DS_Store
   creating: __MACOSX/
   creating: __MACOSX/data/
  inflating: __MACOSX/data/._.DS_Store
  inflating: data/.Rhistory
   creating: data/bams/
  inflating: data/bams/.DS_Store
  etc...
```

Change into the directory that contains the files we want to import

```sh
$ cd data/ref
$ ls
Icon  ref.dict  ref.fasta  ref.fasta.fai  refSDF
```

Next, we want to create our pachyderm repo and then instruct pachyderm to import
those into our repo

```sh
$ pachctl create repo reference
$ pachctl put file reference@master -r -f .
```

First milestone reached! Lets just check and make sure everything looks good

```sh
$ pachctl list repo
NAME                CREATED             SIZE
reference           43 seconds ago      83.68MiB
```

```sh
$ pachctl list file reference@master
NAME                TYPE                SIZE
Icon                file                0B
ref.dict            file                164B
ref.fasta           file                61.11MiB
ref.fasta.fai       file                20B
refSDF              dir                 22.57MiB
$ cd ../../
```

## Committing a sample

Next, we're going to work on commiting the `*.bam` files for Mom. Let's start by
create a repositories for input `*.bam` files to go into:

```sh
$ pachctl create repo samples
```

Add a `*.bam` file (along with it's index file) corresponding to a first sample,
in our case it's the `mother`. Here we will assume that the files corresponding
to each sample are committed to separate directories (e.g., `/mother`).

```sh
$ cd data/bams/
$ pachctl start commit samples@master
$ for f in $(ls mother.*); do pachctl put file samples@master:mother/$f -f $f; done
$ pachctl finish commit samples@master
$ cd ../../
```

You should then be able to see the versioned sample data in Pachyderm:

```sh
$ pachctl list file samples@master
NAME                TYPE                SIZE
mother              dir                 23.79MiB
$ pachctl list file samples@master:mother
NAME                TYPE                SIZE
mother/mother.bai   file                9.047KiB
mother/mother.bam   file                23.79MiB
```

## Variant calling

To call variants for the input sample, we will run the `HaplotypeCaller` using
GATK. Details of the exact GATK command and related scripting are included in
the [likelihoods.json](likelihoods.json) pipeline specification.

To create and run the variant calling pipeline to generate genotype likelihoods:

```sh
$ pachctl create pipeline -f likelihoods.json
```

This will automatically trigger a job:

```sh
$ pachctl list job
ID                                   OUTPUT COMMIT                                STARTED        DURATION   RESTART PROGRESS  DL       UL       STATE
c61c71d1-6544-48ad-8361-b4ad155ba1a0 likelihoods/992393004c5a45c0a35995cf0179f1cb 43 minutes ago 18 seconds 0       1 + 0 / 1 107.5MiB 4.667MiB success
```

And you can view the output of the variant calling as follows:

```sh
$ pachctl list file likelihoods@master
NAME                TYPE                SIZE
mother.g.vcf        file                4.667MiB
mother.g.vcf.idx    file                758B
```

## Joint genotyping

The last step is to joint call all your GVCF files using the GATK tool
GenotypeGVCFs. Details of this GATK command and related scripting are included
in the [joint-call.json](joint-call.json) pipeline specification.

To run the joint genotyping:

```sh
$ pachctl create pipeline -f joint_call.json
```

This will automatically trigger a job and produce our final output:

```sh
$ pachctl list job
ID                                   OUTPUT COMMIT                                STARTED        DURATION   RESTART PROGRESS  DL       UL       STATE
67135f10-4121-4f29-a30b-1eaf6ffe2194 joint_call/c4ebd6dd0c764a97a8d7f3a71f6bb9ce  38 minutes ago 5 seconds  0       1 + 0 / 1 88.35MiB 113.9KiB success
c61c71d1-6544-48ad-8361-b4ad155ba1a0 likelihoods/992393004c5a45c0a35995cf0179f1cb 43 minutes ago 18 seconds 0       1 + 0 / 1 107.5MiB 4.667MiB success
$ pachctl list file joint_call@master
NAME                TYPE                SIZE
joint.vcf           file                103.7KiB
joint.vcf.idx       file                10.21KiB
```

## Adding more samples

Now that we have our pipelines running, out final results will be automatically
updated any time we add new samples. To illustrate this, we can add the `father`
and `son` samples as follows:

```sh
$ cd data/bams/
$ pachctl start commit samples@master
dc963cc9bdc2486798b92d20eead5058
$ for f in $(ls father.*); do pachctl put file samples@master:father/$f -f $f; done
$ pachctl finish commit samples@master
$ pachctl start commit samples@master
84e6615de64f43d8815909fa978bd4bc
$ for f in $(ls son.*); do pachctl put file samples@master:son/$f -f $f; done
$ pachctl finish commit samples@master
$ pachctl list file samples@master
NAME                TYPE                SIZE
father              dir                 9.662MiB
mother              dir                 23.79MiB
son                 dir                 9.58MiB
```

This will trigger new jobs to process the new samples:

```sh
pachctl list job
ID                                   OUTPUT COMMIT                                STARTED            DURATION   RESTART PROGRESS  DL       UL       STATE
73222c06-c444-4dff-b370-6c7dea83258d joint_call/32d3615a036e4c0eadd3ed49435ee7db  About a minute ago 6 seconds  0       1 + 0 / 1 97.64MiB 188.6KiB success
47652b28-a1ba-454e-a74c-7b43740a72f0 likelihoods/a7e160f8418a4bd5a46233bd6b1d84ce 2 minutes ago      16 seconds 0       1 + 2 / 3 93.26MiB 4.729MiB success
55b17b8c-b305-4d4c-b822-f861578dadd8 joint_call/4eb7c19600134af5b9595bf5d9f71edc  2 minutes ago      5 seconds  0       1 + 0 / 1 92.92MiB 166.3KiB success
a350a349-5ddb-4e19-bdda-66d7edbf9447 likelihoods/e783989ca367428ea2df6406a23bea6c 2 minutes ago      19 seconds 0       1 + 1 / 2 93.35MiB 4.564MiB success
67135f10-4121-4f29-a30b-1eaf6ffe2194 joint_call/c4ebd6dd0c764a97a8d7f3a71f6bb9ce  About an hour ago  5 seconds  0       1 + 0 / 1 88.35MiB 113.9KiB success
c61c71d1-6544-48ad-8361-b4ad155ba1a0 likelihoods/992393004c5a45c0a35995cf0179f1cb About an hour ago  18 seconds 0       1 + 0 / 1 107.5MiB 4.667MiB success
```

If you are using the
[Enterprise Edition](https://docs.pachyderm.com/latest/enterprise/), you should
be able to see the DAG and data as shown below:

![alt tag](dash.png)
