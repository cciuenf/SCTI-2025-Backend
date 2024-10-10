# SCTI 2025 BACKEND

Contém a API de backend para o evento da SCTI 2025

## Requisitos

Nem todos os requisitos nesta lista irão virar Issues, está lista só serve como uma maneira de acompanhar uma visão menos especializada do projeto

### Pré-Projeto
    - [ ] Fazer setup do projeto spring boot
    - [ ] Fazer setup do postgres
    - [ ] Colocar ambos para rodar em um container Docker
    - [x] Criar a API base de maneira extensível
    - [ ] Github Actions para PRs e Issues no Trello 
    - [ ] Github Actions para Linting e Formatting 
    - [ ] Criar Issues para todos os requisitos e rotas da API

### Banco de Dados
    - [ ] Modelagem do banco de dados
    - [ ] Não guardar imagens no banco de dados

### Usuários/Administradores
    - [ ] Identificador (Periodo) [Caso aplicável]
    - [ ] Identificador (IsCommission)
        - Não pode estar cadastrado em nenhuma atividade
        - Tem permissões para validar presença
    - [ ] Identificador (IsAdmin)
        - Possuem permissoes administrativas reduzidas
        - Não podem promover outros a administrador
        - Não podem rodar queries
        - Não podem apagar usuários
    - [ ] Identificador (IsMasterAdmin)
        - Possuem permissoes administrativas completas menos uma
    - [ ] Identificador (IsMasterUser)
        - Possue controle total do sistema
        - Deve existir apenas um
    - [ ] Usuários devem possuir foto de perfil
        - Verificaçao de imagem com (Google cloud/Rekognition)
        - A imagem ao chegar ao backend já deve ter sido verificada pelo NSFW JS
    - [ ] Presença em Atividades
        - O sistema de presença deve ser inteligente
            - [ ] Validar a presença mais próxima temporalmente ao scanear o QR code com 20 minutos de segurança 
        - [ ] Usuários não pagos não devem receber email com QR code

### Atividades
    - [ ] Três tipos (MC[Minicurso], PL[Palestra], EV[Evento])
    - [ ] Desinscrição
        - Proibido para quem compareceu
        - Proibido se a atividade já passou
        - Palestras devem ser impossíveis de se desinscrever
    - [ ] Inscrições
        - Proibido se você tem uma atividade no mesmo horário
        - Proibido se já passou
        - Proibido se o usuário não possui mais créditos
        - Todos os usuários devem ser inscritos em todas as palestras quando pagarem o ingresso

### Email
    - [ ] Todos os email devem ser lowercase
    - [ ] Não permitir mais de uma conta com o mesmo email
    - [ ] Para finalizar a criação de conta o email deve ser verificado
    - [ ] 

    E-mail confirmando pagamento
    E-mail de lembrança de inscrição
    E-mail de QR-code
    E-mail de certificados
    E-mail de notícias (convite para eventos, inscrições abertas, etc)

### Segurança
    - [ ] Redirect Automático de http para https pela linguagem
    - [ ] Certificado Let's encrypt automático
    - [ ] Autenticação através de JWTs
    - [ ] JWTs com diferentes níveis de acesso
    - [ ] JWTs revogaveis
    - [ ] Double Token Auth
        - Token de acesso (usuários)
        - Token de validação (banco de dados)

### Crachá
    - [ ] Layout de crachá padronizado
    - [ ] Sistema para gerar crachás automáticos
    - [ ] Foto do usuário no crachá caso o usuário tenha foto, senão sem foto
    - [ ] Detecção de tamanho de nome para quebra de linha

### Middleware
    - [ ] Log padronizado (Rota/Modulo/Arquivo)
    - [ ] Arquivo de Log Global
    - [ ] Arquivo de log Localizado
    - [ ] Arquivo de log por usuário
    - [ ] Contador de Visitas / Server metrics
    - [ ] Tracker de visitante (Localização generalizada {país} para analise futura)

### Produtos
    - [ ] Implementar produtos na API

### Dev Utilities
    - [ ] Github Actions para deploy
    - [ ] Servidor com Nix

### Extras
    - [ ] Query Runner (Master) 
    - [ ] Server command prompt (Master)
    - [ ] Sistema de pontos para brindes
        - Baseado em presença e colocação em atividades que todos tem acesso
    - [ ] CTF durante evento
    - [ ] Load test
    - [ ] Unit Tests Obrigatórios
    - [ ] Documentar durante o processo e nao depois
    - [ ] Não guardar imagens no banco de dados
