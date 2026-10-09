-- ============================================================
-- FINANCEUP - CREACION DE LA BASE DE DATOS
-- Ejecutar conectado a la base "postgres" (la que viene por defecto).
--
-- IMPORTANTE (pgAdmin): ejecuta estas DOS instrucciones POR SEPARADO.
-- PostgreSQL no permite CREATE DATABASE junto con otra instruccion.
--   1) Selecciona solo el bloque CREATE DATABASE y presiona F5.
--   2) Selecciona solo la linea ALTER DATABASE y presiona F5.
--
-- Despues abre una ventana de consultas sobre "financeup" y ejecuta
-- 02_crear_tablas.sql y luego 03_insertar_datos.sql (esos si se
-- ejecutan completos de una vez).
-- ============================================================

-- PASO 1: crear la base
CREATE DATABASE financeup
    WITH ENCODING = 'UTF8'
    TEMPLATE = template0;

-- PASO 2: todas las fechas se guardan en hora UTC (hora universal).
-- Los CRUD en Go tambien trabajan en UTC, asi ninguna fecha se "corre" de hora.
-- Colombia es UTC-5: una fecha 15:00 UTC equivale a las 10:00 en Colombia.
ALTER DATABASE financeup SET timezone = 'UTC';
