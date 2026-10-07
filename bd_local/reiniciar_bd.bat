@echo off
REM ============================================================
REM  FINANCEUP - BORRA Y VUELVE A CREAR LA BASE "financeup"
REM  Util si se da??aron los datos de prueba. Se pierde lo que hayas guardado.
REM  La base local debe estar encendida (iniciar_bd.bat).
REM ============================================================
set "PG_BIN=C:\Program Files\PostgreSQL\18\bin"
set "SQL=%~dp0..\sql"
set PGCLIENTENCODING=UTF8
"%PG_BIN%\psql.exe" -h localhost -p 5433 -U postgres -d postgres -q -c "DROP DATABASE IF EXISTS financeup WITH (FORCE);" || goto error
"%PG_BIN%\psql.exe" -h localhost -p 5433 -U postgres -d postgres -q -v ON_ERROR_STOP=1 -f "%SQL%\01_crear_base_datos.sql" || goto error
"%PG_BIN%\psql.exe" -h localhost -p 5433 -U postgres -d financeup -q -v ON_ERROR_STOP=1 -f "%SQL%\02_crear_tablas.sql" || goto error
"%PG_BIN%\psql.exe" -h localhost -p 5433 -U postgres -d financeup -q -v ON_ERROR_STOP=1 -f "%SQL%\03_insertar_datos.sql" || goto error
echo Base "financeup" creada de nuevo con los datos de prueba.
pause
exit /b 0
:error
echo Ocurrio un error. Verifica que la base este encendida con iniciar_bd.bat
pause
exit /b 1
