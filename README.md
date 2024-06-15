# Ticket Booking System Simulation

This project simulates a ticket booking system to demonstrate various concurrency patterns in Go, including goroutines, channels, mutexes, and worker pools.

## Overview

The goal of this project is to simulate a scenario where multiple users try to book a limited number of tickets concurrently. The system handles concurrent booking requests using a worker pool and ensures that no more tickets are booked than available. Additionally, it tracks the number of successful and failed booking attempts.

## Concurrency Patterns Used

- **Goroutines**: Lightweight threads managed by the Go runtime to handle concurrent booking requests.
- **Channels**: Used for communication between the main function and worker goroutines.
- **Mutex**: Ensures safe access to shared data (total and booked tickets) between multiple goroutines.
- **Worker Pool**: Limits the number of concurrent booking attempts to a manageable level.

## Code Explanation

### TicketBookingSystem Struct

The `TicketBookingSystem` struct maintains the total number of tickets, the number of booked tickets, and the number of failed booking attempts. It uses a mutex to ensure that the ticket booking process is thread-safe.

### WorkerPool Struct

The `WorkerPool` struct manages the workers that process booking requests. It includes the number of workers, channels for requests and results, and a wait group to synchronize the workers.

### Booking Logic

The `BookTicket` method of `TicketBookingSystem` checks if there are any tickets left and, if so, books a ticket. Otherwise, it increments the count of failed booking attempts.

### Worker Function

Each worker processes booking requests from the `request` channel and sends the result to the `result` channel. A random delay simulates real-world conditions where booking requests do not arrive all at once.

### Main Function

The `main` function initializes the ticket booking system and the worker pool, starts the booking process, and prints the results.

## How to Run

1. Clone the repository:
    ```bash
    git clone https://github.com/yourusername/ticket-booking-simulation.git
    ```

2. Navigate to the project directory:
    ```bash
    cd ticket-booking-simulation
    ```

3. Run the Go program:
    ```bash
    go run main.go totalTickets=50 totalUsers=55 workerCount=10
    ```
