### 简介
该 SDK 实现多K8S集群的资源操作，包括创建、查询、删除、更新等。还增加了Crd资源的支持。
支持 restApi 结构体的定义，也支持直接使用yaml操作。

### 使用
```bash
go get -u github.com/xops-infra/multi-k8s-client@main
```

### 更新日志
- 2024-04
    - feat: support sparkOnK8S;
- 2024-02
    - feat: init k8s from kubePath&kubeConfig;
    - feat: support flink crd list with filter;
- 2024-01
    - 支持 Flink CRD，支持 Sesson，Application 集群创建，以及对 session集群提交任务接口。（集群需要预先安装好cert-manager和flink-operator-repo，可以参考官网https://nightlies.apache.org/flink/flink-kubernetes-operator-docs-release-1.7/docs/try-flink-kubernetes-operator/quick-start/）