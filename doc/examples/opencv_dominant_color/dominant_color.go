package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/OwnLocal/PipeDream/aws"
	"github.com/prometheus/common/log"
)

// Text file that has all the image urls
var images_file = flag.String("images_file", "", "path to the s3 urls")
var aws_bucket_name = flag.String("aws_bucket_name", "", "s3 bucket")

func usageAndExit(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	fmt.Println("Flags:")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Parse()

	file, err := os.Open(string(*images_file))
	if err != nil {
		log.Error(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Read the source text file line by line
	for scanner.Scan() {
		fileName := scanner.Text()

		// Setup the s3 file object based on file name and bucket flag
		file := aws.S3Type{Bucket: *aws_bucket_name, FileName: fileName, ContentType: "image"}

		// Download the input file
		err := aws.Download(file)
		if err != nil {
			log.Error(err)
		}

		// Call the c++ algorithm on the file
		cmd := fmt.Sprintf("/home/dominantColor %s ", fileName)
		output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
		if err != nil {
			log.Error(string(output))
		}

		// Print the results of the dominant colors
		fmt.Println(fileName, "|", string(output))
	}

	if err := scanner.Err(); err != nil {
		log.Error(err)
	}
}
