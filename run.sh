#!/bin/bash

# Function to display help message
show_help() {
    echo "Piko - Decentralized Encrypted Messaging with Blockchain Integration"
    echo ""
    echo "Usage: ./run.sh [command]"
    echo ""
    echo "Commands:"
    echo "  start       Start the Piko application with Docker Compose"
    echo "  stop        Stop the Piko application"
    echo "  restart     Restart the Piko application"
    echo "  logs        Show logs from the Piko application"
    echo "  build       Build the Docker images"
    echo "  clean       Remove all containers and volumes"
    echo "  help        Show this help message"
    echo ""
}

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo "Error: Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        echo "Error: Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
}

# Start the application
start_app() {
    echo "Starting Piko application..."
    docker-compose up -d
    echo "Piko is now running at http://localhost:8080"
}

# Stop the application
stop_app() {
    echo "Stopping Piko application..."
    docker-compose down
}

# Restart the application
restart_app() {
    echo "Restarting Piko application..."
    docker-compose restart
}

# Show logs
show_logs() {
    docker-compose logs -f
}

# Build the Docker images
build_images() {
    echo "Building Piko Docker images..."
    docker-compose build
}

# Clean up containers and volumes
clean_up() {
    echo "Removing all containers and volumes..."
    docker-compose down -v
    echo "Cleanup complete."
}

# Main script execution
check_docker

# Process command line arguments
case "$1" in
    start)
        start_app
        ;;
    stop)
        stop_app
        ;;
    restart)
        restart_app
        ;;
    logs)
        show_logs
        ;;
    build)
        build_images
        ;;
    clean)
        clean_up
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        show_help
        exit 1
        ;;
esac

exit 0 