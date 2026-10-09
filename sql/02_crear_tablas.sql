-- ============================================================
-- FINANCEUP - CREACION DE TABLAS
-- Ejecutar conectado a la base de datos "financeup".
-- Cada schema corresponde a un CRUD:
--   auth      -> crud_auth      (puerto 8081)
--   finanzas  -> crud_finanzas  (puerto 8082)
--   educacion -> crud_educacion (puerto 8083)
--   negocio   -> crud_negocio   (puerto 8084)
--   soporte   -> crud_soporte   (puerto 8085)
-- Las columnas marcadas con "-- AJUSTE CLIENTE" no estaban en el
-- SQL original y se agregaron para cubrir lo que usa el frontend.
-- ============================================================

CREATE SCHEMA IF NOT EXISTS "auth";
CREATE SCHEMA IF NOT EXISTS "soporte";
CREATE SCHEMA IF NOT EXISTS "educacion";
CREATE SCHEMA IF NOT EXISTS "finanzas";
CREATE SCHEMA IF NOT EXISTS "negocio";


-- ============================================================
-- FUNCIÓN GLOBAL DE AUDITORÍA
-- ============================================================

CREATE OR REPLACE FUNCTION fn_update_fecha_modificacion()
RETURNS TRIGGER AS $$
BEGIN
    NEW.fecha_modificacion = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- ============================================================
-- ESQUEMA: auth
-- Tablas: rol, tipo_documento, usuario, credencial,
--         usuario_rol, auditoria_login
-- ============================================================

-- auth.rol
CREATE TABLE "auth"."rol" (
    "id_rol"              SERIAL PRIMARY KEY,
    "nombre_rol"          VARCHAR(100) UNIQUE NOT NULL,
    "descripcion"         TEXT,
    "activo"              BOOLEAN     NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_rol_mod
    BEFORE UPDATE ON "auth"."rol"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- auth.tipo_documento
CREATE TABLE "auth"."tipo_documento" (
    "id_tipo_documento"   SERIAL PRIMARY KEY,
    "nombre"              VARCHAR(50)  NOT NULL,
    "codigo"              VARCHAR(10)  UNIQUE NOT NULL,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_tipo_documento_mod
    BEFORE UPDATE ON "auth"."tipo_documento"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- auth.usuario
CREATE TABLE "auth"."usuario" (
    "id_usuario"          SERIAL PRIMARY KEY,
    "tipo_documento"      INT          REFERENCES "auth"."tipo_documento"("id_tipo_documento") ON DELETE RESTRICT,
    "nombre"              VARCHAR(100) NOT NULL,
    "apellido"            VARCHAR(100) NOT NULL,
    "email"               VARCHAR(150) UNIQUE NOT NULL,
    "telefono"            VARCHAR(20),
    "cedula"              VARCHAR(50)  UNIQUE,
    "ciudad"              VARCHAR(100),
    "direccion"           VARCHAR(200),                       -- AJUSTE CLIENTE
    "fecha_nacimiento"    DATE,                               -- AJUSTE CLIENTE
    "estado"              VARCHAR(20)  NOT NULL DEFAULT 'activo'
                              CHECK (estado IN ('activo','inactivo','suspendido')),
    "fecha_registro"      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    "fecha_ultima_sesion" TIMESTAMP,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_usuario_mod
    BEFORE UPDATE ON "auth"."usuario"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- auth.credencial
CREATE TABLE "auth"."credencial" (
    "id_credencial"         SERIAL PRIMARY KEY,
    "id_usuario"            INT         UNIQUE NOT NULL
                                REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "contrasena_hash"       VARCHAR(255) NOT NULL,
    "salt"                  VARCHAR(100) NOT NULL,
    "algoritmo"             VARCHAR(20)  NOT NULL DEFAULT 'bcrypt'
                                CHECK (algoritmo IN ('bcrypt','argon2','sha256')),
    "fecha_actualizacion"   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    "fecha_ultimo_cambio"   TIMESTAMP,
    "intentos_fallidos"     INT          DEFAULT 0,
    "bloqueado_hasta"       TIMESTAMP,
    "requiere_cambio"       BOOLEAN      DEFAULT false,
    "activo"                BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_credencial_mod
    BEFORE UPDATE ON "auth"."credencial"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- auth.usuario_rol
CREATE TABLE "auth"."usuario_rol" (
    "id_usuario_rol"      SERIAL PRIMARY KEY,
    "id_usuario"          INT       NOT NULL
                              REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_rol"              INT       NOT NULL
                              REFERENCES "auth"."rol"("id_rol") ON DELETE RESTRICT,
    "fecha_asignacion"    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "activo"              BOOLEAN   NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE("id_usuario", "id_rol")
);

CREATE TRIGGER trg_usuario_rol_mod
    BEFORE UPDATE ON "auth"."usuario_rol"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- auth.auditoria_login
CREATE TABLE "auth"."auditoria_login" (
    "id_auditoria"        SERIAL PRIMARY KEY,
    "id_usuario"          INT         NOT NULL
                              REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "tipo_evento"         VARCHAR(30) NOT NULL DEFAULT 'login'
                              CHECK (tipo_evento IN ('login','logout','cambio_contrasena','acceso_denegado')),
    "ip_address"          VARCHAR(45),
    "navegador"           VARCHAR(200),
    "fecha_evento"        TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    "estado_evento"       VARCHAR(20) NOT NULL DEFAULT 'exitoso'
                              CHECK (estado_evento IN ('exitoso','fallido')),
    "fecha_creacion"      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
    -- Sin activo ni fecha_modificacion: los logs de auditoria son inmutables
);


-- ============================================================
-- ESQUEMA: soporte
-- Tablas: estado_pqr, pqr, adjunto, registro_actividad
-- (fusión de pqr_schema + auditoria)
-- ============================================================

-- soporte.estado_pqr
CREATE TABLE "soporte"."estado_pqr" (
    "id_estado"           SERIAL PRIMARY KEY,
    "nombre"              VARCHAR(50)  NOT NULL,
    "descripcion"         VARCHAR(255),
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_estado_pqr_mod
    BEFORE UPDATE ON "soporte"."estado_pqr"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- soporte.pqr
CREATE TABLE "soporte"."pqr" (
    "id_pqr"              SERIAL PRIMARY KEY,
    "id_usuario"          INT       NOT NULL
                              REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "radicado"            VARCHAR(30)  UNIQUE,                -- AJUSTE CLIENTE
    "titulo"              VARCHAR(200),                       -- AJUSTE CLIENTE
    "tipo"                VARCHAR(20)  NOT NULL DEFAULT 'peticion'   -- AJUSTE CLIENTE
                              CHECK (tipo IN ('peticion','queja','reclamo','sugerencia')),
    "categoria"           VARCHAR(100),                       -- AJUSTE CLIENTE
    "prioridad"           VARCHAR(10)  NOT NULL DEFAULT 'media'      -- AJUSTE CLIENTE
                              CHECK (prioridad IN ('alta','media','baja')),
    "asesor"              VARCHAR(100),                       -- AJUSTE CLIENTE
    "mensaje_respuesta"   TEXT,                               -- AJUSTE CLIENTE
    "descripcion"         TEXT      NOT NULL,
    "id_estado"           INT       NOT NULL
                              REFERENCES "soporte"."estado_pqr"("id_estado") ON DELETE RESTRICT,
    "activo"              BOOLEAN   NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_pqr_mod
    BEFORE UPDATE ON "soporte"."pqr"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- soporte.adjunto
CREATE TABLE "soporte"."adjunto" (
    "id_adjunto"          SERIAL PRIMARY KEY,
    "id_pqr"              INT          NOT NULL
                              REFERENCES "soporte"."pqr"("id_pqr") ON DELETE RESTRICT,
    "nombre_archivo"      VARCHAR(255) NOT NULL,
    "ruta_archivo"        VARCHAR(500) NOT NULL,
    "tipo_mime"           VARCHAR(100),
    "tamano_bytes"        INT,
    "fecha_carga"         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_adjunto_mod
    BEFORE UPDATE ON "soporte"."adjunto"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- soporte.registro_actividad (antes auditoria.registro_actividad)
CREATE TABLE "soporte"."registro_actividad" (
    "id_actividad"        SERIAL PRIMARY KEY,
    "id_usuario"          INT          REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "tipo_actividad"      VARCHAR(100),
    "descripcion"         TEXT,
    "entidad_afectada"    VARCHAR(100),
    "fecha_actividad"     TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
    -- Sin activo ni fecha_modificacion: los registros de auditoria son inmutables
);


-- ============================================================
-- ESQUEMA: educacion
-- Tablas: modulo_educativo, contenido, leccion,
--         progreso_educativo, progreso_leccion
-- ============================================================

-- educacion.modulo_educativo
CREATE TABLE "educacion"."modulo_educativo" (
    "id_modulo"           SERIAL PRIMARY KEY,
    "titulo"              VARCHAR(200) NOT NULL,
    "descripcion"         TEXT,
    "contenido"           TEXT,
    "nivel"               VARCHAR(20)  NOT NULL DEFAULT 'basico'
                              CHECK (nivel IN ('basico','intermedio','avanzado')),
    "url_thumbnail"       VARCHAR(500),
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_modulo_educativo_mod
    BEFORE UPDATE ON "educacion"."modulo_educativo"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- educacion.contenido
CREATE TABLE "educacion"."contenido" (
    "id_contenido"        SERIAL PRIMARY KEY,
    "titulo"              VARCHAR(200) NOT NULL,
    "descripcion"         TEXT,
    "duracion_minutos"    INT,
    "url_video"           VARCHAR(500),
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_contenido_mod
    BEFORE UPDATE ON "educacion"."contenido"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- educacion.leccion
CREATE TABLE "educacion"."leccion" (
    "id_leccion"          SERIAL PRIMARY KEY,
    "id_modulo"           INT          NOT NULL
                              REFERENCES "educacion"."modulo_educativo"("id_modulo") ON DELETE RESTRICT,
    "id_contenido"        INT
                              REFERENCES "educacion"."contenido"("id_contenido") ON DELETE RESTRICT,
    "titulo"              VARCHAR(200) NOT NULL,
    "descripcion"         TEXT,
    "duracion_minutos"    INT,
    "url_video"           VARCHAR(500),
    "numero_leccion"      INT,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_leccion_mod
    BEFORE UPDATE ON "educacion"."leccion"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- educacion.progreso_educativo
CREATE TABLE "educacion"."progreso_educativo" (
    "id_progreso"             SERIAL PRIMARY KEY,
    "id_usuario"              INT       NOT NULL
                                  REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_modulo"               INT       NOT NULL
                                  REFERENCES "educacion"."modulo_educativo"("id_modulo") ON DELETE RESTRICT,
    "porcentaje_completado"   INT       DEFAULT 0,
    "fecha_inicio"            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "fecha_completado"        TIMESTAMP,
    "calificacion"            INT,
    "activo"                  BOOLEAN   NOT NULL DEFAULT true,
    "fecha_creacion"          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE("id_usuario", "id_modulo")
);

CREATE TRIGGER trg_progreso_educativo_mod
    BEFORE UPDATE ON "educacion"."progreso_educativo"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- educacion.progreso_leccion
CREATE TABLE "educacion"."progreso_leccion" (
    "id_progreso_leccion" SERIAL PRIMARY KEY,
    "id_usuario"          INT       NOT NULL
                              REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_leccion"          INT       NOT NULL
                              REFERENCES "educacion"."leccion"("id_leccion") ON DELETE RESTRICT,
    "completado"          BOOLEAN   DEFAULT false,
    "fecha_inicio"        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "fecha_completado"    TIMESTAMP,
    "activo"              BOOLEAN   NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE("id_usuario", "id_leccion")
);

CREATE TRIGGER trg_progreso_leccion_mod
    BEFORE UPDATE ON "educacion"."progreso_leccion"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();


-- ============================================================
-- ESQUEMA: finanzas
-- Tablas: categoria, movimiento_ingreso_egreso, tipo_ingreso,
--         finanzas, tipo_inversion, nivel_riesgo,
--         movimiento_inversion, tipo_ingreso_inversion,
--         inversion, editar_meta, movimiento_meta,
--         tipo_ingreso_meta, meta
-- (fusión de finanzas + inversion + metas)
-- ============================================================

-- finanzas.categoria
CREATE TABLE "finanzas"."categoria" (
    "id_categoria"        SERIAL PRIMARY KEY,
    "nombre"              VARCHAR(100) NOT NULL,
    "descripcion"         TEXT,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_categoria_mod
    BEFORE UPDATE ON "finanzas"."categoria"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.movimiento_ingreso_egreso
CREATE TABLE "finanzas"."movimiento_ingreso_egreso" (
    "id_movimiento_dinero"  SERIAL PRIMARY KEY,
    "id_usuario"            INT           NOT NULL            -- AJUSTE CLIENTE
                                REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_categoria"          INT                               -- AJUSTE CLIENTE
                                REFERENCES "finanzas"."categoria"("id_categoria") ON DELETE RESTRICT,
    "nombre"                VARCHAR(120)  NOT NULL,
    "monto"                 DECIMAL(18,2) NOT NULL,
    "es_ingreso"            BOOLEAN       NOT NULL,
    "fecha"                 DATE          NOT NULL DEFAULT CURRENT_DATE,  -- AJUSTE CLIENTE
    "metodo_pago"           VARCHAR(50),                      -- AJUSTE CLIENTE
    "observaciones"         TEXT,                             -- AJUSTE CLIENTE
    "activo"                BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"        TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_fin_movimiento_mod
    BEFORE UPDATE ON "finanzas"."movimiento_ingreso_egreso"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.tipo_ingreso
CREATE TABLE "finanzas"."tipo_ingreso" (
    "id_tipo_ingreso"         SERIAL PRIMARY KEY,
    "id_movimiento_dinero"    INT REFERENCES "finanzas"."movimiento_ingreso_egreso"("id_movimiento_dinero") ON DELETE RESTRICT,
    "nombre_movimiento_pago"  VARCHAR(120) NOT NULL,
    "descripcion"             VARCHAR(150),
    "activo"                  BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_fin_tipo_ingreso_mod
    BEFORE UPDATE ON "finanzas"."tipo_ingreso"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.finanzas
CREATE TABLE "finanzas"."finanzas" (
    "id_finanzas"           SERIAL PRIMARY KEY,
    "id_usuario"            INT           NOT NULL
                                REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_movimiento_dinero"  INT           REFERENCES "finanzas"."movimiento_ingreso_egreso"("id_movimiento_dinero") ON DELETE RESTRICT,
    "id_categoria"          INT           REFERENCES "finanzas"."categoria"("id_categoria") ON DELETE RESTRICT,
    "monto_presupuesto"     DECIMAL(12,2),
    "gasto"                 DECIMAL(12,2),
    "disponible"            DECIMAL(12,2),
    "fecha"                 DATE          DEFAULT CURRENT_DATE,
    "activo"                BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"        TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"    TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_finanzas_mod
    BEFORE UPDATE ON "finanzas"."finanzas"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.tipo_inversion (antes inversion.tipo_inversion)
CREATE TABLE "finanzas"."tipo_inversion" (
    "id_tipo_inversion"   SERIAL PRIMARY KEY,
    "nombre"              VARCHAR(60)  NOT NULL,
    "descripcion"         VARCHAR(150),
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_tipo_inversion_mod
    BEFORE UPDATE ON "finanzas"."tipo_inversion"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.nivel_riesgo (antes inversion.nivel_riesgo)
CREATE TABLE "finanzas"."nivel_riesgo" (
    "id_nivel_riesgo"     SERIAL PRIMARY KEY,
    "nombre"              VARCHAR(40) NOT NULL,
    "activo"              BOOLEAN     NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_nivel_riesgo_mod
    BEFORE UPDATE ON "finanzas"."nivel_riesgo"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.movimiento_inversion (antes inversion.movimiento_ingreso_egreso)
CREATE TABLE "finanzas"."movimiento_inversion" (
    "id_movimiento_dinero"  SERIAL PRIMARY KEY,
    "nombre"                VARCHAR(120)  NOT NULL,
    "monto"                 DECIMAL(18,2) NOT NULL,
    "es_ingreso"            BOOLEAN       NOT NULL,
    "activo"                BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"        TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_inv_movimiento_mod
    BEFORE UPDATE ON "finanzas"."movimiento_inversion"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.tipo_ingreso_inversion (antes inversion.tipo_ingreso)
CREATE TABLE "finanzas"."tipo_ingreso_inversion" (
    "id_tipo_ingreso"         SERIAL PRIMARY KEY,
    "id_movimiento_dinero"    INT REFERENCES "finanzas"."movimiento_inversion"("id_movimiento_dinero") ON DELETE RESTRICT,
    "nombre_movimiento_pago"  VARCHAR(120) NOT NULL,
    "descripcion"             VARCHAR(150),
    "activo"                  BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_inv_tipo_ingreso_mod
    BEFORE UPDATE ON "finanzas"."tipo_ingreso_inversion"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.inversion (antes inversion.inversion)
CREATE TABLE "finanzas"."inversion" (
    "id_inversion"          SERIAL PRIMARY KEY,
    "id_usuario"            INT           NOT NULL
                                REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_tipo_inversion"     INT           NOT NULL
                                REFERENCES "finanzas"."tipo_inversion"("id_tipo_inversion") ON DELETE RESTRICT,
    "id_nivel_riesgo"       INT           NOT NULL
                                REFERENCES "finanzas"."nivel_riesgo"("id_nivel_riesgo") ON DELETE RESTRICT,
    "id_movimiento_dinero"  INT           REFERENCES "finanzas"."movimiento_inversion"("id_movimiento_dinero") ON DELETE RESTRICT,
    "nombre"                VARCHAR(120),
    "monto"                 DECIMAL(18,2),
    "rentabilidad"          DECIMAL(8,4),
    "fecha_inicio"          DATE,
    "fecha_fin"             DATE,
    "activo"                BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"        TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_inversion_mod
    BEFORE UPDATE ON "finanzas"."inversion"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.editar_meta (antes metas.editar_meta)
CREATE TABLE "finanzas"."editar_meta" (
    "id_editar_meta"      SERIAL PRIMARY KEY,
    "nombre"              VARCHAR(40),
    "monto_actual"        DECIMAL(18,2),
    "monto_objetivo"      DECIMAL(18,2) NOT NULL,
    "ahorro_mensual"      DECIMAL(18,2),
    "fecha_objetivo"      DATE          DEFAULT CURRENT_DATE,
    "descripcion"         VARCHAR(100),
    "activo"              BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP     DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_editar_meta_mod
    BEFORE UPDATE ON "finanzas"."editar_meta"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.movimiento_meta (antes metas.movimiento_ingreso_egreso)
CREATE TABLE "finanzas"."movimiento_meta" (
    "id_movimiento_dinero"  SERIAL PRIMARY KEY,
    "nombre"                VARCHAR(120)  NOT NULL,
    "monto"                 DECIMAL(18,2) NOT NULL,
    "es_ingreso"            BOOLEAN       NOT NULL,
    "activo"                BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"        TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_meta_movimiento_mod
    BEFORE UPDATE ON "finanzas"."movimiento_meta"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.tipo_ingreso_meta (antes metas.tipo_ingreso)
CREATE TABLE "finanzas"."tipo_ingreso_meta" (
    "id_tipo_ingreso"         SERIAL PRIMARY KEY,
    "id_movimiento_dinero"    INT REFERENCES "finanzas"."movimiento_meta"("id_movimiento_dinero") ON DELETE RESTRICT,
    "nombre_movimiento_pago"  VARCHAR(120) NOT NULL,
    "descripcion"             VARCHAR(150),
    "activo"                  BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_meta_tipo_ingreso_mod
    BEFORE UPDATE ON "finanzas"."tipo_ingreso_meta"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.meta (antes metas.meta)
CREATE TABLE "finanzas"."meta" (
    "id_meta"               SERIAL PRIMARY KEY,
    "id_usuario"            INT           NOT NULL
                                REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_editar_meta"        INT           NOT NULL
                                REFERENCES "finanzas"."editar_meta"("id_editar_meta") ON DELETE RESTRICT,
    "id_movimiento_dinero"  INT           REFERENCES "finanzas"."movimiento_meta"("id_movimiento_dinero") ON DELETE RESTRICT,
    "nombre"                VARCHAR(120),
    "descripcion"           TEXT,
    "monto_objetivo"        DECIMAL(18,2),
    "monto_actual"          DECIMAL(18,2),
    "fecha_limite"          DATE,
    "color"                 VARCHAR(20),
    "icono"                 VARCHAR(50),                      -- AJUSTE CLIENTE
    "activo"                BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"        TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_meta_mod
    BEFORE UPDATE ON "finanzas"."meta"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.solicitud_consolidacion (AJUSTE CLIENTE: modulo "Resuelve tu deuda")
CREATE TABLE "finanzas"."solicitud_consolidacion" (
    "id_solicitud"          SERIAL PRIMARY KEY,
    "id_usuario"            INT           NOT NULL
                                REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "saldo_total"           DECIMAL(18,2) NOT NULL,
    "cuota_actual"          DECIMAL(18,2) NOT NULL,
    "cuota_propuesta"       DECIMAL(18,2) NOT NULL,
    "estado"                VARCHAR(20)   NOT NULL DEFAULT 'pendiente'
                                CHECK (estado IN ('pendiente','aprobada','rechazada')),
    "activo"                BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"        TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"    TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_solicitud_consolidacion_mod
    BEFORE UPDATE ON "finanzas"."solicitud_consolidacion"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();


-- ============================================================
-- ESQUEMA: negocio
-- Tablas: banco, producto_crediticio, asesor_bancario,
--         contacto_asesor, lead, conversacion_usuario_asesor,
--         credito_desembolsado, transaccion_comision
-- (fusión de intermediacion + leads_schema + creditos + transacciones)
-- ============================================================

-- negocio.banco (antes intermediacion.banco)
CREATE TABLE "negocio"."banco" (
    "id_banco"              SERIAL PRIMARY KEY,
    "nombre_banco"          VARCHAR(100) UNIQUE NOT NULL,
    "ciudad"                VARCHAR(100),
    "contacto"              VARCHAR(100),
    "telefono"              VARCHAR(20),
    "email"                 VARCHAR(150),
    "comision_porcentaje"   DECIMAL(5,2),
    "url_logo"              VARCHAR(500),
    "descripcion"           TEXT,
    "sitio_web"             VARCHAR(300),
    "estado"                VARCHAR(20)  NOT NULL DEFAULT 'activo'
                                CHECK (estado IN ('activo','inactivo')),
    "fecha_registro"        TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    "activo"                BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_banco_mod
    BEFORE UPDATE ON "negocio"."banco"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- negocio.producto_crediticio (antes intermediacion.producto_crediticio)
CREATE TABLE "negocio"."producto_crediticio" (
    "id_producto"         SERIAL PRIMARY KEY,
    "id_banco"            INT           NOT NULL
                              REFERENCES "negocio"."banco"("id_banco") ON DELETE RESTRICT,
    "nombre_producto"     VARCHAR(150)  NOT NULL,
    "descripcion"         TEXT,
    "monto_minimo"        DECIMAL(12,2),
    "monto_maximo"        DECIMAL(12,2),
    "tasa_minima"         DECIMAL(5,2),
    "tasa_maxima"         DECIMAL(5,2),
    "plazo_minimo"        INT,
    "plazo_maximo"        INT,
    "requisitos"          TEXT,
    "activo"              BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_producto_crediticio_mod
    BEFORE UPDATE ON "negocio"."producto_crediticio"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- negocio.asesor_bancario (antes intermediacion.asesor_bancario)
CREATE TABLE "negocio"."asesor_bancario" (
    "id_asesor"           SERIAL PRIMARY KEY,
    "id_banco"            INT          NOT NULL
                              REFERENCES "negocio"."banco"("id_banco") ON DELETE RESTRICT,
    "nombre"              VARCHAR(100) NOT NULL,
    "apellido"            VARCHAR(100) NOT NULL,
    "email"               VARCHAR(150) NOT NULL,
    "telefono"            VARCHAR(20),
    "especialidad"        VARCHAR(100),
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_asesor_bancario_mod
    BEFORE UPDATE ON "negocio"."asesor_bancario"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- negocio.contacto_asesor (antes intermediacion.contacto_asesor)
CREATE TABLE "negocio"."contacto_asesor" (
    "id_contacto"         SERIAL PRIMARY KEY,
    "id_asesor"           INT         NOT NULL
                              REFERENCES "negocio"."asesor_bancario"("id_asesor") ON DELETE RESTRICT,
    "whatsapp"            VARCHAR(20),
    "email"               VARCHAR(150),
    "telefono"            VARCHAR(20),
    "disponible_desde"    VARCHAR(5),                     -- AJUSTE BEE: bee no reconoce TIME, se guarda "HH:MM"
    "disponible_hasta"    VARCHAR(5),                     -- AJUSTE BEE
    "dias_disponibles"    VARCHAR(100),
    "activo"              BOOLEAN     NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_contacto_asesor_mod
    BEFORE UPDATE ON "negocio"."contacto_asesor"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- negocio.lead (antes leads_schema.lead)
CREATE TABLE "negocio"."lead" (
    "id_lead"             SERIAL PRIMARY KEY,
    "id_usuario"          INT           NOT NULL
                              REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_producto"         INT           NOT NULL
                              REFERENCES "negocio"."producto_crediticio"("id_producto") ON DELETE RESTRICT,
    "id_asesor"           INT
                              REFERENCES "negocio"."asesor_bancario"("id_asesor") ON DELETE RESTRICT,
    "tipo_credito"        VARCHAR(100),
    "monto_interes"       DECIMAL(12,2),
    "plazo_interes"       INT,
    "estado_lead"         VARCHAR(30)   NOT NULL DEFAULT 'nuevo'
                              CHECK (estado_lead IN ('nuevo','contactado','en_proceso','aprobado','rechazado','cancelado')),
    "fecha_generacion"    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    "fecha_contacto"      TIMESTAMP,
    "observaciones"       TEXT,
    "activo"              BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_lead_mod
    BEFORE UPDATE ON "negocio"."lead"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- negocio.conversacion_usuario_asesor (antes leads_schema.conversacion_usuario_asesor)
CREATE TABLE "negocio"."conversacion_usuario_asesor" (
    "id_conversacion"     SERIAL PRIMARY KEY,
    "id_lead"             INT         NOT NULL
                              REFERENCES "negocio"."lead"("id_lead") ON DELETE RESTRICT,
    "id_usuario"          INT         NOT NULL
                              REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_asesor"           INT         NOT NULL
                              REFERENCES "negocio"."asesor_bancario"("id_asesor") ON DELETE RESTRICT,
    "tipo_contacto"       VARCHAR(30) NOT NULL DEFAULT 'email'
                              CHECK (tipo_contacto IN ('email','telefono','whatsapp','presencial')),
    "asunto"              VARCHAR(200),
    "contenido"           TEXT,
    "fecha_mensaje"       TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    "activo"              BOOLEAN     NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_conversacion_mod
    BEFORE UPDATE ON "negocio"."conversacion_usuario_asesor"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- negocio.credito_desembolsado (antes creditos.credito_desembolsado)
CREATE TABLE "negocio"."credito_desembolsado" (
    "id_credito"            SERIAL PRIMARY KEY,
    "id_lead"               INT           NOT NULL
                                REFERENCES "negocio"."lead"("id_lead") ON DELETE RESTRICT,
    "id_usuario"            INT           NOT NULL
                                REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_producto"           INT           NOT NULL
                                REFERENCES "negocio"."producto_crediticio"("id_producto") ON DELETE RESTRICT,
    "id_banco"              INT           NOT NULL
                                REFERENCES "negocio"."banco"("id_banco") ON DELETE RESTRICT,
    "numero_credito"        VARCHAR(50)   UNIQUE,
    "monto_aprobado"        DECIMAL(12,2) NOT NULL,
    "tasa_interes_final"    DECIMAL(5,2),
    "plazo_meses"           INT,
    "fecha_aprobacion"      DATE,
    "fecha_desembolso"      DATE,
    "estado_credito"        VARCHAR(20)   NOT NULL DEFAULT 'activo'
                                CHECK (estado_credito IN ('activo','pagado','vencido','cancelado')),
    "saldo_actual"          DECIMAL(12,2),
    "activo"                BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"        TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"    TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_credito_mod
    BEFORE UPDATE ON "negocio"."credito_desembolsado"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- negocio.transaccion_comision (antes transacciones.transaccion_comision)
CREATE TABLE "negocio"."transaccion_comision" (
    "id_transaccion"        SERIAL PRIMARY KEY,
    "id_credito"            INT           NOT NULL
                                REFERENCES "negocio"."credito_desembolsado"("id_credito") ON DELETE RESTRICT,
    "id_banco"              INT           NOT NULL
                                REFERENCES "negocio"."banco"("id_banco") ON DELETE RESTRICT,
    "monto_comision"        DECIMAL(12,2) NOT NULL,
    "porcentaje_aplicado"   DECIMAL(5,2),
    "fecha_transaccion"     TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    "estado"                VARCHAR(20)   NOT NULL DEFAULT 'pendiente'
                                CHECK (estado IN ('pendiente','pagada','cancelada')),
    "referencia_pago"       VARCHAR(100),
    "activo"                BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"        TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"    TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_transaccion_mod
    BEFORE UPDATE ON "negocio"."transaccion_comision"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();


-- ============================================================
-- ÍNDICES
-- ============================================================

CREATE INDEX idx_usuario_email      ON "auth"."usuario"("email");
CREATE INDEX idx_usuario_ciudad     ON "auth"."usuario"("ciudad");
CREATE INDEX idx_usuario_cedula     ON "auth"."usuario"("cedula");

CREATE INDEX idx_pqr_usuario        ON "soporte"."pqr"("id_usuario");

CREATE INDEX idx_progreso_usuario   ON "educacion"."progreso_educativo"("id_usuario");

CREATE INDEX idx_meta_usuario       ON "finanzas"."meta"("id_usuario");
CREATE INDEX idx_movimiento_usuario ON "finanzas"."movimiento_ingreso_egreso"("id_usuario");
CREATE INDEX idx_solicitud_usuario  ON "finanzas"."solicitud_consolidacion"("id_usuario");
CREATE INDEX idx_inversion_usuario  ON "finanzas"."inversion"("id_usuario");

CREATE INDEX idx_lead_usuario       ON "negocio"."lead"("id_usuario");
CREATE INDEX idx_lead_estado        ON "negocio"."lead"("estado_lead");
CREATE INDEX idx_credito_usuario    ON "negocio"."credito_desembolsado"("id_usuario");
CREATE INDEX idx_credito_estado     ON "negocio"."credito_desembolsado"("estado_credito");