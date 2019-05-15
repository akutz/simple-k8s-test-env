/*
simple-kubernetes-test-environment

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lvs

const getFreePortCmd = `read _l _u </proc/sys/net/ipv4/ip_local_port_range && ` +
	`while true; do ` +
	`_p=$(shuf -i $_l-$_u -n 1); ss -lpn | grep -q $_p || break; ` +
	`done`

// createRRTCPServiceCmdPatt is the command used to create an LVS service.
//
// The command is formatted with a Go template and expects the following
// object:
//     {
//         NIC     string   // name of the public interface
//         SID     string   // service ID
//         VIP     string   // virtual IP address
//     }
const createRRTCPServiceCmdPatt = `` +
	`sudo mkdir -p /var/run/sk8 && ` +
	`sudo flock /var/run/sk8/lvs.lock sh -c '` +
	`cat /var/run/sk8/{{.SID}}.lvs 2>/dev/null || { ` +
	getFreePortCmd + " && " +
	`iptables -A INPUT -i {{.NIC}} -p tcp -m tcp --dport $_p -j ACCEPT && ` +
	`ipvsadm -A -t {{.VIP}}:$_p -s rr && ` +
	`echo $_p | tee /var/run/sk8/{{.SID}}.lvs` +
	`; }'`

// addRRTCPServiceCmdPatt is the command used to append a target to an
// LVS service.
//
// The command is formatted with a Go template and expects the following
// object:
//     {
//         SID     string          // service ID
//         VIP     string          // virtual IP address
//         Target  ServiceEndpoint // target service
//     }
const addRRTCPServiceCmdPatt = `` +
	`sudo mkdir -p /var/run/sk8 && ` +
	`sudo flock /var/run/sk8/lvs.lock sh -c '` +
	`_p=$(cat /var/run/sk8/{{.SID}}.lvs) && { { ` +
	`ipvsadm -ln -t {{.VIP}}:$_p | grep -q {{.Target.Addr}}:{{.Target.Port}}` +
	`; } || { ` +
	`ipvsadm -a -t {{.VIP}}:$_p -r {{.Target.Addr}}:{{.Target.Port}} -m` +
	`; }; }'`

// setOrGetTCPServiceCmdPatt is the command used to set a single target
// on an LVS service or return the target already set for that service.
//
// The command is formatted with a Go template and expects the following
// object:
//     {
//         SID     string          // service ID
//         VIP     string          // virtual IP address
//         Target  ServiceEndpoint // target service
//     }
const setOrGetTCPServiceCmdPatt = `` +
	`sudo mkdir -p /var/run/sk8 && ` +
	`sudo flock /var/run/sk8/lvs.lock sh -c '` +
	`_p=$(cat /var/run/sk8/{{.SID}}.lvs) && { { ` +
	`_a=$(ipvsadm -ln -t {{.VIP}}:$_p | grep Masq | awk "{print \$2}") && ` +
	`[ -n "${_a}" ] && echo "${_a}"` +
	`; } 2>/dev/null || { ` +
	`ipvsadm -a -t {{.VIP}}:$_p -r {{.Target.Addr}}:{{.Target.Port}} -m && ` +
	`echo {{.Target.Addr}}:{{.Target.Port}}` +
	`; }; }'`

// deleteRRTCPServiceCmdPatt is the command used to delete an LVS service.
//
// The command is formatted with a Go template and expects the following
// object:
//     {
//         NIC     string   // name of the public interface
//         SID     string   // service ID
//         VIP     string   // virtual IP address
//     }
const deleteRRTCPServiceCmdPatt = `` +
	`sudo mkdir -p /var/run/sk8 && ` +
	`sudo flock /var/run/sk8/lvs.lock sh -c '` +
	`[ ! -f '/var/run/sk8/{{.SID}}.lvs' ] || { ` +
	`_p=$(cat /var/run/sk8/{{.SID}}.lvs) && ` +
	`iptables -D INPUT -i {{.NIC}} -p tcp -m tcp --dport $_p -j ACCEPT && ` +
	`ipvsadm -D -t {{.VIP}}:$(cat /var/run/sk8/{{.SID}}.lvs) && ` +
	`rm -f /var/run/sk8/{{.SID}}.lvs` +
	`; }'`
