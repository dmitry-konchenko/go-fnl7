package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

// Функция проверяет количество возвращаемых кафе
func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, min(100, len(cafeList["moscow"]))},
	}

	for _, v := range requests {
		t.Run(fmt.Sprintf("count=%d", v.count), func(t *testing.T) {
			request := fmt.Sprintf("/cafe?city=moscow&count=%d", v.count)
			response := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, request, nil)

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			body := response.Body.String()
			var cafes []string
			if body != "" {
				cafes = strings.Split(body, ",")
			}

			assert.Len(t, cafes, v.want)
		})
	}
}

// Функция проверяет поиск кафе по подстроке без учёта регистра
func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
		{"КОфЕ", 2},
	}

	for _, v := range requests {
		t.Run(v.search, func(t *testing.T) {
			request := "/cafe?city=moscow&search=" + url.QueryEscape(v.search)
			response := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, request, nil)

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			body := response.Body.String()
			var cafes []string
			if body != "" {
				cafes = strings.Split(body, ",")
			}

			assert.Len(t, cafes, v.wantCount)

			for _, cafe := range cafes {
				assert.True(t, strings.Contains(
					strings.ToLower(cafe),
					strings.ToLower(v.search),
				))
			}
		})
	}
}
