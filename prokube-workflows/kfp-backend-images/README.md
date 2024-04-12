# KFP backend images
This directory contains a script to trigger GitHub CI workflow for building KFP backend images. The images then get pushed to Google Artifact Registry (GAR) (see .github/workflows/build-backend-images.yaml for detailed info).

## Usage
### Build Images and Push to GAR
Stage the desired changes using git, then run one of the following:
```make
make build_driver_launcher  # build and push driver and launcher images
make build_apiserver  # build and push apiserver image (launcher and driver get updated automatically)
make build_all  # build and push all KFP backend images
```
These commands push the code to GitHub and trigger image build using GitHub workflows. The images are then pushed to GAR. GAR path is specified in **GCP_ARTIFACT_REGISTRY_PATH** secret, gcloud auth credentials are in **GCP_SA_KEY** secret.
  
The code gets pushed to the current upstream branch. If you want to add custom commit message, see Python script details below. 
  
### Deploy to Cluster
First, you need to add a secret to kubeflow namespace, containing GAR credentials (use service account that is authenticated to pull images from GAR).
```sh
gcloud iam service-accounts keys create key_pull.json --iam-account={SERVICE_ACCOUNT}
kubectl -n=kubeflow create secret docker-registry regcred-gar \
  --docker-server europe-west3-docker.pkg.dev \
  --docker-username _json_key \
  --docker-password="$(cat key_pull.json)"
```

There are 2 ways to deploy the obtained backend images in a cluster:

1. Local deployment patch using Makefile. Specify `REGISTRY_PATH`(e.g. europe-west1-docker.pkg.dev/{GCP_PROJECT_NAME}/prokube/) environment variable and then run `make apiserver` to patch apiserver image. Run `make all` to patch all backend images (this is a dev way: images will be reset to default on cluster restart).
2. (Semi-)automated GitLab CI deployment. Implemented only for apiserver image in **prokube-ai/paas** repository in branch **feature/custom_pipelines_apiserver_image** [here](https://github.com/prokube-ai/paas/blob/feature/custom_pipelines_api_server_image/kubeflow/project-overlays/pipeline/pipeline/pipeline-api-server.yaml). Replace `GITLAB_REGISTRY:MLPIPELINESERVER_IMAGE_TAG` with the actual path to the image in the registry, push these changes to GitLab CI and sync kubeflow using ArgoCD.

## Dev Info
![Untitled-2024-04-05-0940-4](https://github.com/prokube-ai/pipelines/assets/116455436/2c36c3f8-d9f8-4bb0-870c-9e46c5e9159a)  
To create KFP backend images (apiserver, controller, cacheserver, persistenceagent, scheduledworkflow, visualization), we first need to build driver and launcher images (set `V2_DRIVER_IMAGE` and `V2_LAUNCHER_IMAGE` environment variables to point to corresponding images). Those are used at least to build the newest version of apiserver image, possibly for other images too.

The GitHub actions workflow builds these images automatically and pushes them to GAR. This uses a preconfigured service account for Google Cloud with a role **roles/artifactregistry.writer**. Authentication details (described by GCP_SA_KEY secret on GitHub) can be obtained by running `gcloud iam service-accounts keys create key.json --iam-account=github@{GCP_PROJECT_NAME}.iam.gserviceaccount.com` (replace {GCP_PROJECT_NAME} with our project name).

GitHub workflows uses keyword "[apiserver]" anywhere in commit message to trigger apiserver build, uses "[backend-all]" keyword to build all backend images, same logic for launcher and driver images.

### Script Details
Python script does the following:
1. Creates a commit for GitHub. Depending on the command line arguments, either launcher&driver, apiserver, or all backend images are getting built.
2. Pushes the commit to GitHub, triggering images build.

*Optionally*: sync the fork before using the script.

Example usage:
```py
python3 build_backend_images --images "apiserver" --commit-message "Build:"
python3 build_backend_images --images "all" --commit-message "Build:"
```
