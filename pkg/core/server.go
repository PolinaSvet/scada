package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"server-system/pkg/modules/database"
	"server-system/pkg/types"
	"sync"
	"time"
)

func DataInit() {
	// 1. Загрузка конфига
	config := loadConfig("config.json")

	// 2. Создание каналов
	bufferSize := 1000
	genToDb := make(chan types.Message, bufferSize)
	dbToVue := make(chan types.Message, bufferSize)
	systemMessChan := make(chan types.Message, bufferSize)
	statusChan := make(chan types.Message, bufferSize)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// 3. Запуск обработчиков
	wg.Add(1)
	go func() {
		defer wg.Done()
		handleSystemMessages(ctx, systemMessChan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		handleStatus(ctx, statusChan)
	}()

	go generateTestData(genToDb)

	// 4. Запуск database модуля
	dbInit := database.DatabaseInit{
		Ctx:            ctx,
		SystemMessChan: systemMessChan,
		StatusChan:     statusChan,
		InputChan:      genToDb,
		OutputChan:     dbToVue,
		ConfigFile:     config["modules"].(map[string]interface{})["database"].(map[string]interface{})["config_file"].(string),
	}

	db := database.NewModule(dbInit)
	if db != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db.Start()
		}()
	}
	// test for reading data
	wg.Add(1)
	go func() {
		defer wg.Done()

		for _ = range dbToVue {
			//for msg := range dbToVue {
			//log.Println("get msg from dbToVue", msg.InitDT, msg.ID, len(msg.Data))
		}
	}()

	log.Println("Server started")
	wg.Wait()
}

func handleSystemMessages(ctx context.Context, messageChan <-chan types.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-messageChan:
			handleSystemMessage(msg)
		}
	}
}

func handleSystemMessage(msg types.Message) {

	var messageData types.MessageData
	if err := json.Unmarshal(msg.Data, &messageData); err != nil {
		log.Printf("❓ ERROR parsing message data from %s: %v", msg.Source, err)
		return
	}

	// Используем константы из types пакета
	switch msg.Type {
	case types.MessageTypeError:
		log.Printf("🚨 ERROR [%s]: %s", msg.Source, messageData.Message)

	case types.MessageTypeAlarm:
		log.Printf("🔴 ALARM [%s]: %s", msg.Source, messageData.Message)

	case types.MessageTypeWarning:
		log.Printf("🟡 WARNING [%s]: %s", msg.Source, messageData.Message)

	case types.MessageTypeInfo:
		log.Printf("🔵 INFO [%s]: %s", msg.Source, messageData.Message)

	case types.MessageTypeDebug:
		log.Printf("⚪ DEBUG [%s]: %s", msg.Source, messageData.Message)

	case types.MessageTypeStatus:
		log.Printf("🟢 STATUS [%s]: %s", msg.Source, messageData.Message)

	default:
		log.Printf("❓ UNKNOWN [%s] Type: %s, Data: %s", msg.Source, msg.Type, messageData.Message)
	}
}

/*
func handleErrors(ctx context.Context, errorChan <-chan types.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case errMsg := <-errorChan:
			log.Printf("ERROR: %s", string(errMsg.Data))
			// Сохранение в файл...
		}
	}
}*/

func handleStatus(ctx context.Context, statusChan <-chan types.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case statusMsg := <-statusChan:
			var status types.ServiceStatus
			json.Unmarshal(statusMsg.Data, &status)
			log.Printf("STATUS: %+v", status)
		}
	}
}

func loadConfig(filename string) map[string]interface{} {
	data, _ := os.ReadFile(filename)
	var config map[string]interface{}
	json.Unmarshal(data, &config)
	return config
}

// test gen data
func generateTestData(outputChan chan<- types.Message) {
	for j := 0; ; j++ {
		n := rand.Intn(1000)
		for i := 0; i <= n; i++ {
			tagValue := types.TagValue{
				Tag:       fmt.Sprintf("sensor_%d", i),
				Alias:     fmt.Sprintf("alias_sensor_%d", i%2),
				Value:     float64(i) * 10.0,
				Quality:   types.QualityGood,
				Timestamp: time.Now(),
				DataType:  types.DataTypeFLOAT32,
			}

			data, _ := json.Marshal([]types.TagValue{tagValue})
			msg := types.Message{
				ID:       fmt.Sprintf("msg_%d", i),
				Type:     "tag_data",
				Data:     data,
				InitDT:   time.Now(),
				UpdateDT: time.Now(),
				Source:   "generationdata",
			}

			outputChan <- msg
			//log.Printf("Generated message %d", i)
		}
		//t := rand.Intn(1000)
		time.Sleep(1000 * time.Millisecond)
	}
}
