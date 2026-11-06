### 简介

该 SDK 实现多 K8S 集群的资源操作，包括创建、查询、删除、更新等。还增加了 Crd 资源的支持。
支持 restApi 结构体的定义，也支持直接使用 yaml 操作。

### 使用

建议固定经过验证的版本，避免直接依赖 `@main`；将下方占位符替换为实际发布的版本号。本次修复的新 tag 尚未发布。

```bash
go get github.com/xops-infra/multi-k8s-client@<已发布版本号>
```

### Flink 1.12 的 ServiceAccount token 权限

Flink 1.12 的 JobManager（含 sidecar 场景）和 TaskManager Pod 模板设置 `securityContext.fsGroup: 9999` 与 `fsGroupChangePolicy: OnRootMismatch`，使镜像中切换到 UID/GID 9999 的 Flink 进程可读取投射的 ServiceAccount token。容器仍声明 `runAsUser: 0`，并保留原有 postStart 权限初始化钩子。

升级 SDK 不会自动修改既有工作负载；需要通过更新后的模板重新部署才能应用此配置。ServiceAccount 及其所需 RBAC 授权仍须正确配置，token 可读不能替代授权；默认 ServiceAccount 选择和 token 自动挂载行为保持不变。

`fsGroup` 也可能影响支持该机制的 PVC 卷的属组和访问权限，部署前需评估现有数据权限；`OnRootMismatch` 可在卷根目录权限匹配时避免递归修改权限。

### 更新日志

- 2025-05-16

  - feat: 支持 FlinkDeployment 的 labels 设置，支持 app 和 owner 标签；

- 2024-11

  - feat: 支持 Flink operator 方式自动创建 LB service，查询；

- 2024-08

  - feat: 支持 CrdFlinkTMScale 调整 TM 数量
  - feat: 支持 FlinkDeploymentRestart 支持制定重启类型，JM/TM/ALL

- 2024-07

  - feat: support jobmanager sidecar;

- 2024-06
  - feat: support flink v1.12(flink-operator is not support flink version below 1.12);
- 2024-04
  - feat: support sparkOnK8S;
- 2024-02
  - feat: init k8s from kubePath&kubeConfig;
  - feat: support flink crd list with filter;
- 2024-01
  - 支持 Flink CRD，支持 Sesson，Application 集群创建，以及对 session 集群提交任务接口。（集群需要预先安装好 cert-manager 和 flink-operator-repo，可以参考官网https://nightlies.apache.org/flink/flink-kubernetes-operator-docs-release-1.7/docs/try-flink-kubernetes-operator/quick-start/）
