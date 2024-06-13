# Kubernetes Deployment

The Caddy web server, with the `adobe_usage_tracker` plugin, can be deployed via Kubernetes. The sample files for two common deployment strategies can be found in this directory:

1. `https-behind-load-balancer`: A deployment behind a “dumb” (OSI level 4) load balancer where each Caddy instance does its own SSL termination.
2. `http-behind-ingress`: A deployment behind a “smart” (OSI level 7) Ingress controller that does termination of SSL connections and then load balances the HTTP traffic to the Caddy instances.

We will consider each of these scenarios in turn.  Both treatments assume that you are using a prebuilt Caddy+`adobe_usage_tracker` container from an image repository.  If, for any reason, you need to build your own custom container, see the [docker deployment README](../docker/README.md) for instructions.

## SSL-terminating Caddy

As part of its standard operation, Caddy maintains two persistent directories where it caches runtime information: one for its configuration one for served content. In the Kubernetes environment, these two directories are provided by persistent volume claims found in this directory:

* `tracker-config-volume-persistentvolumeclaim.yaml` for configuration.
* `tracker-data-volume-persistentvolumeclaim.yaml` for content.

Both of these should be loaded with `kubectl apply -f` as your first step.

In the TLS-terminating configuration directory `https-behind-load-balancer`, there are two configmaps that contain the needed handshake information:

* `tracker-ulecs-cert-configmap.yaml` which has the certificate in PEM format.
* `tracker-ulecs-key-configmap.yaml` which has the certificate’s private key in PEM format.

You will need to replace the indented PEM content in both of these YAML files with the content of your certificate and key, leaving the four spaces of indentation intact (because it’s required as part of the YAML format). Generally, the easiest way to do this is to replace the entire set of lines with the content of your certificate, and then indent all of the lines by four spaces using a text editor.

Once you have updated these files with your data, load them with `kubectl apply -f`.

In the TLS-terminating configuration directory `https-behind-load-balancer`, there is a configmap file `tracker-caddyfile-configmap.yaml` that contains the Caddyfile.  You will need to replace the four configuration parameters for the `adobe_usage_tracker` plugin in this file with the correct values for your influx database.  Because Caddyfile resources are indented with tabs, but block strings in YAML files must have indentation that starts with spaces, the whitespace in the lines of this file are very carefully formatted.  So you should replace the value of the parameter on each line in a way that doesn’t disturb any other part of the line (including the indentation).

Once you have updated the Caddyfile configmap source, load it with `kubectl apply -f`.

In the TLS-terminating configuration directory `https-behind-load-balancer`, there is a service file `tracker-service.yaml` that contains a Kubernetes load-balanced service configuration. No values in this file need to be altered; just load it with `kubectl apply -f`.

In the TLS-terminating configuration directory `https-behind-load-balancer`, there is a deployment file `tracker-deployment.yaml` that contains the Caddy deployment. This deployment is configured to use a prebuilt Caddy+`adobe_usage_tracker` image (that contains both arm64 and amd64 executables) from the clickonetwo account on Docker hub, but you can configure it to use any Caddy+`adobe_usage_tracker` image you have built and pushed to Docker hub.

Once you load the `tracker-deployment` file with `kubectl apply -f`, one instance of your Caddy service will be running and ready for uploads behind a load balancer provided by your hosting vendor.  You can then scale the deployment to any number of replicas with the `kubectl scale` command.

## Ingress-fronted Caddy

As part of its standard operation, Caddy maintains two persistent directories where it caches runtime information: one for its configuration one for served content. In the Kubernetes environment, these two directories are provided by persistent volume claims found in this directory:

* `tracker-config-volume-persistentvolumeclaim.yaml` for configuration.
* `tracker-data-volume-persistentvolumeclaim.yaml` for content.

Both of these should be loaded with `kubectl apply -f` as your first step.

In the Ingress-fronted configuration directory `http-behind-ingress`, there is a configmap file `tracker-caddyfile-configmap.yaml` that contains the Caddyfile.  You will need to replace the four configuration parameters for the `adobe_usage_tracker` plugin in this file with the correct values for your influx database.  Because Caddyfile resources are indented with tabs, but block strings in YAML files must have indentation that starts with spaces, the whitespace in the lines of this file are very carefully formatted.  So you should replace the value of the parameter on each line in a way that doesn’t disturb any other part of the line (including the indentation).

Once you have updated the Caddyfile configmap source, load it with `kubectl apply -f`.

In the Ingress-fronted configuration directory `http-behind-ingress`, there is a service file `tracker-service.yaml` that contains a Kubernetes ClusterIP service configuration. No values in this file need to be altered; just load it with `kubectl apply -f`.

In the Ingress-fronted configuration directory `http-behind-ingress`, there is a deployment file `tracker-deployment.yaml` that contains the Caddy deployment. This deployment is configured to use a prebuilt Caddy+`adobe_usage_tracker` image (that contains both arm64 and amd64 executables) from the clickonetwo account on Docker hub, but you can configure it to use any Caddy+`adobe_usage_tracker` image you have built and pushed to Docker hub.

Once you load the `tracker-deployment` file with `kubectl apply -f`, one instance of your Caddy service will be running internal to the cluster but not yet reachable from clients.  You can then scale the deployment to any number of replicas with the `kubectl scale` command.

In the Ingress-fronted configuration directory `http-behind-ingress`, there is a secret definition file `lcs-ulecs-secret.yaml` that contains the certificate and private key your Ingress controller will use to terminate SSL connections.  This file contains key-value pairs for two keys, `tls.crt` and `tls.key`, each of which contains the base64-encoded content of the appropriate PEM file (one for the certificate and one for the key). You will need to replace each of these values with the ones for your environment, which can be obtained by running the command `base64 < pemfile` over your certificate pemfile and your private key pemfile. (The starting sequence for the output in both cases will be `LS0tLS1CRUdJTiB`.)

Once you have updated the secret definition, load it with `kubectl apply -f`

In the Ingress-fronted configuration directory `http-behind-ingress`, there is an Ingress definition file  [lcs-ulecs-ingress.yaml](http-behind-ingress/lcs-ulecs-ingress.yaml) that has an `ingressClassName` of `nginx`. If your cloud provider provides an nginx Ingress controller, and you haven’t got any other services being fronted by this controller, you shouldn’t have to edit this file.  You can simply enable the Ingress controller and then apply the definition with `kubectl apply -f`. At that point your Caddy server instances will be available to your clients.

## A note about DNS

Kubernetes pods can be told to use DNS resolution other than that provided by the cluster, and the deployment files in both these scenarios use Google’s public DNS for resolution within the pod.  Thus, even if you are doing DNS spoofing of `lcs-ulecs.adobe.io` on your intranet, you can run the Kubernetes cluster on that network without issues.

