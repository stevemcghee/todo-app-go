# Go To-Do Application

A simple To-Do list application built with Go and PostgreSQL. This application allows users to create, view, update, and delete to-do items.

## Features

*   **Create To-Do Items**: Add new tasks to your list.
*   **View To-Do Items**: See all your pending and completed tasks.
*   **Update To-Do Items**: Mark tasks as completed.
*   **Delete To-Do Items**: Remove tasks from your list.
*   **Persistent Storage**: To-Do items are stored in a PostgreSQL database.

## Prerequisites

Before you begin, ensure you have the following installed:

*   [Docker](https://docs.docker.com/get-docker/)
*   [Docker Compose](https://docs.docker.com/compose/install/)

## Getting Started

Follow these steps to get the application up and running on your local machine.

### 1. Clone the Repository

```bash
git clone https://github.com/gemini/todo-app.git
cd todo-app
```

### 2. Create a `.env` file

Create a file named `.env` in the root directory of the project and add the following content. Replace `your_password` with a strong password for your PostgreSQL database.

```
POSTGRES_USER=user
POSTGRES_PASSWORD=your_password
POSTGRES_DB=todoapp_db
```

### 3. Build and Run with Docker Compose

Navigate to the project root directory and run the following command to build the Docker images and start the services:

```bash
docker-compose up --build
```

This command will:
*   Build the Go application Docker image.
*   Start a PostgreSQL database container.
*   Initialize the database with the `init.sql` script, creating the `todos` table.
*   Start the Go application, connecting it to the PostgreSQL database.

### 4. Access the Application

Once the services are up and running, you can access the application in your web browser at:

[http://localhost:8080](http://localhost:8080)

## API Endpoints

The application exposes the following API endpoints:

*   **`GET /todos`**: Retrieve all to-do items.
*   **`POST /todos`**: Add a new to-do item.
    *   Request Body: `{"task": "New task description"}`
*   **`PUT /todos/{id}`**: Update a to-do item (e.g., mark as completed).
    *   Request Body: `{"completed": true}`
*   **`DELETE /todos/{id}`**: Delete a to-do item.

## Technologies Used

*   **Backend**: Go
*   **Database**: PostgreSQL
*   **Containerization**: Docker, Docker Compose
*   **Frontend**: HTML, CSS, JavaScript (served statically)
