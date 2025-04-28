package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
)

type TempResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type ViaCEPResponse struct {
	Localidade string `json:"localidade"`
	Erro       bool   `json:"erro,omitempty"`
}

// func init() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}
// }

func main() {
	r := chi.NewRouter()
	r.Get("/clima/{cep}", handleClima)

	fmt.Println("Servidor rodando em http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

func handleClima(w http.ResponseWriter, r *http.Request) {
	cep := chi.URLParam(r, "cep")

	if !isValidCEP(cep) {
		respondWithError(w, http.StatusUnprocessableEntity, "CEP inválido")
		return
	}

	city, err := getCityFromCEP(cep)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "CEP não encontrado")
		return
	}

	tempC, err := getTempFromWeatherAPI(city)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Dados do clima indisponíveis neste momento")
		return
	}

	response := TempResponse{
		TempC: tempC,
		TempF: tempC*1.8 + 32,
		TempK: tempC + 273.15,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func isValidCEP(cep string) bool {
	re := regexp.MustCompile(`^\d{5}-?\d{3}$`)
	return re.MatchString(cep)
}

func getCityFromCEP(cep string) (string, error) {
	endpoint := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", err
	}
	defer resp.Body.Close()

	var data ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Erro {
		return "", errors.New("invalid response")
	}

	return data.Localidade, nil
}

func getTempFromWeatherAPI(city string) (float64, error) {
	weatherAPIKey := getWeatherAPIKey()
	city = url.QueryEscape(city)
	endpoint := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&lang=pt", weatherAPIKey, city)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New(resp.Status)
	}

	var data WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	return data.Current.TempC, nil
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func getWeatherAPIKey() string {
	return "628669556f9145dfab1204009252704" //   os.Getenv("WEATHER_API_KEY")
}
