@echo off
setlocal
pushd "%~dp0..\.."
go run ./tools/passhash %*
popd
pause
