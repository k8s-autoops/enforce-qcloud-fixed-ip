# enforce-qcloud-fixed-ip

自动强制为 StatefulSet 类型开启腾讯云 TKE 固定 Pod IP 功能

## 提示

本功能依赖于腾讯云 TKE 的全局 VPC-CNI 模式（非 GlobalRouter + VPC-CNI 模式）

## 使用方式

* 初始化 `admission-bootstrapper`
  本组件使用 `admission-bootstrapper` 安装，首先参照此文档 https://github.com/k8s-autoops/admission-bootstrapper ，完成 `admission-bootstrapper` 的初始化

* 部署以下 YAML

```yaml
# create serviceaccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: enforce-qcloud-fixed-ip
  namespace: autoops
---
# create clusterrole
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: enforce-qcloud-fixed-ip
rules:
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get"]
---
# create clusterrolebinding
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: enforce-qcloud-fixed-ip
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: enforce-qcloud-fixed-ip
subjects:
  - kind: ServiceAccount
    name: enforce-qcloud-fixed-ip
    namespace: autoops
---
# create job
apiVersion: batch/v1
kind: Job
metadata:
  name: install-enforce-qcloud-fixed-ip
  namespace: autoops
spec:
  template:
    spec:
      serviceAccount: admission-bootstrapper
      containers:
        - name: admission-bootstrapper
          image: autoops/admission-bootstrapper
          env:
            - name: ADMISSION_NAME
              value: enforce-qcloud-fixed-ip
            - name: ADMISSION_IMAGE
              value: autoops/enforce-qcloud-fixed-ip
            - name: ADMISSION_ENVS
              value: ""
            - name: ADMISSION_SERVICE_ACCOUNT
              value: "enforce-qcloud-fixed-ip"
            - name: ADMISSION_MUTATING
              value: "true"
            - name: ADMISSION_IGNORE_FAILURE
              value: "false"
            - name: ADMISSION_SIDE_EFFECT
              value: "None"
            - name: ADMISSION_RULES
              value: '[{"operations":["CREATE"],"apiGroups":["apps"], "apiVersions":["*"], "resources":["statefulsets"]}]'
      restartPolicy: OnFailure
```

* 为需要开启此功能的 Namespace 添加注解

  `autoops.enforce-qcloud-fixed-ip=true`

  **可以配合 `enforce-ns-annotations` 自动为新命名空间启用此注解**

## Credits

Guo Y.K., MIT License
