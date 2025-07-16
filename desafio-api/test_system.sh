#!/bin/bash

# Script para testar a autenticação JWT e campos de auditoria
echo "=== Testando autenticação JWT e campos de auditoria ==="
echo ""

# URL base da API
BASE_URL="http://localhost:8080"

# Função para logar erro e sair
log_error() {
    echo "ERRO: $1"
    exit 1
}

# 1. Registrar um novo usuário
echo "1. Registrando um novo usuário de teste..."
REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}')

STATUS_CODE=$(echo "$REGISTER_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$REGISTER_RESPONSE" | sed '$d')

if [ "$STATUS_CODE" -ne 201 ]; then
    echo "Resposta: $RESPONSE_BODY"
    if [ "$STATUS_CODE" -eq 400 ] && [[ "$RESPONSE_BODY" == *"Usuário já existe"* ]]; then
        echo "Usuário testuser já existe, continuando com login..."
    else
        log_error "Falha ao registrar usuário. Status: $STATUS_CODE"
    fi
else
    echo "✅ Usuário testuser registrado com sucesso!"
fi

echo ""

# 2. Fazer login para obter o token JWT
echo "2. Fazendo login para obter token JWT..."
LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}')

STATUS_CODE=$(echo "$LOGIN_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$LOGIN_RESPONSE" | sed '$d')

if [ "$STATUS_CODE" -ne 200 ]; then
    echo "Resposta: $RESPONSE_BODY"
    log_error "Falha ao fazer login. Status: $STATUS_CODE"
fi

# Extrair o token do JSON de resposta
TOKEN=$(echo "$RESPONSE_BODY" | grep -o '"token":"[^"]*' | sed 's/"token":"//')

if [ -z "$TOKEN" ]; then
    log_error "Não foi possível extrair o token da resposta"
fi

echo "✅ Login bem-sucedido! Token JWT obtido."
echo ""

# 3. Criar um item usando o token JWT (com campos de auditoria)
echo "3. Criando um item com token JWT (campos de auditoria)..."
CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/items" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "TEST001",
    "title": "Item de Teste",
    "description": "Este item foi criado para testar campos de auditoria",
    "price": 1500,
    "stock": 10
  }')

STATUS_CODE=$(echo "$CREATE_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$CREATE_RESPONSE" | sed '$d')

if [ "$STATUS_CODE" -ne 201 ]; then
    echo "Resposta: $RESPONSE_BODY"
    
    if [ "$STATUS_CODE" -eq 409 ]; then
        echo "Item com código TEST001 já existe. Vamos continuar com a atualização."
    else
        log_error "Falha ao criar item. Status: $STATUS_CODE"
    fi
else
    echo "✅ Item criado com sucesso!"
    
    # Extrair o ID do item do JSON de resposta
    ITEM_ID=$(echo "$RESPONSE_BODY" | grep -o '"id":[0-9]*' | sed 's/"id"://')
    
    if [ -z "$ITEM_ID" ]; then
        log_error "Não foi possível extrair o ID do item da resposta"
    fi
    
    echo "ID do item: $ITEM_ID"
    
    # Verificar os campos de auditoria na resposta
    echo "Verificando campos de auditoria na resposta..."
    CREATED_BY=$(echo "$RESPONSE_BODY" | grep -o '"created_by":[0-9]*' | sed 's/"created_by"://')
    UPDATED_BY=$(echo "$RESPONSE_BODY" | grep -o '"updated_by":[0-9]*' | sed 's/"updated_by"://')
    
    echo "Campo created_by: $CREATED_BY"
    echo "Campo updated_by: $UPDATED_BY"
    
    if [ -z "$CREATED_BY" ] || [ -z "$UPDATED_BY" ]; then
        log_error "Campos de auditoria não encontrados na resposta"
    fi
    
    if [ "$CREATED_BY" != "$UPDATED_BY" ]; then
        log_error "Os campos created_by e updated_by deveriam ser iguais na criação"
    fi
    
    echo "✅ Campos de auditoria verificados com sucesso!"
else
    # Se o item já existe, vamos buscar seu ID
    echo "Buscando o ID do item existente..."
    LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/items" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json")
    
    ITEM_ID=$(echo "$LIST_RESPONSE" | grep -o '"id":[0-9]*.*"code":"TEST001"' | grep -o '"id":[0-9]*' | head -1 | sed 's/"id"://')
    
    if [ -z "$ITEM_ID" ]; then
        log_error "Não foi possível encontrar o ID do item TEST001"
    fi
    
    echo "ID do item existente: $ITEM_ID"
fi

echo ""

# 4. Atualizar o item para verificar o campo updated_by
echo "4. Atualizando o item para verificar o campo updated_by..."
UPDATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/api/v1/items/$ITEM_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "TEST001",
    "title": "Item de Teste Atualizado",
    "description": "Este item foi atualizado para testar o campo updated_by",
    "price": 2000,
    "stock": 15
  }')

STATUS_CODE=$(echo "$UPDATE_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$UPDATE_RESPONSE" | sed '$d')

if [ "$STATUS_CODE" -ne 200 ]; then
    echo "Resposta: $RESPONSE_BODY"
    log_error "Falha ao atualizar item. Status: $STATUS_CODE"
fi

echo "✅ Item atualizado com sucesso!"

# Verificar os campos de auditoria na resposta de atualização
echo "Verificando campos de auditoria após atualização..."
CREATED_BY=$(echo "$RESPONSE_BODY" | grep -o '"created_by":[0-9]*' | sed 's/"created_by"://')
UPDATED_BY=$(echo "$RESPONSE_BODY" | grep -o '"updated_by":[0-9]*' | sed 's/"updated_by"://')
CREATED_AT=$(echo "$RESPONSE_BODY" | grep -o '"created_at":"[^"]*' | sed 's/"created_at":"//')
UPDATED_AT=$(echo "$RESPONSE_BODY" | grep -o '"updated_at":"[^"]*' | sed 's/"updated_at":"//')

echo "Campo created_by: $CREATED_BY"
echo "Campo updated_by: $UPDATED_BY"
echo "Campo created_at: $CREATED_AT"
echo "Campo updated_at: $UPDATED_AT"

if [ -z "$CREATED_BY" ] || [ -z "$UPDATED_BY" ] || [ -z "$CREATED_AT" ] || [ -z "$UPDATED_AT" ]; then
    log_error "Um ou mais campos de auditoria não encontrados na resposta"
fi

if [ "$CREATED_AT" = "$UPDATED_AT" ]; then
    log_error "Os campos created_at e updated_at não deveriam ser iguais após uma atualização"
fi

echo "✅ Campos de auditoria verificados com sucesso após atualização!"
echo ""

echo "=== Teste concluído com sucesso! ==="
echo "A autenticação JWT e os campos de auditoria estão funcionando corretamente."
