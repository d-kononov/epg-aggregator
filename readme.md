# Astra EPG aggregator

## Pre-Configuration:

Required Astra 5.62-rc8 or newer

1. Install, configure and launch [Astra](https://cesbo.com/astra/quick-start/):
2. In the Stream options (on Astra that receives channels), set EPG Export option. Format: `JSON` and Destination: `http://server-address:8080` (server-address - ip address of the server with this script, 8080 - is a port number defined in the `PORT` variable)

## Configuration

### Environment variables

`PORT` - By default it serves on :8080 unless a `PORT` environment variable was defined.

`OUTPUT_PATH` - By default it uses `./static/epg.xml(.gz)` path to store file unless `OUTPUT_PATH` was defined.

`INTERVAL` - By default it uses 60 seconds interval to put xml data to the file unless `INTERVAL` (in seconds) was defined.

`EXPIRE` - By default it uses 24 hours before exclude channel from xml data unless `EXPIRE` (in hours) was defined.

`STALKER` - By default it rename <sub-title> tag to <desc> in xml data unless `STALKER=false` was set.

## Run

### Docker

[Docker image](https://hub.docker.com/r/freeman1988/epg-aggregator).

`docker run freeman1988/epg-aggregator:latest`

### Kubernetes

`kubectl apply -f deployment.yaml`. Create ingress if needed.

### Systemd

Put `systemd/astra-epg-aggregator.service` file to the `/etc/systemd/system/` folder and run `systemd daemon-reload`.

Download archive file from the latest release, unpack it to `/usr/local/bin/` folder and run `systemd enable astra-epg-aggregator`, `systemd start astra-epg-aggregator` 
