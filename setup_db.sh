#!/bin/bash

# Database connection parameters
DB_USER="root"
DB_PASSWORD=""
DB_HOST="localhost"
DB_NAME="piko"

# Check if MySQL client is installed
if ! command -v mysql &> /dev/null; then
    echo "MySQL client is not installed. Please install it first."
    exit 1
fi

# Drop the database if it exists
echo "Dropping database if it exists..."
mysql -u$DB_USER -p$DB_PASSWORD -h$DB_HOST -e "DROP DATABASE IF EXISTS $DB_NAME;"

# Create the database and initialize schema
echo "Creating database and initializing schema..."
mysql -u$DB_USER -p$DB_PASSWORD -h$DB_HOST < database/init.sql

# Check if the database was created successfully
if [ $? -eq 0 ]; then
    echo "Database setup completed successfully."
else
    echo "Error: Database setup failed."
    exit 1
fi

echo "Database is ready to use."

echo "Setup complete. You can now run the application with 'go run main.go'" 