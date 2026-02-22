# Karlsen Stratum Adapter
[![Join the Karlsen Discord Server](https://img.shields.io/discord/1169939685280337930.svg?label=&logo=discord&logoColor=ffffff)](https://discord.gg/ZPZRvgMJDT)

This is a lightweight daemon that allows mining to a local (or remote)
karlsen node using stratum-based miners.

This daemon is confirmed working with the miners below in both dual-mining
and karlsen-only modes (for those that support it) and Windows, Linux,
macOS and HiveOS.

- **Official miners:**
  - [GPU miner](https://github.com/karlsen-network/karlsen-miner/releases/latest) (reference implementation)
- **Third-party miners with KarlsenHashv2 support:**
  - [lolMiner](https://github.com/Lolliedieb/lolMiner-releases)
  - [BzMiner](https://github.com/bzminer/bzminer)
  - [SRBMiner](https://github.com/doktor83/SRBMiner-Multi)
  - [Rigel](https://github.com/rigelminer/rigel)

**Tips appreciated**: `karlsen:qqe3p64wpjf5y27kxppxrgks298ge6lhu6ws7ndx4tswzj7c84qkjlrspcuxw`

## Hive Setup

[Detailed Instructions here](docs/hive-setup.md)

# Features

Shares-based work allocation with miner-like periodic stat output:

```
===============================================================================
  worker name   |  avg hashrate  |   acc/stl/inv  |    blocks    |    uptime
-------------------------------------------------------------------------------
 lemois         |       0.13GH/s |          3/0/0 |            0 |       6m48s
-------------------------------------------------------------------------------
                |       0.13GH/s |          3/0/0 |            0 |       7m20s
========================================================= kls_bridge_v2.3.0 ===
```

## Variable difficulty engine (vardiff)

Multiple miners with significantly different hashrates can be connected
to the same stratum bridge instance, and the appropriate difficulty
will automatically be decided for each one. Default settings target
15 shares/min, resulting in high confidence decisions regarding
difficulty adjustments, and stable measured hashrates (1hr avg
hashrates within +/- 10% of actual). The minimum share difficulty is 64
and optimized for GPUs.

## Grafana UI

The grafana monitoring UI is an optional component but included for
convenience. It will help to visualize collected statistics.

[Detailed Instructions here](docs/monitoring-setup.md)

## Prometheus API

If the app is run with the `-prom={port}` flag the application will host
stats on the port specified by `{port}`, these stats are documented in
the file [prom.go](src/karlsenstratum/prom.go). This is intended to be use
by prometheus but the stats can be fetched and used independently if
desired. `curl http://localhost:2114/metrics | grep kls_` will get a
listing of current stats. All published stats have a `kls_` prefix for
ease of use.

```
user:~$ curl http://localhost:2114/metrics | grep kls_
# HELP kls_estimated_network_hashrate_gauge Gauge representing the estimated network hashrate
# TYPE kls_estimated_network_hashrate_gauge gauge
kls_estimated_network_hashrate_gauge 2.43428982879776e+14
# HELP kls_network_block_count Gauge representing the network block count
# TYPE kls_network_block_count gauge
kls_network_block_count 271966
# HELP kls_network_difficulty_gauge Gauge representing the network difficulty
# TYPE kls_network_difficulty_gauge gauge
kls_network_difficulty_gauge 1.2526479386202519e+14
# HELP kls_valid_share_counter Number of shares found by worker over time
# TYPE kls_valid_share_counter counter
kls_valid_share_counter{ip="192.168.0.17",miner="SRBMiner-MULTI/2.4.1",wallet="karlsen:qzk3uh2twkhu0fmuq50mdy3r2yzuwqvstq745hxs7tet25hfd4egcafcdmpdl",worker="002"} 276
kls_valid_share_counter{ip="192.168.0.24",miner="SRBMiner-MULTI/2.4.1",wallet="karlsen:qzk3uh2twkhu0fmuq50mdy3r2yzuwqvstq745hxs7tet25hfd4egcafcdmpdl",worker="003"} 43
kls_valid_share_counter{ip="192.168.0.65",miner="SRBMiner-MULTI/2.4.1",wallet="karlsen:qzk3uh2twkhu0fmuq50mdy3r2yzuwqvstq745hxs7tet25hfd4egcafcdmpdl",worker="001"} 307
# HELP kls_worker_job_counter Number of jobs sent to the miner by worker over time
# TYPE kls_worker_job_counter counter
kls_worker_job_counter{ip="192.168.0.17",miner="SRBMiner-MULTI/2.4.1",wallet="karlsen:qzk3uh2twkhu0fmuq50mdy3r2yzuwqvstq745hxs7tet25hfd4egcafcdmpdl",worker="002"} 3471
kls_worker_job_counter{ip="192.168.0.24",miner="SRBMiner-MULTI/2.4.1",wallet="karlsen:qzk3uh2twkhu0fmuq50mdy3r2yzuwqvstq745hxs7tet25hfd4egcafcdmpdl",worker="003"} 3399
kls_worker_job_counter{ip="192.168.0.65",miner="SRBMiner-MULTI/2.4.1",wallet="karlsen:qzk3uh2twkhu0fmuq50mdy3r2yzuwqvstq745hxs7tet25hfd4egcafcdmpdl",worker="001"} 3425
```

# Install

## Build from source (native executable)

Install go 1.18 or later using whatever package manager is approprate
for your system, or from [https://go.dev/doc/install](https://go.dev/doc/install).

```
cd cmd/karlsenbridge
go build .
```

Modify the config file in `./cmd/karlsenbridge/config.yaml` with your setup,
the file comments explain the various flags.

```
./karlsenbridge
```

To recap the entire process of initiating the compilation and launching
the karlsen dridge, follow these steps:

```
cd cmd/karlsenbridge
go build .
./karlsenbridge
```

## Docker (all-in-one)

Best option for users who want access to reporting, and aren't already
using Grafana/Prometheus. Requires a local copy of this repository, and
docker installation.

[Install Docker](https://docs.docker.com/engine/install/) using the
appropriate method for your OS. The docker commands below are assuming a
server type installation - details may be different for a desktop
installation.

The following will run the bridge assuming a local karlsend node with
default port settings, and listen on port 5555 for incoming stratum
connections.

```
git clone https://github.com/karlsen-network/karlsen-stratum-bridge.git
cd karlsen-stratum-bridge
docker compose -f docker-compose-all-src.yml up -d --build
```

These settings can be updated in the [config.yaml](cmd/karlsenbridge/config.yaml)
file, or overridden by modifying, adding or deleting the parameters in the
`command` section of the `docker-compose-all-src.yml` file. Additionally,
Prometheus (the stats database) and Grafana (the dashboard) will be
started and accessible on ports 9090 and 3000 respectively. Once all
services are running, the dashboard should be reachable at
`http://127.0.0.1:3000/d/x7cE7G74k1/klsb-monitoring` with default
username and password `admin`.

These commands builds the bridge component from source, rather than
the previous behavior of pulling down a pre-built image. You may still
use the pre-built image by replacing `docker-compose-all-src.yml` with
`docker-compose-all.yml`, but it is not guaranteed to be up to date, so
compiling from source is the better alternative.

## Docker (bridge only)

Best option for users who want docker encapsulation, and don't need
reporting, or are already using Grafana/Prometheus. Requires a local
copy of this repository, and docker installation.

[Install Docker](https://docs.docker.com/engine/install/) using the
appropriate method for your OS. The docker commands below are assuming a
server type installation - details may be different for a desktop
installation.

The following will run the bridge assuming a local karlsend node with
default port settings, and listen on port 5555 for incoming stratum
connections.

```
git clone https://github.com/karlsen-network/karlsen-stratum-bridge.git
cd karlsen-stratum-bridge
docker compose -f docker-compose-bridge-src.yml up -d --build
```

These settings can be updated in the [config.yaml](cmd/karlsenbridge/config.yaml)
file, or overridden by modifying, adding or deleting the parameters in the
`command` section of the `docker-compose-bridge-src.yml`

These commands builds the bridge component from source, rather than the
previous behavior of pulling down a pre-built image. You may still use
the pre-built image by issuing the command `docker run -p 5555:5555 karlsennetwork/karlsen_bridge:latest`,
but it is not guaranteed to be up to date, so compiling from source is
the better alternative.

## Credits
Huge shoutout to https://github.com/KaffinPX/KStratum for the inspiration
  
Tips appreciated: 
- [@onemorebsmith](https://github.com/onemorebsmith): `kaspa:qp9v6090sr8jjlkq7r3f4h9un5rtfhu3raknfg3cca9eapzee57jzew0kxwlp`
- [@rdugan](https://github.com/rdugan): `kaspa:qrkhyhej7h0gmmvsuf8mmufget4n4xnlwx5j360sz70q7xvu0hlaxfmt9p8j8`
