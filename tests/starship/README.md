# Starship Tests

Multi-chain e2e tests that run against any chain, using chain binaries, relayers
and deploying Mesh-Security contracts.

Starship runs by separating out the infra from the tests that are run against the infra.

## Getting Started

### Setup script

In the `tests/starship` dir, run

```bash
make setup-deps ## Installs dependencies for Starship
```

### Manul install (alternate)

Alternatively to the setup script one can just install the deps directly:

* docker: <https://docs.docker.com/get-docker/>
* kubectl: <https://kubernetes.io/docs/tasks/tools/>
* kind: <https://kind.sigs.k8s.io/docs/user/quick-start/#installation>
* helm: <https://helm.sh/docs/intro/install/>
* yq: <https://github.com/mikefarah/yq/#install>

## Connect to a kubernetes cluster

### Spinup local cluster

On Linux:

```bash
make setup-kind
```

On Mac:
Use Docker Desktop to setup kubernetes cluster: <https://docs.docker.com/desktop/kubernetes/#turn-on-kubernetes>

### Connect to a remote cluster (alternate)

If one has access to a k8s cluster via a `kubeconfig` file one can run Starship directly on the remote cluster.

## Check connection with cluster

Run

```bash
kubectl get nodes
```

## Run Tests

Once the initial connection and setup is done, then one can spin up starship infra with

```bash
make install
# OR if you want to run specific config file
make install FILE=configs/devnet.yaml
```

Once the helm chart is installed, you will have to wait for pods to be in a `Running` state. Usually takes 3-5 mins depending on the resources available.
Can check with

```bash
kubectl get pods
```

When all pods are in `Running` state, run port-forwarding to access the nodes on localhost

```bash
make port-forward
# All exposed endpoints would be printed by this command
```

Now you can run the tests with:

```bash
make test
```

Once done, cleanup with:

```bash
make stop
```

## Configs

Starship configs is the definition of the infra we want to spin up.
Present in `test/starship/configs`, are multiple versions of the similar infra, tweaked to be able to run in different environments

* `configs/local.yaml`: Config file to be able to run locally
* `configs/devnet.yaml`: Supposed to be run on a larger k8s cluster, with more resources and number of validators
* `configs/ci.yaml`: Limited resources on the GH-Action runner, can be adapted for with reducing cpu,memory allocated

All the config files are similar topology, but different resources allocated.
Topology:

* 2 chains: `mesh-1` and `mesh-2` (both running `mesh-security-sdk` demo app)
* 1 hermes relayer: running between the chains, in pull mode (1.6.0)
* Registry service: analogous to cosmos chain-registry, but for only our infra
* Optionally explorer: ping-pub explorer for the mini cosmos

Details of each of arguments in the config file can be found [here](https://starship.cosmology.tech/config/chains)
