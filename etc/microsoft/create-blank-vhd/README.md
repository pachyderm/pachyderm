# azure-create-vhd

Create and upload an ext4 formatted VHD to Microsoft Azure Storage.

## Background

As the Azure CLI currently does not supporting creating a blank VHD ([azure-cli#655](https://github.com/Azure/azure-cli/issues/655)). Much of the logic was inspired https://github.com/colemickens/azure-tools differing slightly so as not to require Azure login.

## Usage

```
docker build . -t azure-create-vhd
docker run -it azure-create-vhd STORAGE_ACCOUNT_NAME STORAGE_ACCOUNT_KEY CONTAINER_NAME VHD_NAME [VHD_SIZE] [--verbose]

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
  ```
