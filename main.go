package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type TicketBookingSystem struct {
	totalTickets, bookedTickets int64
	mx                          sync.Mutex
	totalFailed                 int64
}

type TBS struct {
	totalT, bookedT, failed int64
	m                       sync.Mutex
}

type WorkerPool struct {
	workerCount int
	tbs         *TicketBookingSystem
	request     chan int
	result      chan string
	wg          sync.WaitGroup
}

type WP struct {
	wCount  int
	tbs     *TBS
	request chan int
	result  chan string
	wg      sync.WaitGroup
}

func (tbs *TicketBookingSystem) BookTicket() bool {
	tbs.mx.Lock()
	defer tbs.mx.Unlock()

	if tbs.totalTickets > tbs.bookedTickets {
		tbs.bookedTickets++
		return true
	}
	tbs.totalFailed++

	return false
}

func NewWorkerPool(workerCount int, totalUsers int, tbs *TicketBookingSystem) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		tbs:         tbs,
		request:     make(chan int, totalUsers),
		result:      make(chan string, totalUsers),
	}
}

func NewWP(workerCount int, totalUsers int, tbs *TBS) *WP {
	return &WP{
		wCount:  workerCount,
		tbs:     tbs,
		request: make(chan int, totalUsers),
		result:  make(chan string, totalUsers),
	}
}

// Worker this is actual worker to book a ticket and respond with a result
func (wp *WorkerPool) Worker(id int) {
	defer wp.wg.Done()
	for userID := range wp.request {
		fmt.Printf("worker %d processing user %d\n", id, userID)
		if wp.tbs.BookTicket() {
			wp.result <- fmt.Sprintf("User %d successfully booked a ticket.", userID)
		} else {
			wp.result <- fmt.Sprintf("User %d failed to book a ticket.", userID)
		}
	}
}

func (wp *WorkerPool) Start(totalUsers int) {
	// start worker pool
	for w := 1; w <= wp.workerCount; w++ {
		wp.wg.Add(1)
		go wp.Worker(w)
	}

	// send booking request
	for i := 1; i <= totalUsers; i++ {
		wp.request <- i
		time.Sleep(time.Duration(rand.Intn(400)+100) * time.Millisecond)
	}
	close(wp.request)

	wp.wg.Wait()
	close(wp.result)
}

func (wp *WorkerPool) Done() {
	bookingCount := make(map[int]int)
	for result := range wp.result {
		fmt.Println(result)
		var userID int
		if _, err := fmt.Sscanf(result, "User %d successfully booked a ticket.", &userID); err == nil {
			bookingCount[userID]++
		}
	}

	duplicates := false
	for userID, count := range bookingCount {
		if count > 1 {
			fmt.Printf("User %d booked multiple tickets: %d times\n", userID, count)
			duplicates = true
		}
	}

	if !duplicates {
		fmt.Println("No duplicate ticket bookings found.")
	}
	fmt.Printf("Total tickets booked: %d and total failed bookings are %d\n", wp.tbs.bookedTickets, wp.tbs.totalFailed)
}

func main() {
	var totalTickets int64
	var totalUsers int
	var workerCount int

	// Define command-line flags
	flag.Int64Var(&totalTickets, "totalTickets", 50, "Total number of tickets available")
	flag.IntVar(&totalUsers, "totalUsers", 100, "Total number of users trying to book tickets")
	flag.IntVar(&workerCount, "workerCount", 10, "Number of workers processing the bookings")
	flag.Parse()

	// Print the parsed values
	fmt.Printf("Running with totalTickets=%d totalUsers=%d workerCount=%d\n", totalTickets, totalUsers, workerCount)

	tbs := &TicketBookingSystem{totalTickets: int64(totalTickets)}
	workerPool := NewWorkerPool(workerCount, totalUsers, tbs)

	workerPool.Start(totalUsers)
	workerPool.Done()
}
