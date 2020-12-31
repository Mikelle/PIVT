# Hyperledger Fabric meets Kubernetes
![Fabric Meets K8S](https://raft-fabric-kube.s3-eu-west-1.amazonaws.com/images/fabric_meets_k8s.png)

* [License](#License)
* [Requirements](#requirements)
* [Scaled-up Raft network with TLS PRIVI](#scaled-up-raft-network-with-tls-privi)
* [Limitations](#limitations)
* [FAQ and more](#faq-and-more)
* [Conclusion](#conclusion)

## [License](#License)
This work is licensed under the same license with HL Fabric; [Apache License 2.0](LICENSE).

## [Requirements](#requirements)
* A running Kubernetes(v1.16) cluster, Microk8s also work
* [HL Fabric binaries](https://hyperledger-fabric.readthedocs.io/en/release-1.4/install.html)
* [Helm](https://github.com/helm/helm/releases/tag/v2.16.0), 2.16 or newer 2.xx versions
* [jq](https://stedolan.github.io/jq/download/) 1.5+ and [yq](https://pypi.org/project/yq/) 2.6+
* [Argo](https://github.com/argoproj/argo), both CLI and Controller 2.4.0+
* [Minio](https://github.com/argoproj/argo/blob/master/docs/configure-artifact-repository.md), only required for backup/restore and new-peer-org flows
* Run all the commands in *fabric-kube* folder
* AWS EKS users please also apply this [fix](https://github.com/APGGroeiFabriek/PIVT/issues/1)

### [Scaled-up Raft network with TLS PRIVI](#scaled-up-raft-network-with-tls-privi)

First tear down everything:
```
argo delete --all
helm delete hlf-kube --purge
```
Wait a bit until all pods are terminated:
```
kubectl  get pod --watch
```
Then create necessary stuff:
```
./init.sh ./samples/scaled-raft-tls-privi/ ./samples/chaincode/
```
Lets launch our Raft based Fabric network in _broken_ state:
```
helm install ./hlf-kube --name hlf-kube -f samples/scaled-raft-tls-privi/network.yaml -f samples/scaled-raft-tls-privi/crypto-config.yaml 
```
The pods will start but they cannot communicate to each other since domain names are unknown. You might also want to use the option `--set peer.launchPods=false --set orderer.launchPods=false` to make this process faster.

Run this command to collect the host aliases:
```
kubectl get svc -l addToHostAliases=true -o jsonpath='{"hostAliases:\n"}{range..items[*]}- ip: {.spec.clusterIP}{"\n"}  hostnames: [{.metadata.labels.fqdn}]{"\n"}{end}' > samples/scaled-raft-tls-privi/hostAliases.yaml
```

Or this one, which is much convenient:
```
./collect_host_aliases.sh ./samples/scaled-raft-tls-privi/ 
```

Let's check the created hostAliases.yaml file.
```
cat samples/scaled-raft-tls-privi/hostAliases.yaml
```

The output will be something like:
```
hostAliases:
- ip: 10.152.183.14
  hostnames: [orderer0.cache.com]
- ip: 10.152.183.61
  hostnames: [orderer1.cache.com]
- ip: 10.152.183.166
  hostnames: [orderer2.cache.com]
- ip: 10.152.183.74
  hostnames: [peer0.privi.com]
- ip: 10.152.183.81
  hostnames: [peer1.privi.com]
  ```
The IPs are internal ClusterIPs of related services. Important point here is, as opposed to pod ClusterIPs, service ClusterIPs are stable, they won't change if service is not deleted and re-created.

Next, let's update the network with this host aliases information. These entries goes into pods' `/etc/hosts` file via Pod [hostAliases](https://kubernetes.io/docs/concepts/services-networking/add-entries-to-pod-etc-hosts-with-host-aliases/) spec.
```
helm upgrade hlf-kube ./hlf-kube -f samples/scaled-raft-tls-privi/network.yaml -f samples/scaled-raft-tls-privi/crypto-config.yaml -f samples/scaled-raft-tls-privi/hostAliases.yaml  
```

Again lets wait for all pods are up and running:
```
kubectl get pod --watch
```
Congrulations you have a running scaled up HL Fabric network in Kubernetes, with 3 Raft orderer nodes spanning 1 Orderer organization and 2 peers per organization. But unfortunately, due to TLS, your application cannot use them with transparent load balancing, you need to connect to relevant peer and orderer services separately.

Lets create the channels:
```
helm template channel-flow/ -f samples/scaled-raft-tls-privi/network.yaml -f samples/scaled-raft-tls-privi/crypto-config.yaml -f samples/scaled-raft-tls-privi/hostAliases.yaml | argo submit - --watch
```
And install chaincodes:
```
helm template chaincode-flow/ -f samples/scaled-raft-tls-privi/network.yaml -f samples/scaled-raft-tls-privi/crypto-config.yaml -f samples/scaled-raft-tls-privi/hostAliases.yaml | argo submit - --watch
```

### network.yaml 
This file defines how network is populated regarding channels and chaincodes.

```yaml
network:
  # used by init script to create genesis block and by peer-org-flow to parse consortiums
  genesisProfile: OrdererGenesis
  # used by init script to create genesis block 
  systemChannelID: testchainid

  # defines which organizations will join to which channels
  channels:
    - name: common
      # all peers in these organizations will join the channel
      orgs: [Karga, Nevergreen, Atlantis]
    - name: private-karga-atlantis
      # all peers in these organizations will join the channel
      orgs: [Karga, Atlantis]

  # defines which chaincodes will be installed to which organizations
  chaincodes:
    - name: very-simple
      # if defined, this will override the global chaincode.version value
      version: # "2.0" 
      # chaincode will be installed to all peers in these organizations
      orgs: [Karga, Nevergreen, Atlantis]
      # at which channels are we instantiating/upgrading chaincode?
      channels:
      - name: common
        # chaincode will be instantiated/upgraded using the first peer in the first organization
        # chaincode will be invoked on all peers in these organizations
        orgs: [Karga, Nevergreen, Atlantis]
        policy: OR('KargaMSP.member','NevergreenMSP.member','AtlantisMSP.member')
        
    - name: even-simpler
      orgs: [Karga, Atlantis]
      channels:
      - name: private-karga-atlantis
        orgs: [Karga, Atlantis]
        policy: OR('KargaMSP.member','AtlantisMSP.member')
```

For chart specific configuration, please refer to the comments in the relevant [values.yaml](fabric-kube/hlf-kube/values.yaml) files.

## [Limitations](#limitations)

### TLS

Transparent load balancing is not possible when TLS is globally enabled. So, instead of `Peer-Org`, `Orderer-Org` or `Orderer-LB` services, you need to connect to individual `Peer` and `Orderer` services.

Running Raft orderers without globally enabling TLS is possible since Fabric 1.4.5. See [Scaled-up Raft network without TLS](#scaled-up-raft-network-without-tls) sample for details.

### Multiple Fabric networks in the same Kubernetes cluster

This is possible but they should be run in different namespaces. We do not use Helm release name in names of components, 
so if multiple instances of Fabric network is running in the same namespace, names will conflict.

## [FAQ and more](#faq-and-more)

Please see [FAQ](FAQ.md) page for further details. Also this [post](https://accenture.github.io/blog/2019/06/25/hl-fabric-meets-kubernetes.html) at Accenture's open source blog provides some additional information like motivation, how it works, benefits regarding NFR's, etc.

## [Conclusion](#conclusion)

So happy BlockChaining in Kubernetes :)

And don't forget the first rule of BlockChain club:

**"Do not use BlockChain unless absolutely necessary!"**

*Hakan Eryargi (r a f t)*
