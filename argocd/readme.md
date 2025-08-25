# Argo CD on Minikube (Windows) — Quick Guide

This guide gets you from zero to GitOps with Argo CD on **Minikube (Windows)**. You’ll install Argo CD, open the UI, deploy a sample app, turn on auto‑sync, practice diffs/rollbacks, and scale to many services with ApplicationSet.

> **Prereqs**
> - Windows PowerShell (run as Administrator when needed)
> - `kubectl` and `minikube` installed
> - A running Minikube cluster:  
>   ```powershell
>   minikube start
>   ```

---

## Step 1 — Install Argo CD (Minikube)

1) **Create namespace**
```powershell
kubectl create namespace argocd
```

2) **Install core Argo CD components**
```powershell
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

Check pods:
```powershell
kubectl -n argocd get pods
```

---

## Step 2 — Open the UI (quickest way)

Port-forward the Argo CD server:
```powershell
kubectl -n argocd port-forward svc/argocd-server 8080:443
```

Open http://localhost:8080

Get the initial admin password in **PowerShell**:
```powershell
$secret = kubectl get secret argocd-initial-admin-secret -n argocd -o jsonpath="{.data.password}"
[System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($secret))
```

Login (replace with the password you just printed):
```powershell
argocd login localhost:8080 --username admin --password <PASTE_PASSWORD_HERE> --insecure
```

---

## Step 3 — Create & sync a sample app (manual sync first)

Create a file named **`guestbook-app.yaml`** with this content:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: guestbook
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps
    targetRevision: HEAD
    path: helm-guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: demo
  syncPolicy:
    syncOptions:
      - CreateNamespace=true
```

Apply it:
```powershell
kubectl apply -f guestbook-app.yaml
```

In the Argo CD UI, open **guestbook** → status should be **OutOfSync** → click **SYNC**.

Verify on the cluster:
```powershell
kubectl -n demo get deploy,svc,pods
kubectl -n demo port-forward svc/helm-guestbook 9000:80
# then browse http://localhost:9000
```

---

## Step 4 — Auto-sync, Diffs, Rollbacks

In real life you change **Git** and Argo applies it. For practice, we’ll tweak Helm values in the **Application** to simulate a commit.

### A) Turn on Auto-sync (two easy ways)

**Option 1 — CLI (quickest)**
```powershell
argocd app set guestbook --sync-policy automated --self-heal --auto-prune
argocd app get guestbook
```
- **automated**: sync when Git (or the Application spec) changes
- **selfHeal**: revert out-of-band cluster edits
- **auto-prune**: delete resources removed from Git

**Option 2 — YAML (if you keep Applications in Git)**  
Edit `guestbook-app.yaml` and re-apply:
```yaml
spec:
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```
```powershell
kubectl apply -f guestbook-app.yaml
```

### B) See diffs before/after sync
```powershell
argocd app diff guestbook
# Also handy
argocd app history guestbook
```

**Quick experiment (no repo changes needed):**  
Increase replicas via Helm values stored in the Application:
```powershell
argocd app set guestbook -p replicaCount=2
argocd app diff guestbook
argocd app get guestbook
```
Because **Auto-sync** is on, Argo should apply it and your Deployment should scale to **2**.

Check:
```powershell
kubectl -n demo get deploy/helm-guestbook
```

### C) Roll back
See revisions:
```powershell
argocd app history guestbook
```
Pick an **ID** from the list:
```powershell
argocd app rollback guestbook <ID>
```

### D) Health troubleshooting (when status says “Progressing”)
```powershell
kubectl -n demo rollout status deploy/helm-guestbook
kubectl -n demo get endpoints helm-guestbook -o wide
kubectl -n demo get events --sort-by=.lastTimestamp | Select-Object -Last 20
```
- **Empty endpoints** → Service selector or `targetPort` mismatch  
- **Probe failures** → fix readiness/liveness path/port

**Mini-review (sticky ideas)**
- Auto-sync keeps cluster = desired state automatically.
- Self-heal undoes manual drifts.
- Prune removes deleted objects.
- Diff/History/Rollback = your safe time machine.

---

## Step 5 — Many services via ApplicationSet (hands-on)

### A) Make sure the controller is running
```powershell
kubectl -n argocd get deploy argocd-applicationset-controller
```
If you see a Deployment and it’s **Available**, you’re good.

### B) Create an ApplicationSet that generates multiple apps

We’ll build two apps from the public `argocd-example-apps` repo—one Helm, one Kustomize.

Save as **`apps-demo.yaml`**:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: demo-services
  namespace: argocd
spec:
  generators:
    - list:
        elements:
          - name: helm-guestbook
            path: helm-guestbook
            namespace: svc-helm-guestbook
          - name: kustomize-guestbook
            path: kustomize-guestbook
            namespace: svc-kustomize-guestbook
  template:
    metadata:
      name: '{{name}}'
    spec:
      project: default
      source:
        repoURL: https://github.com/argoproj/argocd-example-apps
        targetRevision: HEAD
        path: '{{path}}'
      destination:
        server: https://kubernetes.default.svc
        namespace: '{{namespace}}'
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
        syncOptions:
          - CreateNamespace=true
```

Apply it:
```powershell
kubectl apply -f apps-demo.yaml
```

Open Argo CD UI → you should see **helm-guestbook** and **kustomize-guestbook** apps created automatically and syncing.

### C) Verify on the cluster
```powershell
kubectl get ns | findstr svc-
kubectl -n svc-helm-guestbook get deploy,svc,pods
kubectl -n svc-kustomize-guestbook get deploy,svc,pods
```

**Why this scales**  
You don’t hand‑craft one Application per service; the generator feeds a template to produce many Applications. Later, switch to a `git.directories` generator to create **one app per `services/*` folder** in your repo, or use a **matrix** to do per‑service × per‑environment.

---

## (Optional) Ingress for services on Minikube
Enable the addon once (Admin PowerShell):
```powershell
minikube addons enable ingress
```
Map a dev host to your Minikube IP (find it with `minikube ip`, e.g., `192.168.49.2`) by editing `C:\Windows\System32\drivers\etc\hosts`:
```
192.168.49.2  post-service-dev.local
```
Then access your Ingress using that host header.

---

## Cheat Sheet
```powershell
# App basics
argocd app list
argocd app get <app>
argocd app sync <app>
argocd app diff <app>
argocd app history <app>
argocd app rollback <app> <ID>

# Health & K8s
kubectl -n <ns> rollout status deploy/<name>
kubectl -n <ns> get endpoints <svc> -o wide
kubectl -n <ns> get events --sort-by=.lastTimestamp | Select-Object -Last 20
```

---

### What to try next
- Convert the sample to **App-of-Apps** for an environment bootstrap.
- Add **sync waves** and a **PreSync** migration Job for a real app.
- Add `ignoreDifferences` for `/spec/replicas` if you use an HPA.
