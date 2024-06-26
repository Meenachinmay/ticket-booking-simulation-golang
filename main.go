package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type TicketBookingSystem struct {
	totalTickets, bookedTickets int64
	mx                          sync.Mutex
	totalFailed                 int64
	tBooked                     map[int]int
	bookingRequests             chan BookingRequest
}

type BookingRequest struct {
	userID int
	result chan bool
}

type WorkerPool struct {
	workerCount int
	tbs         *TicketBookingSystem
	request     chan int
	result      chan string
	wg          sync.WaitGroup
}

func NewTicketBookingSystem(totalTickets int64) *TicketBookingSystem {
	tbs := &TicketBookingSystem{
		totalTickets:    totalTickets,
		tBooked:         make(map[int]int),
		bookingRequests: make(chan BookingRequest, 10000),
	}
	go tbs.processBookings()
	return tbs
}

func (tbs *TicketBookingSystem) processBookings() {
	for request := range tbs.bookingRequests {
		success := false
		tbs.mx.Lock()
		if tbs.totalTickets > tbs.bookedTickets {
			tbs.bookedTickets++
			tbs.tBooked[request.userID]++
			success = true
		} else {
			tbs.totalFailed++
		}
		tbs.mx.Unlock()
		request.result <- success
	}
}

func (tbs *TicketBookingSystem) BookTicket(userID int) bool {
	result := make(chan bool)
	tbs.bookingRequests <- BookingRequest{userID: userID, result: result}
	return <-result

}

func NewWorkerPool(workerCount int, totalUsers int, tbs *TicketBookingSystem) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		tbs:         tbs,
		request:     make(chan int, totalUsers),
		result:      make(chan string, totalUsers),
	}
}

// Worker this is actual worker to book a ticket and respond with a result
func (wp *WorkerPool) Worker(id int) {
	defer wp.wg.Done()
	for userID := range wp.request {
		//fmt.Printf("worker %d processing user %d\n", id, userID)
		if wp.tbs.BookTicket(userID) {
			wp.result <- fmt.Sprintf("User %d successfully booked a ticket.", userID)
		} else {
			wp.result <- fmt.Sprintf("User %d failed to book a ticket.", userID)
		}
	}
}

// Start starting worker in this method: every call for worker inside a go routine
func (wp *WorkerPool) Start(totalUsers int) {
	// assigning a go routine equals to no of worker count
	for w := 1; w <= wp.workerCount; w++ {
		wp.wg.Add(1)
		go wp.Worker(w)
	}

	// pushing a job to the job queue in this case job queue is request channel
	for i := 1; i <= totalUsers; i++ {
		wp.request <- i
		//time.Sleep(time.Duration(rand.Intn(400)+100) * time.Millisecond)
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
	for userID, count := range wp.tbs.tBooked {
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
	flag.Int64Var(&totalTickets, "totalTickets", 500, "Total number of tickets available")
	flag.IntVar(&totalUsers, "totalUsers", 1500, "Total number of users trying to book tickets")
	flag.IntVar(&workerCount, "workerCount", 50, "Number of workers processing the bookings")
	flag.Parse()

	// Print the parsed values
	fmt.Printf("Running with totalTickets=%d totalUsers=%d workerCount=%d\n", totalTickets, totalUsers, workerCount)

	tbs := NewTicketBookingSystem(totalTickets)
	workerPool := NewWorkerPool(workerCount, totalUsers, tbs)

	workerPool.Start(totalUsers)
	workerPool.Done()

}

type Payload struct {
	Name string `json:"name"`
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST method allow ONLY", http.StatusMethodNotAllowed)
		return
	}

	var payload Payload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	//create a response
	response := map[string]string{"status": "success", "name": payload.Name}
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func connectDatabase(dbURL string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var testQuery int
	err = conn.QueryRow("SELECT 1").Scan(&testQuery)
	if err != nil {
		log.Fatal("Database query test failed...", err)
	} else {
		log.Println("Connection test query succeeded")
	}

	return conn, nil
}

type AllowedPositions interface {
	int64 | float64 | string
}

type Position[T AllowedPositions] struct {
	X, Y, Z T
}

type Entity[T AllowedPositions] interface {
	GetPosition() Position[T]
	UpdatePosition(Position[T]) Position[T]
}

type Player[T AllowedPositions] struct {
	position Position[T]
}

func (p *Player[T]) GetPosition() Position[T] {
	return p.position
}

func (p *Player[T]) UpdatePosition(position Position[T]) Position[T] {
	p.position = position
	return p.position
}

type Comparable interface {
	int32 | int64
}

type GenericSlice[T Comparable] []T

func PrintSlice[T Comparable](slice GenericSlice[T]) {
	for _, v := range slice {
		fmt.Println(v)
	}
}
