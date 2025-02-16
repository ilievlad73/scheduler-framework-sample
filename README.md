# scheduler-framework-sample

This repo is a sample for Kubernetes scheduler framework. The `sample` plugin implements `filter` and `prebind` extension points. 
And the custom scheduler name is `scheduler-framework-sample` which defines in `KubeSchedulerConfiguration` object.

## Build

### binary
```shell
$ make local
```

### image
```shell
$ make image
```

## Deploy

```shell
$ kubectl apply -f ./deploy/
```

## Useful commands
```shell
$ kubectl logs $(kubectl get pods -A | grep scheduler-framework | awk -F ' ' '{print $2}') -n kube-system -f
$ kubectl get pods | awk -F ' ' '{print $1}' | tail -n +2 | xargs -n 1 kubectl logs
```