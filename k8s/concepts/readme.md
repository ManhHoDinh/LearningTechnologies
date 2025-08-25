# Kubernetes Core Concepts Study Guide

## Table of Contents

1. [Pod](#1-pod)
2. [ReplicaSet](#2-replicaset) 
3. [Deployment](#3-deployment)
4. [Service](#4-service)
5. [Ingress](#5-ingress)
6. [Persistent Storage](#6-persistent-storage)
7. [StatefulSet](#7-statefulset)
8. [Health Probes](#8-health-probes)
9. [ConfigMap & Secret](#9-configmap--secret)
10. [DaemonSet](#10-daemonset)

---

## 1. Pod

### What it is
The smallest deployable unit in Kubernetes. A pod represents a group of one or more containers that:

- **Share the same lifecycle:** Created and destroyed together
- **Share the same network namespace:** Same IP address, can communicate via localhost
- **Share storage volumes:** Can access the same mounted volumes
- **Are always co-located:** Scheduled on the same node

### Key Characteristics
- **Ephemeral:** Pods can be created, destroyed, and recreated
- **Immutable:** Once created, pod specs cannot be changed (except for specific fields)
- **Single IP:** Each pod gets a unique IP address within the cluster
- **Shared storage:** Containers in a pod can share data via volumes

### Pod Lifecycle States
1. **Pending:** Pod accepted but not yet scheduled or containers not ready
2. **Running:** Pod bound to node, all containers created, at least one running
3. **Succeeded:** All containers terminated successfully (exit 0)
4. **Failed:** All containers terminated, at least one failed (non-zero exit)
5. **Unknown:** Pod state cannot be determined

### Restart Policies
- **Always (default):** Restart containers when they exit
- **OnFailure:** Restart only if container exits with non-zero status
- **Never:** Never restart containers

### Use Cases for Multi-Container Pods

#### 1. Sidecar Pattern
Helper container supports main application:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: web-with-sidecar
spec:
  containers:
  - name: web-server
    image: nginx:1.25
    ports:
    - containerPort: 80
    volumeMounts:
    - name: shared-logs
      mountPath: /var/log/nginx
  - name: log-agent
    image: fluent/fluentd:v1.14
    volumeMounts:
    - name: shared-logs
      mountPath: /var/log/nginx
      readOnly: true
  volumes:
  - name: shared-logs
    emptyDir: {}
```

#### 2. Ambassador Pattern
Proxy container handles external communications:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-proxy
spec:
  containers:
  - name: app
    image: myapp:1.0
    ports:
    - containerPort: 8080
  - name: ambassador
    image: envoyproxy/envoy:v1.24.0
    ports:
    - containerPort: 9901
```

#### 3. Adapter Pattern
Transforms data for the main container:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-adapter
spec:
  containers:
  - name: app
    image: legacy-app:1.0
  - name: monitoring-adapter
    image: prometheus/node-exporter:v1.6.0
    ports:
    - containerPort: 9100
```

### Resource Management
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: resource-demo
spec:
  containers:
  - name: app
    image: nginx:1.25
    resources:
      requests:        # Minimum guaranteed resources
        memory: "128Mi"
        cpu: "100m"    # 0.1 CPU core
      limits:          # Maximum allowed resources
        memory: "256Mi"
        cpu: "500m"    # 0.5 CPU core
```

### Security Context
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: security-demo
spec:
  securityContext:
    runAsUser: 1000
    runAsGroup: 3000
    fsGroup: 2000
  containers:
  - name: app
    image: nginx:1.25
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      runAsNonRoot: true
      capabilities:
        drop:
        - ALL
```

### kubectl Commands
- **Create:** `kubectl apply -f pod.yaml`
- **List pods:** `kubectl get pods` or `kubectl get po`
- **Detailed info:** `kubectl describe pod <pod-name>`
- **Get pod IP:** `kubectl get pod <pod-name> -o wide`
- **Pod logs:** `kubectl logs <pod-name>` 
- **Multi-container logs:** `kubectl logs <pod-name> -c <container-name>`
- **Execute commands:** `kubectl exec -it <pod-name> -- /bin/bash`
- **Multi-container exec:** `kubectl exec -it <pod-name> -c <container-name> -- /bin/bash`
- **Port forwarding:** `kubectl port-forward pod/<pod-name> 8080:80`
- **Copy files:** `kubectl cp <pod-name>:/path/file ./local-file`
- **Delete pod:** `kubectl delete pod <pod-name>`

### Common Troubleshooting
- **ImagePullBackOff:** Check image name, registry access, credentials
- **CrashLoopBackOff:** Check container logs and application configuration
- **Pending state:** Check resource availability and node capacity
- **Ready 0/1:** Check readiness probes and application startup

---

## 2. ReplicaSet

### What it is
A ReplicaSet ensures that a specified number of pod replicas are running at any given time. It's a controller that maintains the desired state of identical pods.

**Key Functions:**
- **Maintains replica count:** Ensures exactly N pods are running
- **Self-healing:** Replaces failed or deleted pods automatically
- **Scaling:** Can increase or decrease the number of replicas
- **Label-based selection:** Uses labels to identify which pods it manages

### How it Works
1. ReplicaSet continuously monitors pods matching its label selector
2. If actual replica count < desired → creates new pods
3. If actual replica count > desired → deletes excess pods
4. Uses the pod template to create new pods when needed

### Core Components

#### 1. Replica Count
```yaml
spec:
  replicas: 3  # Desired number of pod replicas
```

#### 2. Label Selector
```yaml
spec:
  selector:
    matchLabels:      # Simple equality-based selector
      app: nginx
      version: "1.0"
  # OR
  selector:
    matchExpressions:  # More complex expression-based selector
    - key: app
      operator: In
      values: ["nginx", "apache"]
    - key: version
      operator: NotIn
      values: ["beta"]
```

#### 3. Pod Template
```yaml
spec:
  template:
    metadata:
      labels:
        app: nginx    # Must match selector labels
        version: "1.0"
    spec:
      containers:
      - name: nginx
        image: nginx:1.25
```

### Complete Example with Advanced Features
```yaml
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: web-server-rs
  labels:
    app: web-server
    tier: frontend
spec:
  replicas: 4
  selector:
    matchLabels:
      app: web-server
      tier: frontend
  template:
    metadata:
      labels:
        app: web-server
        tier: frontend
        version: "v1.2.0"
    spec:
      containers:
      - name: nginx
        image: nginx:1.25
        ports:
        - containerPort: 80
          name: http
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
```

### ReplicaSet vs. Other Controllers

| Controller | Use Case | Updates | Rollback |
|------------|----------|---------|----------|
| **ReplicaSet** | Basic replication | Manual | No |
| **Deployment** | Production apps | Rolling updates | Yes |
| **StatefulSet** | Stateful apps | Ordered updates | Limited |
| **DaemonSet** | Node-level services | Rolling updates | Limited |

### kubectl Commands

#### Basic Operations
- **Create:** `kubectl apply -f replicaset.yaml`
- **List:** `kubectl get rs` or `kubectl get replicasets`
- **Details:** `kubectl describe rs <replicaset-name>`
- **Delete:** `kubectl delete rs <replicaset-name>`

#### Scaling Operations  
- **Scale up:** `kubectl scale rs <replicaset-name> --replicas=6`
- **Scale down:** `kubectl scale rs <replicaset-name> --replicas=2`
- **Auto-scale:** `kubectl autoscale rs <replicaset-name> --min=2 --max=10 --cpu-percent=80`

#### Monitoring and Debugging
- **Watch pods:** `kubectl get pods -l app=nginx -w`
- **Check events:** `kubectl get events --field-selector involvedObject.kind=ReplicaSet`
- **Pod details:** `kubectl get pods -o wide -l app=nginx`

#### Editing and Updates
- **Edit live:** `kubectl edit rs <replicaset-name>`
- **Update image:** `kubectl set image rs/<replicaset-name> container-name=new-image:tag`

### Important Notes

#### Label Selector Requirements
- Selector labels **MUST** match template labels
- Once created, selector is **immutable** (cannot be changed)
- Template labels can be changed, but new pods will use new labels

#### Ownership and Cleanup
- ReplicaSet owns the pods it creates
- Deleting ReplicaSet deletes all its pods by default
- Use `kubectl delete rs <name> --cascade=orphan` to keep pods

### Common Patterns

#### 1. Blue-Green Deployment (Manual)
```bash
# Create green version
kubectl apply -f green-replicaset.yaml

# Switch traffic (update service selector)
kubectl patch service web-service -p '{"spec":{"selector":{"version":"green"}}}'

# Remove blue version
kubectl delete rs blue-replicaset
```

#### 2. Canary Testing
```bash
# Scale down main version
kubectl scale rs main-app --replicas=8

# Create canary version with fewer replicas
kubectl scale rs canary-app --replicas=2
```

### Troubleshooting

#### Common Issues
- **Pods not starting:** Check image name, resource limits, node capacity
- **Selector mismatch:** Ensure selector matches template labels exactly
- **Stuck scaling:** Check resource quotas and node availability
- **Wrong pod count:** Verify no other controllers are managing same pods

#### Debugging Commands
```bash
# Check ReplicaSet status
kubectl get rs <name> -o wide

# View ReplicaSet events
kubectl describe rs <name>

# Check pod status and events
kubectl get pods -l app=<label> -o wide
kubectl describe pod <pod-name>

# Monitor scaling operations
kubectl get rs <name> -w
```

### Best Practices
- **Use Deployments instead:** ReplicaSets are typically managed by Deployments
- **Meaningful labels:** Use descriptive labels for better organization
- **Resource limits:** Always set resource requests and limits
- **Health checks:** Include liveness and readiness probes
- **Monitoring:** Set up monitoring for replica count and pod health

---

## 3. Deployment

### What it is
A higher-level controller that manages ReplicaSets, and indirectly Pods.

Adds rollout, rollback, and declarative updates features.

When you update the spec (e.g. change image), it creates a new ReplicaSet with the new template and slowly scales down the old one.

### Why it exists
ReplicaSets are stable, but they don't handle updates safely. Imagine replacing 100 pods one by one manually! Deployments automate:

- **Rolling updates** → old pods drained gradually, new ones started
- **Rollback** → go back to a previous version if the new one fails
- **History** → track revisions

### Example YAML: Deployment with 2 replicas
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deploy
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.25
        ports:
        - containerPort: 80
```

### kubectl Commands
- **Create:** `kubectl apply -f nginx-deploy.yaml`
- **Check rollout:** `kubectl rollout status deployment/nginx-deploy`
- **Update image:** `kubectl set image deployment/nginx-deploy nginx=nginx:1.27`
- **Rollback:** `kubectl rollout undo deployment/nginx-deploy`
- **See history:** `kubectl rollout history deployment/nginx-deploy`

---

## 4. Service

### What it is
A stable network identity (name + IP + port) that exposes pods.

- Pods are ephemeral (names/IPs change if they're recreated)
- A Service ensures clients always have a consistent endpoint
- Built on labels + selectors → it automatically load-balances to all matching pods

### Types

#### ClusterIP (default)
- Internal only (within cluster)
- Ex: backend DB service

#### NodePort
- Exposes on `<NodeIP>:<Port>` across all nodes
- Simple, but port range is limited (30000–32767)

#### LoadBalancer
- Integrates with cloud providers to provision an external LB
- Common for web apps in GKE/EKS/AKS

#### ExternalName
- Maps service name to an external DNS name

### Example YAML: ClusterIP service
```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-svc
spec:
  selector:
    app: nginx
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: ClusterIP
```

- `selector: app=nginx` → targets pods from our Deployment
- Clients inside the cluster can just do `curl http://nginx-svc`

### kubectl Commands
- **Create:** `kubectl apply -f nginx-svc.yaml`
- **Check:** `kubectl get svc`
- **Test with port-forward:** `kubectl port-forward svc/nginx-svc 8080:80`
- **If NodePort:** access via `http://<nodeIP>:<nodePort>`

---

## 5. Ingress

### What it is
A Kubernetes API object that manages external HTTP/HTTPS access to Services.

Think of it as a smart router or reverse proxy inside the cluster.

**Handles:**
- Path-based routing (`/api` → backend, `/app` → frontend)
- Host-based routing (`api.example.com` vs `shop.example.com`)
- TLS/SSL termination

### Why it exists
- Services of type LoadBalancer = one external IP per service → expensive and messy
- Ingress lets you expose many services behind one load balancer, with rich routing rules

### Example YAML: Basic Ingress
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx-ingress
spec:
  rules:
  - host: example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: nginx-svc
            port:
              number: 80
```

- Requests to `http://example.com/` go to `nginx-svc:80`
- Requires an Ingress Controller (e.g. Nginx Ingress, Traefik, Istio gateway) installed in the cluster

### kubectl Commands
- **Create:** `kubectl apply -f nginx-ingress.yaml`
- **Check:** `kubectl get ingress`
- **Describe:** `kubectl describe ingress nginx-ingress`

---

## 6. Persistent Storage

### Why this matters
Pods are ephemeral; their container filesystems vanish when pods reschedule. Databases, uploads, and caches need durable volumes that survive pod restarts and can move across nodes.

### The trio
- **PersistentVolume (PV):** A real piece of storage in the cluster (NFS, local disk, cloud disk). Cluster‑scoped, created by admins or dynamically by CSI drivers
- **PersistentVolumeClaim (PVC):** An app's request for storage (size, access mode). When bound, it points to one PV
- **StorageClass (SC):** A template describing how to provision PVs on demand (parameters, reclaim policy). With SC, PVCs auto‑create PVs = dynamic provisioning

### Key concepts

#### AccessModes:
- **ReadWriteOnce (RWO):** one node can mount read/write (most block volumes)
- **ReadOnlyMany (ROX):** many nodes read‑only
- **ReadWriteMany (RWX):** many nodes read/write (NFS, some CSI)

#### Reclaim Policies (what happens when a PVC is deleted)
- **Delete:** underlying storage is deleted (common for cloud-managed disks)
- **Retain:** PV is kept, data remains; requires manual cleanup or reuse (best for production DBs)

#### Volume Binding Modes (when PVCs bind to PVs)
- **Immediate:** PVC binds as soon as it's created
- **WaitForFirstConsumer:** PVC stays Pending until a Pod requests it

### Example: Dynamic provisioning

#### 1) StorageClass
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast
provisioner: kubernetes.io/no-provisioner # replace with your CSI
parameters:
  type: gp3
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Delete
```

#### 2) PVC (requests 10Gi, RWO)
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc
spec:
  storageClassName: fast
  accessModes: [ ReadWriteOnce ]
  resources:
    requests:
      storage: 10Gi
```

#### 3) Pod mounting the PVC
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-data
spec:
  containers:
  - name: app
    image: busybox
    command: ["sh","-c","echo 'hello' >> /data/hello.txt && sleep 3600"]
    volumeMounts:
    - name: app-data
      mountPath: /data
  volumes:
  - name: app-data
    persistentVolumeClaim:
      claimName: data-pvc
```

### kubectl Commands
- **Create and watch:** `kubectl apply -f pvc.yaml && kubectl get pvc`
- **Inspect volumes:** `kubectl get pv && kubectl describe pv <name>`
- **Troubleshoot:** `kubectl describe pvc data-pvc`

---

## 7. StatefulSet

### What it is
A controller for stateful workloads (databases, queues) that need:

- **Stable network IDs:** pods named db-0, db-1, …
- **Sticky storage:** each replica keeps its own volume across restarts
- **Ordered rollout & scaling:** create/terminate in ordinal order

### Key pieces
- **Headless Service** (`clusterIP: None`) gives per‑pod DNS: `db-0.db-svc.default.svc.cluster.local`
- **volumeClaimTemplates:** auto‑create one PVC per pod (e.g., data-db-0)
- **Update strategies:** RollingUpdate (default) with ordered, one‑by‑one replacement

### When to use
- Primary‑replica databases (Postgres, MySQL), Kafka, ZooKeeper, etc.
- Anything that needs identity + durable, per‑replica data

### Example: StatefulSet with per‑replica volumes
```yaml
apiVersion: v1
kind: Service
metadata:
  name: db-svc
spec:
  clusterIP: None            # headless
  selector:
    app: db
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: db
spec:
  serviceName: db-svc        # ties to headless svc
  replicas: 3
  selector:
    matchLabels:
      app: db
  template:
    metadata:
      labels:
        app: db
    spec:
      containers:
      - name: postgres
        image: postgres:16
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: data
          mountPath: /var/lib/postgresql/data
        env:
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: pg-secret
              key: password
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 10Gi
      storageClassName: standard
```

### kubectl Commands
- **Create & watch:** `kubectl apply -f statefulset.yaml && kubectl rollout status sts/db`
- **Scale:** `kubectl scale sts/db --replicas=5`
- **Inspect PVCs:** `kubectl get pvc -l app=db`
- **Per‑pod exec:** `kubectl exec -it db-0 -- psql -U postgres`

---

## 8. Health Probes

### Why they matter
Kubernetes needs to know:
- Is the process alive? (restart it if deadlocked) → **livenessProbe**
- Is it ready to serve traffic? (keep it out of Service endpoints until ready) → **readinessProbe**
- Should we wait longer before judging liveness/readiness? → **startupProbe**

### The three probes
- **readinessProbe:** gates load-balancing. Not ready ⇒ the pod stays out of the Service
- **livenessProbe:** restarts the container if it keeps failing
- **startupProbe:** one-time grace period while the app boots; while this runs, liveness/readiness are suppressed

### Probe types
- **httpGet** (path/port, optional headers)
- **exec** (run a command, exit code 0 = success)
- **tcpSocket** (open succeeds = success)

**Timing knobs:** initialDelaySeconds, periodSeconds, timeoutSeconds, failureThreshold, successThreshold

### Example: Web app with all three probes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 2
  selector:
    matchLabels: { app: web }
  template:
    metadata:
      labels: { app: web }
    spec:
      containers:
      - name: web
        image: ghcr.io/example/web:1.0
        ports:
        - containerPort: 8080
        startupProbe:
          httpGet: { path: /healthz/startup, port: 8080 }
          failureThreshold: 30         # up to 30 * 1s = 30s to start
          periodSeconds: 1
        readinessProbe:
          httpGet: { path: /healthz/ready, port: 8080 }
          initialDelaySeconds: 5
          periodSeconds: 5
          failureThreshold: 2          # 2 consecutive fails => NotReady
        livenessProbe:
          httpGet: { path: /healthz/live, port: 8080 }
          initialDelaySeconds: 10
          periodSeconds: 10
          failureThreshold: 3          # 3 consecutive fails => restart
```

### Debugging probes
- `kubectl describe pod <pod>` → Events show probe failures
- `kubectl get endpoints <svc>` → see if your pod is included (readiness)
- Tune initialDelaySeconds and timeoutSeconds if slow boot or heavy GC pauses

---

## 9. ConfigMap & Secret

### Why they exist
- **ConfigMap:** injects non-sensitive config (env vars, files)
- **Secret:** same idea, but for sensitive data (passwords, keys, certs). Values are base64-encoded

This keeps config separate from code/images, making apps portable and 12-factor compliant.

### Mounting options
- As environment variables (`envFrom` or `valueFrom`)
- As files via a volume mount (kubelet writes keys as filenames, values as contents)

### Example: ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  APP_ENV: "prod"
  APP_DEBUG: "false"
  WELCOME_MSG: "Hello from ConfigMap"
```

### Example: Secret
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-secret
type: Opaque
data:
  username: YWRtaW4=        # "admin"
  password: cGFzc3dvcmQ=    # "password"
```

### Using them in a Pod
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo
spec:
  containers:
  - name: app
    image: busybox
    command: ["sh","-c","echo $APP_ENV && sleep 3600"]
    envFrom:
    - configMapRef:
        name: app-config
    - secretRef:
        name: db-secret
```

### kubectl Commands
- **Create directly:**
  - `kubectl create configmap app-config --from-literal=APP_ENV=prod`
  - `kubectl create secret generic db-secret --from-literal=username=admin`
- **Inspect:** `kubectl get cm,secret && kubectl describe cm app-config`
- **Decode secret:** `kubectl get secret db-secret -o jsonpath="{.data.password}" | base64 -d`

---

## 10. DaemonSet

### What it is
A controller that ensures one pod per node (or per matching node).

Used for node-level agents (logging, monitoring, networking).

**Examples:**
- kube-proxy (runs everywhere)
- CNI plugins (Calico, Flannel)
- Monitoring (Node Exporter, Fluentd, Prometheus node-agent)

### Why it exists
Sometimes you don't want a "pool of N replicas" (Deployment/ReplicaSet). Instead, you need every node to have exactly one copy (or a subset). DaemonSet handles this.

### Features
- By default, 1 pod per schedulable node
- With nodeSelector / tolerations, you can restrict to certain nodes (e.g., GPU nodes)
- When a new node joins, the DaemonSet automatically schedules a pod on it

### Example YAML
```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: log-agent
spec:
  selector:
    matchLabels:
      app: log-agent
  template:
    metadata:
      labels:
        app: log-agent
    spec:
      containers:
      - name: fluentd
        image: fluent/fluentd:v1.14
        resources:
          limits:
            cpu: "100m"
            memory: "200Mi"
```

### kubectl Commands
- **Create:** `kubectl apply -f daemonset.yaml`
- **Check:** `kubectl get ds`
- **See pods:** `kubectl get pods -l app=log-agent -o wide`
- **Delete:** `kubectl delete ds log-agent`

---

## Directory Structure
Your K8s learning materials are organized as follows:

```
01-pods/                   # Basic pod examples
02-replicasets/            # ReplicaSet configurations  
03-deployments/            # Deployment examples
04-services/               # Service definitions
05-ingress/                # Ingress routing examples
06-persistent-storage/     # PV, PVC, StorageClass examples
07-statefulsets/           # StatefulSet configurations
08-health-probes/          # Liveness, readiness, startup probes
09-configmaps-secrets/     # ConfigMap and Secret examples
10-daemonsets/             # DaemonSet configurations
```

Each directory contains practical YAML examples and configurations for hands-on learning.