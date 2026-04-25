package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count int // передаваемое значение count
		want  int // ожидаемое количество кафе в ответе
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: 100},
	}

	city := "moscow"

	for _, tc := range requests {
		// Формируем URL с параметрами
		url := fmt.Sprintf("/cafe?city=%s&count=%d", city, tc.count)

		// Создаем запрос и обработчик ответа
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		// Выполняем запрос и проверяем, что запрос успешно обработан
		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code,
			"a request with count=%d should return a 200 status", tc.count)

		// Получаем тело ответа и убираем пробелы и лишние переводы строк
		body := strings.TrimSpace(response.Body.String())

		var filteredCafes []string

		if body == "" {
			// Если тело ответа пустое, filteredCafes остается пустым слайсом
			filteredCafes = []string{}
		} else {
			// Разбиваем строку с кафе по запятым и удаляем пустые элементы
			cafes := strings.Split(body, ",")
			for _, cafe := range cafes {
				if cafe != "" {
					filteredCafes = append(filteredCafes, cafe)
				}
			}
		}

		// Определяем максимально возможное количество кафе для данного города
		maxCafes := len(cafeList[city])

		// Ожидаемое количество — минимум из запрошенного count и реального количества кафе
		expectedCount := tc.want
		if tc.want > maxCafes {
			expectedCount = maxCafes
		}

		// Проверяем количество возвращённых кафе
		assert.Equal(t, expectedCount, len(filteredCafes),
			"with count=%d, we expect %d cafes, but we received %d",
			tc.count, expectedCount, len(filteredCafes))
	}
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

	city := "moscow" 

	for _, tc := range requests {
		// Формируем URL с параметрами
		url := fmt.Sprintf("/cafe?city=%s&search=%s", city, tc.search)

		// Создаем запрос и обработчик ответа
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)


		// Выполняем запрос и проверяем, что запрос успешно обработан
		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code,
			"a request with search=%s should return a 200 status", tc.search)


		// Получаем тело ответа и убираем пробелы и лишние переводы строк
		body := strings.TrimSpace(response.Body.String())

		var filteredCafes []string

		if body == "" {
			// Если тело ответа пустое, filteredCafes остается пустым слайсом
			filteredCafes = []string{}
		} else {
			// Разбиваем строку с кафе по запятым и удаляем пустые элементы
			cafes := strings.Split(body, ",")

			for _, cafe := range cafes {
				if cafe != "" {
					filteredCafes = append(filteredCafes, cafe)
				}
			}
		}

		// Проверяем количество возвращённых кафе
		assert.Equal(t, tc.wantCount, len(filteredCafes),
			"when search=%s, we expect %d cafes, but we get %d",
			tc.search, tc.wantCount, len(filteredCafes))

		// Если кафе найдены, проверяем, что каждое содержит поисковую подстроку
		if tc.wantCount > 0 {
			// Приводим поисковый запрос к нижнему регистру для сравнения
			searchLower := strings.ToLower(tc.search)

			for _, cafe := range filteredCafes {
				cafeLower := strings.ToLower(cafe)
				assert.True(t, strings.Contains(cafeLower, searchLower),
					"cafe '%s' does not contain the substring '%s''", cafe, tc.search)
			}
		}
	}
}