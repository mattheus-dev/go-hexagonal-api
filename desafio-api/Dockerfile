# Build stage
FROM golang:1.24-alpine AS builder

# Metadados

LABEL maintainer="Mattheus Tibério <mattheustib@gmail.com>"
LABEL description="API RESTful em Go para gerenciamento de itens"
LABEL version="1.0.0"
LABEL org.opencontainers.image.source="https://github.com/mattheus-dev/desafio-api"

WORKDIR /app

# Copiar os arquivos de dependências primeiro para aproveitar o cache do Docker
COPY go.mod go.sum ./
RUN go mod download

# Copiar o código fonte
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o desafio-api ./cmd/api

# Stage final (imagem leve)
FROM alpine:3.18

# Metadados
LABEL maintainer="Mattheus Tibério <mattheustib@gmail.com>"
LABEL description="API RESTful em Go para gerenciamento de itens"
LABEL version="1.0.0"

WORKDIR /app

# Instalar dependências necessárias
RUN apk --no-cache add tzdata ca-certificates && \
    cp /usr/share/zoneinfo/America/Sao_Paulo /etc/localtime && \
    echo "America/Sao_Paulo" > /etc/timezone && \
    apk del tzdata

# Copiar o binário do builder
COPY --from=builder /app/desafio-api .

# Copiar os arquivos de migração
COPY migrations ./migrations

# Copiar o script de configuração do banco de dados
COPY setup_db.sh .

# Tornar o script executável
RUN chmod +x setup_db.sh

# Criar um usuário não-root para executar a aplicação
RUN adduser -D -g '' appuser && \
    chown -R appuser:appuser /app

USER appuser

# Variáveis de ambiente padrão
ENV DB_HOST=db \
    DB_PORT=3306 \
    DB_USER=root \
    DB_PASSWORD=password \
    DB_NAME=desafio_db \
    APP_PORT=8080 \
    GIN_MODE=release \
    JWT_SECRET=your-jwt-secret-key-here

# Expor a porta da aplicação
EXPOSE 8080

# Comando para executar a aplicação
ENTRYPOINT ["/app/desafio-api"]
