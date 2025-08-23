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
        "start_date": "2025-09-01T00:00:00-03:00",
        "end_date": "2025-09-05T23:59:59-03:00",
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
    local level="${12}"
    local requirements="${13}"
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
        --arg level "$level" \
        --arg requirements "$requirements" \
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
            level: $level,
            requirements: $requirements,
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
    create_activity "Abertura da Semana Acadêmica" "Cerimônia de abertura com autoridades e apresentação da semana" "PROGRAD" "Auditório Principal" "palestra" "2025-09-01T09:00:00-03:00" "2025-09-01T12:00:00-03:00" true 0 true false "none" "" && ((success_count++))
    ((total_count++))

    create_activity "Mesa redonda sobre estágios" "Conectar estudantes de Ciência da Computação com o mercado de trabalho através de experiências reais de estágio." "Estagiários de Computação" "Cine Darcy" "palestra" "2025-09-01T14:00:00-03:00" "2025-09-01T15:30:00-03:00" true 0 true false "none" "" && ((success_count++))
    ((total_count++))

    create_activity "Hackathon" "SEM DESCRIÇÂO" "Comissão SCTI" "Cine Darcy" "palestra" "2025-09-01T16:00:00-03:00" "2025-09-01T18:00:00-03:00" true 0 true false "medium" "" && ((success_count++))
    ((total_count++))

    # Day 2 - September 2, 2025
    create_activity "Curso Prático de Pentest usando Kali Linux" "Em um mundo cada vez mais digital, a segurança da informação tornou-se uma prioridade essencial para empresas, governos e indivíduos. A crescente sofisticação dos ataques cibernéticos exige profissionais cada vez mais capacitados para identificar, mitigar e prevenir vulnerabilidades em sistemas e redes. Nesse contexto, o objetivo deste minicurso é proporcionar aos participantes uma introdução prática em Pentest, capacitando-os a compreender o funcionamento de ataques e a aplicar procedimentos de defesa, contribuindo diretamente para a construção de ambientes digitais mais seguros." "Prof. Vinicius Barcelos" "INF-1" "minicurso" "2025-09-02T08:30:00-03:00" "2025-09-02T12:00:00-03:00" false 25 false true "medium" "Virtual box com Kali" && ((success_count++))
    ((total_count++))

    create_activity "MovieTracker: Criando um app com React Native" "Nesse minicurso vamos aprender a fazer um aplicativo em React Native com Expo! Ao final da atividade cada participante terá a sua própria versão do Movie Tracker, um aplicativo que consome uma API pública de filmes e permite interações com os títulos. " "Renan Souza Oliveira" "INF-2" "minicurso" "2025-09-02T08:30:00-03:00" "2025-09-02T12:00:00-03:00" false 25 false true "medium" "NodeJS, NPM/Yarn, Escolha um:, Android Studio e JDK24, Expo GO e um celular" && ((success_count++))
    ((total_count++))

    create_activity "Desburocratizando o mercado de trabalho" "Nessa palestra veremos de forma detalhada as diferentes maneiras de ingressar no mercado de trabalho, porém, numa perspectiva mais ampla e geral, não se restringindo somente aos profissionais de computação. O intuito da atividade é que todos os participantes possam distinguir as modalidades de contratação e identificar qual é a que mais se enquadra no seu atual momento." "Patrick Pereira" "Cine Darcy" "palestra" "2025-09-02T14:00:00-03:00" "2025-09-02T15:30:00-03:00" true 0 true false  "none" "" && ((success_count++))
    ((total_count++))

    create_activity "Como avaliar um modelo de machine learning e seus desafios." "Desafios e métodos para avaliação de uma modelo de previsão de crédito." "Clébio Júnior" "Cine Darcy" "palestra" "2025-09-02T16:00:00-03:00" "2025-09-02T18:00:00-03:00" true 0 true false "easy" ""  && ((success_count++))
    ((total_count++))

    # Day 3 - September 3, 2025
    create_activity "Engenharia e Ciência de Dados com Big Data: Prática com PySpark no Databricks" "Este minicurso prático introduz Engenharia e Ciência de Dados com big data, usando PySpark no Databricks Community Edition. Os alunos criarão pipelines ETL e análises em datasets massivos no laboratório, simulando fluxos do Azure Databricks. Inclui visão do mercado, destacando habilidades e oportunidades com PySpark." "João Paulo Seixas" "INF-1" "minicurso" "2025-09-03T08:30:00-03:00" "2025-09-03T12:00:00-03:00" false 25 false true "easy" "Python3" && ((success_count++))
    ((total_count++))   

    create_activity "Dia a dia de desenvolvimento web: Entendo como funciona na pratica" "Partiremos de um projeto já pronto, um mini-sistema corporativo de Agendamento de Salas de Reunião (ASP.NET Core Web API + JavaScript puro), simulando um produto real. A proposta é mostrar o dia a dia de trabalho pegando duas tarefas: uma feature (busca e paginação no CRUD de salas/reservas, com regras simples) e um bug (depuração e resolução do problema). Durante a execução, apresentarei o fluxo com issues, branch, commits e PRs, além de práticas de validação e tratamento de erros — objetivo: um exemplo conciso, prático e replicável." "Jhulian Pereira Manhães" "INF-2" "minicurso" "2025-09-03T08:30:00-03:00" "2025-09-03T12:00:00-03:00" false 25 false true "medium" "Git, Visual Studio/VSCode" && ((success_count++))
    ((total_count++))

    create_activity "Inteligência Artificial no Trabalho: Aplicações em Ambientes Cloud" "Esta palestra explora o impacto da IA no mercado, apresentando exemplos práticos de desenvolvimento de algoritmos e ciência de dados em ambientes cloud. Casos reais, como chatbots e análises preditivas, são destacados, junto a fundamentos de planejamento e criação de IA escalável. Ideal para quem quer entender como a IA transforma negócios em larga escala." "João Paulo Seixas" "Cine Darcy" "palestra" "2025-09-03T14:00:00-03:00" "2025-09-03T15:30:00-03:00" true 0 true false "easy" "" && ((success_count++))
    ((total_count++))

    create_activity "Matemática Aplicada no Ofício" "Nessa palestra entenderemos como toda a base matemática do curso se aplica de fato no dia a dia de um cientista da computação, com exemplos palpáveis de como essa fundamentação está presente nas mais diversas áreas do universo tecnológico." "Prof. João Luiz" "Cine Darcy" "palestra" "2025-09-03T16:00:00-03:00" "2025-09-03T18:00:00-03:00" true 0 true false "medium" "" && ((success_count++))
    ((total_count++))

    # Day 4 - September 4, 2025
    create_activity "Introdução à UX com IA" "Aprenda sobre design da experiência de usuário e crie um produto digital pro seu portfólio com ajuda de IA." "Diana de Sales" "INF-2" "minicurso" "2025-09-04T08:30:00-03:00" "2025-09-04T12:00:00-03:00" false 25 false true "easy" "Conta no Figma" && ((success_count++))
    ((total_count++))

    create_activity "Manutenção de Hardware" "TEMP DESC" "SEM PALESTRANTE" "INF-1" "minicurso" "2025-09-04T08:30:00-03:00" "2025-09-04T12:00:00-03:00" false 25 false true "easy" "" && ((success_count++))
    ((total_count++))

    create_activity "Erlang, MCP e Kubernetes: Lições de um Sistema Distribuído em Produção" "Uma jornada pelos desafios reais de construir sistemas distribuídos a fim de criar um assistente financeiro inteligente com LLM. Explorando como a BEAM/Erlang se integra com protocolo de modelo contextual (MCP) para criar clusters escaláveis, enfrentando questões de consistência eventual via delta CRDTs e a complexa integração entre Kubernetes e a máquina virtual do Erlang (BEAM). Um mergulho técnico com reviravoltas sobre como teoria encontra prática em sistemas de produção" "Zoey de Souza" "Cine Darcy" "palestra" "2025-09-04T14:00:00-03:00" "2025-09-04T15:30:00-03:00" true 0 true false "hard" "" && ((success_count++))
    ((total_count++))

    create_activity "Guia de Sobrevivência no Mercado Tech" "O mercado de tecnologia mudou e não, não é apenas culpa da IA. Se antes parecia fácil entrar e crescer, hoje a concorrência está mais acirrada e a realidade das empresas é bem diferente do que se vende em redes sociais e cursos milagrosos. Nesta palestra, compartilho quase 20 anos de experiência para alinhar expectativas de quem está começando ou tentando se recolocar." "Mano Deyvin" "Cine Darcy" "palestra" "2025-09-04T16:00:00-03:00" "2025-09-04T18:00:00-03:00" true 0 true false "none" "" && ((success_count++))
    ((total_count++))

    # Day 5 - September 5, 2025
    create_activity "DevOps Desmistificado: Construindo sua Pipeline do Zero" "Minicurso que guia os participantes desde os fundamentos do terminal até a construção de pipelines de integração contínua (CI) e entrega contínua (CD). Através de exemplos e exercícios, exploraremos Docker, automação de deploy e boas práticas de DevOps, revelando como conceitos aparentemente simples se conectam para formar sistemas complexos de automação" "Zoey de Souza" "INF-1" "minicurso" "2025-09-05T08:30:00-03:00" "2025-09-05T12:00:00-03:00" false 25 false true "medium" "Linux/WSL, Docker, VSCode" && ((success_count++))
    ((total_count++))

    create_activity "Autenticação de APIs e Controle de Acesso com Keycloak: Introdução ao RBAC" "Neste minicurso, vamos explorar os fundamentos da autenticação e autorização em APIs modernas, com foco em boas práticas de segurança e gerenciamento de acesso baseado em funções (RBAC). Você aprenderá o que são tokens, roles, escopos, a diferença entre autenticação e autorização, e como integrar o Keycloak — uma poderosa ferramenta de identidade e acesso — com aplicações backend. Ao final, cada participante terá implementado uma API protegida com Keycloak, utilizando tokens JWT e controle de acesso por funções." "Brandon Carvalho" "INF-2" "minicurso" "2025-09-05T08:30:00-03:00" "2025-09-05T12:00:00-03:00" false 25 false true "Hard" "Docker, Docker-compose, VSCode, Postman" && ((success_count++))
    ((total_count++))

    create_activity "Fechamento do Hackathon" "TEMP DESC" "Comissão SCTI" "Cine Darcy" "palestra" "2025-09-05T14:00:00-03:00" "2025-09-05T15:30:00-03:00" true 0 true false "medium" "" && ((success_count++))
    ((total_count++))

    create_activity "Mercado Trabalho Exterior - Programador e Fundador da Operação Código de Ouro" "Programador com 6 anos de carreira, sendo 4 deles trabalhando para empresas do Mercado Internacional, hoje Lucas atua como Software Engineer na Medely, é Especialista em Internacionalização de carreira, alocando Devs do Brasil inteiro em empresas internacionais e BigTechs como Meta, SAP, Santander, Itaú, Ifood, etc. 26 anos, se formou no Instituto Federal Fluminense em Campos - Rio de Janeiro e hoje ajuda múltiplas centenas de Devs a conseguirem a vaga internacional." "Lucas Siqueira" "Cine Darcy" "palestra" "2025-09-05T16:00:00-03:00" "2025-09-05T18:00:00-03:00" true 0 true false "none" "" && ((success_count++))
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
