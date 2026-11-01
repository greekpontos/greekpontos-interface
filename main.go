package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Order представляет структуру заказа
type Order struct {
	ID       int
	Customer string
	Products string
	Total    float64
	Status   string
}

// --- 1. АБСТРАКЦИИ (ИНТЕРФЕЙСЫ) ---

// RepositoryWriter отвечает исключительно за запись заказов (ISP, DIP)
type RepositoryWriter interface {
	Save(customer string, products []string, total float64) error
}

// Notifier определяет единый контракт для отправки уведомлений (LSP, DIP)
type Notifier interface {
	Send(customer string)
}

// --- 2. РЕАЛИЗАЦИЯ РЕПОЗИТОРИЯ ---

// SQLiteRepo инкапсулирует работу с конкретной БД SQLite (SRP)
type SQLiteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) *SQLiteRepo {
	return &SQLiteRepo{db: db}
}

func (r *SQLiteRepo) Save(customer string, products []string, total float64) error {
	_, err := r.db.Exec(
		"INSERT INTO orders (customer, products, total, status) VALUES (?, ?, ?, ?)",
		customer, fmt.Sprintf("%v", products), total, "pending",
	)
	return err
}

// --- 3. РЕАЛИЗАЦИЯ ОТПРАВИТЕЛЕЙ УВЕДОМЛЕНИЙ (OCP, LSP) ---

// EmailSender реализует отправку по электронной почте
type EmailSender struct{}

func (e *EmailSender) Send(customer string) {
	fmt.Printf("Email-уведомление отправлено клиенту %s\n", customer)
}

// SMSSender реализует отправку по SMS
type SMSSender struct{}

func (s *SMSSender) Send(customer string) {
	fmt.Printf("SMS-уведомление отправлено клиенту %s\n", customer)
}

// --- 4. ОСНОВНОЙ СЕРВИС (БИЗНЕС-ЛОГИКА) ---

// OrderService зависит только от абстракций, а не от реализаций (DIP, SRP)
type OrderService struct {
	repo     RepositoryWriter
	notifier Notifier
}

func NewOrderService(repo RepositoryWriter, notifier Notifier) *OrderService {
	return &OrderService{
		repo:     repo,
		notifier: notifier,
	}
}

func (s *OrderService) CreateOrder(customer string, products []string, total float64) error {
	// Сохранение через абстрактный репозиторий
	err := s.repo.Save(customer, products, total)
	if err != nil {
		return err
	}

	// Отправка через абстрактный нотификатор
	s.notifier.Send(customer)
	return nil
}

// --- 5. ДЕМОНСТРАЦИЯ РАБОТЫ ---

func main() {
	// Инициализация базы данных
	db, err := sql.Open("sqlite3", "orders.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		customer TEXT NOT NULL,
		products TEXT NOT NULL,
		total REAL NOT NULL,
		status TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Создаем конкретный репозиторий SQLite
	sqliteRepo := NewSQLiteRepo(db)

	fmt.Println("--- Тест 1: Работа с Email-уведомлениями ---")
	emailSender := &EmailSender{}
	// Внедряем зависимости в сервис
	serviceWithEmail := NewOrderService(sqliteRepo, emailSender)

	err = serviceWithEmail.CreateOrder("Иван", []string{"apple", "banana"}, 10.5)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n--- Тест 2: Работа с SMS-уведомлениями (без изменения кода сервиса) ---")
	smsSender := &SMSSender{}
	// Переиспользуем тот же репозиторий, но меняем тип уведомлений
	serviceWithSMS := NewOrderService(sqliteRepo, smsSender)

	err = serviceWithSMS.CreateOrder("Алексей", []string{"orange", "grape"}, 15.2)
	if err != nil {
		log.Fatal(err)
	}
}
