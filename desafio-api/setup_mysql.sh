#!/bin/bash

# Configurações do MySQL
DB_USER=${DB_USER:-root}
DB_PASSWORD=${DB_PASSWORD:-}
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}
DB_NAME=${DB_NAME:-mercadolibre_challenge}

# Comando para conectar ao MySQL
MYSQL_CMD="mysql -h $DB_HOST -P $DB_PORT -u $DB_USER"
if [ -n "$DB_PASSWORD" ]; then
    MYSQL_CMD="$MYSQL_CMD -p$DB_PASSWORD"
fi

# Verifica se o MySQL está rodando
echo "Verificando conexão com o MySQL..."
if ! $MYSQL_CMD -e "SELECT 1" >/dev/null 2>&1; then
    echo "Erro: Não foi possível conectar ao MySQL. Verifique se o serviço está rodando e as credenciais estão corretas."
    exit 1
fi

# Cria o banco de dados se não existir
echo "Criando banco de dados $DB_NAME se não existir..."
$MYSQL_CMD -e "CREATE DATABASE IF NOT EXISTS $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# Executa o script de migração
echo "Aplicando migrações..."
$MYSQL_CMD $DB_NAME < migrations/001_create_items_table.mysql.sql

echo "Configuração do banco de dados concluída com sucesso!"
