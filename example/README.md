# Prerequisites

- Docker installed on your system
- Two SQL files:
  - `schema.sql` - containing table definitions
  - `data.sql` - containing statements to execute (like data insertion)

# Lazy mode (using predefined script)

Setup `PostgreSQL`:

```bash
./run.sh
```

Extract records:

```bash
POSTGRES_HOST=localhost POSTGRES_PORT=5432 POSTGRES_USER=myuser POSTGRES_PASSWORD=mysecretpassword POSTGRES_DB=mydb go run github.com/desprit-media/traversql-core/cmd/traversql@v0.0.3 traverse --table=order_items --pk-fields=order_item_id --pk-values=1
```

```sql
INSERT INTO public.customers (customer_id, first_name, last_name, email, phone, date_of_birth, loyalty_points, created_at, last_login) VALUES ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'John', 'Doe', 'john.doe@example.com', '555-123-4567', '1985-07-15T00:00:00Z', 170, '2025-05-07T12:51:30.855923+03:00', NULL);
INSERT INTO public.customer_addresses (address_id, customer_id, address_type, street_address, city, state, postal_code, country, is_default) VALUES (1, 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'BOTH', '123 Main St', 'New York', 'NY', '10001', 'USA', true);
INSERT INTO public.orders (order_id, customer_id, order_date, status, shipping_address_id, billing_address_id, shipping_method, payment_method, subtotal, tax, shipping_cost, total_amount, notes) VALUES (1, 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '2025-05-07T12:51:30.85939+03:00', 'DELIVERED', 1, 1, 'Standard Shipping', 'Credit Card', 699.99, 56, 10, 765.99, NULL);
INSERT INTO public.products (product_id, name, description, sku, price, stock_quantity, weight, dimensions, created_at, updated_at) VALUES (1, 'Smartphone X', '6.5-inch display, 128GB storage', 'PHON-X-128', 699.99, 49, 0.35, '{"depth": 0.8, "width": 7.5, "height": 15.0}', '2025-05-07T12:51:30.857838+03:00', '2025-05-07T12:51:30.857838+03:00');
INSERT INTO public.order_items (order_item_id, order_id, product_id, quantity, unit_price, total_price, discount_amount) VALUES (1, 1, 1, 1, 699.99, 699.99, 0);
```

# Manual Steps

Pull the `PostgreSQL` Docker image:

```bash
docker pull postgres:16
```

Create a Docker container with `PostgreSQL`:

```bash
docker run --name postgres-traversql-example -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_USER=myuser -e POSTGRES_DB=mydb -p 5432:5432 -d postgres:16
```

Wait for PostgreSQL to start up:

```bash
sleep 5
```

Copy your SQL files to the container:

```bash
docker cp schema.sql postgres-traversql-example:/schema.sql
docker cp data.sql postgres-traversql-example:/data.sql
```

Execute the schema file to create tables:

```bash
docker exec -it postgres-traversql-example psql -U myuser -d mydb -f /schema.sql
```

Execute the data file to run additional statements:

```bash
docker exec -it postgres-traversql-example psql -U myuser -d mydb -f /data.sql
```

Extract records:

```bash
POSTGRES_HOST=localhost POSTGRES_PORT=5432 POSTGRES_USER=myuser POSTGRES_PASSWORD=mysecretpassword POSTGRES_DB=mydb go run github.com/desprit-media/traversql-core/cmd/traversql@v0.0.3 traverse --table=order_items --pk-fields=order_item_id --pk-values=1
```
