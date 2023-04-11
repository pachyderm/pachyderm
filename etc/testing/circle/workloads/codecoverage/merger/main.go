package main

import (
	"log"
)

func main() {
	logger := log.Default()
	logger.Println("Running merger code")
	// inputFolder := os.Args[1]
	// outputFolder := os.Args[2]

	// formatArgs := []string{"tool", "covdata", "txtfmt", "-i"}

	// files := []string{}
	// filepath.WalkDir(inputFolder, func(path string, d fs.DirEntry, err error) error {
	// 	if !d.IsDir() {
	// 		files = append(files, path)
	// 	}
	// 	return nil
	// })
	// formatArgs = append(formatArgs, strings.Join(files, ","), "-o", outputFolder)
	// // DNJ TODO - do we need explicit merge call
	// cmd := exec.Command("go", formatArgs...)
	// fmt.Println("Running command: ", cmd.String())
	// if err := cmd.Run(); err != nil {
	// 	log.Fatal(err)
	// }
	// check for consumed file right before save
	// put cov file in pfsout - deferred pipeline pushes to codecov
}
