-- ============================================================
-- FINANCEUP - CREACION DE LA BASE DE DATOS
-- Ejecutar conectado a la base "postgres" (la que viene por defecto).
-- Despues de crearla, abrir una ventana de consultas sobre "financeup"
-- y ejecutar 02_crear_tablas.sql y luego 03_insertar_datos.sql.
-- ============================================================

CREATE DATABASE financeup
    WITH ENCODING = 'UTF8'
    TEMPLATE = template0;

-- Todas las fechas se guardan en hora UTC (hora universal).
-- Los CRUD en Go tambien trabajan en UTC, asi ninguna fecha se "corre" de hora.
-- Colombia es UTC-5: una fecha 15:00 UTC equivale a las 10:00 en Colombia.
ALTER DATABASE financeup SET timezone = 'UTC';
