### 简介

该 SDK 实现多 K8S 集群的资源操作，包括创建、查询、删除、更新等。还增加了 Crd 资源的支持。
支持 restApi 结构体的定义，也支持直接使用 yaml 操作。

### 使用

```bash
go get -u github.com/xops-infra/multi-k8s-client@main
```

### 更新日志

- 2024-06
  - feat: support flink v 1.12(flink-operator is not support flink version below 1.12);
- 2024-04
  - feat: support sparkOnK8S;
- 2024-02
  - feat: init k8s from kubePath&kubeConfig;
  - feat: support flink crd list with filter;
- 2024-01
  - 支持 Flink CRD，支持 Sesson，Application 集群创建，以及对 session 集群提交任务接口。（集群需要预先安装好 cert-manager 和 flink-operator-repo，可以参考官网https://nightlies.apache.org/flink/flink-kubernetes-operator-docs-release-1.7/docs/try-flink-kubernetes-operator/quick-start/）
