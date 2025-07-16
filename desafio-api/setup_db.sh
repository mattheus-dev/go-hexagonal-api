#!/bin/bash

# Configurações do banco de dados
MYSQL_DATABASE="mercadolibre_challenge"
MYSQL_USER="root"
MYSQL_PASSWORD="root"
MYSQL_ROOT_PASSWORD="root"
DB_HOST="localhost"

# Aguardar o MySQL estar pronto para aceitar conexões
until mysqladmin ping -h "$DB_HOST" --silent; do
    echo "Aguardando o MySQL estar disponível..."
    sleep 2
done

# Criar o banco de dados se não existir
mysql -h "$DB_HOST" -u root -p"$MYSQL_ROOT_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS $MYSQL_DATABASE;"

# Criar usuário e conceder privilégios
mysql -h "$DB_HOST" -u root -p"$MYSQL_ROOT_PASSWORD" -e "
CREATE USER IF NOT EXISTS '$MYSQL_USER'@'%' IDENTIFIED BY '$MYSQL_PASSWORD';
GRANT ALL PRIVILEGES ON $MYSQL_DATABASE.* TO '$MYSQL_USER'@'%';
FLUSH PRIVILEGES;
"

# Executar migrações
for migration in migrations/*.sql; do
    echo "Aplicando migração: $migration"
    mysql -h "$DB_HOST" -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < "$migration"
done

echo "Banco de dados configurado com sucesso!"
