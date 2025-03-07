package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type CityWeatherData struct {
	City      string    `json:"city"`
	Temp      float64   `json:"temp"`
	Desc      string    `json:"desc"`
	CacheTime time.Time `json:"cache_time"`
}

type Cache struct {
	data        map[string]*list.Element
	orderedList *list.List
	maxSize     int
	expiry      time.Duration
	mu          sync.RWMutex
}

type cacheItem struct {
	city string
	data CityWeatherData
}

var cache Cache

func init() {
	cache = Cache{
		data:        make(map[string]*list.Element),
		orderedList: list.New(),
		maxSize:     100, //size for the cache
		expiry:      30 * time.Minute,
	}

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

// Fetch data from WeatherstackAPI
func fetchWeatherFromAPI(city string) (CityWeatherData, error) {
	// Retrieve the API key from environment variables
	apiKey := os.Getenv("WEATHERSTACK_API_KEY")
	if apiKey == "" {
		return CityWeatherData{}, fmt.Errorf("API key is missing")
	}

	// Create the URL for the API request
	url := fmt.Sprintf("http://api.weatherstack.com/current?access_key=%s&query=%s", apiKey, city)
	/*
	   Request URL: http://api.weatherstack.com/current?access_key=your_api_key_here&query=London
	   Raw Response:
	   {
	       "location": {
	           "name": "London",
	           "country": "United Kingdom",
	           "region": "England",
	           "lat": 51.5074,
	           "lon": -0.1278,
	           "timezone_id": "Europe/London",
	           "localtime": "2025-03-07 16:00",
	           "localtime_epoch": 1678209600
	       },
	       "current": {
	           "temperature": 15,
	           "weather_descriptions": [
	               "Partly cloudy"
	           ],
	           "wind_speed": 14,
	           "humidity": 82
	       }
	   }
	*/
	// Make the HTTP request to Weatherstack API
	resp, err := http.Get(url)
	if err != nil {
		return CityWeatherData{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return CityWeatherData{}, fmt.Errorf("API error: %s", resp.Status)
	}
	// Read and parse the JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CityWeatherData{}, err
	}
	var apiResponse struct {
		Current struct {
			Temperature          float64  `json:"temperature"`
			Weather_descriptions []string `json:"weather_descriptions"`
		} `json:"current"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return CityWeatherData{}, err
	}

	// Extract temperature and description from the API response
	temperature := apiResponse.Current.Temperature
	desc := ""
	if len(apiResponse.Current.Weather_descriptions) > 0 {
		desc = apiResponse.Current.Weather_descriptions[0]
	} else {
		desc = "No description available"
	}
	return CityWeatherData{
		City:      city,
		Temp:      temperature,
		Desc:      desc,
		CacheTime: time.Now(),
	}, nil
}

func getCityWeatherData(city string) (CityWeatherData, error) {
	// Fetch data from Weatherstack API
	weatherData, err := fetchWeatherFromAPI(city)
	if err != nil {
		return CityWeatherData{}, err
	}
	return weatherData, nil
}

func getCachedWeatherData(city string) (CityWeatherData, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	elem, exists := cache.data[city]
	if !exists {
		return CityWeatherData{}, false
	}
	// Move the accessed item to the front of the list (most recent)
	cache.orderedList.MoveToFront(elem)
	item := elem.Value.(*cacheItem)
	if time.Since(item.data.CacheTime) < cache.expiry {
		return item.data, true
	}

	// If expired, remove the item from cache
	cache.orderedList.Remove(elem)
	delete(cache.data, city)
	return CityWeatherData{}, false
}

func updateCache(city string, data CityWeatherData) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// If the cache is at maximum size, evict the least recently used item
	if cache.orderedList.Len() >= cache.maxSize {
		evictOldest()
	}

	// Add the new data to the cache
	item := &cacheItem{city: city, data: data}
	elem := cache.orderedList.PushFront(item)
	cache.data[city] = elem
}

func evictOldest() {
	// Evict the least recently used item (oldest in the list)
	oldest := cache.orderedList.Back()
	if oldest != nil {
		cache.orderedList.Remove(oldest)
		item := oldest.Value.(*cacheItem)
		delete(cache.data, item.city)
	}
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	// Get the 'city' query parameter
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "City parameter is required", http.StatusBadRequest)
		return
	}

	// Check if data is in cache and still valid
	cachedWeatherData, found := getCachedWeatherData(city)
	if found {
		// Serve from cache if data is valid
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cachedWeatherData); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		}
		return
	}
	// Fetch new weather data
	newData, err := getCityWeatherData(city)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch weather data: %v", err), http.StatusInternalServerError)
		return
	}

	// Update cache with the new data
	updateCache(city, newData)

	// Return the new data in JSON format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newData)
}

func main() {
	// Start the HTTP server
	http.HandleFunc("/weather", weatherHandler)

	// Serve on port 8080
	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
