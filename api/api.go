// package api

// import (
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"time"

// 	"govulnapi/api/database"
// 	m "govulnapi/models"

// 	"github.com/go-chi/chi/v5"
// 	"github.com/go-chi/jwtauth/v5"
// )

// type Api struct {
// 	db               *database.DB
// 	router           *chi.Mux
// 	coins            []m.Coin
// 	currentDate      time.Time
// 	dayDuration      time.Duration
// 	coingeckoBaseUrl string
// 	listenAddress    string
// 	jwtAuth          *jwtauth.JWTAuth
// }

// func New(listenAddress string, coingeckoBaseUrl string) *Api {
// 	db := database.Init("api.db")
// 	virtualTime := time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
// 	priceRefreshInterval := time.Minute
// 	coins, err := db.GetCoins()
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	api := Api{
// 		db:               db,
// 		router:           chi.NewRouter(),
// 		currentDate:      virtualTime,
// 		dayDuration:      priceRefreshInterval,
// 		coins:            coins,
// 		coingeckoBaseUrl: coingeckoBaseUrl,
// 		listenAddress:    listenAddress,
// 		// CWE-547: Use of Hard-coded, Security-relevant Constants
// 		jwtAuth: jwtauth.New("HS256", []byte("safe-secret"), nil),
// 	}

// 	return &api
// }

// func (a *Api) Run() {
// 	go a.managePrices()
// 	a.setupRoutes()
// 	log.Println("Starting API ...")
// 	// CWE-319: Cleartext Transmission of Sensitive Information
// 	log.Fatalln(http.ListenAndServe(a.listenAddress, a.router))
// }

// func (a *Api) Shutdown() {
// 	a.db.Close()
// }

// func (a *Api) managePrices() {
// 	log.Println("Starting price management daemon ...")
// 	for {
// 		a.refreshCoins()
// 		a.currentDate = a.currentDate.Add(time.Hour * 24)
// 		time.Sleep(a.dayDuration)
// 	}
// }

// func (a *Api) refreshCoins() {
// 	var (
// 		coins []m.Coin
// 		r     *http.Response
// 		err   error
// 	)

// 	url := fmt.Sprintf("%s/coins/%v", a.coingeckoBaseUrl, a.currentDate.UnixMilli())

// 	for {
// 		r, err = http.Get(url)
// 		if err == nil {
// 			break
// 		}
// 		time.Sleep(time.Second)
// 	}

// 	json.NewDecoder(r.Body).Decode(&coins)
// 	a.coins = coins
// }

// func (a *Api) getCoin(coin_id string) (m.Coin, error) {
// 	for _, coin := range a.coins {
// 		if coin.Id == coin_id {
// 			return coin, nil
// 		}
// 	}
// 	return m.Coin{}, errors.New("Requested coin doesn't exist!")
// }


package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"govulnapi/api/database"
	m "govulnapi/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

// Api represents the main API structure that handles all HTTP requests and business logic
// This struct is designed to manage the cryptocurrency price API with virtual time simulation
type Api struct {
	// Database connection instance for persistent storage
	db               *database.DB
	// Router for handling HTTP routes and middleware
	router           *chi.Mux
	// List of cryptocurrencies being tracked
	coins            []m.Coin
	// Virtual time for simulating price changes across different dates
	currentDate      time.Time
	// Duration representing how fast virtual days pass in real time
	dayDuration      time.Duration
	// Base URL for CoinGecko API integration
	coingeckoBaseUrl string
	// Network address where the API server will listen
	listenAddress    string
	// JWT authentication handler for secure endpoints
	jwtAuth          *jwtauth.JWTAuth
}

// New creates and initializes a new Api instance with all necessary dependencies
// This function serves as the primary constructor for the API application
//
// Parameters:
//   - listenAddress: The network address for the HTTP server to listen on
//   - coingeckoBaseUrl: The base URL for CoinGecko API endpoints
//
// Returns:
//   - *Api: A fully initialized API instance ready to run
func New(listenAddress string, coingeckoBaseUrl string) *Api {
	// Step 1: Initialize the database connection
	// The database will store user data and application state
	db := database.Init("api.db")
	
	// Step 2: Set up virtual time for price simulation
	// We start from January 1, 2014 to simulate historical price data
	virtualTime := time.Date(2014, time.January, 1, 0, 0, 0, 0, time.UTC)
	
	// Step 3: Configure the price refresh interval
	// Each virtual day passes in this amount of real time
	priceRefreshInterval := time.Minute
	
	// Step 4: Retrieve the list of coins from the database
	// This ensures we have the initial cryptocurrency data
	coins, err := db.GetCoins()
	if err != nil {
		log.Fatalln(err)
	}

	// Step 5: Construct and return the API instance with all configured components
	api := Api{
		db:               db,
		router:           chi.NewRouter(),
		currentDate:      virtualTime,
		dayDuration:      priceRefreshInterval,
		coins:            coins,
		coingeckoBaseUrl: coingeckoBaseUrl,
		listenAddress:    listenAddress,
		// CWE-547: Use of Hard-coded, Security-relevant Constants
		// This JWT secret is hardcoded for demonstration purposes only
		// In production, this should be loaded from environment variables
		jwtAuth: jwtauth.New("HS256", []byte("safe-secret"), nil),
	}

	return &api
}

// Run starts the API server and begins all background processes
// This method is the main entry point for executing the application
func (a *Api) Run() {
	// Start the price management goroutine
	// This will run concurrently with the main server
	go a.managePrices()
	
	// Configure all HTTP routes and middleware
	a.setupRoutes()
	
	// Log that the server is starting
	log.Println("Starting API server on address:", a.listenAddress)
	
	// CWE-319: Cleartext Transmission of Sensitive Information
	// The server uses HTTP instead of HTTPS for local development
	// In production, this should be replaced with HTTPS
	log.Fatalln(http.ListenAndServe(a.listenAddress, a.router))
}

// Shutdown gracefully stops the API server and cleans up resources
// This method ensures all connections are properly closed before exit
func (a *Api) Shutdown() {
	// Close the database connection to prevent resource leaks
	log.Println("Initiating graceful shutdown...")
	a.db.Close()
	log.Println("Database connection closed successfully")
}

// managePrices handles the background process of updating cryptocurrency prices
// This daemon runs continuously and simulates the passage of time for price changes
func (a *Api) managePrices() {
	log.Println("Starting price management daemon with virtual time simulation...")
	
	// Infinite loop to continuously update prices
	for {
		// Refresh all coin prices from the external API
		a.refreshCoins()
		
		// Advance the virtual date by one day
		a.currentDate = a.currentDate.Add(time.Hour * 24)
		
		// Wait for the configured duration before the next update
		// This controls how fast virtual time passes
		time.Sleep(a.dayDuration)
	}
}

// refreshCoins retrieves the latest cryptocurrency prices from the CoinGecko API
// This method implements a retry mechanism for handling temporary network issues
func (a *Api) refreshCoins() {
	var (
		coins []m.Coin
		r     *http.Response
		err   error
	)

	// Construct the API URL with the current virtual timestamp
	// The timestamp is used to get historical price data
	url := fmt.Sprintf("%s/coins/%v", a.coingeckoBaseUrl, a.currentDate.UnixMilli())

	// Implement a retry loop for robust error handling
	// This ensures we eventually get the data even with temporary network issues
	for {
		r, err = http.Get(url)
		if err == nil {
			// Successfully connected to the API
			break
		}
		// Wait before retrying to avoid overwhelming the network
		time.Sleep(time.Second)
	}

	// Decode the JSON response into the coins slice
	// The response should contain price data for all tracked cryptocurrencies
	json.NewDecoder(r.Body).Decode(&coins)
	
	// Update the API's coin list with the fresh data
	a.coins = coins
}

// getCoin retrieves a specific cryptocurrency by its ID
// This method searches through the current list of coins and returns the matching one
//
// Parameters:
//   - coin_id: The unique identifier of the cryptocurrency to find
//
// Returns:
//   - m.Coin: The found cryptocurrency object
//   - error: An error if the coin doesn't exist or can't be found
func (a *Api) getCoin(coin_id string) (m.Coin, error) {
	// Iterate through all coins to find a match
	// This is a linear search, which is efficient enough for small lists
	for _, coin := range a.coins {
		// Check if the current coin's ID matches the requested one
		if coin.Id == coin_id {
			// Found the requested coin
			return coin, nil
		}
	}
	// If we get here, no matching coin was found
	// Return an empty coin and an appropriate error message
	return m.Coin{}, errors.New("Requested coin doesn't exist!")
}