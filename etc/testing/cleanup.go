package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// ClusterInfo holds information about a cluster.
type ClusterInfo struct {
	KopsBucket      string `json:"kops_bucket"`
	PachydermBucket string `json:"pachyderm_bucket"`
	Created         string `json:"created"`
}

// KopsBucket is the s3 bucket used by kops.
const KopsBucket = "pachyderm-travis-state-store-v1"

// MaxClusterTime is the maximimum time a cluster can be up.
const MaxClusterTime = time.Hour * 4

// HandleRequest handles the deletion of old clusters.
func HandleRequest() (string, error) {
	cmd := exec.Command("/bin/bash", "-c", "export PATH=$PATH:/var/task; kops --state=s3://"+KopsBucket+" get clusters | tail -n+2 | awk '{print $1}'")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "Failed to get clusters", err
	}
	names := strings.Split(string(out), "\n")
	names = names[:len(names)-1]
	var deleted string
	s, err := session.NewSession()
	if err != nil {
		return "Failed to create session", err
	}
	svc := s3.New(s)
	for _, name := range names {
		infoObject, err := svc.GetObject(
			&s3.GetObjectInput{
				Bucket: aws.String(KopsBucket),
				Key:    aws.String(name + "-info.json"),
			})
		if err != nil {
			return "Failed to get info file", err
		}
		var info ClusterInfo
		if err := json.NewDecoder(infoObject.Body).Decode(&info); err != nil {
			return "Failed to decode info file", err
		}
		createTime, err := time.Parse(time.UnixDate, info.Created)
		if err != nil {
			return "Failed to parse create time", err
		}
		// Cluster has been up for too long
		if createTime.Add(MaxClusterTime).Before(time.Now()) {
			deleted += name + ", "
			cmd := exec.Command("/bin/bash", "-c", "export PATH=$PATH:/var/task; kops --state=s3://"+KopsBucket+" delete cluster --name="+name+" --yes")
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return "Failed to delete cluster", err
			}
			_, err := svc.DeleteBucket(
				&s3.DeleteBucketInput{
					Bucket: aws.String(info.PachydermBucket),
				})
			if err != nil {
				return "Failed to delete pachyderm bucket", err
			}
			_, err = svc.DeleteObject(
				&s3.DeleteObjectInput{
					Bucket: aws.String(KopsBucket),
					Key:    aws.String(name + "-info.json"),
				})
			if err != nil {
				return "Failed to delete info file", err
			}
		}
	}
	if len(deleted) <= 0 {
		return "No clusters deleted", nil
	}
	deleted = "Clusters deleted: " + deleted
	tmp := []rune(deleted)
	return string(tmp[:len(tmp)-2]), nil
}

func main() {
	lambda.Start(HandleRequest)
}
