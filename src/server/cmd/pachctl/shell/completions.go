package shell

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/pretty"
	pps_pretty "github.com/pachyderm/pachyderm/src/server/pps/pretty"

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
		log.Fatal(err)
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
		bis, err := c.ListBranch(partialFile.Commit.Repo.Name)
		if err != nil {
			log.Fatal(err)
		}
		for _, bi := range bis {
			head := "-"
			if bi.Head != nil {
				head = bi.Head.ID
			}
			result = append(result, prompt.Suggest{
				Text:        fmt.Sprintf("%s@%s:", partialFile.Commit.Repo.Name, bi.Branch.Name),
				Description: fmt.Sprintf("(%s)", head),
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
		if err := c.GlobFileF(partialFile.Commit.Repo.Name, partialFile.Commit.ID, partialFile.Path+"*", func(fi *pfs.FileInfo) error {
			if maxCompletions > 0 {
				maxCompletions--
			} else {
				return errutil.ErrBreak
			}
			result = append(result, prompt.Suggest{
				Text: fmt.Sprintf("%s@%s:%s", partialFile.Commit.Repo.Name, partialFile.Commit.ID, fi.File.Path),
			})
			return nil
		}); err != nil {
			log.Fatal(err)
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
		log.Fatal(err)
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
	pis, err := c.ListPipeline()
	if err != nil {
		log.Fatal(err)
	}
	var result []prompt.Suggest
	for _, pi := range pis {
		result = append(result, prompt.Suggest{
			Text:        pi.Pipeline.Name,
			Description: pi.Description,
		})
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
	return fmt.Sprintf("%s: %s - %s", ji.Pipeline.Name, pps_pretty.Progress(ji), statusString)
}

// JobCompletion completes job parameters of the form <job>
func JobCompletion(_, text string, maxCompletions int64) ([]prompt.Suggest, CacheFunc) {
	c := getPachClient()
	var result []prompt.Suggest
	if err := c.ListJobF("", nil, nil, 0, false, func(ji *pps.JobInfo) error {
		if maxCompletions > 0 {
			maxCompletions--
		} else {
			return errutil.ErrBreak
		}
		result = append(result, prompt.Suggest{
			Text:        ji.Job.ID,
			Description: jobDesc(ji),
		})
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	return result, CacheAll
}
