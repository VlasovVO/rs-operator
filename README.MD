# Реализация ReplicaSet при помощи Operator SDK

Кастомная реализация функционала ReplicaSet.

Собиралась и тестировалась на версиях:

> Operator SDK: v0.18.0
> 
> Kubernetes: v0.17.0
> 
> minikube v1.11.0

## Как установить

```bash
# Устанавливаем CRD
$ kubectl apply -f deploy/crds/my.replica.com_myrs_crd.yaml

# Добавляем RBAC и оператор
$ kubectl apply -f deploy/
```


## Запуск

Пример кастомного ресурса [Можно посмотреть тут]() и при необходимости создать свой

```bash
# Добавить предоставленый манифест CR можно командой
$ kubectl apply -f deploy/crds/my.replica.com_v1alpha1_myrs_cr.yaml

# Спустя некоторые время поды появятся и будут доступны
$ kubectl get pods
NAME                             READY   STATUS    RESTARTS   AGE
example-myrs-pod-2g994           1/1     Running   0          68s
example-myrs-pod-55lp5           1/1     Running   0          68s
example-myrs-pod-chxz5           1/1     Running   0          68s
example-myrs-pod-shxsx           1/1     Running   0          68s
```

