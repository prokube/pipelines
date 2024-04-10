# Trigger image build --------------
.PHONY: build_apiserver build_all

build_apiserver:
	python3 build_backend_images.py --images=apiserver

build_all:
	python3 build_backend_images.py --images=all


# Deployment patches --------------
.PHONY: check-env patch_apiserver patch_cacheserver patch_persistenceagent patch_scheduledworkflow patch_viewercontroller patch_visualization patch_all

check-env:
ifndef REGISTRY_PATH
	$(error REGISTRY_PATH is not set)
endif

patch_apiserver: check-env
	kubectl patch deployment -n kubeflow ml-pipeline --patch "$$(envsubst < kfp-deployment-patches/apiserver.yaml)"

patch_cacheserver: check-env
	kubectl patch deployment -n kubeflow cache-server --patch "$$(envsubst < kfp-deployment-patches/cacheserver.yaml)"

patch_persistenceagent: check-env
	kubectl patch deployment -n kubeflow ml-pipeline-persistenceagent --patch "$$(envsubst < kfp-deployment-patches/persistenceagent.yaml)"

patch_scheduledworkflow: check-env
	kubectl patch deployment -n kubeflow ml-pipeline-scheduledworkflow --patch "$$(envsubst < kfp-deployment-patches/scheduledworkflow.yaml)"

patch_viewercontroller: check-env
	kubectl patch deployment -n kubeflow ml-pipeline-viewer-crd --patch "$$(envsubst < kfp-deployment-patches/viewercontroller.yaml)"

patch_visualization: check-env
	kubectl patch deployment -n kubeflow ml-pipeline-visualization-server --patch "$$(envsubst < kfp-deployment-patches/visualization.yaml)"

patch_all: patch_apiserver patch_cacheserver patch_persistenceagent patch_scheduledworkflow patch_viewercontroller patch_visualization
