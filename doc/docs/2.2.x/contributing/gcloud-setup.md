# Google Cloud Cluster Setup 

You can develop Pachyderm against a cluster deployed in Google Cloud using this guide. 


## Before You Start 

- [Install Pachyderm locally](/getting-started/local-installation/#local-installation). 
- Download Google's [Cloud SDK](https://cloud.google.com/sdk/).
- Get your GCP owner/admin to set up a project for you (e.g., `YOURNAME-dev`).
    - This requires permissions from **Project** > **Settings** > **Permissions**. 
  
---

## 1. Set Up Google Cloud Platform (GCP)


1. Go to [console.cloud.google.com](https://console.cloud.google.com/).
2. Sign in via Gmail or G Suite.  
3. Navigate to APIs and enable the Google Compute API.
4. Open a terminal locally and each of the following commands individually:
      ```
      # Log in:
      gcloud auth login 
      gcloud init

      # Connect to your project (requires permissions mentioned in previous section):
      gcloud config set project YOURNAME-dev
      gcloud compute instances list 

      # Create an instance via the bash helper: 
      create_docker_machine

      # Attach to Docker daemon: 
      eval "$(docker-machine env dev)"
      ```


## 2. Set up and Launch Kubectl

1. Open a terminal and run the following commands:
      ```
      # Update kubectl 
      gcloud components update kubectl

      # Set up port forwarding
      portforwarding

      # View client version
      kubectl version

      # Deploy Kubernetes 
      make launch-kube 

      # View client & server version
      kubectl version 

      # Verify processes
      docker ps 
      ```


## 3. Deploy Pachyderm Cluster 

1. Open a terminal and run the following command:
      ```
      make launch
      ```
2. Start developing! 
