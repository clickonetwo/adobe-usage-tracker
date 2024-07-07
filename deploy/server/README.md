# Server Deployment

The Caddy web server, with the `adobe_usage_tracker` plugin, can be deployed directly onto a networked server. The sample files for such a deployment can be found in this directory. Here are the steps you need to go through to make this work in your environment.

First, copy this entire `server` directory onto a build machine that has a working golang ecosystem, and change into that directory.

Second, replace the content of the two certificate files in that directory (`lcs-ulecs.cert.pem` and `lcs-ulecs.key.pem`) with your trusted certificate and key content.

Third, replace the content in the `adobe_usage_tracker` configuration section in the sample Caddyfile with proper parameters for your Influx installation.

Third, build Caddy with the `adobe_usage_tracker` plugin.  This involves using the `xcaddy` command, so start by installing the `xcaddy` command using `go`,  with:

```bash
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
```

Then, invoke the installed `xcaddy` command:

```bash
xcaddy build --with github.com/clickonetwo/tracker@v1.1.0
```

This will create a temporary subdirectory, move into it, fetch and build a `caddy` executable with the plugin, place the resulting `caddy` executable in the original working directory, and then delete the temporary subdirectory.  (See [the xcaddy readme](https://github.com/caddyserver/xcaddy) for full details on how to use `xcaddy`.)

Build notes for developers:

* The above build instructions presume that your build machine has the same OS and processor architecture as your server.  If not, then you will need the xcaddy command to cross-compile caddy for the target, which you can do by prefixing the `xcaddy` command with `env GOOS=<targetos> GOARCH=<targetarch>` .  For example, if the server is an amd64 linux box, the xcaddy-command would be `env GOOS=linux GOARCH=amd64 xcaddy build --with github.com/clickonetwo/tracker@v1.1.0`.
* The `tracker` module file `go.mod` specifies using a version of the Caddy server that matches the version of Caddy used by the Caddy builder image in the Dockerfile in the `docker` deployment sample directory (currently 2.8.1). This is not the latest version of Caddy, but `xcaddy` will fetch the latest version of Caddy available when it does its build.

Fourth (and finally), move the entire `server` directory with the built executable onto your server

Once these steps are complete, you can run the tracker manually as follows:

1. Change into the `server` directory.
2. Start Caddy with `./caddy run`. (If Caddy fails to start because it needs privileges to listen on port 443, the try `sudo ./caddy run`.)
3. Caddy will run, with log output to the terminal, until interrupted.

To set Caddy up as a service, follow the instructions for your OS.  You will want to redirect standard error to a log file or, even better, add this snippet at the very top of your caddy file:

```Caddyfile
{
	log {
		output file caddy.log
	}
}
```

### A note about DNS

The `Caddyfile` in this directory relies on Caddy’s host to do DNS resolution, and specifies the upstream host as `https://lcs-ulecs.adobe.io`.  Thus, if you are using DNS spoofing in your local network to force clients to find your server, and you are running Caddy on a server on your local network, you can’t use this `Caddyfile`.  If you did, the DNS resolution for `lcs-ulecs.adobe.com` would find your own server, *not* the Adobe servers.

If you want to run your Caddy server on a local network with DNS spoofing in place, then use the `Caddyfile-with-public-dns` configuration file in this directory, which tells Caddy to resolve the `lcs-ulecs.adobe.io` address dynamically against Google’s public DNS servers.  You can do this either by renaming this file to `Caddyfile` and using the instructions above, or by invoking Caddy with

```bash
./caddy run --config Caddyfile-with-public-dns
```

