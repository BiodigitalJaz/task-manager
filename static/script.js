 // Function to submit a new task
 function submitTask(event) {
    event.preventDefault(); // Prevent default form submission

    // Capture form input values
    const title = document.getElementById('title').value;
    const description = document.getElementById('description').value;
    const user_id = document.getElementById('user_id').value;
    const start_time = document.getElementById('start_time').value;

    const formattedDate = formatDate(start_time)
    // Create a JSON object with form data
    const taskData = {
        title: title,
        description: description,
        user_id: user_id,
        start_time: formattedDate,
    };
    fetch('/tasks', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(taskData)
    }).then(response => {
        if (response.ok) {
            // If the response is successful, reload the task list
            document.getElementById('task-list').innerHTML = '<hx:include src="/tasks" hx-trigger="load"></hx:include>';
            fetchTaskList();
            event.target.reset();
        } else {
            console.error('Failed to add task');
        }
    }).catch(error => {
        console.error('Error adding task:', error);
    });
    window.onload = fetchTaskList;

    document.getElementById('taskForm').addEventListener('submit', submitTask);
}
// Function to format datetime value as "mm/dd/yy hh:mm:ss"
function formatDate(datetimeValue) {
    // Create a new Date object from the datetime value
    const date = new Date(datetimeValue);

    // Extract date and time components
    const year = date.getFullYear().toString().slice(-2); // Get last two digits of year
    const month = ('0' + (date.getMonth() + 1)).slice(-2); // Month is zero-indexed
    const day = ('0' + date.getDate()).slice(-2);
    const hour = ('0' + date.getHours()).slice(-2);
    const minute = ('0' + date.getMinutes()).slice(-2);
    const second = ('0' + date.getSeconds()).slice(-2);

    // Construct formatted datetime string
    const formattedDatetime = month + '/' + day + '/' + year + ' ' + hour + ':' + minute + ':' + second;


    return formattedDatetime;
}
// Function to fetch and display task list
async function fetchTaskList() {
    try {
        const response = await fetch('/tasks'); // Fetch data from your Go server
        const data = await response.json(); // Extract JSON data from the response

        // Clear previous task list
        const taskListElement = document.getElementById('task-list');
        taskListElement.innerHTML = '';

        // Update task list with data from the server
        data.forEach(task => {

            //create the container and add the task ID
            const taskItem = document.createElement('li');
            taskItem.classList.add('list-group-item');
            taskItem.dataset.taskId = task.id;

            //Add title
            const taskTitle = document.createElement('h5');
            if (task.completed) {
                taskTitle.classList.add('list-group-item-completed')
            } else {
                taskTitle.classList.add('list-group-item')
            }
            taskTitle.textContent = task.id + ': ' + task.title;
            taskItem.appendChild(taskTitle);

            if (task.completed) {
                const completedtime = document.createElement('p');
                completedtime.classList.add('list-group-item-completed')
                completedtime.textContent = 'Completed at: ' + task.completed_time;
                taskItem.appendChild(completedtime);
            }

            if (!task.completed) {
                // Add start time
                const startTime = document.createElement('p');
                startTime.classList.add('list-group-item')
                startTime.textContent = 'Start Time: ' + task.start_time;
                taskItem.appendChild(startTime);

                // Add description
                const taskDescription = document.createElement('p');
                taskDescription.classList.add('list-group-item')
                taskDescription.textContent = 'Description: ' + task.description;
                taskItem.appendChild(taskDescription);

                // Add completed button
                const completeButton = document.createElement('button');
                completeButton.textContent = 'Complete';
                completeButton.classList.add('btn', 'btn-success');
                completeButton.addEventListener('click', handleCompleteButtonClick);
                taskItem.appendChild(completeButton);

                // Add update buttons
                const updateButton = document.createElement('button');
                updateButton.textContent = 'Update';
                updateButton.classList.add('btn', 'btn-primary');
                updateButton.addEventListener('click', handleUpdateButtonClick);
                taskItem.appendChild(updateButton);


            }
            // Add delete buttons
            const deleteButton = document.createElement('button');
            deleteButton.textContent = 'Delete';
            deleteButton.classList.add('btn', 'btn-danger');
            deleteButton.addEventListener('click', handleDeleteButtonClick);
            taskItem.appendChild(deleteButton);

            taskListElement.appendChild(taskItem);
        });
    } catch (error) {
        console.error('Error fetching task list:', error);
    }
}
// Function to handle update button click
async function handleUpdateButtonClick(event) {
    const taskId = event.target.parentElement.dataset.taskId;
    console.log('Update task with ID:', taskId);

    // Fetch updated task details from input fields
    const updatedTitle = prompt('Enter updated title:');
    const updatedDescription = prompt('Enter updated description:');

    // Send PUT request to update task
    try {
        const response = await fetch(`/update-task/${taskId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ title: updatedTitle, description: updatedDescription })
        });
        if (!response.ok) {
            throw new Error('Failed to update task');
        }
        // Fetch and display updated task list
        fetchTaskList();
    } catch (error) {
        console.error('Error updating task:', error);
    }
}
// Function to handle delete button click
async function handleDeleteButtonClick(event) {
    const taskId = event.target.parentElement.dataset.taskId;
    console.log('Delete task with ID:', taskId);

    // Confirm deletion with user
    if (confirm('Are you sure you want to delete this task?')) {
        // Send DELETE request to delete task
        try {
            const response = await fetch(`/delete-task/${taskId}`, {
                method: 'DELETE'
            });
            if (!response.ok) {
                throw new Error('Failed to delete task');
            }
            // Fetch and display updated task list
            fetchTaskList();
        } catch (error) {
            console.error('Error deleting task:', error);
        }
    }
}
//function to delete a task
async function handleCompleteButtonClick(event) {
    const taskId = event.target.parentElement.dataset.taskId;
    console.log('Complete task with ID:', taskId);

    // Send PUT request to update task
    try {
        const response = await fetch(`/complete-task/${taskId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ completed: true })
        });
        if (!response.ok) {
            throw new Error('Failed to complete task');
        }
        // Fetch and display updated task list
        fetchTaskList();
    } catch (error) {
        console.error('Error completing task:', error);
    }
}
// Call fetchTaskList when the page loads
window.onload = fetchTaskList;