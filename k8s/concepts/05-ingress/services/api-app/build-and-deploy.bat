@echo off
echo Building Docker images for minikube...

REM Set minikube docker environment
FOR /f "tokens=*" %%i IN ('minikube docker-env --shell cmd') DO %%i

REM Build all Docker images in minikube's Docker daemon
docker build -t nginx-app:latest ./nginx-app/
docker build -t api-app:latest ./api-app/
docker build -t web-app:latest ./web-app/
docker build -t shop-app:latest ./shop-app/
docker build -t secure-app:latest ./secure-app/

echo Images built in minikube Docker daemon
echo Deploying to Kubernetes...

REM Apply all service manifests
kubectl apply -f nginx-service.yaml
kubectl apply -f api-service.yaml
kubectl apply -f web-service.yaml
kubectl apply -f shop-service.yaml
kubectl apply -f secure-service.yaml

echo All services deployed!
echo You can now apply your ingress configurations:
echo kubectl apply -f ../nginx-ingress.yaml
echo kubectl apply -f ../host-routing.yaml
echo kubectl apply -f ../path-routing.yaml
echo kubectl apply -f ../secure-ingress.yaml