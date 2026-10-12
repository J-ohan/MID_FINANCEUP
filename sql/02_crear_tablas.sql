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
-- Las tablas marcadas con "-- NUEVA (CLIENTE)" no existian y guardan
-- datos que el frontend muestra o edita.
--
-- VERSION UNIFICADA (MID + ajustes del frontend)
--   Parte del SQL del MID y le suma lo que necesita el frontend.
--   Ninguna tabla ni columna del MID se elimino.
--   Se conserva finanzas.solicitud_consolidacion (modulo "Resuelve tu deuda").
--   Tablas nuevas:
--     - auth.cuenta_social            login con Google / Apple / Facebook
--     - auth.codigo_verificacion      codigo 2FA del login y recuperacion de contrasena
--     - soporte.respuesta_pqr         respuestas/seguimiento de cada PQR
--     - soporte.pregunta_frecuente    preguntas frecuentes de Linea de Ayuda
--     - educacion.escuela             escuelas que agrupan los cursos
--     - educacion.instructor          instructor de cada curso
--     - educacion.modulo_detalle      "Lo que aprenderas" y requisitos del curso
--     - educacion.certificado         certificados emitidos al terminar un curso
--     - finanzas.historial_inversion  valor de cada inversion en el tiempo (grafica)
--     - negocio.producto_detalle      etiquetas, requisitos, caracteristicas,
--                                     documentos y perfiles de cada oferta
--     - negocio.producto_tarifa       tabla de tarifas de cada oferta
--     - negocio.producto_pregunta     preguntas frecuentes de cada oferta
--   Columnas que dejaron de ser obligatorias (los formularios no las piden):
--     - finanzas.inversion.id_tipo_inversion
--     - finanzas.meta.id_editar_meta
--   Columnas nuevas: ver "-- AJUSTE CLIENTE" en cada tabla.
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
--         usuario_rol, auditoria_login, cuenta_social,
--         codigo_verificacion
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
    "acepta_terminos"     BOOLEAN      NOT NULL DEFAULT false, -- AJUSTE CLIENTE: casilla de terminos del registro
    "fecha_aceptacion_terminos" TIMESTAMP,                    -- AJUSTE CLIENTE
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
-- Los usuarios que entran solo con Google/Apple/Facebook no tienen fila aqui
-- (ver auth.cuenta_social).
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

-- auth.cuenta_social                                        -- NUEVA (CLIENTE)
-- Botones "Google", "Apple" y "Facebook" del login. Un usuario puede tener
-- varias cuentas vinculadas, pero solo una por proveedor.
CREATE TABLE "auth"."cuenta_social" (
    "id_cuenta_social"    SERIAL PRIMARY KEY,
    "id_usuario"          INT          NOT NULL
                              REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "proveedor"           VARCHAR(20)  NOT NULL
                              CHECK (proveedor IN ('google','apple','facebook')),
    "id_externo"          VARCHAR(255) NOT NULL,              -- identificador que entrega el proveedor
    "email_proveedor"     VARCHAR(150),
    "fecha_vinculacion"   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE("proveedor", "id_externo"),
    UNIQUE("id_usuario", "proveedor")
);

CREATE TRIGGER trg_cuenta_social_mod
    BEFORE UPDATE ON "auth"."cuenta_social"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- auth.codigo_verificacion                                  -- NUEVA (CLIENTE)
-- Codigo de 6 digitos del paso "Verifica tu identidad" (2FA) y del flujo
-- "Recuperar contrasena" (por correo o SMS). Se guarda solo el hash del codigo.
CREATE TABLE "auth"."codigo_verificacion" (
    "id_codigo"           SERIAL PRIMARY KEY,
    "id_usuario"          INT          NOT NULL
                              REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "proposito"           VARCHAR(30)  NOT NULL
                              CHECK (proposito IN ('login_2fa','recuperar_contrasena')),
    "canal"               VARCHAR(10)  NOT NULL DEFAULT 'email'
                              CHECK (canal IN ('email','sms')),
    "codigo_hash"         VARCHAR(255) NOT NULL,
    "fecha_expiracion"    TIMESTAMP    NOT NULL,
    "fecha_uso"           TIMESTAMP,                          -- NULL mientras no se haya usado
    "intentos"            INT          NOT NULL DEFAULT 0,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_codigo_verificacion_mod
    BEFORE UPDATE ON "auth"."codigo_verificacion"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();


-- ============================================================
-- ESQUEMA: soporte
-- Tablas: estado_pqr, pqr, adjunto, respuesta_pqr,
--         registro_actividad, pregunta_frecuente
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
    "progreso"            INT          NOT NULL DEFAULT 0     -- AJUSTE CLIENTE: barra de avance de la tarjeta
                              CHECK (progreso BETWEEN 0 AND 100),
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

-- soporte.respuesta_pqr                                     -- NUEVA (CLIENTE)
-- Hilo de respuestas de una PQR. El frontend muestra cuantas respuestas
-- tiene cada PQR; pqr.mensaje_respuesta sigue guardando la ultima.
CREATE TABLE "soporte"."respuesta_pqr" (
    "id_respuesta"        SERIAL PRIMARY KEY,
    "id_pqr"              INT          NOT NULL
                              REFERENCES "soporte"."pqr"("id_pqr") ON DELETE RESTRICT,
    "id_usuario"          INT          REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "autor"               VARCHAR(100) NOT NULL,              -- nombre que se muestra (ej. 'Soporte técnico')
    "mensaje"             TEXT         NOT NULL,
    "fecha_respuesta"     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_respuesta_pqr_mod
    BEFORE UPDATE ON "soporte"."respuesta_pqr"
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

-- soporte.pregunta_frecuente                                -- NUEVA (CLIENTE)
-- Preguntas frecuentes de la pagina "Linea de ayuda", filtradas por categoria.
CREATE TABLE "soporte"."pregunta_frecuente" (
    "id_pregunta"         SERIAL PRIMARY KEY,
    "categoria"           VARCHAR(50)  NOT NULL,              -- 'Cuenta', 'Seguridad', 'Educación', 'Alianzas', 'PQR'
    "pregunta"            VARCHAR(255) NOT NULL,
    "respuesta"           TEXT         NOT NULL,
    "lectura_minutos"     INT          NOT NULL DEFAULT 1,
    "utilidad"            INT          NOT NULL DEFAULT 0     -- % de usuarios que la marcaron como util
                              CHECK (utilidad BETWEEN 0 AND 100),
    "orden"               INT,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_pregunta_frecuente_mod
    BEFORE UPDATE ON "soporte"."pregunta_frecuente"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();


-- ============================================================
-- ESQUEMA: educacion
-- Tablas: escuela, instructor, modulo_educativo, modulo_detalle,
--         contenido, leccion, progreso_educativo,
--         progreso_leccion, certificado
-- En el frontend un "modulo_educativo" se llama "curso".
-- ============================================================

-- educacion.escuela                                         -- NUEVA (CLIENTE)
-- Agrupa los cursos: "Finanzas Personales", "Inversión".
CREATE TABLE "educacion"."escuela" (
    "id_escuela"          SERIAL PRIMARY KEY,
    "slug"                VARCHAR(60)  UNIQUE NOT NULL,       -- id que usa el frontend ('finanzas-personales')
    "nombre"              VARCHAR(100) NOT NULL,
    "descripcion"         TEXT,
    "orden"               INT,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_escuela_mod
    BEFORE UPDATE ON "educacion"."escuela"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- educacion.instructor                                      -- NUEVA (CLIENTE)
CREATE TABLE "educacion"."instructor" (
    "id_instructor"       SERIAL PRIMARY KEY,
    "nombre"              VARCHAR(150) NOT NULL,
    "cargo"               VARCHAR(150),
    "iniciales"           VARCHAR(5),
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_instructor_mod
    BEFORE UPDATE ON "educacion"."instructor"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- educacion.modulo_educativo
CREATE TABLE "educacion"."modulo_educativo" (
    "id_modulo"           SERIAL PRIMARY KEY,
    "id_escuela"          INT          REFERENCES "educacion"."escuela"("id_escuela") ON DELETE RESTRICT,       -- AJUSTE CLIENTE
    "id_instructor"       INT          REFERENCES "educacion"."instructor"("id_instructor") ON DELETE RESTRICT, -- AJUSTE CLIENTE
    "slug"                VARCHAR(80)  UNIQUE,                -- AJUSTE CLIENTE: id del curso en la URL (/educacion/curso/:cursoId)
    "titulo"              VARCHAR(200) NOT NULL,
    "descripcion"         TEXT,
    "contenido"           TEXT,
    "nivel"               VARCHAR(20)  NOT NULL DEFAULT 'basico'
                              CHECK (nivel IN ('basico','intermedio','avanzado')),
    "categoria"           VARCHAR(20)                         -- AJUSTE CLIENTE: filtro "Categoría" del catalogo
                              CHECK (categoria IN ('fundamentos','credito','ahorro','deuda','fraude')),
    "formato"             VARCHAR(20)  NOT NULL DEFAULT 'video'      -- AJUSTE CLIENTE: filtro "Formato"
                              CHECK (formato IN ('video','articulo','quiz')),
    "duracion"            VARCHAR(30),                        -- AJUSTE CLIENTE: ej. '30 minutos'
    "resumen_lecciones"   VARCHAR(100),                       -- AJUSTE CLIENTE: ej. '4 lecciones + evaluación'
    "otorga_certificado"  BOOLEAN      NOT NULL DEFAULT false,       -- AJUSTE CLIENTE
    "estudiantes"         INT          NOT NULL DEFAULT 0,           -- AJUSTE CLIENTE
    "calificacion"        DECIMAL(2,1)                        -- AJUSTE CLIENTE: estrellas de 0 a 5
                              CHECK (calificacion BETWEEN 0 AND 5),
    "url_thumbnail"       VARCHAR(500),
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_modulo_educativo_mod
    BEFORE UPDATE ON "educacion"."modulo_educativo"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- educacion.modulo_detalle                                  -- NUEVA (CLIENTE)
-- Listas de la ficha del curso: "Lo que aprenderás" y "Requisitos".
CREATE TABLE "educacion"."modulo_detalle" (
    "id_detalle"          SERIAL PRIMARY KEY,
    "id_modulo"           INT          NOT NULL
                              REFERENCES "educacion"."modulo_educativo"("id_modulo") ON DELETE RESTRICT,
    "tipo"                VARCHAR(20)  NOT NULL
                              CHECK (tipo IN ('aprendizaje','requisito')),
    "texto"               VARCHAR(255) NOT NULL,
    "orden"               INT,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_modulo_detalle_mod
    BEFORE UPDATE ON "educacion"."modulo_detalle"
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
    "slug"                VARCHAR(100) UNIQUE,                -- AJUSTE CLIENTE: id de la clase en la URL
    "titulo"              VARCHAR(200) NOT NULL,
    "tipo"                VARCHAR(20)  NOT NULL DEFAULT 'video'      -- AJUSTE CLIENTE
                              CHECK (tipo IN ('video','lectura','quiz')),
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

-- educacion.certificado                                     -- NUEVA (CLIENTE)
-- Certificado que se emite al terminar un curso con otorga_certificado = true.
-- codigo_verificacion es el "Código" impreso en el certificado.
CREATE TABLE "educacion"."certificado" (
    "id_certificado"      SERIAL PRIMARY KEY,
    "id_usuario"          INT          NOT NULL
                              REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_modulo"           INT          NOT NULL
                              REFERENCES "educacion"."modulo_educativo"("id_modulo") ON DELETE RESTRICT,
    "codigo_verificacion" VARCHAR(40)  UNIQUE NOT NULL,
    "fecha_emision"       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE("id_usuario", "id_modulo")
);

CREATE TRIGGER trg_certificado_mod
    BEFORE UPDATE ON "educacion"."certificado"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();


-- ============================================================
-- ESQUEMA: finanzas
-- Tablas: categoria, movimiento_ingreso_egreso, tipo_ingreso,
--         finanzas, tipo_inversion, nivel_riesgo,
--         movimiento_inversion, tipo_ingreso_inversion,
--         inversion, historial_inversion, editar_meta,
--         movimiento_meta, tipo_ingreso_meta, meta,
--         solicitud_consolidacion
-- (fusión de finanzas + inversion + metas)
-- ============================================================

-- finanzas.categoria
CREATE TABLE "finanzas"."categoria" (
    "id_categoria"        SERIAL PRIMARY KEY,
    "nombre"              VARCHAR(100) NOT NULL,
    "descripcion"         TEXT,
    "icono"               VARCHAR(10),                        -- AJUSTE CLIENTE: emoji que muestra el dashboard
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_categoria_mod
    BEFORE UPDATE ON "finanzas"."categoria"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.movimiento_ingreso_egreso
-- Es el "Libro mayor" del frontend: nombre = concepto, monto = valor,
-- es_ingreso = tipo ('ingreso' / 'gasto').
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
    "metodo_pago"           VARCHAR(50),                      -- AJUSTE CLIENTE: 'Efectivo', 'Tarjeta débito', 'Tarjeta crédito', 'Transferencia', 'Otro'
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
    "id_tipo_inversion"     INT                               -- AJUSTE CLIENTE: ya no es NOT NULL, el formulario no pide el tipo
                                REFERENCES "finanzas"."tipo_inversion"("id_tipo_inversion") ON DELETE RESTRICT,
    "id_nivel_riesgo"       INT           NOT NULL
                                REFERENCES "finanzas"."nivel_riesgo"("id_nivel_riesgo") ON DELETE RESTRICT,
    "id_movimiento_dinero"  INT           REFERENCES "finanzas"."movimiento_inversion"("id_movimiento_dinero") ON DELETE RESTRICT,
    "nombre"                VARCHAR(120),
    "monto"                 DECIMAL(18,2),
    "rentabilidad"          DECIMAL(8,4),
    "rendimiento"           DECIMAL(18,2),                    -- AJUSTE CLIENTE: ganancia o perdida en pesos (puede ser negativa)
    "duracion"              VARCHAR(30),                      -- AJUSTE CLIENTE: ej. '2 años', '6 meses', 'Flexible'
    "fecha_inicio"          DATE,
    "fecha_fin"             DATE,
    "activo"                BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"        TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"    TIMESTAMP     DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_inversion_mod
    BEFORE UPDATE ON "finanzas"."inversion"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- finanzas.historial_inversion                              -- NUEVA (CLIENTE)
-- Valor de cada inversion en una fecha. Alimenta las graficas
-- "Crecimiento inversión" (Finanzas) y de rendimiento por periodo (Inversiones).
CREATE TABLE "finanzas"."historial_inversion" (
    "id_historial"        SERIAL PRIMARY KEY,
    "id_inversion"        INT           NOT NULL
                              REFERENCES "finanzas"."inversion"("id_inversion") ON DELETE RESTRICT,
    "fecha"               DATE          NOT NULL,
    "valor"               DECIMAL(18,2) NOT NULL,
    "activo"              BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE("id_inversion", "fecha")
);

CREATE TRIGGER trg_historial_inversion_mod
    BEFORE UPDATE ON "finanzas"."historial_inversion"
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
-- El frontend calcula "porcentaje" y "cumplida" a partir de monto_actual y monto_objetivo.
CREATE TABLE "finanzas"."meta" (
    "id_meta"               SERIAL PRIMARY KEY,
    "id_usuario"            INT           NOT NULL
                                REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_editar_meta"        INT                               -- AJUSTE CLIENTE: ya no es NOT NULL, el formulario "Nueva meta" no lo usa
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
-- Tablas: banco, producto_crediticio, producto_detalle,
--         producto_tarifa, producto_pregunta, asesor_bancario,
--         contacto_asesor, lead, conversacion_usuario_asesor,
--         credito_desembolsado, transaccion_comision
-- (fusión de intermediacion + leads_schema + creditos + transacciones)
-- En el frontend (Alianzas) un banco es un "aliado" y un
-- producto_crediticio es una "oferta".
-- ============================================================

-- negocio.banco (antes intermediacion.banco)
CREATE TABLE "negocio"."banco" (
    "id_banco"              SERIAL PRIMARY KEY,
    "nombre_banco"          VARCHAR(100) UNIQUE NOT NULL,
    "tipo_aliado"           VARCHAR(20)  NOT NULL DEFAULT 'banco'     -- AJUSTE CLIENTE: filtro "Tipo" de Alianzas
                                CHECK (tipo_aliado IN ('banco','fintech','comercio')),
    "ciudad"                VARCHAR(100),
    "contacto"              VARCHAR(100),
    "telefono"              VARCHAR(20),
    "email"                 VARCHAR(150),
    "comision_porcentaje"   DECIMAL(5,2),
    "url_logo"              VARCHAR(500),
    "logo_texto"            VARCHAR(5),                       -- AJUSTE CLIENTE: sigla del logo, ej. 'BA'
    "logo_fondo"            VARCHAR(200),                     -- AJUSTE CLIENTE: degradado CSS del logo
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
-- nombre_producto es la "categoria" que muestra la tarjeta (ej. 'Crédito hipotecario')
-- y tasa_minima es el "Tasa desde".
CREATE TABLE "negocio"."producto_crediticio" (
    "id_producto"         SERIAL PRIMARY KEY,
    "id_banco"            INT           NOT NULL
                              REFERENCES "negocio"."banco"("id_banco") ON DELETE RESTRICT,
    "slug"                VARCHAR(80)   UNIQUE,               -- AJUSTE CLIENTE: id en la URL (/alianzas/producto/:productoId)
    "nombre_producto"     VARCHAR(150)  NOT NULL,
    "familia"             VARCHAR(20)                         -- AJUSTE CLIENTE: pestañas de Alianzas
                              CHECK (familia IN ('creditos','tarjetas','ahorro','comercios')),
    "beneficio"           VARCHAR(20)                         -- AJUSTE CLIENTE: filtro "Beneficio"
                              CHECK (beneficio IN ('cero_interes','cashback','sin_cuota')),
    "descripcion"         TEXT,
    "promesa"             VARCHAR(200),                       -- AJUSTE CLIENTE: frase principal de la ficha
    "vigencia"            VARCHAR(150),                       -- AJUSTE CLIENTE: texto pequeño de la tarjeta, ej. 'Vence en 7 días · 2.100 solicitudes'
    "monto_minimo"        DECIMAL(12,2),
    "monto_maximo"        DECIMAL(12,2),
    "tasa_minima"         DECIMAL(5,2),
    "tasa_maxima"         DECIMAL(5,2),
    "plazo_minimo"        INT,
    "plazo_maximo"        INT,
    "requisitos"          TEXT,                               -- resumen en texto; la lista va en producto_detalle
    "tiempo_aprobacion"   VARCHAR(50),                        -- AJUSTE CLIENTE: ej. '24 horas'
    "calificacion"        DECIMAL(2,1)                        -- AJUSTE CLIENTE: estrellas de 0 a 5
                              CHECK (calificacion BETWEEN 0 AND 5),
    "usuarios"            INT           NOT NULL DEFAULT 0,   -- AJUSTE CLIENTE: solicitudes / usuarios del producto
    "compatibilidad"      INT                                 -- AJUSTE CLIENTE: % de compatibilidad con el usuario
                              CHECK (compatibilidad BETWEEN 0 AND 100),
    "destacada"           BOOLEAN       NOT NULL DEFAULT false,      -- AJUSTE CLIENTE
    "cta_primaria"        VARCHAR(60),                        -- AJUSTE CLIENTE: texto del boton principal
    "cta_secundaria"      VARCHAR(60),                        -- AJUSTE CLIENTE
    "activo"              BOOLEAN       NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_producto_crediticio_mod
    BEFORE UPDATE ON "negocio"."producto_crediticio"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- negocio.producto_detalle                                  -- NUEVA (CLIENTE)
-- Listas de la ficha de cada oferta: etiquetas, requisitos, caracteristicas,
-- documentos y perfiles ('Score Alto', 'Historial Nuevo') del filtro "Perfil".
CREATE TABLE "negocio"."producto_detalle" (
    "id_detalle"          SERIAL PRIMARY KEY,
    "id_producto"         INT          NOT NULL
                              REFERENCES "negocio"."producto_crediticio"("id_producto") ON DELETE RESTRICT,
    "tipo"                VARCHAR(20)  NOT NULL
                              CHECK (tipo IN ('etiqueta','requisito','caracteristica','documento','perfil')),
    "texto"               VARCHAR(255) NOT NULL,
    "orden"               INT,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_producto_detalle_mod
    BEFORE UPDATE ON "negocio"."producto_detalle"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- negocio.producto_tarifa                                   -- NUEVA (CLIENTE)
-- Tabla "Tarifas" de la ficha de cada oferta.
CREATE TABLE "negocio"."producto_tarifa" (
    "id_tarifa"           SERIAL PRIMARY KEY,
    "id_producto"         INT          NOT NULL
                              REFERENCES "negocio"."producto_crediticio"("id_producto") ON DELETE RESTRICT,
    "concepto"            VARCHAR(150) NOT NULL,
    "valor"               VARCHAR(150) NOT NULL,              -- texto libre: 'Desde 14,9% E.A.', 'Sin costo'
    "orden"               INT,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_producto_tarifa_mod
    BEFORE UPDATE ON "negocio"."producto_tarifa"
    FOR EACH ROW EXECUTE FUNCTION fn_update_fecha_modificacion();

-- negocio.producto_pregunta                                 -- NUEVA (CLIENTE)
-- Preguntas frecuentes de la ficha de cada oferta.
CREATE TABLE "negocio"."producto_pregunta" (
    "id_pregunta"         SERIAL PRIMARY KEY,
    "id_producto"         INT          NOT NULL
                              REFERENCES "negocio"."producto_crediticio"("id_producto") ON DELETE RESTRICT,
    "pregunta"            VARCHAR(255) NOT NULL,
    "respuesta"           TEXT         NOT NULL,
    "orden"               INT,
    "activo"              BOOLEAN      NOT NULL DEFAULT true,
    "fecha_creacion"      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "fecha_modificacion"  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_producto_pregunta_mod
    BEFORE UPDATE ON "negocio"."producto_pregunta"
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
-- Es la "Solicitud" que el usuario radica desde la ficha de una oferta.
-- El frontend muestra estado_lead = 'nuevo' como "En estudio".
CREATE TABLE "negocio"."lead" (
    "id_lead"             SERIAL PRIMARY KEY,
    "id_usuario"          INT           NOT NULL
                              REFERENCES "auth"."usuario"("id_usuario") ON DELETE RESTRICT,
    "id_producto"         INT           NOT NULL
                              REFERENCES "negocio"."producto_crediticio"("id_producto") ON DELETE RESTRICT,
    "id_asesor"           INT
                              REFERENCES "negocio"."asesor_bancario"("id_asesor") ON DELETE RESTRICT,
    "radicado"            VARCHAR(20)   UNIQUE,               -- AJUSTE CLIENTE: numero que ve el usuario, ej. 'SOL-482913'
    "tipo_credito"        VARCHAR(100),
    "monto_interes"       DECIMAL(12,2),
    "plazo_interes"       INT,
    "nombre_solicitante"  VARCHAR(150),                       -- AJUSTE CLIENTE: datos del formulario de solicitud
    "documento_solicitante" VARCHAR(50),                      -- AJUSTE CLIENTE
    "correo_solicitante"  VARCHAR(150),                       -- AJUSTE CLIENTE
    "celular_solicitante" VARCHAR(20),                        -- AJUSTE CLIENTE
    "ingresos_mensuales"  DECIMAL(12,2),                      -- AJUSTE CLIENTE
    "acepta_terminos"     BOOLEAN       NOT NULL DEFAULT false,      -- AJUSTE CLIENTE
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
CREATE INDEX idx_codigo_usuario     ON "auth"."codigo_verificacion"("id_usuario", "proposito");

CREATE INDEX idx_pqr_usuario        ON "soporte"."pqr"("id_usuario");
CREATE INDEX idx_respuesta_pqr      ON "soporte"."respuesta_pqr"("id_pqr");
CREATE INDEX idx_faq_categoria      ON "soporte"."pregunta_frecuente"("categoria");

CREATE INDEX idx_progreso_usuario   ON "educacion"."progreso_educativo"("id_usuario");
CREATE INDEX idx_modulo_escuela     ON "educacion"."modulo_educativo"("id_escuela");
CREATE INDEX idx_modulo_detalle     ON "educacion"."modulo_detalle"("id_modulo");
CREATE INDEX idx_leccion_modulo     ON "educacion"."leccion"("id_modulo");
CREATE INDEX idx_certificado_usuario ON "educacion"."certificado"("id_usuario");

CREATE INDEX idx_meta_usuario       ON "finanzas"."meta"("id_usuario");
CREATE INDEX idx_movimiento_usuario ON "finanzas"."movimiento_ingreso_egreso"("id_usuario");
CREATE INDEX idx_solicitud_usuario  ON "finanzas"."solicitud_consolidacion"("id_usuario");
CREATE INDEX idx_inversion_usuario  ON "finanzas"."inversion"("id_usuario");

CREATE INDEX idx_producto_detalle   ON "negocio"."producto_detalle"("id_producto");
CREATE INDEX idx_producto_tarifa    ON "negocio"."producto_tarifa"("id_producto");
CREATE INDEX idx_producto_pregunta  ON "negocio"."producto_pregunta"("id_producto");
CREATE INDEX idx_lead_usuario       ON "negocio"."lead"("id_usuario");
CREATE INDEX idx_lead_estado        ON "negocio"."lead"("estado_lead");
CREATE INDEX idx_credito_usuario    ON "negocio"."credito_desembolsado"("id_usuario");
CREATE INDEX idx_credito_estado     ON "negocio"."credito_desembolsado"("estado_credito");
