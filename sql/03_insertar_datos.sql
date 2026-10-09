-- ============================================================
-- FINANCEUP - DATOS DE PRUEBA
-- Ejecutar conectado a "financeup" DESPUES de 02_crear_tablas.sql.
-- Todos los usuarios de prueba tienen la clave: Financeup2026*
-- ============================================================

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

-- soporte.pqr
INSERT INTO "soporte"."pqr"
    ("id_usuario", "radicado", "titulo", "tipo", "categoria", "prioridad", "asesor", "mensaje_respuesta", "descripcion", "id_estado", "activo")
VALUES
(1, 'PQR-2026-0001', 'No puedo ingresar',        'peticion', 'Plataforma',            'alta',  'Soporte tecnico', NULL,                                          'Problema con acceso a la plataforma',     1, true),
(2, 'PQR-2026-0002', 'Duda sobre tasas',         'peticion', 'Creditos',              'media', 'Camila Rojas',    NULL,                                          'Consulta sobre tasas de interes',         2, true),
(3, 'PQR-2026-0003', 'Cobro de comision',        'reclamo',  'Movimientos y pagos',   'alta',  NULL,              NULL,                                          'Reclamo por comision no autorizada',      1, true),
(4, 'PQR-2026-0004', 'Informacion de creditos',  'peticion', 'Atencion al cliente',   'baja',  'Andres Pena',     'Le enviamos al correo el portafolio vigente.', 'Solicitud de informacion sobre creditos', 3, true);

-- soporte.adjunto
INSERT INTO "soporte"."adjunto"
    ("id_pqr", "nombre_archivo", "ruta_archivo", "tipo_mime", "tamano_bytes", "activo")
VALUES
(1, 'error_screenshot.png',       '/uploads/pqr/error_screenshot.png',       'image/png',       512000, true),
(2, 'consulta_documento.pdf',     '/uploads/pqr/consulta_documento.pdf',     'application/pdf', 256000, true),
(3, 'comprobante_comision.pdf',   '/uploads/pqr/comprobante_comision.pdf',   'application/pdf', 384000, true);

-- soporte.registro_actividad (antes auditoria.registro_actividad)
INSERT INTO "soporte"."registro_actividad"
    ("id_usuario", "tipo_actividad", "descripcion", "entidad_afectada", "fecha_actividad")
VALUES
(1, 'LOGIN',            'Usuario inicio sesion',                   'auth.usuario',       CURRENT_TIMESTAMP - INTERVAL '2 hours'),
(1, 'CREAR_SOLICITUD',  'Se creo una nueva solicitud de credito',  'negocio.lead',       CURRENT_TIMESTAMP - INTERVAL '60 days'),
(2, 'ACTUALIZAR_PERFIL','Se actualizo el perfil de usuario',       'auth.usuario',       CURRENT_TIMESTAMP - INTERVAL '10 days'),
(3, 'CREAR_SIMULACION', 'Se ejecuto una simulacion financiera',    'finanzas.finanzas',  CURRENT_TIMESTAMP - INTERVAL '10 days'),
(4, 'LOGIN',            'Usuario inicio sesion',                   'auth.usuario',       CURRENT_TIMESTAMP - INTERVAL '1 hour');


-- ============================================================
-- ESQUEMA: educacion
-- ============================================================

-- educacion.modulo_educativo
INSERT INTO "educacion"."modulo_educativo"
    ("titulo", "descripcion", "contenido", "nivel", "url_thumbnail", "activo")
VALUES
('Introduccion a Finanzas Personales', 'Conceptos basicos de finanzas personales',  'Contenido teorico basico',              'basico',      'https://example.com/thumb1.jpg', true),
('Creditos y Deudas',                  'Como gestionar creditos y deudas',           'Contenido intermedio sobre gestion',    'intermedio',  'https://example.com/thumb2.jpg', true),
('Inversiones Avanzadas',              'Estrategias avanzadas de inversion',         'Contenido avanzado de portafolio',      'avanzado',    'https://example.com/thumb3.jpg', true),
('Ahorro e Inversion',                 'Tecnicas de ahorro e inversion',             'Contenido sobre ahorro',                'basico',      'https://example.com/thumb4.jpg', true);

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