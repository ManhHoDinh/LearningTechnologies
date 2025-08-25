#!/bin/bash

echo "Starting port forwarding for all services..."
echo

# Function to cleanup background processes
cleanup() {
    echo
    echo "Stopping all port forwarding..."
    jobs -p | xargs -r kill
    exit 0
}

# Set trap to cleanup on script exit
trap cleanup SIGINT SIGTERM

# Start port forwarding for each service in background
echo "Starting API service on http://localhost:8081"
kubectl port-forward service/api-svc 8081:80 &

echo "Starting Nginx service on http://localhost:8082"
kubectl port-forward service/nginx-svc 8082:80 &

echo "Starting Web service on http://localhost:8083"
kubectl port-forward service/web-svc 8083:80 &

echo "Starting Shop service on http://localhost:8084"
kubectl port-forward service/shop-svc 8084:80 &

echo "Starting Secure service on https://localhost:8443"
kubectl port-forward service/secure-svc 8443:443 &

echo
echo "All services are now accessible:"
echo "- API Service:    http://localhost:8081"
echo "- Nginx Service:  http://localhost:8082"
echo "- Web Service:    http://localhost:8083"
echo "- Shop Service:   http://localhost:8084"
echo "- Secure Service: https://localhost:8443"
echo
echo "Press Ctrl+C to stop all port forwarding..."

# Wait for all background jobs
wait