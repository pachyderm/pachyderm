package shell

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
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
			Description: fmt.Sprintf("%s (%s)", ri.Description, units.BytesSize(float64(ri.SizeBytes))),
		})
	}
	return result, samePart(parsePart(text))
}

// BranchCompletion completes branch parameters of the form <repo>@<branch>
func BranchCompletion(flag, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	partialFile := cmdutil.ParsePartialFile(text)
	part := parsePart(text)
	var result []prompt.Suggest
	switch part {
	case repoPart:
		return RepoCompletion(flag, text, maxCompletions)
	case commitOrBranchPart:
		bis, err := c.PfsAPIClient.ListBranch(
			c.Ctx(),
			&pfs.ListBranchRequest{
				Repo: partialFile.Commit.Branch.Repo,
			},
		)
		if err != nil {
			return nil, CacheNone
		}
		for _, bi := range bis.BranchInfo {
			head := "-"
			if bi.Head != nil {
				head = bi.Head.ID
			}
			result = append(result, prompt.Suggest{
				Text:        fmt.Sprintf("%s@%s:", pfsdb.RepoKey(partialFile.Commit.Branch.Repo), bi.Branch.Name),
				Description: fmt.Sprintf("(%s)", head),
			})
		}
		if len(result) == 0 {
			// Master should show up even if it doesn't exist yet
			result = append(result, prompt.Suggest{
				Text:        fmt.Sprintf("%s@master", pfsdb.RepoKey(partialFile.Commit.Branch.Repo)),
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
func FileCompletion(flag, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	partialFile := cmdutil.ParsePartialFile(text)
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
				Text: fmt.Sprintf("%s@%s:%s", pfsdb.RepoKey(partialFile.Commit.Branch.Repo), partialFile.Commit.ID, fi.File.Path),
			})
			return nil
		}); err != nil {
			return nil, CacheNone
		}
	}
	return result, AndCacheFunc(samePart(part), func(_, text string) (result bool) {
		_partialFile := cmdutil.ParsePartialFile(text)
		return path.Dir(_partialFile.Path) == path.Dir(partialFile.Path) &&
			abs(len(_partialFile.Path)-len(partialFile.Path)) < filePathCacheLength

	})
}

// FilesystemCompletion completes file parameters from the local filesystem (not from pfs).
func FilesystemCompletion(_, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	dir := filepath.Dir(text)
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
	return result, func(_, text string) bool {
		return filepath.Dir(text) == dir
	}
}

// PipelineCompletion completes pipeline parameters of the form <pipeline>
func PipelineCompletion(_, _ string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	resp, err := c.PpsAPIClient.ListPipeline(c.Ctx(), &pps.ListPipelineRequest{AllowIncomplete: true})
	if err != nil {
		return nil, CacheNone
	}
	var result []prompt.Suggest
	for _, pi := range resp.PipelineInfo {
		result = append(result, prompt.Suggest{
			Text:        pi.Pipeline.Name,
			Description: pi.Description,
		})
	}
	return result, CacheAll
}

func jobDesc(pji *pps.PipelineJobInfo) string {
	statusString := ""
	if pji.Finished == nil {
		statusString = fmt.Sprintf("%s for %s", pps_pretty.JobState(pji.State), pretty.Since(pji.Started))
	} else {
		statusString = fmt.Sprintf("%s %s", pps_pretty.JobState(pji.State), pretty.Ago(pji.Finished))
	}
	return fmt.Sprintf("%s: %s - %s", pji.Pipeline.Name, pps_pretty.Progress(pji), statusString)
}

// JobCompletion completes job parameters of the form <job>
func JobCompletion(_, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	var result []prompt.Suggest
	if err := c.ListPipelineJobF("", nil, nil, 0, false, func(pji *pps.PipelineJobInfo) error {
		if maxCompletions > 0 {
			maxCompletions--
		} else {
			return errutil.ErrBreak
		}
		result = append(result, prompt.Suggest{
			Text:        pji.PipelineJob.ID,
			Description: jobDesc(pji),
		})
		return nil
	}); err != nil {
		return nil, CacheNone
	}
	return result, CacheAll
}
