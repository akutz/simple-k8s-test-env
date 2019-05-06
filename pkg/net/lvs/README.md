# Linux Virtual Switch

## Director
This section describes how to configure the LVS director.

### iptables
The director's iptables rules should match the following, with the CIDR blocks replaced with values that match the local network settings:

```shell
###############################################################################
##                               General rules                               ##
###############################################################################
*filter
:INPUT DROP [0:0]
:FORWARD DROP [0:0]
:OUTPUT DROP [0:0]

# Allow local-only connections
-A INPUT -i lo -j ACCEPT

# Allow established traffic
-A INPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT

# Allow SSH
-A INPUT -i eth0 -p tcp -m tcp --dport 22 -j ACCEPT

# Allow all outgoing packets on the local-only and public interfaces
-A OUTPUT -o eth0 -j ACCEPT
-A OUTPUT -o lo -j ACCEPT

###############################################################################
##                                  LVS/NAT                                  ##
##                            eth1, 192.168.20.0/24                          ##
###############################################################################
# Allow all traffic coming from the LVS network into the LVS interface
-A INPUT -i eth1 -s 192.168.20.0/24 -j ACCEPT

# Allow established NAT traffic for eth0->eth1
-A FORWARD -i eth0 -o eth1 -m state --state RELATED,ESTABLISHED -j ACCEPT

# Allow forwarding from eth1->eth0
-A FORWARD -i eth1 -o eth0 -j ACCEPT

# Allow outgoing packets destined for the LVS network on the LVS interface.
-A OUTPUT -o eth1 -d 192.168.20.0/24 -j ACCEPT

COMMIT

###############################################################################
##                                  LVS/NAT                                  ##
###############################################################################
*nat
-A POSTROUTING -o eth0 -j MASQUERADE
COMMIT
```