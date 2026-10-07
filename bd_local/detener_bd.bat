@echo off
REM ============================================================
REM  FINANCEUP - APAGA LA BASE DE DATOS LOCAL
REM ============================================================
set "PG_BIN=C:\Program Files\PostgreSQL\18\bin"
"%PG_BIN%\pg_ctl.exe" -D "%~dp0datos" -w stop >nul && echo Base de datos apagada.
pause
