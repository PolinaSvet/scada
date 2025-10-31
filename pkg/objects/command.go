// objects/command.go
package objects

import (
	"encoding/json"
	"fmt"
	"log"
	"server-system/pkg/types"
)

// VueCommand - команда от Vue клиента
type VueCommand struct {
	Command  string                 `json:"command" msgpack:"command"`
	ObjectID string                 `json:"objectId" msgpack:"objectId"`
	UserID   string                 `json:"userId" msgpack:"userId"`
	ClientID string                 `json:"clientId" msgpack:"clientId"`
	Data     map[string]interface{} `json:"data" msgpack:"data"`
	Time     string                 `json:"time" msgpack:"time"`
}

type PendingCommand struct {
	CmdTag          map[string]interface{} `json:"cmdTag" msgpack:"cmdTag"`
	CmdValue        int                    `json:"cmdValue" msgpack:"cmdValue"`
	CmdType         string                 `json:"cmdType" msgpack:"cmdType"`
	CmdMess         string                 `json:"cmdMess" msgpack:"cmdMess"`
	CmdMessQuestion string                 `json:"cmdMessQuestion" msgpack:"cmdMessQuestion"`
	ObjId           string                 `json:"objId" msgpack:"objId"`
	ObjType         string                 `json:"objType" msgpack:"objType"`
}

func CommandExecute(message types.Message) {
	log.Printf("Received message: ID=%s, Type=%s, Source=%s", message.ID, message.Type, message.Source)

	// Обрабатываем только команды
	if message.Type == "command" {
		var vueCmd VueCommand

		// Декодируем данные из json.RawMessage
		if err := json.Unmarshal(message.Data, &vueCmd); err != nil {
			log.Printf("Failed to unmarshal VueCommand: %v, raw data: %s", err, string(message.Data))
			return
		}

		log.Printf("Command received: %+v", vueCmd)

		// Обрабатываем команду sendCommand
		if vueCmd.Command == "sendCommand" {
			processSendCommand(vueCmd)
		}

	} else {
		log.Printf("Unknown message type: %s", message.Type)
	}
}

func processSendCommand(vueCmd VueCommand) {
	// Логируем все полученные данные для отладки
	//log.Printf("Raw VueCommand: %+v", vueCmd)

	data := vueCmd.Data

	// Прямое извлечение значений из data
	if cmdValue, ok := data["cmdValue"]; ok {
		var intCmdValue int64

		// Обрабатываем разные возможные типы числа
		switch v := cmdValue.(type) {
		case int:
			intCmdValue = int64(v)
		case float64:
			intCmdValue = int64(v)
		case uint16:
			intCmdValue = int64(v)
		default:
			log.Printf("Unknown type for cmdValue: %T", cmdValue)
			return
		}

		codeCmd := (int(intCmdValue) >> 12) & 0xF
		idObj := int(intCmdValue) & 0xFFF

		objId, _ := data["objId"].(string)
		objType, _ := data["objType"].(string)
		cmdType, _ := data["cmdType"].(string)
		cmdMess, _ := data["cmdMess"].(string)
		cmdMessQuestion, _ := data["cmdMessQuestion"].(string)
		cmdTagData := parseCmdTag(data["cmdTag"])

		log.Printf("Parsed command: codeCmd=%d, idObj=%d, objId=%s, objType=%s",
			codeCmd, idObj, objId, objType)
		log.Printf("Message: %s, Question: %s", cmdMess, cmdMessQuestion)
		log.Printf("Command tag: %+v, cmdType: %+v", cmdTagData, cmdType)

	} else {
		log.Printf("cmdValue not found in data")
	}
}

// Функция для парсинга cmdTag (без изменений)
func parseCmdTag(cmdTag interface{}) map[string]string {
	result := make(map[string]string)

	if cmdTag == nil {
		return result
	}

	switch tag := cmdTag.(type) {
	case map[string]interface{}:
		for key, value := range tag {
			if strValue, ok := value.(string); ok {
				result[key] = strValue
			} else {
				result[key] = fmt.Sprintf("%v", value)
			}
		}
	default:
		log.Printf("Unknown cmdTag type: %T", cmdTag)
	}

	return result
}
