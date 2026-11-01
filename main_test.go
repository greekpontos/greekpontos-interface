package main

import (
	"errors"
	"testing"
)

// --- МОКИ ДЛЯ ТЕСТИРОВАНИЯ ---

// MockWriter имитирует базу данных для теста, чтобы не зависеть от SQLite и CGO
type MockWriter struct {
	ShouldFail bool
}

func (m *MockWriter) Save(customer string, products []string, total float64) error {
	if m.ShouldFail {
		return errors.New("ошибка базы данных")
	}
	return nil
}

// MockNotifier имитирует отправку уведомлений и запоминает, кому отправлено
type MockNotifier struct {
	CalledWithCustomer string
}

func (m *MockNotifier) Send(customer string) {
	m.CalledWithCustomer = customer
}

// --- ТЕСТЫ ---

// Тест успешного создания заказа
func TestCreateOrder_Success(t *testing.T) {
	// Инициализируем моки
	mockRepo := &MockWriter{ShouldFail: false}
	mockNotify := &MockNotifier{}

	// Внедряем моки в сервис вместо реальных БД и EmailSender
	service := NewOrderService(mockRepo, mockNotify)

	// Вызываем метод создания заказа
	err := service.CreateOrder("ТестКлиент", []string{"item1"}, 100.0)

	// Проверяем, что ошибок нет
	if err != nil {
		t.Fatalf("Ожидался успешный заказ, получена ошибка: %v", err)
	}

	// Проверяем, что нотификатор вызван с правильным именем клиента
	if mockNotify.CalledWithCustomer != "ТестКлиент" {
		t.Errorf("Ожидалось уведомление для ТестКлиент, но отправлено для: %s", mockNotify.CalledWithCustomer)
	}
}

// Тест сценария, когда база данных возвращает ошибку
func TestCreateOrder_DatabaseError(t *testing.T) {
	// Настраиваем мок репозитория на отдачу ошибки
	mockRepo := &MockWriter{ShouldFail: true}
	mockNotify := &MockNotifier{}

	service := NewOrderService(mockRepo, mockNotify)

	err := service.CreateOrder("Иван", []string{"apple"}, 10.5)

	// Проверяем, что сервис вернул ошибку
	if err == nil {
		t.Fatal("Ожидалась ошибка базы данных, но заказ создался успешно")
	}

	// Проверяем, что уведомление НЕ было отправлено при ошибке БД
	if mockNotify.CalledWithCustomer != "" {
		t.Error("Уведомление не должно отправляться, если запись в БД провалилась")
	}
}
