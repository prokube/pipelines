# KFP backend images
This directory contains a script to trigger workflow for building KFP backend images and pushing them to Google Artifact Registry (see .github/workflows/build-backend-images.yaml for detailed info).

## Dev Info
To create KFP backend images (apiserver, controller, cacheserver, persistenceagent, scheduledworkflow, visualization), we first need to build driver and launcher images (set `V2_DRIVER_IMAGE` and `V2_LAUNCHER_IMAGE` environment variables to point to corresponding image). Those are used at least to build the newest possible version of apiserver image, possibly for other images too.

The GitHub actions workflow builds these images automatically and pushes them to GAR. This uses a preconfigured service account for Google Cloud with a role **roles/artifactregistry.writer**. Authentication details (described by GCP_SA_KEY secret on GitHub) can be obtained by running `TODO`.

## Script Usage
Python script does the following:
1. Creates a commit for GitHub. Depending on the command line arguments, either apiserver image, or all backend images are getting built.
2. Pushes the commit to GitHub, triggering images build.

*Optionally*: before using the script:
* Check out to master branch
* Fetch the newest (or specified) available master from pipelines repo
* Check out back to GCP backend images branch, rebase it to new main
* Stage all the necessary files

python3 build_backend_images --images "apiserver" --commit-message "Build:"
python3 build_backend_images --images "all" --commit-message "Build:"