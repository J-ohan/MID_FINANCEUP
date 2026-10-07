@echo off
REM ============================================================
REM  FINANCEUP - CREA LA BASE DE DATOS LOCAL DEL PROYECTO
REM  Se ejecuta UNA SOLA VEZ (doble clic).
REM  - Crea la carpeta "datos" donde PostgreSQL guarda la informacion.
REM  - La deja en el puerto 5433, solo para este equipo y sin contrasena.
REM  - Crea la base "financeup" con sus tablas y datos de prueba.
REM ============================================================
setlocal
set "PG_BIN=C:\Program Files\PostgreSQL\18\bin"
set "AQUI=%~dp0"
set "DATOS=%AQUI%datos"
set "SQL=%AQUI%..\sql"

if exist "%DATOS%\PG_VERSION" (
    echo La base local ya existe. Si quieres empezar de cero, borra la carpeta "datos" y vuelve a ejecutar este archivo.
    pause
    exit /b 0
)

echo [1/4] Creando la carpeta de datos...
"%PG_BIN%\initdb.exe" -D "%DATOS%" -U postgres -A trust -E UTF8 --no-locale >nul || goto error

echo [2/4] Configurando puerto 5433 y acceso solo local...
>> "%DATOS%\postgresql.conf" echo port = 5433
>> "%DATOS%\postgresql.conf" echo listen_addresses = 'localhost'

echo [3/4] Encendiendo la base de datos...
"%PG_BIN%\pg_ctl.exe" -D "%DATOS%" -l "%AQUI%postgres.log" -w start >nul || goto error

echo [4/4] Creando la base financeup, sus tablas y datos de prueba...
set PGCLIENTENCODING=UTF8
"%PG_BIN%\psql.exe" -h localhost -p 5433 -U postgres -d postgres -q -v ON_ERROR_STOP=1 -f "%SQL%\01_crear_base_datos.sql" || goto error
"%PG_BIN%\psql.exe" -h localhost -p 5433 -U postgres -d financeup -q -v ON_ERROR_STOP=1 -f "%SQL%\02_crear_tablas.sql" || goto error
"%PG_BIN%\psql.exe" -h localhost -p 5433 -U postgres -d financeup -q -v ON_ERROR_STOP=1 -f "%SQL%\03_insertar_datos.sql" || goto error

echo.
echo Listo. La base "financeup" quedo encendida en localhost:5433 (usuario postgres, sin contrasena).
pause
exit /b 0

:error
echo.
echo Ocurrio un error. Revisa el mensaje de arriba o el archivo postgres.log
pause
exit /b 1
