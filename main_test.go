package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
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

func splitCafes(body string) []string {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil
	}
	return strings.Split(body, ",")
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}

	for _, v := range requests {
		t.Run(v.search, func(t *testing.T) {
			response := httptest.NewRecorder()
			// QueryEscape нужен, чтобы кириллица корректно попала в URL.
			req := httptest.NewRequest("GET", "/cafe?city=moscow&search="+url.QueryEscape(v.search), nil)
			handler.ServeHTTP(response, req)

			// Запрос должен быть успешно обработан.
			require.Equal(t, http.StatusOK, response.Code)

			cafes := splitCafes(response.Body.String())
			assert.Len(t, cafes, v.wantCount, "ответ сервера: %q", response.Body.String())

			// Каждое возвращённое кафе должно содержать строку поиска
			// без учёта регистра.
			for _, cafe := range cafes {
				assert.Contains(t, strings.ToLower(cafe), strings.ToLower(v.search))
			}
		})
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	moscowCount := len(cafeList["moscow"])

	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: min(100, moscowCount)},
	}

	for _, v := range requests {
		t.Run(fmt.Sprintf("count=%d", v.count), func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/cafe?city=moscow&count=%d", v.count), nil)
			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			cafes := splitCafes(response.Body.String())
			assert.Len(t, cafes, v.want, "ответ сервера: %q", response.Body.String())
		})
	}
}
