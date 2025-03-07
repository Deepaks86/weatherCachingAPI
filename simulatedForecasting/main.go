package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
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
var randomTemperature *rand.Rand

func init() {
	cache = Cache{
		data:        make(map[string]*list.Element),
		orderedList: list.New(),
		maxSize:     100, // Set a maximum size for the cache
		expiry:      30 * time.Minute,
	}
	// Initialize the random number generator with a new source.
	randomTemperature = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func getCityWeatherData(city string) CityWeatherData {
	// Simulate fetching weather data
	temperature := randomTemperature.Float64() * 40 // Random temperature between 0 and 39 degrees Celsius
	desc := ""                                      // Simulated weather description
	switch {
	case temperature >= 0 && temperature < 10:
		desc = "Cold"
	case temperature >= 10 && temperature < 20:
		desc = "Cool"
	case temperature >= 20 && temperature < 30:
		desc = "Warm"
	case temperature >= 30 && temperature < 40:
		desc = "Hot"
	default:
		desc = "Unknown"
	}
	temperature = float64(int(temperature*100)) / 100.0
	return CityWeatherData{
		City:      city,
		Temp:      temperature,
		Desc:      desc,
		CacheTime: time.Now(),
	}
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
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Simulate fetching new weather data
	newData := getCityWeatherData(city)

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
