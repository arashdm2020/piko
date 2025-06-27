#!/bin/bash

# Check if MySQL client is installed
if ! command -v mysql &> /dev/null; then
    echo "MySQL client is not installed. Please install it first."
    exit 1
fi

# Initialize the database
echo "Creating database and tables..."
mysql -u root < database/init.sql

# Check if the operation was successful
if [ $? -eq 0 ]; then
    echo "Database initialization completed successfully."
else
    echo "Failed to initialize database. Please check your MySQL connection and credentials."
    exit 1
fi

echo "Setup complete. You can now run the application with 'go run main.go'" 