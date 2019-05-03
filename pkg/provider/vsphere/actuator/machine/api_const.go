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

package machine

const weaveWorksYAML = `apiVersion: v1
kind: List
items:
  - apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: weave-net
      annotations:
        cloud.weave.works/launcher-info: |-
          {
            "original-request": {
              "url": "/k8s/v1.10/net.yaml?k8s-version=Q2xpZW50IFZlcnNpb246IHZlcnNpb24uSW5mb3tNYWpvcjoiMSIsIE1pbm9yOiIxMSIsIEdpdFZlcnNpb246InYxLjExLjMiLCBHaXRDb21taXQ6ImE0NTI5NDY0ZTQ2MjljMjEyMjRiM2Q1MmVkZmUwZWE5MWIwNzI4NjIiLCBHaXRUcmVlU3RhdGU6ImNsZWFuIiwgQnVpbGREYXRlOiIyMDE4LTA5LTA5VDE4OjAyOjQ3WiIsIEdvVmVyc2lvbjoiZ28xLjEwLjMiLCBDb21waWxlcjoiZ2MiLCBQbGF0Zm9ybToiZGFyd2luL2FtZDY0In0K",
              "date": "Thu Apr 04 2019 20:59:48 GMT+0000 (UTC)"
            },
            "email-address": "support@weave.works"
          }
      labels:
        name: weave-net
      namespace: kube-system
  - apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: ClusterRole
    metadata:
      name: weave-net
      annotations:
        cloud.weave.works/launcher-info: |-
          {
            "original-request": {
              "url": "/k8s/v1.10/net.yaml?k8s-version=Q2xpZW50IFZlcnNpb246IHZlcnNpb24uSW5mb3tNYWpvcjoiMSIsIE1pbm9yOiIxMSIsIEdpdFZlcnNpb246InYxLjExLjMiLCBHaXRDb21taXQ6ImE0NTI5NDY0ZTQ2MjljMjEyMjRiM2Q1MmVkZmUwZWE5MWIwNzI4NjIiLCBHaXRUcmVlU3RhdGU6ImNsZWFuIiwgQnVpbGREYXRlOiIyMDE4LTA5LTA5VDE4OjAyOjQ3WiIsIEdvVmVyc2lvbjoiZ28xLjEwLjMiLCBDb21waWxlcjoiZ2MiLCBQbGF0Zm9ybToiZGFyd2luL2FtZDY0In0K",
              "date": "Thu Apr 04 2019 20:59:48 GMT+0000 (UTC)"
            },
            "email-address": "support@weave.works"
          }
      labels:
        name: weave-net
    rules:
      - apiGroups:
          - ''
        resources:
          - pods
          - namespaces
          - nodes
        verbs:
          - get
          - list
          - watch
      - apiGroups:
          - networking.k8s.io
        resources:
          - networkpolicies
        verbs:
          - get
          - list
          - watch
      - apiGroups:
          - ''
        resources:
          - nodes/status
        verbs:
          - patch
          - update
  - apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: ClusterRoleBinding
    metadata:
      name: weave-net
      annotations:
        cloud.weave.works/launcher-info: |-
          {
            "original-request": {
              "url": "/k8s/v1.10/net.yaml?k8s-version=Q2xpZW50IFZlcnNpb246IHZlcnNpb24uSW5mb3tNYWpvcjoiMSIsIE1pbm9yOiIxMSIsIEdpdFZlcnNpb246InYxLjExLjMiLCBHaXRDb21taXQ6ImE0NTI5NDY0ZTQ2MjljMjEyMjRiM2Q1MmVkZmUwZWE5MWIwNzI4NjIiLCBHaXRUcmVlU3RhdGU6ImNsZWFuIiwgQnVpbGREYXRlOiIyMDE4LTA5LTA5VDE4OjAyOjQ3WiIsIEdvVmVyc2lvbjoiZ28xLjEwLjMiLCBDb21waWxlcjoiZ2MiLCBQbGF0Zm9ybToiZGFyd2luL2FtZDY0In0K",
              "date": "Thu Apr 04 2019 20:59:48 GMT+0000 (UTC)"
            },
            "email-address": "support@weave.works"
          }
      labels:
        name: weave-net
    roleRef:
      kind: ClusterRole
      name: weave-net
      apiGroup: rbac.authorization.k8s.io
    subjects:
      - kind: ServiceAccount
        name: weave-net
        namespace: kube-system
  - apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: Role
    metadata:
      name: weave-net
      annotations:
        cloud.weave.works/launcher-info: |-
          {
            "original-request": {
              "url": "/k8s/v1.10/net.yaml?k8s-version=Q2xpZW50IFZlcnNpb246IHZlcnNpb24uSW5mb3tNYWpvcjoiMSIsIE1pbm9yOiIxMSIsIEdpdFZlcnNpb246InYxLjExLjMiLCBHaXRDb21taXQ6ImE0NTI5NDY0ZTQ2MjljMjEyMjRiM2Q1MmVkZmUwZWE5MWIwNzI4NjIiLCBHaXRUcmVlU3RhdGU6ImNsZWFuIiwgQnVpbGREYXRlOiIyMDE4LTA5LTA5VDE4OjAyOjQ3WiIsIEdvVmVyc2lvbjoiZ28xLjEwLjMiLCBDb21waWxlcjoiZ2MiLCBQbGF0Zm9ybToiZGFyd2luL2FtZDY0In0K",
              "date": "Thu Apr 04 2019 20:59:48 GMT+0000 (UTC)"
            },
            "email-address": "support@weave.works"
          }
      labels:
        name: weave-net
      namespace: kube-system
    rules:
      - apiGroups:
          - ''
        resourceNames:
          - weave-net
        resources:
          - configmaps
        verbs:
          - get
          - update
      - apiGroups:
          - ''
        resources:
          - configmaps
        verbs:
          - create
  - apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: RoleBinding
    metadata:
      name: weave-net
      annotations:
        cloud.weave.works/launcher-info: |-
          {
            "original-request": {
              "url": "/k8s/v1.10/net.yaml?k8s-version=Q2xpZW50IFZlcnNpb246IHZlcnNpb24uSW5mb3tNYWpvcjoiMSIsIE1pbm9yOiIxMSIsIEdpdFZlcnNpb246InYxLjExLjMiLCBHaXRDb21taXQ6ImE0NTI5NDY0ZTQ2MjljMjEyMjRiM2Q1MmVkZmUwZWE5MWIwNzI4NjIiLCBHaXRUcmVlU3RhdGU6ImNsZWFuIiwgQnVpbGREYXRlOiIyMDE4LTA5LTA5VDE4OjAyOjQ3WiIsIEdvVmVyc2lvbjoiZ28xLjEwLjMiLCBDb21waWxlcjoiZ2MiLCBQbGF0Zm9ybToiZGFyd2luL2FtZDY0In0K",
              "date": "Thu Apr 04 2019 20:59:48 GMT+0000 (UTC)"
            },
            "email-address": "support@weave.works"
          }
      labels:
        name: weave-net
      namespace: kube-system
    roleRef:
      kind: Role
      name: weave-net
      apiGroup: rbac.authorization.k8s.io
    subjects:
      - kind: ServiceAccount
        name: weave-net
        namespace: kube-system
  - apiVersion: extensions/v1beta1
    kind: DaemonSet
    metadata:
      name: weave-net
      annotations:
        cloud.weave.works/launcher-info: |-
          {
            "original-request": {
              "url": "/k8s/v1.10/net.yaml?k8s-version=Q2xpZW50IFZlcnNpb246IHZlcnNpb24uSW5mb3tNYWpvcjoiMSIsIE1pbm9yOiIxMSIsIEdpdFZlcnNpb246InYxLjExLjMiLCBHaXRDb21taXQ6ImE0NTI5NDY0ZTQ2MjljMjEyMjRiM2Q1MmVkZmUwZWE5MWIwNzI4NjIiLCBHaXRUcmVlU3RhdGU6ImNsZWFuIiwgQnVpbGREYXRlOiIyMDE4LTA5LTA5VDE4OjAyOjQ3WiIsIEdvVmVyc2lvbjoiZ28xLjEwLjMiLCBDb21waWxlcjoiZ2MiLCBQbGF0Zm9ybToiZGFyd2luL2FtZDY0In0K",
              "date": "Thu Apr 04 2019 20:59:48 GMT+0000 (UTC)"
            },
            "email-address": "support@weave.works"
          }
      labels:
        name: weave-net
      namespace: kube-system
    spec:
      minReadySeconds: 5
      template:
        metadata:
          labels:
            name: weave-net
        spec:
          containers:
            - name: weave
              command:
                - /home/weave/launch.sh
              env:
                - name: HOSTNAME
                  valueFrom:
                    fieldRef:
                      apiVersion: v1
                      fieldPath: spec.nodeName
              image: 'docker.io/weaveworks/weave-kube:2.5.1'
              readinessProbe:
                httpGet:
                  host: 127.0.0.1
                  path: /status
                  port: 6784
              resources:
                requests:
                  cpu: 10m
              securityContext:
                privileged: true
              volumeMounts:
                - name: weavedb
                  mountPath: /weavedb
                - name: cni-bin
                  mountPath: /host/opt
                - name: cni-bin2
                  mountPath: /host/home
                - name: cni-conf
                  mountPath: /host/etc
                - name: dbus
                  mountPath: /host/var/lib/dbus
                - name: lib-modules
                  mountPath: /lib/modules
                - name: xtables-lock
                  mountPath: /run/xtables.lock
            - name: weave-npc
              env:
                - name: HOSTNAME
                  valueFrom:
                    fieldRef:
                      apiVersion: v1
                      fieldPath: spec.nodeName
              image: 'docker.io/weaveworks/weave-npc:2.5.1'
              resources:
                requests:
                  cpu: 10m
              securityContext:
                privileged: true
              volumeMounts:
                - name: xtables-lock
                  mountPath: /run/xtables.lock
          hostNetwork: true
          hostPID: true
          restartPolicy: Always
          securityContext:
            seLinuxOptions: {}
          serviceAccountName: weave-net
          tolerations:
            - effect: NoSchedule
              operator: Exists
          volumes:
            - name: weavedb
              hostPath:
                path: /var/lib/weave
            - name: cni-bin
              hostPath:
                path: /opt
            - name: cni-bin2
              hostPath:
                path: /home
            - name: cni-conf
              hostPath:
                path: /etc
            - name: dbus
              hostPath:
                path: /var/lib/dbus
            - name: lib-modules
              hostPath:
                path: /lib/modules
            - name: xtables-lock
              hostPath:
                path: /run/xtables.lock
                type: FileOrCreate
      updateStrategy:
        type: RollingUpdate
`
