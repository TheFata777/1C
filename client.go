package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

// Структура сообщения
type Message struct {
	Type     string      `json:"type"`
	Data     interface{} `json:"data"`
	ClientID string      `json:"clientID,omitempty"`
}

// Обработчик сообщений
func handleMessage(conn net.Conn, message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Println("Error unmarshaling message:", err)
		return
	}

	switch msg.Type {
	case "startGame":
		fmt.Println("Игра началась!")
		// Запускаем цикл угадывания после получения "startGame"
		for {
			guess := getGuessFromUser()
			sendGuess(conn, guess)

			// Ожидаем ответ от сервера
			message, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				log.Println("Error reading message from server:", err)
				break
			}
			handleMessage(conn, []byte(message))
		}
	case "response":
		response := msg.Data.(string)
		fmt.Println(response)
		// После получения ответа - запрашиваем новое предположение
		// (Это уже находится в цикле в case "startGame")
	}
}

func main() {
	// Ввод адреса сервера
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Введите адрес сервера: ")
	serverAddress, _ := reader.ReadString('\n')
	serverAddress = strings.TrimSpace(serverAddress)

	// Создание соединения с сервером
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		log.Fatal("Error connecting to server:", err)
	}
	defer conn.Close()

	// Отправка сообщения о подключении
	fmt.Print("Введите ваше имя: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	conn.Write([]byte(name + "\n"))

	// Ожидание сообщения "startGame"
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println("Error reading message from server:", err)
			break
		}
		if strings.Contains(message, "startGame") { // Проверка на "startGame"
			handleMessage(conn, []byte(message))
			break // Выходим из цикла, после получения "startGame"
		}
	}
}

// Функция для отправки предположения
func sendGuess(conn net.Conn, guess int) {
	msg := Message{Type: "guess", Data: strconv.Itoa(guess)}
	jsonMsg, _ := json.Marshal(msg)
	conn.Write(append(jsonMsg, '\n'))
}

// Функция для ввода предположения
func getGuessFromUser() int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Введите ваше предположение: ")
	guessStr, _ := reader.ReadString('\n')
	guessStr = strings.TrimSpace(guessStr)
	guess, _ := strconv.Atoi(guessStr)
	return guess
}
