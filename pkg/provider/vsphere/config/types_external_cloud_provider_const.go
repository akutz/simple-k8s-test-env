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

package config

const extCCMDefaultImage = "gcr.io/cloud-provider-vsphere/vsphere-cloud-controller-manager:latest"

const extCCMSecretsFormat = `apiVersion: v1
kind: Secret
metadata:
  name: {{.SecretName}}
  namespace: {{.Namespace}}
data:
  {{.ServerAddr}}.username: "{{base64 .Username}}"
  {{.ServerAddr}}.password: "{{base64 .Password}}"
`

const extCCMConfigFormat = `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.ConfigMapName}}
  namespace: {{.Namespace}}
data:
  {{.ConfigFileName}}: |
    [Global]
    secret-name =      "{{.SecretName}}"
    secret-namespace = "{{.Namespace}}"
    service-account =  "{{.ServiceAccount}}"

    port =             "{{.ServerPort}}"
    insecure-flag =    "{{.Insecure}}"
    datacenters =      "{{join .Datacenters ","}}"

    [VirtualCenter "{{.ServerAddr}}"]{{if .Region}}{{if .Zone}}

    [Labels]
      region = "{{.Region}}"
      zone =   "{{.Zone}}"{{end}}{{end}}
`

const extCCMDeployPodFormat = `apiVersion: v1
kind: Pod
metadata:
  annotations:
    scheduler.alpha.kubernetes.io/critical-pod: ""
  labels:
    component: cloud-controller-manager
    tier: control-plane
  name: vsphere-cloud-controller-manager
  namespace: {{.Namespace}}
spec:
  containers:
  - name: vsphere-cloud-controller-manager
    image: "{{.Image}}"
    args:
    - /bin/vsphere-cloud-controller-manager
    - --v=2
    - --cloud-config=/etc/cloud/{{.ConfigFileName}}
    - --cloud-provider=vsphere
    volumeMounts:
    - mountPath: /etc/cloud
      name: cloud-config-volume
      readOnly: true
    resources:
      requests:
        cpu: 200m
  hostNetwork: true
  tolerations:
  - key: node.cloudprovider.kubernetes.io/uninitialized
    value: "true"
    effect: NoSchedule
  - key: node.kubernetes.io/not-ready
    effect: NoSchedule
  - key: node-role.kubernetes.io/master
    effect: NoSchedule
  securityContext:
    runAsUser: 1001
  serviceAccountName: {{.ServiceAccount}}
  volumes:
  - name: cloud-config-volume
    configMap:
      name: {{.ConfigMapName}}
`

const extCCMRolesFormat = `apiVersion: v1
items:
- apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  metadata:
    name: system:cloud-controller-manager
  rules:
  - apiGroups:
    - ""
    resources:
    - events
    verbs:
    - create
    - patch
    - update
  - apiGroups:
    - ""
    resources:
    - nodes
    verbs:
    - '*'
  - apiGroups:
    - ""
    resources:
    - nodes/status
    verbs:
    - patch
  - apiGroups:
    - ""
    resources:
    - services
    verbs:
    - list
    - patch
    - update
    - watch
  - apiGroups:
    - ""
    resources:
    - serviceaccounts
    verbs:
    - create
    - get
    - list
    - watch
    - update
  - apiGroups:
    - ""
    resources:
    - persistentvolumes
    verbs:
    - get
    - list
    - update
    - watch
  - apiGroups:
    - ""
    resources:
    - endpoints
    verbs:
    - create
    - get
    - list
    - watch
    - update
  - apiGroups:
    - ""
    resources:
    - secrets
    verbs:
    - get
    - list
    - watch
kind: List
metadata: {}
`

const extCCMRoleBindingsFormat = `apiVersion: v1
items:
- apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRoleBinding
  metadata:
    name: system:cloud-controller-manager
  roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: ClusterRole
    name: system:cloud-controller-manager
  subjects:
  - kind: ServiceAccount
    name: {{.ServiceAccount}}
    namespace: {{.Namespace}}
  - kind: User
    name: {{.ServiceAccount}}
kind: List
metadata: {}
`
const extCCMServiceFormat = `apiVersion: v1
kind: Service
metadata:
  labels:
    component: cloud-controller-manager
  name: vsphere-cloud-controller-manager
  namespace: {{.Namespace}}
spec:
  type: NodePort
  ports:
  - port: {{.ServicePort}}
    protocol: TCP
    targetPort: {{.ServicePort}}
  selector:
    component: cloud-controller-manager
`

const extCCMServiceAccountFormat = `apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{.ServiceAccount}}
  namespace: {{.Namespace}}
`
