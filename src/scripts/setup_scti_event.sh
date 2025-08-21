#!/bin/bash

# SCTI Event Setup Script - Simple Bash Version
# This script automates the creation of the SCTI 2025 event with multiple activities

BASE_URL="http://127.0.0.1:8080"
SHOW_REQUEST_BODIES=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --show-request-bodies)
            SHOW_REQUEST_BODIES=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--show-request-bodies]"
            exit 1
            ;;
    esac
done

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Simple logging function
log() {
    echo -e "[$(date +%H:%M:%S)] $1"
}

log_success() {
    log "${GREEN}✅ $1${NC}"
}

log_error() {
    log "${RED}❌ $1${NC}"
}

log_info() {
    log "${BLUE}ℹ️  $1${NC}"
}

log_warn() {
    log "${YELLOW}⚠️  $1${NC}"
}

# Load environment variables from .env file
load_env() {
    log_info "Loading environment variables..."
    
    if [ ! -f "../.env" ]; then
        log_error ".env file not found in src directory"
        log_warn "Please create a .env file with your credentials:"
        log_info "SCTI_EMAIL=your_email@example.com"
        log_info "MASTER_USER_PASS=your_password"
        exit 1
    fi
    
    # Source the .env file, properly handling comments and quotes
    while IFS= read -r line; do
        # Skip empty lines and comments
        if [[ -n "$line" && ! "$line" =~ ^[[:space:]]*# ]]; then
            # Remove any surrounding quotes from the values
            line=$(echo "$line" | sed 's/^[[:space:]]*//' | sed 's/[[:space:]]*$//')
            if [[ "$line" =~ ^([^=]+)=(.*)$ ]]; then
                key="${BASH_REMATCH[1]}"
                value="${BASH_REMATCH[2]}"
                # Remove surrounding quotes if they exist
                value=$(echo "$value" | sed 's/^"\(.*\)"$/\1/' | sed "s/^'\(.*\)'$/\1/")
                export "$key=$value"
            fi
        fi
    done < "../.env"
    
    # Validate required variables
    if [ -z "$SCTI_EMAIL" ] || [ -z "$MASTER_USER_PASS" ]; then
        log_error "Missing required environment variables"
        log_warn "Required: SCTI_EMAIL, MASTER_USER_PASS"
        exit 1
    fi
    
    log_success "Environment variables loaded successfully"
}

# Login to get access token
login() {
    log_info "Logging in..."
    
    # Create proper JSON payload using jq to avoid any quoting issues
    local login_data=$(jq -n \
        --arg email "$SCTI_EMAIL" \
        --arg password "$MASTER_USER_PASS" \
        '{
            email: $email,
            password: $password
        }')
    
    # Show request body if flag is enabled
    if [ "$SHOW_REQUEST_BODIES" = true ]; then
        log_info "Login request body (raw):"
        echo "$login_data"
        echo
        log_info "Login request body (formatted):"
        echo "$login_data" | jq . 2>/dev/null || echo "jq failed, raw data above"
        echo
    fi
    
    local response=$(curl -s -X POST "$BASE_URL/login" \
        -H "Content-Type: application/json" \
        -d "$login_data")
    
    if [ $? -eq 0 ]; then
        local success=$(echo "$response" | jq -r '.success')
        if [ "$success" = "true" ]; then
            ACCESS_TOKEN=$(echo "$response" | jq -r '.data.access_token')
            REFRESH_TOKEN=$(echo "$response" | jq -r '.data.refresh_token')
            log_success "Login successful"
            return 0
        else
            local message=$(echo "$response" | jq -r '.message // "No message"')
            local errors=$(echo "$response" | jq -r '.errors[]? // empty' | tr '\n' ' ')
            log_error "Login failed: $message"
            if [ -n "$errors" ]; then
                log_error "Errors: $errors"
            fi
            log_info "Full response: $response"
            return 1
        fi
    else
        log_error "Login request failed"
        return 1
    fi
}

# Create SCTI Event
create_event() {
    log_info "Creating SCTI Event..."
    
    local event_data='{
        "name": "Semana de Ciência e Tecnologia da Informação",
        "slug": "scti",
        "description": "Bem vindos à SCTI! Uma semana repleta de palestras, minicursos e atividades sobre tecnologia da informação.",
        "location": "UENF - Universidade Estadual do Norte Fluminense",
        "start_date": "2025-09-01T00:00:00Z",
        "end_date": "2025-09-05T23:59:59Z",
        "is_hidden": false,
        "is_blocked": false,
        "max_tokens_per_user": 5
    }'
    
    # Show request body if flag is enabled
    if [ "$SHOW_REQUEST_BODIES" = true ]; then
        log_info "Event creation request body:"
        echo "$event_data" | jq .
        echo
    fi
    
    local response=$(curl -s -X POST "$BASE_URL/events" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Refresh: Bearer $REFRESH_TOKEN" \
        -d "$event_data")
    
    if [ $? -eq 0 ]; then
        local success=$(echo "$response" | jq -r '.success')
        if [ "$success" = "true" ]; then
            log_success "Event created successfully"
            return 0
        else
            local message=$(echo "$response" | jq -r '.message // "No message"')
            local errors=$(echo "$response" | jq -r '.errors[]? // empty' | tr '\n' ' ')
            log_error "Event creation failed: $message"
            if [ -n "$errors" ]; then
                log_error "Errors: $errors"
            fi
            log_info "Full response: $response"
            return 1
        fi
    else
        log_error "Event creation request failed"
        return 1
    fi
}

# Create Activity
create_activity() {
    local name="$1"
    local description="$2"
    local speaker="$3"
    local location="$4"
    local type="$5"
    local start_time="$6"
    local end_time="$7"
    local has_unlimited_capacity="$8"
    local max_capacity="$9"
    local is_mandatory="${10}"
    local has_fee="${11}"
    
    log_info "Creating: $name"
    
    # Generate a simple slug from the name
    local slug=$(echo "$name" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-\|-$//g')
    
    # Use jq to properly escape JSON strings and avoid quoting issues
    local activity_data=$(jq -n \
        --arg name "$name" \
        --arg description "$description" \
        --arg speaker "$speaker" \
        --arg location "$location" \
        --arg type "$type" \
        --arg start_time "$start_time" \
        --arg end_time "$end_time" \
        --argjson has_unlimited_capacity "$has_unlimited_capacity" \
        --argjson max_capacity "$max_capacity" \
        --argjson is_mandatory "$is_mandatory" \
        --argjson has_fee "$has_fee" \
        --arg slug "$slug" \
        '{
            name: $name,
            description: $description,
            speaker: $speaker,
            location: $location,
            type: $type,
            start_time: $start_time,
            end_time: $end_time,
            has_unlimited_capacity: $has_unlimited_capacity,
            max_capacity: $max_capacity,
            is_mandatory: $is_mandatory,
            has_fee: $has_fee,
            is_standalone: false,
            standalone_slug: $slug,
            is_hidden: false,
            is_blocked: false
        }')
    
    # Show request body if flag is enabled
    if [ "$SHOW_REQUEST_BODIES" = true ]; then
        log_info "Activity creation request body for '$name':"
        echo "$activity_data" | jq .
        echo
    fi
    
    local response=$(curl -s -X POST "$BASE_URL/events/scti/activity" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Refresh: Bearer $REFRESH_TOKEN" \
        -d "$activity_data")
    
    if [ $? -eq 0 ]; then
        local success=$(echo "$response" | jq -r '.success')
        if [ "$success" = "true" ]; then
            log_success "Created: $name"
            return 0
        else
            local message=$(echo "$response" | jq -r '.message // "No message"')
            local errors=$(echo "$response" | jq -r '.errors[]? // empty' | tr '\n' ' ')
            log_error "Failed: $name - $message"
            if [ -n "$errors" ]; then
                log_error "Errors: $errors"
            fi
            log_info "Response: $response"
            return 1
        fi
    else
        log_error "Failed: $name - Request failed"
        return 1
    fi
}

# Create all activities
create_all_activities() {
    log_info "Creating activities..."
    
    local success_count=0
    local total_count=0
    
    # Day 1 - September 1, 2025
    create_activity "Abertura da Semana Acadêmica" "Cerimônia de abertura com autoridades e apresentação da semana" "PROGRAD" "Auditório Principal" "palestra" "2025-09-01T09:00:00Z" "2025-09-01T12:00:00Z" true 0 true false && ((success_count++))
    ((total_count++))
    
    create_activity "Mesa redonda sobre estágios" "Conectar estudantes de Ciência da Computação com o mercado de trabalho através de experiências reais de estágio." "Estagiários de Computação" "Cine Darcy" "palestra" "2025-09-01T14:00:00Z" "2025-09-01T15:30:00Z" true 0 true false && ((success_count++))
    ((total_count++))
    
    create_activity "Hackathon" "SEM DESCRIÇÂO" "Comissão SCTI" "Cine Darcy" "2025-09-01T16:00:00Z" "2025-09-01T18:00:00Z" true 0 true false && ((success_count++))
    ((total_count++))
    
    # Day 2 - September 2, 2025
    create_activity "Curso Prático de Pentest usando Kali Linux" "TEMP DESC" "Prof. Vinicius Barcelos" "INF-1" "minicurso" "2025-09-02T08:30:00Z" "2025-09-02T12:00:00Z" false 25 false true && ((success_count++))
    ((total_count++))
    
    create_activity "MovieTracker: Criando um app com React Native" "TEMP DESC" "Renan Souza Oliveira" "INF-2" "minicurso" "2025-09-02T08:30:00Z" "2025-09-02T12:00:00Z" false 25 false true && ((success_count++))
    ((total_count++))
    
    create_activity "Desburocratizando o mercado de trabalho" "TEMP DESC" "Patrick Pereira" "Cine Darcy" "palestra" "2025-09-02T14:00:00Z" "2025-09-02T15:30:00Z" true 0 true false  && ((success_count++)) 
    ((total_count++))

    create_activity "Como avaliar um modelo de machine learning e seus desafios." "TEMP DESC" "Clébio Júnior" "Cine Darcy" "palestra" "2025-09-02T16:00:00Z" "2025-09-02T18:00:00Z" true 0 true false  && ((success_count++)) 
    ((total_count++))

    # Day 3 - September 3, 2025
    create_activity "Engenharia e Ciência de Dados com Big Data: Prática com PySpark no Databricks" "TEMP DESC" "João Paulo Seixas" "INF-1" "minicurso" "2025-09-03T08:30:00Z" "2025-09-03T12:00:00Z" false 25 false true && ((success_count++))
    ((total_count++))
    
    create_activity "Dia a dia de desenvolvimento web: Entendo como funciona na pratica" "TEMP DESC" "Jhulian Pereira Manhães" "INF-2" "minicurso" "2025-09-03T08:30:00Z" "2025-09-03T12:00:00Z" false 25 false true && ((success_count++))
    ((total_count++))
    
    create_activity "Inteligência Artificial no Trabalho: Aplicações em Ambientes Cloud" "TEMP DESC" "João Paulo Seixas" "Cine Darcy" "palestra" "2025-09-03T14:00:00Z" "2025-09-03T15:30:00Z" true 0 true false && ((success_count++))
    ((total_count++))

    create_activity "Matemática Aplicada no Ofício" "TEMP DESC" "Prof. João Luiz" "Cine Darcy" "palestra" "2025-09-03T16:00:00Z" "2025-09-03T18:00:00Z" true 0 true false && ((success_count++))
    ((total_count++))

    # Day 4 - September 4, 2025
    create_activity "Introdução à UX com IA" "TEMP DESC" "Diana de Sales" "INF-2" "minicurso" "2025-09-04T08:30:00Z" "2025-09-04T12:00:00Z" false 25 false true && ((success_count++))
    ((total_count++))
    
    create_activity "Montagem e Desmontagem de PC" "TEMP DESC" "Prof. Luiz Ramirez" "INF-1" "minicurso" "2025-09-04T08:30:00Z" "2025-09-04T12:00:00Z" false 25 false true && ((success_count++))
    ((total_count++))
    
    create_activity "Erlang, MCP e Kubernetes: Lições de um Sistema Distribuído em Produção" "TEMP DESC" "Zoey de Souza" "Cine Darcy" "palestra" "2025-09-04T14:00:00Z" "2025-09-04T15:30:00Z" true 0 true false && ((success_count++))
    ((total_count++))
    
    create_activity "Palestra do Mano Deyvin" "TEMP DESC" "Mano Deyvin" "Cine Darcy" "palestra" "2025-09-04T16:00:00Z" "2025-09-04T18:00:00Z" true 0 true false && ((success_count++))
    ((total_count++))

    # Day 5 - September 5, 2025
    create_activity "DevOps Desmistificado: Construindo sua Pipeline do Zero" "TEMP DESC" "Zoey de Souza" "INF-1" "minicurso" "2025-09-05T08:30:00Z" "2025-09-05T12:00:00Z" false 25 false true && ((success_count++))
    ((total_count++))
    
    create_activity "Autenticação de APIs e Controle de Acesso com Keycloak: Introdução ao RBAC" "TEMP DESC" "Brandon Carvalho" "INF-2" "minicurso" "2025-09-05T08:30:00Z" "2025-09-05T12:00:00Z" false 25 false true && ((success_count++))
    ((total_count++))

    create_activity "Fechamento do Hackathon" "TEMP DESC" "Comissão SCTI" "Cine Darcy" "palestra" "2025-09-05T14:00:00Z" "2025-09-05T15:30:00Z" true 0 true false && ((success_count++))
    ((total_count++))
    
    create_activity "Mercado Trabalho Exterior - Programador e Fundador da Operação Código de Ouro" "TEMP DESC" "Lucas Siqueira" "Cine Darcy" "palestra" "2025-09-05T16:00:00Z" "2025-09-05T18:00:00Z" true 0 true false && ((success_count++))
    ((total_count++)) 

    log_success "Created $success_count out of $total_count activities"
    return $([ $success_count -eq $total_count ] && echo 0 || echo 1)
}

# Main execution
main() {
    log_success "Starting SCTI Event Creation Script..."
    log_info "Server: $BASE_URL"
    echo
    
    # Step 1: Load environment variables
    load_env
    
    # Step 2: Login
    log_info "1️⃣ Logging in..."
    if ! login; then
        log_error "Login failed. Exiting."
        exit 1
    fi
    
    # Step 3: Create SCTI Event
    log_info "2️⃣ Creating SCTI Event..."
    if ! create_event; then
        log_error "Event creation failed. Exiting."
        exit 1
    fi
    
    # Step 4: Create Activities
    log_info "3️⃣ Creating activities..."
    if create_all_activities; then
        log_success "All activities created successfully!"
    else
        log_warn "Some activities failed to create."
    fi
    
    echo
    log_success "🎉 SCTI Event setup completed!"
    log_info "Event Slug: scti"
}

# Run the main function
main