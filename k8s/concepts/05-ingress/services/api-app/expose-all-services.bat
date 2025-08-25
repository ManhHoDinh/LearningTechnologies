@echo off
echo Starting port forwarding for all services...
echo.

REM Start port forwarding for each service in background
echo Starting API service on http://localhost:8081
start "API Service" cmd /c "kubectl port-forward service/api-svc 8081:80"

echo Starting Nginx service on http://localhost:8082
start "Nginx Service" cmd /c "kubectl port-forward service/nginx-svc 8082:80"

echo Starting Web service on http://localhost:8083
start "Web Service" cmd /c "kubectl port-forward service/web-svc 8083:80"

echo Starting Shop service on http://localhost:8084
start "Shop Service" cmd /c "kubectl port-forward service/shop-svc 8084:80"

echo Starting Secure service on https://localhost:8443
start "Secure Service" cmd /c "kubectl port-forward service/secure-svc 8443:443"

echo.
echo All services are being exposed:
echo - API Service:    http://localhost:8081
echo - Nginx Service:  http://localhost:8082
echo - Web Service:    http://localhost:8083
echo - Shop Service:   http://localhost:8084
echo - Secure Service: https://localhost:8443
echo.
echo Press any key to stop all port forwarding...
pause > nul

echo Stopping all port forwarding...
taskkill /f /im kubectl.exe > nul 2>&1
echo All services stopped.