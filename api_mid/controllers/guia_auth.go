package controllers

// ============================================================================
// GUIA DEL FLUJO DE AUTENTICACION (rama barreto)
// Todas las rutas van al MID: http://localhost:8080
// Body en JSON (en Postman: Body -> raw -> JSON).
// "sin token" = ruta publica. "con token" = Authorization: Bearer <token>
// ============================================================================
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
// ERRORES COMUNES (todos responden { "error": "mensaje" })
// ----------------------------------------------------------------------------
//   400  Datos invalidos (contrasena debil, correo mal escrito, falta un campo)
//   401  "Correo o contrasena incorrectos"
//   401  "El codigo no es correcto" / "El codigo vencio. Pide uno nuevo."
//   401  "La sesion no es valida" (sin token, vencido o despues del logout)
//   403  "No puedes ver ni cambiar datos de otro usuario"
//   409  "Ya existe un usuario con ese correo"
//   423  "Cuenta bloqueada..." (5 intentos fallidos: esperar 15 minutos)
//   429  "Ya te enviamos un codigo hace menos de un minuto"
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
// Correos: salen por Gmail si existe api_mid/credentials/credentials.json.
// Sin ese archivo, en modo dev el codigo se muestra en la consola del MID.
// ============================================================================
