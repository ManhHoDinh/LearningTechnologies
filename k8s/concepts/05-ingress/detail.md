# Concept 5 — Ingress (Hands‑on Guide)

> **Goal:** Expose multiple Services over HTTP/HTTPS through a single entrypoint using Kubernetes **Ingress**.

---

## What is an Ingress?

An **Ingress** is a Kubernetes API object that manages **external HTTP/HTTPS** access to Services. Think of it as a **smart router / reverse proxy** living **inside** your cluster.

**It can handle:**
- **Path‑based routing** (e.g., `/api → backend`, `/app → frontend`)
- **Host‑based routing** (e.g., `api.example.com` vs `shop.example.com`)
- **TLS/SSL termination** (HTTPS at the edge)

**Why it exists**
- `Service type: LoadBalancer` → **one external IP per Service** → expensive and messy.
- **Ingress** lets you expose **many** Services behind **one** load balancer, with flexible routing.

> ⚠️ **Important:** An Ingress object is **just the rules**. You **must** have an **Ingress Controller** (e.g., NGINX Ingress, Traefik, Istio gateway) installed. The controller enforces the rules.

---

## Prerequisites
- A running cluster (Minikube, kind, k3d, or managed K8s)
- `kubectl` configured to talk to the cluster
- Services (e.g., `api-svc`, `web-svc`, `shop-svc`, `nginx-svc`, `secure-svc`) already created
- **Ingress Controller installed** (see: [Install a controller](#step-1--install-an-ingress-controller))

---

## Quick Reference: kubectl
- **Create/Update:** `kubectl apply -f <file>.yaml`
- **List:** `kubectl get ingress`
- **Describe:** `kubectl describe ingress <name>`
- **Watch status:** `kubectl get ingress -w`

---

## Minimal Example — Basic Ingress
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
**Behavior:** Requests to `http://example.com/` go to **nginx-svc:80**.

---

## Patterns

### 📝 Example 1 — Host‑based routing
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: host-routing
spec:
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-svc
            port:
              number: 80
  - host: web.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web-svc
            port:
              number: 80
```
**Result:** `api.example.com` → `api-svc`, `web.example.com` → `web-svc`.

### 📝 Example 2 — Path‑based routing
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: path-routing
spec:
  rules:
  - host: example.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: api-svc
            port:
              number: 80
      - path: /shop
        pathType: Prefix
        backend:
          service:
            name: shop-svc
            port:
              number: 80
```
**Result:** `/api/*` → `api-svc`, `/shop/*` → `shop-svc`.

### 📝 Example 3 — TLS termination (HTTPS)
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: secure-ingress
spec:
  tls:
  - hosts:
    - secure.example.com
    secretName: tls-secret
  rules:
  - host: secure.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: secure-svc
            port:
              number: 443
```
**Result:** HTTPS for `secure.example.com` is terminated at the controller using `tls-secret`.

### 📝 Example 4 — Default backend (catch‑all)
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: default-backend
spec:
  defaultBackend:
    service:
      name: catchall-svc
      port:
        number: 80
```
**Result:** Any request that doesn’t match a host/path rule goes to `catchall-svc`.

---

## Local Development Options

### A) Keep port‑forward, add custom hostnames (quick & dirty)
1) Add entries to your hosts file pointing to **localhost**:
   - **Windows:** `C:\\Windows\\System32\\drivers\\etc\\hosts`
   - **macOS/Linux:** `/etc/hosts`

   ```
   127.0.0.1 api.local web.local nginx.local shop.local secure.local
   ```

2) Keep your port‑forward script as‑is. Now browse:
   - `http://api.local:8081`
   - `http://nginx.local:8082`
   - `http://web.local:8083`
   - `http://shop.local:8084`
   - `https://secure.local:8443` *(you’ll see a cert warning unless you trust a local CA)*

> Pros: zero cluster changes. Cons: still need ports in URLs; not true Ingress.

### B) Use an Ingress controller (best: :80/:443 like production)

#### Step 1 — Install an Ingress controller
- **Minikube:**
  ```bash
  minikube addons enable ingress
  ```
- **kind / other:** Install **NGINX Ingress Controller** (Helm or manifests). Example (Helm):
  ```bash
  helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
  helm repo update
  helm install ingress-nginx ingress-nginx/ingress-nginx \
    --create-namespace --namespace ingress-nginx
  ```

#### Step 2 — One Ingress with multiple hosts
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: apps-ingress
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "false"   # remove once you add TLS
spec:
  rules:
  - host: api.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service: { name: api-svc, port: { number: 80 } }
  - host: nginx.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service: { name: nginx-svc, port: { number: 80 } }
  - host: web.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service: { name: web-svc, port: { number: 80 } }
  - host: shop.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service: { name: shop-svc, port: { number: 80 } }
```
Apply it:
```bash
kubectl apply -f apps-ingress.yaml
```

#### Step 3 — Point your custom domains to the cluster IP
- **Minikube:**
  ```bash
  minikube ip   # e.g., 192.168.49.2
  ```
- Add to your hosts file:
  ```
  192.168.49.2 api.local nginx.local web.local shop.local secure.local
  ```
- Browse with no ports: `http://api.local`, `http://web.local`, etc.

#### Optional Step 4 — Add trusted HTTPS (local dev)
1) Generate local certs with **mkcert**
   ```bash
   # macOS
   brew install mkcert nss
   mkcert -install

   # Windows (Chocolatey)
   choco install mkcert
   mkcert -install

   # Then create a cert for your hosts
   mkcert api.local nginx.local web.local shop.local secure.local
   ```
2) Create a TLS secret from the generated files
   ```bash
   kubectl create secret tls local-tls \
     --cert=./api.local+nginx.local+web.local+shop.local+secure.local.pem \
     --key=./api.local+nginx.local+web.local+shop.local+secure.local-key.pem
   ```
3) Update your Ingress to use TLS & **remove** the `ssl-redirect: "false"` annotation so HTTP redirects to HTTPS
```yaml
spec:
  tls:
  - hosts: [api.local, nginx.local, web.local, shop.local, secure.local]
    secretName: local-tls
  rules:
    # (same rules as before)
```

---

## Verify & Debug
- **DNS/hosts resolution**
  ```bash
  ping api.local
  nslookup api.local   # should resolve to your cluster LB / minikube IP
  ```
- **Ingress status & address**
  ```bash
  kubectl get ingress -A -o wide
  kubectl describe ingress apps-ingress
  ```
- **Controller pods & logs** (NGINX example)
  ```bash
  kubectl get pods -n ingress-nginx
  kubectl logs -n ingress-nginx deploy/ingress-nginx-controller
  ```
- **Functional checks**
  ```bash
  curl -i http://api.local/
  curl -kI https://api.local/   # -k for self-signed
  ```

---

## Common Annotations (NGINX Ingress)
- `nginx.ingress.kubernetes.io/rewrite-target: /`
- `nginx.ingress.kubernetes.io/ssl-redirect: "true|false"`
- `nginx.ingress.kubernetes.io/proxy-body-size: 10m`
- `nginx.ingress.kubernetes.io/limit-rps: "5"`
- `nginx.ingress.kubernetes.io/enable-cors: "true"`

> Check your controller’s docs for supported annotations and CRDs.

---

## Troubleshooting
- **404 from Ingress**: path or host mismatch; verify `rules` and `pathType`.
- **502/504**: backend Service/Pod not reachable; check `Service` selectors and Pod readiness.
- **No external IP/Address**: your environment may use NodePort; consult controller installation notes.
- **TLS not working**: secret name mismatch, wrong CN/SANs in cert, or `tls.hosts` not listing the host.

---

## Cleanup
```bash
kubectl delete ingress nginx-ingress host-routing path-routing secure-ingress default-backend apps-ingress
# If installed via Helm
helm uninstall ingress-nginx -n ingress-nginx
```

---

### Key Takeaway
Use **Ingress** to consolidate traffic through a single, flexible edge with **host/path routing** and **TLS**—much closer to production than port‑forwards and multiple load balancers.

