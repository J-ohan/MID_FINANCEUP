-- ============================================================
-- FINANCEUP - DATOS DE PRUEBA
-- Ejecutar conectado a "financeup" DESPUES de 02_crear_tablas.sql.
-- Todos los usuarios de prueba tienen la clave: Financeup2026*
-- ============================================================

-- VERSION UNIFICADA (MID + ajustes del frontend)
--   Los datos originales del MID estan intactos: ninguna fila se borro
--   ni se le cambio un valor.
--   Todo lo agregado esta marcado con "-- NUEVO":
--     * filas nuevas (INSERT), y
--     * valores para las columnas NUEVAS de filas que ya existian (UPDATE
--       que solo toca esas columnas nuevas).
--   Este archivo tiene tildes y emojis: guardalo y abrelo como UTF-8.
SET client_encoding = 'UTF8';

-- ============================================================
-- ESQUEMA: auth
-- ============================================================

-- auth.rol
INSERT INTO "auth"."rol" ("nombre_rol", "descripcion", "activo") VALUES
('Administrador', 'Usuario con acceso total al sistema', true),
('Asesor',        'Asesor bancario que gestiona leads',  true),
('Usuario',       'Usuario estandar del sistema',        true),
('Gerente',       'Gerente de operaciones',              true);

-- auth.tipo_documento
INSERT INTO "auth"."tipo_documento" ("nombre", "codigo", "activo") VALUES
('Cedula de Ciudadania',  'CC',  true),
('Cedula de Extranjeria', 'CE',  true),
('Pasaporte',             'PA',  true),
('NIT',                   'NIT', true);

-- auth.usuario
INSERT INTO "auth"."usuario"
    ("tipo_documento", "nombre", "apellido", "email", "telefono", "cedula", "ciudad", "direccion", "fecha_nacimiento", "estado", "activo")
VALUES
(1, 'harold',  'arciniegas', 'harold.arciniegas@email.com', '3001234567', '1234567890', 'Bogota',       'Calle 10 # 5-20',   '1998-03-14', 'activo', true),
(1, 'fabio',   'zorro',      'fabio.zorro@email.com',       '3109876543', '9876543210', 'Medellin',     'Carrera 45 # 30-12', '1999-07-22', 'activo', true),
(2, 'johan',   'barreto',    'johan.barreto@email.com',     '3201234567', '5555555555', 'Cali',         'Avenida 6 # 15-40',  '2000-11-05', 'activo', true),
(1, 'sharith', 'bermudez',   'sharith.bermudez@email.com',  '3051234567', '4444444444', 'Cartagena',    'Calle 30 # 8-15',    '2001-01-30', 'activo', true),
(3, 'fabian',  'barreto',    'fabian.barreto@email.com',    '3161234567', '6666666666', 'Barranquilla', 'Carrera 50 # 72-10', '1997-09-18', 'activo', true);

-- NUEVO: columnas nuevas de los usuarios de arriba
UPDATE "auth"."usuario" SET "acepta_terminos" = true, "fecha_aceptacion_terminos" = CURRENT_TIMESTAMP - INTERVAL '90 days' WHERE "id_usuario" = 1;
UPDATE "auth"."usuario" SET "acepta_terminos" = true, "fecha_aceptacion_terminos" = CURRENT_TIMESTAMP - INTERVAL '80 days' WHERE "id_usuario" = 2;
UPDATE "auth"."usuario" SET "acepta_terminos" = true, "fecha_aceptacion_terminos" = CURRENT_TIMESTAMP - INTERVAL '70 days' WHERE "id_usuario" = 3;
UPDATE "auth"."usuario" SET "acepta_terminos" = true, "fecha_aceptacion_terminos" = CURRENT_TIMESTAMP - INTERVAL '60 days' WHERE "id_usuario" = 4;
UPDATE "auth"."usuario" SET "acepta_terminos" = true, "fecha_aceptacion_terminos" = CURRENT_TIMESTAMP - INTERVAL '50 days' WHERE "id_usuario" = 5;


-- auth.credencial
-- Hashes bcrypt reales de la clave "Financeup2026*".
-- El salt de bcrypt va dentro del hash (caracteres 8 a 29); se copia en la columna salt.
INSERT INTO "auth"."credencial"
    ("id_usuario", "contrasena_hash", "salt", "algoritmo", "intentos_fallidos", "requiere_cambio", "activo")
VALUES
(1, '$2a$10$nmktU8G5s5XMaON.h.y4O.VIWT2I4lXXR/7HfizLNXsHk9F9f85lq', 'nmktU8G5s5XMaON.h.y4O.', 'bcrypt', 0, false, true),
(2, '$2a$10$txgMfIq5TTvdo10Ysmk8o.OPxHkCGyTiA6LDwceI.I0YlEmD4AY2y', 'txgMfIq5TTvdo10Ysmk8o.', 'bcrypt', 0, false, true),
(3, '$2a$10$HWdPYbYZjCxdXAe29kBxVOTeaW8Igx3DuSJBdxxQhlXlURUj92c4m', 'HWdPYbYZjCxdXAe29kBxVO', 'bcrypt', 1, false, true),
(4, '$2a$10$SQwQ/BYFXUhQ2M.2zD4tC.4BKPNgbugk4dN6963GASwfZEGQqh6YG', 'SQwQ/BYFXUhQ2M.2zD4tC.', 'bcrypt', 0, false, true),
(5, '$2a$10$05AzCcB6GzRV9tZhZL6ZyusAhbfEFJ7gsKilx7vCRYFpgOzSpjTLy', '05AzCcB6GzRV9tZhZL6Zyu', 'bcrypt', 0, false, true);

-- auth.usuario_rol
INSERT INTO "auth"."usuario_rol" ("id_usuario", "id_rol", "activo") VALUES
(1, 1, true),
(2, 3, true),
(3, 2, true),
(4, 3, true),
(5, 4, true);

-- auth.auditoria_login
INSERT INTO "auth"."auditoria_login"
    ("id_usuario", "tipo_evento", "ip_address", "navegador", "estado_evento")
VALUES
(1, 'login',            '192.168.1.1', 'Chrome 120',  'exitoso'),
(2, 'login',            '192.168.1.2', 'Firefox 121', 'exitoso'),
(3, 'login',            '192.168.1.3', 'Safari 17',   'exitoso'),
(1, 'cambio_contrasena','192.168.1.1', 'Chrome 120',  'exitoso'),
(4, 'login',            '192.168.1.4', 'Edge 121',    'exitoso');

-- NUEVO: auth.cuenta_social
-- Cuentas de Google y Facebook vinculadas a usuarios que tambien tienen clave.
INSERT INTO "auth"."cuenta_social"
    ("id_usuario", "proveedor", "id_externo", "email_proveedor", "fecha_vinculacion", "activo")
VALUES
(4, 'google',   '108274615203948571234', 'sharith.bermudez@email.com', CURRENT_TIMESTAMP - INTERVAL '30 days', true),
(2, 'facebook', '10229384756102938',     'fabio.zorro@email.com',      CURRENT_TIMESTAMP - INTERVAL '12 days', true);

-- NUEVO: auth.codigo_verificacion
-- Sin datos de prueba: el backend crea un codigo (y guarda su hash) cada vez
-- que alguien inicia sesion o pide recuperar la contrasena.



-- ============================================================
-- ESQUEMA: soporte
-- (antes pqr_schema + auditoria)
-- ============================================================

-- soporte.estado_pqr
INSERT INTO "soporte"."estado_pqr" ("nombre", "descripcion", "activo") VALUES
('Abierta',     'Peticion recientemente registrada',  true),
('En Revision', 'Siendo revisada por el equipo',      true),
('Resuelta',    'Peticion solucionada',               true),
('Cerrada',     'Peticion archivada',                 true);

-- NUEVO: estado que usa el frontend y no existia (id 5)
INSERT INTO "soporte"."estado_pqr" ("nombre", "descripcion", "activo") VALUES
('Rechazada',   'La solicitud no procede',            true);


-- soporte.pqr
INSERT INTO "soporte"."pqr"
    ("id_usuario", "radicado", "titulo", "tipo", "categoria", "prioridad", "asesor", "mensaje_respuesta", "descripcion", "id_estado", "activo")
VALUES
(1, 'PQR-2026-0001', 'No puedo ingresar',        'peticion', 'Plataforma',            'alta',  'Soporte tecnico', NULL,                                          'Problema con acceso a la plataforma',     1, true),
(2, 'PQR-2026-0002', 'Duda sobre tasas',         'peticion', 'Creditos',              'media', 'Camila Rojas',    NULL,                                          'Consulta sobre tasas de interes',         2, true),
(3, 'PQR-2026-0003', 'Cobro de comision',        'reclamo',  'Movimientos y pagos',   'alta',  NULL,              NULL,                                          'Reclamo por comision no autorizada',      1, true),
(4, 'PQR-2026-0004', 'Informacion de creditos',  'peticion', 'Atencion al cliente',   'baja',  'Andres Pena',     'Le enviamos al correo el portafolio vigente.', 'Solicitud de informacion sobre creditos', 3, true);

-- NUEVO: columna progreso (barra de avance) de las PQR de arriba
UPDATE "soporte"."pqr" SET "progreso" = 10  WHERE "id_pqr" = 1;
UPDATE "soporte"."pqr" SET "progreso" = 60  WHERE "id_pqr" = 2;
UPDATE "soporte"."pqr" SET "progreso" = 10  WHERE "id_pqr" = 3;
UPDATE "soporte"."pqr" SET "progreso" = 100 WHERE "id_pqr" = 4;


-- soporte.adjunto
INSERT INTO "soporte"."adjunto"
    ("id_pqr", "nombre_archivo", "ruta_archivo", "tipo_mime", "tamano_bytes", "activo")
VALUES
(1, 'error_screenshot.png',       '/uploads/pqr/error_screenshot.png',       'image/png',       512000, true),
(2, 'consulta_documento.pdf',     '/uploads/pqr/consulta_documento.pdf',     'application/pdf', 256000, true),
(3, 'comprobante_comision.pdf',   '/uploads/pqr/comprobante_comision.pdf',   'application/pdf', 384000, true);

-- NUEVO: soporte.respuesta_pqr
INSERT INTO "soporte"."respuesta_pqr"
    ("id_pqr", "id_usuario", "autor", "mensaje", "fecha_respuesta", "activo")
VALUES
(2, NULL, 'Camila Rojas',     'Hola, recibimos tu consulta sobre tasas y ya la asignamos a un asesor.',                     CURRENT_TIMESTAMP - INTERVAL '3 days', true),
(2, NULL, 'Camila Rojas',     'Estamos revisando tu caso, en caso de alguna novedad se te notificará por este mismo medio.', CURRENT_TIMESTAMP - INTERVAL '2 days', true),
(4, NULL, 'Andres Pena',      'Recibimos tu solicitud de información sobre créditos.',                                     CURRENT_TIMESTAMP - INTERVAL '7 days', true),
(4, 4,    'Sharith Bermudez', '¿Me pueden enviar el portafolio al correo?',                                                 CURRENT_TIMESTAMP - INTERVAL '6 days', true),
(4, NULL, 'Andres Pena',      'Le enviamos al correo el portafolio vigente.',                                               CURRENT_TIMESTAMP - INTERVAL '5 days', true);


-- soporte.registro_actividad (antes auditoria.registro_actividad)
INSERT INTO "soporte"."registro_actividad"
    ("id_usuario", "tipo_actividad", "descripcion", "entidad_afectada", "fecha_actividad")
VALUES
(1, 'LOGIN',            'Usuario inicio sesion',                   'auth.usuario',       CURRENT_TIMESTAMP - INTERVAL '2 hours'),
(1, 'CREAR_SOLICITUD',  'Se creo una nueva solicitud de credito',  'negocio.lead',       CURRENT_TIMESTAMP - INTERVAL '60 days'),
(2, 'ACTUALIZAR_PERFIL','Se actualizo el perfil de usuario',       'auth.usuario',       CURRENT_TIMESTAMP - INTERVAL '10 days'),
(3, 'CREAR_SIMULACION', 'Se ejecuto una simulacion financiera',    'finanzas.finanzas',  CURRENT_TIMESTAMP - INTERVAL '10 days'),
(4, 'LOGIN',            'Usuario inicio sesion',                   'auth.usuario',       CURRENT_TIMESTAMP - INTERVAL '1 hour');

-- NUEVO: soporte.pregunta_frecuente (pagina "Linea de ayuda")
INSERT INTO "soporte"."pregunta_frecuente"
    ("categoria", "pregunta", "respuesta", "lectura_minutos", "utilidad", "orden", "activo")
VALUES
('Cuenta',    'Olvidé mi contraseña, ¿qué debo hacer?',      'Ingresa a la pantalla de inicio de sesión y selecciona "¿Olvidaste tu contraseña?". Te enviaremos un enlace a tu correo registrado para que puedas restablecerla.', 1, 96, 1, true),
('Seguridad', '¿La información financiera que veo es segura?', 'Sí. Toda la información se transmite de forma cifrada y cumplimos con los estándares de seguridad exigidos para el manejo de datos financieros.',                   2, 93, 2, true),
('Educación', '¿Qué tipo de cursos ofrece Finance Up?',       'Ofrecemos cursos de educación financiera, inversión, ahorro y planeación presupuestal, tanto para principiantes como para usuarios avanzados.',                       2, 89, 3, true),
('Alianzas',  '¿Qué son las alianzas financieras?',           'Son acuerdos con entidades del sector financiero que nos permiten ofrecerte beneficios, tasas preferenciales y productos exclusivos.',                                1, 91, 4, true),
('PQR',       '¿Qué es una PQR?',                             'Una PQR es una Petición, Queja o Reclamo que puedes radicar cuando necesitas reportar un problema o solicitar información sobre nuestros servicios.',                1, 94, 5, true),
('PQR',       '¿Cuánto tarda la respuesta de una PQR?',       'El tiempo máximo de respuesta es de 48 horas hábiles. Puedes hacer seguimiento del avance desde la sección "Mis PQR".',                                              1, 88, 6, true);



-- ============================================================
-- ESQUEMA: educacion
-- ============================================================

-- NUEVO: educacion.escuela
INSERT INTO "educacion"."escuela" ("slug", "nombre", "descripcion", "orden", "activo") VALUES
('finanzas-personales', 'Finanzas Personales', 'Ordena tu plata mes a mes: presupuesto, crédito y deuda.', 1, true),
('inversion',           'Inversión',           'Del primer fondo a un portafolio que puedas sostener.',    2, true);

-- NUEVO: educacion.instructor
INSERT INTO "educacion"."instructor" ("nombre", "cargo", "iniciales", "activo") VALUES
('Mariana Escobar Uribe',  'Asesora financiera, 9 años en banca personal', 'ME', true),
('Julián Restrepo Vélez',  'Ex analista de riesgo crediticio',             'JR', true),
('Daniela Cifuentes Mora', 'Educadora financiera',                         'DC', true),
('Óscar Bermúdez Lara',    'Gestor de portafolios',                        'OB', true);

-- educacion.modulo_educativo
INSERT INTO "educacion"."modulo_educativo"
    ("titulo", "descripcion", "contenido", "nivel", "url_thumbnail", "activo")
VALUES
('Introduccion a Finanzas Personales', 'Conceptos basicos de finanzas personales',  'Contenido teorico basico',              'basico',      'https://example.com/thumb1.jpg', true),
('Creditos y Deudas',                  'Como gestionar creditos y deudas',           'Contenido intermedio sobre gestion',    'intermedio',  'https://example.com/thumb2.jpg', true),
('Inversiones Avanzadas',              'Estrategias avanzadas de inversion',         'Contenido avanzado de portafolio',      'avanzado',    'https://example.com/thumb3.jpg', true),
('Ahorro e Inversion',                 'Tecnicas de ahorro e inversion',             'Contenido sobre ahorro',                'basico',      'https://example.com/thumb4.jpg', true);

-- NUEVO: columnas nuevas de los 4 modulos de arriba
UPDATE "educacion"."modulo_educativo" SET "id_escuela" = 1, "slug" = 'introduccion-finanzas-personales', "categoria" = 'fundamentos', "otorga_certificado" = true WHERE "id_modulo" = 1;
UPDATE "educacion"."modulo_educativo" SET "id_escuela" = 1, "slug" = 'creditos-y-deudas',                "categoria" = 'deuda'                                   WHERE "id_modulo" = 2;
UPDATE "educacion"."modulo_educativo" SET "id_escuela" = 2, "slug" = 'inversiones-avanzadas',            "categoria" = 'fundamentos'                             WHERE "id_modulo" = 3;
UPDATE "educacion"."modulo_educativo" SET "id_escuela" = 2, "slug" = 'ahorro-e-inversion',               "categoria" = 'ahorro'                                  WHERE "id_modulo" = 4;

-- NUEVO: cursos que muestra hoy el frontend (quedan con id_modulo 5 a 11)
INSERT INTO "educacion"."modulo_educativo"
    ("id_escuela", "id_instructor", "slug", "titulo", "descripcion", "contenido", "nivel", "categoria", "formato",
     "duracion", "resumen_lecciones", "otorga_certificado", "estudiantes", "calificacion", "url_thumbnail", "activo")
VALUES
(1, 1, 'presupuesto-50-30-20', 'Presupuesto 50/30/20',                      'Reparte tu sueldo entre necesidades, gustos y ahorro sin llevar una hoja de cálculo interminable.',   NULL, 'basico',     'fundamentos', 'video',    '30 minutos', '4 lecciones + evaluación', true,  4812, 4.7, '/assets/cursos/presupuesto.svg', true),
(1, 2, 'tarjetas-responsable', 'Usa bien tu tarjeta de crédito',            'Entiende el cupo, la cuota mínima y los intereses antes de que ellos te entiendan a ti.',             NULL, 'basico',     'credito',     'video',    '35 minutos', '4 lecciones + evaluación', true,  6390, 4.8, '/assets/cursos/tarjetas.svg',    true),
(1, 2, 'historial-desde-cero', 'Entiende tu puntaje de crédito',            'Qué mira una entidad cuando te consulta en centrales de riesgo y cómo subir tu score.',              NULL, 'basico',     'credito',     'video',    '45 minutos', '4 lecciones + evaluación', true,  8145, 4.9, '/assets/cursos/historial.svg',   true),
(1, 3, 'salir-de-deudas',      'Sal de deudas con el método bola de nieve', 'Ordena tus deudas y decide cuál pagar primero cuando el dinero no alcanza para todas.',               NULL, 'intermedio', 'deuda',       'articulo', '28 minutos', '3 lecciones + evaluación', false, 3127, 4.6, '/assets/cursos/deudas.svg',      true),
(2, 4, 'intro-bolsa',          'Primeros pasos en la bolsa',                'Qué se compra y se vende en una bolsa de valores, y cómo abrir tu primera cuenta en Colombia.',       NULL, 'basico',     'fundamentos', 'video',    '35 minutos', '3 lecciones + evaluación', true,  5674, 4.7, '/assets/cursos/bolsa.svg',       true),
(2, 4, 'intro-criptomonedas',  'Criptomonedas sin humo',                    'Qué hay detrás de una criptomoneda, qué riesgos asumes y cómo evitar las estafas más comunes.',      NULL, 'basico',     'fundamentos', 'video',    '40 minutos', '3 lecciones + evaluación', true,  7203, 4.4, '/assets/cursos/cripto.svg',      true),
(2, 4, 'gestion-portafolio',   'Arma y rebalancea tu portafolio',           'Cómo repartir lo que inviertes y cada cuánto volver a ajustar los pesos.',                           NULL, 'avanzado',   'fundamentos', 'quiz',     '30 minutos', '3 lecciones + evaluación', true,  1948, 4.8, '/assets/cursos/portafolio.svg',  true);

-- NUEVO: educacion.modulo_detalle ("Lo que aprenderas" y "Requisitos")
-- de los cursos 5 a 11
INSERT INTO "educacion"."modulo_detalle" ("id_modulo", "tipo", "texto", "orden", "activo") VALUES
(5, 'aprendizaje', 'Repartir tu ingreso mensual en tres bolsillos',            1, true),
(5, 'aprendizaje', 'Distinguir un gasto fijo de uno que puedes recortar',      2, true),
(5, 'aprendizaje', 'Ajustar la regla cuando tu ingreso es variable',           3, true),
(5, 'requisito',   'Conocer tu ingreso mensual aproximado',                    1, true),
(6, 'aprendizaje', 'Leer un extracto sin perderte',                            1, true),
(6, 'aprendizaje', 'Calcular cuánto te cuesta pagar la cuota mínima',          2, true),
(6, 'aprendizaje', 'Aprovechar el período de gracia a tu favor',               3, true),
(6, 'requisito',   'Tener o estar por pedir una tarjeta de crédito',           1, true),
(7, 'aprendizaje', 'Qué pesa y qué no pesa en tu score',                       1, true),
(7, 'aprendizaje', 'Cada cuánto conviene consultar tu historial',              2, true),
(7, 'aprendizaje', 'Cómo recuperarte de un reporte negativo',                  3, true),
(7, 'requisito',   'Ninguno, puedes empezar sin historial crediticio',         1, true),
(8, 'aprendizaje', 'Inventariar todo lo que debes en una sola tabla',          1, true),
(8, 'aprendizaje', 'Elegir entre bola de nieve y avalancha',                   2, true),
(8, 'aprendizaje', 'Negociar un acuerdo de pago sin miedo',                    3, true),
(8, 'requisito',   'Tener al menos una deuda activa',                          1, true),
(9, 'aprendizaje', 'Diferenciar acción, bono y fondo',                         1, true),
(9, 'aprendizaje', 'Abrir cuenta en una comisionista',                         2, true),
(9, 'aprendizaje', 'Calcular cuánto te cobran por operar',                     3, true),
(9, 'requisito',   'Tener un fondo de emergencia armado',                      1, true),
(10, 'aprendizaje', 'Explicar qué es una blockchain sin tecnicismos',           1, true),
(10, 'aprendizaje', 'Reconocer una estafa piramidal disfrazada',                2, true),
(10, 'aprendizaje', 'Custodiar tus llaves de forma segura',                     3, true),
(10, 'requisito',   'Ninguno',                                                  1, true),
(11, 'aprendizaje', 'Definir pesos objetivo por tipo de activo',                1, true),
(11, 'aprendizaje', 'Rebalancear sin disparar impuestos ni comisiones',         2, true),
(11, 'aprendizaje', 'Medir tu portafolio con tres métricas',                    3, true),
(11, 'requisito',   'Haber invertido antes',                                    1, true),
(11, 'requisito',   'Curso de bolsa completado',                                2, true);


-- educacion.contenido
INSERT INTO "educacion"."contenido"
    ("titulo", "descripcion", "duracion_minutos", "url_video", "activo")
VALUES
('Video Introduccion Finanzas', 'Video introductorio',    15, 'https://example.com/videos/intro1.mp4',       true),
('Video Creditos Explicado',    'Explicacion creditos',   20, 'https://example.com/videos/creditos1.mp4',    true),
('Video Inversiones',           'Guia de inversiones',    25, 'https://example.com/videos/inversiones1.mp4', true),
('Video Ahorro Basico',         'Conceptos de ahorro',    18, 'https://example.com/videos/ahorro1.mp4',      true);

-- educacion.leccion
INSERT INTO "educacion"."leccion"
    ("id_modulo", "id_contenido", "titulo", "descripcion", "duracion_minutos", "url_video", "numero_leccion", "activo")
VALUES
(1, 1,    'Leccion 1: Conceptos Basicos',    'Primera leccion del modulo',   30, 'https://example.com/leccion1.mp4', 1, true),
(1, NULL, 'Leccion 2: Presupuesto Personal', 'Segunda leccion del modulo',   35, 'https://example.com/leccion2.mp4', 2, true),
(2, 2,    'Leccion 1: Tipos de Credito',     'Primer tema de creditos',      40, 'https://example.com/leccion3.mp4', 1, true),
(3, 3,    'Leccion 1: Analisis de Riesgo',   'Analisis avanzado',            45, 'https://example.com/leccion4.mp4', 1, true);

-- NUEVO: slug de las 4 lecciones de arriba (tipo queda en 'video')
UPDATE "educacion"."leccion" SET "slug" = 'introduccion-finanzas-personales-l1' WHERE "id_leccion" = 1;
UPDATE "educacion"."leccion" SET "slug" = 'introduccion-finanzas-personales-l2' WHERE "id_leccion" = 2;
UPDATE "educacion"."leccion" SET "slug" = 'creditos-y-deudas-l1'                WHERE "id_leccion" = 3;
UPDATE "educacion"."leccion" SET "slug" = 'inversiones-avanzadas-l1'            WHERE "id_leccion" = 4;

-- NUEVO: lecciones de los cursos 5 a 11 (quedan con id_leccion 5 a 35)
INSERT INTO "educacion"."leccion"
    ("id_modulo", "id_contenido", "slug", "titulo", "tipo", "descripcion", "duracion_minutos", "url_video", "numero_leccion", "activo")
VALUES
(5, 1,    'presupuesto-50-30-20-l1', 'Qué resuelve la regla 50/30/20',             'video',   'De dónde sale la regla y por qué funciona mejor que anotar cada gasto.',                     6,  NULL, 1, true),
(5, 4,    'presupuesto-50-30-20-l2', 'Clasifica tus gastos en tres bolsillos',     'video',   'Necesidades, gustos y ahorro: dónde entra el arriendo, el plan de datos y las salidas.',     8,  NULL, 2, true),
(5, NULL, 'presupuesto-50-30-20-l3', 'Los tres errores que rompen el presupuesto', 'lectura', 'Presupuestar sobre el bruto, olvidar los gastos anuales y no dejar colchón.',               5,  NULL, 3, true),
(5, NULL, 'presupuesto-50-30-20-l4', 'Arma tu primer presupuesto',                 'video',   'Aterrizamos la regla sobre un sueldo real de 2.800.000 al mes.',                             9,  NULL, 4, true),
(5, NULL, 'presupuesto-50-30-20-l5', 'Evaluación final',                           'quiz',    'Cinco preguntas para confirmar que puedes repartir tu ingreso solo.',                        4,  NULL, 5, true),
(6, 2,    'tarjetas-responsable-l1', 'Cómo funciona una tarjeta por dentro',       'video',   'Cupo, fecha de corte y fecha límite de pago explicados sobre un extracto real.',             7,  NULL, 1, true),
(6, NULL, 'tarjetas-responsable-l2', 'El costo real de la cuota mínima',           'video',   'Qué pasa con una deuda de 1.200.000 si solo pagas el mínimo durante un año.',                9,  NULL, 2, true),
(6, NULL, 'tarjetas-responsable-l3', 'Diferir a cuotas sin perder plata',          'lectura', 'Cuándo diferir sale barato y cuándo es mejor pagar de contado.',                             6,  NULL, 3, true),
(6, NULL, 'tarjetas-responsable-l4', 'Señales de que la tarjeta te está ganando',  'video',   'Los cuatro comportamientos que anteceden a un sobreendeudamiento.',                          8,  NULL, 4, true),
(6, NULL, 'tarjetas-responsable-l5', 'Evaluación final',                           'quiz',    'Pon a prueba lo que aprendiste sobre intereses y fechas de corte.',                          5,  NULL, 5, true),
(7, NULL, 'historial-desde-cero-l1', 'Qué es el puntaje y quién lo calcula',       'video',   'Datacrédito, TransUnion y de dónde salen los datos que reportan sobre ti.',                  8,  NULL, 1, true),
(7, NULL, 'historial-desde-cero-l2', 'Los cinco factores que mueven tu score',     'video',   'Historial de pago, uso del cupo, antigüedad, mezcla de productos y consultas.',              11, NULL, 2, true),
(7, NULL, 'historial-desde-cero-l3', 'Mitos que circulan sobre el crédito',        'lectura', 'Consultar tu score no lo baja, y cerrar tarjetas no siempre ayuda.',                         7,  NULL, 3, true),
(7, NULL, 'historial-desde-cero-l4', 'Plan de 6 meses para subir tu puntaje',      'video',   'Qué hacer mes a mes si vienes de un reporte negativo.',                                      12, NULL, 4, true),
(7, NULL, 'historial-desde-cero-l5', 'Evaluación final',                           'quiz',    'Confirma que sabes leer tu historial antes de pedir un crédito.',                            5,  NULL, 5, true),
(8, NULL, 'salir-de-deudas-l1',      'Inventario: cuánto debes de verdad',         'lectura', 'Saldo, tasa y cuota de cada deuda en una sola tabla.',                                       8,  NULL, 1, true),
(8, NULL, 'salir-de-deudas-l2',      'Bola de nieve contra avalancha',             'lectura', 'Una da motivación temprana, la otra ahorra más intereses. Cuál te sirve.',                   9,  NULL, 2, true),
(8, NULL, 'salir-de-deudas-l3',      'Cómo negociar con la entidad',               'lectura', 'Qué pedir, con quién hablar y qué dejar por escrito.',                                       7,  NULL, 3, true),
(8, NULL, 'salir-de-deudas-l4',      'Evaluación final',                           'quiz',    'Arma el orden de pago de un caso con cuatro deudas.',                                        4,  NULL, 4, true),
(9, 3,    'intro-bolsa-l1',          'Qué es y para qué sirve una bolsa',          'video',   'Quién emite, quién compra y qué papel juega la comisionista.',                               9,  NULL, 1, true),
(9, NULL, 'intro-bolsa-l2',          'Acciones, bonos y fondos',                   'video',   'Tres instrumentos, tres niveles de riesgo, tres horizontes de tiempo.',                      11, NULL, 2, true),
(9, NULL, 'intro-bolsa-l3',          'Abre tu primera cuenta',                     'lectura', 'Documentos, montos mínimos y comisiones de las casas más usadas.',                           10, NULL, 3, true),
(9, NULL, 'intro-bolsa-l4',          'Evaluación final',                           'quiz',    'Identifica qué instrumento encaja en cada objetivo.',                                        5,  NULL, 4, true),
(10, NULL, 'intro-criptomonedas-l1',  'Qué es una criptomoneda',                    'video',   'Por qué existe, quién la emite y qué la respalda.',                                          10, NULL, 1, true),
(10, NULL, 'intro-criptomonedas-l2',  'Blockchain explicada con un ejemplo',        'video',   'Un libro contable que nadie puede borrar, contado con una cuenta de tienda.',                12, NULL, 2, true),
(10, NULL, 'intro-criptomonedas-l3',  'Riesgos y estafas frecuentes',               'lectura', 'Rendimientos garantizados, referidos obligatorios y otras banderas rojas.',                  13, NULL, 3, true),
(10, NULL, 'intro-criptomonedas-l4',  'Evaluación final',                           'quiz',    'Detecta las banderas rojas en tres casos reales.',                                           5,  NULL, 4, true),
(11, NULL, 'gestion-portafolio-l1',   'Pesos objetivo según tu horizonte',          'lectura', 'Cómo cambia la mezcla si tu meta está a 3, 10 o 25 años.',                                   10, NULL, 1, true),
(11, NULL, 'gestion-portafolio-l2',   'Cuándo y cómo rebalancear',                  'video',   'Por calendario o por desviación: dos disciplinas para no improvisar.',                       11, NULL, 2, true),
(11, NULL, 'gestion-portafolio-l3',   'Tres métricas para hacer seguimiento',       'lectura', 'Rentabilidad real, volatilidad y máxima caída.',                                             6,  NULL, 3, true),
(11, NULL, 'gestion-portafolio-l4',   'Evaluación final',                           'quiz',    'Rebalancea un portafolio desviado 12 puntos de su objetivo.',                                5,  NULL, 4, true);


-- educacion.progreso_educativo
INSERT INTO "educacion"."progreso_educativo"
    ("id_usuario", "id_modulo", "porcentaje_completado", "fecha_inicio", "fecha_completado", "calificacion", "activo")
VALUES
(1, 1, 100, CURRENT_TIMESTAMP - INTERVAL '30 days', CURRENT_TIMESTAMP - INTERVAL '5 days',  95,   true),
(2, 1, 50,  CURRENT_TIMESTAMP - INTERVAL '20 days', NULL,                                    NULL, true),
(3, 2, 75,  CURRENT_TIMESTAMP - INTERVAL '15 days', NULL,                                    NULL, true),
(4, 1, 100, CURRENT_TIMESTAMP - INTERVAL '10 days', CURRENT_TIMESTAMP - INTERVAL '2 days',  88,   true);

-- educacion.progreso_leccion
INSERT INTO "educacion"."progreso_leccion"
    ("id_usuario", "id_leccion", "completado", "fecha_inicio", "fecha_completado", "activo")
VALUES
(1, 1, true, CURRENT_TIMESTAMP - INTERVAL '30 days', CURRENT_TIMESTAMP - INTERVAL '28 days', true),
(1, 2, true, CURRENT_TIMESTAMP - INTERVAL '27 days', CURRENT_TIMESTAMP - INTERVAL '25 days', true),
(2, 1, true, CURRENT_TIMESTAMP - INTERVAL '20 days', CURRENT_TIMESTAMP - INTERVAL '19 days', true),
(3, 3, true, CURRENT_TIMESTAMP - INTERVAL '15 days', CURRENT_TIMESTAMP - INTERVAL '14 days', true);

-- NUEVO: educacion.certificado
-- Los usuarios 1 y 4 terminaron el modulo 1 al 100% (ver progreso_educativo).
INSERT INTO "educacion"."certificado"
    ("id_usuario", "id_modulo", "codigo_verificacion", "fecha_emision", "activo")
VALUES
(1, 1, 'FU-INTRO-0001-0001', CURRENT_TIMESTAMP - INTERVAL '5 days', true),
(4, 1, 'FU-INTRO-0001-0004', CURRENT_TIMESTAMP - INTERVAL '2 days', true);



-- ============================================================
-- ESQUEMA: finanzas
-- (antes finanzas + inversion + metas)
-- ============================================================

-- finanzas.categoria
INSERT INTO "finanzas"."categoria" ("nombre", "descripcion", "activo") VALUES
('Arriendo',      'Gastos de vivienda',    true),
('Servicios',     'Servicios publicos',    true),
('Alimentacion',  'Gastos en alimentos',  true),
('Transporte',    'Gastos en transporte',  true),
('Salud',         'Gastos medicos',        true);

-- NUEVO: icono de las 5 categorias de arriba
UPDATE "finanzas"."categoria" SET "icono" = '🏠' WHERE "id_categoria" = 1;
UPDATE "finanzas"."categoria" SET "icono" = '💡' WHERE "id_categoria" = 2;
UPDATE "finanzas"."categoria" SET "icono" = '🍽️' WHERE "id_categoria" = 3;
UPDATE "finanzas"."categoria" SET "icono" = '🚗' WHERE "id_categoria" = 4;
UPDATE "finanzas"."categoria" SET "icono" = '⚕️' WHERE "id_categoria" = 5;

-- NUEVO: categorias del Libro mayor que no existian (id 6 a 12)
-- ("Vivienda" del frontend = "Arriendo" y "Alimentación" = "Alimentacion")
INSERT INTO "finanzas"."categoria" ("nombre", "descripcion", "icono", "activo") VALUES
('Entretenimiento', 'Salidas, streaming y ocio',          '🎮', true),
('Educación',       'Matriculas, cursos y libros',        '🎓', true),
('Gastos hormiga',  'Pequenos gastos diarios',            '☕', true),
('Salario',         'Ingreso por salario',                '💼', true),
('Otros ingresos',  'Freelance, ventas y otros ingresos', '💻', true),
('Ahorro',          'Dinero apartado para ahorro',        '🐷', true),
('Otros',           'Movimientos sin categoria',          '📦', true);


-- finanzas.movimiento_ingreso_egreso
INSERT INTO "finanzas"."movimiento_ingreso_egreso"
    ("id_usuario", "id_categoria", "nombre", "monto", "es_ingreso", "fecha", "metodo_pago", "observaciones", "activo")
VALUES
(1, NULL, 'Salario Enero',      5000000, true,  CURRENT_DATE - 20, 'Transferencia', NULL,                  true),
(1, 1,    'Gasto Arriendo',     1500000, false, CURRENT_DATE - 18, 'Transferencia', 'Pago mes en curso',   true),
(1, NULL, 'Freelance Enero',     800000, true,  CURRENT_DATE - 10, 'Nequi',         NULL,                  true),
(1, 2,    'Gasto Servicios',     300000, false, CURRENT_DATE - 7,  'Debito',        'Luz y agua',          true),
(1, 3,    'Gasto Alimentacion',  600000, false, CURRENT_DATE - 3,  'Tarjeta',       NULL,                  true),
(2, NULL, 'Salario',            3500000, true,  CURRENT_DATE - 15, 'Transferencia', NULL,                  true),
(2, 4,    'Transporte mes',      250000, false, CURRENT_DATE - 5,  'Efectivo',      NULL,                  true);

-- NUEVO: un gasto hormiga para probar ese indicador del Libro mayor (categoria 8)
INSERT INTO "finanzas"."movimiento_ingreso_egreso"
    ("id_usuario", "id_categoria", "nombre", "monto", "es_ingreso", "fecha", "metodo_pago", "observaciones", "activo")
VALUES
(1, 8,    'Café diario',          60000, false, CURRENT_DATE - 2,  'Efectivo',      NULL,                  true);


-- finanzas.tipo_ingreso
INSERT INTO "finanzas"."tipo_ingreso"
    ("id_movimiento_dinero", "nombre_movimiento_pago", "descripcion", "activo")
VALUES
(1,    'Salario',      'Ingreso por salario mensual',              true),
(3,    'Freelance',    'Ingresos por trabajo independiente',       true),
(NULL, 'Bonificacion', 'Bonificacion laboral',                     true);

-- finanzas.finanzas
INSERT INTO "finanzas"."finanzas"
    ("id_usuario", "id_movimiento_dinero", "id_categoria", "monto_presupuesto", "gasto", "disponible", "fecha", "activo")
VALUES
(1, 1,    1, 5000000, 3200000, 1800000, CURRENT_DATE, true),
(2, NULL, 2, 3500000, 2100000, 1400000, CURRENT_DATE, true),
(3, 3,    3, 4200000, 2800000, 1400000, CURRENT_DATE, true),
(4, 1,    4, 6000000, 3500000, 2500000, CURRENT_DATE, true);

-- finanzas.tipo_inversion (antes inversion.tipo_inversion)
INSERT INTO "finanzas"."tipo_inversion" ("nombre", "descripcion", "activo") VALUES
('Acciones',        'Inversion en acciones de bolsa',         true),
('Fondos Mutuales', 'Inversion en fondos mutuales',           true),
('Bonos',           'Inversion en bonos gubernamentales',     true),
('Criptomonedas',   'Inversion en criptomonedas',             true);

-- finanzas.nivel_riesgo (antes inversion.nivel_riesgo)
INSERT INTO "finanzas"."nivel_riesgo" ("nombre", "activo") VALUES
('Bajo',  true),
('Medio', true),
('Alto',  true);

-- finanzas.movimiento_inversion (antes inversion.movimiento_ingreso_egreso)
INSERT INTO "finanzas"."movimiento_inversion"
    ("nombre", "monto", "es_ingreso", "activo")
VALUES
('Inversion Acciones', 1000000, false, true),
('Ganancia Fondos',      50000, true,  true),
('Inversion Bonos',    2000000, false, true);

-- finanzas.tipo_ingreso_inversion (antes inversion.tipo_ingreso)
INSERT INTO "finanzas"."tipo_ingreso_inversion"
    ("id_movimiento_dinero", "nombre_movimiento_pago", "descripcion", "activo")
VALUES
(1, 'Rendimiento', 'Ganancia por rendimiento de inversiones', true),
(2, 'Dividendos',  'Dividendos de acciones',                  true);

-- finanzas.inversion (antes inversion.inversion)
INSERT INTO "finanzas"."inversion"
    ("id_usuario", "id_tipo_inversion", "id_nivel_riesgo", "id_movimiento_dinero",
     "nombre", "monto", "rentabilidad", "fecha_inicio", "fecha_fin", "activo")
VALUES
(1, 1, 2, 1,    'Acciones Tecnologicas', 1000000, 12.50, CURRENT_DATE - INTERVAL '180 days', CURRENT_DATE + INTERVAL '180 days', true),
(1, 2, 1, NULL, 'Fondo Moderado',         500000,  8.00, CURRENT_DATE - INTERVAL '365 days', NULL,                               true),
(2, 3, 1, 3,    'Bonos Estatales',       2000000,  6.75, CURRENT_DATE - INTERVAL '90 days',  CURRENT_DATE + INTERVAL '270 days', true),
(3, 1, 3, NULL, 'Acciones de Riesgo',     500000, 25.00, CURRENT_DATE - INTERVAL '30 days',  NULL,                               true);

-- NUEVO: rendimiento (en pesos = monto x rentabilidad) y duracion de las 4 inversiones de arriba
UPDATE "finanzas"."inversion" SET "rendimiento" = 125000, "duracion" = '1 año'    WHERE "id_inversion" = 1;
UPDATE "finanzas"."inversion" SET "rendimiento" =  40000, "duracion" = 'Flexible' WHERE "id_inversion" = 2;
UPDATE "finanzas"."inversion" SET "rendimiento" = 135000, "duracion" = '1 año'    WHERE "id_inversion" = 3;
UPDATE "finanzas"."inversion" SET "rendimiento" = 125000, "duracion" = '6 meses'  WHERE "id_inversion" = 4;

-- NUEVO: una inversion con perdida, como la que muestra el frontend (id 5)
INSERT INTO "finanzas"."inversion"
    ("id_usuario", "id_tipo_inversion", "id_nivel_riesgo", "id_movimiento_dinero",
     "nombre", "monto", "rentabilidad", "rendimiento", "duracion", "fecha_inicio", "fecha_fin", "activo")
VALUES
(1, 4, 3, NULL, 'Criptomonedas',         1000000, -44.00, -440000, 'Flexible', CURRENT_DATE - INTERVAL '120 days', NULL, true);

-- NUEVO: finanzas.historial_inversion
-- Ultimos meses de las inversiones 1, 2 y 5 del usuario 1 (graficas de crecimiento).
INSERT INTO "finanzas"."historial_inversion" ("id_inversion", "fecha", "valor", "activo") VALUES
(1, (CURRENT_DATE - INTERVAL '5 months')::date, 1000000, true),
(1, (CURRENT_DATE - INTERVAL '4 months')::date, 1020000, true),
(1, (CURRENT_DATE - INTERVAL '3 months')::date, 1045000, true),
(1, (CURRENT_DATE - INTERVAL '2 months')::date, 1070000, true),
(1, (CURRENT_DATE - INTERVAL '1 month')::date,  1100000, true),
(1, CURRENT_DATE,                               1125000, true),
(2, (CURRENT_DATE - INTERVAL '5 months')::date,  513000, true),
(2, (CURRENT_DATE - INTERVAL '4 months')::date,  518000, true),
(2, (CURRENT_DATE - INTERVAL '3 months')::date,  523000, true),
(2, (CURRENT_DATE - INTERVAL '2 months')::date,  528000, true),
(2, (CURRENT_DATE - INTERVAL '1 month')::date,   534000, true),
(2, CURRENT_DATE,                                540000, true),
(5, (CURRENT_DATE - INTERVAL '3 months')::date, 1000000, true),
(5, (CURRENT_DATE - INTERVAL '2 months')::date,  820000, true),
(5, (CURRENT_DATE - INTERVAL '1 month')::date,   690000, true),
(5, CURRENT_DATE,                                560000, true);


-- finanzas.editar_meta (antes metas.editar_meta)
INSERT INTO "finanzas"."editar_meta"
    ("nombre", "monto_actual", "monto_objetivo", "ahorro_mensual", "fecha_objetivo", "descripcion", "activo")
VALUES
('Fondo Emergencia', 4500000,  10000000,  500000, CURRENT_DATE + INTERVAL '365 days', 'Ahorrar 6 meses de gastos', true),
('Vacaciones Europa',5000000,   5000000, 1000000, CURRENT_DATE + INTERVAL '90 days',  'Viaje familiar',            true),
('Compra Casa',     12000000,  50000000, 2000000, CURRENT_DATE + INTERVAL '730 days', 'Cuota inicial vivienda',    true),
('Auto Nuevo',       8000000,  20000000, 1500000, CURRENT_DATE + INTERVAL '365 days', 'Compra de vehiculo',        true);

-- finanzas.movimiento_meta (antes metas.movimiento_ingreso_egreso)
INSERT INTO "finanzas"."movimiento_meta"
    ("nombre", "monto", "es_ingreso", "activo")
VALUES
('Aporte Meta Enero',   500000, true, true),
('Aporte Meta Febrero', 500000, true, true),
('Aporte Meta Marzo',   500000, true, true);

-- finanzas.tipo_ingreso_meta (antes metas.tipo_ingreso)
INSERT INTO "finanzas"."tipo_ingreso_meta"
    ("id_movimiento_dinero", "nombre_movimiento_pago", "descripcion", "activo")
VALUES
(1, 'Aporte Regular',  'Aporte regular a la meta',    true),
(2, 'Aporte Especial', 'Aporte adicional a la meta',  true);

-- finanzas.meta (antes metas.meta)
INSERT INTO "finanzas"."meta"
    ("id_usuario", "id_editar_meta", "id_movimiento_dinero",
     "nombre", "descripcion", "monto_objetivo", "monto_actual", "fecha_limite", "color", "icono", "activo")
VALUES
(1, 1, 1,    'Fondo de Emergencia', 'Ahorrar 6 meses de gastos',    10000000,  4500000, CURRENT_DATE + INTERVAL '365 days', '#FF9999', 'shield',   true),
(1, 2, NULL, 'Vacaciones',          'Viaje familiar a Europa',        5000000,  5000000, CURRENT_DATE + INTERVAL '90 days',  '#99FF99', 'plane',    true),
(2, 3, 2,    'Compra de Casa',      'Cuota inicial para vivienda',   50000000, 12000000, CURRENT_DATE + INTERVAL '730 days', '#9999FF', 'home',     true),
(3, 4, NULL, 'Auto Nuevo',          'Compra de vehiculo',            20000000,  8000000, CURRENT_DATE + INTERVAL '365 days', '#FFFF99', 'car',      true);

-- NUEVO: meta creada desde "Nueva meta" (sin editar_meta, ya se permite NULL) (id 5)
INSERT INTO "finanzas"."meta"
    ("id_usuario", "id_editar_meta", "id_movimiento_dinero",
     "nombre", "descripcion", "monto_objetivo", "monto_actual", "fecha_limite", "color", "icono", "activo")
VALUES
(1, NULL, NULL, 'Educación',           'Pago de la especializacion',     5000000,  5000000, CURRENT_DATE + INTERVAL '30 days',  '#C9E4CA', '🎓',       true);


-- finanzas.solicitud_consolidacion (AJUSTE CLIENTE)
INSERT INTO "finanzas"."solicitud_consolidacion"
    ("id_usuario", "saldo_total", "cuota_actual", "cuota_propuesta", "estado", "activo")
VALUES
(1, 18500000, 1450000,  980000, 'pendiente', true),
(2,  9200000,  820000,  610000, 'aprobada',  true);


-- ============================================================
-- ESQUEMA: negocio
-- (antes intermediacion + leads_schema + creditos + transacciones)
-- ============================================================

-- negocio.banco (antes intermediacion.banco)
INSERT INTO "negocio"."banco"
    ("nombre_banco", "ciudad", "contacto", "telefono", "email",
     "comision_porcentaje", "url_logo", "descripcion", "sitio_web", "estado", "activo")
VALUES
('Banco Colombiano',    'Bogota',      'Juan Gomez',   '6015551234', 'contacto@bancocol.com',    2.50, 'https://example.com/logo1.png', 'Banco con amplia cobertura nacional',     'www.bancocol.com',    'activo', true),
('Banco Metropolitano', 'Medellin',    'Maria Lopez',  '5745551234', 'contacto@bancometro.com',  2.75, 'https://example.com/logo2.png', 'Banco especializado en creditos',          'www.bancometro.com',  'activo', true),
('Banco del Oriente',   'Bucaramanga', 'Carlos Ruiz',  '7685551234', 'contacto@bancoriente.com', 2.25, 'https://example.com/logo3.png', 'Banco regional con buenos servicios',     'www.bancoriente.com', 'activo', true),
('Banco Pacifico',      'Cali',        'Pedro Diaz',   '3105551234', 'contacto@bancopac.com',    2.60, 'https://example.com/logo4.png', 'Banco del sector occidental',             'www.bancopac.com',    'activo', true);

-- NUEVO: aliados que muestra hoy la pagina Alianzas (id_banco 5 a 10)
-- Los 4 bancos de arriba quedan con tipo_aliado = 'banco' (valor por defecto).
INSERT INTO "negocio"."banco"
    ("nombre_banco", "tipo_aliado", "ciudad", "contacto", "telefono", "email",
     "comision_porcentaje", "url_logo", "logo_texto", "logo_fondo", "descripcion", "sitio_web", "estado", "activo")
VALUES
('Banco Andino',    'banco',    'Bogota',       'Juan Gomez',  '6015551234', 'alianzas@bancoandino.com.co', 2.50, NULL, 'BA', 'linear-gradient(135deg, #0f3d22 0%, #0a5c28 100%)', 'Banco con amplia cobertura nacional',                         'https://www.bancoandino.com.co', 'activo', true),
('Fintech Luz',     'fintech',  'Medellin',     'Maria Lopez', '6045551234', 'alianzas@fintechluz.co',      2.75, NULL, 'FL', 'linear-gradient(135deg, #1db954 0%, #0a5c28 100%)', 'Fintech de tarjetas de crédito 100% digitales',               'https://www.fintechluz.co',      'activo', true),
('MicroCrédito Ya', 'fintech',  'Cali',         'Pedro Diaz',  '6025551234', 'alianzas@microcreditoya.co',  3.00, NULL, 'MC', 'linear-gradient(135deg, #3f4a44 0%, #1b2320 100%)', 'Préstamos pequeños para construir historial crediticio',      'https://www.microcreditoya.co',  'activo', true),
('Banco Verde',     'banco',    'Bucaramanga',  'Carlos Ruiz', '6075551234', 'alianzas@bancoverde.com.co',  2.25, NULL, 'BV', 'linear-gradient(135deg, #0a5c28 0%, #17a34a 100%)', 'Banco especializado en crédito de vivienda',                  'https://www.bancoverde.com.co',  'activo', true),
('Tienda Más',      'comercio', 'Barranquilla', 'Laura Diaz',  '6055551234', 'alianzas@tiendamas.co',       1.50, NULL, 'TM', 'linear-gradient(135deg, #17a34a 0%, #0f3d22 100%)', 'Red de más de 300 comercios con compras a cuotas sin interés', 'https://www.tiendamas.co',       'activo', true),
('Ahorro Plus',     'comercio', 'Bogota',       'Andrea Mesa', '6015557788', 'alianzas@ahorroplus.co',      1.00, NULL, 'AP', 'linear-gradient(135deg, #6b7c72 0%, #1b2320 100%)', 'Cuenta de ahorro con rendimiento diario',                     'https://www.ahorroplus.co',      'activo', true);


-- negocio.producto_crediticio (antes intermediacion.producto_crediticio)
INSERT INTO "negocio"."producto_crediticio"
    ("id_banco", "nombre_producto", "descripcion",
     "monto_minimo", "monto_maximo", "tasa_minima", "tasa_maxima",
     "plazo_minimo", "plazo_maximo", "requisitos", "activo")
VALUES
(1, 'Credito Personal',    'Credito para gastos personales',    1000000,    50000000, 10.00, 20.00, 12,  84,  'Cedula, comprobante ingresos',           true),
(1, 'Credito Hipotecario', 'Credito para compra de vivienda',   100000000, 1000000000, 8.00, 12.00, 120, 360, 'Cedula, avaluo, comprobante ingresos',   true),
(2, 'Microcredito',        'Credito para pequenos negocios',    500000,     10000000, 15.00, 25.00,  6,  60,  'Cedula, plan de negocio',                true),
(3, 'Credito de Vehiculo', 'Financiamiento de vehiculos',       10000000,  150000000,  9.00, 18.00, 24,  84,  'Cedula, documento vehiculo',             true);

-- NUEVO: slug y familia (pestana de Alianzas) de los 4 productos de arriba
UPDATE "negocio"."producto_crediticio" SET "slug" = 'credito-personal',    "familia" = 'creditos' WHERE "id_producto" = 1;
UPDATE "negocio"."producto_crediticio" SET "slug" = 'credito-hipotecario', "familia" = 'creditos' WHERE "id_producto" = 2;
UPDATE "negocio"."producto_crediticio" SET "slug" = 'microcredito',        "familia" = 'creditos' WHERE "id_producto" = 3;
UPDATE "negocio"."producto_crediticio" SET "slug" = 'credito-vehiculo',    "familia" = 'creditos' WHERE "id_producto" = 4;

-- NUEVO: ofertas que muestra hoy la pagina Alianzas (id_producto 5 a 10,
-- una por cada aliado nuevo: id_banco 5 a 10)
INSERT INTO "negocio"."producto_crediticio"
    ("id_banco", "slug", "nombre_producto", "familia", "beneficio", "descripcion", "promesa", "vigencia",
     "monto_minimo", "monto_maximo", "tasa_minima", "tasa_maxima", "plazo_minimo", "plazo_maximo", "requisitos",
     "tiempo_aprobacion", "calificacion", "usuarios", "compatibilidad", "destacada", "cta_primaria", "cta_secundaria", "activo")
VALUES
(5, 'banco-andino',    'Crédito de libre inversión', 'creditos',  'cero_interes',
    'Crédito de libre inversión con hasta 2 puntos menos de tasa si completas los módulos básicos de educación financiera.',
    'Baja tu tasa estudiando', 'Vence en 7 días · 2.100 solicitudes · ID 8866',
    1000000, 80000000, 14.90, 24.00, 12, 60,
    'Ingresos desde 2 SMMLV; Puntaje crediticio sobre 700; Módulos básicos de FinanceUp completados',
    '24 horas', 4.8, 2100, 92, true, 'Solicitar crédito', 'Conocer el producto', true),
(6, 'fintech-luz',     'Tarjeta de crédito digital', 'tarjetas',  'cashback',
    'Tarjeta de crédito digital con 5% de cashback en mercado y transporte, y cuota de manejo en cero durante los primeros 6 meses.',
    'Cashback en lo que compras cada semana', 'Recomendada · 8.450 solicitudes · ID 2288',
    500000, 12000000, 18.50, 18.50, 1, 36,
    'Ser mayor de 18 años; Cédula colombiana vigente; Selfie de verificación de identidad',
    '10 minutos', 4.6, 8450, 87, true, 'Solicitar tarjeta', 'Conocer el producto', true),
(7, 'microcredito-ya', 'Microcrédito',               'creditos',  'sin_cuota',
    'Préstamos pequeños para construir historial crediticio desde cero, con cuotas quincenales o mensuales.',
    'Tu primer crédito, sin historial previo', 'Cupo abierto · 1.300 solicitudes · ID 6795',
    300000, 3000000, 22.00, 26.00, 3, 18,
    'Documento de identidad vigente; Cuenta bancaria a tu nombre',
    '1 hora', 4.3, 1300, 74, false, 'Solicitar préstamo', 'Conocer el producto', true),
(8, 'banco-verde',     'Crédito hipotecario',        'creditos',  'cero_interes',
    'Crédito hipotecario con tasa fija durante toda la vigencia y un asesor asignado hasta la escritura.',
    'Tasa fija de principio a fin', 'Vence en 21 días · 940 solicitudes · ID 4410',
    50000000, 350000000, 11.40, 11.40, 60, 240,
    'Ingresos desde 4 SMMLV; Cuota inicial del 20% del valor del inmueble; Antigüedad laboral de un año',
    '5 días hábiles', 4.7, 940, 81, false, 'Solicitar crédito', 'Conocer el producto', true),
(9, 'tienda-mas',      'Compra ahora, paga después', 'comercios', 'cero_interes',
    'Compra ahora y paga después sin intereses en más de 300 comercios aliados en Colombia.',
    'Difiere sin intereses en comercios aliados', 'Vigente todo el año · 5.700 activaciones · ID 1902',
    50000, 5000000, 0, 0, 3, 12,
    'Cuenta FinanceUp activa; Sin moras vigentes',
    'Inmediata', 4.5, 5700, 78, false, 'Activar el cupo', 'Conocer el producto', true),
(10, 'ahorro-plus',     'Cuenta de ahorro',           'ahorro',    'cashback',
    'Cuenta de ahorro con rendimiento diario y devolución del 3% sobre tus compras recurrentes.',
    'Tu saldo rinde todos los días', 'Nuevo aliado · 620 aperturas · ID 7731',
    0, 0, 0, 0, 0, 0,
    'Documento de identidad vigente; Correo verificado',
    '15 minutos', 4.4, 620, 69, false, 'Abrir la cuenta', 'Conocer el producto', true);

-- NUEVO: negocio.producto_detalle (productos 5 a 10)
INSERT INTO "negocio"."producto_detalle" ("id_producto", "tipo", "texto", "orden", "activo") VALUES
-- producto 5: Banco Andino
(5, 'etiqueta',       'Tasa preferencial',                                               1, true),
(5, 'etiqueta',       'Sin codeudor',                                                    2, true),
(5, 'etiqueta',       'Desembolso en 24 h',                                              3, true),
(5, 'requisito',      'Ingresos desde 2 SMMLV',                                          1, true),
(5, 'requisito',      'Puntaje crediticio sobre 700',                                    2, true),
(5, 'requisito',      'Módulos básicos de FinanceUp completados',                        3, true),
(5, 'caracteristica', 'Desembolso a tu cuenta de ahorros en 24 horas hábiles',           1, true),
(5, 'caracteristica', 'Puedes abonar a capital cuando quieras, sin penalidad',           2, true),
(5, 'caracteristica', 'La tasa baja 0,5 puntos por cada módulo que termines, hasta 2 puntos', 3, true),
(5, 'documento',      'Cédula de ciudadanía al día',                                     1, true),
(5, 'documento',      'Certificado laboral no mayor a 30 días',                          2, true),
(5, 'documento',      'Extractos bancarios de los últimos 3 meses',                      3, true),
(5, 'perfil',         'Score Alto',                                                      1, true),
-- producto 6: Fintech Luz
(6, 'etiqueta',       '100% digital',                                                    1, true),
(6, 'etiqueta',       'Cashback 5%',                                                     2, true),
(6, 'etiqueta',       'Sin cuota 6 meses',                                               3, true),
(6, 'requisito',      'Ser mayor de 18 años',                                            1, true),
(6, 'requisito',      'Cédula colombiana vigente',                                       2, true),
(6, 'requisito',      'Selfie de verificación de identidad',                             3, true),
(6, 'caracteristica', 'Tarjeta virtual activa apenas te aprueban, la física llega en 5 días', 1, true),
(6, 'caracteristica', 'Congela y descongela la tarjeta desde la app cuando quieras',     2, true),
(6, 'caracteristica', 'El cashback se abona el primer día hábil del mes siguiente',      3, true),
(6, 'documento',      'Cédula de ciudadanía al día',                                     1, true),
(6, 'documento',      'Selfie de verificación',                                          2, true),
(6, 'perfil',         'Score Alto',                                                      1, true),
(6, 'perfil',         'Historial Nuevo',                                                 2, true),
-- producto 7: MicroCrédito Ya
(7, 'etiqueta',       'Sin historial',                                                   1, true),
(7, 'etiqueta',       'Cuotas quincenales',                                              2, true),
(7, 'requisito',      'Documento de identidad vigente',                                  1, true),
(7, 'requisito',      'Cuenta bancaria a tu nombre',                                     2, true),
(7, 'caracteristica', 'Empiezas con $300.000 y el cupo sube cada vez que pagas a tiempo', 1, true),
(7, 'caracteristica', 'Eliges cuotas quincenales o mensuales según cómo te paguen',      2, true),
(7, 'caracteristica', 'Reportamos tu buen comportamiento a centrales de riesgo',         3, true),
(7, 'documento',      'Cédula de ciudadanía al día',                                     1, true),
(7, 'documento',      'Certificación de cuenta bancaria',                                2, true),
(7, 'perfil',         'Historial Nuevo',                                                 1, true),
-- producto 8: Banco Verde
(8, 'etiqueta',       'Tasa fija',                                                       1, true),
(8, 'etiqueta',       'Asesor asignado',                                                 2, true),
(8, 'etiqueta',       'Compatible con subsidio',                                         3, true),
(8, 'requisito',      'Ingresos desde 4 SMMLV',                                          1, true),
(8, 'requisito',      'Cuota inicial del 20% del valor del inmueble',                    2, true),
(8, 'requisito',      'Antigüedad laboral de un año',                                    3, true),
(8, 'caracteristica', 'Financiamos hasta el 80% del valor comercial del inmueble',       1, true),
(8, 'caracteristica', 'La tasa no cambia aunque suba la del mercado',                    2, true),
(8, 'caracteristica', 'Un asesor te acompaña desde el avalúo hasta la firma de escritura', 3, true),
(8, 'documento',      'Cédula de ciudadanía al día',                                     1, true),
(8, 'documento',      'Certificado laboral no mayor a 30 días',                          2, true),
(8, 'documento',      'Extractos bancarios de los últimos 3 meses',                      3, true),
(8, 'documento',      'Promesa de compraventa firmada',                                  4, true),
(8, 'documento',      'Declaración de renta del último año',                             5, true),
(8, 'perfil',         'Score Alto',                                                      1, true),
-- producto 9: Tienda Más
(9, 'etiqueta',       '0% interés',                                                      1, true),
(9, 'etiqueta',       '300 comercios',                                                   2, true),
(9, 'etiqueta',       'Aprobación inmediata',                                            3, true),
(9, 'requisito',      'Cuenta FinanceUp activa',                                         1, true),
(9, 'requisito',      'Sin moras vigentes',                                              2, true),
(9, 'caracteristica', 'Difiere a 3, 6 o 12 cuotas sin interés en comercios de la red',   1, true),
(9, 'caracteristica', 'El cupo se libera de nuevo a medida que pagas',                   2, true),
(9, 'caracteristica', 'Pagas desde la app, sin ir al comercio',                          3, true),
(9, 'documento',      'Cédula de ciudadanía al día',                                     1, true),
(9, 'perfil',         'Score Alto',                                                      1, true),
(9, 'perfil',         'Historial Nuevo',                                                 2, true),
-- producto 10: Ahorro Plus
(10, 'etiqueta',       'Rendimiento diario',                                              1, true),
(10, 'etiqueta',       'Sin cuota de manejo',                                             2, true),
(10, 'etiqueta',       'Cashback 3%',                                                     3, true),
(10, 'requisito',      'Documento de identidad vigente',                                  1, true),
(10, 'requisito',      'Correo verificado',                                               2, true),
(10, 'caracteristica', 'Los rendimientos se abonan cada día sobre el saldo disponible',   1, true),
(10, 'caracteristica', 'Sin monto mínimo de apertura ni saldo mínimo',                    2, true),
(10, 'caracteristica', 'Retiras sin costo en la red de cajeros aliados',                  3, true),
(10, 'documento',      'Cédula de ciudadanía al día',                                     1, true),
(10, 'perfil',         'Historial Nuevo',                                                 1, true);

-- NUEVO: negocio.producto_tarifa (productos 5 a 10)
INSERT INTO "negocio"."producto_tarifa" ("id_producto", "concepto", "valor", "orden", "activo") VALUES
(5, 'Tasa de interés',                 'Desde 14,9% E.A.',                        1, true),
(5, 'Estudio de crédito',              'Sin costo',                               2, true),
(5, 'Seguro de vida deudor',           '0,08% mensual sobre el saldo',            3, true),
(5, 'Abono extraordinario a capital',  'Sin penalidad',                           4, true),
(6, 'Tasa de interés',                 '18,5% E.A.',                              1, true),
(6, 'Cuota de manejo',                 '$0 los primeros 6 meses, luego $19.900',  2, true),
(6, 'Avance en cajero',                '4,5% del monto, mínimo $9.500',           3, true),
(6, 'Reposición por pérdida',          '$28.000',                                 4, true),
(7, 'Tasa de interés',                 '22% E.A.',                                1, true),
(7, 'Administración por desembolso',   '$12.000',                                 2, true),
(7, 'Pago tardío',                     '$8.500 por cuota vencida',                3, true),
(8, 'Tasa de interés',                 'Desde 11,4% E.A. fija',                   1, true),
(8, 'Avalúo del inmueble',             'Desde $380.000',                          2, true),
(8, 'Estudio de títulos',              '$420.000',                                3, true),
(8, 'Seguro de incendio y terremoto',  'Según valor asegurado',                   4, true),
(9, 'Interés corriente',               '0% en comercios aliados',                 1, true),
(9, 'Cuota de manejo',                 'Sin costo',                               2, true),
(9, 'Mora',                            'Tasa máxima legal vigente',               3, true),
(10, 'Rendimiento',                     '9,2% E.A. sobre el saldo diario',         1, true),
(10, 'Cuota de manejo',                 'Sin costo',                               2, true),
(10, 'Retiro en cajeros de otra red',   '$7.600',                                  3, true),
(10, 'Cuatro por mil',                  'Exenta hasta 350 UVT al mes',             4, true);

-- NUEVO: negocio.producto_pregunta (productos 5 a 10)
INSERT INTO "negocio"."producto_pregunta" ("id_producto", "pregunta", "respuesta", "orden", "activo") VALUES
(5, '¿Cuándo se aplica el descuento en la tasa?',
    'Al cierre de cada mes revisamos tus módulos completados y ajustamos la tasa de la siguiente cuota. El descuento máximo acumulado es de 2 puntos.', 1, true),
(5, '¿Puedo pedirlo si ya tengo otro crédito?',
    'Sí, siempre que la suma de tus cuotas no supere el 40% de tu ingreso mensual comprobable.', 2, true),
(6, '¿Sobre qué compras aplica el 5%?',
    'Supermercados, transporte y estaciones de servicio, con tope de $120.000 de cashback al mes.', 1, true),
(6, '¿Qué pasa después de los 6 meses sin cuota de manejo?',
    'La cuota queda en $19.900 mensuales y se exonera cualquier mes en que factures más de $800.000.', 2, true),
(7, '¿Cada cuánto sube mi cupo?',
    'Después de tres cuotas pagadas a tiempo revisamos tu cupo y puede subir hasta un 40%.', 1, true),
(8, '¿Puedo usar el subsidio de vivienda?',
    'Sí. El subsidio se aplica como parte de la cuota inicial y no reduce el monto máximo que podemos financiarte.', 1, true),
(8, '¿Cuánto tarda el desembolso?',
    'La aprobación toma 5 días hábiles y el desembolso ocurre el día de la firma de escritura.', 2, true),
(9, '¿Dónde puedo usar el cupo?',
    'En los comercios marcados con el sello Tienda Más dentro de la app. La lista se actualiza cada mes.', 1, true),
(10, '¿El rendimiento es fijo?',
    'Es variable y sigue la tasa del mercado. Te avisamos por la app cada vez que cambia.', 1, true);


-- negocio.asesor_bancario (antes intermediacion.asesor_bancario)
INSERT INTO "negocio"."asesor_bancario"
    ("id_banco", "nombre", "apellido", "email", "telefono", "especialidad", "activo")
VALUES
(1, 'Roberto',  'Sanchez',    'roberto.sanchez@bancocol.com',      '3001112222', 'Creditos Personales', true),
(1, 'Diana',    'Valenzuela', 'diana.valenzuela@bancocol.com',      '3009998888', 'Hipotecarios',        true),
(2, 'Fernando', 'Castillo',   'fernando.castillo@bancometro.com',   '3107776666', 'Microcreditos',       true),
(3, 'Claudia',  'Morales',    'claudia.morales@bancoriente.com',    '3104445555', 'Vehiculos',           true);

-- negocio.contacto_asesor (antes intermediacion.contacto_asesor)
INSERT INTO "negocio"."contacto_asesor"
    ("id_asesor", "whatsapp", "email", "telefono",
     "disponible_desde", "disponible_hasta", "dias_disponibles", "activo")
VALUES
(1, '3001112222', 'roberto.sanchez@bancocol.com',    '3001112222', '08:00', '18:00', 'Lunes a Viernes', true),
(2, '3009998888', 'diana.valenzuela@bancocol.com',   '3009998888', '09:00', '17:00', 'Lunes a Viernes', true),
(3, '3107776666', 'fernando.castillo@bancometro.com','3107776666', '08:00', '20:00', 'Lunes a Sabado',  true),
(4, '3104445555', 'claudia.morales@bancoriente.com', '3104445555', '07:00', '19:00', 'Lunes a Viernes', true);

-- negocio.lead (antes leads_schema.lead)
INSERT INTO "negocio"."lead"
    ("id_usuario", "id_producto", "id_asesor", "tipo_credito",
     "monto_interes", "plazo_interes", "estado_lead",
     "fecha_generacion", "fecha_contacto", "observaciones", "activo")
VALUES
(1, 1, 1, 'Personal',    5000000,   36, 'aprobado',   CURRENT_TIMESTAMP - INTERVAL '60 days', CURRENT_TIMESTAMP - INTERVAL '50 days', 'Cliente solvente, aprobado sin inconvenientes',          true),
(2, 2, 2, 'Hipotecario', 200000000, 240,'en_proceso',  CURRENT_TIMESTAMP - INTERVAL '30 days', CURRENT_TIMESTAMP - INTERVAL '25 days', 'Pendiente evaluacion de inmueble',                       true),
(3, 3, 3, 'Microcredito',5000000,   36, 'aprobado',   CURRENT_TIMESTAMP - INTERVAL '45 days', CURRENT_TIMESTAMP - INTERVAL '40 days', 'Negocio informal, requiere seguimiento',                 true),
(4, 4, 4, 'Vehiculo',    50000000,  60, 'contactado', CURRENT_TIMESTAMP - INTERVAL '15 days', CURRENT_TIMESTAMP - INTERVAL '10 days', 'Interesado en financiamiento de camioneta',              true),
(5, 1, 1, 'Personal',    3000000,   24, 'nuevo',      CURRENT_TIMESTAMP - INTERVAL '5 days',  NULL,                                   'Lead recien generado',                                   true);

-- NUEVO: columnas nuevas (formulario de solicitud) de los 5 leads de arriba
UPDATE "negocio"."lead" SET "radicado" = 'SOL-104233', "nombre_solicitante" = 'Harold Arciniegas', "documento_solicitante" = '1234567890', "correo_solicitante" = 'harold.arciniegas@email.com', "celular_solicitante" = '3001234567', "ingresos_mensuales" =  5800000, "acepta_terminos" = true WHERE "id_lead" = 1;
UPDATE "negocio"."lead" SET "radicado" = 'SOL-218764', "nombre_solicitante" = 'Fabio Zorro',       "documento_solicitante" = '9876543210', "correo_solicitante" = 'fabio.zorro@email.com',       "celular_solicitante" = '3109876543', "ingresos_mensuales" = 12000000, "acepta_terminos" = true WHERE "id_lead" = 2;
UPDATE "negocio"."lead" SET "radicado" = 'SOL-330915', "nombre_solicitante" = 'Johan Barreto',     "documento_solicitante" = '5555555555', "correo_solicitante" = 'johan.barreto@email.com',     "celular_solicitante" = '3201234567', "ingresos_mensuales" =  2400000, "acepta_terminos" = true WHERE "id_lead" = 3;
UPDATE "negocio"."lead" SET "radicado" = 'SOL-447120', "nombre_solicitante" = 'Sharith Bermudez',  "documento_solicitante" = '4444444444', "correo_solicitante" = 'sharith.bermudez@email.com',  "celular_solicitante" = '3051234567', "ingresos_mensuales" =  4500000, "acepta_terminos" = true WHERE "id_lead" = 4;
UPDATE "negocio"."lead" SET "radicado" = 'SOL-559382', "nombre_solicitante" = 'Fabian Barreto',    "documento_solicitante" = '6666666666', "correo_solicitante" = 'fabian.barreto@email.com',    "celular_solicitante" = '3161234567', "ingresos_mensuales" =  3800000, "acepta_terminos" = true WHERE "id_lead" = 5;


-- negocio.conversacion_usuario_asesor (antes leads_schema.conversacion_usuario_asesor)
INSERT INTO "negocio"."conversacion_usuario_asesor"
    ("id_lead", "id_usuario", "id_asesor", "tipo_contacto",
     "asunto", "contenido", "fecha_mensaje", "activo")
VALUES
(1, 1, 1, 'email',    'Solicitud de Credito Aprobada',  'Le informamos que su solicitud fue aprobada.',                   CURRENT_TIMESTAMP - INTERVAL '50 days', true),
(1, 1, 1, 'whatsapp', 'Confirmacion de Desembolso',     'El dinero sera depositado en 2 dias habiles.',                   CURRENT_TIMESTAMP - INTERVAL '45 days', true),
(2, 2, 2, 'telefono', 'Solicitud de Informacion',       'Llamada para recabar informacion sobre su propiedad.',           CURRENT_TIMESTAMP - INTERVAL '25 days', true),
(3, 3, 3, 'email',    'Plan de Acompanamiento',         'Iniciamos plan de acompanamiento para su negocio.',              CURRENT_TIMESTAMP - INTERVAL '40 days', true),
(4, 4, 4, 'whatsapp', 'Consulta sobre Vehiculo',        'Informacion sobre financiamiento de su vehiculo.',               CURRENT_TIMESTAMP - INTERVAL '10 days', true);

-- negocio.credito_desembolsado (antes creditos.credito_desembolsado)
INSERT INTO "negocio"."credito_desembolsado"
    ("id_lead", "id_usuario", "id_producto", "id_banco",
     "numero_credito", "monto_aprobado", "tasa_interes_final", "plazo_meses",
     "fecha_aprobacion", "fecha_desembolso", "estado_credito", "saldo_actual", "activo")
VALUES
(1, 1, 1, 1, 'CRED-2024-001',   5000000, 15.50,  36, CURRENT_DATE - INTERVAL '50 days', CURRENT_DATE - INTERVAL '48 days', 'activo',   4800000, true),
(2, 2, 2, 1, 'CRED-2024-002', 200000000, 10.00, 240, CURRENT_DATE - INTERVAL '20 days', CURRENT_DATE - INTERVAL '18 days', 'activo', 199800000, true),
(3, 3, 3, 2, 'CRED-2024-003',   5000000, 18.75,  36, CURRENT_DATE - INTERVAL '40 days', CURRENT_DATE - INTERVAL '38 days', 'activo',   4700000, true),
(4, 4, 4, 3, 'CRED-2024-004',  50000000, 12.50,  60, CURRENT_DATE - INTERVAL '5 days',  NULL,                              'activo',  50000000, true);

-- negocio.transaccion_comision (antes transacciones.transaccion_comision)
INSERT INTO "negocio"."transaccion_comision"
    ("id_credito", "id_banco", "monto_comision", "porcentaje_aplicado",
     "fecha_transaccion", "estado", "referencia_pago", "activo")
VALUES
(1, 1,  125000, 2.50, CURRENT_TIMESTAMP - INTERVAL '48 days', 'pagada',   'TRANS-001-2024', true),
(2, 1, 5000000, 2.50, CURRENT_TIMESTAMP - INTERVAL '18 days', 'pendiente','TRANS-002-2024', true),
(3, 2,  125000, 2.50, CURRENT_TIMESTAMP - INTERVAL '38 days', 'pagada',   'TRANS-003-2024', true),
(4, 3, 1125000, 2.25, CURRENT_TIMESTAMP - INTERVAL '5 days',  'pendiente','TRANS-004-2024', true);
