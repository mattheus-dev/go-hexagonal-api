# Desafio API - Gerenciamento de Itens

API RESTful para gerenciamento de itens seguindo os princípios da arquitetura hexagonal (ports and adapters).

## 🚀 Como Usar com Docker

### Pré-requisitos

- Docker 20.10+ e Docker Compose v2.0+
- Git (opcional)

### Usando a imagem do Docker Hub

1. Crie um arquivo `docker-compose.yml`:

```yaml
version: '3.8'

services:
  app:
    image: seuusuario/desafio-api:latest
    container_name: desafio-api
    restart: always
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=db
      - DB_USER=root
      - DB_PASSWORD=password
      - DB_NAME=desafio_db
    depends_on:
      - db

  db:
    image: mysql:8.0
    container_name: desafio-db
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: desafio_db
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

volumes:
  mysql_data:
```

2. Inicie os contêineres:

```bash
docker-compose up -d
```

3. A API estará disponível em: `http://localhost:8080`

### Variáveis de Ambiente

| Variável         | Padrão     | Descrição                           |
|------------------|------------|-----------------------------------|
| `DB_HOST`        | `db`       | Host do banco de dados            |
| `DB_PORT`        | `3306`     | Porta do banco de dados           |
| `DB_USER`        | `root`     | Usuário do banco de dados         |
| `DB_PASSWORD`    | `password` | Senha do banco de dados           |
| `DB_NAME`        | `desafio_db` | Nome do banco de dados           |
| `APP_PORT`       | `8080`     | Porta em que a API irá rodar      |
| `GIN_MODE`       | `release`  | Modo de execução do Gin (debug/release) |

## 🔧 Desenvolvimento

### Pré-requisitos para Desenvolvimento

- Go 1.21 ou superior
- MySQL 8.0 ou superior
- Git

## 📚 Documentação da API

A documentação da API está disponível em [Documentação da API](API_DOCS.md).

## Configuração

1. Clone o repositório:

```bash
git clone https://github.com/mattheus-dev/desafio-api.git
cd desafio-api
```

2. Instale as dependências:

```bash
go mod download
```

3. Configure as variáveis de ambiente:

Crie um arquivo `.env` na raiz do projeto com base no arquivo `.env.example`:

```bash
cp .env.example .env
```

Edite o arquivo `.env` com as configurações do seu banco de dados MySQL.

4. Execute o script de configuração do banco de dados:

```bash
chmod +x setup_mysql.sh
./setup_mysql.sh
```

## Executando a Aplicação

```bash
go run cmd/api/main.go
```

A API estará disponível em `http://localhost:8080`

## Endpoints

### Criar Item

```http
POST /api/v1/items
Content-Type: application/json

{
  "code": "SAM27324354",
  "title": "Tablet Samsung Galaxy Tab S7",
  "description": "Galaxy Tab S7 with S Pen SM-t733 12.4 polegadas e 4GB de memória RAM",
  "price": 150000,
  "stock": 15
}
```

### Listar Itens

```http
GET /api/v1/items?status=ACTIVE&limit=10&page=1
```

### Buscar Item por ID

```http
GET /api/v1/items/1
```

### Atualizar Item

```http
PUT /api/v1/items/1
Content-Type: application/json

{
  "code": "SAM27324354",
  "title": "Tablet Samsung Galaxy Tab S7 (2023)",
  "description": "Galaxy Tab S7 with S Pen SM-t733 12.4 polegadas e 4GB de memória RAM - Nova versão",
  "price": 140000,
  "stock": 10
}
```

### Excluir Item

```http
DELETE /api/v1/items/1
```

## Estrutura do Projeto

```
desafio-api/
├── cmd/
│   └── api/
│       └── main.go           # Ponto de entrada da aplicação
├── internal/
│   ├── adapters/            # Implementações concretas
│   │   ├── database/        # Configuração do banco de dados
│   │   └── repository/      # Implementação do repositório
│   ├── application/         # Casos de uso
│   │   └── service/         # Serviços de aplicação
│   ├── domain/              # Entidades e regras de negócio
│   └── ports/               # Interfaces (portas)
│       └── http/           # Handlers HTTP
├── migrations/              # Migrações do banco de dados
├── pkg/
│   └── utils/             # Utilitários
└── README.md                # Documentação
```

## Testando

Para executar os testes:

```bash
go test -v ./...
```

## Variáveis de Ambiente

| Variável    | Descrição                     | Padrão           |
|-------------|-----------------------------|------------------|
| DB_HOST     | Host do banco de dados       | localhost        |
| DB_PORT     | Porta do banco de dados      | 3306             |
| DB_USER     | Usuário do banco de dados    | root             |
| DB_PASSWORD | Senha do banco de dados      | (vazio)          |
| DB_NAME     | Nome do banco de dados       | mercadolibre_challenge |
| PORT        | Porta da aplicação           | 8080             |

## Licença

Este projeto está licenciado sob a licença MIT - veja o arquivo [LICENSE](LICENSE) para mais detalhes.
