package main

import (
	"math"
	"time"
	"encoding/json"
	"net/http"
	"log"
	"strconv"
)

type fileLogRecord struct {
	fileId int
	shardId int
}

type Epoch struct {
	startTime time.Time
	shardAmount int
}

type ServerState struct {
	epochs []Epoch
	currentShardAmount int
}

type Server struct {
	state ServerState
	maxVirtualShardAmount int
	lastFileId int
	fileLog []fileLogRecord
}

func (server *Server) init(maxVirtualShardAmount int) {
	server.maxVirtualShardAmount = maxVirtualShardAmount
}

func (server *Server) hashFunc(id int) int {
	return id % server.maxVirtualShardAmount
}

func (server *Server) shardId(virtualShardId int, shardAmount int) int {
	return int(math.Floor(float64(virtualShardId) /
	 (float64(server.maxVirtualShardAmount) / float64(shardAmount))))
}

func (server *Server) addFile(fileId int) (shardId int) {
	return server.shardId(server.hashFunc(fileId), server.state.currentShardAmount)
}

func (server *Server) addFiles(n int) {
	for i := 0; i < n; i++ {
		server.fileLog = append(server.fileLog,
			fileLogRecord{
				fileId: server.lastFileId,
				shardId: server.addFile(server.lastFileId + 1),
			})
	}
}

// todo add n shards
func (server *Server) addShard() bool {
	new_epochs := make([]Epoch, len(server.state.epochs))
	copy(new_epochs, server.state.epochs)
	server.state.epochs = append(new_epochs, Epoch{time.Now(), server.state.currentShardAmount + 1})
	// todo err handle
	return true
}

func (server *Server) getFile(fileId int, addTime time.Time) (shardId int) {
	virtualShardId := server.hashFunc(fileId)

	for i := len(server.state.epochs) - 1; i >=0; i-- {
		if server.state.epochs[i].startTime.Before(addTime) {
			return server.shardId(virtualShardId, server.state.epochs[i].shardAmount)
		}
	}
	return -1
}

type initRequest struct {
	maxVirtualShardAmount int
}

type addFilesRequest struct {
	fileAmount int
}

type getFileRequest struct {
	fileId int
	addTime time.Time
}

// func decodeRequestBody(w http.ResponseWriter, r *http.Request, body interface{}) {
// 	err := decodeJSONBody(w, r, &body)
// 	if err != nil {
//         var mr *malformedRequest
//         if errors.As(err, &mr) {
//             http.Error(w, mr.msg, mr.status)
//         } else {
//             http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
//         }
//         return
//     }
// }


func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Отображение текущих данных о корзинах
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>Baskets</h1>"))
	w.Write([]byte("<div id=\"shard-container\">"))
	w.Write([]byte("</div>"))
	w.Write([]byte(`
		<button onclick="addBasket()">Add Basket</button>
		<button onclick="addData()">Add Data</button>
		<button onclick="getData()">Get Data</button>
		<div id="logging"></div>
		<style>
			.highlight  { background-color:yellow; }
		</style>
		<script>
			function getBasketElement(basketNumber) {
				const basketDiv = document.createElement('div');
				basketDiv.class = 'basket';
				basketDiv.setAttribute('data-basketid', basketNumber);
				basketDiv.innerHTML = '<strong>Basket ' +  basketNumber + ' (<span class="len">' + 0 +'</span> items):</strong>';
				return basketDiv;
			}
			function addLog(text) {
				let loggingDiv = document.getElementById("logging");
				var newParagraph = document.createElement("p");
				newParagraph.innerHTML = text;

				// Вставляем новый элемент в начало <div>
				loggingDiv.insertBefore(newParagraph, loggingDiv.firstChild);
			}
			function addBasket() {
				fetch('/add-basket', { method: 'POST' })
					.then(response => response.json())
					.then(data => {
						if (data.success) {
							const newShardId = data.newShardId;
							const basketContainer = document.getElementById('shard-container');
							const newShardDiv = getBasketElement(newShardId, []);
							basketContainer.appendChild(newShardDiv);
							addLog("New basket added: " + newShardId);
						}
					})
					.catch(error => console.error('Error:', error));
			}

			function addData() {

				fetch('/add-data', {
					method: 'POST',
					headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
				})
					.then(response => response.json())
					.then(data => {
						if (data.success) {
							addLog("New data " + " added")
						}
					})
					.catch(error => console.error('Error:', error));
			}

			function getData() {
				const dataID = prompt('Enter the data ID:');
				if (!dataID) return;

				fetch('/get-data', {
					method: 'POST',
					headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
					body: 'dataId=' + dataID,
				})
					.then(response => response.json())
					.then(data => {
						if (data.success) {
							const shardId = data.shardId;
							const basketDiv = document.querySelector('div[data-basketid="' + shardId + '"]');
			
							// Добавляем класс для подсветки на две секунды
							basketDiv.classList.add('highlight');
							setTimeout(() => {
								// Удаляем класс подсветки после двух секунд
								basketDiv.classList.remove('highlight');
							}, 2000);
							addLog("Data " + dataID +" found in busket " + shardId)
						} else {
							alert('Error: ' + data.error);
						}
					})
					.catch(error => console.error('Error:', error));
			}

			function initBaskets() {
				fetch('/init', { method: 'POST' })
					.then(response => response.json())
					.then(data => {
						if (data.success) {
							const basketsContainer = document.getElementById('shard-container');
							basketsContainer.innerHTML = ''; // Очистка контейнера перед добавлением новых корзин
	
							amount = data.shardAmount || 0;
							for (let i = 0; i < amount; i++) {
								const basketElement = getBasketElement(i);
								basketsContainer.appendChild(basketElement);
							}
						}
					})
					.catch(error => console.error('Error:', error));
			}
	
			// Инициализация при загрузке страницы
			initBaskets();

		</script>
	`))
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
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error parsing form"})
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
		server.init(maxVirtualShardAmount)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"shardAmount": server.state.currentShardAmount})
}

func (server *Server) addFilesHandler(w http.ResponseWriter, r *http.Request) {
	// var requestData addFilesRequest
	// decodeRequestBody(w, r, &requestData)

	server.addFiles(1)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func (server *Server) getFileHandler(w http.ResponseWriter, r *http.Request) {
	// var requestData getFileRequest
	// decodeRequestBody(w, r, &requestData)

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
	fileIdStr, addTimeStr := r.FormValue("fileId"), r.FormValue("addTime")
	if fileIdStr == "" {
		log.Printf("Error: fileId is empty")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "fileId is empty"})
		return
	}
	if addTimeStr == "" {
		log.Printf("Error: addTime is empty")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "addTime is empty"})
		return
	}

	addTime, err := time.Parse(timeFormat, addTimeStr)
	if err != nil {
		log.Printf("Error parsing time")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error parsing time"})
		return
	}
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
		"shardId": server.getFile(fileId, addTime),
	})
}

func (server *Server) addShardHandler(w http.ResponseWriter, r *http.Request) {
	success := server.addShard()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"newShardId": server.state.currentShardAmount,
	})
}

func (server *Server) getLogRecordHandler(w http.ResponseWriter, r *http.Request) {
	var logRecord fileLogRecord
	if len(server.fileLog) > 0 {
		logRecord = server.fileLog[0]
		server.fileLog = server.fileLog[1:]
	} else {
		logRecord = fileLogRecord{-1, -1}
	}
	

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"success": true,
			"fileId": logRecord.fileId,
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
		lastFileId: -1,
		fileLog: make([]fileLogRecord, 0, 1000),
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