#!/bin/bash

echo "=== Diagnóstico do Sistema de API ==="
echo

# 1. Verificar conectividade com o banco de dados
echo "1. Testando conectividade com o banco de dados..."
docker exec desafio-db mysql -uroot -ppassword -e "SHOW DATABASES;" desafio_db
if [ $? -eq 0 ]; then
    echo "✓ Conectividade com o banco de dados OK"
else
    echo "✗ Erro de conectividade com o banco de dados"
fi

# 2. Verificar estrutura das tabelas
echo -e "\n2. Verificando estrutura da tabela users..."
docker exec desafio-db mysql -uroot -ppassword desafio_db -e "DESCRIBE users;"
echo -e "\n3. Verificando estrutura da tabela items..."
docker exec desafio-db mysql -uroot -ppassword desafio_db -e "DESCRIBE items;"

# 3. Verificar as chaves estrangeiras para os campos de auditoria
echo -e "\n4. Verificando as chaves estrangeiras..."
docker exec desafio-db mysql -uroot -ppassword desafio_db -e "
    SELECT CONSTRAINT_NAME, TABLE_NAME, COLUMN_NAME, REFERENCED_TABLE_NAME
    FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
    WHERE TABLE_NAME = 'items' AND REFERENCED_TABLE_NAME IS NOT NULL;
"

# 4. Inserir um usuário diretamente no banco para teste
echo -e "\n5. Inserindo usuário de teste diretamente no banco..."
docker exec desafio-db mysql -uroot -ppassword desafio_db -e "
    INSERT INTO users (username, password, created_at, updated_at)
    VALUES ('direct_test_user', '\$2a\$10\$JqSMvjWGNxnf2MbO83fYc.lWvfn4h8Fwn.UQGtN9XaAQyE6H/uINW', NOW(), NOW())
    ON DUPLICATE KEY UPDATE username=username;
"
echo "Usuário inserido"

# 5. Verificar se o usuário foi inserido
echo -e "\n6. Verificando se o usuário foi inserido..."
docker exec desafio-db mysql -uroot -ppassword desafio_db -e "SELECT * FROM users WHERE username = 'direct_test_user';"

# 6. Testar registro via API com curl detalhado
echo -e "\n7. Testando registro via API..."
curl -v -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "api_test_user", "password": "password123"}'
echo -e "\n"

# 7. Testar login com o usuário inserido diretamente
echo -e "\n8. Testando login com usuário inserido diretamente..."
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "direct_test_user", "password": "password123"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "✗ Login falhou, nenhum token recebido"
else
    echo "✓ Login bem-sucedido!"
    echo "Token JWT recebido: ${TOKEN:0:20}..."

    # 8. Testar criação de item com o token
    echo -e "\n9. Testando criação de item com autenticação..."
    ITEM_RESP=$(curl -s -X POST http://localhost:8080/api/v1/items \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "code": "TEST_AUDIT",
        "title": "Item de Teste Diagnóstico",
        "description": "Item para teste de campos de auditoria",
        "price": 1999,
        "stock": 50
      }')
    
    echo "Resposta: $ITEM_RESP"
    
    # 9. Verificar se os campos de auditoria foram salvos
    echo -e "\n10. Verificando campos de auditoria no banco..."
    docker exec desafio-db mysql -uroot -ppassword desafio_db -e "
        SELECT i.id, i.code, i.created_by, u.username as created_by_user
        FROM items i
        JOIN users u ON i.created_by = u.id
        WHERE i.code = 'TEST_AUDIT';
    "
fi

echo -e "\n=== Diagnóstico concluído ==="
