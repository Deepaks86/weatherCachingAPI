# weatherCachingAPI
# Weather API Caching

This repository contains two different weather API implementations with caching mechanisms:
1. **Simulated Weather API Caching** - Simulates random weather data for cities.
2. **Real-time Weather API Caching** - Fetches real-time weather data using the Weatherstack API, with caching for efficient retrieval.

Both implementations use a Least Recently Used (LRU) cache to store and serve weather data, reducing the number of requests made to the weather data source and improving performance.

## Table of Contents
- [Simulated Weather API Caching](#simulated-weather-api-caching)
- [Real-time Weather API Caching](#real-time-weather-api-caching)
- [How to Run](#how-to-run)
- [Environment Setup (for Real-time API)](#environment-setup-for-real-time-api)
- [Cache Structure](#cache-structure)
- [Contributing](#contributing)

---

## Simulated Weather API Caching

This implementation simulates weather data with random temperature and descriptions (e.g., Cold, Cool, Warm, Hot). The data is cached and served with an expiry time, and the Least Recently Used (LRU) cache ensures that the most recent weather data is retained.

### Features:
- Simulated weather data (random temperatures between 0 and 39°C).
- Weather descriptions based on temperature ranges.
- LRU caching mechanism to store weather data with expiry times.
- Cache eviction when the cache reaches its maximum size.

---

## Real-time Weather API Caching

This implementation fetches real-time weather data from the [Weatherstack API](https://weatherstack.com/), caching the results to avoid redundant API calls. The weather data is retrieved for cities via an HTTP request and includes the temperature and a weather description.

### Features:
- Fetches real-time weather data from Weatherstack API.
- Caches the weather data with an expiry time of 30 minutes.
- Cache eviction when the cache reaches its maximum size (100 entries).
- Serves weather data for a given city based on the query parameter `city`.

### External Dependencies:
- [Weatherstack API](https://weatherstack.com/) for real-time weather data.
- `github.com/joho/godotenv` for loading environment variables.

---

## How to Run

### Prerequisites:
- Go 1.18+ installed.
- For the **Real-time Weather API Caching** version, you will need to sign up at [Weatherstack](https://weatherstack.com/) and get an API key.

### Running the Simulated Weather API Caching:
cd simulatedForecasting

Run the server:

go run main.go

The server will start on http://localhost:8080. You can query the weather for a city like this:

curl "http://localhost:8080/weather?city=Pune"

### Running the Real-time Weather API Caching:

Sign up at https://weatherstack.com and get your API key.

cd realtimeForecasting
Create a .env file and add your Weatherstack API key:

Run the server:

go run main.go

The server will start on http://localhost:8080. You can query the weather for a city like this:

curl "http://localhost:8080/weather?city=Pune"

### Cache Structure

Both implementations use an LRU (Least Recently Used) cache to store weather data. The cache works as follows:

    A cache item stores the city name, weather data (temperature and description), and the timestamp when it was cached.
    When a city’s weather data is requested, the system first checks if the data is cached and whether it is still valid (not expired).
    If the data is not found or has expired, the system fetches new data (simulated or from the Weatherstack API).
    Once the data is retrieved, it is added to the cache.
    If the cache exceeds the maximum size, the least recently used data is evicted to make room for new data.
