# Docker Deployment

The Caddy web server, with the `adobe_usage_tracker` plugin, can be deployed in a Docker container that is directly exposed to the network. The sample files for such a deployment can be found in this directory. Here are the steps you need to go through to make this work in your environment.

First, copy this entire `docker` directory to the server that will run your container, and switch into that directory.

Second, replace the content of the two certificate files in that directory (`lcs-ulecs.pem.cert` and `lcs-ulecs.pem.key`) with your trusted certificate and key content.

Third, replace the content in the `adobe_usage_tracker` configuration section in the sample Caddyfile with proper parameters for your Influx installation.

Fourth, fetch or build a container image of Caddy with the `adobe_usage_tracker` plugin installed.

* If you are running a Linux OS environment on an arm64 or amd64 architecture, you can simply fetch the pre-built container mentioned in the `compose.yaml` file from the Docker hub with
   `docker pull clickonetwo/adobe_usage_tracker:1.0.0`.
  In this case, no edit of the compose file is required.
* If you are running a non-Linux OS or an unusual processor architecture, you will need to build the container yourself using the Dockerfile in this directory, using the command
  `docker build -t my_docker_account/adobe_usage_tracker:1.0.0 .`
  and push it to your image repository with
  `docker push my_docker_account/adobe_usage_tracker:1.0.0`
  In this case, you will also need to edit the `compose.yaml` file to pull that image.

At this point, you are ready to run Caddy in its container with the command:

```bash
docker compose up -d
```

The container will run detached and be connected to port 443 on all addresses of the host machine.  You can see the last 1000 lines of the Caddy logs at any time by issuing the command

```bash
docker compose logs -n=1000
```

Add the `-f` flag to that command if you want to see new log entries as they appear.

## A note about DNS

Docker containers can be told to use DNS resolution other than that provided by the host machine, and the `compose.yaml` file in this directory directs Docker to use Google’s public DNS for resolution within the container.  Thus, even if you are doing DNS spoofing of `lcs-ulecs.adobe.io` on your intranet, you can run the containerized Caddy on that network without issues.

