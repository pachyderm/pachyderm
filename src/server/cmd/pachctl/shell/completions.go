package shell

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/clientsdk"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pretty"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pps_pretty "github.com/pachyderm/pachyderm/v2/src/server/pps/pretty"

	prompt "github.com/c-bata/go-prompt"
	units "github.com/docker/go-units"
)

type partEnum int

const (
	repoPart partEnum = iota
	commitOrBranchPart
	filePart
)

func parsePart(text string) partEnum {
	switch {
	case !strings.ContainsRune(text, '@'):
		return repoPart
	case !strings.ContainsRune(text, ':'):
		return commitOrBranchPart
	default:
		return filePart
	}
}

func samePart(p partEnum) CacheFunc {
	return func(state CacheState) bool {
		return parsePart(state.Text) == p
	}
}

var (
	pachClient     *client.APIClient
	pachClientOnce sync.Once
)

func getPachClient() *client.APIClient {
	pachClientOnce.Do(func() {
		c, err := client.NewOnUserMachine("user-completion")
		if err != nil {
			log.Fatal(err)
		}
		pachClient = c
	})
	return pachClient
}

func closePachClient() error {
	if pachClient == nil {
		return nil
	}
	return pachClient.Close()
}

// RepoCompletion completes repo parameters of the form <repo>
func RepoCompletion(state CacheState, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	ris, err := c.ListRepo()
	if err != nil {
		return nil, CacheNone
	}
	var result []prompt.Suggest
	for _, ri := range ris {
		result = append(result, prompt.Suggest{
			Text:        ri.Repo.Name,
			Description: fmt.Sprintf("%s (<= %s)", ri.Description, units.BytesSize(float64(ri.SizeBytesUpperBound))),
		})
	}
	return result, samePart(parsePart(state.Text))
}

// BranchCompletion completes branch parameters of the form <repo>@<branch>
func BranchCompletion(state CacheState, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	partialFile := cmdutil.ParsePartialFile(state.Text)
	part := parsePart(state.Text)
	var result []prompt.Suggest
	switch part {
	case repoPart:
		return RepoCompletion(state, maxCompletions)
	case commitOrBranchPart:
		client, err := c.PfsAPIClient.ListBranch(
			c.Ctx(),
			&pfs.ListBranchRequest{
				Repo: partialFile.Commit.Branch.Repo,
			},
		)
		if err != nil {
			return nil, CacheNone
		}
		if err := clientsdk.ForEachBranchInfo(client, func(bi *pfs.BranchInfo) error {
			head := "-"
			if bi.Head != nil {
				head = bi.Head.ID
			}
			result = append(result, prompt.Suggest{
				Text:        fmt.Sprintf("%s@%s:", partialFile.Commit.Branch.Repo, bi.Branch.Name),
				Description: fmt.Sprintf("(%s)", head),
			})
			return nil
		}); err != nil {
			return nil, CacheNone
		}
		if len(result) == 0 {
			// Master should show up even if it doesn't exist yet
			result = append(result, prompt.Suggest{
				Text:        fmt.Sprintf("%s@master", partialFile.Commit.Branch.Repo),
				Description: "(nil)",
			})
		}
	}
	return result, samePart(part)
}

const (
	// filePathCacheLength is how many new characters must be typed in a file
	// path before we go to the server for new results.
	filePathCacheLength = 4
)

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// FileCompletion completes file parameters of the form <repo>@<branch>:/file
func FileCompletion(state CacheState, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	partialFile := cmdutil.ParsePartialFile(state.Text)
	part := parsePart(state.Text)
	var result []prompt.Suggest
	switch part {
	case repoPart:
		return RepoCompletion(state, maxCompletions)
	case commitOrBranchPart:
		return BranchCompletion(state, maxCompletions)
	case filePart:
		if err := c.GlobFile(partialFile.Commit, partialFile.Path+"*", func(fi *pfs.FileInfo) error {
			if maxCompletions > 0 {
				maxCompletions--
			} else {
				return errutil.ErrBreak
			}
			result = append(result, prompt.Suggest{
				Text: fmt.Sprintf("%s@%s:%s", partialFile.Commit.Branch.Repo, partialFile.Commit.ID, fi.File.Path),
			})
			return nil
		}); err != nil {
			return nil, CacheNone
		}
	}
	return result, AndCacheFunc(samePart(part), func(state CacheState) (result bool) {
		_partialFile := cmdutil.ParsePartialFile(state.Text)
		return path.Dir(_partialFile.Path) == path.Dir(partialFile.Path) &&
			abs(len(_partialFile.Path)-len(partialFile.Path)) < filePathCacheLength

	})
}

// FilesystemCompletion completes file parameters from the local filesystem (not from pfs).
func FilesystemCompletion(state CacheState, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	dir := filepath.Dir(state.Text)
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, CacheNone
	}
	var result []prompt.Suggest
	for _, fi := range fis {
		result = append(result, prompt.Suggest{
			Text: filepath.Join(dir, fi.Name()),
		})
	}
	return result, func(state CacheState) bool {
		return filepath.Dir(state.Text) == dir
	}
}

// PipelineCompletion completes pipeline parameters of the form <pipeline>
func PipelineCompletion(_ CacheState, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	client, err := c.PpsAPIClient.ListPipeline(c.Ctx(), &pps.ListPipelineRequest{Details: true})
	if err != nil {
		return nil, CacheNone
	}
	var result []prompt.Suggest
	if err := clientsdk.ForEachPipelineInfo(client, func(pi *pps.PipelineInfo) error {
		result = append(result, prompt.Suggest{
			Text:        pi.Pipeline.Name,
			Description: pi.Details.Description,
		})
		return nil
	}); err != nil {
		return nil, CacheNone
	}
	return result, CacheAll
}

func jobSetDesc(jsi *pps.JobSetInfo) string {
	failure := 0
	var created *types.Timestamp
	var modified *types.Timestamp
	for _, job := range jsi.Jobs {
		if job.State != pps.JobState_JOB_SUCCESS && pps.IsTerminal(job.State) {
			failure++
		}

		if created == nil {
			created = job.Created
			modified = job.Created
		} else {
			if job.Created.Compare(created) < 0 {
				created = job.Created
			}
			if job.Created.Compare(modified) > 0 {
				modified = job.Created
			}
		}
	}
	return fmt.Sprintf("%s, %d subjobs %d failure(s)", pretty.Ago(created), len(jsi.Jobs), failure)
}

// JobCompletion completes job parameters of the form <job-set>
func JobSetCompletion(_ CacheState, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	var result []prompt.Suggest
	listJobSetClient, err := c.PpsAPIClient.ListJobSet(c.Ctx(), &pps.ListJobSetRequest{})
	if err != nil {
		return nil, CacheNone
	}
	if err := clientsdk.ForEachJobSet(listJobSetClient, func(jsi *pps.JobSetInfo) error {
		result = append(result, prompt.Suggest{
			Text:        jsi.JobSet.ID,
			Description: jobSetDesc(jsi),
		})
		return nil
	}); err != nil {
		return nil, CacheNone
	}
	return result, CacheAll
}

func jobDesc(ji *pps.JobInfo) string {
	statusString := ""
	if ji.Finished == nil {
		statusString = fmt.Sprintf("%s for %s", pps_pretty.JobState(ji.State), pretty.Since(ji.Started))
	} else {
		statusString = fmt.Sprintf("%s %s", pps_pretty.JobState(ji.State), pretty.Ago(ji.Finished))
	}
	return fmt.Sprintf("%s: %s - %s", ji.Job.Pipeline.Name, pps_pretty.Progress(ji), statusString)
}

// JobCompletion completes job parameters of the form <pipeline>@<job-ID>
func JobCompletion(state CacheState, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	partialFile := cmdutil.ParsePartialFile(state.Text)
	part := parsePart(state.Text)
	var result []prompt.Suggest
	switch part {
	case repoPart:
		return RepoCompletion(state, maxCompletions)
	case commitOrBranchPart:
		if err := c.ListJobF(partialFile.Commit.Branch.Repo.Name, nil, 0, false, func(info *pps.JobInfo) error {
			result = append(result, prompt.Suggest{
				Text:        fmt.Sprint(info.Job),
				Description: jobDesc(info),
			})
			return nil
		}); err != nil {
			return nil, CacheNone
		}
		return result, samePart(commitOrBranchPart)
	}
	return nil, CacheNone
}
