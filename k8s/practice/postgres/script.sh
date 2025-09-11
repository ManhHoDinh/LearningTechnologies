#!/bin/bash -i
set -e

kubectl apply -f postgres-password-cm.yaml
kubectl delete cm postgres-master-configmap  || true
kubectl create cm postgres-master-configmap  --from-file=postgresql.conf --from-file=pg_hba.conf || true
kubectl apply -f postgres-master-sts.yaml
kubectl exec -it postgres-master-0 -- bash
su - postgres
psql
SET password_encryption = 'scram-sha-256';
CREATE ROLE repuser WITH REPLICATION PASSWORD 'postgres' LOGIN;
SELECT * FROM pg_create_physical_replication_slot('replica_1_slot');
exit
exit
exit

kubectl apply -f postgres-master-svc.yaml

cd slave
kubectl apply -f pvc-slave.yaml
kubectl apply -f sync-master-data.yaml
kubectl create cm postgres-slave-configmap --from-file=slave-config
kubectl apply -f postgres-slave-sts.yaml

# 🔥 OR create with helm
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install postgresql bitnami/postgresql-ha --set postgresqlPassword=postgres --set replication.password=postgres
