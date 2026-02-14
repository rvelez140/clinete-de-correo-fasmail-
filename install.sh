#!/usr/bin/env bash
# =============================================================================
# FasMail Panel - Script de Auto-Instalación con Detección de Puertos
# =============================================================================
# Uso desde una VPS limpia:
#   curl -fsSL https://raw.githubusercontent.com/rvelez140/clinete-de-correo-fasmail-/main/install.sh | bash
#
# O clonando primero:
#   git clone https://github.com/rvelez140/clinete-de-correo-fasmail-.git
#   cd clinete-de-correo-fasmail-
#   ./install.sh
#
# Forzar un puerto especifico:
#   APP_PORT=9090 ./install.sh
# =============================================================================

set -euo pipefail

# Detectar si se ejecuta desde pipe (curl | bash) o interactivo
if [ ! -t 0 ]; then
    INTERACTIVE=false
else
    INTERACTIVE=true
fi

# ── Constantes ───────────────────────────────────────────────────────────────
REPO_URL="https://github.com/rvelez140/clinete-de-correo-fasmail-.git"
REPO_DIR="clinete-de-correo-fasmail-"
APP_NAME="FasMail Panel"
APP_PORT=${APP_PORT:-8080}
MAX_PORT_ATTEMPTS=50
HEALTH_TIMEOUT=120

# ── Colores ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# ── Funciones de log ─────────────────────────────────────────────────────────
info()    { printf "${BLUE}[INFO]${NC}    %s\n" "$1"; }
success() { printf "${GREEN}[OK]${NC}      %s\n" "$1"; }
warn()    { printf "${YELLOW}[WARN]${NC}    %s\n" "$1"; }
error()   { printf "${RED}[ERROR]${NC}   %s\n" "$1" >&2; }

banner() {
    printf "\n${CYAN}${BOLD}"
    printf "╔══════════════════════════════════════════════════════════╗\n"
    printf "║             FasMail Panel - Auto Installer              ║\n"
    printf "║          Panel de Administracion de Correo              ║\n"
    printf "║        Con Auto-Deteccion de Puertos Disponibles        ║\n"
    printf "╚══════════════════════════════════════════════════════════╝\n"
    printf "${NC}\n"
}

# ── Verificar prerequisitos ──────────────────────────────────────────────────
check_requirements() {
    info "Verificando prerequisitos..."
    local missing=0

    # Docker
    if command -v docker &>/dev/null; then
        local docker_version
        docker_version=$(docker --version 2>/dev/null | head -1)
        success "Docker encontrado: $docker_version"
    else
        error "Docker no esta instalado."
        error "Instala Docker: https://docs.docker.com/engine/install/"
        missing=1
    fi

    # Docker Compose (v2 plugin)
    if docker compose version &>/dev/null 2>&1; then
        local compose_version
        compose_version=$(docker compose version 2>/dev/null | head -1)
        success "Docker Compose encontrado: $compose_version"
    else
        error "Docker Compose (v2) no esta disponible."
        error "Instala Docker Compose: https://docs.docker.com/compose/install/"
        missing=1
    fi

    # Git
    if command -v git &>/dev/null; then
        success "Git encontrado: $(git --version)"
    else
        error "Git no esta instalado."
        error "Instala Git: apt install git / yum install git"
        missing=1
    fi

    # Docker daemon running
    if docker info &>/dev/null 2>&1; then
        success "Docker daemon esta corriendo"
    else
        error "Docker daemon no esta corriendo. Ejecuta: sudo systemctl start docker"
        missing=1
    fi

    if [ "$missing" -eq 1 ]; then
        printf "\n"
        error "Faltan prerequisitos. Instala lo necesario e intenta de nuevo."
        exit 1
    fi

    success "Todos los prerequisitos estan disponibles"
    printf "\n"
}

# ── Clonar o detectar repositorio ────────────────────────────────────────────
clone_or_detect_repo() {
    # Si ya estamos dentro del repo (existe docker-compose.yml y Dockerfile)
    if [ -f "docker-compose.yml" ] && [ -f "Dockerfile" ] && [ -f "go.mod" ]; then
        success "Repositorio detectado en el directorio actual: $(pwd)"
        return 0
    fi

    # Si existe el subdirectorio del repo
    if [ -d "$REPO_DIR" ]; then
        info "Directorio $REPO_DIR encontrado. Actualizando..."
        cd "$REPO_DIR"
        git pull origin main || warn "No se pudo actualizar. Usando version existente."
        success "Repositorio listo en: $(pwd)"
        return 0
    fi

    # Clonar
    info "Clonando repositorio desde GitHub..."
    git clone "$REPO_URL"
    cd "$REPO_DIR"
    success "Repositorio clonado en: $(pwd)"
}

# ── Limpiar contenedores previos si existen ──────────────────────────────────
cleanup_existing() {
    if docker compose ps -q 2>/dev/null | grep -q .; then
        warn "Contenedores FasMail existentes detectados. Deteniendo..."
        docker compose down 2>/dev/null || true
        success "Contenedores anteriores detenidos"
        sleep 2
    fi
}

# ── Verificar si un puerto esta disponible ───────────────────────────────────
check_port_available() {
    local port=$1

    # Metodo 1: ss (Linux moderno, mas comun en VPS Ubuntu/Debian)
    if command -v ss &>/dev/null; then
        if ss -tln 2>/dev/null | grep -qE ":${port}\b"; then
            return 1
        fi
        return 0
    fi

    # Metodo 2: netstat (sistemas mas antiguos)
    if command -v netstat &>/dev/null; then
        if netstat -tln 2>/dev/null | grep -qE ":${port}\b"; then
            return 1
        fi
        return 0
    fi

    # Metodo 3: /dev/tcp bash built-in (fallback universal)
    if (echo >/dev/tcp/localhost/"${port}") &>/dev/null; then
        return 1
    fi

    return 0
}

# ── Encontrar un puerto disponible ───────────────────────────────────────────
find_available_port() {
    local preferred_port=${1:-8080}
    local port=$preferred_port

    for ((i=0; i<MAX_PORT_ATTEMPTS; i++)); do
        if check_port_available "$port"; then
            echo "$port"
            return 0
        fi
        warn "Puerto $port esta en uso, probando siguiente..."
        port=$((port + 1))
    done

    error "No se encontro un puerto disponible en el rango ${preferred_port}-$((preferred_port + MAX_PORT_ATTEMPTS - 1))"
    exit 1
}

# ── Detectar y asignar puerto ────────────────────────────────────────────────
detect_port() {
    printf "\n"
    info "=== Auto-Deteccion de Puertos ==="
    info "Verificando disponibilidad del puerto ${APP_PORT}..."

    if check_port_available "$APP_PORT"; then
        success "Puerto ${APP_PORT} esta disponible"
    else
        warn "Puerto ${APP_PORT} esta en uso por otro servicio"
        info "Buscando puerto alternativo disponible..."
        APP_PORT=$(find_available_port "$APP_PORT")
        success "Puerto alternativo encontrado: ${APP_PORT}"
    fi

    info "FasMail Panel usara el puerto: ${APP_PORT}"
    printf "\n"
}

# ── Generar secretos seguros ─────────────────────────────────────────────────
generate_secret() {
    local length=${1:-32}
    if command -v openssl &>/dev/null; then
        openssl rand -base64 "$length" | tr -d '/+=' | head -c "$length"
    else
        tr -dc 'A-Za-z0-9' < /dev/urandom | head -c "$length"
    fi
}

# ── Crear archivo .env ───────────────────────────────────────────────────────
create_env_file() {
    if [ -f ".env" ]; then
        if [ "$INTERACTIVE" = true ]; then
            warn "Archivo .env ya existe."
            printf "    Deseas sobreescribirlo? (s/N): "
            read -r response
            if [[ ! "$response" =~ ^[sS]$ ]]; then
                info "Usando .env existente"
                if grep -q '^APP_PORT=' .env 2>/dev/null; then
                    APP_PORT=$(grep '^APP_PORT=' .env | cut -d= -f2)
                    info "Puerto del .env existente: ${APP_PORT}"
                fi
                return 0
            fi
        else
            warn "Archivo .env ya existe. En modo no-interactivo, se sobreescribe."
        fi
        cp .env ".env.backup.$(date +%Y%m%d_%H%M%S)"
        success "Backup del .env anterior creado"
    fi

    info "Generando secretos seguros..."
    local db_password
    local jwt_secret
    db_password=$(generate_secret 32)
    jwt_secret=$(generate_secret 48)

    cat > .env <<EOF
# FasMail Panel - Configuracion generada automaticamente
# Fecha: $(date -u +"%Y-%m-%d %H:%M:%S UTC")

# Puerto del host (auto-detectado por el instalador)
APP_PORT=${APP_PORT}

# Contrasena de PostgreSQL (generada automaticamente)
DB_PASSWORD=${db_password}

# Secreto JWT para firmar tokens (generado automaticamente)
JWT_SECRET=${jwt_secret}
EOF

    chmod 600 .env
    success "Archivo .env creado con secretos seguros (permisos: 600)"
    success "Puerto configurado en .env: ${APP_PORT}"
}

# ── Construir y levantar servicios ───────────────────────────────────────────
build_and_start() {
    info "Construyendo imagen Docker y levantando servicios..."
    info "Esto puede tardar unos minutos en la primera ejecucion..."
    info "El panel estara disponible en el puerto: ${APP_PORT}"
    printf "\n"

    docker compose up --build -d

    printf "\n"
    success "Contenedores iniciados"
}

# ── Esperar a que la app este saludable ──────────────────────────────────────
wait_for_healthy() {
    info "Esperando a que los servicios esten listos (max ${HEALTH_TIMEOUT}s)..."

    local elapsed=0
    local interval=5

    while [ $elapsed -lt $HEALTH_TIMEOUT ]; do
        local status
        status=$(docker inspect --format='{{.State.Health.Status}}' fasmail-panel 2>/dev/null || echo "starting")

        if [ "$status" = "healthy" ]; then
            printf "\n"
            success "FasMail Panel esta corriendo y saludable"
            return 0
        fi

        printf "."
        sleep $interval
        elapsed=$((elapsed + interval))
    done

    printf "\n"
    warn "Timeout esperando health check. Verificando estado..."
    docker compose ps
    docker compose logs --tail 20 app
    warn "La aplicacion puede seguir iniciando. Verifica con: docker compose logs -f"
}

# ── Detectar IP publica ──────────────────────────────────────────────────────
get_public_ip() {
    local ip=""
    # Intentar varios servicios para obtener la IP publica
    for service in "ifconfig.me" "api.ipify.org" "icanhazip.com"; do
        ip=$(curl -s --max-time 5 "$service" 2>/dev/null || true)
        if [[ "$ip" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            echo "$ip"
            return 0
        fi
    done
    echo "localhost"
}

# ── Mostrar resumen final ────────────────────────────────────────────────────
show_summary() {
    local ip
    ip=$(get_public_ip)

    printf "\n"
    printf "${GREEN}${BOLD}"
    printf "╔══════════════════════════════════════════════════════════╗\n"
    printf "║           FasMail Panel - Instalacion Completa          ║\n"
    printf "╚══════════════════════════════════════════════════════════╝\n"
    printf "${NC}\n"

    printf "${BOLD}Puerto asignado:${NC}\n"
    if [ "$APP_PORT" -ne 8080 ] 2>/dev/null; then
        printf "  ${YELLOW}NOTA: El puerto por defecto (8080) estaba en uso.${NC}\n"
        printf "  ${YELLOW}Se asigno automaticamente el puerto: ${APP_PORT}${NC}\n\n"
    fi

    printf "${BOLD}Acceso:${NC}\n"
    printf "  URL:  ${CYAN}http://%s:%s${NC}\n" "$ip" "$APP_PORT"
    if [ "$ip" = "localhost" ]; then
        printf "  URL:  ${CYAN}http://localhost:%s${NC}\n" "$APP_PORT"
    fi
    printf "\n"

    printf "${BOLD}Primer acceso:${NC}\n"
    printf "  1. Abre la URL en tu navegador\n"
    printf "  2. Se mostrara el wizard de instalacion\n"
    printf "  3. Configura la base de datos, Docker y crea tu usuario admin\n"
    printf "\n"

    printf "${BOLD}Datos de conexion a PostgreSQL (ya configurados):${NC}\n"
    printf "  Host:     postgres (interno) / localhost:5432 (externo)\n"
    printf "  Usuario:  fasmail\n"
    printf "  Base:     fasmail\n"
    printf "  Password: (ver archivo .env)\n"
    printf "\n"

    printf "${BOLD}Comandos utiles:${NC}\n"
    printf "  Ver logs:          ${CYAN}docker compose logs -f${NC}\n"
    printf "  Reiniciar:         ${CYAN}docker compose restart${NC}\n"
    printf "  Detener:           ${CYAN}docker compose down${NC}\n"
    printf "  Reconstruir:       ${CYAN}docker compose up --build -d${NC}\n"
    printf "  Estado:            ${CYAN}docker compose ps${NC}\n"
    printf "\n"

    printf "${BOLD}Archivos importantes:${NC}\n"
    printf "  Configuracion:     ${CYAN}.env${NC} (puerto: ${APP_PORT})\n"
    printf "  Docker Compose:    ${CYAN}docker-compose.yml${NC}\n"
    printf "\n"

    printf "${YELLOW}SEGURIDAD: Las contrasenas fueron generadas automaticamente.${NC}\n"
    printf "${YELLOW}Revisa el archivo .env para ver los valores generados.${NC}\n"
    printf "${YELLOW}Nunca compartas el archivo .env ni lo subas a repositorios publicos.${NC}\n"
    printf "\n"
}

# ── Main ─────────────────────────────────────────────────────────────────────
main() {
    banner
    check_requirements
    clone_or_detect_repo
    cleanup_existing
    detect_port
    create_env_file
    build_and_start
    wait_for_healthy
    show_summary
}

main "$@"
