@echo off
echo Generating a new Ed25519 keypair...
echo.

go run tools/keygen/keygen.go

echo.
pause
