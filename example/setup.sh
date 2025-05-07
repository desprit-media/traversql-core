#!/bin/bash

# Configuration variables
CONTAINER_NAME="postgres-traversql-example"
POSTGRES_USER="myuser"
POSTGRES_PASSWORD="mysecretpassword"
POSTGRES_DB="mydb"
SCHEMA_FILE="schema.sql"
DATA_FILE="data.sql"
PORT=5432

# Check if required files exist
if [ ! -f "$SCHEMA_FILE" ]; then
    echo "Error: Schema file '$SCHEMA_FILE' not found!"
    exit 1
fi

if [ ! -f "$DATA_FILE" ]; then
    echo "Error: Data file '$DATA_FILE' not found!"
    exit 1
fi

# Check if container already exists
if [ "$(docker ps -a -q -f name=$CONTAINER_NAME)" ]; then
    echo "Container $CONTAINER_NAME already exists. Removing it..."
    docker stop $CONTAINER_NAME >/dev/null 2>&1
    docker rm $CONTAINER_NAME >/dev/null 2>&1
fi

# Pull PostgreSQL image
echo "Pulling PostgreSQL image..."
docker pull postgres:16

# Create and run the PostgreSQL container
echo "Creating PostgreSQL container..."
docker run --name $CONTAINER_NAME \
    -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
    -e POSTGRES_USER=$POSTGRES_USER \
    -e POSTGRES_DB=$POSTGRES_DB \
    -p $PORT:5432 \
    -d postgres:16

# Wait for PostgreSQL to start up
echo "Waiting for PostgreSQL to start up..."
sleep 10

# Check if container is running
if [ ! "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
    echo "Error: Container failed to start!"
    exit 1
fi

# Copy SQL files to the container
echo "Copying SQL files to the container..."
docker cp $SCHEMA_FILE $CONTAINER_NAME:/$SCHEMA_FILE
docker cp $DATA_FILE $CONTAINER_NAME:/$DATA_FILE

# Execute schema file to create tables
echo "Creating database schema..."
docker exec -it $CONTAINER_NAME psql -U $POSTGRES_USER -d $POSTGRES_DB -f /$SCHEMA_FILE

# Execute data file to run additional statements
echo "Loading data..."
docker exec -it $CONTAINER_NAME psql -U $POSTGRES_USER -d $POSTGRES_DB -f /$DATA_FILE

echo "PostgreSQL setup complete!"
echo "Connection details:"
echo "  Host: localhost"
echo "  Port: $PORT"
echo "  Database: $POSTGRES_DB"
echo "  Username: $POSTGRES_USER"
echo "  Password: $POSTGRES_PASSWORD"
