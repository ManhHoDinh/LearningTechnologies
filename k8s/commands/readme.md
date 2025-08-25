# 📘 Kubernetes Command Guide

## 📍 Step 1 – Core Concepts

Before touching YAML and clusters, understand these key building blocks:

- **Cluster** = a set of machines (nodes) managed by Kubernetes.  
- **Node** = one machine (VM or physical server).  
- **Pod** = the smallest deployable unit (usually wraps 1 container).  
- **Deployment** = manages Pods, keeps the right number running.  
- **Service** = gives Pods a stable IP/DNS, load balancing.  
- **ConfigMap / Secret** = externalized configuration.  
- **Ingress** = HTTP/HTTPS routing to your services.  

👉 **Mini-exercise:** Draw this picture:  
`Client → Ingress → Service → Pod (container)`

---

## 🧪 Module 0 – Sanity Checks

### 1. `kubectl version`
- **What it does:** Shows client and server versions of Kubernetes.  
- **Why:** Confirms cluster connectivity, spots mismatches.  
- **Example:**  
  ```bash
  kubectl version --short
  ```  
- **Gotcha:** If only client shows and server is blank → your cluster/context isn’t reachable.

---

### 2. `kubectl config current-context`
- **What it does:** Shows which cluster+user combo you’re targeting.  
- **Why:** Prevents accidental changes in the wrong cluster.  
- **Example:**  
  ```bash
  kubectl config current-context
  ```  
- **Tip:** Switch with `kubectl config use-context <name>`.

---

### 3. `kubectl cluster-info`
- **What it does:** Prints API server and core service endpoints.  
- **Why:** Quick connectivity check.  
- **Example:**  
  ```bash
  kubectl cluster-info
  ```

---

### 4. `kubectl get nodes`
- **What it does:** Lists worker/master nodes.  
- **Why:** Confirms cluster health at the node level.  
- **Example:**  
  ```bash
  kubectl get nodes -o wide
  ```  
- **Read:** STATUS should be **Ready**.

---

## 🧪 Module 1 – Pods & Debugging

### 5. `kubectl run`
```bash
kubectl run hello-pod --image=nginx --restart=Never
```
- Quick pod for demos/tests.  
- `--image=nginx` → container image.  
- `--restart=Never` → ensures it’s a **Pod**, not a Deployment.

---

### 6. `kubectl get pods`
```bash
kubectl get pods
kubectl get pods -n kube-system
```
- Lists pods in a namespace.  
- **STATUS** column shows health: Running, Pending, CrashLoopBackOff…

---

### 7. `kubectl describe pod <pod-name>`
- Full details: containers, events, errors.  
- **Tip:** Bottom “Events” section is gold for debugging.

---

### 8. `kubectl logs <pod-name>`
```bash
kubectl logs myapp-1234
kubectl logs myapp-1234 -c sidecar
kubectl logs -f myapp-1234   # follow logs
```

---

### 9. `kubectl exec -it <pod-name> -- <command>`
```bash
kubectl exec -it myapp-1234 -- sh
```
- Opens shell inside a pod.  
- Useful for debugging containers.

---

## 🧪 Module 2 – Deployments & Services

### 10. `kubectl apply -f <file.yaml>`
- Creates/updates objects from YAML (GitOps style).  
- Example **nginx-deploy.yaml**:
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
- Apply:  
  ```bash
  kubectl apply -f nginx-deploy.yaml
  ```

---

### 11. `kubectl get all`
- Lists all resources in namespace.  
- ```bash
  kubectl get all -n default
  ```

---

### 12. `kubectl delete -f <file.yaml>`
- Removes resources from YAML (safer than deleting by name).

---

### 13. `kubectl scale`
- ```bash
  kubectl scale deployment/nginx-deploy --replicas=4
  ```
- Scales replicas quickly.

---

### 14. `kubectl rollout status`
```bash
    kubectl rollout status deployment/nginx-deploy
```
- Watches Deployment progress.

---

### 15. `kubectl set image`
```bash
kubectl set image deploy/nginx-deploy nginx=nginx:1.27
```
- Updates image → triggers rollout.

---

### 16. `kubectl rollout history` & `undo`
```bash
kubectl rollout history deploy/nginx-deploy
kubectl rollout undo deploy/nginx-deploy --to-revision=2
```

---

### 17. `kubectl expose`
```bash
kubectl expose deployment/nginx-deploy --port=80 --type=NodePort
```
- Creates a Service for network access.  
- Types: ClusterIP / NodePort / LoadBalancer.

---

### 18. `kubectl port-forward`
```bash
kubectl port-forward deployment/nginx-deploy 8080:80
```
- Local testing via `http://localhost:8080`.

---

### 19. `kubectl get svc`
- Lists Services with name, type, IP, ports.

---

## 🧪 Module 3 – Debugging & Monitoring

### 20. `kubectl get events`
```bash
kubectl get events --sort-by=.metadata.creationTimestamp
```
- Shows cluster events (pods failing, scheduled, restarted).

---

### 21. `kubectl top`
```bash
kubectl top nodes
kubectl top pods -n kube-system
```
- Resource usage (requires metrics-server).

---

### 22. `kubectl explain`
```bash
kubectl explain pods
kubectl explain deployment.spec.template.spec.containers
```
- Built-in manual for YAML fields.

---

## 🧪 Module 4 – Namespaces & Access

- `kubectl get ns` → list namespaces.  
- `kubectl config get-contexts` → list contexts.  
- `kubectl config use-context <name>` → switch clusters.  
- `kubectl auth can-i create pods -n default` → check permissions.

---

## 🧪 Module 5 – Config & Jobs

- `kubectl create configmap app-config --from-literal=ENV=prod`  
- `kubectl create secret generic db-pass --from-literal=password=MyP@ssw0rd`  
- `kubectl get configmap,secret`  
- `kubectl create job pi --image=perl -- perl -Mbignum=bpi -wle 'print bpi(2000)'`  
- `kubectl create cronjob hello --image=busybox --schedule="*/5 * * * *" -- echo hello`  

---

## 🧪 Module 6 – Node Management

- `kubectl get all --all-namespaces` → cluster-wide view.  
- `kubectl describe node <node-name>` → details.  
- `kubectl drain <node-name>` → safe maintenance.  
- `kubectl cordon <node-name>` → unschedulable.  
- `kubectl uncordon <node-name>` → schedulable again.  

---

✅ You now know **36 essential kubectl commands** grouped by use-case:  
- Cluster checks  
- Pods & debugging  
- Deployments & rollouts  
- Services & networking  
- Events & monitoring  
- Namespaces & RBAC  
- Config & Jobs  
- Node ops  
