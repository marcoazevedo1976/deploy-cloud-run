package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestCEPValidation(t *testing.T) {
	valid := []string{"78048250", "78048-250"}
	invalid := []string{"abc12345", "1234", "123456789"}

	for _, cep := range valid {
		assert.True(t, isValidCEP(cep))
	}
	for _, cep := range invalid {
		assert.False(t, isValidCEP(cep))
	}
}

func TestNotFoundCEP(t *testing.T) {
	req := httptest.NewRequest("GET", "/clima/00000-000", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/clima/{cep}", handleClima)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFoundCEP(t *testing.T) {
	cep := "78048-250" // Cuiabá, MT

	req := httptest.NewRequest("GET", "/clima/"+cep, nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/clima/{cep}", handleClima)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response TempResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)

	// Realizamos os cálculos de conversão manualmente
	calculatedTempF := response.TempC*1.8 + 32
	calculatedTempK := response.TempC + 273.15

	// Asseguramos que a conversão de Fahrenheit e Kelvin estão corretas
	assert.InDelta(t, calculatedTempF, response.TempF, 0.1, "Conversão de Fahrenheit está incorreta")
	assert.InDelta(t, calculatedTempK, response.TempK, 0.1, "Conversão de Kelvin está incorreta")
}
