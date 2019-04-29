# Simple Go Web Service

## Overview

This example shows the stages of creating an application that is deployed with Barrelman. This is the 
progression through the different layers of tooling that all lead to a simplified and consistent 
deployment across environments.

## Process

### First Layer

The first layer is the application. In this case we are dealing with a simple Go web server that
outputs `Hello, APP_NAME. You are running on NODE_NAME`. This could be using any language and 
framework.

In Go, we can run unit tests and build the application in the following way

```
go test ./src/
go build -o go-web-service ./src/
```

### Second Layer

The second layer is the container image. At this time we want to create the Docker container image
and test running it. We are using a simple multi-stage Dockerfile to build our container image and
capture the binary.

```
docker build -t go-web-service:v1.0.0 .
docker run --rm -p 8080:8080 go-web-service:v1.0.0 go-web-service
```

At this point you can browse to `http://localhost:8080/`.

### Third Layer

The third layer is the application and container image defined as Kubernetes deployment package
using Helm. Helm allows templating and overriding configuration of an application while maintining
an immutable version of the application that can be deployed to a Kubernetes cluster.

```
helm install --namespace barrelman-go-web-service --name go-web-service charts/go-web-service/
```

Port forward to the service

```
kubectl -n barrelman-go-web-service port-forward svc/go-web-service 8080:8080
```

At this point you can browse to `http://localhost:8080/`.

### Final Layer

The final layer is the application, container image, and application package defined as a complete
application manifest. This manifest can contain multiple microservices and independent shared
services like databases and distributed caches.

```
barrelman apply manifest.yaml
```

Port forward to the service

```
kubectl -n barrelman-go-web-service port-forward svc/go-web-service 8080:8080
```

At this point you can browse to `http://localhost:8080/`.
