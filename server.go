package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Структура клиента
type Client struct {
	conn     net.Conn
	ID       string
	Name     string
	Attempts int
}

// Структура сообщения
type Message struct {
	Type     string      `json:"type"`
	Data     interface{} `json:"data"`
	ClientID string      `json:"clientID,omitempty"`
}

// Глобальные переменные
var (
	clients       = make(map[string]*Client) // Карта клиентов
	leaderboard   = make(map[string]int)     // Таблица лидеров
	secretNumber  int                        // Загаданное число
	isGameStarted bool                       // Флаг начала игры
	mutex         sync.Mutex                 // Мьютекс для защиты доступа к клиентам
	serverAddress string                     // Адрес сервера
)

// Обработчик запроса от клиента
func handleClient(conn net.Conn) {
	defer conn.Close()

	// Генерация уникального ID для клиента
	clientID := strconv.Itoa(rand.Int())
	client := &Client{conn: conn, ID: clientID}

	// Сохранение клиента в карту
	mutex.Lock()
	clients[clientID] = client
	mutex.Unlock()

	// Чтение имени клиента
	reader := bufio.NewReader(conn)
	name, _ := reader.ReadString('\n')
	client.Name = strings.TrimSpace(name)
	fmt.Printf("Клиент %s с ID %s подключился\n", client.Name, client.ID)

	// Обработка сообщений от клиента
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println("Error reading message from client:", err)
			break
		}

		// Обработка сообщения
		handleMessage(conn, clientID, message)
	}

	// Удаление клиента из карты при отключении
	mutex.Lock()
	delete(clients, clientID)
	mutex.Unlock()
}

// Отправка сообщения всем клиентам
func broadcast(msgType string, data interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, client := range clients {
		send(client.conn, msgType, data, client.ID)
	}
}

// Отправка сообщения определенному клиенту
func sendToClient(clientID string, msgType string, data interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	client := clients[clientID]
	if client != nil {
		send(client.conn, msgType, data, clientID)
	}
}

// Отправка сообщения клиенту
func send(conn net.Conn, msgType string, data interface{}, clientID string) {
	msg := Message{Type: msgType, Data: data, ClientID: clientID}
	jsonMsg, _ := json.Marshal(msg)
	conn.Write([]byte(jsonMsg))
}

// Обработчик сообщений от клиента
func handleMessage(conn net.Conn, clientID string, message string) {
	message = strings.TrimSpace(message)
	var msg Message
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		log.Println("Error unmarshaling message:", err)
		return
	}

	switch msg.Type {
	case "guess":
		// Обработка предположения
		guess, _ := strconv.Atoi(msg.Data.(string))
		handleGuess(clientID, guess)
	}
}

// Функция для обработки предположения клиента
func handleGuess(clientID string, guess int) {
	client := clients[clientID]
	if client == nil {
		return
	}

	client.Attempts++
	if guess == secretNumber {
		sendToClient(clientID, "response", "Вы угадали!")
		leaderboard[client.Name] = client.Attempts
	} else if guess > secretNumber {
		sendToClient(clientID, "response", "Число больше загаданного")
	} else {
		sendToClient(clientID, "response", "Число меньше загаданного")
	}
}

// Функция для запуска нового эксперимента
func startGame() {
	// Загадываем новое число
	secretNumber = rand.Intn(100) + 1 // Число от 1 до 100

	// Отправляем сообщение о начале игры всем клиентам
	broadcast("startGame", "Игра началась!") // Исправленный вызов broadcast
	isGameStarted = true
	fmt.Println("Игра запущена!")
}

// Функция для вывода таблицы лидеров
func printLeaderboard() {
	fmt.Println("Таблица лидеров:")
	for name, attempts := range leaderboard {
		fmt.Printf("%s: %d попыток\n", name, attempts)
	}
}

func main() {
	// Запуск сервера
	listener, err := net.Listen("tcp", ":8089")
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	defer listener.Close()

	// Получаем адрес сервера
	serverAddress = listener.Addr().String()
	fmt.Printf("Сервер запущен. Адрес для подключения: %s\n", serverAddress)

	reader := bufio.NewReader(os.Stdin)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error accepting connection:", err)
		}

		go handleClient(conn)

		fmt.Print("Введите команду (start, leaderboard): ")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)

		switch command {
		case "start":
			if !isGameStarted { // Проверка, запущена ли игра
				startGame()
			} else {
				fmt.Println("Игра уже запущена.")
			}
		case "leaderboard":
			printLeaderboard()
		default:
			fmt.Println("Неизвестная команда.")
		}
	}
}
