package controllers

// ============================================================================
// GUIA DEL FLUJO DE AUTENTICACION (rama barreto)
// Todas las rutas van al MID: http://localhost:8080
// Body en JSON (en Postman: Body -> raw -> JSON).
// "sin token" = ruta publica. "con token" = Authorization: Bearer <token>
// ============================================================================
//
// ----------------------------------------------------------------------------
// 0. ANTES DE PROBAR
// ----------------------------------------------------------------------------
//
//   1. Encender crud_auth (8081). Para la parte E tambien crud_soporte (8085),
//      ahi queda el historial de los cambios que hace el admin.
//   2. Encender el MID desde la carpeta api_mid:  go run .
//      No debe salir el aviso de JWT_SECRETO (llenarlo en MID_FINANCEUP/.env).
//   3. En Postman crear un Environment con:  url = http://localhost:8080/v1
//      y token (vacio). En la coleccion: Authorization -> Bearer Token -> {{token}}
//   4. En el paso 1 y el paso 3, en Scripts -> Post-response, pegar:
//        const r = pm.response.json();
//        if (r.token) pm.environment.set("token", r.token);
//   5. Los usuarios de prueba de la base tienen correos falsos (@email.com):
//      el codigo nunca les llega. Usar un correo real (registrarse en el paso 1
//      o cambiar el correo de uno en pgAdmin). Su clave es Financeup2026*
//
// ----------------------------------------------------------------------------
// A. REGISTRO (solo la primera vez)
// ----------------------------------------------------------------------------
//
// PASO 1. Registrarse (deja la sesion iniciada: entrega el token de una vez)
//   POST /v1/registro   (sin token)
//   Body (obligatorios: nombre, apellido, email, contrasena, acepta_terminos):
//     {
//       "nombre": "Laura",
//       "apellido": "Gomez",
//       "email": "laura.gomez@email.com",
//       "contrasena": "Financeup2026",
//       "acepta_terminos": true,
//       "id_tipo_documento": 1,
//       "cedula": "1098765432",
//       "telefono": "3157778899",
//       "ciudad": "Bogota",
//       "direccion": "Calle 45 # 12-30"
//     }
//   Reglas: nombre y apellido minimo 3 letras; contrasena de 8 a 72
//   caracteres con una mayuscula y un numero; la cedula no se puede repetir.
//   id_tipo_documento: 1 = CC, 2 = CE, 3 = Pasaporte, 4 = NIT.
//   Al registrarse queda con el rol "Usuario".
//   Respuesta 201 (igual que el paso 3 del login, con otro mensaje):
//     {
//       "mensaje": "Usuario registrado",
//       "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
//       "expira_en": 28800,
//       "id_usuario": 7,
//       "nombre": "Laura",
//       "apellido": "Gomez",
//       "email": "laura.gomez@email.com",
//       "rol": "user",
//       "estado": "activo"
//     }
//
// ----------------------------------------------------------------------------
// B. INICIO DE SESION (cada vez que se entra)
// ----------------------------------------------------------------------------
//
// PASO 2. Correo y contrasena (envia un codigo que vence en 10 minutos)
//   POST /v1/login   (sin token)
//   Body:
//     { "email": "laura.gomez@email.com", "contrasena": "Financeup2026" }
//   Respuesta 200 (guardar el id_usuario para el paso 3):
//     { "id_usuario": 7, "mensaje": "Te enviamos un codigo de verificacion a tu correo",
//       "requiere_codigo": true }
//
// PASO 3. Codigo del login (entrega el token de la sesion, dura 8 horas)
//   POST /v1/login/verificar   (sin token)
//   Body:
//     { "id_usuario": 7, "codigo": "654321" }
//   Respuesta 200:
//     {
//       "mensaje": "Inicio de sesion exitoso",
//       "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
//       "expira_en": 28800,
//       "id_usuario": 7,
//       "nombre": "Laura",
//       "apellido": "Gomez",
//       "email": "laura.gomez@email.com",
//       "rol": "user",
//       "estado": "activo"
//     }
//   "rol" es "admin" si el usuario tiene el rol Administrador; si no, "user".
//
// PASO 3b. Reenviar el codigo (solo si no llego; esperar 1 minuto entre codigos)
//   POST /v1/login/reenviar   (sin token)
//   Body:
//     { "id_usuario": 7 }
//   Respuesta 200:
//     { "mensaje": "Te enviamos un codigo nuevo a tu correo", "id_usuario": 7 }
//
// PASO 3c. Entrar con Google (no se prueba en Postman: el id_token lo da el
// boton de Google en la pagina). Necesita GOOGLE_CLIENT_ID en el .env.
//   POST /v1/login/social   (sin token)
//   Body:
//     { "proveedor": "google", "id_token": "<el que entrega Google>" }
//   Responde igual que el paso 3. "apple" y "facebook" responden 501.
//
// ----------------------------------------------------------------------------
// C. CON LA SESION INICIADA (todas con token)
// ----------------------------------------------------------------------------
//
// PASO 4. Ver el perfil (el id de la URL debe ser el propio; el admin ve todos)
//   GET /v1/perfil/7   (con token, sin body)
//   Respuesta 200:
//     { "id_usuario": 7, "nombre": "Laura", "apellido": "Gomez",
//       "email": "laura.gomez@email.com", "tipo_documento": "CC",
//       "cedula": "1098765432", "telefono": "3157778899", "ciudad": "Bogota",
//       "direccion": "Calle 45 # 12-30", "fecha_nacimiento": "",
//       "estado": "activo", "rol": "user" }
//
// PASO 5. Cambiar la contrasena (opcional)
//   PUT /v1/perfil/7/contrasena   (con token)
//   Body:
//     { "contrasena_actual": "Financeup2026", "contrasena_nueva": "Financeup2027" }
//   Respuesta 200:
//     { "mensaje": "Contrasena actualizada" }
//
// PASO 6. Cerrar sesion (despues el token ya no sirve: volver al paso 2)
//   POST /v1/logout   (con token, sin body)
//   Respuesta 200:
//     { "mensaje": "Sesion cerrada" }
//
// ----------------------------------------------------------------------------
// D. RECUPERAR LA CONTRASENA (sin token)
// ----------------------------------------------------------------------------
//
// PASO 7. Pedir el codigo (vence en 15 minutos; siempre responde lo mismo
// para no revelar que correos estan registrados)
//   POST /v1/recuperar-contrasena   (sin token)
//   Body:
//     { "email": "laura.gomez@email.com" }
//   Respuesta 200:
//     { "mensaje": "Si el correo esta registrado, te enviamos un codigo" }
//
// PASO 8. Poner la contrasena nueva (tambien quita el bloqueo de 15 minutos)
//   PUT /v1/recuperar-contrasena   (sin token)
//   Body:
//     { "email": "laura.gomez@email.com", "codigo": "112233",
//       "contrasena_nueva": "Financeup2028" }
//   Respuesta 200:
//     { "mensaje": "Contrasena actualizada" }
//   Despues se vuelve al paso 2 con la contrasena nueva.
//
// ----------------------------------------------------------------------------
// E. ADMINISTRAR USUARIOS (con token de un usuario con rol Administrador)
// ----------------------------------------------------------------------------
// Si el token no es de admin responde 403. El admin no puede cambiarse
// su propio rol ni su propio estado. Cada cambio queda en el historial
// (soporte.registro_actividad) y al usuario se le cierran las sesiones:
// debe volver a entrar (paso 2) para que el cambio se le note.
//
// PASO 9. Ver todos los usuarios con su rol
//   GET /v1/admin/usuarios   (con token de admin, sin body)
//   Respuesta 200 (lista, del mas nuevo al mas viejo):
//     [ { "id_usuario": 7, "nombre": "Laura", "apellido": "Gomez",
//         "email": "laura.gomez@email.com", "telefono": "3157778899",
//         "cedula": "1098765432", "estado": "activo", "activo": true,
//         "rol": "user", "roles": ["Usuario"],
//         "fecha_registro": "2026-10-22T15:04:05Z", "ultima_sesion": null } ]
//   "roles" trae el nombre real (Administrador, Asesor, Usuario, Gerente).
//   "ultima_sesion" es null si nunca ha entrado.
//
// PASO 10. Cambiar el rol (queda con un solo rol)
//   PUT /v1/admin/usuarios/7/rol   (con token de admin)
//   Body ("admin", "user" o el nombre real del rol):
//     { "rol": "admin" }
//   Respuesta 200:
//     { "mensaje": "Rol actualizado. El usuario debe iniciar sesion de nuevo.",
//       "id_usuario": 7, "rol": "admin", "roles": ["Administrador"] }
//
// PASO 11. Activar, suspender o inactivar una cuenta
//   PUT /v1/admin/usuarios/7/estado   (con token de admin)
//   Body (estado: activo, suspendido o inactivo; el motivo es obligatorio
//   para suspender o inactivar, minimo 5 caracteres):
//     { "estado": "suspendido", "motivo": "Actividad sospechosa" }
//   Respuesta 200:
//     { "mensaje": "Estado actualizado", "id_usuario": 7, "estado": "suspendido" }
//   "inactivo" se usa en vez de borrar la cuenta. Para reactivarla:
//     { "estado": "activo" }
//
//   El resto de la pagina de admin no necesita rutas nuevas:
//     editar datos      -> directo al CRUD: GET /v1/usuario/7, cambiar campos
//                          y PUT /v1/usuario/7 con el objeto COMPLETO
//     crear usuario     -> paso 1 y luego paso 10
//     restablecer clave -> paso 7
//     ver su historial  -> CRUD soporte: GET /v1/registro_actividad?query=id_usuario:7
//
// ----------------------------------------------------------------------------
// ERRORES COMUNES (todos responden { "error": "mensaje" })
// ----------------------------------------------------------------------------
//   400  Datos invalidos (contrasena debil, correo mal escrito, falta un campo)
//   400  "No se pudo guardar..." (la cedula o el correo ya existen en la base)
//   400  "No puedes cambiar tu propio rol" / "...el estado de tu propia cuenta"
//   400  "Ese rol no existe" / "El estado debe ser activo, inactivo o suspendido"
//   401  "Correo o contrasena incorrectos"
//   401  "El codigo no es correcto" / "El codigo vencio. Pide uno nuevo."
//   401  "La sesion no es valida" (sin token, vencido, despues del logout
//        o despues de que un admin le cambio el rol o el estado)
//   403  "No puedes ver ni cambiar datos de otro usuario"
//   403  "Solo un administrador puede ver esta informacion"
//   403  "Tu cuenta esta suspendida..." / "...inactiva..." (al iniciar sesion)
//   404  "El usuario no existe"
//   409  "Ya existe un usuario con ese correo" / "El usuario ya esta suspendido"
//   423  "Cuenta bloqueada..." (5 intentos fallidos: esperar 15 minutos)
//   429  "Ya te enviamos un codigo hace menos de un minuto"
//   501  Login con Apple o Facebook (todavia no existe)
//   502  "No fue posible comunicarse con un CRUD" (el CRUD esta apagado)
//   503  Falta GOOGLE_CLIENT_ID en el .env (login con Google), o Gmail no
//        esta configurado y el MID no esta en modo dev
//
// ----------------------------------------------------------------------------
// TIEMPOS (se cambian en services/auth.go)
// ----------------------------------------------------------------------------
//   Codigo del login ................... 10 minutos (minutosCodigoLogin)
//   Codigo para recuperar contrasena ... 15 minutos (minutosCodigoRecuperar)
//   Espera entre codigos ............... 60 segundos (segundosEntreCodigos)
//   Bloqueo por intentos fallidos ...... 15 minutos tras 5 intentos
//   Duracion de la sesion (token) ...... 8 horas (helpers/token.go)
//
// Correos: salen por Gmail si existen api_mid/credentials/credentials.json y
// api_mid/token/token.json. Sin ellos, en modo dev el codigo sale en la consola
// del MID. Si Gmail deja de enviar (pasa cada 7 dias mientras la app de Google
// esta "en prueba"): borrar token/token.json, arrancar el MID y autorizar otra vez.
// ============================================================================
