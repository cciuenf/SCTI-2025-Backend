# SCTI 2025 BACKEND

Contém a API de backend para o evento da SCTI 2025

## Requisitos

### Pré-Projeto
- [ ] Fazer setup inicial (Docker com Postgres) + go
- [x] Criar .env.example
- [x] Criar a API base de maneira extensível
- [ ] Github Actions para PRs e Issues no Trello
- [x] Github Actions para Linting e Formatting
- [ ] Criar Issues para os requisitos e rotas da API


### Autenticação -> Go-Auth
- [ ] Pacote para autenticação modularizado
- [x] JWT + Refresh token por padrão
- [ ] Possui uma interface para conectar seu próprio DB
- [ ] Possui uma interface para interagir com seu usuário
- [x] Proporciona o middleware de  autenticação
- [ ] Proporciona uma função para trocar senha
- [ ] Proporciona verificação de conta
- [ ] Possui rate limiting pra evitar ataques de força bruta


### Banco de Dados
- [ ] Modelagem do banco de dados
- [ ] Usar transações em código crítico
- [ ] Modelagem do banco no próprio códgio usando GORM


#### User
- [ ] UUID pk
- [ ] IsVerified
- [ ] IsPaid
- [ ] Tokens
- [ ] Nome
- [ ] Sobrenome
- [ ] Email
- [ ] PFP (Path)
- [ ] IsUenf
- [ ] Curso
- [ ] Periodo
- [ ] Redes
- [ ] IsAdmin
- [ ] IsMaster


- Usuários que possuam foto de perfil
    - Verificaçao de imagem com (Google cloud/Rekognition) se possível
    - A imagem ao chegar ao backend já deve ter sido verificada pelo NSFW JS
    - Fotos de perfis são guardadas em disco, < 10mb e apenas .png .jpg .webp

- Email de usuario deve ser limpado antes de processado (toLower, ...)
- Não deve exisitir mais de uma conta com o mesmo email
- Para acessar a dashboard o email deve ser verificado
- Emails que devem ser enviados ao usuário
    - Email confirmando compra
    - Email para lembrar o usuário de se inscrever nas atividades
    - Email de QR-Code (Caso alguma comissão não tenha feito os crachas fisicos)
    - Email de certificados
    - Email de notícias

- Cracha
    - Layout de crachá deve ser padronizado
        - Layout com PFP
        - Layout sem PFP
    - Sistema para gerar crachás automáticos
    - Foto de usuário no cracha é opcional
    - Detecção de tamanho de nome para quebra de linha

- Presença em Atividades
    - O sistema de presença deve ser inteligente
        - Validar a presença mais próxima temporalmente ao scanear o QR code com 20 minutos de segurança

- Os links que o usuário colocar em redes devem ser verificados
    - Não podem ser links minificados
    - Usar um regex para proibir links especificos


#### Activity
- [ ] ID pk
- [ ] Tipo (MC|Minicurso, PL|Palestra, EV|Evento[Extra])
- [ ] Nome
- [ ] Descrição
- [ ] Horário
- [ ] Ministrante
- [ ] Sala
- [ ] DiaEvento
- [ ] Vagas
- [ ] Thumbnail (Path)

- Desinscrição
    - Proibido para quem compareceu
    - Proibido se a atividade já passou
    - Palestras devem ser impossíveis de se desinscrever
- Inscrições
    - Proibido se você tem uma atividade no mesmo horário
    - Proibido se já passou
    - Proibido se o usuário não possui mais créditos
    - Todos os usuários devem ser inscritos em todas as palestras quando pagarem o ingresso
- Vagas
    - Se a atividade for Palestra/Evento ela ignora a variavel vagas
- Thumbnail
    - Atividades devem possuir thumbnail (Ou foto do ministrante ou thumb customizada)


#### Registrations
- [ ] UserID fk
- [ ] ActivityID fk
- [ ] Attended

- Race condition
    - A implementação atual possui uma condição de corrida, achar uma implementação que diminua ou remova essa condição de corrida


#### ActionLog
- [ ] ActionID AutoIncrement
- [ ] PerformedBy (User, Admin, Master)
- [ ] UserID
- [ ] Time
- [ ] ActionType
- [ ] ActionData


### Segurança
- [ ] Redirect Automático de http para https pela linguagem
- [ ] Certificado Let's encrypt automático


### Logging Middleware
- [ ] Log padronizado (Rota/Modulo/Arquivo)
- [ ] Arquivo de Log Global
- [ ] Arquivo de log Localizado
- [ ] Arquivo de log por usuário
- [ ] Contador de Visitas / Server metrics


--!! NAO REVISADOS !!--
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

