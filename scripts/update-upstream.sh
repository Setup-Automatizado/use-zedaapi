#!/bin/bash

# Script para atualizar whatsmeow do upstream de forma segura
# Uso: ./scripts/update-upstream.sh [versão]

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para logging
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Verificar se estamos no diretório correto
if [ ! -f "go.mod" ] || [ ! -d "api" ]; then
    error "Execute este script a partir da raiz do projeto whatsmeow-private"
    exit 1
fi

# Verificar se upstream está configurado
if ! git remote get-url upstream >/dev/null 2>&1; then
    error "Remote 'upstream' não está configurado. Execute:"
    echo "git remote add upstream git@github.com:tulir/whatsmeow.git"
    exit 1
fi

# Criar backup
BACKUP_BRANCH="backup-before-update-$(date +%Y%m%d-%H%M%S)"
log "Criando backup: $BACKUP_BRANCH"
git checkout develop
git branch "$BACKUP_BRANCH"
success "Backup criado: $BACKUP_BRANCH"

# Criar branch para atualização
UPDATE_BRANCH="update-whatsmeow-$(date +%Y%m%d)"
log "Criando branch de atualização: $UPDATE_BRANCH"
git checkout -b "$UPDATE_BRANCH"

# Buscar atualizações
log "Buscando atualizações do upstream..."
git fetch upstream

# Mostrar tags disponíveis
log "Tags disponíveis no upstream:"
git tag -l | grep -E "^v[0-9]+\.[0-9]+\.[0-9]+$" | sort -V | tail -10

# Verificar mudanças
log "Mudanças desde última atualização:"
git log --oneline HEAD..upstream/main | head -20

# Perguntar se deve continuar
echo
warning "Você está prestes a fazer merge das mudanças do upstream."
read -p "Deseja continuar? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log "Operação cancelada pelo usuário"
    git checkout develop
    git branch -D "$UPDATE_BRANCH"
    exit 0
fi

# Fazer merge
log "Fazendo merge das mudanças do upstream..."
if git merge upstream/main; then
    success "Merge realizado com sucesso!"
else
    error "Conflitos encontrados durante o merge!"
    log "Resolva os conflitos e execute:"
    echo "git add ."
    echo "git commit -m 'resolve: merge conflicts from upstream'"
    echo "git checkout develop"
    echo "git merge $UPDATE_BRANCH"
    exit 1
fi

# Atualizar dependências
log "Atualizando dependências..."
go mod tidy
go mod verify

# Executar testes
log "Executando testes..."
if go test ./...; then
    success "Todos os testes passaram!"
else
    error "Alguns testes falharam!"
    log "Verifique os erros antes de continuar"
    exit 1
fi

# Verificar se compila
log "Verificando se o projeto compila..."
if go build -v ./...; then
    success "Projeto compila com sucesso!"
else
    error "Erro de compilação encontrado!"
    exit 1
fi

# Testar API especificamente
log "Testando API..."
cd api
if go test ./... && go build ./cmd/server; then
    success "API testada com sucesso!"
else
    error "Erro nos testes da API!"
    exit 1
fi
cd ..

# Commit das mudanças
log "Fazendo commit das atualizações..."
git add .
git commit -m "chore: update whatsmeow from upstream

- Updated core whatsmeow library
- Resolved dependencies
- All tests passing"

# Voltar para develop e fazer merge
log "Fazendo merge para develop..."
git checkout develop
if git merge "$UPDATE_BRANCH"; then
    success "Atualização concluída com sucesso!"
    log "Branch de atualização: $UPDATE_BRANCH"
    log "Backup disponível em: $BACKUP_BRANCH"
else
    error "Erro ao fazer merge para develop!"
    exit 1
fi

# Limpeza
log "Limpando branch temporário..."
git branch -d "$UPDATE_BRANCH"

echo
success "🎉 Atualização do whatsmeow concluída!"
echo
log "Próximos passos:"
echo "1. Teste sua aplicação em ambiente de desenvolvimento"
echo "2. Execute testes de integração"
echo "3. Faça deploy em ambiente de teste"
echo "4. Se tudo estiver funcionando, faça deploy em produção"
echo
log "Se algo der errado, você pode voltar ao backup:"
echo "git checkout $BACKUP_BRANCH"
echo "git checkout -b develop-from-backup"
