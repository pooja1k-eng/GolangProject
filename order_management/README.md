Order Management Microservices

A microservices system built in Go featuring an API Gateway, an Order Service with PostgreSQL persistence, an inter-service resilient Product client, and a unified GraphQL interface.

1. Prerequisites

Go: Version 1.22+ installed

Docker: For running the local PostgreSQL container

2. Environment Configuration

Create a file named .env in the root of your project:

# Database Settings
DB_HOST=127.0.0.1
DB_PORT=5432
DB_NAME=orderdb
DB_USER=postgres
DB_PASSWORD=postgres

# Ports
GATEWAY_PORT=8082
PRODUCT_SERVICE_PORT=8081
PRODUCT_SERVICE_URL=http://localhost:8081
ORDER_SERVICE_PORT=10004
ORDER_SERVICE_URL=http://localhost:10004

# Application
LOG_LEVEL=info
AUTH_TOKEN=secret-bearer-token


3. How to Run

Step 1: Start PostgreSQL

docker run --name order-postgres -e POSTGRES_DB=orderdb -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:15-alpine


(If the container is already created, run: docker start order-postgres)

Step 2: Start the Microservices (3 Separate Terminals)

Terminal 1 — Product Service:

go run product/main.go


Terminal 2 — Order Service:

go run order/main.go


Terminal 3 — API Gateway:

go run gateway/main.go


4. API Workflow (Run in Order)

1. Health Check (Readiness)

Method: GET

URL: http://localhost:8082/health/ready

Headers: None

Response: 200 OK

{
  "status": "READY"
}


2. Login & Generate JWT Token

Method: POST

URL: http://localhost:8082/api/v1/auth/login

Headers:

Content-Type: application/json

Body:

{
  "email": "admin@example.com",
  "password": "Admin@123"
}


Response: 200 OK

{
  "expiresIn": 7200,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "tokenType": "Bearer"
}


(Copy the string inside "token" for all subsequent requests)

3. Create Order

Method: POST

URL: http://localhost:8082/api/v1/orders

Headers:

Content-Type: application/json

Authorization: Bearer <PASTE_TOKEN_HERE>

Body:

{
  "customerId": 101,
  "productId": 501,
  "quantity": 2
}


Response: 201 Created

{
  "id": 1,
  "customerId": 101,
  "productId": 501,
  "quantity": 2,
  "unitPrice": 89.99,
  "totalPrice": 179.98,
  "status": "CREATED"
}


4. Get Order by ID

Method: GET

URL: http://localhost:8082/api/v1/orders/1

Headers:

Authorization: Bearer <PASTE_TOKEN_HERE>

Response: 200 OK

5. Update Order Status

Method: PUT

URL: http://localhost:8082/api/v1/orders/1/status

Headers:

Content-Type: application/json

Authorization: Bearer <PASTE_TOKEN_HERE>

Body:

{
  "status": "CONFIRMED"
}


Response: 200 OK

{
  "message": "Status updated successfully"
}


6. GraphQL Query & Mutation

Method: POST

URL: http://localhost:8082/graphql

Headers:

Content-Type: application/json

Query (Fetch Orders)

{
  "query": "query { orders { id customerId totalPrice status } }"
}


Mutation (Update Status)

{
  "query": "mutation { updateOrderStatus(id: 1, status: \"CONFIRMED\") { id status totalPrice } }"
}
