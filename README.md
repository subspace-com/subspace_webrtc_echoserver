# subspace_webrtc_echoserver
This application listens on HTTP(S) for WebRTC offers, accepts incoming PeerConnections, accepts DataChannels, then echoes back any data received on a DataChannel.

Certificates can be configured with `CERT_FILE` and `KEY_FILE` as ENV variables (see pion/.env).

It can be configured with a STUN server (`STUN_URL`) or directly with an external IP address
(`EXTERNAL_IP`) in case of 1:1 NAT mapping like in AWS or GCP.

In order to support Cross-Origin Source requests, it's possible to define a list of allowed origins inside `ALLOWED_ORIGINS`.

## Environment variables

### SERVER_PORT

Optional setting. Port where the server will listen for HTTP(S) requests.
Defaults to `443`.


### SERVER_ADDR

Optional setting. IP address the server will listen on for HTTP(S) requests.
Defaults to `127.0.0.1`.

### STUN_URL

Optional setting. If a STUN URL is provided and `EXTERNAL_IP` is not provided, then the server will use this URL to discover its server reflexive transport address.

### CERT_FILE

Optional setting. If a file path is provided, then the server will use this file as TLS certificate and will server over HTTPS. If omitted the server will serve over HTTP.

### KEY_FILE

Optional setting. If a file path is provided, then the server will use this file as private key for the TLS certificate.

### EXTERNAL_IP

Optional setting. If an IP address is provided, then the server will use this to build server reflexive candidates and will not use STUN to discover it dynamically. This can be used for example with AWS EC2 and GCP, setting the instance elastic IP.

### ALLOWED_ORIGINS

Optional setting. A comma separated list of domains that are allowed as Cross-Origin source. If missing then all domains will be allowed.

### DEBUG

Optional setting. If set to `true`, the server will log debug information.
Default: `false`.

### MIN_PORT

Optional setting. An integer identifying the beginning of the port range to be used for RTP.

### MIN_PORT

Optional setting. An integer identifying the end of the port range to be used for RTP.

### MAX_REQ_RATE

Optional setting. The max request rate allowed from a single IP address, in units per second.
Default: unlimited.


### Example of environment variables

```
    SERVER_PORT=443
    SERVER_ADDR=0.0.0.0
    STUN_URL=stun:stun.l.google.com:19302
    CERT_FILE=certs/certfile.pem
    KEY_FILE=certs/keyfile.pem
    EXTERNAL_IP=1.2.3.4
    ALLOWED_ORIGINS=http://one.com,https://two.com
    MIN_PORT=5000
    MAX_PORT=15000
    MAX_REQ_RATE=10
```


Build with:

```
docker-compose build
```

or

```
docker build -t pion-echo -f pion/Dockerfile .
```

Run with:

```
docker-compose up -d
```

or

```
docker run -i -p 8080:8080 pion-echo
```

## Firewall settings

Ensure `SERVER_PORT` is reachable over TCP, and the range `MIN_PORT` to `MAX_PORT` (if set) is reachable over UDP.

# Sources and references

[Pion WebRTC](https://github.com/pion/webrtc)

[WebRTC echoes, pion](https://github.com/sipsorcery/webrtc-echoes/tree/master/pion)
