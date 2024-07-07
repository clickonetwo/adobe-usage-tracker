# Adobe Usage Tracker

This repository houses the `adobe_usage_tracker`, which is a plugin for the [Caddy web server](https://caddyserver.com). The `adobe_usage_tracker` plugin acts as a reverse proxy for the Adobe log upload server `lcs-cops.adobe.io`.  It analyzes the logs being uploaded and sends the resulting information about application usage to an Influx database.

The rest of this documentation assumes you know how to configure, deploy, and use the Caddy web server with plugins.  The [Caddy webserver documentation](https://caddyserver.com/docs/) provides an excellent overview, sample deployment information, and in-depth reference material that you can use to get started. 

## Configuration

There are two types of configuration required for proper function of the `adobe_usage_tracker`: DNS configuration and API configuration.

### DNS Configuration

Adobe applications hardcode the `https://lcs-ulecs.adobe.io` URL as the endpoint to which they upload logs. In order for the `adobe_usage_tracker` plugin to see these log uploads, you will have to set up your Caddy web server as a “transparent proxy” that the Adobe applications think is the Adobe log server.  This means that:

*  your Caddy web server must have a certificate for `lcs-ulecs.adobe.io`,
*  your client machines must trust that certificate, and
* you must configure Caddy so that the certificate is used when clients contact `lcs-ulecs.adobe.io`.  

The Caddy configuration command for configuring a particular certificate looks like this, where the certificate file and the certificate key file are both in PEM format:

```Caddyfile
tls <certificate-file> <certificate-key-file>
```

and you will need to place this command in the section of your Caddyfile that proxies `https` calls to `lcs-ulecs.adobe.io`. For example, if you are using Caddy exclusively to proxy log traffic, and your certificates are called `lcs-ulecs.cert.pem` and `lcs-ulecs.key.pem`, your Caddyfile might look like this:

```Caddyfile
:443 {
	tls lcs-ulecs.cert.pem lcs-ulecs.key.pem
	... other configuration described below ...
	reverse-proxy https://lcs-ulecs.adobe.io
}
```

This section says to accept incoming traffic on port 443 (https), use the cert/key combo to respond to the TLS challenge, and then to proxy all the traffic to (the real) `lcs-ulecs.adobe.io`.  The Caddy server will do its own connect to the Adobe log server, and do its own verification of its certificate.

One subtlety that matters here: notice that the Caddy server is both masquerading as `lcs-ulecs.adobe.io` and then connecting to (the real) `lcs-ulecs.adobe.io`.  If you are spoofing DNS for your client machines in order for them to connect to your Caddy-based proxy, you must ensure that the host runninig your Caddy server correctly resolves `lcs-ulecs.adobe.io` against Adobe’s nameservers, *not* against its own address. Otherwise the Caddy server will attempt to forward each request to itself.

### API Configuration

The `adobe_usage_tracker` plugin uses the Influx v1 API to upload log measurements to the Influx database.  This API requires four values:

* The API host URL for the target Influx database. This URL should include protocol (either `http` or `https` and hostname and (optionally) a port. It should not include any path components.
* The name of the Infux database to which the log measurements should be uploaded.
* The retention policy for use with the log measurements in that database.
* An authorization token for the given host and database that has upload permissions.

All versions of Influx support uploads via the v1 API.  But if your Influx installation uses “buckets” (Influx v2.7 and higher) you will need to establish a [*DBRP mapping*](https://docs.influxdata.com/influxdb/v2/reference/api/influxdb-1x/dbrp/) before you can configure your plugin parameters.  The docs for using the Influx CLI to establish a mapping can be found [here for self-hosted configurations](https://docs.influxdata.com/influxdb/v2/reference/cli/influx/v1/dbrp/), and [here for cloud-hosted configurations](https://docs.influxdata.com/influxdb/cloud/query-data/influxql/dbrp/).

The `adobe_usage_tracker` plugin includes in the timestream data the remote host address from which each log is uploaded. Determination of this address is done by looking for an `X-Forwarded-For` header in the request and using the first address found in that header; if no such header is found then it uses the address from which the Caddy server received the request.  Depending on your deployment environment, you may want the plugin to use a different header (such as `Via` or `X-Real-IP`) or to use the last address found in that header rather than the first.  Both of these can be configured.

Once you have determined the correct Influx API parameters and header controls for your usage, you configure your `adobe_usage_tracker` plugin by adding a snippet like this to your Caddyfile (replacing all the values in angle brackets with values appropriate to your environment):

```Caddyfile
adobe_usage_tracker {
    endpoint <https://influxUploadHost.mydomain.com>
    database <influxDatabaseName>
    policy <infuxRetentionPolicyName<
    token <influxApiTokenWithUploadPrivilege>
    header <headerName or "" for no header>
    position <first or last>
}
```

This snippet, as with the `tls` snippet shown above, should be placed in your Caddyfile in the entry for log upload.  Working Caddyfiles with instructions may be found in the deploy directory in this repository (see next section). The four Influx API parameters _must_ be supplied, but the `header` and `position` parameters are both optional (defaulting to `X-Forwarded-For` and `first`, respectively).

## Deployment Scenarios

There are instructions and sample files for different types of deployments in this repository:

* deployment of Caddy and `adobe_usage_tracker` directly to a  [network web server](deploy/server/README.md).
* deployment of Caddy and `adobe_usage_tracker` in [a single container exposed to the network](deploy/docker/README.md) .
* deployment of Caddy and `adobe_usage_tracker` via [Kubernetes](deploy/k8s/README.md) . 

Each of these deployment scenarios has been tested successfully using their example files, so you can be assured that all the syntax and structure of the files is correct.  But you will, of course, have to change the content of the files so that they are relevant to your environment. For example, you will need to generate an `lcs-ulecs.adobe.io` certificate (with key) that is trusted by your client machines. And you will need influx upload parameters that work with your database.

## License and Attribution

The material in this repository is licensed under the [MIT license](https://opensource.org/license/mit), which is reproduced in full in the [LICENSE](LICENSE) file.

The code in this repository makes use of the [Caddy web server](https://caddyserver.com) SDK and its sample code, both of which are also licensed under the MIT license.

Adobe is either a registered trademark or a trademark of Adobe in the United States and/or other countries.
