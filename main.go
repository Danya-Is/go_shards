package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

//	type initRequest struct {
//		maxVirtualShardAmount int
//	}
type addFilesRequest struct {
	SourceFileName string `json:"source_file_name"`
}

//
//type getFileRequest struct {
//	fileId  int
//	addTime time.Time
//}

func decodeRequestBody(w *http.ResponseWriter, r *http.Request, body interface{}) bool {
	err := decodeJSONBody(*w, r, &body)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(*w, mr.msg, mr.status)
		} else {
			http.Error(*w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return false
	}
	return true
}

func readIdsFromFile(filePath string) ([]int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var ids []int
	for scanner.Scan() {
		line := scanner.Text()
		num, err := strconv.Atoi(line)
		if err != nil {
			return nil, err
		}
		ids = append(ids, num)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Отображение текущих данных о корзинах
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>Baskets</h1>"))
	w.Write([]byte("<div id=\"shard-container\">"))
	w.Write([]byte("</div>"))
	w.Write([]byte(clientContent))
}

func (server *Server) initHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Initialized.")
	// var requestData initRequest
	// decodeRequestBody(w, r, &requestData)

	// Считывание данных из тела запроса
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %s", err)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error parsing form"})
		if err != nil {
			log.Printf("Error parsing error response: %s", err)
			return
		}
		return
	}

	maxVirtualShardAmountStr := r.FormValue("maxVirtualShardAmount")
	if maxVirtualShardAmountStr != "" {
		// Инициализация - создание начальных данных о шардах
		maxVirtualShardAmount, err := strconv.Atoi(maxVirtualShardAmountStr)
		if err != nil {
			log.Printf("Error parsing maxShardAmount: %s", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error parsing maxShardAmount"})
			return
		}
		server.Init(maxVirtualShardAmount)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"shardAmount": server.state.currentShardAmount,
	})
}

func (server *Server) addFilesHandler(w http.ResponseWriter, r *http.Request) {
	var body []byte
	r.Body.Read(body)
	log.Printf("Got ADD_FILES request %s", string(body))

	var requestData addFilesRequest
	if ok := decodeRequestBody(&w, r, &requestData); !ok {
		return
	}

	fileIds, err := readIdsFromFile(requestData.SourceFileName)
	if err != nil {
		log.Fatal(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	server.AddFiles(fileIds)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func (server *Server) getFileHandler(w http.ResponseWriter, r *http.Request) {
	// var requestData getFileRequest
	// decodeRequestBody(w, r, &requestData)
	log.Println("Got GET_FILE request")

	const timeFormat = "2006-01-02T15:04:05.000Z"

	// Считывание данных из тела запроса
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %s", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error parsing form"})
		return
	}

	// Получение dataId из формы
	fileIdStr := r.FormValue("fileId")
	if fileIdStr == "" {
		log.Printf("Error: fileId is empty")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "fileId is empty"})
		return
	}
	//if addTimeStr == "" {
	//	log.Printf("Error: addTime is empty")
	//	w.Header().Set("Content-Type", "application/json")
	//	json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "addTime is empty"})
	//	return
	//}
	//
	//addTime, err := time.Parse(timeFormat, addTimeStr)
	//if err != nil {
	//	log.Printf("Error parsing time")
	//	w.Header().Set("Content-Type", "application/json")
	//	json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error parsing time"})
	//	return
	//}
	fileId, err := strconv.Atoi(fileIdStr)
	if err != nil {
		log.Printf("Error parsing fileId")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error parsing fileId"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"shardId": server.GetFile(fileId),
	})
}

func (server *Server) addShardHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Got ADD_SHARD request")

	success := server.AddShard()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    success,
		"newShardId": server.state.currentShardAmount - 1,
	})
}

func (server *Server) getLogRecordHandler(w http.ResponseWriter, r *http.Request) {
	var logRecord FileLogRecord
	var success = true

	if len(server.fileLog) > 0 {
		logRecord = server.fileLog[0]
		server.fileLog = server.fileLog[1:]
	} else {
		logRecord = FileLogRecord{-1, -1}
		success = false
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"success": success,
			"fileId":  logRecord.fileId,
			"shardId": logRecord.shardId,
		})
}

func main() {
	server := Server{
		state: ServerState{
			currentShardAmount: 1,
			epochs: []Epoch{
				Epoch{time.Now(), 1},
			},
		},
		maxVirtualShardAmount: 65536,
		lastFileId:            -1,
		fileLog:               make([]FileLogRecord, 0, 1000),
		fileMap:               make(map[int]time.Time),
	}
	mux := http.NewServeMux()
	// todo is it a good idea to use routes for button click?
	//  maybe check url vars or route by info in body
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/init", server.initHandler)
	mux.HandleFunc("/add-basket", server.addShardHandler)
	mux.HandleFunc("/add-data", server.addFilesHandler)
	mux.HandleFunc("/get-data", server.getFileHandler)
	mux.HandleFunc("/get-log-record", server.getLogRecordHandler)

	log.Print("Starting server on :8000...")
	err := http.ListenAndServe(":8000", mux)
	log.Fatal(err)
}
