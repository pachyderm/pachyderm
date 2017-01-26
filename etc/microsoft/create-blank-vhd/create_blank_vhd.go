package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/docopt/docopt-go"
)

var isVerbose = false

func main() {
	usage := `

Usage:
  create_blank_vhd STORAGE_ACCOUNT_NAME STORAGE_ACCOUNT_KEY CONTAINER_NAME VHD_NAME [VHD_SIZE] [--verbose]

Arguments:
  STORAGE_ACCOUNT_NAME  Azure storage account name
  STORAGE_ACCOUNT_KEY   Azure storage account key
  CONTAINER_NAME        Name of blob container to store VHD
  VHD_NAME              Name of VHD to create. Must end in .vhd extension

Options:
  -h --help          	Show this help message and exit
  --vhd_size N          Optional parameter denoting size in bytes of VHD (Default: 10G).
                        Suffixes "k" or "K" (kilobyte, 1024) "M" (megabyte, 1024k) 
                        "G" (gigabyte, 1024M) and T (terabyte, 1024G) are supported.
  --verbose             Output logs (Default: false)

`

	args, _ := docopt.Parse(usage, nil, true, "", false)

	strgAccountName := args["STORAGE_ACCOUNT_NAME"].(string)
	strgAccountKey := args["STORAGE_ACCOUNT_KEY"].(string)
	containerName := args["CONTAINER_NAME"].(string)
	vhdName := args["VHD_NAME"].(string)
	vhdSize := args["VHD_SIZE"]
	isVerbose = args["--verbose"].(bool)

	if vhdSize == nil {
		vhdSize = "10G"
	}

	var cmdName string
	var cmdArgs []string
	var err error

	// Create raw disk
	cmdName = "qemu-img"
	cmdArgs = []string{"create", "-f", "raw", "image.raw", vhdSize.(string)}
	if _, err = execCommand("Create raw disk", cmdName, cmdArgs); err != nil {
		os.Exit(1)
	}

	// Format disk as ext4
	cmdName = "mkfs.ext4"
	cmdArgs = []string{"./image.raw"}
	if _, err = execCommand("Format disk", cmdName, cmdArgs); err != nil {
		os.Exit(1)
	}

	// Convert to vhd
	cmdName = "qemu-img"
	cmdArgs = []string{"convert", "-f", "raw", "-o", "subformat=fixed,force_size", "-O", "vpc", "image.raw", "image.vhd"}
	if _, err = execCommand("Convert to vhd", cmdName, cmdArgs); err != nil {
		os.Exit(1)
	}

	// Upload
	cmdName = "azure-vhd-utils-for-go"
	cmdArgs = []string{
		"upload",
		"--localvhdpath=image.vhd",
		"--stgaccountname=" + strgAccountName,
		"--stgaccountkey=" + strgAccountKey,
		"--containername=" + containerName,
		"--blobname=" + vhdName,
	}
	if _, err = execCommand("Upload", cmdName, cmdArgs); err != nil {
		os.Exit(1)
	}

	// Get Blob Url
	client, err := storage.NewBasicClient(strgAccountName, strgAccountKey)
	if err != nil {
		print(err)
		os.Exit(1)
	}

	url := client.GetBlobService().GetBlobURL(containerName, vhdName)

	fmt.Print(url)

	return
}

func execCommand(friendlyName string, cmdName string, cmdArgs []string) (string, error) {
	print("--- " + friendlyName + " ---")
	stdOut, err := exec.Command(cmdName, cmdArgs...).Output()
	if err != nil {
		print(err)
		return "", err
	}
	print(string(stdOut))
	return string(stdOut), nil
}

func print(message interface{}) {
	if isVerbose {
		fmt.Println(message)
	}
}
