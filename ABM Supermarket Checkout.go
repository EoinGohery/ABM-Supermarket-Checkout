/*The Bri4field:
Eoin Gohery - 17206413,
Donal O’Connell - 17241499,
Cian McInerney - 17232724,
William Cummins – 17234956 */

package main

import (
	"fmt"
	"math/rand"
	"time"
)

var expressTills, allTills []*till
var lostCustomers, custsProcessed []*customer
var customersLeft, id, customersProcessed, stockProcessed, seeCustomers, numItemLimit int
var totalWait float64
var customerWaitTime = make(map[int]float64)

// Customer struct
type customer struct {
	NoItems int
	id      int
	Created time.Time
}

// Till struct
type till struct {
	queue             chan *customer
	tillNo            int
	delay             int
	tillType          int
	totalUtilised     float64
	customerProcessed int
}

// Returns weather, switch case with all weather types and their values
func getWeather(weatherInput int) int {
	switch weatherInput {
	case 1:
		return 2
	case 2:
		return 1
	case 3:
		return 3
	case 4:
		return 100
	case 5:
		return 10
	default:
		return 2
	}
}

// Returns number of tills that the user requests during input
func getTillNum(tillnumber int) int {
	switch tillnumber {
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 3
	case 4:
		return 4
	case 5:
		return 5
	case 6:
		return 6
	case 7:
		return 7
	case 8:
		return 8

	default:
		return 4
	}
}

// Allows the user to check customers choices through terminal
func seeCustomersChoice(seeCustomers int) int {
	switch seeCustomers {
	case 1:
		return 1
	case 2:
		return 2
	default:
		return 1
	}
}

// Generates customers and place them into a channel of customer pointers,
func customersGenerator(customers chan<- *customer, weatherVal int, productsInputFirst int, productsInputSecond int, spawnRateFirst int, spawnRateSecond int, expressCustomer chan<- *customer, quit <-chan bool) {
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				stockMult := rand.Intn(productsInputSecond-productsInputFirst) + productsInputFirst
				spawnRate := rand.Intn(spawnRateSecond-spawnRateFirst) + spawnRateFirst
				id++
				newCustomer := customer{stockMult, id, time.Now()}
				if stockMult <= numItemLimit {
					if expressFull() {
						if tillsFull() {
							customersLeft++
							//fmt.Println("customer left")
							lostCustomer := &newCustomer
							lostCustomers = append(lostCustomers, lostCustomer)
						} else {
							customers <- &newCustomer
							stockProcessed = stockProcessed + stockMult
							customersProcessed++
							custsProcessed = append(custsProcessed, &newCustomer)
						}
					} else {
						expressCustomer <- &newCustomer
						stockProcessed = stockProcessed + stockMult
						customersProcessed++
						custsProcessed = append(custsProcessed, &newCustomer)
					}
				} else {
					if tillsFull() {
						customersLeft++
						//fmt.Println("customer left")
						lostCustomer := &newCustomer
						lostCustomers = append(lostCustomers, lostCustomer)
					} else {
						customers <- &newCustomer
						stockProcessed = stockProcessed + stockMult
						customersProcessed++
						custsProcessed = append(custsProcessed, &newCustomer)
					}
				}
				//wait after each customer generation based on user input
				time.Sleep(time.Millisecond * time.Duration(spawnRate+weatherVal))
			}
		}
	}()
}

// Each till is ran in seperate versions of this function. the queus are buffered channels of size 6 . These channels are filled from the main customers channel.
// Checks the type of till that will be used (0 -> Normal Till) (1 -> Express)
func runTill(myTill *till, customers <-chan *customer, timePerItemFirst float64, timePerItemSecond float64, customerData chan<- []float64) {
	go func() {
		for {
			if myTill.tillType == 0 {
				if len(myTill.queue) == getShortestQueue() {
					n := <-customers
					myTill.queue <- n
					//fmt.Println("Customer on Till:", myTill.tillNo)
				}
			} else if myTill.tillType == 1 {
				if len(myTill.queue) == getShortestExpressQueue() {
					n := <-customers
					myTill.queue <- n
					//fmt.Println("Customer on Till:", myTill.tillNo)
				}
			}

		}
	}()

	//Pulls customers from the queue and processes them
	go func() {
		for {
			m := <-myTill.queue
			customer := *m
			timePerItem := timePerItemFirst + rand.Float64()*(timePerItemSecond-timePerItemFirst)
			//fmt.Println("Customer Processed", customer)
			time.Sleep(time.Millisecond * time.Duration((float64(customer.NoItems)*timePerItem)+60.0*float64(myTill.delay)))
			currentTime := time.Now()
			diff := float64(currentTime.Sub(customer.Created))
			totalWait = totalWait + (diff / 1e+6 / 60)
			myTill.customerProcessed++
			myTill.totalUtilised = myTill.totalUtilised + (diff / 1e+6 / 60)
			data := make([]float64, 2)
			data[0] = float64(id)
			data[1] = (diff / 1e+6 / 60)
			customerData <- data
		}
	}()

}

// Gets the range of two ints which the user inputs
func userInput() (int, int) {
	fmt.Println("Enter First number:")
	first := 0
	fmt.Scan(&first)
	fmt.Println("Enter Second number:")
	second := 0
	fmt.Scan(&second)
	firstType := fmt.Sprintf("%T", first)
	secondType := fmt.Sprintf("%T", second)
	if first < 0 || first >= second || second > 60 || firstType != "int" || secondType != "int" {
		first = 0
		second = 60
	}
	return first, second

}

func customerManager(customerData <-chan []float64) {
	for {
		data := <-customerData
		id := int(data[0])
		processTime := data[1]
		customerWaitTime[id] = processTime
	}
}

// Gets the range of two floats which the user inputs
func getScanTime() (float64, float64) {
	fmt.Println("Enter First number:")
	first := 0.0
	fmt.Scan(&first)
	fmt.Println("Enter Second number:")
	second := 0.0
	fmt.Scan(&second)
	firstType := fmt.Sprintf("%T", first)
	secondType := fmt.Sprintf("%T", second)
	if first < 0.5 || first >= second || second > 6.0 || firstType != "float64" || secondType != "float64" {
		first = 0.5
		second = 6.0
	}
	return first, second

}

func setMaxItemLimit(maxitems int) int {
	maxitemType := fmt.Sprintf("%T", maxitems)
	if maxitems < 5 || maxitems > 15 || maxitemType != "int" {
		maxitems = 15
	}
	return maxitems

}

// Gets shortest queue for customers using normal tills, this function is needed for a customer wanting to join the shortest queue
func getShortestQueue() int {
	shortest := 6
	for i := 0; i < len(allTills); i++ {
		x := len(allTills[i].queue)
		if x < shortest {
			shortest = x
		}
	}
	return shortest
}

// Gets shortest queue for customers using express tills, this function is needed for a customer wanting to join the shortest queue
func getShortestExpressQueue() int {
	shortest := 6
	for i := 0; i < len(expressTills); i++ {
		x := len(expressTills[i].queue)
		if x < shortest {
			shortest = x
		}
	}
	return shortest
}

// Check if the normal tills are full
func tillsFull() bool {
	full := true
	for i := 0; i < len(allTills); i++ {
		x := len(allTills[i].queue)
		if x != 6 {
			full = false
		}
	}
	return full
}

// Checks if the express tills are full
func expressFull() bool {
	full := true
	for i := 0; i < len(expressTills); i++ {
		x := len(expressTills[i].queue)
		if x != 6 {
			full = false
		}
	}
	return full
}

func main() {
	customers := make(chan *customer)
	expressCustomer := make(chan *customer)
	customerData := make(chan []float64)
	quit := make(chan bool)

	customersLeft, customersProcessed, id, seeCustomers = 0, 0, 0, 0
	var newTills, oldTills, noExpressTills, weatherInput int

	// Weather input
	fmt.Println("Weather Type: 1: Mild, 2: Sunny, 3:  Wet, 4: Avalanche, 5: Thunderstorm. Default: Mild")
	fmt.Scan(&weatherInput)
	weatherInput = getWeather(weatherInput)

	// How often should the customers spawn
	fmt.Println("How Often should customers spawn: Enter two numbers in a range from 0-60:")
	spawnRateFirst, spawnRateSecond := userInput()
	if spawnRateSecond > 60 {
		spawnRateSecond = 59
	}

	// Range of products
	fmt.Println("How many products for each trolley: Enter Two Numbers for the range 0 to 200")
	productsInputFirst, productsInputSecond := userInput()
	if productsInputSecond > 200 {
		productsInputSecond = 200
	}
	//Number of new Tills
	fmt.Println("How many new tills? (range 1-8), Default: 4")
	fmt.Scan(&newTills)
	newTills = getTillNum(newTills)

	//Number of old tills
	fmt.Println("How many old tills? (range 1-8), Default: 4")
	fmt.Scan(&oldTills)
	oldTills = getTillNum(oldTills)

	//Number of express tills
	fmt.Println("How many express tills? (range 1-8), Default: 4")
	fmt.Scan(&noExpressTills)
	noExpressTills = getTillNum(noExpressTills)

	//Max number of items that can enter the express queue
	fmt.Println("What will the max number of items allowed in the express lane")
	fmt.Scan(&numItemLimit)
	numItemLimit = setMaxItemLimit(numItemLimit)

	// Time for each product to be entered at the checkout
	fmt.Println("How much time do you want for each product at checkout: Enter Two Numbers for the range 0.5 to 6")
	timePerItemFirst, timePerItemSecond := getScanTime()
	if timePerItemSecond > 6 {
		timePerItemSecond = 5.9
	}

	totalTills := oldTills + newTills

	//Generates tills based on user input, adds them to a slice and runs them on seperate go routines
	for i := 0; i < totalTills; i++ {
		if i < newTills {
			allTills = append(allTills, &till{make(chan *customer, 6), i, 1, 0, 0.0, 0})
			go runTill(allTills[i], customers, timePerItemFirst, timePerItemSecond, customerData)
		} else {
			allTills = append(allTills, &till{make(chan *customer, 6), i, 2, 0, 0.0, 0})
			go runTill(allTills[i], customers, timePerItemFirst, timePerItemSecond, customerData)
		}
	}

	for i := 0; i < noExpressTills; i++ {
		expressTills = append(expressTills, &till{make(chan *customer, 6), i + totalTills, 1, 1, 0.0, 0})
		go runTill(expressTills[i], expressCustomer, timePerItemFirst, timePerItemSecond, customerData)
	}
	// Run the customer generator on a seperate Till
	go customersGenerator(customers, weatherInput, productsInputFirst, productsInputSecond, spawnRateFirst, spawnRateSecond, expressCustomer, quit)
	go customerManager(customerData)

	time.Sleep(time.Millisecond * 840) // 14 hours, 60 minutes an hour. each milisecond repersents a minute
	quit <- true
	time.Sleep(time.Millisecond * 100)

	fmt.Println("\nCustomers lost: ", customersLeft)
	fmt.Println("Customers processed: ", customersProcessed)
	fmt.Println("Stock processed: ", stockProcessed)
	fmt.Println("Average stock per trolley: ", stockProcessed/customersProcessed)
	s := fmt.Sprintf("%.2f", totalWait/float64(customersProcessed))
	fmt.Println("Average customers wait: ", s, "minutes")

	// Prints out all normal tills and express tills that are being used and their utilisation
	for i := 0; i < totalTills; i++ {
		currentTillType := ""
		if allTills[i].delay == 2 {
			currentTillType = "Old"
		} else {
			currentTillType = "New"
		}
		fmt.Println("\nTill No:", allTills[i].tillNo, "Type:", currentTillType)
		s := fmt.Sprintf("%.2f", allTills[i].totalUtilised)
		fmt.Println("Total checkout Utilisation: ", s, "minutes")
		f := fmt.Sprintf("%.2f", allTills[i].totalUtilised/float64(allTills[i].customerProcessed))
		fmt.Println("Average checkout Utilisation: ", f, "minutes per customer")
	}
	for i := 0; i < noExpressTills; i++ {
		fmt.Println("\nTill No:", expressTills[i].tillNo, "Type: Express")
		s := fmt.Sprintf("%.2f", expressTills[i].totalUtilised)
		fmt.Println("Total checkout Utilisation: ", s, "minutes")
		f := fmt.Sprintf("%.2f", expressTills[i].totalUtilised/float64(expressTills[i].customerProcessed))
		fmt.Println("Average checkout Utilisation: ", f, "minutes per customer")
	}

	// Giving the user an option to see the list of processed and lost customers during a 14 hour simulation of a supermarket
	fmt.Println("\nWould you like to see the list of processed and lost customers?\n1)Yes 2)No")
	fmt.Scan(&seeCustomers)
	if seeCustomersChoice(seeCustomers) == 1 {
		fmt.Println("\nCustomers Processed:")
		for z := 0; z < len(custsProcessed); z++ {
			fmt.Printf("Customer id: %v  Number of items:%2v  Time Taken to process : %.2f\n", custsProcessed[z].id, custsProcessed[z].NoItems, customerWaitTime[custsProcessed[z].id])
		}
		fmt.Println("\nCustomers Lost:")
		for z := 0; z < len(lostCustomers); z++ {
			fmt.Printf("Lost Customer id: %v  Number of items:%v\n", lostCustomers[z].id, lostCustomers[z].NoItems)
		}
	}
}
