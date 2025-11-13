document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('todo-form');
    const input = document.getElementById('todo-input');
    const list = document.getElementById('todo-list');

    const fetchTodos = async () => {
        const response = await fetch('/todos');
        const todos = await response.json();
        list.innerHTML = '';
        if (todos) {
            todos.forEach(todo => {
                renderTodo(todo);
            });
        }
    };

    const renderTodo = (todo) => {
        const item = document.createElement('li');
        item.dataset.id = todo.id;
        if (todo.completed) {
            item.classList.add('completed');
        }

        const taskSpan = document.createElement('span');
        taskSpan.textContent = todo.task;
        taskSpan.addEventListener('click', () => toggleComplete(todo));

        const deleteBtn = document.createElement('button');
        deleteBtn.textContent = '×';
        deleteBtn.className = 'delete-btn';
        deleteBtn.addEventListener('click', () => deleteTodo(todo.id));

        item.appendChild(taskSpan);
        item.appendChild(deleteBtn);
        list.appendChild(item);
    };

    const addTodo = async (task) => {
        const response = await fetch('/todos', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ task }),
        });
        const newTodo = await response.json();
        renderTodo(newTodo);
    };

    const toggleComplete = async (todo) => {
        const response = await fetch(`/todos/${todo.id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ...todo, completed: !todo.completed }),
        });
        if (response.ok) {
            const li = document.querySelector(`[data-id='${todo.id}']`);
            li.classList.toggle('completed');
        }
    };

    const deleteTodo = async (id) => {
        const response = await fetch(`/todos/${id}`, {
            method: 'DELETE',
        });
        if (response.ok) {
            const li = document.querySelector(`[data-id='${id}']`);
            li.remove();
        }
    };

    form.addEventListener('submit', (e) => {
        e.preventDefault();
        const task = input.value.trim();
        if (task) {
            addTodo(task);
            input.value = '';
        }
    });

    fetchTodos();
});
