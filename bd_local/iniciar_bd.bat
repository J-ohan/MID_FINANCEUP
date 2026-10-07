@echo off
REM ============================================================
REM  FINANCEUP - ENCIENDE LA BASE DE DATOS LOCAL
REM  Doble clic cada vez que prendas el PC, antes de usar los CRUD.
REM ============================================================
set "PG_BIN=C:\Program Files\PostgreSQL\18\bin"
set "DATOS=%~dp0datos"

if not exist "%DATOS%\PG_VERSION" (
    echo La base local aun no existe. Ejecuta primero crear_bd.bat
    pause
    exit /b 1
)

"%PG_BIN%\pg_ctl.exe" -D "%DATOS%" status >nul 2>&1
if %errorlevel%==0 (
    echo La base de datos ya estaba encendida en localhost:5433
) else (
    "%PG_BIN%\pg_ctl.exe" -D "%DATOS%" -l "%~dp0postgres.log" -w start >nul && echo Base de datos encendida en localhost:5433
)
pause
