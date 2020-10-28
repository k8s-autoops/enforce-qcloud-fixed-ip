# enforce-qcloud-fixed-ip

自动强制为 StatefulSet 类型开启腾讯云 TKE 固定 Pod IP 功能

## 使用方式

本组件使用 `admission-bootstrapper` 安装，首先参照此文档 https://github.com/k8s-autoops/admission-bootstrapper ，完成 `admission-bootstrapper` 的初始化

然后，部署以下 YAML 即可

```yaml
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

## Credits

Guo Y.K., MIT License
