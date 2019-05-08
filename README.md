# Devoxx UK 2019 - Knative demo

## Setup

1. Create GKE cluster fronted by DNS `*.knative.currie.cloud`.
1. Update [docker-secret.yaml](setup/docker-secret.yaml) with Docker Hub credentials.
1. `kubectl apply -f setup` to create secret with Docker Hub credentials associated with `build-bot` service account.
1. Install `kaniko` build template `kubectl apply -f https://raw.githubusercontent.com/knative/build-templates/master/kaniko/kaniko.yaml`
1. Follow the setup instructions in the [GCP Pub/Sub Knative sample](https://knative.dev/docs/eventing/samples/gcp-pubsub-source/).

## Serving

### Deploy Knative service

1. [helloworld.go](helloworld.go) is a simple Go web server. Note the use of `PORT` passed in by Knative.
1. It has been built with [Dockerfile](Dockerfile) and pushed to Docker Hub at `dcurrie/helloworld-go:latest`.
1. [service.yaml](service.yaml) defines a Knative service which specifies the Docker image in its configuration along with an environment variable that is output in the response.
1. `kubectl apply -f service.yaml`
1. `curl helloworld-go.default.knative.currie.cloud` should return `v1: Hello World!`.

### Deploy a second revision of the service

1. Change `v1` to `v2` in `service.yaml` and re-apply.
1. ```curl helloworld-go.default.knative.currie.cloud``` should now return `v2: Hello World!`. A new revision of the service has been created and, because we are using `runLatest` in the service definition, requests are automatically routed to the new revision once it has become available.

### Auto-scaling

Unlike, say, a Function-as-a-Service platform, container instances are expected to handle multiple requests and, by default, handle concurrent requests.


1. By default, the number of pods for a revision scales down to zero. This can be great as it means old revisions to which traffic is no longer being routed don't cost anything. It may not be desirable though if traffic is the container for a service takes time to become ready. Add the annotation `autoscaling.knative.dev/minScale: "2"` and re-apply the `service.yaml`.
1. `kubectl get pod` should now show two pods for the latest revision.
1. By default, auto-scaling is triggered based on a target maximum concurrency of 100. We'll lower that target to make triggering a scaling decision easier. Remove the `minScale` annotation and add the annotation `autoscaling.knative.dev/target: "1"` and re-apply the `service.yaml`.
1. We'll use [hey](https://github.com/rakyll/hey) to drive some load. Run `hey -z 10s -c 100 http://helloworld-go.default.knative.currie.cloud && kubectl get pods`. Although the default averaging window is 60 seconds, when the concurrency breaches double the target then the auto-scaler enters panic mode and starts scaling up the number of instances so you should see additional pods.

### Manual blue-green deployment

1. Execute `kubectl get revision` and note the name of the revision for the latest generation.
1. In `service.yaml`, change `runLatest` to `release` and `v2` to `v3`. Add a `revisions` stanza under `release` which lists the latest revision noted in the previous step.
1. `curl helloworld-go.default.knative.currie.cloud` should still return `v2: Hello World!`.
1. Execute `kubectl get revision` and note the name of the revision for the third generation.
1. Add the new revision into `service.yaml` under the existing one. Then add a `rolloutPercent: 0` as a sibling of the `revisions` element. Apply the updated YAML.
1.`curl helloworld-go.default.knative.currie.cloud` is still returning `v2: Hello World!`.
1. Although no default traffic is being routed to the new service revision, it is now tagged `candidate` and is available via `curl candidate.helloworld-go.default.knative.currie.cloud`.
1. Increase the rollout percentage to `50` and re-apply.
1. Run `curl helloworld-go.default.knative.currie.cloud` repeatedly and note that the workload is now balanced across the two revisions.
1. Remove the `rolloutPercent` and the old revision and re-apply.
1. Run `curl helloworld-go.default.knative.currie.cloud` repeatedly and you should now only see the latest version.

### Cleanup

1. `kubectl delete -f service.yaml`

## Build

1. Take a look at [service-with-build.yaml](service-with-build.yaml). It adds a `build` configuration that points to the GitHub repo containing this application. It also specifies a service account configured with Docker Hub credentials and a Kaniko build template.
1. Retrieve the build template with `kubectl get buildtemplate kaniko -o yaml`. Note how it takes the image name as a parameter. This template only contains a single step which specifies the kaniko `executor` image that will build an image and push it to a registry.
1. Apply the service configuration with `kubectl apply -f service-with-build.yaml`.
1. Watch the pods with `kubectl get pods -w`.
1. Note that the build pod contains three init containers. Two are always injected: one to set up credentials (e.g. mounting in the Docker and/or Git credentials) and the second to check out the source from Git. These are followed by a container for each step in the build. They are run as init containers so that execute sequentially.
1. Once the build has complete, the service will be deployed.
1. Run `curl helloworld-go.default.knative.currie.cloud`

## Eventing

1. Create eventing source with `kubectl apply -f gcp-pubsub-source.yaml`.
1. Create the trigger to subscribe with `kubectl apply -f trigger.yaml`.
1. Publish a message `gcloud pubsub topics publish devoxx --message "Devoxx"`
1. Look at the logs for the application and note that it has been driven by the message.
1. Look at [helloworld.go](helloworld.go) again to see how it handles the event.

### Cleanup

1. `kubectl delete -f service-with-build.yaml`
1. `kubectl delete -f trigger.yaml`
1. `kubectl delete -f gcp-pubsub-source.yaml`