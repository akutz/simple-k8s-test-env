#!/bin/sh

tdnf upgrade -y && \
tdnf install -y bindutils \
                cni \
                ethtool \
                gawk \
                inotify-tools \
                ipset \
                iputils \
                ipvsadm \
                libnetfilter_conntrack \
                libnetfilter_cthelper \
                libnetfilter_cttimeout \
                libnetfilter_queue \
                jq \
                lsof \
                socat \
                sudo \
                tar \
                unzip && \
curl -sSL https://raw.githubusercontent.com/vmware/cloud-init-vmware-guestinfo/master/install.sh | sh - && \
rm -f /etc/cloud/cloud.cfg.d/99-disable-networking-config.cfg && \
{ cat >/etc/systemd/scripts/ip4save <<EOF
*filter
:INPUT DROP [0:0]
:FORWARD DROP [0:0]
:OUTPUT DROP [0:0]

# Block all null packets.
-A INPUT -p tcp --tcp-flags ALL NONE -j DROP

# Reject a syn-flood attack.
-A INPUT -p tcp ! --syn -m state --state NEW -j DROP

# Block XMAS/recon packets.
-A INPUT -p tcp --tcp-flags ALL ALL -j DROP

# Allow all incoming packets on the loopback interface.
-A INPUT -i lo -j ACCEPT

# Allow incoming packets for established connections.
-A INPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT

# Allow SSH on all interfaces.
-A INPUT -p tcp -m tcp --dport 22 -j ACCEPT

# Allow all outgoing packets.
-A OUTPUT -j ACCEPT

# Enable the rules.
COMMIT
EOF
} && \
cp -f /etc/systemd/scripts/ip4save /etc/systemd/scripts/ip6save && \
printf 'changeme\nchangeme' | passwd && \
printf '' >/etc/machine-id && \
rm -fr /var/lib/cloud/instances && \
rm -rf /etc/ssh/*key* /root/.ssh && \
rm -fr /var/log && mkdir -p /var/log && \
echo 'clearing history & sealing the VM...' && \
unset HISTFILE && history -c && rm -fr /root/.bash_history && \
shutdown -P now