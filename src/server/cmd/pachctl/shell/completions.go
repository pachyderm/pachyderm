package shell

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pretty"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pps_pretty "github.com/pachyderm/pachyderm/v2/src/server/pps/pretty"
	"google.golang.org/protobuf/types/known/timestamppb"

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
	return func(_, text string) bool {
		return parsePart(text) == p
	}
}

var (
	pachClient     *client.APIClient
	pachClientOnce sync.Once
)

func getPachClient() *client.APIClient {
	pachClientOnce.Do(func() {
		c, err := client.NewOnUserMachine(pctx.TODO(), "user-completion")
		if err != nil {
			Fatal(err)
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
func RepoCompletion(_, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
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
	return result, samePart(parsePart(text))
}

// ProjectBranchCompletion completes branch parameters of the form
// <repo>@<branch> in a project-aware fashion.
func ProjectBranchCompletion(project, flag, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	partialFile := cmdutil.ParsePartialFile(project, text)
	part := parsePart(text)
	var result []prompt.Suggest
	switch part {
	case repoPart:
		return RepoCompletion(flag, text, maxCompletions)
	case commitOrBranchPart:
		client, err := c.PfsAPIClient.ListBranch(
			c.Ctx(),
			&pfs.ListBranchRequest{
				Repo: partialFile.Commit.Repo,
			},
		)
		if err != nil {
			return nil, CacheNone
		}
		if err := grpcutil.ForEach[*pfs.BranchInfo](client, func(bi *pfs.BranchInfo) error {
			head := "-"
			if bi.Head != nil {
				head = bi.Head.Id
			}
			result = append(result, prompt.Suggest{
				Text:        fmt.Sprintf("%s@%s:", partialFile.Commit.Repo, bi.Branch.Name),
				Description: fmt.Sprintf("(%s)", head),
			})
			return nil
		}); err != nil {
			return nil, CacheNone
		}
		if len(result) == 0 {
			// Master should show up even if it doesn't exist yet
			result = append(result, prompt.Suggest{
				Text:        fmt.Sprintf("%s@master", partialFile.Commit.Repo),
				Description: "(nil)",
			})
		}
	}
	return result, samePart(part)
}

// BranchCompletion completes branch parameters of the form <repo>@<branch>
func BranchCompletion(flag, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	return ProjectBranchCompletion(pfs.DefaultProjectName, flag, text, maxCompletions)
}

func ProjectCompletion(flag, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	pis, err := c.ListProject()
	if err != nil {
		return nil, CacheNone
	}
	var result []prompt.Suggest
	for _, pi := range pis {
		result = append(result, prompt.Suggest{
			Text:        pi.Project.Name,
			Description: pi.Description,
		})
	}
	return result, CacheNone
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

// ProjectFileCompletion completes file parameters of the form
// <repo>@<branch>:/file in a project-aware manner.
func ProjectFileCompletion(project, flag, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	partialFile := cmdutil.ParsePartialFile(project, text)
	part := parsePart(text)
	var result []prompt.Suggest
	switch part {
	case repoPart:
		return RepoCompletion(flag, text, maxCompletions)
	case commitOrBranchPart:
		return BranchCompletion(flag, text, maxCompletions)
	case filePart:
		if err := c.GlobFile(partialFile.Commit, partialFile.Path+"*", func(fi *pfs.FileInfo) error {
			if maxCompletions > 0 {
				maxCompletions--
			} else {
				return errutil.ErrBreak
			}
			result = append(result, prompt.Suggest{
				Text: fmt.Sprintf("%s@%s:%s", partialFile.Commit.Repo, partialFile.Commit.Id, fi.File.Path),
			})
			return nil
		}); err != nil {
			return nil, CacheNone
		}
	}
	return result, AndCacheFunc(samePart(part), func(_, text string) (result bool) {
		_partialFile := cmdutil.ParsePartialFile(project, text)
		return path.Dir(_partialFile.Path) == path.Dir(partialFile.Path) &&
			abs(len(_partialFile.Path)-len(partialFile.Path)) < filePathCacheLength

	})
}

// FileCompletion completes file parameters of the form <repo>@<branch>:/file
func FileCompletion(flag, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	return ProjectFileCompletion(pfs.DefaultProjectName, flag, text, maxCompletions)
}

// FilesystemCompletion completes file parameters from the local filesystem (not from pfs).
func FilesystemCompletion(_, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	dir := filepath.Dir(text)
	fis, err := os.ReadDir(dir)
	if err != nil {
		return nil, CacheNone
	}
	var result []prompt.Suggest
	for _, fi := range fis {
		result = append(result, prompt.Suggest{
			Text: filepath.Join(dir, fi.Name()),
		})
	}
	return result, func(_, text string) bool {
		return filepath.Dir(text) == dir
	}
}

// PipelineCompletion completes pipeline parameters of the form <pipeline>
func PipelineCompletion(_, _ string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	client, err := c.PpsAPIClient.ListPipeline(c.Ctx(), &pps.ListPipelineRequest{Details: true})
	if err != nil {
		return nil, CacheNone
	}
	var result []prompt.Suggest
	if err := grpcutil.ForEach[*pps.PipelineInfo](client, func(pi *pps.PipelineInfo) error {
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
	var created *timestamppb.Timestamp
	var modified *timestamppb.Timestamp
	for _, job := range jsi.Jobs {
		if job.State != pps.JobState_JOB_SUCCESS && pps.IsTerminal(job.State) {
			failure++
		}

		if created == nil {
			created = job.Created
			modified = job.Created
		} else {
			if job.Created.AsTime().Before(created.AsTime()) {
				created = job.Created
			}
			if job.Created.AsTime().After(modified.AsTime()) {
				modified = job.Created
			}
		}
	}
	return fmt.Sprintf("%s, %d subjobs %d failure(s)", pretty.Ago(created), len(jsi.Jobs), failure)
}

// JobCompletion completes job parameters of the form <job-set>
func JobSetCompletion(_, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	var result []prompt.Suggest
	listJobSetClient, err := c.PpsAPIClient.ListJobSet(c.Ctx(), &pps.ListJobSetRequest{})
	if err != nil {
		return nil, CacheNone
	}
	if err := grpcutil.ForEach[*pps.JobSetInfo](listJobSetClient, func(jsi *pps.JobSetInfo) error {
		result = append(result, prompt.Suggest{
			Text:        jsi.JobSet.Id,
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

// JobCompletion completes job parameters of the form <job>
func JobCompletion(_, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	var result []prompt.Suggest
	if err := c.ListJobF("", "", nil, 0, false, func(ji *pps.JobInfo) error {
		if maxCompletions > 0 {
			maxCompletions--
		} else {
			return errutil.ErrBreak
		}
		result = append(result, prompt.Suggest{
			Text:        ji.Job.Id,
			Description: jobDesc(ji),
		})
		return nil
	}); err != nil {
		return nil, CacheNone
	}
	return result, CacheAll
}
