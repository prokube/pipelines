# KFP backend images
This directory contains a script to trigger GitHub CI workflow for building KFP backend images. The images then get pushed to Google Artifact Registry (GAR) (see .github/workflows/build-backend-images.yaml for detailed info).

## Usage
Stage the desired changes using git, then run `make build_apiserver` to build and push apiserver image to GAR using GitHub CI. Run `make build_all` to build and push all KFP backend images. Both commands commit to **GCP-artifact-registry** branch of this repository. If you want to add custom commit message, see Python script details below. Google artifact registry path is specified in **GCP_ARTIFACT_REGISTRY_PATH** secret, gcloud auth credentials are in **GCP_SA_KEY** secret.
  
There are 2 ways to deploy the obtained backend images in cluster:

1. (Semi-)automated GitLab CI deployment of custom apiserver image. Implemented for apiserver image in **prokube-ai/paas** repository in branch **feature/custom_pipelines_apiserver_image**. Replace `GITLAB_REGISTRY:MLPIPELINESERVER_IMAGE_TAG` in **paas/kubeflow/project-overlays/pipeline/pipeline-api-server.yaml** with the actual path to the image registry.
2. Local deployment patch using Makefile. Specify `REGISTRY_PATH`(e.g. europe-west1-docker.pkg.dev/{GCP_PROJECT_NAME}/prokube/) environment variable and then run `make apiserver` to patch apiserver image. Run `make all` to patch all backend images (this is a dev way: images will be reset to default on cluster restart).

## Dev Info
To create KFP backend images (apiserver, controller, cacheserver, persistenceagent, scheduledworkflow, visualization), we first need to build driver and launcher images (set `V2_DRIVER_IMAGE` and `V2_LAUNCHER_IMAGE` environment variables to point to corresponding image). Those are used at least to build the newest possible version of apiserver image, possibly for other images too.

The GitHub actions workflow builds these images automatically and pushes them to GAR. This uses a preconfigured service account for Google Cloud with a role **roles/artifactregistry.writer**. Authentication details (described by GCP_SA_KEY secret on GitHub) can be obtained by running `gcloud iam service-accounts keys create key.json --iam-account=github@{GCP_PROJECT_NAME}.iam.gserviceaccount.com` (replace {GCP_PROJECT_NAME} with our project name).

GitHub workflows uses keyword "[apiserver]" anywhere in commit message to trigger apiserver build, uses "[backend-all]" keyword to build all backend images.

### Script Details
Python script does the following:
1. Creates a commit for GitHub. Depending on the command line arguments, either apiserver image, or all backend images are getting built.
2. Pushes the commit to GitHub, triggering images build.

*Optionally*: before using the script:
* Check out to master branch
* Fetch the newest (or specified) available master from pipelines repo
* Check out back to GCP backend images branch, rebase it to new main
* Stage all the necessary files

Example usage:
```py
python3 build_backend_images --images "apiserver" --commit-message "Build:"
python3 build_backend_images --images "all" --commit-message "Build:"
```