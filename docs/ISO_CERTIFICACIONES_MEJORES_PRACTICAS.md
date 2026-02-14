# Mejores Practicas de Certificaciones ISO - FasMail Panel

## Indice

1. [Introduccion](#1-introduccion)
2. [Resumen del Proyecto FasMail Panel](#2-resumen-del-proyecto-fasmail-panel)
3. [ISO 9001:2015 - Sistema de Gestion de Calidad](#3-iso-90012015---sistema-de-gestion-de-calidad-sgc)
4. [ISO/IEC 27001:2022 - Seguridad de la Informacion](#4-isoiec-270012022---sistema-de-gestion-de-seguridad-de-la-informacion-sgsi)
5. [ISO/IEC 27701:2019 - Gestion de Privacidad](#5-isoiec-277012019---gestion-de-privacidad-de-la-informacion)
6. [ISO 22301:2019 - Continuidad del Negocio](#6-iso-223012019---gestion-de-continuidad-del-negocio)
7. [ISO/IEC 20000-1:2018 - Gestion de Servicios de TI](#7-isoiec-20000-12018---gestion-de-servicios-de-ti)
8. [Matriz Consolidada de Cumplimiento](#8-matriz-consolidada-de-cumplimiento)
9. [Plan de Accion Priorizado](#9-plan-de-accion-priorizado)
10. [Referencias](#10-referencias)

---

## 1. Introduccion

### 1.1 Proposito

Este documento establece las mejores practicas de certificaciones ISO aplicables al proyecto **FasMail Panel**, un panel de administracion web para la gestion de servidores de correo electronico. Su objetivo es servir como guia de referencia para alinear el desarrollo, operacion y mantenimiento del sistema con los estandares internacionales de calidad, seguridad, privacidad, continuidad del negocio y gestion de servicios de TI.

### 1.2 Alcance

El alcance de este documento cubre:

- La aplicacion web FasMail Panel (backend en Go, frontend HTML/CSS/JS)
- La infraestructura Docker asociada (PostgreSQL, Redis)
- Los procesos de desarrollo, despliegue y operacion
- La gestion de datos personales de los usuarios del sistema de correo

### 1.3 Audiencia

- Equipo de desarrollo de FasMail Panel
- Administradores de sistemas
- Responsables de calidad y cumplimiento
- Auditores internos y externos

### 1.4 Certificaciones Cubiertas

| Certificacion | Nombre Completo | Relevancia para FasMail |
|---------------|----------------|------------------------|
| **ISO 9001:2015** | Sistema de Gestion de Calidad | Procesos de desarrollo, documentacion, mejora continua |
| **ISO/IEC 27001:2022** | Seguridad de la Informacion | Proteccion de datos de correo, autenticacion, cifrado |
| **ISO/IEC 27701:2019** | Gestion de Privacidad | Datos personales de usuarios, cumplimiento RGPD |
| **ISO 22301:2019** | Continuidad del Negocio | Disponibilidad del servicio de correo |
| **ISO/IEC 20000-1:2018** | Gestion de Servicios de TI | Correo electronico como servicio gestionado |

---

## 2. Resumen del Proyecto FasMail Panel

### 2.1 Arquitectura del Sistema

FasMail Panel es una aplicacion web desarrollada en **Go 1.23** utilizando el framework **Gin** para HTTP. El sistema opera como un panel de administracion para servidores de correo electronico, con la siguiente infraestructura:

- **Backend**: Go 1.23 con Gin Web Framework
- **Base de datos principal**: PostgreSQL 16 (Alpine)
- **Cache/Sesiones**: Redis 7 (Alpine)
- **Contenedores**: Docker con Docker Compose
- **Autenticacion**: JWT con HMAC-SHA256
- **Cifrado de contrasenas**: bcrypt (costo 12)

### 2.2 Mapa de Componentes

| Componente | Ruta | Responsabilidad |
|-----------|------|-----------------|
| Punto de entrada | `cmd/server/main.go` | Servidor HTTP, rutas, middleware, health check, shutdown graceful |
| Autenticacion | `internal/auth/` | JWT, bcrypt, RBAC, sesiones, middleware de acceso |
| Administracion | `internal/admin/` | Dashboard, configuracion del sistema |
| Configuracion | `internal/config/` | Carga de configuracion, valores por defecto, persistencia |
| Base de datos | `internal/database/` | PostgreSQL pool, Redis, migraciones con Goose |
| Docker | `internal/docker/` | Gestion de contenedores PostgreSQL/Redis |
| Instalador | `internal/installer/` | Asistente de instalacion web multi-paso |
| Modelos | `internal/models/` | Repositorios de User, Session, SystemConfig |
| Interfaz web | `web/` | Plantillas HTML, CSS (Pico), JavaScript |
| Migraciones | `migrations/` | Scripts SQL versionados |
| Infraestructura | `Dockerfile`, `docker-compose.yml` | Build multi-stage, orquestacion de servicios |

---

## 3. ISO 9001:2015 - Sistema de Gestion de Calidad (SGC)

La norma ISO 9001:2015 establece los requisitos para un Sistema de Gestion de Calidad (SGC). Para un proyecto de software como FasMail Panel, esto se traduce en procesos de desarrollo controlados, documentacion adecuada, pruebas sistematicas y mejora continua.

### 3.1 Clausula 4 - Contexto de la Organizacion

#### Mejores Practicas

- **Comprension de la organizacion y su contexto**: Documentar el proposito de FasMail Panel como solucion de gestion de correo electronico, identificando las necesidades del mercado y los factores internos/externos que afectan su calidad.
- **Comprension de las partes interesadas**: Identificar y documentar las necesidades de administradores de correo, usuarios finales, proveedores de hosting y reguladores.
- **Alcance del SGC**: Definir claramente que el SGC cubre el desarrollo, despliegue y mantenimiento del panel de administracion y sus servicios asociados.

#### Estado en FasMail Panel

| Requisito | Estado | Detalle |
|-----------|--------|---------|
| Documentacion de contexto | ❌ NO IMPLEMENTADO | No existe documentacion formal del contexto organizacional |
| Identificacion de partes interesadas | ❌ NO IMPLEMENTADO | No hay registro de stakeholders ni sus requisitos |
| Alcance del SGC | ❌ NO IMPLEMENTADO | No hay definicion formal del alcance |

#### Recomendaciones

1. Crear un documento `docs/CONTEXTO_ORGANIZACIONAL.md` que defina el proposito, vision y partes interesadas del proyecto.
2. Mantener un registro de requisitos de partes interesadas que se actualice periodicamente.

### 3.2 Clausula 5 - Liderazgo y Compromiso

#### Mejores Practicas

- **Politica de calidad**: Establecer una politica de calidad documentada que sea comunicada, entendida y aplicada.
- **Roles y responsabilidades**: Definir claramente las responsabilidades del equipo de desarrollo, operaciones y soporte.
- **Enfoque al cliente**: Asegurar que los requisitos de los usuarios del sistema de correo se determinan y se cumplen.

#### Recomendaciones

1. Documentar una politica de calidad en `docs/POLITICA_CALIDAD.md`.
2. Definir una matriz RACI (Responsable, Aprobador, Consultado, Informado) para las actividades del proyecto.

### 3.3 Clausula 6 - Planificacion

#### Mejores Practicas

- **Riesgos y oportunidades**: Mantener un registro de riesgos tecnnicos y de negocio, con planes de mitigacion.
- **Objetivos de calidad**: Definir objetivos medibles como porcentaje de cobertura de pruebas, tiempo de respuesta del sistema, y tasa de defectos.
- **Planificacion de cambios**: Establecer un proceso formal para gestionar cambios en el sistema.

#### Estado en FasMail Panel

| Requisito | Estado | Detalle |
|-----------|--------|---------|
| Registro de riesgos | ❌ NO IMPLEMENTADO | No existe analisis de riesgos documentado |
| Objetivos de calidad medibles | ❌ NO IMPLEMENTADO | No hay KPIs definidos |
| Gestion de cambios | ⚠️ PARCIAL | Las migraciones de BD estan versionadas con Goose (`internal/database/migrate.go`), pero no existe un proceso formal de gestion de cambios de codigo |

### 3.4 Clausula 7 - Soporte y Recursos

#### Mejores Practicas

- **Documentacion**: Mantener documentacion actualizada del sistema, APIs, procesos de despliegue y operacion.
- **Competencias**: Registrar las competencias necesarias del equipo y los planes de formacion.
- **Informacion documentada**: Controlar la creacion, actualizacion y distribucion de documentos.

#### Estado en FasMail Panel

| Requisito | Estado | Detalle |
|-----------|--------|---------|
| Documentacion del sistema | ❌ NO IMPLEMENTADO | No existe README ni documentacion tecnica |
| Documentacion de API | ❌ NO IMPLEMENTADO | Endpoints no documentados |
| Guia de despliegue | ⚠️ PARCIAL | `Makefile` y `docker-compose.yml` proveen automatizacion basica pero sin documentacion explicativa |
| Guia de desarrollo | ❌ NO IMPLEMENTADO | No hay CONTRIBUTING ni guias de estilo |

#### Recomendaciones

1. Crear un `README.md` comprehensivo con instrucciones de instalacion, configuracion y uso.
2. Documentar los endpoints de la API REST.
3. Crear una guia de contribucion (`CONTRIBUTING.md`).

### 3.5 Clausula 8 - Operacion

#### Mejores Practicas

- **Planificacion y control operacional**: Implementar procesos de desarrollo controlados con revision de codigo, pruebas automatizadas y despliegue controlado.
- **Control de productos/servicios externos**: Gestionar las dependencias de software (Go modules) y servicios externos (PostgreSQL, Redis, Docker).
- **Produccion y provision del servicio**: Establecer procedimientos para el despliegue, monitoreo y mantenimiento del sistema.

#### Estado en FasMail Panel

| Requisito | Estado | Detalle |
|-----------|--------|---------|
| Gestion de dependencias | ✅ IMPLEMENTADO | `go.mod` y `go.sum` gestionan 15 dependencias directas y ~73 transitivas |
| Migraciones versionadas | ✅ IMPLEMENTADO | Goose en `internal/database/migrate.go` con 3 migraciones SQL embebidas |
| Build reproducible | ✅ IMPLEMENTADO | `Dockerfile` multi-stage y `docker-compose.yml` para despliegue consistente |
| Pruebas automatizadas | ❌ NO IMPLEMENTADO | `Makefile` define target `test` (`go test ./... -v -race`) pero no existen archivos `*_test.go` en el proyecto |
| Revision de codigo | ❌ NO IMPLEMENTADO | No hay proceso de code review documentado ni pipeline CI/CD |

#### Recomendaciones

1. **Prioridad alta**: Implementar pruebas unitarias para los paquetes criticos:
   - `internal/auth/password_test.go` - Pruebas de hashing y validacion de contrasenas
   - `internal/auth/jwt_test.go` - Pruebas de generacion y validacion de tokens
   - `internal/auth/service_test.go` - Pruebas de autenticacion y sesiones
   - `internal/models/` - Pruebas de repositorios con base de datos de prueba
2. Configurar un pipeline CI/CD (GitHub Actions, GitLab CI) que ejecute pruebas automaticamente.
3. Establecer un flujo de code review obligatorio antes de mergear cambios.

### 3.6 Clausula 9 - Evaluacion del Desempeno

#### Mejores Practicas

- **Monitoreo y medicion**: Implementar metricas de rendimiento del sistema, tiempos de respuesta, errores y disponibilidad.
- **Auditoria interna**: Realizar auditorias periodicas del sistema y procesos.
- **Revision por la direccion**: Revisar periodicamente la efectividad del SGC.

#### Estado en FasMail Panel

| Requisito | Estado | Detalle |
|-----------|--------|---------|
| Health check | ✅ IMPLEMENTADO | Endpoint `/health` en `cmd/server/main.go:91-108` reporta estado de PostgreSQL, Redis e instalacion |
| Docker healthchecks | ✅ IMPLEMENTADO | `Dockerfile:34-35` y `docker-compose.yml:27-31,45-50,61-65` con intervalos de 10-30s |
| Metricas de aplicacion | ❌ NO IMPLEMENTADO | No hay metricas de rendimiento, latencia o uso |
| Logging estructurado | ❌ NO IMPLEMENTADO | Solo `log.Printf` basico, sin niveles ni formato estructurado |
| Auditoria interna | ❌ NO IMPLEMENTADO | No hay proceso de auditoria definido |

#### Recomendaciones

1. Integrar un sistema de metricas (Prometheus + Grafana) para monitorear:
   - Tiempo de respuesta de endpoints
   - Tasa de errores HTTP
   - Uso de conexiones de base de datos (pool min=2, max=10 en `internal/database/postgres.go:19-20`)
   - Sesiones activas en Redis
2. Implementar logging estructurado con niveles (info, warn, error) y formato JSON.

### 3.7 Clausula 10 - Mejora Continua

#### Mejores Practicas

- **No conformidades y acciones correctivas**: Registrar y dar seguimiento a defectos, incidentes y oportunidades de mejora.
- **Mejora continua**: Establecer procesos para identificar e implementar mejoras de forma sistematica.

#### Recomendaciones

1. Utilizar un sistema de issue tracking (GitHub Issues, Jira) para registrar defectos y mejoras.
2. Realizar retrospectivas periodicas del proceso de desarrollo.
3. Mantener un registro de lecciones aprendidas.

---

## 4. ISO/IEC 27001:2022 - Sistema de Gestion de Seguridad de la Informacion (SGSI)

La norma ISO/IEC 27001 es **critica** para FasMail Panel dado que gestiona un servicio de correo electronico, el cual procesa y almacena informacion confidencial de comunicaciones.

### 4.1 Contexto y Alcance (Clausulas 4-5)

#### Mejores Practicas

- **Activos criticos a proteger**:
  - Credenciales de usuarios (email, password hash, tokens)
  - Datos de sesion (IP, user agent)
  - Configuracion del sistema (JWT secret, credenciales de BD)
  - Comunicaciones de correo electronico gestionadas
- **Declaracion de aplicabilidad**: Documentar que controles del Anexo A son aplicables y cuales no, con justificacion.

### 4.2 Evaluacion y Tratamiento de Riesgos (Clausula 6)

#### Matriz de Riesgos para FasMail Panel

| Riesgo | Probabilidad | Impacto | Nivel | Mitigacion |
|--------|-------------|---------|-------|-----------|
| Compromiso de credenciales de usuario | Media | Alto | Alto | bcrypt cost=12, validacion de fuerza de contrasena |
| Interceptacion de tokens de sesion | Alta | Alto | Critico | Falta TLS - mitigacion pendiente |
| Inyeccion SQL | Baja | Alto | Medio | Consultas parametrizadas implementadas |
| Acceso no autorizado al panel | Media | Alto | Alto | JWT + RBAC implementados |
| Compromiso de la base de datos | Media | Alto | Alto | SSL deshabilitado por defecto - brecha |
| Fuga de datos de correo | Media | Alto | Alto | Sin cifrado en reposo - brecha |

### 4.3 Controles del Anexo A - Mapeo al Proyecto

#### A.5 Politicas de Seguridad de la Informacion

| Control | Estado | Detalle |
|---------|--------|---------|
| A.5.1 Politicas para la seguridad de la informacion | ❌ NO IMPLEMENTADO | No existe politica de seguridad documentada |
| A.5.2 Revision de las politicas | ❌ NO IMPLEMENTADO | No hay proceso de revision |

**Recomendacion**: Crear `docs/POLITICA_SEGURIDAD.md` con politica de seguridad formal.

#### A.6 Organizacion de la Seguridad de la Informacion

| Control | Estado | Detalle |
|---------|--------|---------|
| A.6.1 Organizacion interna | ❌ NO IMPLEMENTADO | No hay roles de seguridad definidos |
| A.6.2 Dispositivos moviles y teletrabajo | ❌ NO IMPLEMENTADO | No hay politica de acceso remoto |

#### A.8 Gestion de Activos

| Control | Estado | Detalle |
|---------|--------|---------|
| A.8.1 Inventario de activos | ⚠️ PARCIAL | Dependencias en `go.mod`, pero no hay inventario formal de activos de informacion |
| A.8.2 Clasificacion de la informacion | ❌ NO IMPLEMENTADO | No hay esquema de clasificacion |

**Recomendacion**: Clasificar los datos del sistema:
- **Confidencial**: password_hash, JWT secret, refresh_token_hash, credenciales de BD
- **Interno**: email, display_name, configuracion del sistema
- **Publico**: endpoints de health check

#### A.9 Control de Acceso

| Control | Estado | Archivo | Detalle |
|---------|--------|---------|---------|
| A.9.1 Politica de control de acceso | ✅ IMPLEMENTADO | `internal/auth/middleware.go:12-59` | `AuthRequired()` valida JWT y sesion Redis |
| A.9.1 Control de acceso basado en roles | ✅ IMPLEMENTADO | `internal/auth/middleware.go:61-78` | `AdminRequired()` verifica rol "admin" |
| A.9.2 Gestion de acceso de usuarios | ✅ IMPLEMENTADO | `internal/models/user.go:35-49` | Creacion de usuarios con roles, estados activo/inactivo |
| A.9.3 Responsabilidad de los usuarios | ✅ IMPLEMENTADO | `internal/auth/middleware.go:80-110` | `ForcePasswordChange()` obliga cambio de contrasena inicial |
| A.9.4 Autenticacion segura | ✅ IMPLEMENTADO | `internal/auth/password.go:10-23` | bcrypt con costo 12, validacion de fuerza (8+ chars, mayuscula, minuscula, digito, especial) |
| A.9.4 Gestion de sesiones | ✅ IMPLEMENTADO | `internal/auth/service.go:84-122` | Sesiones en PostgreSQL + Redis con expiracion |
| A.9.4 Invalidacion de sesiones | ✅ IMPLEMENTADO | `internal/auth/service.go:142-189` | Al cambiar contrasena se eliminan todas las sesiones del usuario |
| A.9.4 Rate limiting en login | ❌ NO IMPLEMENTADO | `internal/auth/handler.go:24-69` | No hay limitacion de intentos de login |
| A.9.4 Bloqueo de cuenta | ❌ NO IMPLEMENTADO | - | No hay mecanismo de bloqueo tras intentos fallidos |
| A.9.4 Expiracion de contrasenas | ❌ NO IMPLEMENTADO | - | Campo `password_changed_at` existe (`internal/models/user.go:22`) pero no se usa para forzar cambio periodico |

**Recomendaciones de prioridad alta**:
1. Implementar rate limiting en el endpoint POST `/auth/login` (maximo 5 intentos por IP en 15 minutos).
2. Implementar bloqueo de cuenta temporal tras 5 intentos fallidos consecutivos.
3. Implementar politica de expiracion de contrasenas utilizando el campo `password_changed_at` que ya existe en el modelo User.

#### A.10 Criptografia

| Control | Estado | Archivo | Detalle |
|---------|--------|---------|---------|
| A.10.1 Politica de uso de controles criptograficos | ⚠️ PARCIAL | - | Implementado en codigo pero no documentado como politica |
| A.10.1 Hash de contrasenas | ✅ IMPLEMENTADO | `internal/auth/password.go:10-18` | bcrypt con costo 12 usando `golang.org/x/crypto/bcrypt` |
| A.10.1 Firma de tokens | ✅ IMPLEMENTADO | `internal/auth/jwt.go:58` | JWT con HMAC-SHA256 (`jwt.SigningMethodHS256`) |
| A.10.1 Hash de refresh tokens | ✅ IMPLEMENTADO | `internal/auth/service.go:226-229` | SHA-256 antes de almacenar en base de datos |
| A.10.1 Generacion de secretos | ✅ IMPLEMENTADO | `internal/config/loader.go:129-135` | `crypto/rand` con 64 bytes para JWT secret |
| A.10.1 Permisos de archivos de configuracion | ✅ IMPLEMENTADO | `internal/config/loader.go:122` | Archivo de configuracion con permisos 0600 |
| A.10.1 TLS/HTTPS | ❌ NO IMPLEMENTADO | `cmd/server/main.go:257` | Usa `ListenAndServe` en lugar de `ListenAndServeTLS` |
| A.10.1 SSL para PostgreSQL | ❌ NO IMPLEMENTADO | `internal/config/config.go:69` | `SSLMode` predeterminado en `"disable"` |

**Brechas criticas**:
1. **Sin TLS/HTTPS**: El servidor en `cmd/server/main.go:257` usa `srv.ListenAndServe()` sin cifrado. Todas las comunicaciones (incluyendo credenciales y tokens JWT) viajan en texto plano.
2. **SSL deshabilitado en PostgreSQL**: En `internal/config/config.go:69`, el `SSLMode` predeterminado es `"disable"`, lo que significa que las consultas a la base de datos (incluyendo password hashes) viajan sin cifrado.

**Recomendaciones criticas**:
1. Implementar TLS/HTTPS con certificados (Let's Encrypt o certificados internos). Cambiar `ListenAndServe` por `ListenAndServeTLS`.
2. Cambiar el `SSLMode` predeterminado de `"disable"` a `"require"` o `"verify-full"`.
3. Documentar la politica criptografica del proyecto.

#### A.12 Seguridad de las Operaciones

| Control | Estado | Archivo | Detalle |
|---------|--------|---------|---------|
| A.12.1 Procedimientos operacionales documentados | ⚠️ PARCIAL | `Makefile` | Targets de build/deploy pero sin documentacion de procedimientos |
| A.12.1 Gestion de cambios | ✅ IMPLEMENTADO | `internal/database/migrate.go:14-29` | Migraciones versionadas con Goose, rollback disponible |
| A.12.1 Prevencion de reinstalacion | ✅ IMPLEMENTADO | `cmd/server/main.go:111-137` | Middleware guardia del instalador bloquea acceso post-instalacion |
| A.12.4 Registro de eventos (logging) | ⚠️ PARCIAL | `cmd/server/main.go` | Solo `log.Printf` basico, sin logging de auditoria |
| A.12.4 Tracking de IP y User Agent | ✅ IMPLEMENTADO | `internal/models/session.go:12-20` | Sesiones almacenan `UserAgent` e `IPAddress` |
| A.12.6 Gestion de vulnerabilidades tecnicas | ⚠️ PARCIAL | `Dockerfile:1-13` | Multi-stage build con Alpine reduce superficie de ataque |
| A.12.6 Ejecucion con privilegios minimos | ✅ IMPLEMENTADO | `Dockerfile:20-28` | `adduser -D fasmail` y `USER fasmail` - ejecucion no-root |

**Recomendaciones**:
1. Implementar un sistema de logging de auditoria que registre:
   - Intentos de login (exitosos y fallidos) con IP y timestamp
   - Cambios de contrasena
   - Cambios de configuracion del sistema
   - Acciones administrativas
2. Implementar escaneo automatizado de vulnerabilidades en dependencias (`govulncheck`, `nancy`).
3. Agregar escaneo de imagenes Docker (`trivy`, `grype`).

#### A.13 Seguridad de las Comunicaciones

| Control | Estado | Detalle |
|---------|--------|---------|
| A.13.1 Controles de red | ⚠️ PARCIAL | Red Docker aislada `fasmail-network` en `docker-compose.yml:72-74`, pero sin TLS |
| A.13.1 Seguridad de servicios de red | ❌ NO IMPLEMENTADO | Redis sin contrasena en `docker-compose.yml:55` |
| A.13.2 Seguridad del correo electronico | ❌ NO IMPLEMENTADO | No hay configuracion de DKIM, SPF ni DMARC |

**Recomendaciones**:
1. Configurar Redis con contrasena en produccion (el soporte ya existe en `internal/database/redis.go:23`).
2. Implementar DKIM, SPF y DMARC para autenticacion de correo electronico.
3. Segmentar la red Docker para aislar servicios de base de datos de la red publica.

#### A.14 Desarrollo y Mantenimiento de Sistemas

| Control | Estado | Archivo | Detalle |
|---------|--------|---------|---------|
| A.14.1 Requisitos de seguridad | ⚠️ PARCIAL | - | Seguridad implementada en codigo pero no documentada como requisitos |
| A.14.2 Prevencion de inyeccion SQL | ✅ IMPLEMENTADO | `internal/models/user.go`, `session.go`, `system.go` | Todas las consultas usan placeholders `$1`, `$2`, etc. |
| A.14.2 Cookies HttpOnly | ✅ IMPLEMENTADO | `internal/auth/handler.go:60-61` | `c.SetCookie(..., true)` - ultimo parametro `httpOnly=true` |
| A.14.2 Cookie Secure flag | ❌ NO IMPLEMENTADO | `internal/auth/handler.go:60-61` | `c.SetCookie(..., false, true)` - parametro `secure=false` |
| A.14.2 Proteccion CSRF | ❌ NO IMPLEMENTADO | - | No hay tokens CSRF en formularios POST |
| A.14.2 Headers de seguridad HTTP | ❌ NO IMPLEMENTADO | - | No hay X-Frame-Options, CSP, X-Content-Type-Options, HSTS |
| A.14.2 Validacion de entrada | ⚠️ PARCIAL | `internal/admin/handler.go:59-85` | Settings acepta key/value sin validacion mas alla de keys protegidas |
| A.14.2 Exclusion de password_hash en JSON | ✅ IMPLEMENTADO | `internal/models/user.go:16` | `json:"-"` evita serializacion del hash |

**Brechas criticas**:
1. **Cookie `Secure=false`**: En `internal/auth/handler.go:60`, la cookie `access_token` se establece con `secure=false`, permitiendo envio sobre HTTP no cifrado.
2. **Sin proteccion CSRF**: Los formularios POST de login, cambio de contrasena y settings no tienen proteccion contra Cross-Site Request Forgery.

**Recomendaciones criticas**:
1. Cambiar el flag `Secure` de las cookies a `true` cuando se implemente TLS.
2. Implementar tokens CSRF en todos los formularios POST (Gin tiene middleware disponible).
3. Agregar headers de seguridad HTTP:
   ```
   X-Frame-Options: DENY
   X-Content-Type-Options: nosniff
   X-XSS-Protection: 1; mode=block
   Content-Security-Policy: default-src 'self'
   Strict-Transport-Security: max-age=31536000; includeSubDomains
   ```
4. Validar y sanitizar todas las entradas en el handler de settings.

#### A.16 Gestion de Incidentes de Seguridad

| Control | Estado | Detalle |
|---------|--------|---------|
| A.16.1 Gestion de incidentes | ❌ NO IMPLEMENTADO | No hay procedimiento de respuesta a incidentes |
| A.16.1 Reporte de incidentes | ❌ NO IMPLEMENTADO | No hay mecanismo de notificacion de incidentes |
| A.16.1 Recopilacion de evidencias | ⚠️ PARCIAL | IP y user agent se almacenan en sesiones, pero no hay log de auditoria |

**Recomendaciones**:
1. Crear un procedimiento de respuesta a incidentes de seguridad documentado en `docs/PLAN_RESPUESTA_INCIDENTES.md`.
2. Implementar alertas automaticas para eventos criticos (multiples intentos de login fallidos, cambios de configuracion).

#### A.18 Cumplimiento

| Control | Estado | Detalle |
|---------|--------|---------|
| A.18.1 Identificacion de requisitos legales | ❌ NO IMPLEMENTADO | No hay analisis de requisitos legales aplicables |
| A.18.2 Revision de seguridad de la informacion | ❌ NO IMPLEMENTADO | No hay auditorias ni revisiones de seguridad |

---

## 5. ISO/IEC 27701:2019 - Gestion de Privacidad de la Informacion

La norma ISO 27701 extiende ISO 27001 para incluir requisitos especificos de gestion de privacidad, alineados con regulaciones como el RGPD europeo y la LOPDGDD espanola. Es especialmente relevante para FasMail Panel ya que procesa datos personales de usuarios de correo electronico.

### 5.1 Datos Personales en FasMail Panel

El sistema almacena los siguientes datos personales identificados en el codigo:

| Dato Personal | Ubicacion | Tipo | Base Legal Sugerida |
|---------------|----------|------|-------------------|
| `email` | `internal/models/user.go:15` | Identificador directo | Ejecucion de contrato |
| `display_name` | `internal/models/user.go:17` | Identificador directo | Ejecucion de contrato |
| `ip_address` | `internal/models/session.go:17` | Identificador indirecto | Interes legitimo (seguridad) |
| `user_agent` | `internal/models/session.go:16` | Dato tecnico | Interes legitimo (seguridad) |
| `last_login_at` | `internal/models/user.go:21` | Dato de actividad | Interes legitimo (seguridad) |
| `password_changed_at` | `internal/models/user.go:22` | Dato de actividad | Interes legitimo (seguridad) |
| `created_at` | `internal/models/user.go:23` | Dato temporal | Ejecucion de contrato |

### 5.2 Derechos del Titular de los Datos

| Derecho (RGPD) | Estado | Detalle |
|----------------|--------|---------|
| Acceso (Art. 15) | ❌ NO IMPLEMENTADO | No hay funcionalidad para que un usuario exporte sus datos |
| Rectificacion (Art. 16) | ⚠️ PARCIAL | Solo se puede cambiar contrasena, no email ni display_name |
| Supresion (Art. 17) | ❌ NO IMPLEMENTADO | No hay funcionalidad para eliminar cuenta de usuario |
| Portabilidad (Art. 20) | ❌ NO IMPLEMENTADO | No hay exportacion de datos en formato estructurado |
| Consentimiento informado | ❌ NO IMPLEMENTADO | No hay aviso de privacidad ni solicitud de consentimiento |

### 5.3 Mejores Practicas de Privacidad

#### Recomendaciones

1. **Aviso de privacidad**: Agregar una pagina de politica de privacidad accesible desde el login.
2. **Exportacion de datos**: Implementar un endpoint que permita al usuario descargar sus datos personales en formato JSON.
3. **Eliminacion de cuenta**: Implementar funcionalidad de eliminacion de cuenta que:
   - Elimine el usuario de la tabla `users`
   - Elimine todas las sesiones asociadas de `sessions` y Redis
   - Registre la eliminacion con fines de auditoria (anonimizado)
4. **Retencion de datos**: Definir periodos de retencion para:
   - Sesiones expiradas (actualmente se mantienen indefinidamente en PostgreSQL, aunque hay funcion `DeleteExpired` en `internal/models/session.go:72-78`)
   - Datos de usuarios inactivos
5. **Minimizacion de datos**: Evaluar si todos los datos recopilados son necesarios para el proposito del servicio.

### 5.4 Evaluacion de Impacto en la Privacidad (EIPD/DPIA)

Para un servicio de correo electronico, se recomienda realizar una Evaluacion de Impacto en la Privacidad que considere:

- Volumenes de datos personales procesados
- Transferencias internacionales de datos
- Riesgos para los derechos de los interesados
- Medidas de mitigacion implementadas

---

## 6. ISO 22301:2019 - Gestion de Continuidad del Negocio

El correo electronico es un servicio critico para las organizaciones. ISO 22301 establece los requisitos para garantizar la continuidad del servicio ante incidentes disruptivos.

### 6.1 Analisis de Impacto en el Negocio (BIA)

| Componente | Impacto si falla | Tiempo Maximo de Interrupcion Tolerable |
|-----------|-----------------|--------------------------------------|
| Aplicacion FasMail Panel | Los administradores no pueden gestionar el correo | 4 horas (sugerido) |
| PostgreSQL | Perdida de acceso a configuracion y datos de usuario | 1 hora (sugerido) |
| Redis | Sesiones activas se pierden (reconexion requerida) | 15 minutos (sugerido) |
| Docker Engine | Todos los servicios se detienen | 30 minutos (sugerido) |

### 6.2 Controles de Continuidad Implementados

| Control | Estado | Archivo | Detalle |
|---------|--------|---------|---------|
| Apagado graceful | ✅ IMPLEMENTADO | `cmd/server/main.go:262-286` | Captura SIGINT/SIGTERM, shutdown con timeout de 10s, cierre de pool y Redis |
| Health checks de aplicacion | ✅ IMPLEMENTADO | `cmd/server/main.go:91-108` | Endpoint `/health` con estado de PostgreSQL, Redis e instalacion |
| Health checks de Docker | ✅ IMPLEMENTADO | `docker-compose.yml:27-31,45-50,61-65` | Checks para app (wget), PostgreSQL (pg_isready) y Redis (redis-cli ping) |
| Reinicio automatico | ✅ IMPLEMENTADO | `docker-compose.yml:26,44,60` | `restart: unless-stopped` en los 3 servicios |
| Persistencia de datos | ✅ IMPLEMENTADO | `docker-compose.yml:67-70` | Volumenes Docker: `app-data`, `pgdata`, `redis-data` |
| Persistencia de Redis | ✅ IMPLEMENTADO | `docker-compose.yml:55` | `redis-server --appendonly yes` habilita AOF |
| Connection pooling | ✅ IMPLEMENTADO | `internal/database/postgres.go:19-23` | min=2, max=10 conexiones, lifetime=30min, idle=5min, healthcheck=30s |
| Inicio condicional de servicios | ✅ IMPLEMENTADO | `docker-compose.yml:19-23` | `depends_on` con `condition: service_healthy` asegura orden de arranque |

### 6.3 Brechas de Continuidad

| Control | Estado | Detalle |
|---------|--------|---------|
| Backup automatizado de PostgreSQL | ❌ NO IMPLEMENTADO | No hay procedimiento de backup de `pgdata` |
| Backup de configuracion | ❌ NO IMPLEMENTADO | No hay backup del volumen `app-data` (contiene `config.json`) |
| Replicacion de base de datos | ❌ NO IMPLEMENTADO | PostgreSQL ejecuta como instancia unica |
| Plan de recuperacion ante desastres (DRP) | ❌ NO IMPLEMENTADO | No existe plan documentado |
| RTO/RPO definidos | ❌ NO IMPLEMENTADO | No se han establecido objetivos formales |
| Pruebas de recuperacion | ❌ NO IMPLEMENTADO | No se realizan simulacros |

### 6.4 Recomendaciones de Continuidad

1. **Backup automatizado de PostgreSQL**:
   ```bash
   # Ejemplo de backup diario con pg_dump
   docker exec fasmail-postgres pg_dump -U fasmail fasmail > backup_$(date +%Y%m%d).sql
   ```
   Programar con cron y almacenar en ubicacion externa.

2. **Backup de configuracion**:
   - Incluir el volumen `app-data` en la estrategia de backup (contiene `config.json` con el JWT secret).

3. **Definir RTO y RPO**:
   - **RTO** (Recovery Time Objective): Tiempo maximo aceptable para restaurar el servicio.
   - **RPO** (Recovery Point Objective): Cantidad maxima aceptable de perdida de datos (determina frecuencia de backups).

4. **Documentar procedimiento de recuperacion**:
   Crear `docs/PLAN_RECUPERACION.md` con pasos detallados para restaurar el servicio desde backups.

5. **Considerar alta disponibilidad**:
   - PostgreSQL con replicacion streaming (primario/replica)
   - Redis Sentinel o Redis Cluster para alta disponibilidad de sesiones
   - Multiples instancias de FasMail Panel detras de un load balancer

---

## 7. ISO/IEC 20000-1:2018 - Gestion de Servicios de TI

ISO 20000-1 establece requisitos para un sistema de gestion de servicios de TI (SGS). FasMail Panel, como herramienta de gestion de un servicio de correo electronico, debe alinearse con estas practicas.

### 7.1 Sistema de Gestion de Servicios

#### Mejores Practicas

- **Catalogo de servicios**: Documentar los servicios que FasMail Panel provee.
- **Acuerdos de nivel de servicio (SLA)**: Definir compromisos medibles de disponibilidad, rendimiento y soporte.

#### Estado Actual

| Requisito | Estado | Detalle |
|-----------|--------|---------|
| Catalogo de servicios | ❌ NO IMPLEMENTADO | No hay documentacion de servicios ofrecidos |
| SLAs definidos | ❌ NO IMPLEMENTADO | No hay compromisos de nivel de servicio |
| Politica de gestion de servicios | ❌ NO IMPLEMENTADO | No existe politica formal |

### 7.2 Gestion de la Disponibilidad

| Requisito | Estado | Archivo | Detalle |
|-----------|--------|---------|---------|
| Monitoreo de disponibilidad | ✅ IMPLEMENTADO | `cmd/server/main.go:91-108` | Endpoint `/health` con verificacion de dependencias |
| Healthchecks automaticos | ✅ IMPLEMENTADO | `docker-compose.yml` | Intervalos de 10-30s para cada servicio |
| Reinicio automatico | ✅ IMPLEMENTADO | `docker-compose.yml` | `restart: unless-stopped` |
| Monitoreo externo | ❌ NO IMPLEMENTADO | - | No hay monitoreo externo (UptimeRobot, Pingdom, etc.) |
| Dashboard de disponibilidad | ❌ NO IMPLEMENTADO | - | No hay visualizacion historica de uptime |

### 7.3 Gestion de la Capacidad

| Parametro | Valor Configurado | Archivo |
|-----------|------------------|---------|
| Pool PostgreSQL min | 2 conexiones | `internal/database/postgres.go:19` |
| Pool PostgreSQL max | 10 conexiones | `internal/database/postgres.go:20` |
| Redis pool size | 10 conexiones | `internal/database/redis.go:28` |
| Redis max memory | 256 MB | `docker-compose.yml:55` |
| Redis eviction policy | allkeys-lru | `docker-compose.yml:55` |
| HTTP read timeout | 30 segundos | `internal/config/config.go:61` |
| HTTP write timeout | 30 segundos | `internal/config/config.go:62` |
| JWT access token TTL | 15 minutos | `internal/config/config.go:76` |
| JWT refresh token TTL | 24 horas | `internal/config/config.go:77` |

**Recomendaciones**:
1. Documentar los limites de capacidad y los criterios de escalamiento.
2. Implementar alertas cuando se acerquen los umbrales (pool de conexiones al 80%, Redis al 80% de memoria).

### 7.4 Gestion del Cambio

| Requisito | Estado | Archivo | Detalle |
|-----------|--------|---------|---------|
| Migraciones versionadas | ✅ IMPLEMENTADO | `internal/database/migrate.go` | Goose con migraciones embebidas, soporte de rollback |
| Docker Compose reproducible | ✅ IMPLEMENTADO | `docker-compose.yml` | Infraestructura como codigo |
| Build determinista | ✅ IMPLEMENTADO | `Dockerfile` | Multi-stage build con versiones fijas (Go 1.23, Alpine 3.20) |
| Pipeline CI/CD | ❌ NO IMPLEMENTADO | - | No hay automatizacion de build/test/deploy |
| Registro de cambios (changelog) | ❌ NO IMPLEMENTADO | - | No hay CHANGELOG ni versionamiento semantico |

**Recomendaciones**:
1. Implementar un pipeline CI/CD con las siguientes etapas:
   - Lint (golangci-lint)
   - Test (go test ./... -race)
   - Build (Docker build)
   - Security scan (govulncheck, trivy)
   - Deploy (a staging, luego produccion)
2. Adoptar versionamiento semantico (SemVer) y mantener un CHANGELOG.
3. Establecer un proceso de aprobacion de cambios (pull requests con revision obligatoria).

### 7.5 Gestion de Incidentes y Problemas

| Requisito | Estado | Detalle |
|-----------|--------|---------|
| Registro de incidentes | ❌ NO IMPLEMENTADO | No hay sistema de ticketing |
| Categorizacion de incidentes | ❌ NO IMPLEMENTADO | No hay clasificacion definida |
| Analisis de causa raiz | ❌ NO IMPLEMENTADO | No hay proceso formal |
| Base de conocimiento | ❌ NO IMPLEMENTADO | No hay documentacion de problemas conocidos |

---

## 8. Matriz Consolidada de Cumplimiento

### Resumen por Certificacion

| Certificacion | Controles Evaluados | Implementados | Parciales | No Implementados | % Cumplimiento |
|---------------|-------------------|---------------|-----------|-------------------|---------------|
| ISO 9001:2015 | 15 | 4 | 3 | 8 | 27% |
| ISO/IEC 27001:2022 | 35 | 15 | 8 | 12 | 43% |
| ISO/IEC 27701:2019 | 8 | 0 | 1 | 7 | 0% |
| ISO 22301:2019 | 12 | 7 | 0 | 5 | 58% |
| ISO/IEC 20000-1:2018 | 12 | 5 | 0 | 7 | 42% |

### Controles Criticos - Estado Detallado

| # | Control | ISO | Archivo | Estado | Prioridad |
|---|---------|-----|---------|--------|-----------|
| 1 | TLS/HTTPS | 27001 A.10 | `cmd/server/main.go:257` | ❌ | **CRITICA** |
| 2 | SSL en PostgreSQL | 27001 A.10 | `internal/config/config.go:69` | ❌ | **CRITICA** |
| 3 | Cookie Secure flag | 27001 A.14 | `internal/auth/handler.go:60-61` | ❌ | **CRITICA** |
| 4 | Proteccion CSRF | 27001 A.14 | Formularios POST | ❌ | **CRITICA** |
| 5 | Rate limiting login | 27001 A.9 | `internal/auth/handler.go:24` | ❌ | **CRITICA** |
| 6 | Logging de auditoria | 27001 A.12 | - | ❌ | **ALTA** |
| 7 | Headers de seguridad HTTP | 27001 A.14 | - | ❌ | **ALTA** |
| 8 | Backup automatizado | 22301 | - | ❌ | **ALTA** |
| 9 | Contrasena Redis | 27001 A.13 | `docker-compose.yml:55` | ❌ | **ALTA** |
| 10 | DKIM/SPF/DMARC | 27001 A.13 | - | ❌ | **ALTA** |
| 11 | Pruebas unitarias | 9001 | `Makefile:22` | ❌ | **MEDIA** |
| 12 | Documentacion sistema | 9001 | - | ❌ | **MEDIA** |
| 13 | Exportacion datos usuario | 27701 | - | ❌ | **MEDIA** |
| 14 | Eliminacion de cuenta | 27701 | - | ❌ | **MEDIA** |
| 15 | Pipeline CI/CD | 9001, 20000 | - | ❌ | **MEDIA** |
| 16 | Monitoreo externo | 22301, 20000 | - | ❌ | **MEDIA** |
| 17 | Expiracion de contrasenas | 27001 A.9 | - | ❌ | **MEDIA** |
| 18 | Bloqueo de cuenta | 27001 A.9 | - | ❌ | **MEDIA** |
| 19 | Plan de recuperacion | 22301 | - | ❌ | **BAJA** |
| 20 | SLAs definidos | 20000 | - | ❌ | **BAJA** |

---

## 9. Plan de Accion Priorizado

### 9.1 Prioridad Critica - Implementar Inmediatamente

Estas brechas representan riesgos de seguridad activos que deben abordarse antes de cualquier despliegue en produccion.

1. **Habilitar TLS/HTTPS**
   - Archivo: `cmd/server/main.go`
   - Cambio: Reemplazar `srv.ListenAndServe()` por `srv.ListenAndServeTLS(certFile, keyFile)`
   - Agregar configuracion de certificados en `config.Config`
   - Integrar con Let's Encrypt o certificados internos

2. **Activar SSL en conexion a PostgreSQL**
   - Archivo: `internal/config/config.go:69`
   - Cambio: Modificar `SSLMode` predeterminado de `"disable"` a `"require"`
   - Generar certificados para PostgreSQL

3. **Activar flag Secure en cookies**
   - Archivo: `internal/auth/handler.go:60-61,80-81,137-138`
   - Cambio: Modificar el parametro `secure` de `false` a `true` en todas las llamadas `c.SetCookie()`

4. **Implementar proteccion CSRF**
   - Agregar middleware CSRF de Gin
   - Incluir tokens CSRF en todos los formularios POST (login, cambio de contrasena, settings)

5. **Implementar rate limiting en login**
   - Archivo: `internal/auth/handler.go`
   - Agregar middleware de rate limiting al endpoint POST `/auth/login`
   - Limite sugerido: 5 intentos por IP en 15 minutos

### 9.2 Prioridad Alta - Implementar a Corto Plazo

6. **Implementar logging de auditoria**
   - Crear paquete `internal/audit/` con logger estructurado
   - Registrar eventos: login, logout, cambio de contrasena, cambios de configuracion
   - Almacenar en tabla PostgreSQL dedicada o archivo de log rotado

7. **Agregar headers de seguridad HTTP**
   - Crear middleware Gin que inyecte: X-Frame-Options, CSP, X-Content-Type-Options, HSTS

8. **Configurar backup automatizado**
   - Script de backup para PostgreSQL (pg_dump) y volumen app-data
   - Programar con cron (diario como minimo)
   - Almacenar en ubicacion externa/remota

9. **Configurar contrasena de Redis**
   - Archivo: `docker-compose.yml` - Agregar `--requirepass` al comando Redis
   - Configurar `FASMAIL_REDIS_PASSWORD` en variables de entorno

10. **Configurar DKIM, SPF y DMARC**
    - Generar claves DKIM
    - Configurar registros DNS SPF y DMARC
    - Implementar firma de correos salientes

### 9.3 Prioridad Media - Implementar a Mediano Plazo

11. **Implementar pruebas unitarias**
    - Comenzar con paquetes criticos: `auth/`, `models/`, `config/`
    - Objetivo: cobertura minima del 70% en paquetes de seguridad
    - Integrar en pipeline CI/CD

12. **Implementar derechos de datos personales (RGPD)**
    - Endpoint de exportacion de datos del usuario (JSON)
    - Funcionalidad de eliminacion de cuenta
    - Aviso de privacidad en la interfaz

13. **Definir SLAs y metricas de servicio**
    - Documentar compromisos de disponibilidad (ej. 99.5%)
    - Implementar metricas Prometheus para latencia, errores, saturacion

14. **Configurar monitoreo externo**
    - Integrar con servicio de monitoreo (UptimeRobot, Healthchecks.io)
    - Configurar alertas por correo/SMS

15. **Implementar politica de expiracion de contrasenas**
    - Usar campo `password_changed_at` existente en `internal/models/user.go:22`
    - Forzar cambio cada 90 dias (configurable)

16. **Configurar pipeline CI/CD**
    - GitHub Actions o GitLab CI
    - Etapas: lint, test, build, security scan, deploy

### 9.4 Prioridad Baja - Mejora Continua

17. **Documentar plan de respuesta a incidentes**
    - Crear `docs/PLAN_RESPUESTA_INCIDENTES.md`
    - Definir roles, escalacion, comunicacion

18. **Realizar pruebas de penetracion**
    - Ejecutar analisis de seguridad periodico
    - Herramientas recomendadas: OWASP ZAP, Burp Suite

19. **Programa de formacion en seguridad**
    - Capacitar al equipo en OWASP Top 10
    - Establecer practicas de desarrollo seguro

20. **Implementar replicacion de base de datos**
    - PostgreSQL streaming replication
    - Redis Sentinel para alta disponibilidad

---

## 10. Referencias

### Normas ISO

- **ISO 9001:2015** - Sistemas de gestion de la calidad - Requisitos
- **ISO/IEC 27001:2022** - Seguridad de la informacion, ciberseguridad y proteccion de la privacidad - Sistemas de gestion de la seguridad de la informacion - Requisitos
- **ISO/IEC 27701:2019** - Tecnicas de seguridad - Extension de ISO/IEC 27001 e ISO/IEC 27002 para la gestion de la privacidad de la informacion
- **ISO 22301:2019** - Seguridad y resiliencia - Sistemas de gestion de la continuidad del negocio - Requisitos
- **ISO/IEC 20000-1:2018** - Tecnologia de la informacion - Gestion de servicios - Parte 1: Requisitos del sistema de gestion de servicios

### Recursos Tecnicos

- **OWASP Top 10 (2021)** - Las 10 vulnerabilidades de seguridad web mas criticas
- **CIS Docker Benchmark** - Guia de configuracion segura de Docker
- **NIST Cybersecurity Framework** - Marco de ciberseguridad del NIST
- **Go Security Best Practices** - Practicas de seguridad para desarrollo en Go

### Regulaciones de Privacidad

- **RGPD (Reglamento General de Proteccion de Datos)** - Reglamento (UE) 2016/679
- **LOPDGDD** - Ley Organica 3/2018 de Proteccion de Datos Personales y Garantia de los Derechos Digitales

---

*Documento generado para FasMail Panel. Ultima actualizacion: Febrero 2026.*
*Este documento debe revisarse y actualizarse al menos una vez al ano o ante cambios significativos en el sistema.*
