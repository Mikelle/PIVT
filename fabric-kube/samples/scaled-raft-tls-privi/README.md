A scaled up HL Fabric network with 3 nodes of Raft orderer spanning 2 organizations and 2 peers per organization. 
Also demonstrates how TLS can be enabled and actual domain names (instead of internal Kubernetes service names) can be used.

Note, when TLS is enabled, transparent load balancing is not posssible.

See scaled-raft-no-tls sample for how to use Raft orderer without enabling TLS globally.