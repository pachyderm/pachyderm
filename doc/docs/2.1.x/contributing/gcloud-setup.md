# Google Cloud Cluster Setup 

You can develop Pachyderm against a cluster deployed in Google Cloud using this guide.


## Before You Start 

- [Install Pachyderm locally](../../../getting-started/local-installation) 
- Download Google's [Cloud SDK](https://cloud.google.com/sdk/).
- Get your GCP owner/admin to set up a project for you (e.g., `YOURNAME-dev`).
    - This requires permissions from **Project** > **Settings** > **Permissions**. 
---

## 1. Set Up Google Cloud Platform


1. Go to [console.cloud.google.com](https://console.cloud.google.com/).
2. Sign in via Gmail or G Suite.  
3. Navigate to **APIs & Services** > **Enable APIs & Services** > **Library**.
4. Find the **Compute Engine API**.
5. Select **Enable**.
6. Open a terminal locally and each of the following commands individually:
      1. **Log in**:
      ```
      gcloud auth login 
      gcloud init
      ```
      2. **Connect to your project**:
      ```
      gcloud config set project YOURNAME-dev
      gcloud compute instances list 
      ```
      3. **Create an instance via the bash helper**: 
      ```
      create_docker_machine
      ```
      4. **Attach to Docker daemon**: 
      ```
      eval "$(docker-machine env dev)"
      ```

## 2. Set up and Launch Kubectl

Open a terminal and run the following commands:

1. **Update kubectl**: `gcloud components update kubectl`.
2. **Set up port forwarding**: `portforwarding`.
3. **View client version**: `kubectl version`.
4. **Deploy Kubernetes**: `make launch-kube`.
5. **View client & server version**:  `kubectl version`.
6. **Verify processes**: `docker ps`.

## 3. Deploy Pachyderm Cluster 

Open a terminal and run the following command:
      ```
      make launch
      ```
