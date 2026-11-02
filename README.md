# Full Cycle - Pós Go Expert - Clean Architecture

Projeto desenvolvido em Go com foco em **Clean Architecture e listagem de orders**, expondo um único caso de uso através de REST, gRPC e GraphQL, disponíveis simultaneamente.

A aplicação parte do [exemplo de Clean Architecture do curso](https://github.com/devfullcycle/goexpert/tree/main/20-CleanArch) e acrescenta a funcionalidade de listagem e a execução automatizada com Docker Compose.

O objetivo deste projeto é demonstrar:

* Separação entre entidades, casos de uso e infraestrutura
* Utilização de interfaces para desacoplar o acesso ao banco
* Compartilhamento do `ListOrdersUseCase` entre três interfaces de comunicação
* Criação e listagem de orders
* Injeção de dependências utilizando Wire
* Configuração utilizando Viper e variáveis de ambiente
* Provisionamento do MySQL e RabbitMQ com Docker
* Aplicação automática de migrações
* Controle da ordem de inicialização dos serviços

Este repositório contém somente o código deste desafio.

---

# Tecnologias utilizadas

* Go 1.26
* Clean Architecture
* REST e HTTP Server
* Chi Router
* gRPC e Protocol Buffers
* GraphQL e gqlgen
* MySQL 8.4
* RabbitMQ 4.4.0
* Wire
* Viper
* golang-migrate
* Docker e Docker Compose

---

# Como executar

## 1. Clonar o repositório

```bash
git clone https://github.com/danielsantello/fc-pos-desafio-clean-architecture.git
cd fc-pos-desafio-clean-architecture
```

## 2. Iniciar a aplicação

Com Docker e Docker Compose instalados, execute na raiz do projeto:

```bash
docker compose up
```

Esse é o único comando necessário para construir e iniciar o ambiente. Não é necessário instalar Go, executar geradores, criar tabelas ou preparar um arquivo `.env`.

Na primeira execução, o Compose baixa as imagens e compila a aplicação. Aguarde as mensagens de inicialização dos três servidores antes de enviar requisições.

As portas `8000`, `50051`, `8080` e `15673` precisam estar disponíveis na máquina.

## 3. Portas e serviços

| Serviço | Endereço na máquina | Finalidade |
| --- | --- | --- |
| REST | `http://localhost:8000/order` | Criar e listar orders |
| gRPC | `localhost:50051` | RPCs `CreateOrder` e `ListOrders` |
| GraphQL | `http://localhost:8080/query` | Mutation de criação e query de listagem |
| GraphQL Playground | `http://localhost:8080/` | Interface para executar operações GraphQL |
| RabbitMQ Management | `http://localhost:15673` | Painel de administração |

O painel do RabbitMQ utiliza usuário `labs` e senha `labs`.

O MySQL e a conexão AMQP do RabbitMQ ficam disponíveis na rede interna do Compose, respectivamente em `mysql:3306` e `rabbitmq:5672`. Essas duas portas não são publicadas na máquina.

---

# Inicialização automática e migrações

O Compose coordena quatro serviços:

1. **mysql:** inicia o banco e verifica sua disponibilidade com uma consulta SQL autenticada.
2. **migrate:** aguarda o MySQL saudável e aplica as migrações pendentes.
3. **rabbitmq:** inicia o broker e verifica sua disponibilidade.
4. **app:** inicia após o banco e o RabbitMQ estarem saudáveis e a migração terminar com sucesso.

Essa sequência utiliza `healthcheck` e as condições `service_healthy` e `service_completed_successfully` do `depends_on`, evitando que a aplicação inicie antes de suas dependências estarem prontas.

As migrações estão na pasta `migrations`:

| Arquivo | Operação |
| --- | --- |
| `000001_create_orders.up.sql` | Cria a tabela `orders` |
| `000001_create_orders.down.sql` | Remove a tabela `orders` em uma reversão explícita |

A tabela contém `id`, `price`, `tax` e `final_price`. O ID é a chave primária.

O golang-migrate registra a versão aplicada no banco. Nas próximas inicializações, a mensagem `no change`, seguida de saída com código `0`, indica que não existem migrações pendentes.

O Compose executa somente `up`. A migração `down` não é executada ao desligar os containers.

Os dados do MySQL e do RabbitMQ são armazenados em volumes nomeados e sobrevivem à recriação dos containers.

---

# Funcionamento da aplicação

Uma order possui os seguintes dados:

```json
{
  "id": "order-1",
  "price": 100.5,
  "tax": 0.5,
  "final_price": 101
}
```

O preço final é calculado somando `price` e `tax`.

A criação já fazia parte do exemplo do curso. O desafio acrescenta o `ListOrdersUseCase`, que consulta o contrato do repositório e transforma as entidades em DTOs de saída.

REST, gRPC e GraphQL recebem a mesma instância desse caso de uso na inicialização da aplicação. Cada adaptador converte o resultado para seu formato de resposta.

Todos utilizam o mesmo banco: uma order criada por qualquer uma das interfaces pode ser consultada pelas outras.

A listagem retorna todas as orders, ordenadas por `id`, sem paginação. Quando não existem registros, retorna uma lista vazia.

---

# REST

## Criar uma order

```text
POST http://localhost:8000/order
```

```bash
curl -i http://localhost:8000/order \
  -H 'Content-Type: application/json' \
  -d '{"id":"order-1","price":100.5,"tax":0.5}'
```

Resposta esperada: `HTTP 200 OK`, com conteúdo JSON.

```json
{
  "id": "order-1",
  "price": 100.5,
  "tax": 0.5,
  "final_price": 101
}
```

Utilize um ID diferente para cada nova order. O banco não permite IDs duplicados.

## Listar orders

```text
GET http://localhost:8000/order
```

```bash
curl -i http://localhost:8000/order
```

Resposta esperada: `HTTP 200 OK`.

```json
[
  {
    "id": "order-1",
    "price": 100.5,
    "tax": 0.5,
    "final_price": 101
  }
]
```

Sem registros, a resposta é `[]`.

O arquivo [api.http](api.http), na raiz, contém as requisições de criação e listagem, além de uma consulta GraphQL. Ele pode ser utilizado em clientes compatíveis, como a extensão REST Client do VS Code.

---

# gRPC

O contrato está em [order.proto](internal/infra/grpc/protofiles/order.proto).

O serviço `pb.OrderService` disponibiliza:

| RPC | Requisição | Resposta |
| --- | --- | --- |
| `CreateOrder` | `CreateOrderRequest` | `CreateOrderResponse` |
| `ListOrders` | `ListOrdersRequest` | `ListOrdersResponse` |

`ListOrdersRequest` é vazio. A resposta contém o campo repetido `orders`.

## Consultar utilizando Evans

O servidor possui reflection habilitada. Com o [Evans](https://github.com/ktr0731/evans) instalado, conecte-se:

```bash
evans -r --host localhost --port 50051 repl
```

No terminal interativo do Evans:

```text
package pb
service OrderService
call ListOrders
```

Exemplo de resposta:

```json
{
  "orders": [
    {
      "id": "order-1",
      "price": 100.5,
      "tax": 0.5,
      "finalPrice": 101
    }
  ]
}
```

Para criar uma order, execute `call CreateOrder` e preencha `id`, `price` e `tax` quando solicitado.

Sem registros, o Evans pode exibir `{}`, pois campos repetidos vazios podem ser omitidos na representação JSON do Protobuf. Isso representa uma lista vazia.

O Evans é apenas um cliente de teste opcional. Ele não é necessário para iniciar a aplicação.

---

# GraphQL

O schema está em [schema.graphqls](internal/infra/graph/schema.graphqls).

Abra o Playground em:

```text
http://localhost:8080/
```

## Listar orders

```graphql
query {
  ListOrders {
    id
    Price
    Tax
    FinalPrice
  }
}
```

Exemplo de resposta:

```json
{
  "data": {
    "ListOrders": [
      {
        "id": "order-1",
        "Price": 100.5,
        "Tax": 0.5,
        "FinalPrice": 101
      }
    ]
  }
}
```

Também é possível consultar pelo terminal:

```bash
curl http://localhost:8080/query \
  -H 'Content-Type: application/json' \
  -d '{"query":"query { ListOrders { id Price Tax FinalPrice } }"}'
```

Sem registros, o campo `data.ListOrders` contém `[]`.

## Criar uma order

```graphql
mutation {
  createOrder(input: {id: "order-graphql-1", Price: 200, Tax: 10}) {
    id
    Price
    Tax
    FinalPrice
  }
}
```

GraphQL diferencia maiúsculas de minúsculas. A query utiliza `ListOrders`, conforme o enunciado; os campos `Price`, `Tax` e `FinalPrice` mantêm os nomes do exemplo da aula.

---

# Clean Architecture

A separação de responsabilidades da listagem está organizada da seguinte forma:

| Componente | Responsabilidade |
| --- | --- |
| `entity.Order` | Representar a order e suas regras |
| `entity.OrderRepositoryInterface` | Definir o contrato de persistência e consulta |
| `usecase.ListOrdersUseCase` | Coordenar a listagem e produzir os DTOs de saída |
| `database.OrderRepository` | Implementar o contrato utilizando SQL |
| `web.WebListOrdersHandler` | Adaptar a listagem para HTTP e JSON |
| `service.OrderService` | Adaptar a listagem para mensagens gRPC |
| `graph.Resolver` | Disponibilizar os casos de uso aos resolvers GraphQL |
| `cmd/ordersystem` | Montar as dependências e iniciar os servidores |

O caso de uso depende da interface do repositório. Ele não importa MySQL, HTTP, gRPC ou GraphQL.

O Wire gera o código que monta o repositório e o caso de uso. O `main` compartilha o caso de uso de listagem com os três adaptadores.

O evento `OrderCreated` e seu handler de publicação no RabbitMQ foram mantidos da aplicação base. A listagem não publica eventos. O projeto não inclui um consumidor RabbitMQ nem o provisionamento de uma fila vinculada ao exchange utilizado pelo handler; portanto, iniciar o broker não significa que as publicações serão armazenadas em uma fila.

---

# Estrutura do projeto

Principais arquivos e diretórios:

```text
.
├── cmd/ordersystem/
│   ├── main.go
│   ├── wire.go
│   └── wire_gen.go
├── configs/
│   └── config.go
├── internal/
│   ├── entity/
│   ├── event/
│   ├── infra/
│   │   ├── database/
│   │   ├── graph/
│   │   ├── grpc/
│   │   └── web/
│   └── usecase/
│       ├── create_order.go
│       └── list_orders.go
├── migrations/
│   ├── 000001_create_orders.up.sql
│   └── 000001_create_orders.down.sql
├── pkg/events/
├── api.http
├── Dockerfile
├── docker-compose.yaml
├── gqlgen.yml
├── go.mod
├── go.sum
└── README.md
```

---

# Variáveis de ambiente

O Compose fornece as configurações ao container da aplicação. O Viper permite carregar um `.env` opcional para execução local; as variáveis de ambiente têm prioridade sobre o arquivo.

| Variável | Valor no Compose | Finalidade |
| --- | --- | --- |
| `DB_DRIVER` | `mysql` | Driver SQL |
| `DB_HOST` | `mysql` | Host do banco na rede Docker |
| `DB_PORT` | `3306` | Porta interna do MySQL |
| `DB_USER` | `orders` | Usuário da aplicação |
| `DB_PASSWORD` | `orders` | Senha do banco |
| `DB_NAME` | `orders` | Nome do banco |
| `WEB_SERVER_PORT` | `:8000` | Endereço de escuta REST |
| `GRPC_SERVER_PORT` | `50051` | Porta gRPC |
| `GRAPHQL_SERVER_PORT` | `8080` | Porta GraphQL |
| `RABBITMQ_URL` | `amqp://labs:labs@rabbitmq:5672/` | Endereço e credenciais do RabbitMQ |

Para alterar essas configurações no ambiente Docker, ajuste o bloco `environment` do serviço `app` no `docker-compose.yaml`. Se alterar as portas internas da aplicação, ajuste também os respectivos mapeamentos em `ports`.

As credenciais apresentadas são destinadas ao ambiente local do desafio.

---

# Comandos úteis

## Acompanhar os containers

```bash
docker compose ps -a
```

O container `migrate` terminar com código `0` é esperado: sua função é aplicar migrações e encerrar.

## Consultar os logs da aplicação

```bash
docker compose logs -f app
```

## Reconstruir após alterações no código

```bash
docker compose up --build
```

## Encerrar preservando os dados

```bash
docker compose down
```

## Remover o ambiente e seus dados

```bash
docker compose down -v
```

**Atenção:** a opção `-v` remove os volumes e apaga os dados do MySQL e do RabbitMQ deste projeto.

## Executar os testes localmente

Com Go 1.26 e um compilador C instalado para os testes que utilizam SQLite:

```bash
go test ./...
```

Esses requisitos locais não são necessários para executar a aplicação pelo Docker Compose.

---

# Autor

Daniel Santello
