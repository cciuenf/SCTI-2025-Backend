# Como Contribuir com o Projeto

Este guia explica como configurar seu ambiente para contribuir com o código do projeto.

## Pré-requisitos

Antes de começar, você precisará:
- Uma conta no GitHub
- Git instalado em sua máquina
- Conhecimentos básicos de linha de comando

## Passo a Passo

### 1. Instale o Nix

O Nix é um gerenciador de pacotes que nos ajuda a criar ambientes de desenvolvimento consistentes.

**Para instalação multiusuário (recomendada)**:
```
sh <(curl -L https://nixos.org/nix/install) --daemon 
```

Para mais detalhes, visite: [https://nixos.org/download/](https://nixos.org/download/)

---

### 2. Faça um Fork do Projeto

1. Visite a página do repositório no GitHub.
2. Clique no botão **Fork** no canto superior direito.
3. Aguarde a criação do fork em sua conta.

---

### 3. Clone o Repositório

Após o fork, clone o repositório para sua máquina local:
```
git clone https://github.com/SEU-USUARIO/SCTI-2025-Backend.git &&
cd SCTI-2025-Backend
```
### 4. Configure o Ambiente de Desenvolvimento

Use o Nix para configurar seu ambiente de desenvolvimento:

```
nix develop 
```

Isso configurará automaticamente todas as dependências necessárias para o desenvolvimento.

---

### 5. Crie uma Branch

Antes de fazer alterações, crie uma branch para seu trabalho:


```
git checkout -b nome-da-sua-feature
```

---

### 6. Faça suas Alterações

Agora você pode fazer alterações no código conforme necessário.

---

### 7. Commit e Push

Após concluir suas alterações:


```
git add .
git commit -m "Descrição clara da sua alteração"
git push origin nome-da-sua-feature
```

---

### 8. Crie um Pull Request

1. Volte ao seu fork no GitHub.
2. Clique no botão **Compare & pull request**.
3. Adicione uma descrição detalhada das suas alterações.
4. Clique em **Create pull request**.

---

## Dicas Adicionais

- Mantenha seu fork sincronizado com o repositório original.
- Teste suas alterações antes de enviar.
- Siga as convenções de código do projeto.
- Documente suas alterações adequadamente.

Obrigado por contribuir com nosso projeto!

---


