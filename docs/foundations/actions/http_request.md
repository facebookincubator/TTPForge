# TTPForge Actions: `http_request`

The `http_request` action can be used to send http requests to a target host and
store the response without the need to invoke a shell and use `curl`, `wget`,
`nc`, etc. Check out the TTP below to see how it works:

https://github.com/facebookincubator/TTPForge/blob/93206fad43ae0087b1d136dffc9ea7a3f3dd95d3/example-ttps/actions/http-request/get-parameters.yaml#L1-L15

You can experiment with the above TTP by installing the `examples` TTP
repository (skip this if `ttpforge list repos` shows that the `examples` repo is
already installed):

```bash
ttpforge install repo https://github.com/facebookincubator/TTPForge --name examples
```

and then running the below command:

```bash
ttpforge run examples//actions/http-request/get-parameters.yaml
```

## Fields

You can specify the following YAML fields for the `http_request:` action:

- `http_request:` (type: `string`) URL to which the request is made.
- `type:` (type: `string`) The http request type (`GET`, `POST`, `PUT`,`PATCH`,
  `DELETE`).
- `response_headers:` (type: `bool`) Whether or not the http response headers
  are included with the response (default: `false`).
  - **Note:** Set this to `true` for `HEAD` requests, or no response will
    appear!
- `disable_redirects:` (type: `bool`) Don't follow http redirects (default:
  `false`).
- `headers:` (type: `header`) The http request headers:
  - `field:` (type: `string`) HTTP header field.
  - `value:` (type: `string`) HTTP header value.
- `parameters:` (type: `parameter`) The http request parameters:
  - `name:` (type: `string`) Name of the http parameter
  - `value:` (type: `string`) Value of the http parameter.
- `body:` (type: `string`) String for request body data.
- `proxy:` (type: `string`) The http proxy to use for requests
- `regex:` (type: `string`) Regular expression, if specified return only
  matching string.
- `response:` (type: `string`) Shell variable name to store request's response.
- `cleanup:` You can define a custom
  [cleanup action](https://github.com/facebookincubator/TTPForge/blob/main/docs/foundations/cleanup.md#cleanup-basics).
